package util

import (
	"math/big"
	"sort"
)

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

func SumInts(items []int) int {
	sum := 0
	for _, item := range items {
		sum += item
	}
	return sum
}

func MinInt(items []int) int {
	min := items[0]
	for _, item := range items {
		if item < min {
			min = item
		}
	}
	return min
}

func AddBigFloat(x *big.Float, y float64) *big.Float {
	var newVal big.Float
	newVal.Add(x, big.NewFloat(y))
	return &newVal
}
