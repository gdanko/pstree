package pstree

import (
	"log/slog"
	"os"
	"testing"

	"github.com/shirou/gopsutil/v4/process"
	"github.com/stretchr/testify/assert"
)

// setupTestLogger creates a logger for testing
func setupTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
}

// createMockProcesses creates a slice of mock Process structs for testing
func createMockProcesses() []Process {
	return []Process{
		{
			PID:      1,
			PPID:     0,
			Command:  "init",
			Username: "root",
			PGID:     1,
			Parent:   -1,
			Child:    1,
			Sister:   -1,
			Print:    true,
			UIDs:     []uint32{0},
			MemoryInfo: &process.MemoryInfoStat{
				RSS: 1024 * 1024, // 1MB
			},
			CPUPercent: 0.5,
			NumThreads: 1,
			Age:        3600, // 1 hour
		},
		{
			PID:      2,
			PPID:     1,
			Command:  "systemd",
			Username: "root",
			PGID:     2,
			Parent:   0,
			Child:    2,
			Sister:   -1,
			Print:    true,
			UIDs:     []uint32{0},
			MemoryInfo: &process.MemoryInfoStat{
				RSS: 2 * 1024 * 1024, // 2MB
			},
			CPUPercent: 1.0,
			NumThreads: 2,
			Age:        1800, // 30 minutes
		},
		{
			PID:      3,
			PPID:     1,
			Command:  "kworker",
			Username: "root",
			PGID:     3,
			Parent:   0,
			Child:    -1,
			Sister:   1,
			Print:    true,
			UIDs:     []uint32{0},
			MemoryInfo: &process.MemoryInfoStat{
				RSS: 512 * 1024, // 512KB
			},
			CPUPercent: 0.2,
			NumThreads: 1,
			Age:        900, // 15 minutes
		},
		{
			PID:      4,
			PPID:     2,
			Command:  "bash",
			Username: "user",
			PGID:     4,
			Parent:   1,
			Child:    -1,
			Sister:   -1,
			Print:    true,
			UIDs:     []uint32{1000},
			HasUIDTransition: true,
			ParentUID:        0,
			ParentUsername:   "root",
			MemoryInfo: &process.MemoryInfoStat{
				RSS: 3 * 1024 * 1024, // 3MB
			},
			CPUPercent: 0.8,
			NumThreads: 1,
			Age:        600, // 10 minutes
		},
	}
}

func TestSortByPid(t *testing.T) {
	// Create a slice of process.Process pointers in unsorted order
	procs := []*process.Process{
		{Pid: 3},
		{Pid: 1},
		{Pid: 4},
		{Pid: 2},
	}

	// Sort the slice
	sorted := SortByPid(procs)

	// Verify the sorting
	assert.Equal(t, int32(1), sorted[0].Pid)
	assert.Equal(t, int32(2), sorted[1].Pid)
	assert.Equal(t, int32(3), sorted[2].Pid)
	assert.Equal(t, int32(4), sorted[3].Pid)
}

func TestGetPidFromIndex(t *testing.T) {
	processes := createMockProcesses()

	tests := []struct {
		name     string
		index    int
		expected int32
	}{
		{"Valid index 0", 0, 1},
		{"Valid index 1", 1, 2},
		{"Valid index 2", 2, 3},
		{"Valid index 3", 3, 4},
		{"Invalid index", 10, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pid := GetPidFromIndex(&processes, tt.index)
			assert.Equal(t, tt.expected, pid)
		})
	}
}

func TestFindPrintable(t *testing.T) {
	processes := createMockProcesses()

	// Set some processes as not printable
	processes[1].Print = false
	processes[3].Print = false

	printable := FindPrintable(&processes)

	// Verify only printable processes are returned
	assert.Equal(t, 2, len(printable))
	assert.Equal(t, int32(1), printable[0].PID)
	assert.Equal(t, int32(3), printable[1].PID)
}

func TestGetProcessByPid(t *testing.T) {
	processes := createMockProcesses()

	tests := []struct {
		name        string
		pid         int32
		expectError bool
	}{
		{"Find existing PID 1", 1, false},
		{"Find existing PID 2", 2, false},
		{"Find existing PID 3", 3, false},
		{"Find existing PID 4", 4, false},
		{"Find non-existent PID", 99, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proc, err := GetProcessByPid(&processes, tt.pid)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.pid, proc.PID)
			}
		})
	}
}

func TestSortProcsByAge(t *testing.T) {
	processes := createMockProcesses()

	// Sort by age
	SortProcsByAge(&processes)

	// Verify sorting (ascending order)
	assert.Equal(t, int32(4), processes[0].PID) // 10 minutes
	assert.Equal(t, int32(3), processes[1].PID) // 15 minutes
	assert.Equal(t, int32(2), processes[2].PID) // 30 minutes
	assert.Equal(t, int32(1), processes[3].PID) // 1 hour
}

func TestSortProcsByCpu(t *testing.T) {
	processes := createMockProcesses()

	// Sort by CPU usage
	SortProcsByCpu(&processes)

	// Verify sorting (ascending order)
	assert.Equal(t, int32(3), processes[0].PID) // 0.2%
	assert.Equal(t, int32(1), processes[1].PID) // 0.5%
	assert.Equal(t, int32(4), processes[2].PID) // 0.8%
	assert.Equal(t, int32(2), processes[3].PID) // 1.0%
}

func TestSortProcsByMemory(t *testing.T) {
	processes := createMockProcesses()

	// Sort by memory usage
	SortProcsByMemory(&processes)

	// Verify sorting (ascending order)
	assert.Equal(t, int32(3), processes[0].PID) // 512KB
	assert.Equal(t, int32(1), processes[1].PID) // 1MB
	assert.Equal(t, int32(2), processes[2].PID) // 2MB
	assert.Equal(t, int32(4), processes[3].PID) // 3MB
}

func TestSortProcsByUsername(t *testing.T) {
	processes := createMockProcesses()

	// Sort by username
	SortProcsByUsername(&processes)

	// Verify sorting (alphabetical order)
	assert.Equal(t, "root", processes[0].Username)
	assert.Equal(t, "root", processes[1].Username)
	assert.Equal(t, "root", processes[2].Username)
	assert.Equal(t, "user", processes[3].Username)
}

func TestSortProcsByPid(t *testing.T) {
	processes := []Process{
		{PID: 3},
		{PID: 1},
		{PID: 4},
		{PID: 2},
	}

	// Sort by PID
	SortProcsByPid(&processes)

	// Verify sorting (ascending order)
	assert.Equal(t, int32(1), processes[0].PID)
	assert.Equal(t, int32(2), processes[1].PID)
	assert.Equal(t, int32(3), processes[2].PID)
	assert.Equal(t, int32(4), processes[3].PID)
}

func TestSortProcsByNumThreads(t *testing.T) {
	processes := createMockProcesses()

	// Sort by number of threads
	SortProcsByNumThreads(&processes)

	// Verify sorting (ascending order)
	// PIDs 1, 3, and 4 all have 1 thread, but their order should be preserved
	assert.Equal(t, int32(1), processes[0].PID)
	assert.Equal(t, int32(3), processes[1].PID)
	assert.Equal(t, int32(4), processes[2].PID)
	assert.Equal(t, int32(2), processes[3].PID) // 2 threads
}

func TestGetPIDIndex(t *testing.T) {
	logger := setupTestLogger()
	processes := createMockProcesses()

	tests := []struct {
		name     string
		pid      int32
		expected int
	}{
		{"Find existing PID 1", 1, 0},
		{"Find existing PID 2", 2, 1},
		{"Find existing PID 3", 3, 2},
		{"Find existing PID 4", 4, 3},
		{"Find non-existent PID", 99, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			index := GetPIDIndex(logger, processes, tt.pid)
			assert.Equal(t, tt.expected, index)
		})
	}
}

func TestMarkCurrentAndAncestors(t *testing.T) {
	logger := setupTestLogger()
	processes := createMockProcesses()

	// Mark process 4 and its ancestors
	MarkCurrentAndAncestors(logger, &processes, 4)

	// Verify marking
	assert.True(t, processes[0].IsCurrentOrAncestor) // PID 1 (grandparent)
	assert.True(t, processes[1].IsCurrentOrAncestor) // PID 2 (parent)
	assert.False(t, processes[2].IsCurrentOrAncestor) // PID 3 (unrelated)
	assert.True(t, processes[3].IsCurrentOrAncestor) // PID 4 (current)
}

func TestMarkUIDTransitions(t *testing.T) {
	logger := setupTestLogger()
	processes := createMockProcesses()

	// Reset the HasUIDTransition flag
	for i := range processes {
		processes[i].HasUIDTransition = false
	}

	// Mark UID transitions
	MarkUIDTransitions(logger, &processes)

	// Verify marking
	assert.False(t, processes[0].HasUIDTransition) // PID 1 (root process)
	assert.False(t, processes[1].HasUIDTransition) // PID 2 (same UID as parent)
	assert.False(t, processes[2].HasUIDTransition) // PID 3 (same UID as parent)
	assert.True(t, processes[4-1].HasUIDTransition) // PID 4 (different UID from parent)
}
