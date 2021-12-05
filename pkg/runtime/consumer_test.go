package runtime

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const lookupURL = "pulsar://localhost:6650"

func TestProducerConsumer(t *testing.T) {
	topic := "my-topic"

	consumer, err := NewConsumer(lookupURL, topic)
	assert.NoError(t, err)
	defer consumer.Close()

	producer, err := NewProducer(lookupURL, topic)
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
