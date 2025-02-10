package pstree

import (
	"log"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gdanko/pstree/util"

	"github.com/shirou/gopsutil/v4/process"
)

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

func GetTreeData(username string, contains string, level int) map[int32][]int32 {
	tree := make(map[int32][]int32)

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

		if addProc {
			pid := int32(p.Pid)
			tree[ppid] = append(tree[ppid], pid)
		}
	}
	// We need to set the starting PID to the first PID in the tree, which may not always be 0
	startingPid := util.FindFirstPid(tree)

	// Limit depth using BFS
	return pruneTree(tree, level, startingPid)
}

func GetTreeDataFromFile(filename string, username string, contains string, level int) (tree map[int32][]int32, err error) {
	var (
		line    string
		pattern string
	)
	tree = make(map[int32][]int32)
	abs, err := filepath.Abs(filename)
	if err != nil {
		return tree, err
	}

	lines, err := util.ReadFileToSlice(abs)
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
	startingPid := util.FindFirstPid(tree)

	// Limit depth using BFS
	return pruneTree(tree, level, startingPid), nil
}
