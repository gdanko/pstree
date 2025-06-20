package pstree

import (
	"fmt"
	"strings"
)

// ProcessGroup represents a group of identical processes
type ProcessGroup struct {
	FirstIndex int   // Index of the first process in the group
	Count      int   // Number of identical processes
	Indices    []int // Indices of all processes in the group
	IsThread   bool  // Whether this is a thread group
}

// processGroups stores information about groups of identical processes
// Key is the parent PID, value is a map of command -> ProcessGroup
var processGroups map[int32]map[string]ProcessGroup

// skipProcesses tracks which processes should be skipped during printing
var skipProcesses map[int]bool

// InitCompactMode initializes the compact mode by identifying identical processes
// This should be called before printing the tree when compact mode is enabled
func InitCompactMode(processes []Process) {
	// Initialize the maps
	processGroups = make(map[int32]map[string]ProcessGroup)
	skipProcesses = make(map[int]bool)

	// Group processes with identical commands under the same parent
	for i := range processes {
		// Skip processes that are already part of a group
		if skipProcesses[i] {
			continue
		}

		// Get parent PID
		parentPID := processes[i].PPID

		// Get the command and extract the base name (without path)
		cmd := processes[i].Command
		baseName := cmd

		// Extract just the base command name without the path
		if lastSlash := strings.LastIndex(cmd, "/"); lastSlash != -1 {
			baseName = cmd[lastSlash+1:]
		}

		// Special case: normalize gsleep to sleep for consistent grouping
		if baseName == "gsleep" {
			baseName = "sleep"
		}

		// Determine if this is a thread
		isThread := processes[i].NumThreads > 0 && parentPID > 0

		// Initialize map for this parent if needed
		if _, exists := processGroups[parentPID]; !exists {
			processGroups[parentPID] = make(map[string]ProcessGroup)
		}

		// Use the base command name as the key for grouping
		// This is how Linux pstree works - it groups processes with the same base name
		group, exists := processGroups[parentPID][baseName]
		if !exists {
			// Create a new group
			group = ProcessGroup{
				FirstIndex: i,
				Count:      1,
				Indices:    []int{i},
				IsThread:   isThread,
			}
		} else {
			// Add to existing group
			group.Count++
			group.Indices = append(group.Indices, i)

			// Mark this process to be skipped during printing
			skipProcesses[i] = true
		}

		// Update the group in the map
		processGroups[parentPID][baseName] = group
	}
}

// ShouldSkipProcess returns true if the process should be skipped during printing
func ShouldSkipProcess(processIndex int) bool {
	return skipProcesses[processIndex]
}

// GetProcessCount returns the count of identical processes for the given process
func GetProcessCount(processes []Process, processIndex int) (int, bool) {
	// Get parent PID and command
	parentPID := processes[processIndex].PPID
	cmd := processes[processIndex].Command

	// Extract the base command name
	baseName := cmd
	if lastSlash := strings.LastIndex(cmd, "/"); lastSlash != -1 {
		baseName = cmd[lastSlash+1:]
	}

	// Special case: normalize gsleep to sleep for consistent grouping
	if baseName == "gsleep" {
		baseName = "sleep"
	}

	// Check if we have a group for this process
	if groups, exists := processGroups[parentPID]; exists {
		// Look up by base name, not full path
		if group, exists := groups[baseName]; exists && group.FirstIndex == processIndex {
			return group.Count, group.IsThread
		}
	}

	// No group or not the first process in the group
	return 1, false
}

// FormatCompactOutput formats the command with count for compact mode
func FormatCompactOutput(command string, count int, isThread bool, hideThreads bool) string {
	if count <= 1 {
		return command
	}

	// Extract the base command name without path
	baseCmd := command
	// If command contains a path, extract just the last component
	if lastSlash := strings.LastIndex(command, "/"); lastSlash != -1 {
		baseCmd = command[lastSlash+1:]
	}

	if isThread {
		// Format for threads: N*[{command}]
		if hideThreads {
			return ""
		} else {
			return fmt.Sprintf("%d*[{%s}]", count, baseCmd)
		}
	} else {
		// Format for processes: N*[command]
		return fmt.Sprintf("%d*[%s]", count, baseCmd)
	}
}
