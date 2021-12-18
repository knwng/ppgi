package graph

import (
	"errors"
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	nebula "github.com/vesoft-inc/nebula-go/v2"
)

type GraphReadWriter interface {}

type PrincipleNode struct {
    Name string
    Prop string
    Time string
}

func ParsePrincipleNodes(nodeStrs []string) ([]PrincipleNode, error) {
    nodes := make([]PrincipleNode, len(nodeStrs))
    for i, node := range nodeStrs {
        items := strings.Split(node, ".")
        if len(items) != 3 {
            return []PrincipleNode{}, errors.New(fmt.Sprintf("Principle node config format: {node_name}.{id_prop}.{time_prop}, but got: %s", node))
        }
        name, prop, time := items[0], items[1], items[2]
        nodes[i] = PrincipleNode{
            Name: name,
            Prop: prop,
            Time: time,
        }
    }
    return nodes, nil
}

// NebulaReader uses nebula-go sdk to connect to nebula-graphd service
type NebulaReadWriter struct {
    address         string  // address of nebula-graphd
    port            int     // port of nebula-graphd
    username        string
    password        string
    graphName       string  // name of graph(or space), all the operations are applied to this graph
    pool            *nebula.ConnectionPool
}

func NewNebulaReadWriter(address string, port int,
        username, password, graphName string) (*NebulaReadWriter, error) {
    log := nebula.DefaultLogger{}
    hostAddress := nebula.HostAddress{Host: address, Port: port}
    hostList := []nebula.HostAddress{hostAddress}
    defaultPoolConfig := nebula.GetDefaultConf()
    pool, err := nebula.NewConnectionPool(hostList, defaultPoolConfig, log)
    if err != nil {
        return nil, errors.New(fmt.Sprintf(
            "Fail to initialize the connection pool, host: %s, port: %d, %s",
            address, port, err))
    }

    return &NebulaReadWriter{
        address: address,
        port: port,
        username: username,
        password: password,
        graphName: graphName,
        pool: pool,
    }, nil
}

func (s *NebulaReadWriter) Close() {
    s.pool.Close()
}

func (s *NebulaReadWriter) LookupWithTimeLimit(node *PrincipleNode, timeRange [2]time.Time) ([]string, error) {
    startTime := fmt.Sprintf("datetime(\"%s\")", timeRange[0].Format("2006-01-02T15:04:05.000000"))
    endTime := fmt.Sprintf("datetime(\"%s\")", timeRange[1].Format("2006-01-02T15:04:05.000000"))

    timeProp := fmt.Sprintf("%s.%s", node.Name, node.Time)
    idProp := fmt.Sprintf("%s.%s", node.Name, node.Prop)

    query := fmt.Sprintf("USE %s; LOOKUP ON %s WHERE %s > %s and %s < %s YIELD %s;",
        s.graphName, node.Name, timeProp, startTime, timeProp, endTime, idProp)

    resultSet, err := s.Query(query)

    if err != nil {
        return []string{}, err
    }

    colNames := resultSet.GetColNames()
    rowNum := resultSet.GetRowSize()
    log.WithFields(log.Fields{
        "col_names": colNames,
        "num_row": rowNum,
        "result": resultSet,
    }).Debug("Got results")

    unwrappedData := make([]string, rowNum)
    dataList, err := resultSet.GetValuesByColName(idProp)
    if err != nil {
        return nil, errors.New(fmt.Sprintf("Reading data of col %s failed, err: %s\n", idProp, err))
    }

    for i, data := range dataList {
        unwrappedData[i], err = data.AsString()
        if err != nil {
            log.WithFields(log.Fields{
                "col": idProp,
                "row": i,
                "error": err,
            }).Warn("Reading data failed, skip")
            continue
        }
    }

    return unwrappedData, nil
}

/*
GetMultiNeighbors gets all the neighbors of a given node.
Params:
    @name: String, node's name, such as '121904329086390421'
    @edgeProp: map[string]string, whose key is the name of
               edge and value is the name of neighbor's property. For example,
               {'id_email': 'email', 'id_telephone':'telephone',
               'id_province':'province'}

Return:
    @: map[string][]string, whose key is the property of neighbor and value is
       the data list, assuming all the data is string. For example,
       {'email': ['abc@gmail'], 'telephone': ['13133445432'],
       'province': ['Shandong', 'Beijing']}
*/
func (s *NebulaReadWriter) GetMultiNeighbors(name string, edgeProp map[string]string) (map[string][]string, error) {
    propList := make([]string, 0)
    queryList := make([]string, 0)
    for edge, prop := range edgeProp {
        propList = append(propList, prop)
        queryList = append(queryList, fmt.Sprintf(
            "USE %s; GO FROM hash(\"%s\") OVER %s YIELD properties($$).%s AS %s;",
            s.graphName, name, edge, prop, prop))
    }

    resultSetList, err := s.MultiQuery(queryList)
    if err != nil {
        return nil, err
    }

    unwrappedDataDict := make(map[string][]string)
    for i, resultSet := range resultSetList {
        result, err := resultSet.GetValuesByColName(propList[i])
        if err != nil {
            log.WithFields(log.Fields{
                "query": queryList[i],
                "error": err,
            }).Warn("Unwraping the result of query failed")
            continue
        }
        if result == nil {
            log.WithField("query", queryList[i]).Warn("The result of query is empty")
            continue
        }
        unwrappedDataDict[propList[i]] = make([]string, 0)
        for j, data := range result {
            unwrappedData, err := data.AsString()
            if err != nil {
                log.WithFields(log.Fields{
                    "col": propList[i],
                    "row": j,
                    "error": err,
                }).Warn("Unwraping data failed")
            }
            unwrappedDataDict[propList[i]] = append(unwrappedDataDict[propList[i]], unwrappedData)
        }
    }

    return unwrappedDataDict, nil
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

func checkResultSet(prefix string, res *nebula.ResultSet) error {
    if !res.IsSucceed() {
        return errors.New(fmt.Sprintf("%s, ErrorCode: %v, ErrorMsg: %s",
                          prefix, res.GetErrorCode(), res.GetErrorMsg()))
    }
    return nil
}
