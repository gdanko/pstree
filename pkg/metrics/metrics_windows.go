//go:build windows
// +build windows

package metrics

import (
	"errors"

	"github.com/shirou/gopsutil/v4/process"
)

type PGIDFunc func(proc *process.Process) (int, error)

func getPGIDFunc() PGIDFunc {
	return func(proc *process.Process) (int, error) {
		return 0, errors.New("getpgid not supported on Windows")
	}
}
