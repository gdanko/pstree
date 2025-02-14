package util

import (
	"bytes"
	"math"
	"os/exec"
	"os/user"
	"sort"
	"strconv"
	"strings"

	terminal "github.com/wayneashleyberry/terminal-dimensions"
)

func ExecuteCommand(name string, args ...string) (int, string, string, error) {
	var (
		cmd      *exec.Cmd
		err      error
		exitCode int = 0
		exitErr  *exec.ExitError
		ok       bool
	)
	cmd = exec.Command(name, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()

	if err != nil {
		if exitErr, ok = err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return -1, "", "", err // Return error if not an ExitError
		}
	}
	return exitCode, strings.TrimRight(stdout.String(), "\n"), strings.TrimRight(stderr.String(), "\n"), nil
}

func StrToInt32(input string) int32 {
	num, _ := strconv.ParseInt(input, 10, 32)
	return int32(num)
}

func Int32toStr(input int32) string {
	output := strconv.Itoa(int(input))
	return output
}

func SortSlice(unsorted []int32) []int32 {
	sort.Slice(unsorted, func(i, j int) bool {
		return unsorted[i] < unsorted[j]
	})
	return unsorted
}

func GetScreenWidth() int {
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

func TruncateString(s string, length int) string {
	if len(s) > length {
		return s[:length]
	}
	return s
}

func HasColorSupport() (bool, int) {
	returncode, stdout, _, err := ExecuteCommand("/usr/bin/tput", "colors")
	if err != nil || returncode != 0 {
		return false, 0
	}
	colors, err := strconv.Atoi(stdout)
	if err != nil {
		return false, 0
	}
	if colors < 8 {
		return false, 0
	}
	return true, colors
}

func UserExists(username string) bool {
	_, err := user.Lookup(username)
	return err == nil
}

func RoundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}
