package naive

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"

	"github.com/knwng/ppgi/pkg/runtime"
)

const (
	SERVER = "server"
	CLIENT = "client"
)

func NaivePSI(conn runtime.Conn, role string, elements [][]byte) [][]byte {
	if role != SERVER && role != CLIENT {
		panic(fmt.Sprintf("role must be server or client, found %v\n", role))
	}
	numElement := len(elements)
	// Hash the elements
	hashes := make([][]byte, 0, numElement)
	for _, element := range elements {
		hash := sha256.Sum256(element)
		hashes = append(hashes, hash[:])
	}

	intersectionHashSet := make(map[string]bool)
	result := make([][]byte, 0)
	if role == CLIENT {
		// 1. client sends her hash values to server
		for _, hash := range hashes {
			n, err := conn.Write(hash)
			if err != nil || n != len(hash) {
				panic("")
			}
		}
		// 2. read intersection hash
		numBuf := make([]byte, 4)
		n, err := conn.Read(numBuf)
		if err != nil || n != 4 {
			panic("")
		}
		numIntersection := int(binary.LittleEndian.Uint32(numBuf))

		for i := 0; i < numIntersection; i++ {
			hash := make([]byte, 32)
			n, err := conn.Read(hash)
			if err != nil || n != 32 {
				panic("")
			}
			intersectionHashSet[string(hash)] = true
		}
	} else {
		// 1. server receives the client hashes
		peerHashes := make([][]byte, 0, numElement)
		// TODO(zhuzilin) Here I hardcoded for sha256, may need the user
		// to provide a bit size for the hash functions in the future.
		for i := 0; i < numElement; i++ {
			hash := make([]byte, 32)
			n, err := conn.Read(hash)
			if err != nil || n != 32 {
				panic("")
			}
			peerHashes = append(peerHashes, hash)
		}

		// 2. server compare the client hash with her own hashes.
		hashSet := make(map[string]bool)
		for _, hash := range hashes {
			hashSet[string(hash)] = true
		}
		for _, hash := range peerHashes {
			if _, ok := hashSet[string(hash)]; ok {
				intersectionHashSet[string(hash)] = true
			}
		}

		// 3. send the intersection hash back to clent.
		numIntersection := len(intersectionHashSet)
		numBuf := make([]byte, 4)
		binary.LittleEndian.PutUint32(numBuf, uint32(numIntersection))
		conn.Write(numBuf)
		for hashStr, _ := range intersectionHashSet {
			hash := []byte(hashStr)
			n, err := conn.Write(hash)
			if err != nil || n != len(hash) {
				panic("")
			}
		}
	}
	// Read the origin values from the intersection hashes.
	for i := 0; i < numElement; i++ {
		if _, ok := intersectionHashSet[string(hashes[i])]; ok {
			result = append(result, elements[i])
		}
	}
	return result
}
