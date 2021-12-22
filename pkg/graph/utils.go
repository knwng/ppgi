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

// func ValueWrapperToString(key string, val *nebula.ValueWrapper) ([]string, error) {
//     valType := val.GetType()
//     switch valType {
//     case "null":
//         return []string{key, "null", ""}, nil
//     case "bool":
//         val, err := val.AsBool()
//         if err != nil {
//             return []string{}, err
//         }
//         if val {
//             return []string{key, "bool", "true"}, nil
//         } else {
//             return []string{key, "bool", "false"}, nil
//         }
//     case "int":
//         val, err := val.AsInt()
//         if err != nil {
//             return []string{}, err
//         }
//         return []string{key, "int", strconv.FormatInt(val, 10)}, nil
//     case "float":
//         val, err := val.AsFloat()
//         if err != nil {
//             return []string{}, err
//         }
//         return []string{key, "float", strconv.FormatFloat(val, 'E', -1, 64)}, nil
//     case "string":
//         val, err := val.AsString()
//         if err != nil {
//             return []string{}, err
//         }
//         return []string{key, "string", val}, nil
//     case "date":
//         val, err := val.AsDate()
//         if err != nil {
//             return []string{}, err
//         }

//         dateStr := fmt.Sprintf("%d-%02d-%02d", val.Year, val.Month, val.Day)
//         return []string{key, "date", dateStr}, nil
//     case "time":
//         val, err := val.AsTime()
//         if err != nil {
//             return []string{}, err
//         }

//         dateStr := fmt.Sprintf("")
//     case "datetime":
//         val, err := val.AsDateTime()
//         if err != nil {
//             return []string{}, err
//         }
//     default:
//         return []string{}, errors.New(fmt.Sprintf("Type %s not implemented", valType))
//     }
// }
