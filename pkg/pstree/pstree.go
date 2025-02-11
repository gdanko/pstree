package pstree

import (
	"bytes"
	"errors"
	"log"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	"github.com/gdanko/pstree/util"

	"github.com/shirou/gopsutil/v4/process"
)

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

func GetTreeData(username string, contains string, level int, excludeRoot bool) (tree map[int32][]int32) {
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

func GetTreeDataFromPs(username string, contains string, level int) (tree map[int32][]int32, err error) {
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

func GetTreeDataFromFile(filename string, username string, contains string, level int) (tree map[int32][]int32, err error) {
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
