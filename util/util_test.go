package util

import (
	"testing"
	"github.com/stretchr/testify/assert"
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
