package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCommandOutput tests that the pstree command produces the expected output format
// based on the command-line flags provided.
func TestCommandOutput(t *testing.T) {
	// Skip this test if we're not running in the full test environment
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Find the path to the pstree binary
	binPath, err := findPstreeBinary()
	if err != nil {
		t.Skipf("Skipping test: %v", err)
	}

	// Test cases for different command-line flags
	testCases := []struct {
		name          string
		args          []string
		expectedRegex []string // Regular expressions that should match the output
	}{
		{
			name: "Show CPU percentage (-c)",
			args: []string{"-c"},
			expectedRegex: []string{
				`\(c:\d+\.\d+%\)`, // Match CPU percentage format (c:X.XX%)
			},
		},
		{
			name: "Show memory usage (-m)",
			args: []string{"-m"},
			expectedRegex: []string{
				`\(m:\d+\.\d+ (?:B|KiB|MiB|GiB)\)`, // Match memory usage format with any unit
			},
		},
		{
			name: "Show process age (-G)",
			args: []string{"-G"},
			expectedRegex: []string{
				`\(\d{2}:\d{2}:\d{2}:\d{2}\)`, // Match process age format (DD:HH:MM:SS)
			},
		},
		{
			name: "Show PIDs (-p)",
			args: []string{"-p"},
			expectedRegex: []string{
				`\(\d+\)`, // Match PID format (X)
			},
		},
		{
			name: "Show PGIDs (-g)",
			args: []string{"-g"},
			expectedRegex: []string{
				`\(\d+\)`, // Match PGID format (X)
			},
		},
		{
			name: "Show PIDs and PGIDs (-p, -g)",
			args: []string{"-p", "-g"},
			expectedRegex: []string{
				`\(\d+\,\d+\)`, // Match PID format (X)
			},
		},
		{
			name: "Show threads (-t)",
			args: []string{"-t"},
			expectedRegex: []string{
				`\(t:\d+\)`, // Match thread count format (t:X)
			},
		},
		{
			name: "Show owner (-O)",
			args: []string{"-O"},
			expectedRegex: []string{
				`\w+\s+\S+`, // Match username followed by process name
			},
		},
		{
			name: "All options (-A)",
			args: []string{"-A"},
			expectedRegex: []string{
				`\(\d+,\d+\)`,                      // PIDs and PGIDs combined
				`\(t:\d+\)`,                        // Thread count
				`\(\d{2}:\d{2}:\d{2}:\d{2}\)`,      // Process age
				`\(c:\d+\.\d+%\)`,                  // CPU percentage
				`\(m:\d+\.\d+ (?:B|KiB|MiB|GiB)\)`, // Memory usage with any unit
			},
		},
	}

	// Run each test case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Run the pstree command with the specified arguments
			cmd := exec.Command(binPath, tc.args...)
			var stdout bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = os.Stderr

			err := cmd.Run()
			if err != nil {
				t.Fatalf("Failed to run pstree command: %v", err)
			}

			// Get the command output
			output := stdout.String()

			// Check that the output matches all expected regex patterns
			for _, pattern := range tc.expectedRegex {
				re := regexp.MustCompile(pattern)
				assert.True(t, re.MatchString(output),
					"Output should match pattern '%s' when using %v flags\nOutput: %s",
					pattern, tc.args, output)
			}
		})
	}
}

// findPstreeBinary attempts to locate the pstree binary in common locations
func findPstreeBinary() (string, error) {
	// Try common locations
	locations := []string{
		"../bin/pstree", // Relative to cmd directory
		"/Users/gary.danko/gitlab/pstree/bin/pstree", // Absolute path
	}

	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			return loc, nil
		}
	}

	// If not found, try to build it
	cmd := exec.Command("make", "build")
	cmd.Dir = "/Users/gary.danko/gitlab/pstree"
	if err := cmd.Run(); err != nil {
		return "", err
	}

	// Check if the build succeeded
	if _, err := os.Stat("/Users/gary.danko/gitlab/pstree/bin/pstree"); err == nil {
		return "/Users/gary.danko/gitlab/pstree/bin/pstree", nil
	}

	return "", os.ErrNotExist
}
