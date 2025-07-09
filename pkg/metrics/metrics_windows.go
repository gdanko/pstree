//go:build windows
// +build windows

package metrics

import (
	"errors"

	"github.com/shirou/gopsutil/v4/process"
)

// PGIDFunc is a function type that retrieves the process group ID for a given process.
type PGIDFunc func(proc *process.Process) (int, error)

// getPGIDFunc returns a function that attempts to get the process group ID (PGID)
// for a given process on Windows systems.
//
// Since Windows does not support process groups in the same way as Unix-like systems,
// this function always returns an error indicating that the operation is not supported.
//
// Returns:
//   - PGIDFunc: A function that returns (0, error) when called on Windows
func getPGIDFunc() PGIDFunc {
	return func(proc *process.Process) (int, error) {
		return 0, errors.New("getpgid not supported on Windows")
	}
}
