package pstree

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// createMockProcessesForCompact creates a slice of mock Process structs for testing compact mode
func createMockProcessesForCompact() []Process {
	return []Process{
		{
			PID:      1,
			PPID:     0,
			Command:  "/sbin/init",
			Args:     []string{},
			Parent:   -1,
			NumThreads: 1,
		},
		{
			PID:      2,
			PPID:     1,
			Command:  "/usr/bin/bash",
			Args:     []string{"-c", "sleep 100"},
			Parent:   0,
			NumThreads: 1,
		},
		{
			PID:      3,
			PPID:     1,
			Command:  "/usr/bin/bash",
			Args:     []string{"-c", "sleep 100"},
			Parent:   0,
			NumThreads: 1,
		},
		{
			PID:      4,
			PPID:     1,
			Command:  "/usr/bin/bash",
			Args:     []string{"-c", "echo hello"},
			Parent:   0,
			NumThreads: 1,
		},
		{
			PID:      5,
			PPID:     2,
			Command:  "/bin/sleep",
			Args:     []string{"100"},
			Parent:   1,
			NumThreads: 1,
		},
		{
			PID:      6,
			PPID:     2,
			Command:  "/bin/sleep",
			Args:     []string{"100"},
			Parent:   1,
			NumThreads: 1,
		},
		{
			PID:      7,
			PPID:     2,
			Command:  "/bin/sleep",
			Args:     []string{"200"},
			Parent:   1,
			NumThreads: 1,
		},
		{
			PID:      8,
			PPID:     4,
			Command:  "/bin/echo",
			Args:     []string{"hello"},
			Parent:   3,
			NumThreads: 3, // This is a thread
		},
	}
}

func TestInitCompactMode(t *testing.T) {
	processes := createMockProcessesForCompact()
	
	// Initialize compact mode
	InitCompactMode(processes)
	
	// Check that processGroups and skipProcesses are initialized
	assert.NotNil(t, processGroups)
	assert.NotNil(t, skipProcesses)
	
	// Check that processes with identical commands and args are grouped
	// PIDs 2 and 3 have identical command and args
	assert.True(t, skipProcesses[2], "Process with PID 3 should be marked to skip")
	
	// PIDs 5 and 6 have identical command and args
	assert.True(t, skipProcesses[5], "Process with PID 6 should be marked to skip")
	
	// PID 7 has different args, so it shouldn't be skipped
	assert.False(t, skipProcesses[6], "Process with PID 7 should not be marked to skip")
}

func TestShouldSkipProcess(t *testing.T) {
	processes := createMockProcessesForCompact()
	InitCompactMode(processes)
	
	tests := []struct {
		name        string
		processIndex int
		expected    bool
	}{
		{"Process 0 (PID 1) - unique", 0, false},
		{"Process 1 (PID 2) - first of duplicate", 1, false},
		{"Process 2 (PID 3) - duplicate", 2, true},
		{"Process 3 (PID 4) - unique", 3, false},
		{"Process 4 (PID 5) - first of duplicate", 4, false},
		{"Process 5 (PID 6) - duplicate", 5, true},
		{"Process 6 (PID 7) - unique args", 6, false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldSkipProcess(tt.processIndex)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetProcessCount(t *testing.T) {
	processes := createMockProcessesForCompact()
	InitCompactMode(processes)
	
	tests := []struct {
		name        string
		processIndex int
		expectedCount int
		expectedIsThread bool
	}{
		{"Process 0 (PID 1) - unique", 0, 1, false},
		{"Process 1 (PID 2) - has duplicate", 1, 2, true},
		{"Process 3 (PID 4) - unique", 3, 1, true},
		{"Process 4 (PID 5) - has duplicate", 4, 2, true},
		{"Process 6 (PID 7) - unique args", 6, 1, true},
		{"Process 7 (PID 8) - thread", 7, 1, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, isThread := GetProcessCount(processes, tt.processIndex)
			assert.Equal(t, tt.expectedCount, count)
			assert.Equal(t, tt.expectedIsThread, isThread)
		})
	}
}

func TestFormatCompactOutput(t *testing.T) {
	tests := []struct {
		name        string
		command     string
		count       int
		isThread    bool
		hideThreads bool
		expected    string
	}{
		{"Single process", "/usr/bin/bash", 1, false, false, "/usr/bin/bash"},
		{"Multiple processes", "/usr/bin/bash", 3, false, false, "3*[bash]"},
		{"Single thread", "/usr/bin/chrome", 1, true, false, "/usr/bin/chrome"},
		{"Multiple threads", "/usr/bin/chrome", 5, true, false, "5*[{chrome}]"},
		{"Multiple threads hidden", "/usr/bin/chrome", 5, true, true, ""},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatCompactOutput(tt.command, tt.count, tt.isThread, tt.hideThreads)
			assert.Equal(t, tt.expected, result)
		})
	}
}
