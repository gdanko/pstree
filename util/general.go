package util

import (
	"bytes"
	"fmt"

	"math"
	"os/exec"
	"os/user"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/mem"
	terminal "github.com/wayneashleyberry/terminal-dimensions"
)

type Duration struct {
	Days    int64
	Hours   int64
	Minutes int64
	Seconds int64
}

func ExecutePipeline(commandStr string) (int, string, string, error) {
	commands := strings.Split(commandStr, "|")
	var cmds []*exec.Cmd

	// Trim spaces and create command slices
	for _, cmdStr := range commands {
		parts := strings.Fields(strings.TrimSpace(cmdStr))
		if len(parts) == 0 {
			continue
		}
		cmds = append(cmds, exec.Command(parts[0], parts[1:]...))
	}

	if len(cmds) == 0 {
		return -1, "", "No commands provided", fmt.Errorf("empty command pipeline")
	}

	// Set up pipes for the pipeline
	var stdoutBuf, stderrBuf bytes.Buffer
	var previousCmd *exec.Cmd

	for _, cmd := range cmds {
		cmd.Stderr = &stderrBuf // Capture stderr for each command

		if previousCmd != nil {
			// Create pipe between previous and current command
			pipeIn, err := previousCmd.StdoutPipe()
			if err != nil {
				return -1, "", "", fmt.Errorf("failed to create stdout pipe: %v", err)
			}
			cmd.Stdin = pipeIn
		}

		previousCmd = cmd // Move to the next command
	}

	// Capture output of the last command
	cmds[len(cmds)-1].Stdout = &stdoutBuf

	// Start and wait for all commands
	for _, cmd := range cmds {
		if err := cmd.Start(); err != nil {
			return -1, "", stderrBuf.String(), err
		}
	}

	// Ensure all commands complete execution
	for _, cmd := range cmds {
		if err := cmd.Wait(); err != nil {
			return -1, "", stderrBuf.String(), err
		}
	}

	// Get the exit code of the last command
	exitCode := 0
	if exitErr, ok := cmds[len(cmds)-1].ProcessState.Sys().(interface{ ExitCode() int }); ok {
		exitCode = exitErr.ExitCode()
	}

	return exitCode, strings.TrimRight(stdoutBuf.String(), "\n"), strings.TrimRight(stderrBuf.String(), "\n"), nil
}

func GetTotalMemory() (*mem.VirtualMemoryStat, error) {
	v, err := mem.VirtualMemory()
	if err != nil {
		return &mem.VirtualMemoryStat{}, err
	}
	return v, nil
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

func Contains(elems []string, v string) bool {
	for _, s := range elems {
		if v == s {
			return true
		}
	}
	return false
}

func GetScreenWidth() int {
	var (
		err   error
		width uint
	)
	width, err = terminal.Width()
	if err != nil {
		return 132
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
	returncode, stdout, _, err := ExecutePipeline("/usr/bin/tput colors")
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

func ByteConverter(num uint64) string {
	var (
		absolute float64
		suffix   string = "B"
		unit     string
	)
	absolute = math.Abs(float64(num))

	for _, unit = range []string{"", "Ki", "Mi", "Gi", "Ti", "Pi", "Ei", "Zi"} {
		if absolute < 1024.0 {
			return fmt.Sprintf("%.2f %s%s", RoundFloat(absolute, 2), unit, suffix)
		}
		absolute = absolute / 1024
	}
	return fmt.Sprintf("%.2f Yi%s", RoundFloat(absolute, 2), suffix)
}

func BtoI(b bool) int {
	if b {
		return 1
	}
	return 0
}

func StoI(s string) int {
	if s != "" {
		return 1
	}
	return 0
}

func GetUnixTimestamp() int64 {
	return time.Now().Unix()
}

func DetermineUsername() string {
	username, err := user.Current()
	if err != nil {
		return ""
	}
	return username.Username
}

func FindDuration(seconds int64) (duration Duration) {
	days := int64(seconds / 86400)
	hours := int64(((seconds - (days * 86400)) / 3600))
	minutes := int64(((seconds - days*86400 - hours*3600) / 60))
	secs := int64((seconds - (days * 86400) - (hours * 3600) - (minutes * 60)))
	return Duration{
		Days:    days,
		Hours:   hours,
		Minutes: minutes,
		Seconds: secs,
	}
}

func DeleteSliceElement(slice []string, index int) []string {
	return append(slice[:index], slice[index+1:]...)
}
