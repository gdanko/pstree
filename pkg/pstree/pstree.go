package pstree

import (
	"log"
	"os"
	"sort"
	"strings"
	"syscall"

	"github.com/kr/pretty"

	"github.com/shirou/gopsutil/v4/process"
)

type Process struct {
	Args     string
	Child    int
	Command  string
	Parent   int
	PGID     int32
	PID      int32
	PPID     int32
	Print    bool
	Sister   int
	UID      int
	Username string
}

// func indexOf(tree []Process, pid int32) int {
// 	for i, p := range tree {
// 		if p.PID == pid {
// 			return i
// 		}
// 	}
// 	return -1
// }

// func returnLastElement(input []int32) (last int32) {
// 	return input[len(input)-1]
// }

// func FindFirstPid(tree map[int32][]int32) int32 {
// 	keys := make([]int32, 0, len(tree))
// 	for k := range tree {
// 		keys = append(keys, k)
// 	}
// 	return util.SortSlice(keys)[0]
// }

// func pruneTree(tree map[int32][]int32, maxDepth int, startingPid int32) map[int32][]int32 {
// 	if maxDepth <= 0 {
// 		// return map[int32][]int32{} // No depth means no results
// 		return tree // Return entire tree on no depth
// 	}

// 	result := make(map[int32][]int32)
// 	queue := []struct {
// 		pid   int32
// 		depth int
// 	}{{startingPid, 0}} // Start from the root process (PID 0)

// 	for len(queue) > 0 {
// 		node := queue[0]
// 		queue = queue[1:]

// 		if node.depth >= maxDepth {
// 			continue
// 		}

// 		if children, exists := tree[node.pid]; exists {
// 			result[node.pid] = children
// 			for _, child := range children {
// 				queue = append(queue, struct {
// 					pid   int32
// 					depth int
// 				}{child, node.depth + 1})
// 			}
// 		}
// 	}
// 	return result
// }

func sortByPid(procs []*process.Process) []*process.Process {
	sort.Slice(procs, func(i, j int) bool {
		return procs[i].Pid < procs[j].Pid // Ascending order
	})
	return procs
}

func GetPIDIndex(processes []Process, pid int32) int {
	for i := range processes {
		if processes[i].PID == pid {
			return i
		}
	}
	return -1
}

func getProcInfo(proc *process.Process) (username string, command string, args string) {
	var (
		argsSlice []string
		err       error
	)
	username, err = proc.Username()
	if err != nil {
		username = "?"
	}

	command, err = proc.Exe()
	if err != nil {
		command = "?"
	}

	argsSlice, err = proc.CmdlineSlice()
	if err != nil {
		args = ""
	}

	if len(argsSlice) > 1 {
		args = strings.Join(argsSlice[1:], " ")
	}

	return username, command, args
}

func GetProcesses(processes *[]Process) {
	var (
		args     string
		command  string
		err      error
		pid      int32
		pgid     int
		ppid     int32
		sorted   []*process.Process
		unsorted []*process.Process
		username string
	)

	// Get all processes
	unsorted, err = process.Processes()
	if err != nil {
		log.Fatalf("Failed to get processes: %v", err)
	}

	sorted = sortByPid(unsorted)

	// Map of parent to children
	for _, p := range sorted {
		username, command, args = getProcInfo(p)

		ppid, err = p.Ppid()
		if err != nil {
			continue
		}

		pid = p.Pid

		pgid, err = syscall.Getpgid(int(pid))
		if err != nil {
			pgid = -1
		}

		*processes = append(*processes,
			Process{
				Args:     args,
				Child:    -1,
				Command:  command,
				Parent:   -1,
				PGID:     int32(pgid),
				PID:      pid,
				PPID:     ppid,
				Print:    false,
				Sister:   -1,
				UID:      0,
				Username: username,
			})
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

func markParents(processes *[]Process, me int) {
	parent := (*processes)[me].Parent
	for parent != -1 {
		(*processes)[parent].Print = true
		parent = (*processes)[parent].Parent
	}
}

func markChildren(processes *[]Process, me int) {
	if (*processes)[me].Username == "root" {
		pretty.Println((*processes)[me])
	}
	var child int
	(*processes)[me].Print = true
	if (*processes)[me].Username == "root" {
		pretty.Println((*processes)[me])
	}
	child = (*processes)[me].Child
	for child != -1 {
		markChildren(processes, child)
		child = (*processes)[child].Sister
	}
}

func MarkProcs(processes *[]Process, flagContains string, flagUsername string, flagExcludeRoot bool, flagPid int32) {
	var (
		me      int
		myPid   int32
		showAll bool = false
	)

	myPid = int32(os.Getpid())
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
