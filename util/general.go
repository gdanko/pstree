package util

import (
	"sort"
	"strconv"
	"unicode"

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

func TruncateEllipsis(text string, maxLength int) string {
	spaceBeforeLast, lastSpace := -1, -1
	iMinus1, iMinus2, iMinus3 := -1, -1, -1
	len := 0
	for i, r := range text {
		if unicode.IsSpace(r) || unicode.IsPunct(r) {
			spaceBeforeLast, lastSpace = lastSpace, i
		}
		len++
		if len > maxLength {
			if lastSpace != -1 && lastSpace <= iMinus3 {
				return text[:lastSpace] + "..."
			}
			if spaceBeforeLast != -1 && spaceBeforeLast <= iMinus3 {
				return text[:spaceBeforeLast] + "..."
			}
			return text[:iMinus3] + "..."
		}
		iMinus3, iMinus2, iMinus1 = iMinus2, iMinus1, i
	}
	return text
}
