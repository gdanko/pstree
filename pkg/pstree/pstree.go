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
	Age                 int64
	Args                []string
	Child               int
	Children            *[]Process
	Command             string
	Connections         []net.ConnectionStat
	CPUPercent          float64
	CPUTimes            *cpu.TimesStat
	CreateTime          int64
	Environment         []string
	GIDs                []uint32
	Groups              []uint32
	HasUIDTransition    bool
	IsCurrentOrAncestor bool
	MemoryInfo          *process.MemoryInfoStat
	MemoryPercent       float32
	NumFDs              int32
	NumThreads          int32
	OpenFiles           []process.OpenFilesStat
	Parent              int
	ParentProcess       *Process
	ParentUID           uint32
	ParentUsername      string
	PGID                int32
	PID                 int32
	PPID                int32
	Print               bool
	Sister              int
	Status              []string
	UIDs                []uint32
	Username            string
}

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

// GetPidFromIndex retrieves the PID of a process at the specified index in the processes slice.
//
// Parameters:
//   - processes: Pointer to a slice of Process structs
//   - index: The index of the process in the slice
//
// Returns:
//   - The PID of the process at the specified index, or -1 if the index is out of bounds
func GetPidFromIndex(processes *[]Process, index int) (pid int32) {
	for i := range *processes {
		if i == index {
			return (*processes)[i].PID
		}
	}
	return int32(-1)
}

// FindPrintable returns a slice containing only the processes that have their Print flag set to true.
//
// Parameters:
//   - processes: Pointer to a slice of Process structs
//
// Returns:
//   - A new slice containing only the processes marked as printable
func FindPrintable(processes *[]Process) (printable []Process) {
	for i := range *processes {
		if (*processes)[i].Print {
			printable = append(printable, (*processes)[i])
		}
	}
	return printable
}

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

// GetPIDIndex finds the index of a process with the specified PID in the processes slice.
//
// Parameters:
//   - logger: Logger instance for debug information
//   - processes: Slice of Process structs to search
//   - pid: The PID to search for
//
// Returns:
//   - The index of the process with the specified PID, or -1 if not found
func GetPIDIndex(logger *slog.Logger, processes []Process, pid int32) int {
	for i := range processes {
		if processes[i].PID == pid {
			return i
		}
	}
	return -1
}

// MarkCurrentAndAncestors marks the current process and all its ancestors.
// This function identifies the current process by its PID and marks it and all
// its ancestors with IsCurrentOrAncestor=true for highlighting in the display.
//
// Parameters:
//   - logger: Logger instance for debug information
//   - processes: Pointer to a slice of Process structs
//   - currentPid: The PID of the current process to highlight
func MarkCurrentAndAncestors(logger *slog.Logger, processes *[]Process, currentPid int32) {
	if currentPid <= 0 {
		return
	}

	logger.Debug(fmt.Sprintf("Marking current process %d and its ancestors", currentPid))

	// Find the current process index
	currentIndex := GetPIDIndex(logger, *processes, currentPid)
	if currentIndex == -1 {
		logger.Debug(fmt.Sprintf("Current process %d not found", currentPid))
		return
	}

	// Mark the current process
	(*processes)[currentIndex].IsCurrentOrAncestor = true

	// Mark all ancestors
	parent := (*processes)[currentIndex].Parent
	for parent != -1 {
		logger.Debug(fmt.Sprintf("Marking pid %d as ancestor of current process", GetPidFromIndex(processes, parent)))
		(*processes)[parent].IsCurrentOrAncestor = true
		parent = (*processes)[parent].Parent
	}
}

// MarkUIDTransitions identifies and marks processes where the user ID changes from the parent process.
// This function compares the UIDs of each process with its parent and sets HasUIDTransition=true
// when a transition is detected. It also stores the parent UID for display purposes.
//
// Parameters:
//   - logger: Logger instance for debug information
//   - processes: Pointer to a slice of Process structs
func MarkUIDTransitions(logger *slog.Logger, processes *[]Process) {
	logger.Debug("Marking UID transitions between processes - START")

	for i := range *processes {
		// Skip the root process (which has no parent)
		if (*processes)[i].Parent == -1 {
			continue
		}

		// Get parent index
		parentIdx := (*processes)[i].Parent

		// Compare UIDs between process and its parent
		if len((*processes)[i].UIDs) > 0 && len((*processes)[parentIdx].UIDs) > 0 {
			// Store parent UID regardless of transition
			(*processes)[i].ParentUID = (*processes)[parentIdx].UIDs[0]
			(*processes)[i].ParentUsername = (*processes)[parentIdx].Username

			// Compare the first UID (effective UID)
			if (*processes)[i].UIDs[0] != (*processes)[parentIdx].UIDs[0] {
				logger.Debug(fmt.Sprintf("UID transition detected: Process %d (UID %d) has different UID from parent %d (UID %d)",
					(*processes)[i].PID, (*processes)[i].UIDs[0],
					(*processes)[parentIdx].PID, (*processes)[parentIdx].UIDs[0]))
				(*processes)[i].HasUIDTransition = true
			}
			if (*processes)[i].Username != (*processes)[parentIdx].Username {
				logger.Debug(fmt.Sprintf("Username transition detected: Process %d (%s) has different username from parent %d (%s)",
					(*processes)[i].PID, (*processes)[i].Username,
					(*processes)[parentIdx].PID, (*processes)[parentIdx].Username))
				(*processes)[i].HasUIDTransition = true
			}
		} else if (*processes)[i].Username != (*processes)[parentIdx].Username {
			// Fallback to username comparison if UIDs are not available
			logger.Debug(fmt.Sprintf("Username transition detected: Process %d (%s) has different username from parent %d (%s)",
				(*processes)[i].PID, (*processes)[i].Username,
				(*processes)[parentIdx].PID, (*processes)[parentIdx].Username))
			(*processes)[i].HasUIDTransition = true
		}
	}
}

// generateProcess creates a Process struct from a process.Process pointer.
// It collects various process attributes using goroutines and channels for concurrent execution
// to improve performance when gathering process information.
//
// Parameters:
//   - proc: Pointer to a process.Process struct from which to generate the Process
//
// Returns:
//   - A new Process struct populated with information from the input process
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
	go ProcessUIDs(uidsChannel)
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

// markParents marks all parent processes of a given process as printable.
// This function recursively traverses up the process tree, marking each parent
// process with Print=true until it reaches the root process (or a process with no parent).
//
// Parameters:
//   - logger: Logger instance for debug information
//   - processes: Pointer to a slice of Process structs
//   - me: Index of the process whose parents should be marked
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

// markChildren marks a process and all its child processes as printable.
// This function recursively traverses down the process tree, marking each child
// process with Print=true, and continues with any sibling processes.
//
// Parameters:
//   - logger: Logger instance for debug information
//   - processes: Pointer to a slice of Process structs
//   - me: Index of the process whose children should be marked
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

// GetProcesses retrieves information about all processes running on the system.
// It populates the provided processes slice with detailed information about each process,
// including their relationships, resource usage, and other attributes.
//
// Parameters:
//   - logger: Logger instance for debug information
//   - processes: Pointer to a slice of Process structs to populate
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

// MakeTree builds the process tree structure by establishing parent-child relationships.
// It organizes processes into a hierarchical tree structure based on their parent-child
// relationships, setting up the Parent, Child, and Sister fields for each process.
//
// Parameters:
//   - logger: Logger instance for debug information
//   - processes: Pointer to a slice of Process structs to organize into a tree
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

// MarkProcs marks processes that should be displayed based on filtering criteria.
// It applies various filters such as process name pattern matching, username filtering,
// root process exclusion, and PID filtering to determine which processes should be displayed.
//
// Parameters:
//   - logger: Logger instance for debug information
//   - processes: Pointer to a slice of Process structs to mark
//   - flagContains: String pattern to match in process command lines
//   - flagUsername: Slice of usernames to filter processes by
//   - flagExcludeRoot: Boolean indicating whether to exclude root processes
//   - flagPid: PID to filter processes by (0 means no filtering by PID)
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

// DropProcs removes processes that are not marked for display from the process tree.
// It modifies the process tree structure to maintain proper parent-child relationships
// while excluding processes that should not be displayed.
//
// Parameters:
//   - logger: Logger instance for debug information
//   - processes: Pointer to a slice of Process structs to filter
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
