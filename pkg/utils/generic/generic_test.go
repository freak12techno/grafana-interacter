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

func TestPaginate(t *testing.T) {
	t.Parallel()

	inputs := []struct {
		name        string
		input       []int
		page        int
		perPage     int
		result      []int
		resultPages int
	}{
		{"TakeAll", []int{1, 2, 3}, 0, 3, []int{1, 2, 3}, 1},
		{"FirstPage", []int{1, 2, 3, 4, 5, 6}, 0, 3, []int{1, 2, 3}, 2},
		{"LastPage", []int{1, 2, 3, 4, 5, 6}, 1, 3, []int{4, 5, 6}, 2},
		{"Overflow1", []int{1, 2, 3, 4}, 1, 3, []int{4}, 2},
		{"Overflow2", []int{1, 2, 3, 4}, 2, 3, []int{}, 2},
	}

	for _, input := range inputs {
		t.Run(input.name, func(t *testing.T) {
			t.Parallel()

			result, pages := Paginate(input.input, input.page, input.perPage)
			require.Equal(t, input.result, result)
			require.Equal(t, input.resultPages, pages)
		})
	}
}
