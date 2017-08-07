package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUniqueInts(t *testing.T) {
	assert.Equal(
		t,
		[]int{5},
		UniqueInts([]int{5}),
	)

	assert.Equal(
		t,
		[]int{1, 2, 3},
		UniqueInts([]int{1, 2, 3}),
	)

	assert.Equal(
		t,
		[]int{1},
		UniqueInts([]int{1, 1, 1}),
	)

	assert.Equal(
		t,
		[]int{1, 2, 3},
		UniqueInts([]int{1, 1, 2, 2, 3, 3}),
	)
}

func TestSumInts(t *testing.T) {
	assert.Equal(t, 1, SumInts([]int{1}))
	assert.Equal(t, 2, SumInts([]int{1, 1}))
	assert.Equal(t, 6, SumInts([]int{1, 2, 3}))
	assert.Equal(t, 17, SumInts([]int{5, 5, 5, 2}))
}
