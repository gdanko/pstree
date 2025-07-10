package tree

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
// If any process in a potential group has threads and thread display is enabled
// (HideThreads is false), that group of processes will not be compacted.
//
// This function should be called before printing the tree when compact mode is enabled.
//
// Returns:
//   - error: nil if successful, or an error if initialization fails
func (processTree *ProcessTree) InitCompactMode() error {
	processTree.Logger.Debug("Entering processTree.InitCompactMode()")
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
			// Check if any process in the group has threads
			hasThreads := false
			if len(processTree.Nodes[pidIndex].Threads) > 0 {
				hasThreads = true
			}
			for _, idx := range group.Indices {
				if len(processTree.Nodes[idx].Threads) > 0 {
					hasThreads = true
					break
				}
			}

			// Only add to group if either:
			// 1. No threads in the group and current process, or
			// 2. Threads are hidden
			if !hasThreads || processTree.DisplayOptions.HideThreads {
				// Add to existing group
				group.Count++
				group.Indices = append(group.Indices, pidIndex)

				// Mark this process to be skipped during printing
				processTree.SkipProcesses[pidIndex] = true
			}
		}

		// Update the group in the map
		processTree.ProcessGroups[parentPID][compositeKey][processOwner] = group
	}
	return nil
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
func (processTree *ProcessTree) GetProcessCount(pidIndex int) (int, []int32, bool) {
	var (
		args            []string
		cmd             string
		compositeKey    string
		groupHasThreads bool
		groupPIDs       []int32
		parentPID       int32
		processOwner    string
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
				groupProcess := processTree.Nodes[group.Indices[i]]
				if len(groupProcess.Threads) > 0 {
					groupHasThreads = true
				}
				groupPIDs = append(groupPIDs, processTree.Nodes[group.Indices[i]].PID)
			}
			return group.Count, groupPIDs, groupHasThreads
		}
	}

	// No group or not the first process in the group
	return 1, []int32{}, false
}

//------------------------------------------------------------------------------
// OUTPUT FORMATTING
//------------------------------------------------------------------------------

// FormatCompactOutput formats the compacted processes.
//
// This function creates a formatted string representation of a process group
// in the style of Linux pstree. For regular processes, the format is "N*[command]",
// where N is the count.
//
// Parameters:
//   - command: The command name to format
//   - count: Number of identical processes
//   - groupPIDs: The list of PIDs for this process group
//
// Returns:
//   - Formatted string for display
func (processTree *ProcessTree) FormatCompactOutput(command string, count int, groupPIDs []int32) string {
	if count <= 1 {
		return command
	}
	if processTree.DisplayOptions.ShowPIDs {
		return fmt.Sprintf("───%d*[%s] (%s)", count, filepath.Base(command), strings.Join(processTree.PIDsToString(groupPIDs), ","))
	} else {
		return fmt.Sprintf("───%d*[%s]", count, filepath.Base(command))
	}
}

// FormatCompactedThreads formats the compacted threads.
//
// This function creates a formatted string representation of a process group's
// threads in the style of Linux pstree. For regular processes, the format
// is "{command}" for 1 thread and "N*[{command}]" for > 1 threads,
// where N is the count.
//
// Parameters:
//   - command: The command name to format
//   - compactableThreads: Number of threads that can be compacted
//
// Returns:
//   - Formatted string for display
func (processTree *ProcessTree) FormatCompactedThreads(command string, compactableThreads int32) string {
	var (
		compactedThread string
	)

	if compactableThreads == 1 {
		compactedThread = fmt.Sprintf("───{%s}", filepath.Base(command))
	} else if compactableThreads > 1 {
		compactedThread = fmt.Sprintf("───%d*[{%s}]", compactableThreads, filepath.Base(command))
	}
	return compactedThread
}

// PIDsToString converts a slice of process IDs to a slice of their string representations.
//
// This function is used in compact mode when displaying process groups with PIDs.
// Each PID is converted to a string representation that can be joined together
// for display in the process tree.
//
// Parameters:
//   - pids: Slice of int32 process IDs to convert
//
// Returns:
//   - []string: Slice of string representations of the PIDs
func (processTree *ProcessTree) PIDsToString(pids []int32) []string {
	pidStrings := make([]string, len(pids))
	for i, pid := range pids {
		pidStrings[i] = fmt.Sprintf("%d", pid)
	}
	return pidStrings
}
