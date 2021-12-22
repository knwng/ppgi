package graph

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestQuery(t *testing.T) {
	target := []string{"efkhb@gmail.com", "uYAcuqoX@gmail.com",
					   "CDWc@gmail.com", "eDJWvJZwUk@gmail.com",
					   "ieBaArcxUR@gmail.com", "uLi@gmail.com"}

	nebula, err := NewNebulaReadWriter("192.168.31.147", 9669, "root", "nebula", "relation_graph", []int{1, 2})
	assert.NoError(t, err)
	defer nebula.Close()

	data, err := nebula.GetSingleNeighbor("121904329086390421", "id_email", "email")
	assert.NoError(t, err)
	t.Logf("data: %+v", data)
	
	assert.Equal(t, target, data)
}

// TODO(knwng): Fix this test later
// func TestMultiQuery(t *testing.T) {
// 	target := make(map[string][]string)
// 	target["email"] = []string{"efkhb@gmail.com", "uYAcuqoX@gmail.com",
// 							   "CDWc@gmail.com", "eDJWvJZwUk@gmail.com",
// 							   "ieBaArcxUR@gmail.com", "uLi@gmail.com"}
	
// 	target["province"] = []string{"Shandong", "Hainan", "Beijing", "Shanghai",
// 								  "Liaoning", "Chongqing"}

// 	target["telephone"] = []string{"10927866150", "14326585168", "18595097300",
// 								   "11660492056", "13992190315"}

// 	nebula, err := NewNebulaReadWriter("192.168.31.147", 9669, "root", "nebula", "relation_graph", []int{1, 2})
// 	assert.NoError(t, err)
// 	defer nebula.Close()

// 	edgeProp := map[string]string{
// 		"id_email": "email",
// 		"id_telephone": "telephone",
// 		"id_province": "province",
// 	}

// 	data, err := nebula.GetMultiNeighbors("121904329086390421", edgeProp)
// 	assert.NoError(t, err)
// 	t.Logf("data: %+v", data)

// 	assert.Equal(t, target, data)
// }

func TestLookup(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Shanghai")
	assert.NoError(t, err)

	startTime := time.Date(2021, 12, 18, 23, 54, 0, 0, loc)
	endTime := time.Date(2021, 12, 18, 23, 55, 0, 0, loc)

	node := Node{
		Type: "id",
		RelatedEdges: []string{"id_email", "id_telephone", "id_province"},
		Props: []string{"data", "collect_time"},
		// DataProp: "data",
		TimeProp: "collect_time",
	}

	nebula, err := NewNebulaReadWriter("192.168.31.147", 9669, "root", "nebula", "relation_graph", []int{1, 2})
	assert.NoError(t, err)
	defer nebula.Close()

	data, err := nebula.LookupWithTimeLimit(&node, &startTime, &endTime)
	assert.NoError(t, err)
	fmt.Printf("data: %+v\n", data)
}

func TestGoEdges(t *testing.T) {
	ids := []string{
		"-8677519361643378587",
		"-8591648351703748711",
		"-6298695566440239494",
	}

	nebula, err := NewNebulaReadWriter("192.168.31.147", 9669, "root", "nebula", "relation_graph", []int{1, 2})
	assert.NoError(t, err)
	defer nebula.Close()

	data, err := nebula.GetAllNeighborEdges(ids)
	assert.NoError(t, err)

	fmt.Printf("data: %+v\n", data)
}

// func TestFetchVertexProps(t *testing.T) {
// 	ids := []string{
// 		"-8677519361643378587",
// 		"-8591648351703748711",
// 		"-6298695566440239494",
// 	}

// 	nebula, err := NewNebulaReadWriter("192.168.31.147", 9669, "root", "nebula", "relation_graph", []int{1, 2})
// 	assert.NoError(t, err)
// 	defer nebula.Close()

// 	vertices, err := nebula.FetchVertexProps("id", ids)
// 	assert.NoError(t, err)

// 	t.Logf("data: %+v\n", vertices)
// }

// func TestFetchEdgeProps(t *testing.T) {
// 	edges := [][2]string{
// 		[2]string{
// 			"-3724521808495006073",
// 			"1739437233305531149",
// 		},
// 		[2]string{
// 			"7134922016542360759",
// 			"-3788099225632984055",
// 		},
// 		[2]string{
// 			"-3791265973878437116",
// 			"-3788099225632984055",
// 		},
// 	}

// 	nebula, err := NewNebulaReadWriter("192.168.31.147", 9669, "root", "nebula", "relation_graph", []int{1, 2})
// 	assert.NoError(t, err)
// 	defer nebula.Close()

// 	ret, err := nebula.FetchEdgeProps("id_email", edges)
// 	assert.NoError(t, err)

// 	t.Logf("data: %+v\n", ret)
// }
