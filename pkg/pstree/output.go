package pstree

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/gdanko/pstree/util"
	"github.com/shirou/gopsutil/v4/process"
)

func GenerateTree(parent int32, tree map[int32][]int32, currentSymbol string, currentIndent, indent string, arguments bool, wide bool, showPids bool, useAscii bool) {
	var (
		cmdArgs       []string
		cmdArgsJoined string = ""
		cmdName       string
		err           error
		line          string
		lineLength    int
		ok            bool
		pid           int32
		pipe          string
		pgid          int
		symbol        string
		username      string
	)
	lineLength = util.GetLineLength()
	lineLength = lineLength - len(indent) - 6 // Accomodate the indentation, etc

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

		if len(cmdArgs) > 0 {
			if arguments {
				cmdArgsJoined = strings.Join(cmdArgs, " ")
			}
		}

		fmt.Fprintf(os.Stdout, currentSymbol, currentIndent)
		line = fmt.Sprintf("%s %s %s", username, cmdName, cmdArgsJoined)
		if showPids {
			line = line + fmt.Sprintf("(%d)", parent)
		}
		if wide {
			fmt.Fprintln(os.Stdout, line)
		} else {
			fmt.Fprintln(os.Stdout, util.TruncateEllipsis(line, lineLength))
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
		GenerateTree(child, tree, symbol, indent, indent+pipe, arguments, wide, showPids, useAscii)
	}
	child := returnLastElement(tree[parent])
	if useAscii {
		symbol = "%s  `-- "
	} else {
		symbol = "%s  └── "
	}
	GenerateTree(child, tree, symbol, indent, indent+"    ", arguments, wide, showPids, useAscii)
}
