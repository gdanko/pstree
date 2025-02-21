package pstree

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/gdanko/pstree/util"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

type Process struct {
	Age           int64
	Args          []string
	Child         int
	Children      *[]Process
	Command       string
	Connections   []net.ConnectionStat
	CPUPercent    float64
	CPUTimes      *cpu.TimesStat
	CreateTime    int64
	Environment   []string
	GIDs          []uint32
	Groups        []uint32
	MemoryInfo    *process.MemoryInfoStat
	MemoryPercent float32
	NumFDs        int32
	NumThreads    int32
	OpenFiles     []process.OpenFilesStat
	Parent        int
	ParentProcess *Process
	PGID          int32
	PID           int32
	PPID          int32
	Print         bool
	Sister        int
	Status        []string
	UIDs          []uint32
	Username      string
}

func SortByPid(procs []*process.Process) []*process.Process {
	sort.Slice(procs, func(i, j int) bool {
		return procs[i].Pid < procs[j].Pid // Ascending order
	})
	return procs
}

func GetPidFromIndex(processes *[]Process, index int) (pid int32) {
	for i := range *processes {
		if i == index {
			return (*processes)[i].PID
		}
	}
	return int32(-1)
}

func FindPrintable(processes *[]Process) (printable []Process) {
	for i := range *processes {
		if (*processes)[i].Print {
			printable = append(printable, (*processes)[i])
		}
	}
	return printable
}

func GetProcessByPid(processes *[]Process, pid int32) (proc Process, err error) {
	for i := range *processes {
		if (*processes)[i].PID == pid {
			return (*processes)[i], nil
		}
	}
	errorMessage := fmt.Sprintf("the process with the PID %d was not found", pid)
	return Process{}, errors.New(errorMessage)
}

func SortProcsByAge(processes *[]Process) {
	sort.Slice(*processes, func(i, j int) bool {
		return (*processes)[i].Age < (*processes)[j].Age
	})
}

func SortProcsByCpu(processes *[]Process) {
	sort.Slice(*processes, func(i, j int) bool {
		return (*processes)[i].CPUPercent < (*processes)[j].CPUPercent
	})
}

func SortProcsByMemory(processes *[]Process) {
	sort.Slice(*processes, func(i, j int) bool {
		return float64((*processes)[i].MemoryInfo.RSS) < float64((*processes)[j].MemoryInfo.RSS)
	})
}

func SortProcsByUsername(processes *[]Process) {
	sort.Slice(*processes, func(i, j int) bool {
		return (*processes)[i].Username < (*processes)[j].Username
	})
}

func SortProcsByPid(processes *[]Process) {
	sort.Slice(*processes, func(i, j int) bool {
		return (*processes)[i].PID < (*processes)[j].PID
	})
}

func SortProcsByNumThreads(processes *[]Process) {
	sort.Slice(*processes, func(i, j int) bool {
		return (*processes)[i].NumThreads < (*processes)[j].NumThreads
	})
}

func GetPIDIndex(logger *slog.Logger, processes []Process, pid int32) int {
	for i := range processes {
		if processes[i].PID == pid {
			return i
		}
	}
	return -1
}

func generateProcess(proc *process.Process) Process {
	var (
		args          []string
		command       string
		cpuPercent    float64
		cpuTimes      *cpu.TimesStat
		createTime    int64
		environment   []string
		err           error
		gids          []uint32
		groups        []uint32
		pgid          int
		pid           int32
		ppid          int32
		memoryInfo    *process.MemoryInfoStat
		memoryPercent float32
		numFDs        int32
		numThreads    int32
		openFiles     []process.OpenFilesStat
		uids          []uint32
		username      string
	)

	pid = proc.Pid
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	argsChannel := make(chan func(ctx context.Context, proc *process.Process) ([]string, error))
	go ProcessArgs(argsChannel)
	argsOut, err := (<-argsChannel)(ctx, proc)
	if err != nil {
		args = []string{}
	} else {
		args = argsOut
	}

	commandNameChannel := make(chan func(ctx context.Context, proc *process.Process) (string, error))
	go ProcessCommandName(commandNameChannel)
	commandOut, err := (<-commandNameChannel)(ctx, proc)
	if err != nil {
		command = "?"
	} else {
		command = commandOut
	}

	// Expensive
	// connectionsChannel := make(chan func(ctx context.Context, proc *process.Process) ([]net.ConnectionStat, error))
	// go ProcessConnections(connectionsChannel)
	// connectionsOut, err := (<-connectionsChannel)(ctx, proc)
	// if err != nil {
	// 	connections = []net.ConnectionStat{}
	// } else {
	// 	connections = connectionsOut
	// }

	cpuPercentChannel := make(chan func(ctx context.Context, proc *process.Process) (float64, error))
	go ProcessCpuPercent(cpuPercentChannel)
	cpuPercentOut, err := (<-cpuPercentChannel)(ctx, proc)
	if err != nil {
		cpuPercent = -1
	} else {
		cpuPercent = cpuPercentOut
	}

	cpuTimesChannel := make(chan func(ctx context.Context, proc *process.Process) (*cpu.TimesStat, error))
	go ProcessCpuTimes(cpuTimesChannel)
	cpuTimesOut, err := (<-cpuTimesChannel)(ctx, proc)
	if err != nil {
		cpuTimes = &cpu.TimesStat{}
	} else {
		cpuTimes = cpuTimesOut
	}

	createTimeChannel := make(chan func(ctx context.Context, proc *process.Process) (int64, error))
	go ProcessCreateTime(createTimeChannel)
	createTimeOut, err := (<-createTimeChannel)(ctx, proc)
	if err != nil {
		createTime = -1
	} else {
		createTime = createTimeOut
	}

	environmentChannel := make(chan func(ctx context.Context, proc *process.Process) ([]string, error))
	go ProcessEnvironment(environmentChannel)
	environmentOut, err := (<-environmentChannel)(ctx, proc)
	if err != nil {
		environment = []string{}
	} else {
		environment = environmentOut
	}

	gidsChannel := make(chan func(ctx context.Context, proc *process.Process) ([]uint32, error))
	go ProcessGIDs(gidsChannel)
	gidsOut, err := (<-gidsChannel)(ctx, proc)
	if err != nil {
		gids = []uint32{}
	} else {
		gids = gidsOut
	}

	groupsChannel := make(chan func(ctx context.Context, proc *process.Process) ([]uint32, error))
	go ProcessGroups(groupsChannel)
	groupsOut, err := (<-groupsChannel)(ctx, proc)
	if err != nil {
		groups = []uint32{}
	} else {
		groups = groupsOut
	}

	memoryInfoChannel := make(chan func(ctx context.Context, proc *process.Process) (*process.MemoryInfoStat, error))
	go ProcessMemoryInfo(memoryInfoChannel)
	memoryInfoOut, err := (<-memoryInfoChannel)(ctx, proc)
	if err != nil {
		memoryInfo = &process.MemoryInfoStat{}
	} else {
		memoryInfo = memoryInfoOut
	}

	memoryPercentChannel := make(chan func(ctx context.Context, proc *process.Process) (float32, error))
	go ProcessMemoryPercent(memoryPercentChannel)
	memoryPercentOut, err := (<-memoryPercentChannel)(ctx, proc)
	if err != nil {
		memoryPercent = -1.0
	} else {
		memoryPercent = memoryPercentOut
	}

	numFDsChannel := make(chan func(ctx context.Context, proc *process.Process) (int32, error))
	go ProcessNumFDs(numFDsChannel)
	numFDsOut, err := (<-numFDsChannel)(ctx, proc)
	if err != nil {
		numFDs = -1
	} else {
		numFDs = numFDsOut
	}

	openFilesChannel := make(chan func(ctx context.Context, proc *process.Process) ([]process.OpenFilesStat, error))
	go ProcessOpenFiles(openFilesChannel)
	openFilesOut, err := (<-openFilesChannel)(ctx, proc)
	if err != nil {
		openFiles = []process.OpenFilesStat{}
	} else {
		openFiles = openFilesOut
	}

	numThreadsChannel := make(chan func(ctx context.Context, proc *process.Process) (int32, error))
	go ProcessNumThreads(numThreadsChannel)
	numThreadsOut, err := (<-numThreadsChannel)(ctx, proc)
	if err != nil {
		numThreads = -1
	} else {
		numThreads = numThreadsOut
	}

	pgidChannel := make(chan func(proc *process.Process) (int, error))
	go ProcessPGID(pgidChannel)
	pgidOut, err := (<-pgidChannel)(proc)
	if err != nil {
		pgid = -1
	} else {
		pgid = pgidOut
	}

	// Expensive
	// parentChannel := make(chan func(ctx context.Context, proc *process.Process) (*process.Process, error))
	// go ProcessParent(parentChannel)
	// parentOut, err := (<-parentChannel)(ctx, proc)
	// if err != nil {
	// 	parentProcess = Process{}
	// } else {
	// 	parentProcess = generateProcess(parentOut)
	// }

	ppidChannel := make(chan func(ctx context.Context, proc *process.Process) (int32, error))
	go ProcessPPID(ppidChannel)
	ppidOut, err := (<-ppidChannel)(ctx, proc)
	if err != nil {
		ppid = -1
	} else {
		ppid = ppidOut
	}

	// Expensive
	// statusChannel := make(chan func(ctx context.Context, proc *process.Process) ([]string, error))
	// go ProcessStatus(statusChannel)
	// statusOut, err := (<-statusChannel)(ctx, proc)
	// if err != nil {
	// 	status = []string{}
	// } else {
	// 	status = statusOut
	// }

	usernameChannel := make(chan func(ctx context.Context, proc *process.Process) (string, error))
	go ProcessUsername(usernameChannel)
	usernameOut, err := (<-usernameChannel)(ctx, proc)
	if err != nil {
		username = "?"
	} else {
		username = usernameOut
	}

	uidsChannel := make(chan func(ctx context.Context, proc *process.Process) ([]uint32, error))
	go ProcessGIDs(uidsChannel)
	uidsOut, err := (<-uidsChannel)(ctx, proc)
	if err != nil {
		uids = []uint32{}
	} else {
		uids = uidsOut
	}

	if len(args) > 0 {
		if args[0] == command {
			if len(args) == 1 {
				args = []string{}
			} else if len(args) > 1 {
				args = args[1:]
			}
		}
	}

	return Process{
		Age:           util.GetUnixTimestamp() - createTime,
		Args:          args,
		Child:         -1,
		Children:      &[]Process{},
		Command:       command,
		Connections:   []net.ConnectionStat{},
		CPUPercent:    util.RoundFloat(cpuPercent, 2),
		CPUTimes:      cpuTimes,
		CreateTime:    createTime,
		Environment:   environment,
		GIDs:          gids,
		Groups:        groups,
		MemoryInfo:    memoryInfo,
		MemoryPercent: memoryPercent,
		NumFDs:        numFDs,
		NumThreads:    numThreads,
		OpenFiles:     openFiles,
		Parent:        -1,
		PGID:          int32(pgid),
		PID:           pid,
		PPID:          ppid,
		Print:         false,
		Sister:        -1,
		UIDs:          uids,
		Username:      username,
	}
}

func markParents(logger *slog.Logger, processes *[]Process, me int) {
	logger.Debug(fmt.Sprintf("Entering markParents with with me=%d", GetPidFromIndex(processes, me)))
	parent := (*processes)[me].Parent
	logger.Debug(fmt.Sprintf("Marking %d as a parent of %d", GetPidFromIndex(processes, parent), GetPidFromIndex(processes, me)))
	for parent != -1 {
		logger.Debug(fmt.Sprintf("Marking pid %d's Print attribute as true", GetPidFromIndex(processes, parent)))
		(*processes)[parent].Print = true
		parent = (*processes)[parent].Parent
	}
}

func markChildren(logger *slog.Logger, processes *[]Process, me int) {
	logger.Debug(fmt.Sprintf("Entering markChildren with with me=%d", GetPidFromIndex(processes, me)))
	var child int
	logger.Debug(fmt.Sprintf("Marking pid %d's Print attribute as true", GetPidFromIndex(processes, me)))
	(*processes)[me].Print = true
	child = (*processes)[me].Child
	for child != -1 {
		markChildren(logger, processes, child)
		child = (*processes)[child].Sister
	}
}

func GetProcesses(logger *slog.Logger, processes *[]Process) {
	var (
		err      error
		sorted   []*process.Process
		unsorted []*process.Process
	)
	unsorted, err = process.Processes()
	if err != nil {
		log.Fatalf("Failed to get processes: %v", err)
	}

	sorted = SortByPid(unsorted)

	for _, p := range sorted {
		*processes = append(*processes, generateProcess(p))
	}
}

func MakeTree(logger *slog.Logger, processes *[]Process) {
	logger.Debug("Entering MakeTree")
	for me := range *processes {
		parent := GetPIDIndex(logger, *processes, (*processes)[me].PPID)
		if parent != me && parent != -1 { // Ensure it's not self-referential
			(*processes)[me].Parent = parent
			if (*processes)[parent].Child == -1 {
				(*processes)[parent].Child = me
			} else {
				sister := (*processes)[parent].Child
				for (*processes)[sister].Sister != -1 {
					sister = (*processes)[sister].Sister
				}
				(*processes)[sister].Sister = me
			}
		}
	}
}

func MarkProcs(logger *slog.Logger, processes *[]Process, flagContains string, flagUsername []string, flagExcludeRoot bool, flagPid int32) {
	logger.Debug("Entering MakeProcs")
	var (
		me      int
		myPid   int32
		showAll bool = false
	)

	if flagContains == "" && len(flagUsername) == 0 && !flagExcludeRoot && flagPid < 1 {
		showAll = true
	}
	for me = range *processes {
		if showAll {
			(*processes)[me].Print = true
		} else {
			if len(flagUsername) > 0 {
				for _, username := range flagUsername {
					if (*processes)[me].Username == username {
						markParents(logger, processes, me)
						markChildren(logger, processes, me)
					}
				}
			} else if (*processes)[me].PID == flagPid {
				logger.Debug("flagPid == process.PID")
				if (flagExcludeRoot && (*processes)[me].Username != "root") || (!flagExcludeRoot) {
					logger.Debug("(flagExcludeRoot && process.Username != root) || !flagExcludeRoot")
					markParents(logger, processes, me)
					markChildren(logger, processes, me)
				}
			} else if flagContains != "" && strings.Contains((*processes)[me].Command, flagContains) && ((*processes)[me].PID != myPid) {
				logger.Debug("flagContains is set && process.Command contains flagContains && process.PID != myPid")
				if (flagExcludeRoot && (*processes)[me].Username != "root") || (!flagExcludeRoot) {
					logger.Debug("(flagExcludeRoot && process.Username != root) || !flagExcludeRoot")
					markParents(logger, processes, me)
					markChildren(logger, processes, me)
				}
			} else if flagContains != "" && !strings.Contains((*processes)[me].Command, flagContains) && ((*processes)[me].PID != myPid) {
				logger.Debug("flagContains is set && process.Command does not contain flagContains && process.PID != myPid")
			} else if flagExcludeRoot && (*processes)[me].Username != "root" {
				logger.Debug("flagExcludeRoot && process.Username != root")
				markParents(logger, processes, me)
				markChildren(logger, processes, me)
			}
			// }
			// if (*processes)[me].Username == flagUsername ||
			// 	flagExcludeRoot && (*processes)[me].Username != "root" ||
			// 	(*processes)[me].PID == flagPid ||
			// 	(flagContains != "" && strings.Contains((*processes)[me].Command, flagContains) && ((*processes)[me].PID != myPid)) {
			// 	markParents(logger, flagExcludeRoot, processes, me)
			// 	markChildren(logger, flagExcludeRoot, processes, me)
			// }
		}
	}
}

func DropProcs(logger *slog.Logger, processes *[]Process) {
	logger.Debug("Entering DropProcs")
	for me := range *processes {
		if (*processes)[me].Print {
			var child, sister int
			// Drop children that won't print
			child = (*processes)[me].Child
			for child != -1 && !(*processes)[child].Print {
				child = (*processes)[child].Sister
			}
			(*processes)[me].Child = child

			// Drop sisters that won't print
			sister = (*processes)[me].Sister
			for sister != -1 && !(*processes)[sister].Print {
				sister = (*processes)[sister].Sister
			}
			(*processes)[me].Sister = sister
		}
	}
}
