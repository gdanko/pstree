package util

import (
	"sort"
	"strconv"

	terminal "github.com/wayneashleyberry/terminal-dimensions"
)

func StrToInt32(input string) int32 {
	num, _ := strconv.ParseInt(input, 10, 32)
	return int32(num)
}

func SortSlice(unsorted []int32) []int32 {
	sort.Slice(unsorted, func(i, j int) bool {
		return unsorted[i] < unsorted[j]
	})
	return unsorted
}

func GetLineLength() int {
	var (
		err    error
		length int = 132
		width  uint
	)
	width, err = terminal.Width()
	if err != nil {
		return length
	}

	return int(width)
}

func ToggleBool(input *bool) {
	if *input {
		*input = false
	} else {
		*input = true
	}
}
