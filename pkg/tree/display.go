package tree

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gdanko/pstree/util"
	"github.com/giancarlosio/gorainbow"
	"golang.org/x/term"
)

//------------------------------------------------------------------------------
// TREE TRAVERSAL AND DISPLAY
//------------------------------------------------------------------------------
// Functions in this section handle the recursive traversal of the process tree
// and the display of processes with their relationships.

// PrintTree recursively prints a process tree with customizable formatting options.
//
// This function displays a process and all its children in a tree-like structure,
// with various display options such as process age, CPU usage, memory usage, etc.
// The tree is formatted using different graphical styles based on the display options.
//
// Parameters:
//   - pidIndex: Index of the current process to print
//   - head: String representing the indentation and tree structure for the current line
//
// Refactoring opportunity: This function could be split into:
// - printCurrentNode: Print just the current node
// - printChildNodes: Handle the recursive printing of child nodes
func (processTree *ProcessTree) PrintTree(pidIndex int, head string) {
	processTree.Logger.Debug(fmt.Sprintf("Entering processTree.PrintTree() with %d nodes", len(processTree.Nodes)))
	processTree.Logger.Debug(fmt.Sprintf("processTree.PrintTree(pidIndex=%d, head=\"%s\", atDepth=%d)", pidIndex, head, processTree.AtDepth))
	// https://github.com/FredHucht/pstree/blob/main/pstree.c#L721-L777
	// Skip if we've reached the maximum depth
	if processTree.DisplayOptions.MaxDepth > 0 && processTree.AtDepth > processTree.DisplayOptions.MaxDepth {
		processTree.Logger.Debug(fmt.Sprintf("Skipping process %d at depth %d (max depth %d)", processTree.Nodes[pidIndex].PID, processTree.AtDepth, processTree.DisplayOptions.MaxDepth))
		return
	}

	// Initialize compact mode if enabled and at the root level
	if processTree.AtDepth == 0 {
		// Always initialize compact mode to identify duplicates
		// But we'll respect the CompactMode flag when displaying
		processTree.Logger.Debug("Initializing compact mode")
		processTree.InitCompactMode()
	}

	// Skip this process if it's been marked as a duplicate in compact mode
	// Only skip if compact mode is actually enabled
	if processTree.DisplayOptions.CompactMode && processTree.ShouldSkipProcess(pidIndex) {
		processTree.Logger.Debug(fmt.Sprintf("Skipping PID %d in compact mode", processTree.Nodes[pidIndex].PID))
		return
	}

	var (
		line    string
		newHead string
	)

	if processTree.AtDepth > processTree.DisplayOptions.MaxDepth {
		processTree.Logger.Debug(fmt.Sprintf("Skipping process %d at depth %d (max depth %d)", processTree.Nodes[pidIndex].PID, processTree.AtDepth, processTree.DisplayOptions.MaxDepth))
		return
	}

	if head == "" && !processTree.Nodes[pidIndex].Print {
		processTree.Logger.Debug(fmt.Sprintf("Skipping process %d because head is empty and Print is false", processTree.Nodes[pidIndex].PID))
		return
	}

	line = processTree.buildLineItem(head, pidIndex)

	// If output is not a terminal, strip color
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		line = processTree.stripANSI(line)
		if len(line) > processTree.DisplayOptions.ScreenWidth {
			if !processTree.DisplayOptions.WideDisplay {
				line = processTree.truncatePlain(line)
			}
		}
	} else {
		if !processTree.DisplayOptions.WideDisplay {
			if len(line) > processTree.DisplayOptions.ScreenWidth {
				if processTree.DisplayOptions.RainbowOutput {
					line = processTree.truncateANSI(gorainbow.Rainbow(line))
				} else {
					line = processTree.truncateANSI(line)
				}
			} else {
				if processTree.DisplayOptions.RainbowOutput {
					line = gorainbow.Rainbow(line)
				}
			}
		} else {
			if processTree.DisplayOptions.RainbowOutput {
				line = gorainbow.Rainbow(line)
			}
		}
	}

	newHead = processTree.buildNewHead(head, pidIndex)

	processTree.Logger.Debug(fmt.Sprintf("processTree.PrintTree(): printing line for node.PID=%d, head=\"%s\"", processTree.Nodes[pidIndex].PID, head))
	fmt.Fprintln(os.Stdout, line)

	// Print threads for this process if any exist and threads are not hidden
	if !processTree.DisplayOptions.HideThreads && len(processTree.Nodes[pidIndex].Threads) > 0 {
		processTree.PrintThreads(pidIndex, newHead)
	}

	// Iterate over children and determine sibling status
	childme := processTree.Nodes[pidIndex].Child
	for childme != -1 {
		nextChild := processTree.Nodes[childme].Sister
		processTree.AtDepth++
		processTree.PrintTree(childme, newHead)
		processTree.AtDepth--
		childme = nextChild
	}
}

//------------------------------------------------------------------------------
// TREE VISUALIZATION AND FORMATTING
//------------------------------------------------------------------------------
// Functions in this section handle the visual representation of the process tree,
// including line formatting, tree branch construction, and display options.

// buildLinePrefix constructs the tree visualization prefix for a process node in the tree display.
// It creates the branch connectors (├, └, etc.) that show the hierarchical relationship between processes.
//
// Parameters:
//   - head: The accumulated prefix string from parent levels
//   - pidIndex: Index of the current process in the Nodes array
//
// Returns:
//   - A formatted string containing tree branch characters that represent the process's position in the hierarchy
//
// Refactoring opportunity: This function is complex and could be broken down into:
// - determineNodePosition: Determine if node is last child, has siblings, etc.
// - selectBranchCharacters: Select appropriate branch characters based on position
// - formatPrefix: Format the final prefix string
func (processTree *ProcessTree) buildLinePrefix(head string, pidIndex int) string {
	processTree.Logger.Debug(fmt.Sprintf("processTree.buildLinePrefix(head=\"%s\", pidIndex=%d, atDepth=%d)", head, pidIndex, processTree.AtDepth))

	// Create a strings.Builder with an estimated capacity
	// This helps avoid reallocations as the builder grows
	var builder strings.Builder

	// Pre-allocate capacity based on expected size
	// This is an optimization to avoid reallocations
	// You can adjust the capacity based on typical usage patterns
	builder.Grow(len(head) + 50) // Estimate based on typical usage

	// Append initialization sequences
	builder.WriteString(processTree.TreeChars.Init)
	builder.WriteString(processTree.TreeChars.SG)
	builder.WriteString(head)

	if processTree.Nodes[pidIndex].PID == 1 {
		// This is a worakround
		builder.WriteString(processTree.TreeChars.P)
		if processTree.DisplayOptions.ShowPGLs {
			builder.WriteString(processTree.TreeChars.PGL)
		} else {
			builder.WriteString(processTree.TreeChars.NPGL)
		}
		builder.WriteString(processTree.TreeChars.EG)
		return builder.String()
	}

	if head == "" {
		return ""
	} else {
		// Check if this process has a visible sibling
		hasVisibleSibling := false
		sibling := processTree.Nodes[pidIndex].Sister

		// In compact mode, we need to check if all siblings are going to be skipped
		if processTree.DisplayOptions.CompactMode {
			for sibling != -1 {
				if !processTree.ShouldSkipProcess(sibling) {
					hasVisibleSibling = true
					break
				}
				sibling = processTree.Nodes[sibling].Sister
			}
		} else {
			// In normal mode, just check if there's a sibling
			hasVisibleSibling = (sibling != -1)
		}

		if hasVisibleSibling {
			builder.WriteString(processTree.TreeChars.BarC) // T-connector for processes with visible siblings
		} else {
			builder.WriteString(processTree.TreeChars.BarL) // L-connector for processes without visible siblings (last child)
		}
	}

	// Check if this process has children or threads
	hasChildren := processTree.Nodes[pidIndex].Child != -1 && processTree.AtDepth < processTree.DisplayOptions.MaxDepth
	hasThreads := !processTree.DisplayOptions.HideThreads && len(processTree.Nodes[pidIndex].Threads) > 0

	// Add branch character if the process has children or threads
	if hasChildren || hasThreads {
		builder.WriteString(processTree.TreeChars.P)
	} else {
		builder.WriteString(processTree.TreeChars.S2)
	}

	if processTree.Nodes[pidIndex].PID == processTree.Nodes[pidIndex].PGID {
		if !processTree.DisplayOptions.ShowPGLs {
			builder.WriteString(processTree.TreeChars.NPGL)
		} else {
			builder.WriteString(processTree.TreeChars.PGL)
		}
	} else {
		builder.WriteString(processTree.TreeChars.NPGL)
	}
	builder.WriteString(processTree.TreeChars.EG)

	// Return the completed string
	return builder.String()
}

// buildLineItem constructs a complete formatted line for a process in the tree display.
// It combines the tree structure prefix with various process information based on display options.
//
// Parameters:
//   - head: The accumulated prefix string from parent levels
//   - pidIndex: Index of the current process in the Nodes array
//
// Returns:
//   - A fully formatted string containing the process information with appropriate formatting and coloring.
//     The string includes elements such as tree structure, process IDs, resource usage, and command information
//     based on the configured display options.
//
// Refactoring opportunity: This function is very large and could be broken down into:
// - formatProcessIDs: Format PID, PPID, PGID information
// - formatResourceUsage: Format CPU, memory, thread information
// - formatCommandInfo: Format command and arguments
// - formatOwnerInfo: Format username and UID transition information
func (processTree *ProcessTree) buildLineItem(head string, pidIndex int) string {
	processTree.Logger.Debug(fmt.Sprintf("processTree.buildLineItem(head=\"%s\", pidIndex=%d, atDepth=%d)", head, pidIndex, processTree.AtDepth))
	var (
		ageString       string
		args            string
		commandStr      string
		compactStr      string
		connector       string
		cpuPercent      string
		linePrefix      string
		memoryUsage     string
		owner           string
		ownerTransition string
		pgidString      string
		pidPgidSlice    []string
		pidPgidString   string
		pidString       string
		ppidString      string
		threads         string
	)

	// Create a strings.Builder with an estimated capacity
	// This helps avoid reallocations as the builder grows
	var builder strings.Builder

	// Pre-allocate capacity based on expected size
	// This is an optimization to avoid reallocations
	// You can adjust the capacity based on typical usage patterns
	builder.Grow(len(head) + 260) // Estimate based on typical usage

	linePrefix = processTree.buildLinePrefix(head, pidIndex)
	processTree.colorizeField("prefix", &linePrefix, pidIndex)

	builder.WriteString(linePrefix)
	builder.WriteString(" ")

	if processTree.DisplayOptions.ShowOwner {
		owner = processTree.Nodes[pidIndex].Username
		processTree.colorizeField("owner", &owner, pidIndex)
		builder.WriteString(owner)
		builder.WriteString(" ")
	}

	if processTree.DisplayOptions.ShowPPIDs {
		ppidString = util.Int32toStr(processTree.Nodes[pidIndex].PPID)
		pidPgidSlice = append(pidPgidSlice, ppidString)
	}

	if processTree.DisplayOptions.ShowPIDs {
		pidString = util.Int32toStr(processTree.Nodes[pidIndex].PID)
		pidPgidSlice = append(pidPgidSlice, pidString)
	}

	if processTree.DisplayOptions.ShowPGIDs {
		pgidString = util.Int32toStr(processTree.Nodes[pidIndex].PGID)
		pidPgidSlice = append(pidPgidSlice, pgidString)
	}

	if len(pidPgidSlice) > 0 {
		pidPgidString = fmt.Sprintf("(%s)", strings.Join(pidPgidSlice, ","))
		processTree.colorizeField("pidPgid", &pidPgidString, pidIndex)
		builder.WriteString(pidPgidString)
		builder.WriteString(" ")
	}

	if processTree.DisplayOptions.ShowProcessAge {
		duration := util.FindDuration(processTree.Nodes[pidIndex].Age)
		ageSlice := []string{}
		ageSlice = append(ageSlice, fmt.Sprintf("%02d", duration.Days))
		ageSlice = append(ageSlice, fmt.Sprintf("%02d", duration.Hours))
		ageSlice = append(ageSlice, fmt.Sprintf("%02d", duration.Minutes))
		ageSlice = append(ageSlice, fmt.Sprintf("%02d", duration.Seconds))
		ageString = fmt.Sprintf(
			"(%s)",
			strings.Join(ageSlice, ":"),
		)
		processTree.colorizeField("age", &ageString, pidIndex)
		builder.WriteString(ageString)
		builder.WriteString(" ")
	}

	if processTree.DisplayOptions.ShowCpuPercent {
		cpuPercent = fmt.Sprintf("(c:%.2f%%)", processTree.Nodes[pidIndex].CPUPercent)
		processTree.colorizeField("cpu", &cpuPercent, pidIndex)
		builder.WriteString(cpuPercent)
		builder.WriteString(" ")
	}

	if processTree.DisplayOptions.ShowMemoryUsage {
		memoryUsage = fmt.Sprintf("(m:%s)", util.ByteConverter(processTree.Nodes[pidIndex].MemoryInfo.RSS))
		processTree.colorizeField("memory", &memoryUsage, pidIndex)
		builder.WriteString(memoryUsage)
		builder.WriteString(" ")
	}

	if processTree.DisplayOptions.ShowNumThreads {
		// Always show thread count, even when showing compact format
		threads = fmt.Sprintf("(t:%d)", processTree.Nodes[pidIndex].NumThreads)
		processTree.colorizeField("threads", &threads, pidIndex)
		builder.WriteString(threads)
		builder.WriteString(" ")
	}

	if processTree.DisplayOptions.ShowUIDTransitions && processTree.Nodes[pidIndex].HasUIDTransition {
		// Add UID transition notation {parentUID→currentUID}
		if len(processTree.Nodes[pidIndex].UIDs) > 0 {
			ownerTransition = fmt.Sprintf("(%d→%d)", processTree.Nodes[pidIndex].ParentUID, processTree.Nodes[pidIndex].UIDs[0])
		}
	} else if processTree.DisplayOptions.ShowUserTransitions && processTree.Nodes[pidIndex].HasUIDTransition {
		// Add user transition notation {parentUser→currentUser}
		if processTree.Nodes[pidIndex].ParentUsername != "" {
			ownerTransition = fmt.Sprintf("(%s→%s)", processTree.Nodes[pidIndex].ParentUsername, processTree.Nodes[pidIndex].Username)
		}
	}

	if ownerTransition != "" {
		processTree.colorizeField("ownerTransition", &ownerTransition, pidIndex)
		builder.WriteString(ownerTransition)
		builder.WriteString(" ")
	}

	// Get the command - use full path when compact mode is disabled
	commandStr = processTree.Nodes[pidIndex].Command

	// Determine if this is a thread
	// isThread := processTree.Nodes[pidIndex].NumThreads > 0 && processTree.Nodes[pidIndex].PPID > 0

	// In compact mode, format the command with count for the first process in a group
	if processTree.DisplayOptions.CompactMode {
		// Get the count of identical processes
		count, groupPIDs := processTree.GetProcessCount(pidIndex)

		// If there are multiple identical processes, format with count
		if count > 1 {
			// Format in Linux pstree style
			compactStr = processTree.FormatCompactOutput(commandStr, count, groupPIDs)

			if compactStr != "" {
				// Create the connector string
				connector = "───"

				// Colorize the connector and compact format indicator in green if color support is available
				if processTree.DisplayOptions.ColorSupport {
					// The util.ColorGreen function modifies the string in place via pointer
					processTree.colorizeField("connector", &connector, pidIndex)
					// builder.WriteString(connector)
					processTree.colorizeField("compactStr", &compactStr, pidIndex)
					// builder.WriteString(compactStr)
				}

				// In Linux pstree, the format is just the count and brackets, not repeating the command
				commandStr = fmt.Sprintf("%s%s%s", commandStr, connector, compactStr)
			}
		}
	}

	// For threads in non-compact mode, wrap the command in curly braces
	// if !processTree.DisplayOptions.CompactMode && isThread && processTree.Nodes[pidIndex].NumThreads == 0 {
	// 	// This is likely a thread of a process
	// 	commandStr = fmt.Sprintf("{%s}", commandStr)
	// }

	processTree.colorizeField("command", &commandStr, pidIndex)
	builder.WriteString(commandStr)
	builder.WriteString(" ")

	if processTree.DisplayOptions.ShowArguments {
		if len(processTree.Nodes[pidIndex].Args) > 0 {
			// psutil.Process sometimes prepends the first argument with the name of the binary,
			// e.g., /opt/brave.com/brave/brave /opt/brave.com/brave/brave --arg1 --arg2
			// we want to strip the binary name out
			if strings.HasPrefix(processTree.Nodes[pidIndex].Args[0], processTree.Nodes[pidIndex].Command) {
				processTree.Nodes[pidIndex].Args[0] = processTree.Nodes[pidIndex].Args[0][len(processTree.Nodes[pidIndex].Command)+1:]
			} else if processTree.Nodes[pidIndex].Args[0] == filepath.Base(processTree.Nodes[pidIndex].Command) {
				// psutil.Process sometimes calls the argument filepath.Base(command),
				// e.g., Command is /usr/bin/cat and Args[0] is cat

				if len(processTree.Nodes[pidIndex].Args) == 1 {
					// If this is the only arg, Args is an empty slice
					processTree.Nodes[pidIndex].Args = []string{}
					// If there is more than one arg, remove it from the slice
				} else if len(processTree.Nodes[pidIndex].Args) > 1 {
					processTree.Nodes[pidIndex].Args = processTree.Nodes[pidIndex].Args[1:]
				}
			}
		}
		args = strings.Join(processTree.Nodes[pidIndex].Args, " ")
		processTree.colorizeField("args", &args, pidIndex)
		builder.WriteString(args)
		builder.WriteString(" ")
	}

	return builder.String()
}

// buildNewHead constructs a new head string for child processes based on the current process's position.
//
// Parameters:
//   - head: The accumulated prefix string from parent levels
//   - pidIndex: Index of the current process in the Nodes array
//
// Returns:
//   - A string to be used as the head for child processes, including appropriate vertical bars
//     or spaces based on whether the current process has visible siblings.
func (processTree *ProcessTree) buildNewHead(head string, pidIndex int) string {
	newHead := fmt.Sprintf("%s%s ",
		head,
		func() string {
			if head == "" {
				return ""
			}
			// In compact mode, we need to check if any visible siblings exist
			if processTree.DisplayOptions.CompactMode {
				sibling := processTree.Nodes[pidIndex].Sister
				for sibling != -1 {
					if !processTree.ShouldSkipProcess(sibling) {
						return processTree.TreeChars.Bar // Only add vertical bar if there's a visible sibling
					}
					sibling = processTree.Nodes[sibling].Sister
				}
				return " " // No visible siblings
			} else {
				// In normal mode, just check if there's a sibling
				if processTree.Nodes[pidIndex].Sister != -1 {
					return processTree.TreeChars.Bar
				}
				return " "
			}
		}(),
	)
	return newHead
}

// PrintThreads displays the threads of a process in a tree-like structure.
// It formats each thread with its thread ID and PGID. This only works on
// Linux because macOS does not provide thread IDs.
//
// Parameters:
//   - pidIndex: Index of the parent process whose threads to display
//   - head: The accumulated prefix string from parent levels
func (processTree *ProcessTree) PrintThreads(pidIndex int, head string) {
	if len(processTree.Nodes[pidIndex].Threads) == 0 {
		return
	}

	processTree.Logger.Debug(fmt.Sprintf("Printing %d threads for process %d", len(processTree.Nodes[pidIndex].Threads), processTree.Nodes[pidIndex].PID))

	// Get the thread head with proper spacing
	threadHead := processTree.buildThreadHead(head)

	for i, thread := range processTree.Nodes[pidIndex].Threads {
		var (
			line       string
			threadLine strings.Builder
			prefix     string
			threadInfo string
		)

		// Always use T-connector (├) for threads except for the last thread when there are no child processes
		// This ensures that when a thread is followed by a process, the thread uses the correct connector
		isLastThread := i == len(processTree.Nodes[pidIndex].Threads)-1
		hasChildProcess := processTree.Nodes[pidIndex].Child != -1

		// Create thread line prefix with appropriate branch characters
		if isLastThread && !hasChildProcess {
			// Last thread with no child processes uses └──── style connector (L shape)
			prefix = threadHead + processTree.TreeChars.BarL + processTree.TreeChars.EG + strings.Repeat(processTree.TreeChars.S2, 1) + processTree.TreeChars.NPGL
		} else {
			// Other threads or last thread with child processes use ├──── style connector (T shape)
			prefix = threadHead + processTree.TreeChars.BarC + processTree.TreeChars.EG + strings.Repeat(processTree.TreeChars.S2, 1) + processTree.TreeChars.NPGL
		}

		// Format thread name with curly braces like {processname}
		threadName := fmt.Sprintf(" {%s}", filepath.Base(thread.Command))

		// Format thread ID and PGID as (ThreadID, PGID)
		threadInfo = fmt.Sprintf(" (%d,%d)", thread.TID, thread.PGID)

		// Build the complete thread line
		threadLine.WriteString(prefix)
		threadLine.WriteString(threadName)
		threadLine.WriteString(threadInfo)

		line = threadLine.String()

		// Apply color if supported
		if processTree.DisplayOptions.ColorSupport {
			if processTree.DisplayOptions.ColorizeOutput {
				processTree.colorizeField("prefix", &prefix, pidIndex)
				processTree.colorizeField("command", &threadName, pidIndex)
				processTree.colorizeField("pidPgid", &threadInfo, pidIndex)
				line = prefix + threadName + threadInfo
			}
		}

		// Handle terminal width and coloring
		if !term.IsTerminal(int(os.Stdout.Fd())) {
			line = processTree.stripANSI(line)
			if len(line) > processTree.DisplayOptions.ScreenWidth && !processTree.DisplayOptions.WideDisplay {
				line = processTree.truncatePlain(line)
			}
		} else if !processTree.DisplayOptions.WideDisplay && len(line) > processTree.DisplayOptions.ScreenWidth {
			if processTree.DisplayOptions.RainbowOutput {
				line = processTree.truncateANSI(gorainbow.Rainbow(line))
			} else {
				line = processTree.truncateANSI(line)
			}
		} else if processTree.DisplayOptions.RainbowOutput {
			line = gorainbow.Rainbow(line)
		}

		// Print the thread line
		fmt.Fprintln(os.Stdout, line)
	}
}

// buildThreadHead constructs a head string specifically for thread display.
// It ensures the correct spacing and vertical bars for thread hierarchy.
//
// Parameters:
//   - head: The accumulated prefix string from parent levels
//
// Returns:
//   - A string to be used as the head for thread display
func (processTree *ProcessTree) buildThreadHead(head string) string {
	// Remove the trailing space from the head if it exists
	head = strings.TrimSuffix(head, " ")

	// For thread display, we need to ensure the correct spacing
	// The format should be "│ " (vertical bar followed by space)
	if len(head) > 0 {
		// Replace the last space with a vertical bar + space
		return head + " "
	}

	return head
}
