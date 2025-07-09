package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"

	"github.com/gdanko/pstree/pkg/globals"
	"github.com/gdanko/pstree/pkg/logger"
	"github.com/gdanko/pstree/pkg/pstree"
	"github.com/gdanko/pstree/pkg/tree"
	"github.com/gdanko/pstree/util"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/spf13/cobra"
)

var (
	colorCount              int
	colorSupport            bool
	debugLevel              int
	displayOptions          tree.DisplayOptions
	errorMessage            string
	flagAge                 bool
	flagArguments           bool
	flagColor               bool
	flagColorAttr           string
	flagColorScheme         string
	flagCompactNot          bool
	flagContains            string
	flagCpu                 bool
	flagExcludeRoot         bool
	flagGenerateThreads     bool // Generate threads for testing purposes
	flagHideThreads         bool
	flagIBM850              bool
	flagLevel               int
	flagMapBasedTree        bool // Experimental map-based tree structure
	flagMemory              bool
	flagOrderBy             string
	flagPid                 int32
	flagRainbow             bool
	flagShowAll             bool
	flagShowGroup           bool
	flagShowOwner           bool
	flagShowPGIDs           bool
	flagShowPGLs            bool
	flagShowPIDs            bool
	flagShowPPIDs           bool
	flagShowUIDTransitions  bool
	flagShowUserTransitions bool
	flagThreads             bool
	flagUsername            []string
	flagUTF8                bool
	flagVersion             bool
	flagVT100               bool
	flagWide                bool
	installedMemory         *mem.VirtualMemoryStat
	processes               []tree.Process
	processTree             *tree.ProcessTree
	processMap              *tree.ProcessMap // New variable for the map-based tree
	screenWidth             int
	sorted                  []tree.Process
	unicodeSupport          bool
	usageTemplate           string
	username                string
	validAttributes         []string = []string{"age", "cpu", "mem"}
	validColorSchemes       []string = []string{"darwin", "linux", "powershell", "windows10", "xterm"}
	validOrderBy            []string = []string{"age", "cmd", "cpu", "mem", "pid", "threads", "user"}
	version                 string   = "0.8.1"
	versionString           string
	rootCmd                 = &cobra.Command{
		Use:    "pstree",
		Short:  "",
		Long:   fmt.Sprintf("pstree $Revision: %s $ by Gary Danko (C) 2025", version),
		PreRun: pstreePreRunCmd,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			globals.SetDebugLevel(debugLevel)
			if debugLevel > 0 {
				fmt.Printf("Debug level: %d\n", debugLevel)
			}
		},
		RunE: pstreeRunCmd,
	}
)

// Execute runs the root command of the pstree application.
// It serves as the entry point for the CLI application.
// Returns any error encountered during command execution.
func Execute() error {
	return rootCmd.Execute()
}

// init initializes the root command with appropriate flags and usage template.
// It determines the current username and color support capabilities of the terminal,
// then sets up the command-line interface with appropriate usage instructions.
func init() {
	username = util.DetermineUsername()
	colorSupport, colorCount = util.HasColorSupport()
	unicodeSupport = util.HasUnicodeSupport()

	GetPersistentFlags(rootCmd, colorSupport, colorCount, unicodeSupport, username)

	usageTemplate = fmt.Sprintf(`Usage: pstree [OPTIONS]

Display a tree of processes.

Application Options:
{{.Flags.FlagUsages}}
Process group leaders are marked with '%s' for ASCII, '%s' for IBM-850, '%s' for VT-100, and '%s' for UTF-8.
`, tree.TreeStyles["ascii"].PGL, tree.TreeStyles["pc850"].PGL, tree.TreeStyles["vt100"].PGL, tree.TreeStyles["utf8"].PGL)

	rootCmd.SetUsageTemplate(usageTemplate)
}

// pstreePreRunCmd is executed before the main run command.
// This function is a hook provided by cobra that runs before the main command execution.
// It can be used for pre-execution setup tasks such as initializing resources,
// validating command-line arguments, or setting up the environment.
//
// Parameters:
//   - cmd: The cobra.Command being executed
//   - args: Command line arguments passed to the command
func pstreePreRunCmd(cmd *cobra.Command, args []string) {
}

// pstreeRunCmd is the main execution function for the pstree command.
// It initializes the logger, validates command flags, processes system information,
// and displays the process tree according to the specified options.
//
// Parameters:
//   - cmd: The command being executed
//   - args: Command line arguments passed to the command
//
// Returns:
//   - error: Any error encountered during execution
func pstreeRunCmd(cmd *cobra.Command, args []string) error {
	if debugLevel > 0 {
		logger.Init(slog.LevelDebug)
	} else {
		logger.Init(slog.LevelInfo)
	}
	globals.SetLogger(logger.Logger)
	installedMemory, _ = util.GetTotalMemory()

	// Flag conflict rules
	// to show if a flag is set, use cmd.Flags().Changed("flag")
	//
	// 1. --user cannot be used with --exclude-root
	// 2. only one of --color-attr, --colorize, and --rainbow can be used
	// 3. only one of --ibm-850, --utf-8, and --vt-100 can be use
	// 4. valid options for --color-attr are: age, cpu, mem
	// 5. only one of --uid-transitions and --user-transitions can be used
	// 6. --level cannot be set to less than 1
	// 7. valid options for --color-scheme are: darwin, linux, windows10, xterm

	// Rule 1: --user root cannot be used with --exclude-root
	if cmd.Flags().Changed("user") && flagExcludeRoot {
		return errors.New("--user and --exclude-root cannot be used together")
	}

	// Rule 2: only one of --color-attr, --color, and --rainbow can be used
	if (util.BtoI(flagColor) + util.BtoI(flagRainbow) + util.StoI(flagColorAttr)) > 1 {
		return errors.New("only one of --color-attr, --color, and --rainbow can be used")
	}

	// Rule 3: only one of --ibm-850, --utf-8, and --vt-100 can be used
	if (util.BtoI(flagIBM850) + util.BtoI(flagUTF8) + util.BtoI(flagVT100)) > 1 {
		return errors.New("only one of --ibm-850, --utf-8, and --vt-100 can be used")
	}

	// Rule 4: valid options for --color-attr are: age, cpu, mem
	if flagColorAttr != "" && !slices.Contains(validAttributes, flagColorAttr) {
		return fmt.Errorf("valid options for --color-attr are: %s", strings.Join(validAttributes, ", "))
	}

	// Rule 5: only one of --uid-transitions and --user-transitions can be used
	if (util.BtoI(flagShowUIDTransitions) + util.BtoI(flagShowUserTransitions)) > 1 {
		return errors.New("only one of --uid-transitions and --user-transitions can be used")
	}

	// Rule 6: --level cannot be set to less than 1
	if cmd.Flags().Changed("level") && flagLevel < 1 {
		return errors.New("--level cannot be set to less than 1")
	}

	// Rule 7: valid options for --color-scheme are: darwin, linux, windows10, xterm
	if flagColorScheme != "" && !slices.Contains(validColorSchemes, flagColorScheme) {
		return fmt.Errorf("valid options for --color-scheme are: %s", strings.Join(validColorSchemes, ", "))
	}

	if flagVersion {
		versionString = fmt.Sprintf(`pstree %s
Copyright (C) 2025 Gary Danko

pstree comes with ABSOLUTELY NO WARRANTY.
This is free software, and you are welcome to redistribute it under
the terms of the GNU General Public License.
For more information about these matters, see the file named LICENSE.`,
			version,
		)
		fmt.Fprintln(os.Stdout, versionString)
		os.Exit(0)
	}

	for i, username := range flagUsername {
		if !util.UserExists(username) {
			excluded := []int{}
			excluded = append(excluded, i)
			logger.Logger.Warn(fmt.Sprintf("user '%s' does not exist, excluding", username))
			if len(excluded) > 0 {
				slices.Reverse(excluded)
				for i := range excluded {
					flagUsername = util.DeleteSliceElement(flagUsername, i)
				}
			}
		}
	}

	screenWidth = util.GetScreenWidth()
	pstree.GetProcesses(&processes, flagGenerateThreads)

	if flagOrderBy != "" {
		if !slices.Contains(validOrderBy, flagOrderBy) {
			errorMessage = fmt.Sprintf("valid options for --order-by are: %s", strings.Join(validOrderBy, ", "))
			return errors.New(errorMessage)
		}
		proc, err := pstree.GetProcessByPid(&processes, 1)
		if err != nil {
			panic(err)
		}
		sorted = []tree.Process{proc}
		switch flagOrderBy {
		case "age":
			flagAge = true
			pstree.SortProcsByAge(&processes)
		case "cmd":
			pstree.SortProcsByCmd(&processes)
		case "cpu":
			flagCpu = true
			pstree.SortProcsByCpu(&processes)
		case "mem":
			flagMemory = true
			pstree.SortProcsByMemory(&processes)
		case "pid":
			flagShowPIDs = true
			pstree.SortProcsByPid(&processes)
		case "threads":
			flagThreads = true
			pstree.SortProcsByNumThreads(&processes)
		case "user":
			flagShowOwner = true
			pstree.SortProcsByUsername(&processes)
		default:
			sorted = processes
		}

		for _, proc := range processes {
			if proc.PID != 1 {
				sorted = append(sorted, proc)
			}
		}
		processes = sorted
	}

	if flagLevel == 0 {
		flagLevel = 999
	}

	// If any of the following flags are set, then compact mode should be disabled
	// This is because some of the results or offenders may be buried in collapsed subtrees
	if flagColorAttr != "" || flagCpu || flagMemory || flagContains != "" {
		flagCompactNot = true
	}

	if flagShowAll {
		flagAge = true
		flagArguments = true
		flagCpu = true
		flagMemory = true
		flagShowGroup = true
		flagShowOwner = true
		flagShowPGIDs = true
		flagShowPIDs = true
		flagThreads = true
	}

	displayOptions = tree.DisplayOptions{
		ColorAttr:           flagColorAttr,
		ColorCount:          colorCount,
		ColorizeOutput:      flagColor,
		ColorScheme:         flagColorScheme,
		ColorSupport:        colorSupport,
		CompactMode:         !flagCompactNot,
		Contains:            flagContains,
		ExcludeRoot:         flagExcludeRoot,
		HideThreads:         flagHideThreads,
		IBM850Graphics:      flagIBM850,
		InstalledMemory:     installedMemory.Total,
		MaxDepth:            flagLevel,
		OrderBy:             flagOrderBy,
		RainbowOutput:       flagRainbow,
		RootPID:             flagPid,
		ScreenWidth:         screenWidth,
		ShowArguments:       flagArguments,
		ShowCpuPercent:      flagCpu,
		ShowGroup:           flagShowGroup,
		ShowMemoryUsage:     flagMemory,
		ShowNumThreads:      flagThreads,
		ShowOwner:           flagShowOwner,
		ShowPGIDs:           flagShowPGIDs,
		ShowPGLs:            flagShowPGLs,
		ShowPIDs:            flagShowPIDs,
		ShowPPIDs:           flagShowPPIDs,
		ShowProcessAge:      flagAge,
		ShowUIDTransitions:  flagShowUIDTransitions,
		ShowUserTransitions: flagShowUserTransitions,
		Usernames:           flagUsername,
		UTF8Graphics:        flagUTF8,
		VT100Graphics:       flagVT100,
		WideDisplay:         flagWide,
	}

	// Choose between traditional array-based tree or new map-based tree
	// Filtering by PID, username, etc. is not currently working with the map-based implementation
	if flagMapBasedTree {
		// Use the new map-based tree structure
		logger.Logger.Debug("Using map-based tree structure")

		// Build the process map
		processMap = tree.NewProcessMap(logger.Logger, processes, displayOptions)

		// Mark processes to be displayed
		processMap.FindPrintable()
		// pretty.Println(processMap.Nodes)

		// Drop unmarked processes
		// processMap.DropUnmarked()

		// Show processes that will be displayed
		processMap.ShowPrintable()

		// Print the process tree with simple indentation based on depth
		processMap.PrintTree()

	} else {
		// Use the traditional array-based tree structure
		logger.Logger.Debug("Using traditional array-based tree structure")

		// Generate the process tree
		processTree = tree.NewProcessTree(debugLevel, logger.Logger, processes, displayOptions)

		// Mark processes to be displayed
		processTree.MarkProcesses()

		// Drop unmarked processes
		processTree.DropUnmarked()

		// Show processes that will be displayed
		if processTree.DebugLevel > 2 {
			processTree.ShowPrintable()
			os.Exit(0)
		}

		// Print the tree
		processTree.PrintTree(0, "")
	}

	return nil
}
