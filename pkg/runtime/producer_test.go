package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const serverURL = "pulsar://localhost:6650"

func TestSimpleProducer(t *testing.T) {
	producer, err := NewPulsarProducer(serverURL, "my-topic")
	assert.NoError(t, err)
	defer producer.Close()

	for i := 0; i < 10; i++ {
		err := producer.Send([]byte("hello"))
		assert.NoError(t, err)
	}
}
