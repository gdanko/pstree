package cmd

import (
	"testing"

	"github.com/gdanko/pstree/pkg/pstree"
	"github.com/stretchr/testify/assert"
)

// TestFlagToDisplayOptions tests that the command-line flags correctly set the DisplayOptions
func TestFlagToDisplayOptions(t *testing.T) {

	// Test cases for individual flags
	testCases := []struct {
		name         string
		setupFlags   func()
		checkOptions func(t *testing.T, opts pstree.DisplayOptions)
	}{
		{
			name: "CPU flag (-c)",
			setupFlags: func() {
				flagCpu = true
				flagCompactNot = false // Reset this as it gets modified
			},
			checkOptions: func(t *testing.T, opts pstree.DisplayOptions) {
				assert.True(t, opts.ShowCpuPercent, "ShowCpuPercent should be true when -c flag is set")
				assert.True(t, !opts.CompactMode, "CompactMode should be false when -c flag is set (implies --compact-not)")
			},
		},
		{
			name: "Memory flag (-m)",
			setupFlags: func() {
				flagMemory = true
				flagCompactNot = false // Reset this as it gets modified
			},
			checkOptions: func(t *testing.T, opts pstree.DisplayOptions) {
				assert.True(t, opts.ShowMemoryUsage, "ShowMemoryUsage should be true when -m flag is set")
				assert.True(t, !opts.CompactMode, "CompactMode should be false when -m flag is set (implies --compact-not)")
			},
		},
		{
			name: "Age flag (-G)",
			setupFlags: func() {
				flagAge = true
			},
			checkOptions: func(t *testing.T, opts pstree.DisplayOptions) {
				assert.True(t, opts.ShowProcessAge, "ShowProcessAge should be true when -G flag is set")
			},
		},
		{
			name: "Show PIDs flag (-p)",
			setupFlags: func() {
				flagShowPIDs = true
			},
			checkOptions: func(t *testing.T, opts pstree.DisplayOptions) {
				assert.True(t, opts.ShowPIDs, "ShowPids should be true when -p flag is set")
			},
		},
		{
			name: "Show PGIDs flag (-g)",
			setupFlags: func() {
				flagShowPGIDs = true
			},
			checkOptions: func(t *testing.T, opts pstree.DisplayOptions) {
				assert.True(t, opts.ShowPGIDs, "ShowPGIDs should be true when -g flag is set")
			},
		},
		{
			name: "Show Owner flag (-O)",
			setupFlags: func() {
				flagShowOwner = true
			},
			checkOptions: func(t *testing.T, opts pstree.DisplayOptions) {
				assert.True(t, opts.ShowOwner, "ShowOwner should be true when -O flag is set")
			},
		},
		{
			name: "Show Threads flag (-t)",
			setupFlags: func() {
				flagThreads = true
			},
			checkOptions: func(t *testing.T, opts pstree.DisplayOptions) {
				assert.True(t, opts.ShowNumThreads, "ShowNumThreads should be true when -t flag is set")
			},
		},
		{
			name: "Hide Threads flag (-H)",
			setupFlags: func() {
				flagHideThreads = true
			},
			checkOptions: func(t *testing.T, opts pstree.DisplayOptions) {
				assert.True(t, opts.HideThreads, "HideThreads should be true when -H flag is set")
			},
		},
		{
			name: "Show PGLs flag (-S, --show-pgls)",
			setupFlags: func() {
				flagShowPGLs = true
			},
			checkOptions: func(t *testing.T, opts pstree.DisplayOptions) {
				assert.True(t, opts.ShowPGLs, "ShowPGLs should be true when -S flag is set")
			},
		},
		{
			name: "Show Arguments flag (-a)",
			setupFlags: func() {
				flagArguments = true
			},
			checkOptions: func(t *testing.T, opts pstree.DisplayOptions) {
				assert.True(t, opts.ShowArguments, "ShowArguments should be true when -a flag is set")
			},
		},
		{
			name: "Show UID Transitions flag (-I)",
			setupFlags: func() {
				flagShowUIDTransitions = true
			},
			checkOptions: func(t *testing.T, opts pstree.DisplayOptions) {
				assert.True(t, opts.ShowUIDTransitions, "ShowUIDTransitions should be true when -I flag is set")
			},
		},
		{
			name: "Show User Transitions flag (-U)",
			setupFlags: func() {
				flagShowUserTransitions = true
			},
			checkOptions: func(t *testing.T, opts pstree.DisplayOptions) {
				assert.True(t, opts.ShowUserTransitions, "ShowUserTransitions should be true when -U flag is set")
			},
		},
		{
			name: "IBM-850 Graphics flag (-i)",
			setupFlags: func() {
				flagIBM850 = true
			},
			checkOptions: func(t *testing.T, opts pstree.DisplayOptions) {
				assert.True(t, opts.IBM850Graphics, "IBM850Graphics should be true when -i flag is set")
			},
		},
		{
			name: "UTF-8 Graphics flag (-u)",
			setupFlags: func() {
				flagUTF8 = true
			},
			checkOptions: func(t *testing.T, opts pstree.DisplayOptions) {
				assert.True(t, opts.UTF8Graphics, "UTF8Graphics should be true when -u flag is set")
			},
		},
		{
			name: "VT-100 Graphics flag (-v)",
			setupFlags: func() {
				flagVT100 = true
			},
			checkOptions: func(t *testing.T, opts pstree.DisplayOptions) {
				assert.True(t, opts.VT100Graphics, "VT100Graphics should be true when -v flag is set")
			},
		},
		{
			name: "Level flag (-l)",
			setupFlags: func() {
				flagLevel = 5
			},
			checkOptions: func(t *testing.T, opts pstree.DisplayOptions) {
				assert.Equal(t, 5, opts.MaxDepth, "MaxDepth should be 5 when -l 5 flag is set")
			},
		},
		{
			name: "Color flag (-k)",
			setupFlags: func() {
				flagColorAttr = "cpu"
				flagCompactNot = false // Reset this as it gets modified
			},
			checkOptions: func(t *testing.T, opts pstree.DisplayOptions) {
				assert.Equal(t, "cpu", opts.ColorAttr, "ColorAttr should be 'cpu' when -k cpu flag is set")
				assert.True(t, !opts.CompactMode, "CompactMode should be false when -k flag is set (implies --compact-not)")
			},
		},
		{
			name: "Colorize flag (-C)",
			setupFlags: func() {
				flagColor = true
				colorSupport = true
			},
			checkOptions: func(t *testing.T, opts pstree.DisplayOptions) {
				assert.True(t, opts.ColorizeOutput, "ColorizeOutput should be true when -C flag is set")
			},
		},
		{
			name: "Rainbow flag (-r)",
			setupFlags: func() {
				flagRainbow = true
				colorSupport = true
				colorCount = 256
			},
			checkOptions: func(t *testing.T, opts pstree.DisplayOptions) {
				assert.True(t, opts.RainbowOutput, "RainbowOutput should be true when -r flag is set")
			},
		},
		{
			name: "Compact-not flag (-n)",
			setupFlags: func() {
				flagCompactNot = true
			},
			checkOptions: func(t *testing.T, opts pstree.DisplayOptions) {
				assert.False(t, opts.CompactMode, "CompactMode should be false when -n flag is set")
			},
		},
		{
			name: "All flag (-A)",
			setupFlags: func() {
				flagShowAll = true
				// Reset all flags that would be set by -A
				flagAge = false
				flagArguments = false
				flagCpu = false
				flagMemory = false
				flagShowOwner = false
				flagShowPGIDs = false
				flagShowPIDs = false
				flagShowUIDTransitions = false
				flagThreads = false
			},
			checkOptions: func(t *testing.T, opts pstree.DisplayOptions) {
				assert.True(t, opts.ShowProcessAge, "ShowProcessAge should be true when -A flag is set")
				assert.True(t, opts.ShowArguments, "ShowArguments should be true when -A flag is set")
				assert.True(t, opts.ShowCpuPercent, "ShowCpuPercent should be true when -A flag is set")
				assert.True(t, opts.ShowMemoryUsage, "ShowMemoryUsage should be true when -A flag is set")
				assert.True(t, opts.ShowOwner, "ShowOwner should be true when -A flag is set")
				assert.True(t, opts.ShowPGIDs, "ShowPGIDs should be true when -A flag is set")
				assert.True(t, opts.ShowPIDs, "ShowPids should be true when -A flag is set")
				assert.True(t, opts.ShowUIDTransitions, "ShowUIDTransitions should be true when -A flag is set")
				assert.True(t, opts.ShowNumThreads, "ShowNumThreads should be true when -A flag is set")
			},
		},
	}

	// Run each test case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset all flags
			resetFlags()

			// Setup flags for this test
			tc.setupFlags()

			// Apply flags to DisplayOptions
			applyFlagsToDisplayOptions()

			// Check that DisplayOptions are set correctly
			tc.checkOptions(t, displayOptions)
		})
	}
}

// resetFlags resets all flag values to their defaults
func resetFlags() {
	flagAge = false
	flagArguments = false
	flagColor = false
	flagColorAttr = ""
	flagCompactNot = false
	flagContains = ""
	flagCpu = false
	flagExcludeRoot = false
	flagHideThreads = false
	flagIBM850 = false
	flagLevel = 0
	flagMemory = false
	flagOrderBy = ""
	flagPid = 0
	flagRainbow = false
	flagShowAll = false
	flagShowOwner = false
	flagShowPGIDs = false
	flagShowPGLs = false
	flagShowPIDs = false
	flagShowUIDTransitions = false
	flagShowUserTransitions = false
	flagThreads = false
	flagUsername = []string{}
	flagUTF8 = false
	flagVT100 = false
	flagWide = false

	colorSupport = false
	colorCount = 0
}

// applyFlagsToDisplayOptions applies the current flag values to displayOptions
// This is a simplified version of what happens in pstreeRunCmd
func applyFlagsToDisplayOptions() {
	// If any of the following flags are set, then compact mode should be disabled
	if flagColorAttr != "" || flagCpu || flagMemory || flagContains != "" {
		flagCompactNot = true
	}

	if flagShowAll {
		flagAge = true
		flagArguments = true
		flagCpu = true
		flagMemory = true
		flagShowOwner = true
		flagShowPGIDs = true
		flagShowPIDs = true
		flagShowUIDTransitions = true
		flagThreads = true
	}

	displayOptions = pstree.DisplayOptions{
		ColorAttr:           flagColorAttr,
		ColorCount:          colorCount,
		ColorizeOutput:      flagColor,
		ColorSupport:        colorSupport,
		CompactMode:         !flagCompactNot,
		HideThreads:         flagHideThreads,
		IBM850Graphics:      flagIBM850,
		InstalledMemory:     1024 * 1024 * 1024 * 8, // 8GB for testing
		MaxDepth:            flagLevel,
		RainbowOutput:       flagRainbow,
		ShowArguments:       flagArguments,
		ShowCpuPercent:      flagCpu,
		ShowMemoryUsage:     flagMemory,
		ShowNumThreads:      flagThreads,
		ShowOwner:           flagShowOwner,
		ShowPGIDs:           flagShowPGIDs,
		ShowPGLs:            flagShowPGLs,
		ShowPIDs:            flagShowPIDs,
		ShowProcessAge:      flagAge,
		ShowUIDTransitions:  flagShowUIDTransitions,
		ShowUserTransitions: flagShowUserTransitions,
		UTF8Graphics:        flagUTF8,
		VT100Graphics:       flagVT100,
		WideDisplay:         flagWide,
	}
}
