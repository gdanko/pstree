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
	ColorAttr           string
	ColorCount          int
	ColorizeOutput      bool
	ColorSupport        bool
	CompactMode         bool
	GraphicsMode        int
	HidePGL             bool
	HideThreads         bool
	IBM850Graphics      bool
	InstalledMemory     uint64
	MaxDepth            int
	RainbowOutput       bool
	ShowArguments       bool
	ShowCpuPercent      bool
	ShowMemoryUsage     bool
	ShowNumThreads      bool
	ShowOwner           bool
	ShowPGIDs           bool
	ShowPids            bool
	ShowProcessAge      bool
	ShowUIDTransitions  bool
	ShowUserTransitions bool
	UTF8Graphics        bool
	VT100Graphics       bool
	WideDisplay         bool
}

// Tree characters struct (equivalent to C's `TreeChars`)
type TreeChars struct {
	S2, P, PGL, NPGL, BarC, Bar, BarL, SG, EG, Init string
}

// Define different graphical styles
var TreeStyles = map[string]TreeChars{
	"ascii": {
		Bar:  "|",
		BarC: "|",
		BarL: "\\",
		EG:   "",
		Init: "",
		NPGL: "-",
		P:    "-+",
		PGL:  "=",
		S2:   "--",
		SG:   "",
	},
	"pc850": {
		Bar:  "│",
		BarC: "├",
		BarL: "└",
		EG:   "",
		Init: "",
		NPGL: "─",
		P:    "─",
		PGL:  "¤",
		S2:   "─",
		SG:   "",
	},
	"vt100": {
		Bar:  "│",
		BarC: "├",
		BarL: "└",
		EG:   "\x0F",
		Init: "\033(B\033)0",
		NPGL: "─",
		P:    "─┬",
		PGL:  "◆",
		S2:   "──",
		SG:   "\x0E",
	},
	"utf8": {
		Bar:  "\342\224\202",
		BarC: "\342\224\234",
		BarL: "\342\224\224",
		EG:   "",
		Init: "",
		NPGL: "\342\224\200",
		P:    "\342\224\200\342\224\254",
		PGL:  "●",
		S2:   "\342\224\200\342\224\200",
		SG:   "",
	},
}

// colorizeField applies appropriate color formatting to a specific field in the process tree output.
//
// This function handles coloring of different elements in the process tree display based on the
// current display options and the field type. It supports two main coloring modes:
//
//  1. Standard colorization (--colorize): Each field type gets a predefined color
//  2. Attribute-based colorization (--color): Colors are applied based on process attributes
//     like CPU or memory usage, with thresholds determining the color (green/yellow/red)
//
// Parameters:
//   - fieldName: String identifying which field is being colored (e.g., "cpu", "memory", "command")
//   - value: Pointer to the string value to be colored (modified in-place)
//   - displayOptions: Pointer to the current display options configuration
//   - process: Pointer to the Process struct containing process information
func colorizeField(fieldName string, value *string, displayOptions *DisplayOptions, process *Process) {
	// Only apply colors if the terminal supports them
	if displayOptions.ColorSupport {
		// Standard colorization mode (--colorize flag)
		if displayOptions.ColorizeOutput {
			// Apply specific colors based on the field type
			switch fieldName {
			case "age":
				util.ColorBoldGreen(value)
			case "args":
				util.ColorRed(value)
			case "connector":
				util.ColorGray(value)
			case "command":
				util.ColorBlue(value)
			case "compactStr":
				util.ColorGray(value)
			case "cpu":
				util.ColorYellow(value)
			case "memory":
				util.ColorOrange(value)
			case "owner":
				util.ColorCyan(value)
			case "ownerTransition":
				util.ColorGray(value)
			case "pidPgid":
				util.ColorPurple(value)
			case "prefix":
				util.ColorGreen(value)
			case "threads":
				util.ColorBoldWhite(value)
			case "username":
				util.ColorBlue(value)
			}
		} else if displayOptions.ColorAttr != "" {
			// Attribute-based colorization mode (--color flag)
			// Don't apply attribute-based coloring to the tree prefix
			if fieldName != "prefix" {
				switch displayOptions.ColorAttr {
				case "age":
					// Ensure process age is shown when coloring by age
					displayOptions.ShowProcessAge = true

					// Apply color based on process age thresholds in seconds
					if process.Age < 60 {
						// Low age (< 1 minute)
						util.ColorRed(value)
					} else if process.Age >= 60 && process.Age < 3600 {
						// Medium age (< 1 hour)
						util.ColorOrange(value)
					} else if process.Age >= 3600 && process.Age < 86400 {
						// High age (> 1 hour and < 1 day)
						util.ColorYellow(value)
					} else if process.Age >= 86400 {
						// Very high age (> 1 day)
						util.ColorGreen(value)
					}
				case "cpu":
					// Ensure CPU percentage is shown when coloring by CPU
					displayOptions.ShowCpuPercent = true

					// Apply color based on CPU usage thresholds in percentage
					if process.CPUPercent < 5 {
						// Low CPU usage (< 5%)
						util.ColorGreen(value)
					} else if process.CPUPercent >= 5 && process.CPUPercent < 15 {
						// Medium CPU usage (5-15%)
						util.ColorYellow(value)
					} else if process.CPUPercent >= 15 {
						// High CPU usage (> 15%)
						util.ColorRed(value)
					}
				case "mem":
					// Ensure memory usage is shown when coloring by memory
					displayOptions.ShowMemoryUsage = true

					// Calculate memory usage as percentage of total system memory
					percent := (process.MemoryInfo.RSS / displayOptions.InstalledMemory) * 100

					// Apply color based on memory usage thresholds in percentage
					if percent < 10 {
						// Low memory usage (< 10%)
						util.ColorGreen(value)
					} else if percent >= 10 && percent < 20 {
						// Medium memory usage (10-20%)
						util.ColorOrange(value)
					} else if percent >= 20 {
						// High memory usage (> 20%)
						util.ColorRed(value)
					}
				}
			}
		}
	}
}

// PrintTree recursively prints a process tree with customizable formatting options.
//
// This function displays a process and all its children in a tree-like structure,
// with various display options such as process age, CPU usage, memory usage, etc.
// The tree is formatted using different graphical styles based on the display options.
//
// Parameters:
//   - logger: Logger instance for debug information
//   - processes: Slice of Process structs representing the process tree
//   - me: Index of the current process to print
//   - head: String representing the indentation and tree structure for the current line
//   - screenWidth: Width of the screen for truncating output if needed
//   - currentLevel: Current depth level in the process tree
//   - displayOptions: Configuration options for display formatting
func PrintTree(logger *slog.Logger, processes []Process, me int, head string, screenWidth int, currentLevel int, displayOptions *DisplayOptions) {
	// Skip if we've reached the maximum depth
	if displayOptions.MaxDepth > 0 && currentLevel > displayOptions.MaxDepth {
		return
	}

	// Initialize compact mode if enabled and at the root level
	if currentLevel == 0 {
		// Always initialize compact mode to identify duplicates
		// But we'll respect the CompactMode flag when displaying
		InitCompactMode(processes)
	}

	// Skip this process if it's been marked as a duplicate in compact mode
	// Only skip if compact mode is actually enabled
	if displayOptions.CompactMode && ShouldSkipProcess(me) {
		return
	}

	var (
		ageString       string = ""
		args            string = ""
		C               TreeChars
		cpuPercent      string
		line            string
		linePrefix      string
		memoryUsage     string
		newHead         string
		outputSlice     []string
		owner           string
		ownerTransition string = ""
		pgidString      string
		pidPgidSlice    []string
		pidPgidString   string
		pidString       string
		threads         string
	)

	if currentLevel > displayOptions.MaxDepth {
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
		// Check if this process has a visible sibling
		hasVisibleSibling := false
		sibling := processes[me].Sister

		// In compact mode, we need to check if all siblings are going to be skipped
		if displayOptions.CompactMode {
			for sibling != -1 {
				if !ShouldSkipProcess(sibling) {
					hasVisibleSibling = true
					break
				}
				sibling = processes[sibling].Sister
			}
		} else {
			// In normal mode, just check if there's a sibling
			hasVisibleSibling = (sibling != -1)
		}

		if hasVisibleSibling {
			part1 = C.BarC // T-connector for processes with visible siblings
		} else {
			part1 = C.BarL // L-connector for processes without visible siblings (last child)
		}
	}

	var part2 string
	if processes[me].Child != -1 && currentLevel < displayOptions.MaxDepth {
		part2 = C.P
	} else {
		part2 = C.S2
	}

	var part3 string
	if processes[me].PID == processes[me].PGID {
		if displayOptions.HidePGL {
			part3 = C.NPGL
		} else {
			part3 = C.PGL
		}
	} else {
		part3 = C.NPGL
	}

	linePrefix = fmt.Sprintf("%s%s%s%s%s%s", C.SG, head, part1, part2, part3, C.EG)

	colorizeField("prefix", &linePrefix, displayOptions, &processes[me])
	outputSlice = append(outputSlice, linePrefix)

	if displayOptions.ShowOwner {
		owner = processes[me].Username
		colorizeField("owner", &owner, displayOptions, &processes[me])
		outputSlice = append(outputSlice, owner)
	}

	if displayOptions.ShowPids {
		pidString = util.Int32toStr(processes[me].PID)
		pidPgidSlice = append(pidPgidSlice, pidString)
	}

	if displayOptions.ShowPGIDs {
		pgidString = util.Int32toStr(processes[me].PGID)
		pidPgidSlice = append(pidPgidSlice, pgidString)
	}

	if len(pidPgidSlice) > 0 {
		pidPgidString = fmt.Sprintf("(%s)", strings.Join(pidPgidSlice, ","))
		colorizeField("pidPgid", &pidPgidString, displayOptions, &processes[me])
		outputSlice = append(outputSlice, pidPgidString)
	}

	if displayOptions.ShowProcessAge {
		duration := util.FindDuration(processes[me].Age)
		ageSlice := []string{}
		ageSlice = append(ageSlice, fmt.Sprintf("%02d", duration.Days))
		ageSlice = append(ageSlice, fmt.Sprintf("%02d", duration.Hours))
		ageSlice = append(ageSlice, fmt.Sprintf("%02d", duration.Minutes))
		ageSlice = append(ageSlice, fmt.Sprintf("%02d", duration.Seconds))
		ageString = fmt.Sprintf(
			"(%s)",
			strings.Join(ageSlice, ":"),
		)
		colorizeField("age", &ageString, displayOptions, &processes[me])
		outputSlice = append(outputSlice, ageString)
	}

	if displayOptions.ShowCpuPercent {
		cpuPercent = fmt.Sprintf("(c:%.2f%%)", processes[me].CPUPercent)
		colorizeField("cpu", &cpuPercent, displayOptions, &processes[me])
		outputSlice = append(outputSlice, cpuPercent)
	}

	if displayOptions.ShowMemoryUsage {
		memoryUsage = fmt.Sprintf("(m:%s)", util.ByteConverter(processes[me].MemoryInfo.RSS))
		colorizeField("memory", &memoryUsage, displayOptions, &processes[me])
		outputSlice = append(outputSlice, memoryUsage)
	}

	if displayOptions.ShowNumThreads {
		// Always show thread count, even when showing compact format
		threads = fmt.Sprintf("(t:%d)", processes[me].NumThreads)
		colorizeField("threads", &threads, displayOptions, &processes[me])
		outputSlice = append(outputSlice, threads)
	}

	if displayOptions.ShowUIDTransitions && processes[me].HasUIDTransition {
		// Add UID transition notation {parentUID→currentUID}
		if len(processes[me].UIDs) > 0 {
			ownerTransition = fmt.Sprintf("(%d→%d)", processes[me].ParentUID, processes[me].UIDs[0])
		}
	} else if displayOptions.ShowUserTransitions && processes[me].HasUIDTransition {
		// Add user transition notation {parentUser→currentUser}
		if processes[me].ParentUsername != "" {
			ownerTransition = fmt.Sprintf("(%s→%s)", processes[me].ParentUsername, processes[me].Username)
		}
	}

	if ownerTransition != "" {
		colorizeField("ownerTransition", &ownerTransition, displayOptions, &processes[me])
		outputSlice = append(outputSlice, ownerTransition)
	}

	// Get the command - use full path when compact mode is disabled
	commandStr := processes[me].Command

	// Determine if this is a thread
	isThread := processes[me].NumThreads > 0 && processes[me].PPID > 0

	// In compact mode, format the command with count for the first process in a group
	if displayOptions.CompactMode {
		// Get the count of identical processes
		count, processIsThread := GetProcessCount(processes, me)

		// If there are multiple identical processes, format with count
		if count > 1 {
			// For Linux pstree style format: command---N*[command] or command---N*[{command}]
			// Use the thread status from the process group if available
			if processIsThread {
				isThread = true
			}

			// Format in Linux pstree style
			compactStr := FormatCompactOutput(commandStr, count, isThread, displayOptions.HideThreads)

			if compactStr != "" {
				// Create the connector string
				connector := "───"

				// Colorize the connector and compact format indicator in green if color support is available
				if displayOptions.ColorSupport {
					// The util.ColorGreen function modifies the string in place via pointer
					colorizeField("connector", &connector, displayOptions, &processes[me])
					colorizeField("compactStr", &compactStr, displayOptions, &processes[me])
				}

				// In Linux pstree, the format is just the count and brackets, not repeating the command
				commandStr = fmt.Sprintf("%s%s%s", commandStr, connector, compactStr)
			}
		}
	}

	// For threads in non-compact mode, wrap the command in curly braces
	if !displayOptions.CompactMode && isThread && processes[me].NumThreads == 0 {
		// This is likely a thread of a process
		commandStr = fmt.Sprintf("{%s}", commandStr)
	}

	colorizeField("command", &commandStr, displayOptions, &processes[me])
	outputSlice = append(outputSlice, commandStr)

	if displayOptions.ShowArguments {
		if len(processes[me].Args) > 0 {
			args = strings.Join(processes[me].Args, " ")
			colorizeField("args", &args, displayOptions, &processes[me])
			outputSlice = append(outputSlice, args)
		}
	}

	line = strings.Join(outputSlice, " ")

	if !displayOptions.WideDisplay {
		if len(line) > screenWidth {
			if displayOptions.RainbowOutput {
				line = util.TruncateANSI(logger, gorainbow.Rainbow(line), screenWidth)
			} else {
				line = util.TruncateANSI(logger, line, screenWidth)
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
			// In compact mode, we need to check if any visible siblings exist
			if displayOptions.CompactMode {
				sibling := processes[me].Sister
				for sibling != -1 {
					if !ShouldSkipProcess(sibling) {
						return C.Bar // Only add vertical bar if there's a visible sibling
					}
					sibling = processes[sibling].Sister
				}
				return " " // No visible siblings
			} else {
				// In normal mode, just check if there's a sibling
				if processes[me].Sister != -1 {
					return C.Bar
				}
				return " "
			}
		}(),
	)

	// Iterate over children and determine sibling status
	childme := processes[me].Child
	for childme != -1 {
		nextChild := processes[childme].Sister
		PrintTree(logger, processes, childme, newHead, screenWidth, currentLevel+1, displayOptions)
		childme = nextChild
	}
}
