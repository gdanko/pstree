package pstree

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
	"github.com/stretchr/testify/assert"
)

// setupChannelTestLogger creates a logger for testing the channel functions
func setupChannelTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
}

// TestProcessArgs tests the ProcessArgs function
func TestProcessArgs(t *testing.T) {
	logger := setupChannelTestLogger()
	
	// Create a channel to receive the function
	c := make(chan func(ctx context.Context, proc *process.Process) ([]string, error))
	
	// Call ProcessArgs in a goroutine
	go ProcessArgs(c, logger)
	
	// Get the function from the channel
	fn := <-c
	
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	
	// Create a mock process (we can't actually call the function with a real process in unit tests)
	proc := &process.Process{Pid: int32(os.Getpid())}
	
	// Call the function - we're just testing that it doesn't panic
	assert.NotPanics(t, func() {
		_, _ = fn(ctx, proc)
	})
}

// TestProcessCommandName tests the ProcessCommandName function
func TestProcessCommandName(t *testing.T) {
	logger := setupChannelTestLogger()
	
	// Create a channel to receive the function
	c := make(chan func(ctx context.Context, proc *process.Process) (string, error))
	
	// Call ProcessCommandName in a goroutine
	go ProcessCommandName(c, logger)
	
	// Get the function from the channel
	fn := <-c
	
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	
	// Create a mock process
	proc := &process.Process{Pid: int32(os.Getpid())}
	
	// Call the function - we're just testing that it doesn't panic
	assert.NotPanics(t, func() {
		_, _ = fn(ctx, proc)
	})
}

// TestProcessConnections tests the ProcessConnections function
func TestProcessConnections(t *testing.T) {
	logger := setupChannelTestLogger()
	
	// Create a channel to receive the function
	c := make(chan func(ctx context.Context, proc *process.Process) ([]net.ConnectionStat, error))
	
	// Call ProcessConnections in a goroutine
	go ProcessConnections(c, logger)
	
	// Get the function from the channel
	fn := <-c
	
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	
	// Create a mock process
	proc := &process.Process{Pid: int32(os.Getpid())}
	
	// Call the function - we're just testing that it doesn't panic
	assert.NotPanics(t, func() {
		_, _ = fn(ctx, proc)
	})
}

// TestProcessCpuPercent tests the ProcessCpuPercent function
func TestProcessCpuPercent(t *testing.T) {
	logger := setupChannelTestLogger()
	
	// Create a channel to receive the function
	c := make(chan func(ctx context.Context, proc *process.Process) (float64, error))
	
	// Call ProcessCpuPercent in a goroutine
	go ProcessCpuPercent(c, logger)
	
	// Get the function from the channel
	fn := <-c
	
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	
	// Create a mock process
	proc := &process.Process{Pid: int32(os.Getpid())}
	
	// Call the function - we're just testing that it doesn't panic
	assert.NotPanics(t, func() {
		_, _ = fn(ctx, proc)
	})
}

// TestProcessCpuTimes tests the ProcessCpuTimes function
func TestProcessCpuTimes(t *testing.T) {
	logger := setupChannelTestLogger()
	
	// Create a channel to receive the function
	c := make(chan func(ctx context.Context, proc *process.Process) (*cpu.TimesStat, error))
	
	// Call ProcessCpuTimes in a goroutine
	go ProcessCpuTimes(c, logger)
	
	// Get the function from the channel
	fn := <-c
	
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	
	// Create a mock process
	proc := &process.Process{Pid: int32(os.Getpid())}
	
	// Call the function - we're just testing that it doesn't panic
	assert.NotPanics(t, func() {
		_, _ = fn(ctx, proc)
	})
}

// TestProcessCreateTime tests the ProcessCreateTime function
func TestProcessCreateTime(t *testing.T) {
	logger := setupChannelTestLogger()
	
	// Create a channel to receive the function
	c := make(chan func(ctx context.Context, proc *process.Process) (int64, error))
	
	// Call ProcessCreateTime in a goroutine
	go ProcessCreateTime(c, logger)
	
	// Get the function from the channel
	fn := <-c
	
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	
	// Create a mock process
	proc := &process.Process{Pid: int32(os.Getpid())}
	
	// Call the function - we're just testing that it doesn't panic
	assert.NotPanics(t, func() {
		_, _ = fn(ctx, proc)
	})
}

// TestProcessUIDs tests the ProcessUIDs function
func TestProcessUIDs(t *testing.T) {
	logger := setupChannelTestLogger()
	
	// Create a channel to receive the function
	c := make(chan func(ctx context.Context, proc *process.Process) ([]uint32, error))
	
	// Call ProcessUIDs in a goroutine
	go ProcessUIDs(c, logger)
	
	// Get the function from the channel
	fn := <-c
	
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	
	// Create a mock process
	proc := &process.Process{Pid: int32(os.Getpid())}
	
	// Call the function - we're just testing that it doesn't panic
	assert.NotPanics(t, func() {
		_, _ = fn(ctx, proc)
	})
}

// TestProcessUsername tests the ProcessUsername function
func TestProcessUsername(t *testing.T) {
	logger := setupChannelTestLogger()
	
	// Create a channel to receive the function
	c := make(chan func(ctx context.Context, proc *process.Process) (string, error))
	
	// Call ProcessUsername in a goroutine
	go ProcessUsername(c, logger)
	
	// Get the function from the channel
	fn := <-c
	
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	
	// Create a mock process
	proc := &process.Process{Pid: int32(os.Getpid())}
	
	// Call the function - we're just testing that it doesn't panic
	assert.NotPanics(t, func() {
		_, _ = fn(ctx, proc)
	})
}

// TestProcessPGID tests the ProcessPGID function
func TestProcessPGID(t *testing.T) {
	logger := setupChannelTestLogger()
	
	// Create a channel to receive the function
	c := make(chan func(proc *process.Process) (int, error))
	
	// Call ProcessPGID in a goroutine
	go ProcessPGID(c, logger)
	
	// Get the function from the channel
	fn := <-c
	
	// Create a mock process
	proc := &process.Process{Pid: int32(os.Getpid())}
	
	// Call the function - we're just testing that it doesn't panic
	assert.NotPanics(t, func() {
		_, _ = fn(proc)
	})
}
