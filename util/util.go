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

// MinInt gets the minimum from a slice of ints.
func MinInt(items []int) int {
	if len(items) == 0 {
		return 0
	}
	min := items[0]
	if len(items) == 1 {
		return min
	}
	for _, item := range items[1:] {
		if item < min {
			min = item
		}
	}
	return min
}

// MaxInt gets the maximum from a slice of ints.
func MaxInt(items []int) int {
	if len(items) == 0 {
		return 0
	}
	max := items[0]
	if len(items) == 1 {
		return max
	}
	for _, item := range items[1:] {
		if item > max {
			max = item
		}
	}
	return max
}

// IntsContain check if a slice of ints contains a given int.
func IntsContain(needle int, haystack []int) bool {
	for _, hay := range haystack {
		if hay == needle {
			return true
		}
	}
	return false
}

// And a float64 to a big.Float.
func AddBigFloat(x *big.Float, y float64) *big.Float {
	var newVal big.Float
	newVal.Add(x, big.NewFloat(y))
	return &newVal
}
