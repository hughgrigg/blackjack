package util

import (
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
