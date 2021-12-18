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

	nebula, err := NewNebulaReadWriter("192.168.31.147", 9669, "root", "nebula", "relation_graph")
	assert.NoError(t, err)
	defer nebula.Close()

	data, err := nebula.GetSingleNeighbor("121904329086390421", "id_email", "email")
	assert.NoError(t, err)
	t.Logf("data: %+v", data)
	
	assert.Equal(t, target, data)
}

func TestMultiQuery(t *testing.T) {
	target := make(map[string][]string)
	target["email"] = []string{"efkhb@gmail.com", "uYAcuqoX@gmail.com",
							   "CDWc@gmail.com", "eDJWvJZwUk@gmail.com",
							   "ieBaArcxUR@gmail.com", "uLi@gmail.com"}
	
	target["province"] = []string{"Shandong", "Hainan", "Beijing", "Shanghai",
								  "Liaoning", "Chongqing"}

	target["telephone"] = []string{"10927866150", "14326585168", "18595097300",
								   "11660492056", "13992190315"}

	nebula, err := NewNebulaReadWriter("192.168.31.147", 9669, "root", "nebula", "relation_graph")
	assert.NoError(t, err)
	defer nebula.Close()

	edgeProp := map[string]string{
		"id_email": "email",
		"id_telephone": "telephone",
		"id_province": "province",
	}

	data, err := nebula.GetMultiNeighbors("121904329086390421", edgeProp)
	assert.NoError(t, err)
	t.Logf("data: %+v", data)

	assert.Equal(t, target, data)
}

func TestLookup(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Shanghai")
	assert.NoError(t, err)

	startTime := time.Date(2021, 12, 17, 15, 41, 0, 0, loc)
	endTime := time.Date(2021, 12, 17, 15, 42, 0, 0, loc)
	timeRange := [2]time.Time{startTime, endTime}

	node := PrincipleNode{
		Name: "id",
		Prop: "id",
		Time: "collect_time",	
	}

	nebula, err := NewNebulaReadWriter("192.168.31.147", 9669, "root", "nebula", "relation_graph")
	assert.NoError(t, err)
	defer nebula.Close()

	data, err := nebula.LookupWithTimeLimit(&node, timeRange)
	assert.NoError(t, err)
	fmt.Printf("data: %+v\n", data)
}
