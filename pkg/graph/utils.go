package graph

import (
	"fmt"
	"errors"
	"strings"

	nebula "github.com/vesoft-inc/nebula-go/v2"
)

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

func getIntFromCol(row *nebula.Record, colName string) (int64, error) {
    raw, err := row.GetValueByColName(colName)
    if err != nil {
        return -1, err
    }

    data, err := raw.AsInt()
    if err != nil {
        return -1, err
    }

    return data, nil
}

func getStringFromCol(row *nebula.Record, colName string) (string, error) {
    raw, err := row.GetValueByColName(colName)
    if err != nil {
        return "", err
    }

    data, err := raw.AsString()
    if err != nil {
        return "", err
    }

    return data, nil
}

func getMapFromCol(row *nebula.Record, colName string) ([][3]string, error) {
    raw, err := row.GetValueByColName(colName)
    if err != nil {
        return [][3]string{}, err
    }

    data, err := raw.AsMap()
    if err != nil {
        return [][3]string{}, err
    }

    ret := make([][3]string, 0)
    for key, prop := range data {
        ret = append(ret, [3]string{key, prop.GetType(), prop.String()})
    }

    return ret, nil
}

func varToNebulaExpr(vType, val string) string {
    switch vType {
    case "string":
        return val // not fmt.Sprintf("\"%s\"", val)
    case "date":
        return fmt.Sprintf("date(\"%s\")", val)
    case "time":
        return fmt.Sprintf("time(\"%s\")", val)
    case "datetime":
        return fmt.Sprintf("datetime(\"%s\")", val)
    default:
        return val
    }
}

func checkResultSet(prefix string, res *nebula.ResultSet) error {
    if !res.IsSucceed() {
        return errors.New(fmt.Sprintf("%s, ErrorCode: %v, ErrorMsg: %s",
                          prefix, res.GetErrorCode(), res.GetErrorMsg()))
    }
    return nil
}
