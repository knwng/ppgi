package graph

import (
	"errors"
	"fmt"

	nebula "github.com/vesoft-inc/nebula-go/v2"
)

type GraphReader interface {}

// NebulaReader uses nebula-go sdk to connect to nebula-graphd service
type NebulaReader struct {
    address     string  // address of nebula-graphd
    port        int     // port of nebula-graphd
    username    string
    password    string
    graphName   string  // name of graph(or space), all the operations are applied to this graph
    pool        *nebula.ConnectionPool
}

func NewNebulaReader(address string, port int,
        username, password, graphName string) (*NebulaReader, error) {
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

    return &NebulaReader{
        address: address,
        port: port,
        username: username,
        password: password,
        graphName: graphName,
        pool: pool,
    }, nil
}

func (s *NebulaReader) Close() {
    s.pool.Close()
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
func (s *NebulaReader) GetMultiNeighbors(name string, edgeProp map[string]string) (map[string][]string, error) {
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
            fmt.Printf("Unwrap the result of query '%s' failed, err: %s", queryList[i], err)
            continue
        }
        if result == nil {
            fmt.Printf("The result of query '%s' is empty, skip", queryList[i])
            continue
        }
        unwrappedDataDict[propList[i]] = make([]string, 0)
        for j, data := range result {
            unwrappedData, err := data.AsString()
            if err != nil {
                fmt.Printf("Unwrap data of (col %s, row %d) failed, err: %s, skip", propList[i], j, err)
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
func (s *NebulaReader) GetSingleNeighbor(name, edge, prop string) ([]string, error) {
    query := fmt.Sprintf(
                "USE %s; GO FROM hash(\"%s\") OVER %s YIELD properties($$).%s AS %s;",
                s.graphName, name, edge, prop, prop)

    result, err := s.Query(query)
    if err != nil {
        return nil, err
    }

    colNames := result.GetColNames()
    rowNum := result.GetRowSize()
    fmt.Printf("col name: %+v, num of row: %d\n", colNames, rowNum)

    unwrappedData := make([]string, rowNum)
    dataList, err := result.GetValuesByColName(prop)
    if err != nil {
        return nil, errors.New(fmt.Sprintf("Read data of col %s failed, err: %s, skip\n", prop, err))
    }

    for i, data := range dataList {
        unwrappedData[i], err = data.AsString()
        if err != nil {
            fmt.Printf("Read data of (col %s, row %d) failed, err: %s, skip\n", prop, i, err)
            continue
        }
    }

    return unwrappedData, nil
}

// MultiQuery executes a list of queries sequentially and return a list of ResultSets
func (s *NebulaReader) MultiQuery(queryList []string) ([]*nebula.ResultSet, error) {
    session, err := s.pool.GetSession(s.username, s.password)
    if err != nil {
        return nil, err
    }
    defer session.Release()

    resultSets := make([]*nebula.ResultSet, len(queryList))
    for i, query := range queryList {
        result, err := session.Execute(query)
        if err != nil {
            fmt.Printf("Failed to execute query '%s', err: %s, skip", query, err)
            resultSets[i] = nil
            continue
        }
        if err := checkResultSet(query, result); err != nil {
            fmt.Printf("Failed to check the result of query '%s', err: %s, skip", query, err)
            resultSets[i] = nil
            continue
        }
        resultSets[i] = result
    }
    return resultSets, nil
}

// Query executes a single query and return the ResultSet
func (s *NebulaReader) Query(query string) (*nebula.ResultSet, error) {
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
