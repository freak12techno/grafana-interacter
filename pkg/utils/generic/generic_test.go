package generic

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilter(t *testing.T) {
	t.Parallel()

	array := []int64{1, 2, 3}

	filtered := Filter(array, func(value int64) bool {
		return value == 2
	})

	assert.Len(t, filtered, 1)
	assert.Equal(t, int64(2), filtered[0])
}

func TestMap(t *testing.T) {
	t.Parallel()

	array := []int64{1, 2, 3}

	filtered := Map(array, func(value int64) int64 {
		return value * 2
	})

	assert.Len(t, filtered, 3)
	assert.Equal(t, int64(2), filtered[0])
	assert.Equal(t, int64(4), filtered[1])
	assert.Equal(t, int64(6), filtered[2])
}

func TestFind(t *testing.T) {
	t.Parallel()

	array := []int64{1, 2, 3}

	_, found := Find(array, func(value int64) bool {
		return value == 2
	})
	require.True(t, found)

	value, found2 := Find(array, func(value int64) bool {
		return value == 4
	})
	require.Nil(t, value)
	require.False(t, found2)
}

func TestMergeMaps(t *testing.T) {
	t.Parallel()

	map1 := map[string]string{"a": "1"}
	map2 := map[string]string{"b": "2"}

	resultMap := MergeMaps(map1, map2)
	require.Equal(t, map[string]string{"a": "1", "b": "2"}, resultMap)
}

func TestSplitArrayIntoChunks(t *testing.T) {
	t.Parallel()

	array := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}

	chunks1 := SplitArrayIntoChunks(array, 3)
	require.Equal(t, [][]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}, chunks1)

	chunks2 := SplitArrayIntoChunks(array, 5)
	require.Equal(t, [][]int{{1, 2, 3, 4, 5}, {6, 7, 8, 9}}, chunks2)
}
