package pstree

import (
	"context"
	"syscall"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/process"
)

func ProcessArgs(c chan func(ctx context.Context, proc *process.Process) (args []string, err error)) {
	c <- (func(ctx context.Context, proc *process.Process) (args []string, err error) {
		args, err = proc.CmdlineSliceWithContext(ctx)
		return args, err
	})
}

func ProcessCommandName(c chan func(ctx context.Context, proc *process.Process) (command string, err error)) {
	c <- (func(ctx context.Context, proc *process.Process) (command string, err error) {
		command, err = proc.ExeWithContext(ctx)
		return command, err
	})
}

func ProcessCpuPercent(c chan func(ctx context.Context, proc *process.Process) (cpuPercent float64, err error)) {
	c <- (func(ctx context.Context, proc *process.Process) (cpuPercent float64, err error) {
		cpuPercent, err = proc.CPUPercentWithContext(ctx)
		return cpuPercent, err
	})
}

func ProcessCpuTimes(c chan func(ctx context.Context, proc *process.Process) (cpuTimes *cpu.TimesStat, err error)) {
	c <- (func(ctx context.Context, proc *process.Process) (cpuTimes *cpu.TimesStat, err error) {
		cpuTimes, err = proc.TimesWithContext(ctx)
		return cpuTimes, err
	})
}

func ProcessGIDs(c chan func(ctx context.Context, proc *process.Process) (gids []uint32, err error)) {
	c <- (func(ctx context.Context, proc *process.Process) (gids []uint32, err error) {
		gids, err = proc.GidsWithContext(ctx)
		return gids, err
	})
}

func ProcessGroups(c chan func(ctx context.Context, proc *process.Process) (groups []uint32, err error)) {
	c <- (func(ctx context.Context, proc *process.Process) (groups []uint32, err error) {
		groups, err = proc.GroupsWithContext(ctx)
		return groups, err
	})
}

func ProcessMemoryInfo(c chan func(ctx context.Context, proc *process.Process) (memoryInfo *process.MemoryInfoExStat, err error)) {
	c <- (func(ctx context.Context, proc *process.Process) (memoryInfo *process.MemoryInfoExStat, err error) {
		memoryInfo, err = proc.MemoryInfoExWithContext(ctx)
		return memoryInfo, err
	})
}

func ProcessMemoryPercent(c chan func(ctx context.Context, proc *process.Process) (memoryPercent float32, err error)) {
	c <- (func(ctx context.Context, proc *process.Process) (memoryPercent float32, err error) {
		memoryPercent, err = proc.MemoryPercentWithContext(ctx)
		return memoryPercent, err
	})
}

func ProcessPGID(c chan func(proc *process.Process) (pgid int, err error)) {
	c <- (func(proc *process.Process) (pgid int, err error) {
		pgid, err = syscall.Getpgid(int(proc.Pid))
		return pgid, err
	})
}

func ProcessPPID(c chan func(ctx context.Context, proc *process.Process) (ppid int32, err error)) {
	c <- (func(ctx context.Context, proc *process.Process) (ppid int32, err error) {
		ppid, err = proc.PpidWithContext(ctx)
		return ppid, err
	})
}

func ProcessNumFDs(c chan func(ctx context.Context, proc *process.Process) (numFDs int32, err error)) {
	c <- (func(ctx context.Context, proc *process.Process) (numFDs int32, err error) {
		numFDs, err = proc.NumFDsWithContext(ctx)
		return numFDs, err
	})
}

func ProcessOpenFiles(c chan func(ctx context.Context, proc *process.Process) (openFiles []process.OpenFilesStat, err error)) {
	c <- (func(ctx context.Context, proc *process.Process) (openFiles []process.OpenFilesStat, err error) {
		openFiles, err = proc.OpenFilesWithContext(ctx)
		return openFiles, err
	})
}

func ProcessNumThreads(c chan func(ctx context.Context, proc *process.Process) (numThreads int32, err error)) {
	c <- (func(ctx context.Context, proc *process.Process) (numThreads int32, err error) {
		numThreads, err = proc.NumThreadsWithContext(ctx)
		return numThreads, err
	})
}

func ProcessUsername(c chan func(ctx context.Context, proc *process.Process) (username string, err error)) {
	c <- (func(ctx context.Context, proc *process.Process) (username string, err error) {
		username, err = proc.UsernameWithContext(ctx)
		return username, err
	})
}

func ProcessUIDs(c chan func(ctx context.Context, proc *process.Process) (uids []uint32, err error)) {
	c <- (func(ctx context.Context, proc *process.Process) (uids []uint32, err error) {
		uids, err = proc.UidsWithContext(ctx)
		return uids, err
	})
}
