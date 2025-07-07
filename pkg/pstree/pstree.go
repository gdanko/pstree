// Package pstree provides functionality for building and displaying process trees.
//
// This file contains core process handling functions including process collection,
// sorting, and data transformation. It serves as the foundation for the process tree
// visualization by gathering and organizing the raw process data.
package pstree

import (
	"context"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"sort"
	"time"

	"github.com/gdanko/pstree/pkg/metrics"
	"github.com/gdanko/pstree/util"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

//------------------------------------------------------------------------------
// PROCESS SORTING FUNCTIONS
//------------------------------------------------------------------------------
// Functions in this section handle sorting processes by various attributes.

// SortByPid sorts a slice of process.Process pointers by their PID in ascending order.
//
// Parameters:
//   - procs: Slice of process pointers to be sorted
//
// Returns:
//   - Sorted slice of process pointers
func SortByPid(procs []*process.Process) []*process.Process {
	sort.Slice(procs, func(i, j int) bool {
		return procs[i].Pid < procs[j].Pid // Ascending order
	})
	return procs
}

//------------------------------------------------------------------------------
// PROCESS LOOKUP FUNCTIONS
//------------------------------------------------------------------------------
// Functions in this section handle finding processes by specific attributes.

// GetProcessByPid finds and returns a process with the specified PID from the processes slice.
//
// Parameters:
//   - processes: Pointer to a slice of Process structs
//   - pid: The PID of the process to find
//
// Returns:
//   - The Process struct for the specified PID
//   - An error if the process with the given PID was not found
func GetProcessByPid(processes *[]Process, pid int32) (proc Process, err error) {
	for i := range *processes {
		if (*processes)[i].PID == pid {
			return (*processes)[i], nil
		}
	}
	errorMessage := fmt.Sprintf("the process with the PID %d was not found", pid)
	return Process{}, errors.New(errorMessage)
}

// SortProcsByAge sorts the processes slice by process age in ascending order.
//
// Parameters:
//   - processes: Pointer to a slice of Process structs to be sorted
func SortProcsByAge(processes *[]Process) {
	sort.Slice(*processes, func(i, j int) bool {
		return (*processes)[i].Age < (*processes)[j].Age
	})
}

// SortProcsByCpu sorts the processes slice by CPU usage percentage in ascending order.
//
// Parameters:
//   - processes: Pointer to a slice of Process structs to be sorted
func SortProcsByCpu(processes *[]Process) {
	sort.Slice(*processes, func(i, j int) bool {
		return (*processes)[i].CPUPercent < (*processes)[j].CPUPercent
	})
}

// SortProcsByMemory sorts the processes slice by memory usage (RSS) in ascending order.
//
// Parameters:
//   - processes: Pointer to a slice of Process structs to be sorted
func SortProcsByMemory(processes *[]Process) {
	sort.Slice(*processes, func(i, j int) bool {
		return float64((*processes)[i].MemoryInfo.RSS) < float64((*processes)[j].MemoryInfo.RSS)
	})
}

// SortProcsByUsername sorts the processes slice by username in ascending alphabetical order.
//
// Parameters:
//   - processes: Pointer to a slice of Process structs to be sorted
func SortProcsByUsername(processes *[]Process) {
	sort.Slice(*processes, func(i, j int) bool {
		return (*processes)[i].Username < (*processes)[j].Username
	})
}

// SortProcsByPid sorts the processes slice by PID in ascending order.
//
// Parameters:
//   - processes: Pointer to a slice of Process structs to be sorted
func SortProcsByPid(processes *[]Process) {
	sort.Slice(*processes, func(i, j int) bool {
		return (*processes)[i].PID < (*processes)[j].PID
	})
}

// SortProcsByNumThreads sorts the processes slice by the number of threads in ascending order.
//
// Parameters:
//   - processes: Pointer to a slice of Process structs to be sorted
func SortProcsByNumThreads(processes *[]Process) {
	sort.Slice(*processes, func(i, j int) bool {
		return (*processes)[i].NumThreads < (*processes)[j].NumThreads
	})
}

//------------------------------------------------------------------------------
// PROCESS DATA COLLECTION
//------------------------------------------------------------------------------
// Functions in this section handle gathering detailed process information.

// generateProcess creates a Process struct from a process.Process pointer.
// It collects various process attributes using goroutines and channels for concurrent execution
// to improve performance when gathering process information.
//
// Parameters:
//   - proc: Pointer to a process.Process struct from which to generate the Process
//
// Returns:
//   - A new Process struct populated with information from the input process
func GenerateProcess(proc *process.Process) Process {
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
		threads       map[int32]*cpu.TimesStat
		uids          []uint32
		username      string
	)

	pid = proc.Pid
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	argsChannel := make(chan func(ctx context.Context, proc *process.Process) (args []string, err error))
	go metrics.ProcessArgs(argsChannel)
	argsOut, err := (<-argsChannel)(ctx, proc)
	if err != nil {
		args = []string{}
	} else {
		args = argsOut
	}

	commandNameChannel := make(chan func(ctx context.Context, proc *process.Process) (string, error))
	go metrics.ProcessCommandName(commandNameChannel)
	commandOut, err := (<-commandNameChannel)(ctx, proc)
	if err != nil {
		command = "?"
	} else {
		command = commandOut
	}

	cpuPercentChannel := make(chan func(ctx context.Context, proc *process.Process) (cpuPercent float64, err error))
	go metrics.ProcessCpuPercent(cpuPercentChannel)
	cpuPercentOut, err := (<-cpuPercentChannel)(ctx, proc)
	if err != nil {
		cpuPercent = -1
	} else {
		cpuPercent = cpuPercentOut
	}

	cpuTimesChannel := make(chan func(ctx context.Context, proc *process.Process) (cpuTimes *cpu.TimesStat, err error))
	go metrics.ProcessCpuTimes(cpuTimesChannel)
	cpuTimesOut, err := (<-cpuTimesChannel)(ctx, proc)
	if err != nil {
		cpuTimes = &cpu.TimesStat{}
	} else {
		cpuTimes = cpuTimesOut
	}

	createTimeChannel := make(chan func(ctx context.Context, proc *process.Process) (createTime int64, err error))
	go metrics.ProcessCreateTime(createTimeChannel)
	createTimeOut, err := (<-createTimeChannel)(ctx, proc)
	if err != nil {
		createTime = -1
	} else {
		createTime = createTimeOut
	}

	environmentChannel := make(chan func(ctx context.Context, proc *process.Process) (environment []string, err error))
	go metrics.ProcessEnvironment(environmentChannel)
	environmentOut, err := (<-environmentChannel)(ctx, proc)
	if err != nil {
		environment = []string{}
	} else {
		environment = environmentOut
	}

	gidsChannel := make(chan func(ctx context.Context, proc *process.Process) (gids []uint32, err error))
	go metrics.ProcessGIDs(gidsChannel)
	gidsOut, err := (<-gidsChannel)(ctx, proc)
	if err != nil {
		gids = []uint32{}
	} else {
		gids = gidsOut
	}

	groupsChannel := make(chan func(ctx context.Context, proc *process.Process) (groups []uint32, err error))
	go metrics.ProcessGroups(groupsChannel)
	groupsOut, err := (<-groupsChannel)(ctx, proc)
	if err != nil {
		groups = []uint32{}
	} else {
		groups = groupsOut
	}

	memoryInfoChannel := make(chan func(ctx context.Context, proc *process.Process) (memoryInfo *process.MemoryInfoStat, err error))
	go metrics.ProcessMemoryInfo(memoryInfoChannel)
	memoryInfoOut, err := (<-memoryInfoChannel)(ctx, proc)
	if err != nil {
		memoryInfo = &process.MemoryInfoStat{}
	} else {
		memoryInfo = memoryInfoOut
	}

	memoryPercentChannel := make(chan func(ctx context.Context, proc *process.Process) (memoryPercent float32, err error))
	go metrics.ProcessMemoryPercent(memoryPercentChannel)
	memoryPercentOut, err := (<-memoryPercentChannel)(ctx, proc)
	if err != nil {
		memoryPercent = -1.0
	} else {
		memoryPercent = memoryPercentOut
	}

	numFDsChannel := make(chan func(ctx context.Context, proc *process.Process) (numFDs int32, err error))
	go metrics.ProcessNumFDs(numFDsChannel)
	numFDsOut, err := (<-numFDsChannel)(ctx, proc)
	if err != nil {
		numFDs = -1
	} else {
		numFDs = numFDsOut
	}

	openFilesChannel := make(chan func(ctx context.Context, proc *process.Process) ([]process.OpenFilesStat, error))
	go metrics.ProcessOpenFiles(openFilesChannel)
	openFilesOut, err := (<-openFilesChannel)(ctx, proc)
	if err != nil {
		openFiles = []process.OpenFilesStat{}
	} else {
		openFiles = openFilesOut
	}

	numThreadsChannel := make(chan func(ctx context.Context, proc *process.Process) (numThreads int32, err error))
	go metrics.ProcessNumThreads(numThreadsChannel)
	numThreadsOut, err := (<-numThreadsChannel)(ctx, proc)
	if err != nil {
		numThreads = -1
	} else {
		numThreads = numThreadsOut
	}

	pgidChannel := make(chan func(proc *process.Process) (pgid int, err error))
	go metrics.ProcessPGID(pgidChannel)
	pgidOut, err := (<-pgidChannel)(proc)
	if err != nil {
		pgid = -1
	} else {
		pgid = pgidOut
	}

	ppidChannel := make(chan func(ctx context.Context, proc *process.Process) (ppid int32, err error))
	go metrics.ProcessPPID(ppidChannel)
	ppidOut, err := (<-ppidChannel)(ctx, proc)
	if err != nil {
		ppid = -1
	} else {
		ppid = ppidOut
	}

	threadsChannel := make(chan func(ctx context.Context, proc *process.Process) (threads map[int32]*cpu.TimesStat, err error))
	go metrics.ProcessThreads(threadsChannel)
	threadsOut, err := (<-threadsChannel)(ctx, proc)
	if err != nil {
		threads = map[int32]*cpu.TimesStat{}
	} else {
		threads = threadsOut
	}

	usernameChannel := make(chan func(ctx context.Context, proc *process.Process) (username string, err error))
	go metrics.ProcessUsername(usernameChannel)
	usernameOut, err := (<-usernameChannel)(ctx, proc)
	if err != nil {
		username = "?"
	} else {
		username = usernameOut
	}

	uidsChannel := make(chan func(ctx context.Context, proc *process.Process) (uids []uint32, err error))
	go metrics.ProcessUIDs(uidsChannel)
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

	processThreads := []Thread{}
	for threadID, thread := range threads {
		if threadID != pid {
			processThreads = append(processThreads, Thread{
				Args:     args,
				Command:  filepath.Base(command),
				CPUTimes: thread,
				PID:      pid,
				PPID:     ppid,
				TID:      threadID,
			})
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
		Sister:        -1,
		Threads:       processThreads,
		UIDs:          uids,
		Username:      username,
	}
}

func GenerateThread(proc *process.Process, thread Thread) Process {
	return Process{
		TID: thread.TID,
	}
}

// GetProcesses retrieves all system processes and populates the provided processes slice.
//
// This function uses the gopsutil library to get a list of all processes running on the system,
// sorts them by PID, and then generates detailed Process structs for each one using the
// generateProcess function.
//
// Parameters:
//   - processes: A pointer to a slice that will be populated with Process structs
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

	sorted = SortByPid(unsorted)

	// stuff to simulate threads
	// rand.Seed(time.Now().UnixNano())

	for _, p := range sorted {
		newProcess := GenerateProcess(p)
		// randomThreadCount := rand.Intn(10) + 1
		// randomThreadID := rand.Intn(1000000)

		// for i := 0; i < randomThreadCount; i++ {
		// 	newThread := Thread{
		// 		Args:    newProcess.Args,
		// 		Command: newProcess.Command,
		// 		PID:     int32(newProcess.PID),
		// 		PPID:    int32(newProcess.PPID),
		// 		TID:     int32(randomThreadID),
		// 	}
		// 	newProcess.Threads = append(newProcess.Threads, newThread)
		// }

		*processes = append(*processes, newProcess)
	}
}
