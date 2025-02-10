package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"unicode"

	flags "github.com/jessevdk/go-flags"
	"github.com/shirou/gopsutil/v4/process"
	terminal "github.com/wayneashleyberry/terminal-dimensions"
)

const VERSION = "0.1.0"

type Options struct {
	ShowArgs bool   `short:"a" long:"arguments" description:"Show command line arguments."`
	Version  func() `short:"V" long:"version" description:"Display version information."`
	Long     bool   `short:"l" long:"long" description:"Display long lines."`
	User     string `short:"u" long:"user" description:"Show only branches containing processes of <user>."`
	Start    int32  `long:"start" description:"Start at PID <start>."`
}

func findFirstPid(tree map[int32][]int32) int32 {
	keys := make([]int32, 0, len(tree))
	for k := range tree {
		keys = append(keys, k)
	}
	return sortSlice(keys)[0]
}

func returnLastElement(input []int32) (last int32) {
	return input[len(input)-1]
}

func truncateEllipsis(text string, maxLength int) string {
	spaceBeforeLast, lastSpace := -1, -1
	iMinus1, iMinus2, iMinus3 := -1, -1, -1
	len := 0
	for i, r := range text {
		if unicode.IsSpace(r) || unicode.IsPunct(r) {
			spaceBeforeLast, lastSpace = lastSpace, i
		}
		len++
		if len > maxLength {
			if lastSpace != -1 && lastSpace <= iMinus3 {
				return text[:lastSpace] + "..."
			}
			if spaceBeforeLast != -1 && spaceBeforeLast <= iMinus3 {
				return text[:spaceBeforeLast] + "..."
			}
			return text[:iMinus3] + "..."
		}
		iMinus3, iMinus2, iMinus1 = iMinus2, iMinus1, i
	}
	return text
}

func getLineLength() int {
	var (
		err    error
		length int = 132
		width  uint
	)
	width, err = terminal.Width()
	if err != nil {
		return length
	}

	return int(width)
}

func sortSlice(unsorted []int32) []int32 {
	sort.Slice(unsorted, func(i, j int) bool {
		return unsorted[i] < unsorted[j]
	})
	return unsorted
}

func getProcInfo(proc *process.Process) (string, string, string) {
	var (
		cmdArgs       []string
		cmdArgsJoined string
		cmdName       string
		err           error
		username      string
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
		cmdArgsJoined = ""
	} else {
		cmdArgsJoined = strings.Join(cmdArgs, " ")
	}
	return username, cmdName, cmdArgsJoined
}

func generateTree(parent int32, tree map[int32][]int32, indent string, opts Options) {
	var (
		cmdArgsJoined string
		cmdName       string
		line          string
		lineLength    int
		ok            bool
		username      string
	)
	lineLength = getLineLength()
	lineLength = lineLength - len(indent) - 6 // Accomodate the indentation, etc

	proc, err := process.NewProcess(parent)
	if err == nil {
		username, cmdName, cmdArgsJoined = getProcInfo(proc)
		if parent == 0 && cmdName == "unknown command" {
			cmdName = "kernel_task"
		}
		if !opts.ShowArgs {
			cmdArgsJoined = ""
		}

		line = fmt.Sprintf("%d %s %s %s", int(parent), username, cmdName, cmdArgsJoined)
		if opts.Long {
			fmt.Fprintln(os.Stdout, line)
		} else {
			fmt.Fprintln(os.Stdout, truncateEllipsis(line, lineLength))
		}
	} else {
		if parent == 0 {
			fmt.Fprintf(os.Stdout, "%d %s\n", parent, "kernel_task")
		}
	}
	_, ok = tree[parent]
	if !ok {
		return
	}

	children := tree[parent][:len(tree[parent])-1]
	for _, child := range children {
		fmt.Fprintf(os.Stdout, "%s  ├── ", indent)
		generateTree(child, tree, indent+"  │ ", opts)
	}

	child := returnLastElement(tree[parent])
	fmt.Fprintf(os.Stdout, "%s  └── ", indent)
	generateTree(child, tree, indent+"    ", opts)
}

func getTreeData2(opts Options) map[int32][]int32 {
	tree := make(map[int32][]int32)

	// Get a list of all processes
	processes, err := process.Processes()
	if err != nil {
		log.Fatalf("Failed to get processes: %v", err)
	}

	for _, p := range processes {
		ppid, err := p.Ppid()
		if err != nil {
			// Ignore processes that have disappeared
			continue
		}
		procUser, _ := p.Username()
		if opts.User != "" && procUser == opts.User {
			pid := int(p.Pid)
			tree[ppid] = append(tree[ppid], int32(pid))
		}
	}

	// On systems where PID 0 exists, ensure it doesn't reference itself
	if children, exists := tree[0]; exists {
		for i, pid := range children {
			if pid == 0 {
				tree[0] = append(children[:i], children[i+1:]...)
				break
			}
		}
	}

	return tree
}

func main() {
	var (
		indent      string = ""
		startingPid int32
		tree        map[int32][]int32
	)
	opts := Options{}
	opts.Version = func() {
		fmt.Println(VERSION)
		os.Exit(0)
	}

	// Parse the options
	parser := flags.NewParser(&opts, flags.Default)
	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}

	tree = getTreeData2(opts)
	startingPid = findFirstPid(tree)
	if opts.Start > 0 {
		startingPid = opts.Start
	}
	generateTree(startingPid, tree, indent, opts)
}
