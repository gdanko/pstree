package tree

import (
	"fmt"
	"strings"
)

//------------------------------------------------------------------------------
// PROCESS MARKING AND FILTERING
//------------------------------------------------------------------------------
// Functions in this section handle the identification and marking of processes
// that should be included in the display, based on various filtering criteria.

// MarkProcesses marks processes that should be displayed based on filtering criteria.
// It applies various filters such as process name pattern matching, username filtering,
// root process exclusion, and PID filtering to determine which processes should be displayed.
//
// Refactoring opportunity: This function could be broken down into smaller functions:
// - applyUsernameFilter: Mark processes matching username criteria
// - applyRootPIDFilter: Mark processes based on root PID
// - applyCommandFilter: Mark processes matching command pattern
// - applyRootExclusionFilter: Apply root user exclusion filter
func (processTree *ProcessTree) MarkProcesses() {
	// https://github.com/FredHucht/pstree/blob/main/pstree.c#L662-L684
	processTree.Logger.Debug("Entering processTree.MarkProcesses()")
	var (
		myPid    int32
		process  Process
		pidIndex int
		showAll  bool
		username string
	)

	if processTree.DisplayOptions.Contains == "" && len(processTree.DisplayOptions.Usernames) == 0 && !processTree.DisplayOptions.ExcludeRoot && processTree.DisplayOptions.RootPID < 1 {
		showAll = true
	}

	for pidIndex = range processTree.Nodes {
		if showAll {
			processTree.Nodes[pidIndex].Print = true
		} else {
			process = processTree.Nodes[pidIndex]
			if len(processTree.DisplayOptions.Usernames) > 0 {
				for _, username = range processTree.DisplayOptions.Usernames {
					if process.Username == username {
						processTree.markParents(pidIndex)
						processTree.markChildren(pidIndex)
					}
				}
			} else if processTree.Nodes[pidIndex].PID == processTree.DisplayOptions.RootPID {
				// processTree.Logger.Debug("--pid == processTree.DisplayOptions.RootPID")
				if (processTree.DisplayOptions.ExcludeRoot && processTree.Nodes[pidIndex].Username != "root") || (!processTree.DisplayOptions.ExcludeRoot) {
					// processTree.Logger.Debug("(processTree.DisplayOptions.ExcludeRoot && processTree.Nodes[pidIndex].Username != root) || !processTree.DisplayOptions.ExcludeRoot")
					processTree.markParents(pidIndex)
					processTree.markChildren(pidIndex)
				}
			} else if processTree.DisplayOptions.Contains != "" && strings.Contains(process.Command, processTree.DisplayOptions.Contains) && (process.PID != myPid) {
				// processTree.Logger.Debug("processTree.DisplayOptions.Contains is set && process.Command contains processTree.DisplayOptions.Contains && process.PID != myPid")
				if (processTree.DisplayOptions.ExcludeRoot && process.Username != "root") || (!processTree.DisplayOptions.ExcludeRoot) {
					// processTree.Logger.Debug("(processTree.DisplayOptions.ExcludeRoot && process.Username != root) || !processTree.DisplayOptions.ExcludeRoot")
					processTree.markParents(pidIndex)
					processTree.markChildren(pidIndex)
				}
			} else if processTree.DisplayOptions.Contains != "" && !strings.Contains(process.Command, processTree.DisplayOptions.Contains) && (process.PID != myPid) {
				// processTree.Logger.Debug("processTree.DisplayOptions.Contains is set && process.Command does not contain processTree.DisplayOptions.Contains && process.PID != myPid")
			} else if processTree.DisplayOptions.ExcludeRoot && process.Username != "root" {
				// processTree.Logger.Debug("processTree.DisplayOptions.ExcludeRoot && process.Username != root")
				processTree.markParents(pidIndex)
				processTree.markChildren(pidIndex)
			}
		}
	}
}

// MarkThreads marks threads that should be displayed based on filtering criteria.
// It ensures that threads are properly associated with their parent processes and
// marked for display when appropriate.
func (processTree *ProcessTree) MarkThreads() {
	processTree.Logger.Debug("Entering processTree.MarkThreads()")

	// If threads are hidden, no need to mark them
	if processTree.DisplayOptions.HideThreads {
		return
	}

	// Iterate through all processes
	for pidIndex := range processTree.Nodes {
		// Only mark threads for processes that are marked for display
		if processTree.Nodes[pidIndex].Print && len(processTree.Nodes[pidIndex].Threads) > 0 {
			processTree.Logger.Debug(fmt.Sprintf("Marking %d threads for process %d",
				len(processTree.Nodes[pidIndex].Threads), processTree.Nodes[pidIndex].PID))
		}
	}
}

// DropUnmarked removes processes that are not marked for display from the process tree.
// It modifies the process tree structure to maintain proper parent-child relationships
// while excluding processes that should not be displayed.
//
// Refactoring opportunity: This function could be split into:
// - dropUnmarkedChildren: Remove unmarked children from each node
// - dropUnmarkedSiblings: Remove unmarked siblings from each node
func (processTree *ProcessTree) DropUnmarked() {
	// https://github.com/FredHucht/pstree/blob/main/pstree.c#L706-L717
	processTree.Logger.Debug("Entering processTree.DropUnmarked()")
	var (
		childPidIndex  int
		pidIndex       int
		sisterPidIndex int
	)

	for pidIndex = range processTree.Nodes {
		if processTree.Nodes[pidIndex].Print {
			// Drop children that won't print
			childPidIndex = processTree.Nodes[pidIndex].Child
			for childPidIndex != -1 && !processTree.Nodes[childPidIndex].Print {
				childPidIndex = processTree.Nodes[childPidIndex].Sister
			}
			processTree.Nodes[pidIndex].Child = childPidIndex

			// Drop sisters that won't print
			sisterPidIndex = processTree.Nodes[pidIndex].Sister
			for sisterPidIndex != -1 && !processTree.Nodes[sisterPidIndex].Print {
				sisterPidIndex = processTree.Nodes[sisterPidIndex].Sister
			}
			processTree.Nodes[pidIndex].Sister = sisterPidIndex
		}
	}
}

//------------------------------------------------------------------------------
// PROCESS ATTRIBUTE MARKING
//------------------------------------------------------------------------------
// Functions in this section handle marking special attributes of processes,
// such as UID transitions and highlighting the current process.

// MarkUIDTransitions identifies and marks processes where the user ID changes from the parent process.
// This function compares the UIDs of each process with its parent and sets HasUIDTransition=true
// when a transition is detected. It also stores the parent UID for display purposes.
//
// Parameters:
//   - logger: Logger instance for debug information
//   - processes: Pointer to a slice of Process structs
func (processTree *ProcessTree) MarkUIDTransitions() {
	var (
		pidIndex  int
		ppidIndex int
	)

	processTree.Logger.Debug("Marking UID transitions between processes - START")

	for pidIndex = range processTree.Nodes {
		// Skip the root process (which has no parent)
		if processTree.Nodes[pidIndex].Parent == -1 {
			continue
		}

		// Get parent index
		ppidIndex = processTree.Nodes[pidIndex].Parent

		// Compare UIDs between process and its parent
		if len(processTree.Nodes[pidIndex].UIDs) > 0 && len(processTree.Nodes[ppidIndex].UIDs) > 0 {
			// Store parent UID regardless of transition
			processTree.Nodes[pidIndex].ParentUID = processTree.Nodes[ppidIndex].UIDs[0]
			processTree.Nodes[pidIndex].ParentUsername = processTree.Nodes[ppidIndex].Username

			// Compare the first UID (effective UID
			if processTree.Nodes[pidIndex].UIDs[0] != processTree.Nodes[ppidIndex].UIDs[0] {
				if processTree.DebugLevel > 1 {
					processTree.Logger.Debug(fmt.Sprintf("UID transition detected: Process %d (UID %d) has different UID from parent %d (UID %d)",
						processTree.Nodes[pidIndex].PID, processTree.Nodes[pidIndex].UIDs[0],
						processTree.Nodes[ppidIndex].PID, processTree.Nodes[ppidIndex].UIDs[0]))
				}
				processTree.Nodes[pidIndex].HasUIDTransition = true
			}
		} else {
			// Fallback to username comparison if UIDs are not available
			if processTree.Nodes[pidIndex].Username != processTree.Nodes[ppidIndex].Username {
				if processTree.DebugLevel > 1 {
					processTree.Logger.Debug(fmt.Sprintf("Username transition detected: Process %d (%s) has different username from parent %d (%s)",
						processTree.Nodes[pidIndex].PID, processTree.Nodes[pidIndex].Username,
						processTree.Nodes[ppidIndex].PID, processTree.Nodes[ppidIndex].Username))
				}
				processTree.Nodes[pidIndex].HasUIDTransition = true
			}
		}
	}
}

// MarkCurrentAndAncestors marks the current process and all its ancestors.
// This function identifies the current process by its PID and marks it and all
// its ancestors with IsCurrentOrAncestor=true for highlighting in the display.
//
// Parameters:
//   - currentPid: The PID of the current process to highlight
func (processTree *ProcessTree) MarkCurrentAndAncestors(currentPid int32) {
	if currentPid <= 0 {
		return
	}

	processTree.Logger.Debug(fmt.Sprintf("Marking current process %d and its ancestors", currentPid))
	var (
		currentIndex int
		parentIndex  int
	)

	// Find the current process index
	currentIndex = processTree.getPidIndex(currentPid)
	if currentIndex == -1 {
		processTree.Logger.Debug(fmt.Sprintf("Current process %d not found", currentPid))
		return
	}

	// Mark the current process
	processTree.Nodes[currentIndex].IsCurrentOrAncestor = true

	// Mark all ancestors

	parentIndex = processTree.Nodes[currentIndex].Parent
	for parentIndex != -1 {
		processTree.Logger.Debug(fmt.Sprintf("Marking pid %d as ancestor of current process", processTree.IndexToPidMap[parentIndex]))
		processTree.Nodes[parentIndex].IsCurrentOrAncestor = true
		parentIndex = processTree.Nodes[parentIndex].Parent
	}
}

//------------------------------------------------------------------------------
// TREE TRAVERSAL HELPERS
//------------------------------------------------------------------------------
// Helper functions for traversing the process tree in different directions.

// markParents marks all parent processes of a given process as printable.
// This function recursively traverses up the process tree, marking each parent
// process with Print=true until it reaches the root process (or a process with no parent).
//
// Parameters:
//   - pidIndex: Index of the process whose parents should be marked
func (processTree *ProcessTree) markParents(pidIndex int) {
	processTree.Logger.Debug(fmt.Sprintf("Entering markParents(), pidIndex=%d, pid=%d", pidIndex, processTree.IndexToPidMap[pidIndex]))
	var (
		ppidIndex int
	)

	ppidIndex = processTree.Nodes[pidIndex].Parent
	processTree.Logger.Debug(fmt.Sprintf("Marking %d as a parent of %d", processTree.IndexToPidMap[ppidIndex], processTree.IndexToPidMap[pidIndex]))
	for ppidIndex != -1 {
		processTree.Logger.Debug(fmt.Sprintf("Marking PID %d's Print attribute as true", processTree.IndexToPidMap[ppidIndex]))
		processTree.Nodes[ppidIndex].Print = true
		ppidIndex = processTree.Nodes[ppidIndex].Parent
	}
}

// markChildren marks a process and all its child processes as printable.
// This function recursively traverses down the process tree, marking each child
// process with Print=true, and continues with any sibling processes.
//
// Parameters:
//   - pidIndex: Index of the process whose children should be marked
func (processTree *ProcessTree) markChildren(pidIndex int) {
	processTree.Logger.Debug(fmt.Sprintf("Entering markChildren(), pidIndex=%d, pid=%d", pidIndex, processTree.IndexToPidMap[pidIndex]))
	var (
		childPidIndex int
	)

	processTree.Logger.Debug(fmt.Sprintf("Marking PID %d's Print attribute as true", processTree.IndexToPidMap[pidIndex]))
	processTree.Nodes[pidIndex].Print = true
	childPidIndex = processTree.Nodes[pidIndex].Child
	for childPidIndex != -1 {
		processTree.markChildren(childPidIndex)
		childPidIndex = processTree.Nodes[childPidIndex].Sister
	}
}
