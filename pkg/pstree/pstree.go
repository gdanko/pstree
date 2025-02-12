package pstree

import (
	"bytes"
	"errors"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"syscall"

	"github.com/gdanko/pstree/util"
	"github.com/kr/pretty"

	"github.com/shirou/gopsutil/v4/process"
)

type Process struct {
	Args     []string
	Children []Process
	Command  string
	Hide     bool
	Parent   int
	Pgid     int32
	Pid      int32
	Ppid     int32
	Print    bool
	Username string
}

func returnLastElement(input []int32) (last int32) {
	return input[len(input)-1]
}

func FindFirstPid(tree map[int32][]int32) int32 {
	keys := make([]int32, 0, len(tree))
	for k := range tree {
		keys = append(keys, k)
	}
	return util.SortSlice(keys)[0]
}

func pruneTree(tree map[int32][]int32, maxDepth int, startingPid int32) map[int32][]int32 {
	if maxDepth <= 0 {
		// return map[int32][]int32{} // No depth means no results
		return tree // Return entire tree on no depth
	}

	result := make(map[int32][]int32)
	queue := []struct {
		pid   int32
		depth int
	}{{startingPid, 0}} // Start from the root process (PID 0)

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]

		if node.depth >= maxDepth {
			continue
		}

		if children, exists := tree[node.pid]; exists {
			result[node.pid] = children
			for _, child := range children {
				queue = append(queue, struct {
					pid   int32
					depth int
				}{child, node.depth + 1})
			}
		}
	}
	return result
}

func getProcInfo(proc *process.Process) (username string, cmdName string, cmdArgs []string) {
	var (
		err error
	)
	username, err = proc.Username()
	if err != nil {
		username = "unknown user"
	}

	cmdName, err = proc.Exe()
	if err != nil {
		cmdName = "unknown command"
	}

	cmdArgs, err = proc.CmdlineSlice()
	if err != nil {
		cmdArgs = []string{}
	}

	if len(cmdArgs) > 1 {
		cmdArgs = cmdArgs[1:]
	}

	return username, cmdName, cmdArgs
}

func MakeTree(username string, contains string, level int, excludeRoot bool) (tree map[int32][]int32) {
	tree = make(map[int32][]int32)

	// Get all processes
	processes, err := process.Processes()
	if err != nil {
		log.Fatalf("Failed to get processes: %v", err)
	}

	// Map of parent to children
	for _, p := range processes {
		var addProc bool = true

		ppid, err := p.Ppid()
		if err != nil {
			continue
		}

		name, err := p.Exe()
		if err != nil {
			continue
		}

		procUser, _ := p.Username()
		if username != "" {
			if procUser != username {
				addProc = false
			}
		}

		if contains != "" {
			if !strings.Contains(name, contains) {
				addProc = false
			}
		}

		if excludeRoot {
			if procUser == "root" {
				addProc = false
			}
		}

		if addProc {
			pid := int32(p.Pid)
			tree[ppid] = append(tree[ppid], pid)
		}
	}
	// We need to set the starting PID to the first PID in the tree, which may not always be 0
	startingPid := FindFirstPid(tree)

	// Limit depth using BFS
	return pruneTree(tree, level, startingPid)
}

func AddToTree(tree *[]Process, newProc Process) {
	for i, p := range *tree {
		if newProc.Ppid == (*tree)[i].Pid {
			(*tree)[i].Children = append((*tree)[i].Children, newProc)
		} else {
			if len((*tree)[i].Children) > 0 {
				AddToTree(&p.Children, newProc)
			}
		}
	}
}

func sortByPid(procs []*process.Process) []*process.Process {
	sort.Slice(procs, func(i, j int) bool {
		return procs[i].Pid < procs[j].Pid // Ascending order
	})
	return procs
}

func MarkProcs(tree *[]Process, numProcs int, username string, contains string, excludeRoot bool) {
	username = "gdanko"
	var (
		// myPid int32
		// parent  int32
		// proc    Process
		i       int
		showAll bool = false
	)
	// myPid = int32(os.Getpid())

	if username == "" && contains == "" && !excludeRoot {
		showAll = true
	}

	for i, _ = range *tree {
		if showAll {
			(*tree)[i].Print = true
		} else {
			if username != "" {
				if (*tree)[i].Username == username {
					(*tree)[i].Print = true
				}
			}

			// if (username != "" && (*tree)[i].Username == username) || (excludeRoot && (*tree)[i].Username != "root") || (contains != "" && (strings.Contains((*tree)[i].Command, contains) && (*tree)[i].Pid != myPid)) {
			// 	indexOfProcess := indexOf(tree, (*tree)[i].Pid)
			// 	os.Exit(0)
			// 	MarkAncestors(tree, indexOfProcess)
			// }
		}
	}
}

func indexOf(tree []Process, pid int32) int {
	for i, p := range tree {
		if p.Pid == pid {
			return i
		}
	}
	return -1
}

func MakeTree2(username string, contains string, level int, excludeRoot bool) (tree []Process) {
	var (
		i   int
		err error
		// numProcs  int
		processes []*process.Process
		sorted    []*process.Process
	)
	// Get all processes
	processes, err = process.Processes()
	if err != nil {
		log.Fatalf("Failed to get processes: %v", err)
	}
	sorted = sortByPid(processes)

	for _, p := range sorted {
		pid := p.Pid
		ppid, err := p.Ppid()
		if err != nil {
			continue
		}

		pgid, err := syscall.Getpgid(int(pid))
		if err != nil {
			pgid = -1
		}

		username, cmdName, cmdArgs := getProcInfo(p)

		newProc := Process{
			Args:     cmdArgs,
			Children: []Process{},
			Command:  cmdName,
			Pgid:     int32(pgid),
			Pid:      pid,
			Ppid:     ppid,
			Print:    false,
			Username: username,
		}

		tree = append(tree, newProc)
	}

	for i, _ = range tree {
		var parent int
		parent = indexOf(tree, tree[i].Ppid)
		if parent != i && parent != -1 {
			tree[i].Parent = parent
		}
	}

	for _, p := range tree {
		if p.Pid == 64345 {
			pretty.Println(p)
			pretty.Println(tree[p.Parent])
			os.Exit(0)
		}
	}

	pretty.Println(tree)
	os.Exit(0)
	return tree
}

func MakeTreeFromPs(username string, contains string, level int) (tree map[int32][]int32, err error) {
	var (
		cmd       *exec.Cmd
		exitCode  int
		line      string
		lines     []string
		pattern   string
		stderr    bytes.Buffer
		stderrStr string
		stdout    bytes.Buffer
		stdoutStr string
	)
	tree = make(map[int32][]int32)

	// Get all processes
	cmd = exec.Command("ps", "-axwwo", "user,pid,ppid,pgid,command")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.Sys().(syscall.WaitStatus).ExitStatus()
		} else {
			exitCode = -1 // Unknown error
		}
	}
	stdoutStr = strings.TrimSuffix(stdout.String(), "\n")
	stderrStr = strings.TrimSuffix(stderr.String(), "\n")

	if exitCode != 0 {
		return tree, errors.New(stderrStr)
	}

	lines = strings.Split(stdoutStr, "\n")[1:]
	pattern = `^(\S+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(.+)$`
	re := regexp.MustCompile(pattern)
	for _, line = range lines {
		match := re.FindStringSubmatch(line)
		if match != nil {
			procUser := match[1]
			pid := util.StrToInt32(match[2])
			ppid := util.StrToInt32(match[3])
			name := match[5]

			var addProc bool = true
			if username != "" {
				if procUser != username {
					addProc = false
				}
			}

			if contains != "" {
				if !strings.Contains(name, contains) {
					addProc = false
				}
			}

			if addProc {
				tree[ppid] = append(tree[ppid], pid)
			}
		}
	}
	// We need to set the starting PID to the first PID in the tree, which may not always be 0
	startingPid := FindFirstPid(tree)

	// Limit depth using BFS
	return pruneTree(tree, level, startingPid), nil
}

func MakeTreeFromFile(filename string, username string, contains string, level int) (tree map[int32][]int32, err error) {
	var (
		line    string
		lines   []string
		pattern string
	)
	tree = make(map[int32][]int32)
	abs, err := filepath.Abs(filename)
	if err != nil {
		return tree, err
	}

	lines, err = util.ReadFileToSlice(abs)
	if err != nil {
		return tree, err
	}
	lines = lines[1:]
	pattern = `^(\S+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(.+)$`
	re := regexp.MustCompile(pattern)
	for _, line = range lines {
		match := re.FindStringSubmatch(line)
		if match != nil {
			procUser := match[1]
			pid := util.StrToInt32(match[2])
			ppid := util.StrToInt32(match[3])
			name := match[5]

			var addProc bool = true
			if username != "" {
				if procUser != username {
					addProc = false
				}
			}

			if contains != "" {
				if !strings.Contains(name, contains) {
					addProc = false
				}
			}

			if addProc {
				tree[ppid] = append(tree[ppid], pid)
			}
		}
	}
	// We need to set the starting PID to the first PID in the tree, which may not always be 0
	startingPid := FindFirstPid(tree)

	// Limit depth using BFS
	return pruneTree(tree, level, startingPid), nil
}
