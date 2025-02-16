package pstree

import (
	"fmt"
	"os"
	"strings"

	"github.com/gdanko/pstree/util"
	"github.com/giancarlosio/gorainbow"
)

type DisplayOptions struct {
	ColorAttr       string
	ColorizeOutput  bool
	GraphicsMode    int
	IBM850Graphics  bool
	InstalledMemory uint64
	MaxDepth        int
	RainbowOutput   bool
	ShowArguments   bool
	ShowCpuPercent  bool
	ShowMemoryUsage bool
	ShowNumThreads  bool
	ShowPGIDs       bool
	ShowPIDs        bool
	UTF8Graphics    bool
	VT100Graphics   bool
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

func colorGreen(text *string) {
	coloredText := "\033[32m" + *text + "\033[0m"
	*text = coloredText
}

func colorYellow(text *string) {
	coloredText := "\033[33m" + *text + "\033[0m"
	*text = coloredText
}

func colorRed(text *string) {
	coloredText := "\033[31m" + *text + "\033[0m"
	*text = coloredText
}

func PrintTree(processes []Process, me int, head string, screenWidth int, currentLevel int, displayOptions DisplayOptions) {
	var (
		args        string = ""
		C           TreeChars
		cpuPercent  string
		line        string
		linePrefix  string
		memoryUsage string
		newHead     string
		pidString   string
		pgidString  string
		threads     string
	)

	if currentLevel == displayOptions.MaxDepth {
		return
	}

	if displayOptions.IBM850Graphics {
		C = TreeStyles["pc850"]
	} else if displayOptions.UTF8Graphics {
		C = TreeStyles["utf8"]
	} else if displayOptions.VT100Graphics {
		C = TreeStyles["vt100"]
	} else {
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

	if displayOptions.ShowPIDs {
		pidString = fmt.Sprintf(" (%05s)", util.Int32toStr(processes[me].PID))
	}

	if displayOptions.ShowPGIDs {
		pgidString = fmt.Sprintf(" (%05s)", util.Int32toStr(processes[me].PGID))
	}

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
	} else if displayOptions.ColorAttr != "" {
		if displayOptions.ColorAttr == "age" {
			for _, flag := range []*string{&cpuPercent, &memoryUsage, &threads, &processes[me].Username, &pidString, &processes[me].Command, &args} {
				if processes[me].Age < 60 {
					colorGreen(flag)
				} else if processes[me].Age > 60 && processes[me].Age < 3600 {
					colorYellow(flag)
				} else {
					colorRed(flag)
				}
			}
		} else if displayOptions.ColorAttr == "cpu" {
			displayOptions.ShowCpuPercent = true
			for _, flag := range []*string{&cpuPercent, &memoryUsage, &threads, &processes[me].Username, &pidString, &processes[me].Command, &args} {
				if processes[me].CPUPercent < 5 {
					colorGreen(flag)
				} else if processes[me].CPUPercent > 5 && processes[me].CPUPercent < 15 {
					colorYellow(flag)
				} else {
					colorRed(flag)
				}
			}
		} else if displayOptions.ColorAttr == "mem" {
			displayOptions.ShowMemoryUsage = true
			for _, flag := range []*string{&cpuPercent, &memoryUsage, &threads, &processes[me].Username, &pidString, &processes[me].Command, &args} {
				percent := (processes[me].MemoryInfo.RSS / displayOptions.InstalledMemory) * 100
				if percent < 10 {
					colorGreen(flag)
				} else if processes[me].MemoryInfo.RSS > 10 && processes[me].MemoryInfo.RSS < 20 {
					colorYellow(flag)
				} else {
					colorRed(flag)
				}
			}
		}
	}

	line = fmt.Sprintf("%s%s%s%s%s%s %s %s %s", linePrefix, pidString, pgidString, cpuPercent, memoryUsage, threads, processes[me].Username, processes[me].Command, args)

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
