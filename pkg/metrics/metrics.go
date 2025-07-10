package metrics

import (
	"context"
	"fmt"
	"os/user"
	"strconv"

	"github.com/gdanko/pstree/pkg/globals"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

// ProcessArgs sends a function to the provided channel that retrieves command line arguments for a process.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessArgs(c chan func(ctx context.Context, proc *process.Process) (args []string, err error)) {
	c <- (func(ctx context.Context, proc *process.Process) (args []string, err error) {
		args, err = proc.CmdlineSliceWithContext(ctx)
		return args, err
	})
}

// ProcessCommandName sends a function to the provided channel that retrieves the executable path of a process.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessCommandName(c chan func(ctx context.Context, proc *process.Process) (string, error)) {
	c <- (func(ctx context.Context, proc *process.Process) (command string, err error) {
		// First check for exe, which should be the full path to the
		exe, err := proc.ExeWithContext(ctx)
		if err == nil && exe != "" {
			// Return the full path
			if globals.GetDebugLevel() > 1 {
				globals.GetLogger().Debug(fmt.Sprintf("ProcessCommandName, PID %d (ExeWithContext): %s", proc.Pid, exe))
			}
			return exe, nil
		}

		// Either there was en error or exe was empty so let's try to get the command slice
		cmdLine, err := proc.CmdlineSliceWithContext(ctx)
		if err == nil && len(cmdLine) > 0 {
			// Return the first element of the command line slice, which is the executable
			if globals.GetDebugLevel() > 1 {
				globals.GetLogger().Debug(fmt.Sprintf("ProcessCommandName, PID %d (CmdlineSliceWithContext): %s", proc.Pid, cmdLine[0]))
			}
			return cmdLine[0], nil
		}

		// Crud, we still don't have a command name so let's try to get the command basename
		name, err := proc.NameWithContext(ctx)
		if err == nil && name != "" {
			// Return name, which is the basename of the command
			if globals.GetDebugLevel() > 1 {
				globals.GetLogger().Debug(fmt.Sprintf("ProcessCommandName, PID %d (NameWithContext): %s", proc.Pid, name))
			}
			return name, nil
		}

		// Well crap, I give up, let's return the PID
		if globals.GetDebugLevel() > 1 {
			globals.GetLogger().Debug(fmt.Sprintf("ProcessCommandName, PID %d (PID): %d", proc.Pid, proc.Pid))
		}
		return fmt.Sprintf("[PID %d]", proc.Pid), nil
	})
}

// ProcessConnections sends a function to the provided channel that retrieves network connections for a process.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessConnections(c chan func(ctx context.Context, proc *process.Process) (connections []net.ConnectionStat, err error)) {
	c <- (func(ctx context.Context, proc *process.Process) (connections []net.ConnectionStat, err error) {
		connections, err = proc.ConnectionsWithContext(ctx)
		return connections, err
	})
}

// ProcessCpuPercent sends a function to the provided channel that retrieves CPU usage percentage for a process.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessCpuPercent(c chan func(ctx context.Context, proc *process.Process) (cpuPercent float64, err error)) {
	c <- (func(ctx context.Context, proc *process.Process) (cpuPercent float64, err error) {
		cpuPercent, err = proc.CPUPercentWithContext(ctx)
		return cpuPercent, err
	})
}

// ProcessCpuTimes sends a function to the provided channel that retrieves CPU time statistics for a process.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessCpuTimes(c chan func(ctx context.Context, proc *process.Process) (cpuTimes *cpu.TimesStat, err error)) {
	c <- (func(ctx context.Context, proc *process.Process) (cpuTimes *cpu.TimesStat, err error) {
		cpuTimes, err = proc.TimesWithContext(ctx)
		return cpuTimes, err
	})
}

// ProcessCreateTime sends a function to the provided channel that retrieves the creation time of a process.
// This function is designed to be used with goroutines to gather process information concurrently.
// The creation time is converted from milliseconds to seconds before being returned.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessCreateTime(c chan func(ctx context.Context, proc *process.Process) (createTime int64, err error)) {
	c <- (func(ctx context.Context, proc *process.Process) (createTime int64, err error) {
		createTime, err = proc.CreateTimeWithContext(ctx)
		return createTime / 1000, err
	})
}

// ProcessEnvironment sends a function to the provided channel that retrieves environment variables for a process.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessEnvironment(c chan func(ctx context.Context, proc *process.Process) (environment []string, err error)) {
	c <- (func(ctx context.Context, proc *process.Process) (environment []string, err error) {
		environment, err = proc.EnvironWithContext(ctx)
		return environment, err
	})
}

// ProcessGIDs sends a function to the provided channel that retrieves group IDs for a process.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessGIDs(c chan func(ctx context.Context, proc *process.Process) (gids []uint32, groupsMap map[uint32]string, err error)) {
	c <- (func(ctx context.Context, proc *process.Process) (gids []uint32, groupsMap map[uint32]string, err error) {
		gids, err = proc.GidsWithContext(ctx)
		if err != nil {
			return []uint32{}, make(map[uint32]string), err
		}
		groupsMap = make(map[uint32]string, len(gids))
		for _, gid := range gids {
			groupName, err := user.LookupGroupId(strconv.FormatUint(uint64(gid), 10))
			if err == nil {
				groupsMap[gid] = groupName.Name
			}
		}
		return gids, groupsMap, nil
	})
}

// ProcessGroups sends a function to the provided channel that retrieves supplementary group IDs for a process.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessGroups(c chan func(ctx context.Context, proc *process.Process) (groups []uint32, err error)) {
	c <- (func(ctx context.Context, proc *process.Process) (groups []uint32, err error) {
		groups, err = proc.GroupsWithContext(ctx)
		return groups, err
	})
}

// ProcessMemoryInfo sends a function to the provided channel that retrieves memory usage statistics for a process.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessMemoryInfo(c chan func(ctx context.Context, proc *process.Process) (memoryInfo *process.MemoryInfoStat, err error)) {
	c <- (func(ctx context.Context, proc *process.Process) (memoryInfo *process.MemoryInfoStat, err error) {
		memoryInfo, err = proc.MemoryInfoWithContext(ctx)
		return memoryInfo, err
	})
}

// ProcessMemoryPercent sends a function to the provided channel that retrieves memory usage percentage for a process.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessMemoryPercent(c chan func(ctx context.Context, proc *process.Process) (memoryPercent float32, err error)) {
	c <- (func(ctx context.Context, proc *process.Process) (memoryPercent float32, err error) {
		memoryPercent, err = proc.MemoryPercentWithContext(ctx)
		return memoryPercent, err
	})
}

// ProcessParent sends a function to the provided channel that retrieves the parent process of a process.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessParent(c chan func(ctx context.Context, proc *process.Process) (parent *process.Process, err error)) {
	c <- (func(ctx context.Context, proc *process.Process) (parent *process.Process, err error) {
		parent, err = proc.ParentWithContext(ctx)
		return parent, err
	})
}

// ProcessPGID sends a function to the provided channel that retrieves the process group ID of a process.
// This function is designed to be used with goroutines to gather process information concurrently.
// Unlike other functions, this one uses syscall.Getpgid directly instead of a context-aware method.
// This functionality is not supported on Windows.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessPGID(c chan func(proc *process.Process) (int, error)) {
	c <- getPGIDFunc()
}

// ProcessPPID sends a function to the provided channel that retrieves the parent process ID of a process.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessPPID(c chan func(ctx context.Context, proc *process.Process) (ppid int32, err error)) {
	c <- (func(ctx context.Context, proc *process.Process) (ppid int32, err error) {
		ppid, err = proc.PpidWithContext(ctx)
		return ppid, err
	})
}

// ProcessNumFDs sends a function to the provided channel that retrieves the number of file descriptors used by a process.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessNumFDs(c chan func(ctx context.Context, proc *process.Process) (numFDs int32, err error)) {
	c <- (func(ctx context.Context, proc *process.Process) (numFDs int32, err error) {
		numFDs, err = proc.NumFDsWithContext(ctx)
		return numFDs, err
	})
}

// ProcessNumThreads sends a function to the provided channel that retrieves the number of threads used by a process.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessNumThreads(c chan func(ctx context.Context, proc *process.Process) (numThreads int32, err error)) {
	c <- (func(ctx context.Context, proc *process.Process) (numThreads int32, err error) {
		numThreads, err = proc.NumThreadsWithContext(ctx)
		return numThreads, err
	})
}

// ProcessOpenFiles sends a function to the provided channel that retrieves information about files opened by a process.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessOpenFiles(c chan func(ctx context.Context, proc *process.Process) (openFiles []process.OpenFilesStat, err error)) {
	c <- (func(ctx context.Context, proc *process.Process) (openFiles []process.OpenFilesStat, err error) {
		openFiles, err = proc.OpenFilesWithContext(ctx)
		return openFiles, err
	})
}

// ProcessStatus sends a function to the provided channel that retrieves the status of a process.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessStatus(c chan func(ctx context.Context, proc *process.Process) (status []string, err error)) {
	c <- (func(ctx context.Context, proc *process.Process) (status []string, err error) {
		status, err = proc.StatusWithContext(ctx)
		return status, err
	})
}

// ProcessThreads sends a function to the provided channel that retrieves the threads of a process.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessThreads(c chan func(ctx context.Context, proc *process.Process) (threads map[int32]*cpu.TimesStat, err error)) {
	c <- (func(ctx context.Context, proc *process.Process) (threads map[int32]*cpu.TimesStat, err error) {
		threads, err = proc.ThreadsWithContext(ctx)
		return threads, err
	})
}

// ProcessUsername sends a function to the provided channel that retrieves the username of the process owner.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessUsername(c chan func(ctx context.Context, proc *process.Process) (username string, err error)) {
	c <- (func(ctx context.Context, proc *process.Process) (username string, err error) {
		username, err = proc.UsernameWithContext(ctx)
		return username, err
	})
}

// ProcessUIDs sends a function to the provided channel that retrieves user IDs for a process.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessUIDs(c chan func(ctx context.Context, proc *process.Process) (uids []uint32, err error)) {
	c <- (func(ctx context.Context, proc *process.Process) (uids []uint32, err error) {
		uids, err = proc.UidsWithContext(ctx)
		return uids, err
	})
}
