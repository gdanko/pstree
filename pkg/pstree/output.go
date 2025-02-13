package pstree

import (
	"fmt"
	"os"

	"github.com/gdanko/pstree/util"
)

// Tree characters struct (equivalent to C's `TreeChars`)
type TreeChars struct {
	S2, P, PGL, NPGL, BarC, Bar, BarL, SG, EG, Init string
}

// Define different graphical styles
var TreeStyles = map[string]TreeChars{
	"ascii": {
		S2:   "--",
		P:    "-+",
		PGL:  "=",
		NPGL: "-",
		BarC: "|",
		Bar:  "|",
		BarL: "\\",
		SG:   "",
		EG:   "",
		Init: "",
	},
	"pc850": {
		S2:   "─",
		P:    "├",
		PGL:  "¤",
		NPGL: "─",
		BarC: "│",
		Bar:  "│",
		BarL: "└",
		SG:   "",
		EG:   "",
		Init: "",
	},
	"vt100": {
		S2:   "qq",
		P:    "qw",
		PGL:  "`",
		NPGL: "q",
		BarC: "t",
		Bar:  "x",
		BarL: "m",
		SG:   "\x0E",
		EG:   "\x0F",
		Init: "\033(B\033)0",
	},
	"utf8": {
		S2:   "──",
		P:    "├─",
		PGL:  "●",
		NPGL: "─",
		BarC: "│",
		Bar:  "│",
		BarL: "└─",
		SG:   "",
		EG:   "",
		Init: "",
	},
	"utf8a": {
		S2:   "\342\224\200\342\224\200",
		P:    "\342\224\200\342\224\254",
		NPGL: "\342\224\200",
		BarC: "\342\224\234",
		Bar:  "\342\224\202",
		BarL: "\342\224\224",
		SG:   "",
		EG:   "",
		Init: "",
		// PGL:  "●",
		PGL: "=",
	},
}

func PrintTree(processes []Process, idx int, head string, screenWidth int, flagArguments bool, flagShowPids bool, flagGraphicsMode int, flagWide bool, currentLevel int, flagLevel int, flagColorize bool) {
	var (
		args       string = ""
		C          TreeChars
		line       string
		linePrefix string
		pidString  string
	)

	if currentLevel == flagLevel {
		return
	}

	switch flagGraphicsMode {
	case 1:
		C = TreeStyles["pc850"]
	case 2:
		C = TreeStyles["vt100"]
	case 3:
		C = TreeStyles["utf8a"]
	default:
		C = TreeStyles["ascii"]
	}
	if head == "" && !processes[idx].Print {
		return
	}

	var part1 string
	if head == "" {
		part1 = " "
	} else {
		if processes[idx].Sister != -1 {
			part1 = C.BarC
		} else {
			part1 = C.BarL
		}
	}

	var part2 string
	if processes[idx].Child != -1 {
		part2 = C.P
	} else {
		part2 = C.S2
	}

	var part3 string
	if processes[idx].PID == processes[idx].PGID {
		part3 = C.PGL
	} else {
		part3 = C.NPGL
	}

	linePrefix = fmt.Sprintf("%s%s%s%s%s%s", C.SG, head, part1, part2, part3, C.EG)
	pidString = util.Int32toStr(processes[idx].PID)

	if flagArguments {
		args = processes[idx].Args
	}

	if flagColorize {
		linePrefix = util.ColorYellow(linePrefix)
		processes[idx].Username = util.ColorCyan(processes[idx].Username)
		pidString = util.ColorPurple(pidString)
		processes[idx].Command = util.ColorBlue(processes[idx].Command)
		args = util.ColorRed(args)
	}

	if flagShowPids {
		line = fmt.Sprintf("%s %s %s %s %s", linePrefix, pidString, processes[idx].Username, processes[idx].Command, args)
	} else {
		line = fmt.Sprintf("%s %s %s %s", linePrefix, processes[idx].Username, processes[idx].Command, args)
	}

	if flagWide {
		fmt.Fprintln(os.Stdout, line)
	} else {
		if len(line) > screenWidth {
			fmt.Fprintln(os.Stdout, util.TruncateANSI(line, screenWidth))
		} else {
			fmt.Fprintln(os.Stdout, line)
		}
	}

	var newHead string
	newHead = head
	if processes[idx].Sister != -1 {
		newHead += C.Bar + " "
	} else {
		newHead += "  "
	}

	// Iterate over children and determine sibling status
	childIdx := processes[idx].Child
	for childIdx != -1 {
		nextChild := processes[childIdx].Sister
		PrintTree(processes, childIdx, newHead, screenWidth, flagArguments, flagShowPids, flagGraphicsMode, flagWide, currentLevel+1, flagLevel, flagColorize)
		childIdx = nextChild
	}
}
