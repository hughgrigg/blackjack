package util

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Should be able to remove duplicates from a slice of ints.
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

// Should be able to get the sum of a slice of ints.
func TestSumInts(t *testing.T) {
	assert.Equal(t, 1, SumInts([]int{1}))
	assert.Equal(t, 2, SumInts([]int{1, 1}))
	assert.Equal(t, 6, SumInts([]int{1, 2, 3}))
	assert.Equal(t, 17, SumInts([]int{5, 5, 5, 2}))
}

// Should be able to get the maximum of a slice of ints.
func TestMaxInt(t *testing.T) {
	assert.Equal(t, 3, MaxInt([]int{1, 2, 3}))
	assert.Equal(t, -1, MaxInt([]int{-1, -2, -3}))
	assert.Equal(t, 5, MaxInt([]int{5}))
}

// Should be able to get the minimum of a slice of ints.
func TestMinInt(t *testing.T) {
	assert.Equal(t, 1, MinInt([]int{1, 2, 3}))
	assert.Equal(t, -3, MinInt([]int{-1, -2, -3}))
	assert.Equal(t, 5, MinInt([]int{5}))
}

// Should be able to add a float to a big.Float.
func TestAddBigFloat(t *testing.T) {
	bf := big.NewFloat(12.5)
	assert.Equal(t, big.NewFloat(17.75), AddBigFloat(bf, 5.25))
}
