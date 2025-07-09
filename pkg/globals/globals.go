package globals

import (
	"log/slog"
	"sync"
)

var (
	mu         sync.RWMutex
	logger     *slog.Logger
	debugLevel int
)

// SetDebugLevel sets the global debug level with thread safety.
//
// Parameters:
//   - x: The debug level to set
func SetDebugLevel(x int) {
	mu.Lock()
	debugLevel = x
	mu.Unlock()
}

// GetDebugLevel retrieves the current global debug level with thread safety.
//
// Returns:
//   - int: The current debug level
func GetDebugLevel() (x int) {
	mu.Lock()
	x = debugLevel
	mu.Unlock()
	return x
}

// SetLogger sets the global logger instance with thread safety.
//
// Parameters:
//   - x: The logger instance to set
func SetLogger(x *slog.Logger) {
	mu.Lock()
	logger = x
	mu.Unlock()
}

// GetLogger retrieves the current global logger instance with thread safety.
//
// Returns:
//   - *slog.Logger: The current logger instance
func GetLogger() (x *slog.Logger) {
	mu.Lock()
	x = logger
	mu.Unlock()
	return x
}
