// Package pstree provides functionality for building and displaying process trees.
//
// This file contains the core tree building and display logic, including:
// - Process tree construction and hierarchy management
// - Process marking and filtering
// - Tree visualization with various display options
// - Support for different graphical styles (ASCII, UTF-8, etc.)
//
// NOTE: This file has been identified as being too large and complex.
// Functions like BuildTree, MarkProcesses, and PrintTree could benefit from
// being broken down into smaller, more focused functions.
package tree

import (
	"fmt"
	"log/slog"
	"os"
	"runtime"

	"github.com/gdanko/pstree/pkg/color"
)

//------------------------------------------------------------------------------
// INITIALIZATION AND TREE CONSTRUCTION
//------------------------------------------------------------------------------
// Functions in this section handle the creation of the process tree structure
// and establishing the hierarchical relationships between processes.

// NewProcessTree creates a new process tree from a slice of processes.
//
// This function initializes a ProcessTree structure, populates it with ProcessNode objects
// created from the provided processes, and builds the hierarchical relationships between them.
// The resulting tree can be used for traversal, filtering, and visualization of the process hierarchy.
//
// Parameters:
//   - logger: Logger instance for debug and informational messages
//   - processes: Slice of Process objects containing the process information
//   - displayOptions: Configuration options controlling how the tree will be displayed
//
// Returns:
//   - A pointer to the newly created ProcessTree
func NewProcessTree(debugLevel int, logger *slog.Logger, processes []Process, displayOptions DisplayOptions) (processTree *ProcessTree) {
	var (
		idx  int
		proc Process
	)

	processTree = &ProcessTree{
		AtDepth:        0,
		DebugLevel:     debugLevel,
		DisplayOptions: displayOptions,
		IndexToPidMap:  make(map[int]int32, len(processes)),
		Logger:         logger,
		Nodes:          make([]Process, 0, len(processes)),
		PidToIndexMap:  make(map[int32]int, len(processes)),
		ProcessGroups:  make(map[int32]map[string]map[string]ProcessGroup),
		RootPID:        displayOptions.RootPID,
		SkipProcesses:  make(map[int]bool),
	}

	// Create nodes
	for _, proc = range processes {
		// Add to tree
		idx = len(processTree.Nodes)
		processTree.Nodes = append(processTree.Nodes, proc)
		processTree.PidToIndexMap[proc.PID] = idx
		processTree.IndexToPidMap[idx] = proc.PID
	}

	// If PID is not set via --pid, we want to look for PID 1...
	// https://github.com/FredHucht/pstree/blob/main/pstree.c#L558-L587

	// Define the tree characters
	if processTree.DisplayOptions.IBM850Graphics {
		processTree.TreeChars = TreeStyles["pc850"]
	} else if processTree.DisplayOptions.UTF8Graphics {
		processTree.TreeChars = TreeStyles["utf8"]
	} else if processTree.DisplayOptions.VT100Graphics {
		processTree.TreeChars = TreeStyles["vt100"]
	} else {
		processTree.TreeChars = TreeStyles["ascii"]
	}

	// Initialize the color scheme
	// if 8 bit color (8-16) is detected, we will use the ansi8 color scheme
	if processTree.DisplayOptions.ColorCount >= 8 && processTree.DisplayOptions.ColorCount <= 16 {
		processTree.ColorScheme = color.ColorSchemes["ansi8"]
	} else if processTree.DisplayOptions.ColorCount >= 256 {
		if processTree.DisplayOptions.ColorScheme != "" {
			processTree.ColorScheme = color.ColorSchemes[processTree.DisplayOptions.ColorScheme]
		} else {
			switch runtime.GOOS {
			case "windows":
				if os.Getenv("PSModulePath") != "" {
					processTree.ColorScheme = color.ColorSchemes["powershell"]
				} else {
					processTree.ColorScheme = color.ColorSchemes["windows10"]
				}
			case "linux":
				processTree.ColorScheme = color.ColorSchemes["linux"]
			case "darwin":
				processTree.ColorScheme = color.ColorSchemes["darwin"]
			default:
				processTree.ColorScheme = color.ColorSchemes["xterm"]
			}
		}
	}

	// Initialize colorizer
	if processTree.DisplayOptions.ColorizeOutput || processTree.DisplayOptions.ColorAttr != "" {
		if processTree.DisplayOptions.ColorCount >= 8 && processTree.DisplayOptions.ColorCount <= 16 {
			processTree.Colorizer = color.Colorizers["8color"]
		} else if processTree.DisplayOptions.ColorCount >= 256 {
			processTree.Colorizer = color.Colorizers["256color"]
		}
	}

	// Build the tree
	processTree.BuildTree()

	// Mark UID transitions
	processTree.MarkUIDTransitions()

	return processTree
}

// BuildTree constructs the hierarchical relationships between processes in the tree.
//
// This method establishes the parent-child relationships between processes by connecting
// each process to its parent based on the PPID (Parent Process ID). It creates a tree structure
// where each node can have one parent and multiple children, with siblings linked in a list.
// The resulting tree structure enables efficient traversal for operations like marking,
// filtering, and visualization.
//
// The method handles cases where a parent process might not exist in the tree (e.g., if the
// parent was not included in the original process list or if it's the process itself).
//
// Refactoring opportunity: This function could be broken down into smaller functions:
// - initializeNodes: Initialize all nodes with default values
// - buildParentChildRelationships: Establish the parent-child connections
func (processTree *ProcessTree) BuildTree() {
	// https://github.com/FredHucht/pstree/blob/main/pstree.c#L635-L652
	processTree.Logger.Debug("Entering processTree.BuildTree()")

	// Initialize all nodes with -1 for Child, Parent, and Sister fields
	for i := range processTree.Nodes {
		processTree.Nodes[i].Child = -1
		processTree.Nodes[i].Parent = -1
		processTree.Nodes[i].Sister = -1
		processTree.Nodes[i].Print = false
	}

	// Build the tree using the PidToIndexMap for O(1) lookups
	for pidIndex := range processTree.Nodes {
		ppid := processTree.Nodes[pidIndex].PPID

		// Look up parent index directly from the map
		ppidIndex, exists := processTree.PidToIndexMap[ppid]

		// Skip if parent doesn't exist or is the process itself
		if !exists || ppidIndex == pidIndex {
			continue
		}

		// Set parent relationship
		processTree.Nodes[pidIndex].Parent = ppidIndex

		// Add as child
		if processTree.Nodes[ppidIndex].Child == -1 {
			// First child
			processTree.Nodes[ppidIndex].Child = pidIndex
		} else {
			// Find the last sibling
			sisterIndex := processTree.Nodes[ppidIndex].Child
			for processTree.Nodes[sisterIndex].Sister != -1 {
				sisterIndex = processTree.Nodes[sisterIndex].Sister
			}
			// Add as sister to the last child
			processTree.Nodes[sisterIndex].Sister = pidIndex
		}
	}
}

//------------------------------------------------------------------------------
// DEBUGGING UTILITIES
//------------------------------------------------------------------------------
// Functions in this section provide debugging capabilities for the process tree.

// ShowPrintable logs all processes that are marked for display.
//
// This method is primarily used for debugging purposes. It iterates through all nodes
// in the process tree and logs detailed information about each process that has been
// marked for display (Print=true). The output includes all fields of the Process struct.
//
// The method uses pretty-printing to format the output in a readable way.
func (processTree *ProcessTree) ShowPrintable() {
	for i := range processTree.Nodes {
		if processTree.Nodes[i].Print {
			fmt.Printf("PID %d is printable\n", processTree.IndexToPidMap[i])
		}
	}
}

//------------------------------------------------------------------------------
// UTILITY FUNCTIONS
//------------------------------------------------------------------------------
// General utility functions used throughout the process tree implementation.

// getPidIndex finds the index of a process with the specified PID in the processes slice.
//
// Parameters:
//   - pid: The PID to search for
//
// Returns:
//   - The index of the process with the specified PID, or -1 if not found
func (processTree *ProcessTree) getPidIndex(pid int32) int {
	for pidIndex := range processTree.Nodes {
		if processTree.Nodes[pidIndex].PID == pid {
			return pidIndex
		}
	}
	return -1
}
