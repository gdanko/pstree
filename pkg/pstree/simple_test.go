package pstree

import (
	"testing"

	"github.com/gdanko/pstree/pkg/tree"
	"github.com/stretchr/testify/assert"
)

// TestSimpleSort tests the basic sorting functions
func TestSimpleSort(t *testing.T) {
	// Create test processes with different PIDs
	proc1 := tree.Process{PID: 300}
	proc2 := tree.Process{PID: 100}
	proc3 := tree.Process{PID: 200}

	// Create a slice with the processes
	processes := []tree.Process{proc1, proc2, proc3}

	// Sort the processes by PID
	SortProcsByPid(&processes)

	// Verify that the processes are sorted by PID in ascending order
	assert.Equal(t, int32(100), processes[0].PID)
	assert.Equal(t, int32(200), processes[1].PID)
	assert.Equal(t, int32(300), processes[2].PID)
}
