package graph

import (
	"testing"
)

func TestQuery(t *testing.T) {
	nebula, err := NewNebulaReader("192.168.31.147", 9669, "root", "nebula", "relation_graph")
	if err != nil {
		t.Fatal(err)
	}
	defer nebula.Close()

	data, err := nebula.GetSingleNeighbor("121904329086390421", "id_email", "email")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("data: %+v", data)
}

func TestMultiQuery(t *testing.T) {
	nebula, err := NewNebulaReader("192.168.31.147", 9669, "root", "nebula", "relation_graph")
	if err != nil {
		t.Fatal(err)
	}
	defer nebula.Close()

	edgeProp := map[string]string{
		"id_email": "email",
		"id_telephone": "telephone",
		"id_province": "province",
	}

	data, err := nebula.GetMultiNeighbors("121904329086390421", edgeProp)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("data: %+v", data)
}
