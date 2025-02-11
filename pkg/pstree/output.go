package pstree

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/gdanko/pstree/util"
	"github.com/shirou/gopsutil/v4/process"
)

func GenerateTree(parent int32, tree map[int32][]int32, currentSymbol string, currentIndent, indent string, arguments bool, wide bool, showPids bool, useAscii bool, screenWidth int) {
	var (
		cmdArgs       []string
		cmdArgsJoined string = ""
		cmdName       string
		// currentsSreenWidth int
		err        error
		line       string
		linePrefix string
		ok         bool
		pgid       int
		pid        int32
		pipe       string
		symbol     string
		username   string
	)

	proc, err := process.NewProcess(parent)
	if err == nil {
		pid = proc.Pid

		pgid, err = syscall.Getpgid(int(pid))
		if err != nil {
			pgid = -1
		}

		if currentSymbol == "%s  └── " {
			if pid == int32(pgid) {
				currentSymbol = "%s  └─= "
			}
		} else if currentSymbol == "%s  `-- " {
			if pid == int32(pgid) {
				currentSymbol = "%s  `-= "
			}
		} else if currentSymbol == "%s  ├── " {
			if pid == int32(pgid) {
				currentSymbol = "%s  ├─= "
			}
		} else if currentSymbol == "%s  |-- " {
			if pid == int32(pgid) {
				currentSymbol = "%s  |-= "
			}
		}

		username, cmdName, cmdArgs = getProcInfo(proc)
		if arguments {
			if len(cmdArgs) > 0 {
				cmdArgsJoined = strings.Join(cmdArgs, " ")
			}
		}

		if currentSymbol != "" {
			linePrefix = fmt.Sprintf(currentSymbol, currentIndent)
		} else {
			linePrefix = fmt.Sprintln(currentIndent)
		}

		if showPids {
			line = fmt.Sprintf("%s%d %s %s %s", linePrefix, parent, username, cmdName, cmdArgsJoined)
		} else {
			line = fmt.Sprintf("%s%s %s %s", linePrefix, username, cmdName, cmdArgsJoined)
		}

		if wide {
			fmt.Fprintln(os.Stdout, line)
		} else {
			if len(line) > screenWidth {
				fmt.Fprintln(os.Stdout, util.TruncateString(line, screenWidth))
			} else {
				fmt.Fprintln(os.Stdout, line)
			}
		}
	}
	_, ok = tree[parent]
	if !ok {
		return
	}

	children := tree[parent][:len(tree[parent])-1]
	for _, child := range children {
		if useAscii {
			symbol = "%s  |-- "
			pipe = "  | "
		} else {
			symbol = "%s  ├── "
			pipe = "  │ "
		}
		GenerateTree(child, tree, symbol, indent, indent+pipe, arguments, wide, showPids, useAscii, screenWidth)
	}
	child := returnLastElement(tree[parent])
	if useAscii {
		symbol = "%s  `-- "
	} else {
		symbol = "%s  └── "
	}
	GenerateTree(child, tree, symbol, indent, indent+"    ", arguments, wide, showPids, useAscii, screenWidth)
}
