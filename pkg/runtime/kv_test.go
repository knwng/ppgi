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

func TestHashPutSet(t *testing.T) {
	kv := NewRedisKV(redisURL, "", 0)
	key := "test"

	num := 10

	keys := make([]string, num)
	values := make([]string, num)
	for i := 0; i < num; i++ {
		keys[i] = fmt.Sprintf("mkey-%d", i)
		values[i] = fmt.Sprintf("mval-%d", i)
	}

	maps := make(map[string]string)
	for i, k := range keys {
		maps[k] = values[i]
	}

	err := kv.HashPut(key, maps)
	assert.NoError(t, err)

	ret, err := kv.HashMultiGet(key, keys)
	assert.NoError(t, err)
	// assert.Equal(t, values, ret)
	t.Logf("Result: %+v\n", ret)

	lackedKeys := []string{"mkey-1", "no-mkey-2", "mkey-3", "no-mkey-4", "mkey-5"}
	targetVals := []string{"mval-1", "mval-3", "mval-5"}
	targetIndices := []int{0, 2, 4}

	lackRet, err := kv.HashMultiGet(key, lackedKeys)
	assert.NoError(t, err)
	t.Logf("Result of lacked keys: %+v\n", lackRet)

	vals, indices := GetExistingStringAndIndex(lackRet)

	assert.Equal(t, targetVals, vals)
	assert.Equal(t, targetIndices, indices)
}
