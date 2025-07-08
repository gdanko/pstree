package pstree

import (
	"fmt"
	"path/filepath"
	"strings"
)

//------------------------------------------------------------------------------
// INITIALIZATION
//------------------------------------------------------------------------------

// InitCompactMode initializes the compact mode by identifying identical processes.
//
// This function analyzes the provided processes slice and groups processes that have
// identical commands and arguments under the same parent. It populates the
// processTree.ProcessGroups map with information about these groups and marks
// processes that should be skipped during printing (all except the first process
// in each group).
//
// This function should be called before printing the tree when compact mode is enabled.
func (processTree *ProcessTree) InitCompactMode() {
	var (
		args         []string
		cmd          string
		exists       bool
		group        ProcessGroup
		parentPID    int32
		pidIndex     int
		processOwner string
	)

	// Group processes with identical commands under the same parent
	for pidIndex = range processTree.Nodes {
		// Skip processes that are already part of a group
		if processTree.SkipProcesses[pidIndex] {
			continue
		}

		// Get parent PID
		parentPID = processTree.Nodes[pidIndex].PPID

		// Get the owner of the process
		processOwner = processTree.Nodes[pidIndex].Username

		// Get the command and arguments to create a composite key
		// This ensures processes are only grouped if both command AND arguments match exactly
		cmd = processTree.Nodes[pidIndex].Command
		args = processTree.Nodes[pidIndex].Args

		// Create a composite key with both command and arguments
		compositeKey := cmd
		if len(args) > 0 {
			compositeKey = fmt.Sprintf("%s %s", cmd, strings.Join(args, " "))
		}

		// Initialize map for this parent if needed
		if _, exists := processTree.ProcessGroups[parentPID]; !exists {
			// ProcessGroups map[int32]map[string]map[string]ProcessGroup
			processTree.ProcessGroups[parentPID] = make(map[string]map[string]ProcessGroup)
		}

		if _, exists = processTree.ProcessGroups[parentPID][compositeKey]; !exists {
			processTree.ProcessGroups[parentPID][compositeKey] = make(map[string]ProcessGroup)
		}

		// Use the composite key (command + args) for grouping
		// This ensures that processes are only grouped if both command AND arguments AND owner match exactly
		group, exists = processTree.ProcessGroups[parentPID][compositeKey][processOwner]
		if !exists {
			// Create a new group
			group = ProcessGroup{
				Count:      1,
				FirstIndex: pidIndex,
				FullPath:   cmd,
				Indices:    []int{pidIndex},
				Owner:      processTree.Nodes[pidIndex].Username,
			}
		} else {
			// Add to existing group
			group.Count++
			group.Indices = append(group.Indices, pidIndex)

			// Mark this process to be skipped during printing
			processTree.SkipProcesses[pidIndex] = true
		}

		// Update the group in the map
		processTree.ProcessGroups[parentPID][compositeKey][processOwner] = group
	}
	// Find the first example of groups of two identical processes
	// and print them for debugging purposes
	// for _, v := range processTree.ProcessGroups {
	// 	for _, v := range v {
	// 		if len(v.Indices) > 1 {
	// 			pretty.Println(v.Indices)
	// 			for pidIndex := range v.Indices {
	// 				pretty.Println(processTree.Nodes[v.Indices[pidIndex]])
	// 			}
	// 		}
	// 	}
	// }
}

//------------------------------------------------------------------------------
// PROCESS FILTERING
//------------------------------------------------------------------------------

// ShouldSkipProcess returns true if the process should be skipped during printing.
//
// In compact mode, only the first process of each identical group is displayed,
// with a count indicator. This function checks if a process has been marked to
// be skipped during the initialization phase.
//
// Parameters:
//   - pidIndex: Index of the process to check
//
// Returns:
//   - true if the process should be skipped, false otherwise
func (processTree *ProcessTree) ShouldSkipProcess(pidIndex int) bool {
	return processTree.SkipProcesses[pidIndex]
}

//------------------------------------------------------------------------------
// PROCESS GROUP INFORMATION
//------------------------------------------------------------------------------

// GetProcessCount returns the count of identical processes for the given process.
//
// For processes that are the first in their group, this returns the total number
// of identical processes in that group. For processes that are not the first in
// their group, or are not part of a group, this returns 1.
//
// Parameters:
//   - processes: Slice of Process structs
//   - processIndex: Index of the process to get the count for
//
// Returns:
//   - count: Number of identical processes in the group
//   - isThread: Whether the process group represents threads
func (processTree *ProcessTree) GetProcessCount(pidIndex int) (int, []int32) {
	var (
		args         []string
		cmd          string
		compositeKey string
		groupPIDs    []int32
		parentPID    int32
		processOwner string
	)

	// Get parent PID and command
	args = processTree.Nodes[pidIndex].Args
	cmd = processTree.Nodes[pidIndex].Command
	parentPID = processTree.Nodes[pidIndex].PPID
	processOwner = processTree.Nodes[pidIndex].Username

	// Create the same composite key used in InitCompactMode
	compositeKey = cmd
	if len(args) > 0 {
		compositeKey = cmd + " " + strings.Join(args, " ")
	}

	// Check if we have a group for this process
	if groups, exists := processTree.ProcessGroups[parentPID]; exists {
		// Look up by composite key (command + args)
		if group, exists := groups[compositeKey][processOwner]; exists && group.FirstIndex == pidIndex {
			// Find PIDs for each member of the group
			for i := range group.Indices {
				groupPIDs = append(groupPIDs, processTree.Nodes[group.Indices[i]].PID)
			}
			return group.Count, groupPIDs
		}
	}

	// No group or not the first process in the group
	return 1, []int32{}
}

//------------------------------------------------------------------------------
// OUTPUT FORMATTING
//------------------------------------------------------------------------------

// FormatCompactOutput formats the command with count for compact mode.
//
// This function creates a formatted string representation of a process group
// in the style of Linux pstree. For regular processes, the format is "N*[command]",
// and for threads, the format is "N*[{command}]", where N is the count.
//
// Parameters:
//   - command: The command name to format
//   - count: Number of identical processes/threads
//   - isThread: Whether this is a thread group
//   - hideThreads: Whether threads should be hidden
//
// Returns:
//   - Formatted string for display, or empty string if threads should be hidden
func (processTree *ProcessTree) FormatCompactOutput(command string, count int, groupPIDs []int32) string {
	if count <= 1 {
		return command
	}
	return fmt.Sprintf("%d*[%s] {%s}", count, filepath.Base(command), strings.Join(processTree.PIDsToString(groupPIDs), ", "))
}

func (processTree *ProcessTree) PIDsToString(pids []int32) []string {
	pidStrings := make([]string, len(pids))
	for i, pid := range pids {
		pidStrings[i] = fmt.Sprintf("%d", pid)
	}
	return pidStrings
}
