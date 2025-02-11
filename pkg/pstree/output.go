package pstree

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/shirou/gopsutil/v4/process"
)

func TruncateString(s string, length int) string {
	if len(s) > length {
		return s[:length]
	}
	return s
}

func GenerateTree(parent int32, tree map[int32][]int32, currentSymbol string, currentIndent, indent string, arguments bool, wide bool, showPids bool, useAscii bool, lineLength int) {
	var (
		cmdArgs           []string
		cmdArgsJoined     string = ""
		cmdName           string
		currentLineLength int
		err               error
		line              string
		ok                bool
		pid               int32
		pipe              string
		pgid              int
		symbol            string
		username          string
	)
	currentLineLength = lineLength - (len(currentSymbol) + 6)

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
			fmt.Fprintf(os.Stdout, currentSymbol, currentIndent)
		} else {
			fmt.Fprint(os.Stdout, currentIndent)
		}

		if showPids {
			line = fmt.Sprintf("%d %s %s %s", parent, username, cmdName, cmdArgsJoined)
		} else {
			line = fmt.Sprintf("%s %s %s", username, cmdName, cmdArgsJoined)
		}

		if wide {
			fmt.Fprintln(os.Stdout, line)
		} else {
			fmt.Fprintln(os.Stdout, TruncateString(line, currentLineLength+9))
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
		GenerateTree(child, tree, symbol, indent, indent+pipe, arguments, wide, showPids, useAscii, lineLength)
	}
	child := returnLastElement(tree[parent])
	if useAscii {
		symbol = "%s  `-- "
	} else {
		symbol = "%s  └── "
	}
	GenerateTree(child, tree, symbol, indent, indent+"    ", arguments, wide, showPids, useAscii, lineLength)
}
