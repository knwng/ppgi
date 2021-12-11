package runtime

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const redisURL = "localhost:6379"

func TestKV(t *testing.T) {
	kv := NewRedisKV(redisURL, "", 0)

	for i := 0; i < 10; i++ {
		err := kv.Put(fmt.Sprintf("key-%d", i), fmt.Sprintf("val-%d", i))
		assert.NoError(t, err)
	}

	for i := 0; i < 10; i++ {
		val, err := kv.Get(fmt.Sprintf("key-%d", i))
		assert.NoError(t, err)
		expectedVal := fmt.Sprintf("val-%d", i)
		assert.Equal(t, expectedVal, val)
	}
}
