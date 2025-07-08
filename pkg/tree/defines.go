// Package pstree provides functionality for building and displaying process trees.
//
// This file contains the core data structures and type definitions used throughout the package.
// It defines the Process, DisplayOptions, ProcessTree, and TreeChars types that form the foundation
// of the process tree visualization system.
package tree

import (
	"log/slog"
	"regexp"

	"github.com/gdanko/pstree/pkg/color"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

//------------------------------------------------------------------------------
// CORE DATA STRUCTURES
//------------------------------------------------------------------------------

// Process represents a system process with all its attributes and relationships.
// It combines information gathered from both gopsutil and direct ps command calls
// to provide comprehensive details about the process.
type Process struct {
	// Process age in seconds since creation
	Age int64
	// Command line arguments
	Args []string
	// Index of the first child process in the process tree
	Child int
	// Pointer to a slice of child processes
	Children *[]Process
	// Command name (executable name)
	Command string
	// Network connections associated with this process
	Connections []net.ConnectionStat
	// CPU usage percentage
	CPUPercent float64
	// CPU time statistics
	CPUTimes *cpu.TimesStat
	// Process creation time as Unix timestamp
	CreateTime int64
	// Environment variables
	Environment []string
	// Group IDs associated with this process
	GIDs []uint32
	// Groups associated with this process
	Groups []uint32
	// Indicates if this process has a different UID from its parent
	HasUIDTransition bool
	// Indicates if this process is the current process or an ancestor
	IsCurrentOrAncestor bool
	// Memory usage information
	MemoryInfo *process.MemoryInfoStat
	// Memory usage as percentage of total system memory
	MemoryPercent float32
	// Number of file descriptors
	NumFDs int32
	// Number of threads
	NumThreads int32
	// Open files
	OpenFiles []process.OpenFilesStat
	// Index of the parent process in the process tree
	Parent int
	// Pointer to the parent process
	ParentProcess *Process
	// UID of the parent process
	ParentUID uint32
	// Username of the parent process
	ParentUsername string
	// Process group ID
	PGID int32
	// Process ID
	PID int32
	// Parent process ID
	PPID int32
	// Whether or not we plan to display this process
	Print bool
	// Index of the next sibling process in the process tree
	Sister int
	// Process status information
	Status []string
	// A map of threads for the process
	Threads []Thread
	// Thread ID (if this is a thread)
	TID int32
	// User IDs associated with this process
	UIDs []uint32
	// Username of the process owner
	Username string
}

type Thread struct {
	// Command line arguments
	Args []string
	// Process group ID
	PGID int32
	// PID
	PID int32
	// Parent PID
	PPID int32
	// Thread ID
	TID int32
	// Command name (executable name)
	Command string
	// CPU Times
	CPUTimes *cpu.TimesStat
}

//------------------------------------------------------------------------------
// DISPLAY CONFIGURATION
//------------------------------------------------------------------------------

// DisplayOptions controls how the process tree is displayed, including formatting,
// coloring, and which information is shown for each process.
type DisplayOptions struct {
	// Attribute to color by ("age", "cpu", or "mem")
	ColorAttr string
	// Number of colors to use in rainbow mode
	ColorCount int
	// Whether to colorize the output with predefined colors
	ColorizeOutput bool
	// The system color scheme to use
	ColorScheme string
	// Whether the terminal supports color output
	ColorSupport bool
	// Whether to compact identical processes in the tree
	CompactMode bool
	// String to search for in process names
	Contains string
	// Whether to exclude processes owned by root
	ExcludeRoot bool
	// Whether to hide threads in the output
	HideThreads bool
	// Whether to use IBM850 graphics characters for tree lines
	IBM850Graphics bool
	// Total installed system memory in bytes
	InstalledMemory uint64
	// Maximum depth of the tree to display (0 for unlimited)
	MaxDepth int
	// Sort the results by a number of fields
	OrderBy string
	// Whether to use rainbow colors for output
	RainbowOutput bool
	// Root process PID
	RootPID int32
	// Width of the terminal screen in characters
	ScreenWidth int
	// Whether to show command line arguments
	ShowArguments bool
	// Whether to show CPU usage percentage
	ShowCpuPercent bool
	// Whether to show memory usage
	ShowMemoryUsage bool
	// Whether to show thread count
	ShowNumThreads bool
	// Whether to show process owner
	ShowOwner bool
	// Whether to highlight process group leaders
	ShowPGLs bool
	// Whether to show process group IDs
	ShowPGIDs bool
	// Whether to show process IDs
	ShowPIDs bool
	// Whether to show parent process IDs
	ShowPPIDs bool
	// Whether to show process age
	ShowProcessAge bool
	// Whether to show UID transitions
	ShowUIDTransitions bool
	// Whether to show username transitions
	ShowUserTransitions bool
	// Whether to use UTF-8 graphics characters for tree lines
	UTF8Graphics bool
	// List of usernames to filter by
	Usernames []string
	// Whether to use VT100 graphics characters for tree lines
	VT100Graphics bool
	// Whether to display wide output (not truncated to screen width)
	WideDisplay bool
}

//------------------------------------------------------------------------------
// TREE STRUCTURE
//------------------------------------------------------------------------------

// ProcessTree handles the construction and display of the process tree.
// It maintains the tree structure and provides methods for building,
// manipulating, and displaying the process hierarchy.
type ProcessTree struct {
	// Logger for debug and informational messages
	Logger *slog.Logger
	// Current depth in the tree during traversal
	AtDepth int
	// Display options controlling how the tree is rendered
	DisplayOptions DisplayOptions
	// Array of process nodes in the tree
	Nodes []Process
	// Map from PID to index in the Nodes array for quick lookups
	PidToIndexMap map[int32]int
	// Map from index in the Nodes array to PID
	IndexToPidMap map[int]int32
	// PID of the root process for the tree
	RootPID int32
	// Tree characters for drawing the tree
	TreeChars TreeChars
	// Enable debugging
	DebugLevel int
	// Colorizer for applying colors to text
	Colorizer color.Colorizer
	// Color scheme for applying colors to text
	ColorScheme color.ColorScheme
	// Process groups for compact mode
	ProcessGroups map[int32]map[string]map[string]ProcessGroup
	// Map to track processes that should be skipped during printing
	SkipProcesses map[int]bool
}

//------------------------------------------------------------------------------
// TREE VISUALIZATION
//------------------------------------------------------------------------------

// TreeChars defines the characters used for drawing the tree.
// Different character sets are available for different terminal types and preferences.
type TreeChars struct {
	// Bar represents the vertical bar character (│) used for drawing process tree lines
	Bar string
	// BarC represents the T-junction character (├) used where a process branches off
	BarC string
	// BarL represents the L-junction character (└) used for the last child process in a branch
	BarL string
	// EG represents the End Graphics character sequence for terminating graphic mode
	EG string
	// Init represents the initialization sequence for the terminal
	Init string
	// NPGL represents the character sequence to initialize the graphic set for non-process group leaders
	NPGL string
	// P represents the horizontal line character (─) used for connecting processes
	P string
	// PGL represents the character sequence used to highlight process group leaders
	PGL string
	// S2 represents the secondary process horizontal line character (─) for alternative styling
	S2 string
	// SG represents the Start Graphics character sequence for entering graphic mode
	SG string
}

// TreeStyles defines different graphical styles for tree visualization.
// Each style uses a different set of characters for drawing the tree structure,
// allowing users to choose the style that works best with their terminal.
var TreeStyles = map[string]TreeChars{
	// https://github.com/FredHucht/pstree/blob/main/pstree.c#L192-L207
	"ascii": {
		Bar:  "|",  // B
		BarC: "|",  // C
		BarL: "\\", // L
		EG:   "",   // eg
		Init: "",   // init
		NPGL: "-",  // N
		P:    "-+", // PP
		PGL:  "=",  // G
		S2:   "--", // ss
		SG:   "",   // sg
	},
	"pc850": {
		Bar:  string([]byte{0xB3}),       // B
		BarC: string([]byte{0xC3}),       // C
		BarL: string([]byte{0xB4}),       // L
		EG:   string([]byte{}),           // eg
		Init: string([]byte{}),           // init
		NPGL: string([]byte{0xDA}),       // N
		P:    string([]byte{0xDA, 0xC2}), // PP
		PGL:  "¤",                        // G
		S2:   string([]byte{0xDA, 0xDA}), // ss
		SG:   string([]byte{}),           // sg
	},
	"vt100": {
		Bar:  "\x0Ex\x0F",    // B
		BarC: "\x0Et\x0F",    // C
		BarL: "\x0Em\x0F",    // L
		EG:   "\x0F",         // eg
		Init: "\033(B\033)0", // init
		NPGL: "\x0Eq\x0F",    // N
		P:    "\x0Eqw\x0F",   // PP
		PGL:  "◆",            // G
		S2:   "\x0Eqq\x0F",   // ss
		SG:   "\x0E",         // sg
	},
	"utf8": {
		Bar:  "\342\224\202",             // B
		BarC: "\342\224\234",             // C
		BarL: "\342\224\224",             // L
		EG:   "",                         // eg
		Init: "",                         // init
		NPGL: "\342\224\200",             // N
		P:    "\342\224\200\342\224\254", // PP
		PGL:  "●",                        // G
		S2:   "\342\224\200\342\224\200", // ss
		SG:   "",                         // sg
	},
}

var AnsiEscape = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// ProcessGroup represents a group of identical processes
type ProcessGroup struct {
	Count      int    // Number of identical processes
	FirstIndex int    // Index of the first process in the group
	FullPath   string // Full path of the command
	Indices    []int  // Indices of all processes in the group
	Owner      string // Owner of the process group
}
