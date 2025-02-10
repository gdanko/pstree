package util

import (
	"bufio"
	"os"
)

func ReadFileToSlice(path string) (lines []string, err error) {
	file, err := os.Open(path)
	if err != nil {
		return lines, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		return lines, err
	}
	return lines, nil
}
