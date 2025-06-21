package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"

	"github.com/gdanko/pstree/pkg/logger"
	"github.com/gdanko/pstree/pkg/pstree"
	"github.com/gdanko/pstree/util"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/spf13/cobra"
)

var (
	colorCount              int
	colorSupport            bool
	displayOptions          pstree.DisplayOptions
	errorMessage            string
	flagAge                 bool
	flagArguments           bool
	flagColor               string
	flagColorize            bool
	flagCompactNot          bool
	flagContains            string
	flagCpu                 bool
	flagDebug               bool
	flagExcludeRoot         bool
	flagGraphicsMode        int
	flagHideThreads         bool
	flagIBM850              bool
	flagLevel               int
	flagMemory              bool
	flagOrderBy             string
	flagPid                 int32
	flagRainbow             bool
	flagShowAll             bool
	flagShowOwner           bool
	flagShowPgids           bool
	flagShowPGL             bool
	flagShowPids            bool
	flagShowUIDTransitions  bool
	flagShowUserTransitions bool
	flagThreads             bool
	flagUsername            []string
	flagUTF8                bool
	flagVersion             bool
	flagVT100               bool
	flagWide                bool
	installedMemory         *mem.VirtualMemoryStat
	processes               []pstree.Process
	screenWidth             int
	sorted                  []pstree.Process
	usageTemplate           string
	username                string
	validAttributes         []string = []string{"age", "cpu", "mem"}
	validOrderBy            []string = []string{"age", "cpu", "mem", "pid", "threads", "user"}
	version                 string   = "0.7.7"
	versionString           string
	rootCmd                 = &cobra.Command{
		Use:    "pstree",
		Short:  "",
		Long:   fmt.Sprintf("pstree $Revision: %s $ by Gary Danko (C) 2025", version),
		PreRun: pstreePreRunCmd,
		RunE:   pstreeRunCmd,
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

	GetPersistentFlags(rootCmd, colorSupport, colorCount, username)

	usageTemplate = fmt.Sprintf(`Usage: pstree [OPTIONS]

Display a tree of processes.

Application Options:
{{.Flags.FlagUsages}}
Process group leaders are marked with '%s' for ASCII, '%s' for IBM-850, '%s' for VT-100, and '%s' for UTF-8.
`, pstree.TreeStyles["ascii"].PGL, pstree.TreeStyles["pc850"].PGL, pstree.TreeStyles["vt100"].PGL, pstree.TreeStyles["utf8"].PGL)

	rootCmd.SetUsageTemplate(usageTemplate)
}

// pstreePreRunCmd is executed before the main run command.
// This function is a hook provided by cobra that runs before the main command execution.
// It can be used for pre-execution setup tasks.
//
// Parameters:
//   - cmd: The command being executed
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
	if flagDebug {
		logger.Init(slog.LevelDebug)
	} else {
		logger.Init(slog.LevelInfo)
	}
	installedMemory, _ = util.GetTotalMemory()

	// Flag conflict rules
	//
	// 1. --user root cannot be used with --exclude-root
	// 2. only one of --color, --rainbow, and --color-attr can be used
	// 3. only one of --ibm-850, --utf-8, and --vt-100 can be use
	// 4. valid options for --color-attr are: cpu, mem
	// 5. only one of --uid-transitions and --user-transitions can be used
	// 6. --all and --user-transitions cannot be used together
	// 7. --all and --uid-transitions cannot be used together

	// Rule 1: --user root cannot be used with --exclude-root
	if slices.Contains(flagUsername, "root") && flagExcludeRoot {
		return errors.New("--user root and --exclude-root cannot be used together")
	}

	// Rule 2: only one of --color, --rainbow, and --color-attr can be used
	if (util.BtoI(flagColorize) + util.BtoI(flagRainbow) + util.StoI(flagColor)) > 1 {
		return errors.New("only one of --color, --colorize, and --rainbow can be used")
	}

	// Rule 3: only one of --ibm-850, --utf-8, and --vt-100 can be used
	if (util.BtoI(flagIBM850) + util.BtoI(flagUTF8) + util.BtoI(flagVT100)) > 1 {
		return errors.New("only one of --ibm-850, --utf-8, and --vt-100 can be used")
	}

	// Rule 4: valid options for --color-attr are: age, cpu, mem
	if flagColor != "" && !slices.Contains(validAttributes, flagColor) {
		return fmt.Errorf("valid options for --color are: %s", strings.Join(validAttributes, ", "))
	}

	// Rule 5: only one of --uid-transitions and --user-transitions can be used
	if (util.BtoI(flagShowUIDTransitions) + util.BtoI(flagShowUserTransitions)) > 1 {
		return errors.New("only one of --uid-transitions and --user-transitions can be used")
	}

	// Rule 6: --all and --user-transitions cannot be used together
	if flagShowAll && flagShowUserTransitions {
		return errors.New("--all and --user-transitions cannot be used together")
	}

	// Rule 7: --all and --uid-transitions cannot be used together
	if flagShowAll && flagShowUIDTransitions {
		return errors.New("--all and --uid-transitions cannot be used together")
	}

	if flagVersion {
		versionString = fmt.Sprintf(`pstree %s
Copyright (C) 2025 Gary Danko

pstree comes with ABSOLUTELY NO WARRANTY.
This is free software, and you are welcome to redistribute it under
the terms of the GNU General Public License.
For more information about these matters, see the files named COPYING.`,
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
	pstree.GetProcesses(logger.Logger, &processes)

	if flagOrderBy != "" {
		if !slices.Contains(validOrderBy, flagOrderBy) {
			errorMessage = fmt.Sprintf("valid options for --order-by are: %s", strings.Join(validOrderBy, ", "))
			return errors.New(errorMessage)
		}
		proc, err := pstree.GetProcessByPid(&processes, 1)
		if err != nil {
			panic(err)
		}
		sorted = []pstree.Process{proc}
		switch flagOrderBy {
		case "age":
			flagAge = true
			pstree.SortProcsByAge(&processes)
		case "cpu":
			flagCpu = true
			pstree.SortProcsByCpu(&processes)
		case "mem":
			flagMemory = true
			pstree.SortProcsByMemory(&processes)
		case "pid":
			flagShowPids = true
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

	pstree.MakeTree(logger.Logger, &processes)
	pstree.MarkProcs(logger.Logger, &processes, flagContains, flagUsername, flagExcludeRoot, flagPid)
	pstree.DropProcs(logger.Logger, &processes)

	if flagLevel == 0 {
		flagLevel = 999
	}

	// If any of the following flags are set, then compact mode should be disabled
	// This is because some of the results or offenders may be buried in collapsed subtrees
	if flagColor != "" || flagCpu || flagMemory || flagContains != "" {
		flagCompactNot = true
	}

	if flagShowAll {
		flagAge = true
		flagArguments = true
		flagCpu = true
		flagMemory = true
		flagShowOwner = true
		flagShowPgids = true
		flagShowPids = true
		flagShowUIDTransitions = true
		flagThreads = true
	}

	// If show UID transitions or show User transitions is enabled, mark processes with different UIDs from their parent
	if flagShowUIDTransitions || flagShowUserTransitions {
		pstree.MarkUIDTransitions(logger.Logger, &processes)
	}

	displayOptions = pstree.DisplayOptions{
		ColorAttr:           flagColor,
		ColorCount:          colorCount,
		ColorizeOutput:      flagColorize,
		ColorSupport:        colorSupport,
		CompactMode:         !flagCompactNot,
		GraphicsMode:        flagGraphicsMode,
		HidePGL:             !flagShowPGL,
		HideThreads:         flagHideThreads,
		IBM850Graphics:      flagIBM850,
		InstalledMemory:     installedMemory.Total,
		MaxDepth:            flagLevel,
		RainbowOutput:       flagRainbow,
		ShowArguments:       flagArguments,
		ShowCpuPercent:      flagCpu,
		ShowMemoryUsage:     flagMemory,
		ShowNumThreads:      flagThreads,
		ShowOwner:           flagShowOwner,
		ShowPGIDs:           flagShowPgids,
		ShowPids:            flagShowPids,
		ShowProcessAge:      flagAge,
		ShowUIDTransitions:  flagShowUIDTransitions,
		ShowUserTransitions: flagShowUserTransitions,
		UTF8Graphics:        flagUTF8,
		VT100Graphics:       flagVT100,
		WideDisplay:         flagWide,
	}
	pstree.PrintTree(logger.Logger, processes, 0, "", screenWidth, 0, &displayOptions)

	return nil
}
