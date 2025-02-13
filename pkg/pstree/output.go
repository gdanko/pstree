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

func PrintTree(processes []Process, me int, head string, screenWidth int, flagArguments bool, flagNoPids bool, flagGraphicsMode int, flagWide bool, currentLevel int, flagLevel int, flagColorize bool) {
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
	if head == "" && !processes[me].Print {
		return
	}

	var part1 string
	if head == "" {
		part1 = ""
	} else {
		if processes[me].Sister != -1 {
			part1 = C.BarC
		} else {
			part1 = C.BarL
		}
	}

	var part2 string
	if processes[me].Child != -1 {
		part2 = C.P
	} else {
		part2 = C.S2
	}

	var part3 string
	if processes[me].PID == processes[me].PGID {
		part3 = C.PGL
	} else {
		part3 = C.NPGL
	}

	linePrefix = fmt.Sprintf("%s%s%s%s%s%s", C.SG, head, part1, part2, part3, C.EG)
	pidString = fmt.Sprintf("%05s", util.Int32toStr(processes[me].PID))

	if flagArguments {
		args = processes[me].Args
	}

	if flagColorize {
		linePrefix = util.ColorYellow(linePrefix)
		processes[me].Username = util.ColorCyan(processes[me].Username)
		pidString = util.ColorPurple(pidString)
		processes[me].Command = util.ColorBlue(processes[me].Command)
		args = util.ColorRed(args)
	}

	if flagNoPids {
		line = fmt.Sprintf("%s %s %s %s", linePrefix, processes[me].Username, processes[me].Command, args)
	} else {
		line = fmt.Sprintf("%s %s %s %s %s", linePrefix, pidString, processes[me].Username, processes[me].Command, args)
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
	newHead = fmt.Sprintf("%s%s ", head,
		func() string {
			if head == "" {
				return ""
			}
			if processes[me].Sister != -1 {
				return C.Bar
			}
			return " "
		}(),
	)

	// Iterate over children and determine sibling status
	childme := processes[me].Child
	for childme != -1 {
		nextChild := processes[childme].Sister
		PrintTree(processes, childme, newHead, screenWidth, flagArguments, flagNoPids, flagGraphicsMode, flagWide, currentLevel+1, flagLevel, flagColorize)
		childme = nextChild
	}
}
