package collections

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToArr(t *testing.T) {
	intArray := MapToArray[int, int](map[int]int{1: 1, 2: 2, 3: 3})
	sort.Ints(intArray)
	assert.Equal(t, []int{1, 2, 3}, intArray)

	stringArray := MapToArray[string, string](map[string]string{"1": "A", "2": "B", "3": "C", "4": "D"})
	sort.Strings(stringArray)
	assert.Equal(t, []string{"A", "B", "C", "D"}, stringArray)
}
