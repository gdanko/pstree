package util

import (
	"sort"
	"strconv"
)

func StrToInt32(input string) int32 {
	num, _ := strconv.ParseInt(input, 10, 32)
	return int32(num)
}

func sortSlice(unsorted []int32) []int32 {
	sort.Slice(unsorted, func(i, j int) bool {
		return unsorted[i] < unsorted[j]
	})
	return unsorted
}

func FindFirstPid(tree map[int32][]int32) int32 {
	keys := make([]int32, 0, len(tree))
	for k := range tree {
		keys = append(keys, k)
	}
	return sortSlice(keys)[0]
}
