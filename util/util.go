package util

import (
	"math/big"
	"sort"
)

// Remove duplicate items from a slice of ints.
func UniqueInts(items []int) []int {
	set := map[int]bool{}
	for _, item := range items {
		set[item] = true
	}
	var unique []int
	for item := range set {
		unique = append(unique, item)
	}
	sort.Ints(unique)
	return unique
}

// Get the sum of a slice of ints.
func SumInts(items []int) int {
	sum := 0
	for _, item := range items {
		sum += item
	}
	return sum
}

// Get the minimum from a slice of ints.
func MinInt(items []int) int {
	min := items[0]
	for _, item := range items {
		if item < min {
			min = item
		}
	}
	return min
}

// And a float64 to a big.Float.
func AddBigFloat(x *big.Float, y float64) *big.Float {
	var newVal big.Float
	newVal.Add(x, big.NewFloat(y))
	return &newVal
}
