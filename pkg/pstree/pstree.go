package pstree

import (
	"context"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/gdanko/pstree/util"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/process"
)

type Process struct {
	Args          []string
	Child         int
	Command       string
	CPUPercent    float64
	CPUTimes      *cpu.TimesStat
	GIDs          []uint32
	Groups        []uint32
	MemoryInfo    *process.MemoryInfoExStat
	MemoryPercent float32
	NumFDs        int32
	NumThreads    int32
	OpenFiles     []process.OpenFilesStat
	Parent        int
	PGID          int32
	PID           int32
	PPID          int32
	Print         bool
	Sister        int
	UIDs          []uint32
	Username      string
}

func sortByPid(procs []*process.Process) []*process.Process {
	sort.Slice(procs, func(i, j int) bool {
		return procs[i].Pid < procs[j].Pid // Ascending order
	})
	return procs
}

func generateProcess(proc *process.Process) Process {
	var (
		args          []string
		command       string
		cpuPercent    float64
		cpuTimes      *cpu.TimesStat
		err           error
		gids          []uint32
		groups        []uint32
		pgid          int
		pid           int32
		ppid          int32
		memoryInfo    *process.MemoryInfoExStat
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

	memoryInfoChannel := make(chan func(ctx context.Context, proc *process.Process) (*process.MemoryInfoExStat, error))
	go ProcessMemoryInfo(memoryInfoChannel)
	memoryInfoOut, err := (<-memoryInfoChannel)(ctx, proc)
	if err != nil {
		memoryInfo = &process.MemoryInfoExStat{}
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
	go ProcessPPID(numThreadsChannel)
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

	ppidChannel := make(chan func(ctx context.Context, proc *process.Process) (int32, error))
	go ProcessPPID(ppidChannel)
	ppidOut, err := (<-ppidChannel)(ctx, proc)
	if err != nil {
		ppid = -1
	} else {
		ppid = ppidOut
	}

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

	if len(args) > 1 {
		args = args[1:]
	}

	return Process{
		Args:          args,
		Child:         -1,
		Command:       command,
		CPUPercent:    util.RoundFloat(cpuPercent, 2),
		CPUTimes:      cpuTimes,
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

func markParents(processes *[]Process, me int) {
	parent := (*processes)[me].Parent
	for parent != -1 {
		(*processes)[parent].Print = true
		parent = (*processes)[parent].Parent
	}
}

func markChildren(processes *[]Process, me int) {
	if (*processes)[me].Username == "root" {
	}
	var child int
	(*processes)[me].Print = true
	if (*processes)[me].Username == "root" {
	}
	child = (*processes)[me].Child
	for child != -1 {
		markChildren(processes, child)
		child = (*processes)[child].Sister
	}
}

func GetPIDIndex(processes []Process, pid int32) int {
	for i := range processes {
		if processes[i].PID == pid {
			return i
		}
	}
	return -1
}

func GetProcesses(processes *[]Process) {
	var (
		err      error
		sorted   []*process.Process
		unsorted []*process.Process
	)
	unsorted, err = process.Processes()
	if err != nil {
		log.Fatalf("Failed to get processes: %v", err)
	}

	sorted = sortByPid(unsorted)

	for _, p := range sorted {
		*processes = append(*processes, generateProcess(p))
	}
}

func MakeTree(processes *[]Process) {
	for me := range *processes {
		parent := GetPIDIndex(*processes, (*processes)[me].PPID)
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

func MarkProcs(processes *[]Process, flagContains string, flagUsername string, flagExcludeRoot bool, flagPid int32) {
	var (
		me      int
		myPid   int32
		showAll bool = false
	)

	if flagContains == "" && flagUsername == "" && flagExcludeRoot == false {
		showAll = true
	}
	for me = range *processes {
		if showAll {
			(*processes)[me].Print = true
		} else {
			if (*processes)[me].Username == flagUsername ||
				flagExcludeRoot && (*processes)[me].Username != "root" ||
				(*processes)[me].PID == flagPid ||
				(flagContains != "" && strings.Contains((*processes)[me].Command, flagContains) && ((*processes)[me].PID != myPid)) {
				markParents(processes, me)
				markChildren(processes, me)
			}
		}
	}
}

func DropProcs(processes *[]Process) {
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
