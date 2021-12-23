package graph

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	nebula "github.com/vesoft-inc/nebula-go/v2"
)

// NebulaReader uses nebula-go sdk to connect to nebula-graphd service
type NebulaReadWriter struct {
    address         string  // address of nebula-graphd
    port            int     // port of nebula-graphd
    username        string
    password        string
    graphName       string  // name of graph(or space), all the operations are applied to this graph
    neighborSteps   []int
    pool            *nebula.ConnectionPool
}

func NewNebulaReadWriter(address string, port int,
        username, password, graphName string, neighborSteps []int) (*NebulaReadWriter, error) {
    nebulaLog := nebula.DefaultLogger{}
    hostAddress := nebula.HostAddress{Host: address, Port: port}
    hostList := []nebula.HostAddress{hostAddress}
    defaultPoolConfig := nebula.GetDefaultConf()
    pool, err := nebula.NewConnectionPool(hostList, defaultPoolConfig, nebulaLog)
    if err != nil {
        return nil, errors.New(fmt.Sprintf(
            "Fail to initialize the connection pool, host: %s, port: %d, %s",
            address, port, err))
    }

    if len(neighborSteps) > 2 {
        log.WithField("neighbor_steps", neighborSteps).Fatal("neighbor_steps should have only 1 or 2 elements.")
    }

    return &NebulaReadWriter{
        address: address,
        port: port,
        username: username,
        password: password,
        graphName: graphName,
        neighborSteps: neighborSteps,
        pool: pool,
    }, nil
}

func (s *NebulaReadWriter) Close() {
    s.pool.Close()
}

func (s *NebulaReadWriter) LookupWithTimeLimit(node *Node, startTime, endTime *time.Time) ([]string, error) {
    timeProp := fmt.Sprintf("%s.%s", node.Type, node.TimeProp)
    var (
        idProp, yield string
        useVID bool
    )
    if len(node.DataProp) == 0 {
        idProp = "VertexID"
        yield = ""
        useVID = true
    } else {
        idProp = fmt.Sprintf("%s.%s", node.Type, node.DataProp)
        yield = fmt.Sprintf("YIELD %s", idProp)
        useVID = false
    }

    var query string
    if endTime == nil {
        return []string{}, errors.New("endTime should not be nil")
    }
    if startTime != nil {
        startTimeStr := fmt.Sprintf("datetime(\"%s\")", startTime.Format("2006-01-02T15:04:05.000000"))
        endTimeStr := fmt.Sprintf("datetime(\"%s\")", endTime.Format("2006-01-02T15:04:05.000000"))

        query = fmt.Sprintf("USE %s; LOOKUP ON %s WHERE %s > %s and %s < %s %s;",
        s.graphName, node.Type, timeProp, startTimeStr, timeProp, endTimeStr, yield)
    } else {
        endTimeStr := fmt.Sprintf("datetime(\"%s\")", endTime.Format("2006-01-02T15:04:05.000000"))
        query = fmt.Sprintf("USE %s; LOOKUP ON %s WHERE %s < %s %s;",
        s.graphName, node.Type, timeProp, endTimeStr, yield)
    }

    resultSet, err := s.Query(query)
    if err != nil {
        return []string{}, err
    }

    colNames := resultSet.GetColNames()
    rowNum := resultSet.GetRowSize()
    log.WithFields(log.Fields{
        "query": query,
        "col_names": colNames,
        "num_row": rowNum,
        "result": resultSet,
    }).Debug("Got results")

    unwrappedData := make([]string, rowNum)
    dataList, err := resultSet.GetValuesByColName(idProp)
    if err != nil {
        return nil, errors.New(fmt.Sprintf("Failed to read data at col %s, err: %s\n", idProp, err))
    }

    for i, data := range dataList {
        var err error
        if useVID {
            var vidInt int64
            vidInt, err = data.AsInt()
            unwrappedData[i] = strconv.FormatInt(vidInt, 10)
        } else {
            unwrappedData[i], err = data.AsString()
        }

        if err != nil {
            log.WithFields(log.Fields{
                "col": idProp,
                "row": i,
                "error": err,
            }).Warn("Failed to read data from value wrapper, skip")
            continue
        }
    }

    return unwrappedData, nil
}

func (s *NebulaReadWriter) getVertices(ids []string, vertexRef string) (map[string]VertexData, error) {
    var steps string
    if len(s.neighborSteps) == 2 {
        steps = fmt.Sprintf("%d TO %d", s.neighborSteps[0], s.neighborSteps[1])
    } else {
        steps = fmt.Sprintf("%d", s.neighborSteps[0])
    }

    yield := fmt.Sprintf("id(%[1]s) AS src_id, properties(%[1]s) AS src_prop," +
                         " head(tags(%[1]s)) AS src_tag;", vertexRef)

    query := fmt.Sprintf("USE %s; GO %s STEPS FROM %s OVER * BIDIRECT YIELD " +
                         "DISTINCT %s", s.graphName, steps,
                         strings.Join(ids, ","), yield)

    result, err := s.Query(query)
    if err != nil {
        log.WithFields(log.Fields{
            "query": query,
            "error": err,
        }).Error("Failed to execute query")
        return nil, err
    }

    colNames := result.GetColNames()
    rowNum := result.GetRowSize()
    log.WithFields(log.Fields{
        "col_names": colNames,
        "row_num": rowNum,
    }).Debug()

    ret := make(map[string]VertexData)

    for i := 0; i < rowNum; i++ {
        row, err := result.GetRowValuesByIndex(i)
        if err != nil {
            log.WithFields(log.Fields{
                "row": i,
                "error": err,
            }).Warn("Failed to get row from result data, skip")
            continue
        }

        vidInt, err := getIntFromCol(row, "src_id")
        if err != nil {
            log.WithFields(log.Fields{
                "row_data": row,
                "error": err,
            }).Warn("Failed to get src_id col from row, skip")
            continue
        }
        vid := strconv.FormatInt(vidInt, 10)

        props, err := getMapFromCol(row, "src_prop")
        if err != nil {
            log.WithFields(log.Fields{
                "row_data": row,
                "error": err,
            }).Warn("Failed to get src_id col from row, skip")
            continue
        }

        tag, err := getStringFromCol(row, "src_tag")
        if err != nil {
            log.WithFields(log.Fields{
                "row_data": row,
                "error": err,
            }).Warn("Failed to get src_tag col from row, skip")
            continue
        }

        ret[vid] = VertexData{
            VID: vid,
            Tag: tag,
            Props: props,
        }
        
    }

    return ret, nil
}

func (s *NebulaReadWriter) GetAllNeighborVertices(ids []string) (map[string]VertexData, error) {
    unwrappedData := make(map[string]VertexData, 0)

    srcData, err := s.getVertices(ids, "$^")
    if err != nil {
        return nil, err
    }

    for k, v := range srcData {
        unwrappedData[k] = v
    }

    dstData, err := s.getVertices(ids, "$$")
    if err != nil {
        return nil, err
    }

    for k, v := range dstData {
        unwrappedData[k] = v
    }

    return unwrappedData, nil
}

func (s *NebulaReadWriter) GetAllNeighborEdges(ids []string) ([]EdgeData, error) {
    var steps string
    if len(s.neighborSteps) == 2 {
        steps = fmt.Sprintf("%d TO %d", s.neighborSteps[0], s.neighborSteps[1])
    } else {
        steps = fmt.Sprintf("%d", s.neighborSteps[0])
    }

    query := fmt.Sprintf("USE %s; GO %s STEPS FROM %s OVER * BIDIRECT YIELD " +
                         "DISTINCT src(edge) AS edge_src, dst(edge) AS edge_dst, " +
                         "type(edge) AS edge_type, properties(edge) AS edge_prop",
                         s.graphName, steps, strings.Join(ids, ","))

    result, err := s.Query(query)
    if err != nil {
        return nil, err
    }

    colNames := result.GetColNames()
    rowNum := result.GetRowSize()
    log.WithFields(log.Fields{
        "col_names": colNames,
        "row_num": rowNum,
    }).Debug()

    unwrappedData := make([]EdgeData, 0)
    for i := 0; i < rowNum; i++ {
        row, err := result.GetRowValuesByIndex(i)
        if err != nil {
            log.WithFields(log.Fields{
                "row": i,
                "error": err,
            }).Warn("Failed to get row from result data, skip")
            continue
        }

        src, err := getIntFromCol(row, "edge_src")
        if err != nil {
            log.WithFields(log.Fields{
                "row_data": row,
                "error": err,
            }).Warn("Failed to get edge_src col from row, skip")
            continue
        }

        dst, err := getIntFromCol(row, "edge_dst")
        if err != nil {
            log.WithFields(log.Fields{
                "row_data": row,
                "error": err,
            }).Warn("Failed to get edge_dst col from row, skip")
            continue
        }

        eType, err := getStringFromCol(row, "edge_type")
        if err != nil {
            log.WithFields(log.Fields{
                "row_data": row,
                "error": err,
            }).Warn("Failed to get edge_type col from row, skip")
            continue
        }

        props, err := getMapFromCol(row, "edge_prop")
        if err != nil {
            log.WithFields(log.Fields{
                "row_data": row,
                "error": err,
            }).Warn("Failed to get edge_prop col from row, skip")
            continue
        }

        unwrappedData = append(unwrappedData, EdgeData{
            Source: strconv.FormatInt(src, 10),
            Destination: strconv.FormatInt(dst, 10),
            Type: eType,
            Props: props,
        })

    }

    return unwrappedData, nil
}

func (s *NebulaReadWriter) AddVertexData(vertices []VertexData) error {
    // classify vertex according to its tag
    vertexDefine := make(map[string]string)
    values := make(map[string][]string)

    for _, v := range vertices {
        if _, ok := vertexDefine[v.Tag]; !ok {
            propNames := make([]string, len(v.Props))
            for i, props := range v.Props {
                propNames[i] = props[0]
            }
            vertexDefine[v.Tag] = fmt.Sprintf("%s (%s)", v.Tag, strings.Join(propNames, ", "))
        }
        if _, ok := values[v.Tag]; !ok {
            values[v.Tag] = make([]string, 0)
        }

        propDatas := make([]string, len(v.Props))
        for i, props := range v.Props {
            // Prop: (key, type, value)
            propDatas[i] = varToNebulaExpr(props[1], props[2])
        }

        values[v.Tag] = append(values[v.Tag], fmt.Sprintf("hash(\"%s\"):(%s)", v.VID, strings.Join(propDatas, ", ")))
    }

    for tag, define := range vertexDefine {
        query := fmt.Sprintf("USE %s; INSERT VERTEX IF NOT EXISTS %s VALUES %s;", s.graphName, define, strings.Join(values[tag], ", "))

        _, err := s.Query(query)
        if err != nil {
            log.WithFields(log.Fields{
                "query": query,
                "error": err,
            }).Error("Failed to insert vertex data to graph database")
            return err
        }
    }

    return nil
}

func (s *NebulaReadWriter) AddEdgeData(edges []EdgeData) error {
    edgeDefine := make(map[string]string)
    values := make(map[string][]string)

    for _, e := range edges {
        if _, ok := edgeDefine[e.Type]; !ok {
            propNames := make([]string, len(e.Props))
            for i, props := range e.Props {
                propNames[i] = props[0]
            }
            edgeDefine[e.Type] = fmt.Sprintf("%s (%s)", e.Type, strings.Join(propNames, ", "))
        }
        if _, ok := values[e.Type]; !ok {
            values[e.Type] = make([]string, 0)
        }

        propDatas := make([]string, len(e.Props))
        for i, props := range e.Props {
            // Prop: (key, type, value)
            propDatas[i] = varToNebulaExpr(props[1], props[2])
        }

        edgeVIDs := fmt.Sprintf("hash(\"%s\") -> hash(\"%s\")", e.Source, e.Destination)

        values[e.Type] = append(values[e.Type], fmt.Sprintf("%s:(%s)", edgeVIDs, strings.Join(propDatas, ", ")))
    }

    for eType, define := range edgeDefine {
        query := fmt.Sprintf("USE %s; INSERT EDGE IF NOT EXISTS %s VALUES %s;", s.graphName, define, strings.Join(values[eType], ", "))

        _, err := s.Query(query)
        if err != nil {
            log.WithFields(log.Fields{
                "query": query,
                "error": err,
            }).Error("Failed to insert edge data to graph database")
            return err
        }
    }

    return nil
}

/*
GetSingleNeighbor gets a specific kind of neighbor of a given node
Params:
    @name: String, the name of given node
    @edge: String, the name of edge to the desired neighbor, for example, 'id_email'
    @prop: String, the property of neighbor, for example, 'email'

Return:
*/
func (s *NebulaReadWriter) GetSingleNeighbor(name, edge, prop string) ([]string, error) {
    query := fmt.Sprintf(
                "USE %s; GO FROM hash(\"%s\") OVER %s YIELD properties($$).%s AS %s;",
                s.graphName, name, edge, prop, prop)

    result, err := s.Query(query)
    if err != nil {
        return nil, err
    }

    colNames := result.GetColNames()
    rowNum := result.GetRowSize()
    log.WithFields(log.Fields{
        "col_names": colNames,
        "row_num": rowNum,
    }).Debug()

    unwrappedData := make([]string, rowNum)
    dataList, err := result.GetValuesByColName(prop)
    if err != nil {
        return nil, errors.New(fmt.Sprintf("Read data of col %s failed, err: %s, skip\n", prop, err))
    }

    for i, data := range dataList {
        unwrappedData[i], err = data.AsString()
        if err != nil {
            log.WithFields(log.Fields{
                "col": prop,
                "row": i,
                "error": err,
            }).Warn("Reading data failed, skip")
            continue
        }
    }

    return unwrappedData, nil
}

// MultiQuery executes a list of queries sequentially and return a list of ResultSets
func (s *NebulaReadWriter) MultiQuery(queryList []string) ([]*nebula.ResultSet, error) {
    session, err := s.pool.GetSession(s.username, s.password)
    if err != nil {
        return nil, err
    }
    defer session.Release()

    resultSets := make([]*nebula.ResultSet, len(queryList))
    for i, query := range queryList {
        result, err := session.Execute(query)
        if err != nil {
            log.WithFields(log.Fields{
                "query": query,
                "error": err,
            }).Warn("Failed to execute query, skip")
            resultSets[i] = nil
            continue
        }
        if err := checkResultSet(query, result); err != nil {
            log.WithFields(log.Fields{
                "query": query,
                "error": err,
            }).Warn("Failed to check the result of query")
            resultSets[i] = nil
            continue
        }
        resultSets[i] = result
    }
    return resultSets, nil
}

// Query executes a single query and return the ResultSet
func (s *NebulaReadWriter) Query(query string) (*nebula.ResultSet, error) {
    session, err := s.pool.GetSession(s.username, s.password)
    if err != nil {
        return nil, err
    }
    defer session.Release()

    result, err := session.Execute(query)
    if err != nil {
        return nil, err
    }
    if err := checkResultSet(query, result); err != nil {
        return nil, err
    }

    return result, nil
}
