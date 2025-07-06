//go:build linux || darwin || freebsd
// +build linux darwin freebsd

package pstree

import (
	"syscall"

	"github.com/shirou/gopsutil/v4/process"
)

// getPGIDFunc sends a function to the provided channel that retrieves the process group ID of a process.
// This function is designed to be used with goroutines to gather process information concurrently.
// Unlike other functions, this one uses syscall.Getpgid directly instead of a context-aware method.
//
// Returns:
//   - pgid: The process group ID of the process
//   - err: Error if the process group ID could not be retrieved
func getPGIDFunc() func(proc *process.Process) (int, error) {
	return func(proc *process.Process) (int, error) {
		return syscall.Getpgid(int(proc.Pid))
	}
}
