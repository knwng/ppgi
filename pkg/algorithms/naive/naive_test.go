package algorithms

import (
	"encoding/binary"
	"reflect"
	"testing"

	"github.com/knwng/ppgi/pkg/runtime"
	"github.com/stretchr/testify/assert"
)

func intArray2ByteArray(array []uint32) [][]byte {
	result := make([][]byte, 0, len(array))
	for _, val := range array {
		numBuf := make([]byte, 4)
		binary.LittleEndian.PutUint32(numBuf, val)
		result = append(result, numBuf)
	}
	return result
}

func byteArray2IntArray(array [][]byte) []uint32 {
	result := make([]uint32, 0, len(array))
	for _, val := range array {
		result = append(result, binary.LittleEndian.Uint32(val))
	}
	return result
}

func TestNaivePSI(t *testing.T) {

	serverData := []uint32{1, 2, 3, 4}
	clientData := []uint32{6, 5, 4, 3}

	ch := make(chan []byte)
	conn := runtime.NewChannelConn(ch)

	serverCh := make(chan [][]byte)
	go func() {
		data := intArray2ByteArray(serverData)
		result := NaivePSI(conn, SERVER, data)
		serverCh <- result
	}()

	clientCh := make(chan [][]byte)
	go func() {
		data := intArray2ByteArray(clientData)
		result := NaivePSI(conn, CLIENT, data)
		clientCh <- result
	}()

	serverResult := <-serverCh
	clientResult := <-clientCh

	serverIntersection := byteArray2IntArray(serverResult)
	clientIntersection := byteArray2IntArray(clientResult)

	assert.Equal(t, true, reflect.DeepEqual(serverIntersection, []uint32{3, 4}))
	assert.Equal(t, true, reflect.DeepEqual(clientIntersection, []uint32{4, 3}))
}
