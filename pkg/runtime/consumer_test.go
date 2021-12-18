package runtime

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	topic 		string
	lookupURL 	string
)

func parseFlags() {
	flag.StringVar(&topic, "topic", "my-topic", "topic")
	flag.StringVar(&lookupURL, "url", "pulsar://localhost:6650", "url for pulsar service")
	flag.Parse()
}

func TestMain(m *testing.M) {
	parseFlags()
	os.Exit(m.Run())
}

func TestProducerConsumer(t *testing.T) {
	consumer, err := NewPulsarConsumer(lookupURL, topic, nil)
	assert.NoError(t, err)
	defer consumer.Close()

	producer, err := NewPulsarProducer(lookupURL, topic, nil)
	assert.NoError(t, err)
	defer producer.Close()

	for i := 0; i < 10; i++ {
		err := producer.Send([]byte(fmt.Sprintf("hello-%d", i)))
		assert.NoError(t, err)
	}

	for i := 0; i < 10; i++ {
		msg, err := consumer.Receive()
		assert.NoError(t, err)
		expectedMsg := fmt.Sprintf("hello-%d", i)
		assert.Equal(t, []byte(expectedMsg), msg)
	}
}

func TestSchema(t *testing.T) {
	schemaFile := "../../conf/pulsar_schema.json"
	schema, err := ioutil.ReadFile(schemaFile)
	assert.NoError(t, err)

	schemaStr := strings.ReplaceAll(string(schema), " ", "")
	schemaStr = strings.ReplaceAll(schemaStr, "\n", "")

	t.Logf("%+v", schemaStr)

	producer, err := NewPulsarProducer(lookupURL, topic, &schemaStr)
	assert.NoError(t, err)
	defer producer.Close()

	consumer, err := NewPulsarConsumer(lookupURL, topic, &schemaStr)
	assert.NoError(t, err)
	defer consumer.Close()

	const dataNum = 3
	data := make([][]byte, dataNum)
	for i := 0; i < dataNum; i++ {
		data[i] = []byte(fmt.Sprintf("data-%d", i))
	}

	target1 := Message{
		Algorithm: "rsa",
		Step: "data",
		Data: data,
	}

	err = producer.SendStruct(&target1)
	assert.NoError(t, err)

	msg1, err := consumer.ReceiveStruct()
	assert.NoError(t, err)
	t.Logf("Received struct: %+v", msg1)

	assert.Equal(t, target1, msg1)

	target2 := Message{
		Algorithm: "rsa",
		Step: "pubkey",
		Key: Key{
			N: []byte("this is a key"),
			E: 65535,
		},
	}

	err = producer.SendStruct(&target2)
	assert.NoError(t, err)

	msg2, err := consumer.ReceiveStruct()
	assert.NoError(t, err)
	t.Logf("Received struct: %+v", msg2)

	assert.Equal(t, target2, msg2)
}
