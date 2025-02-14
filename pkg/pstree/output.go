package pstree

import (
	"fmt"
	"os"
	"strings"

	"github.com/gdanko/pstree/util"
	"github.com/giancarlosio/gorainbow"
)

type DisplayOptions struct {
	ColorizeOutput  bool
	GraphicsMode    int
	HidePids        bool
	MaxDepth        int
	RainbowOutput   bool
	ShowArguments   bool
	ShowCpuPercent  bool
	ShowMemoryUsage bool
	ShowNumThreads  bool
	WideDisplay     bool
}

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
		P:    "─",
		PGL:  "¤",
		NPGL: "─",
		BarC: "├",
		Bar:  "│",
		BarL: "└",
		SG:   "",
		EG:   "",
		Init: "",
	},
	"vt100": {
		S2:   "──",
		P:    "─┬",
		PGL:  "◆",
		NPGL: "─",
		BarC: "├",
		Bar:  "│",
		BarL: "└",
		SG:   "\x0E",
		EG:   "\x0F",
		Init: "\033(B\033)0",
	},
	"utf8": {
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
	// "pc850-unused": {
	// 	S2:   "─",
	// 	P:    "├",
	// 	PGL:  "¤",
	// 	NPGL: "─",
	// 	BarC: "│",
	// 	Bar:  "│",
	// 	BarL: "└",
	// 	SG:   "",
	// 	EG:   "",
	// 	Init: "",
	// },
	// "utf8-unused": {
	// 	S2:   "──",
	// 	P:    "├─",
	// 	PGL:  "●",
	// 	NPGL: "─",
	// 	BarC: "│",
	// 	Bar:  "│",
	// 	BarL: "└─",
	// 	SG:   "",
	// 	EG:   "",
	// 	Init: "",
	// },
	// "vt100-unused": {
	// 	S2:   "qq",
	// 	P:    "qw",
	// 	PGL:  "`",
	// 	NPGL: "q",
	// 	BarC: "t",
	// 	Bar:  "x",
	// 	BarL: "m",
	// 	SG:   "\x0E",
	// 	EG:   "\x0F",
	// 	Init: "\033(B\033)0",
	// },
}

// pstree.PrintTree(processes, startingPidIndex, "", screenWidth, currentLevel, displayOptions)
// func PrintTree(processes []Process, me int, head string, screenWidth int, flagArguments bool, flagNoPids bool, flagGraphicsMode int, flagWide bool, currentLevel int, flagLevel int, flagCpu bool, flagThreads bool, flagColor bool) {
func PrintTree(processes []Process, me int, head string, screenWidth int, currentLevel int, displayOptions DisplayOptions) {
	var (
		args        string = ""
		C           TreeChars
		cpuPercent  string = ""
		line        string
		linePrefix  string
		memoryUsage string
		pidString   string
		threads     string = ""
	)

	if currentLevel == displayOptions.MaxDepth {
		return
	}

	switch displayOptions.GraphicsMode {
	case 1:
		C = TreeStyles["pc850"]
	case 2:
		C = TreeStyles["vt100"]
	case 3:
		C = TreeStyles["utf8"]
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
	pidString = fmt.Sprintf(" %05s", util.Int32toStr(processes[me].PID))

	if displayOptions.ShowArguments {
		if len(processes[me].Args) > 0 {
			args = strings.Join(processes[me].Args, "")
		}
	}

	if displayOptions.ShowCpuPercent {
		cpuPercent = fmt.Sprintf(" (c: %.2f%%)", processes[me].CPUPercent)
	}

	if displayOptions.ShowMemoryUsage {
		memoryUsage = fmt.Sprintf(" (m: %s)", util.ByteConverter(processes[me].MemoryInfo.RSS))
	}

	if displayOptions.ShowNumThreads {
		threads = fmt.Sprintf(" (t: %d)", processes[me].NumThreads)
	}

	if displayOptions.ColorizeOutput {
		linePrefix = util.ColorGreen(linePrefix)
		cpuPercent = util.ColorYellow(cpuPercent)
		memoryUsage = util.ColorOrange(memoryUsage)
		threads = util.ColorWhite(threads)
		processes[me].Username = util.ColorCyan(processes[me].Username)
		pidString = util.ColorPurple(pidString)
		processes[me].Command = util.ColorBlue(processes[me].Command)
		args = util.ColorRed(args)
	}

	if displayOptions.HidePids {
		pidString = ""
	}

	line = fmt.Sprintf("%s%s%s%s%s %s %s %s", linePrefix, pidString, cpuPercent, memoryUsage, threads, processes[me].Username, processes[me].Command, args)

	if !displayOptions.WideDisplay {
		if len(line) > screenWidth {
			if displayOptions.RainbowOutput {
				line = util.TruncateANSI(gorainbow.Rainbow(line), screenWidth)
			} else {
				line = util.TruncateANSI(line, screenWidth)
			}
		} else {
			if displayOptions.RainbowOutput {
				line = gorainbow.Rainbow(line)
			}
		}
	} else {
		if displayOptions.RainbowOutput {
			line = gorainbow.Rainbow(line)
		}
	}
	fmt.Fprintln(os.Stdout, line)

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
		PrintTree(processes, childme, newHead, screenWidth, currentLevel, displayOptions)
		childme = nextChild
	}
}
