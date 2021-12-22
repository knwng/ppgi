package runtime

import (
	"time"
	"strings"
	"io/ioutil"
	"crypto/sha256"
	"encoding/base64"
	"github.com/knwng/ppgi/pkg/algorithms/rsa_blind"
)

type Key struct {
	N	[]byte	`json:"n"`
	E	int		`json:"e"`
}

type Message struct {
	Algorithm 	string 				`json:"algorithm"`
	Step		rsa_blind.RSAStep	`json:"step"`
	SessionKey	string				`json:"session_key"`
	Data		[][]byte 			`json:"data"`
	Key			Key					`json:"key"`
}

func ReadSchema(filename string) (string, error) {
	schema, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}

	schemaStr := strings.ReplaceAll(string(schema), " ", "")
	schemaStr = strings.ReplaceAll(schemaStr, "\n", "")
	return schemaStr, nil
}

func generateSessionKey(algorithm string, step rsa_blind.RSAStep) string {
	current := time.Now().String()
	key := sha256.Sum256([]byte(strings.Join([]string{algorithm, string(step), current}, "-")))
	return "" + base64.StdEncoding.EncodeToString(key[:])
}
