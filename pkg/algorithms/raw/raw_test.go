package raw

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRawIntersection(t *testing.T) {
	res := RawIntersection([]int{1, 2, 3, 4}, []int{6, 5, 4, 3})
	expected := map[int]bool{3: true, 4: true}
	assert.Equal(t, true, reflect.DeepEqual(res, expected))
}
