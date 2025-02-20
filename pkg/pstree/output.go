package pstree

import (
	"fmt"
	"log/slog"
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
	NoPids          bool
	RainbowOutput   bool
	ShowArguments   bool
	ShowCpuPercent  bool
	ShowMemoryUsage bool
	ShowNumThreads  bool
	ShowPGIDs       bool
	ShowProcessAge  bool
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
}

func PrintTree(logger *slog.Logger, processes []Process, me int, head string, screenWidth int, currentLevel int, displayOptions DisplayOptions) {
	var (
		ageString   string = ""
		args        string = ""
		C           TreeChars
		cpuPercent  string
		flag        *string
		flags       []*string
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

	if displayOptions.ShowProcessAge {
		duration := util.FindDuration(processes[me].Age)
		ageSlice := []string{}
		ageSlice = append(ageSlice, fmt.Sprintf("%02d", duration.Days))
		ageSlice = append(ageSlice, fmt.Sprintf("%02d", duration.Hours))
		ageSlice = append(ageSlice, fmt.Sprintf("%02d", duration.Minutes))
		ageSlice = append(ageSlice, fmt.Sprintf("%02d", duration.Seconds))
		ageString = fmt.Sprintf(
			" (%s)",
			strings.Join(ageSlice, ":"),
		)
	}

	if !displayOptions.NoPids {
		pidString = fmt.Sprintf(" %05s", util.Int32toStr(processes[me].PID))

	}

	if displayOptions.ShowPGIDs {
		pgidString = fmt.Sprintf(" %05s", util.Int32toStr(processes[me].PGID))
	}

	if displayOptions.ShowArguments {
		if len(processes[me].Args) > 0 {
			args = strings.Join(processes[me].Args, "")
		}
	}

	if displayOptions.ShowCpuPercent {
		cpuPercent = fmt.Sprintf(" (c:%.2f%%)", processes[me].CPUPercent)
	}

	if displayOptions.ShowMemoryUsage {
		memoryUsage = fmt.Sprintf(" (m:%s)", util.ByteConverter(processes[me].MemoryInfo.RSS))
	}

	if displayOptions.ShowNumThreads {
		threads = fmt.Sprintf(" (t:%d)", processes[me].NumThreads)
	}

	if displayOptions.ColorizeOutput {
		util.ColorGreen(&linePrefix)
		util.ColorYellow(&cpuPercent)
		util.ColorOrange(&memoryUsage)
		util.ColorWhite(&threads)
		util.ColorCyan(&processes[me].Username)
		util.ColorBoldBlue(&pgidString)
		util.ColorPurple(&pidString)
		util.ColorBoldGreen(&ageString)
		util.ColorBlue(&processes[me].Command)
		util.ColorRed(&args)
	} else if displayOptions.ColorAttr != "" {
		flags = []*string{&cpuPercent, &memoryUsage, &threads, &processes[me].Username, &pidString, &pgidString, &ageString, &processes[me].Command, &args}
		if displayOptions.ColorAttr == "age" {
			displayOptions.ShowProcessAge = true
			for _, flag = range flags {
				if processes[me].Age < 60 {
					util.ColorGreen(flag)
				} else if processes[me].Age > 60 && processes[me].Age < 3600 {
					util.ColorYellow(flag)
				} else {
					util.ColorRed(flag)
				}
			}
		} else if displayOptions.ColorAttr == "cpu" {
			displayOptions.ShowCpuPercent = true
			for _, flag = range flags {
				if processes[me].CPUPercent < 5 {
					util.ColorGreen(flag)
				} else if processes[me].CPUPercent > 5 && processes[me].CPUPercent < 15 {
					util.ColorYellow(flag)
				} else {
					util.ColorRed(flag)
				}
			}
		} else if displayOptions.ColorAttr == "mem" {
			displayOptions.ShowMemoryUsage = true
			for _, flag = range flags {
				percent := (processes[me].MemoryInfo.RSS / displayOptions.InstalledMemory) * 100
				if percent < 10 {
					util.ColorGreen(flag)
				} else if processes[me].MemoryInfo.RSS > 10 && processes[me].MemoryInfo.RSS < 20 {
					util.ColorYellow(flag)
				} else {
					util.ColorRed(flag)
				}
			}
		}
	}

	line = fmt.Sprintf("%s%s%s%s%s%s%s %s %s %s", linePrefix, pidString, pgidString, ageString, cpuPercent, memoryUsage, threads, processes[me].Username, processes[me].Command, args)

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
		PrintTree(logger, processes, childme, newHead, screenWidth, currentLevel, displayOptions)
		childme = nextChild
	}
}
