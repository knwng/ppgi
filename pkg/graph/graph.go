package graph

import (
	"time"
)

// Graph Structure storage
type Node struct {
	Type         string   `yaml:"type"`
	RelatedEdges []string `yaml:"related_edges"`
	Props        []string `yaml:"props"`
	TimeProp     string   `yaml:"time_prop"`
	DataProp     string   `yaml:"data_prop"`
}

type Edge struct {
	Type     string   `yaml:"type"`
	Props    []string `yaml:"props"`
	TimeProp string   `yaml:"time_prop"`
}

type Graph struct {
	Nodes 			[]Node `yaml:"nodes"`
	Edges 			[]Edge `yaml:"edges"`
	ReverseNodeMap 	map[string]*Node
}

// Graph data storage
type VertexData struct{
    VID     string      `json:"vid"`
    Tag     string      `json:"tag"`
    Props   [][3]string `json:"props"`
}

type EdgeData struct{
    Source      string      `json:"source"`
    Destination string      `json:"destination"`
    Type        string      `json:"type"`
    Props       [][3]string `json:"props"`
}


func (s *Graph) GetAllNodeType() []string {
	nodeTypes := make([]string, len(s.Nodes))
	for i, node := range s.Nodes {
		nodeTypes[i] = node.Type
	}

	return nodeTypes
}

type GraphStrategy interface {
	LookupNodeID(nodes []Node, startTime, endTime *time.Time) ([]string, error)
	GetNeighbors(nodes []Node) ([]Node, []Edge, error)
}

type PrincipleNodeStrategy struct {

}
