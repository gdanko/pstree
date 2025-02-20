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
	colorCount            int
	colorizeString        string = ""
	colorString           string = ""
	colorSupport          bool
	currentLevel          int = 0
	displayOptions        pstree.DisplayOptions
	errorMessage          string
	flagAge               bool
	flagArguments         bool
	flagColor             string
	flagColorize          bool
	flagContains          string
	flagCpu               bool
	flagDebug             bool
	flagExcludeRoot       bool
	flagGraphicsMode      int
	flagIBM850            bool
	flagLevel             int
	flagMemory            bool
	flagPid               int32
	flagRainbow           bool
	flagShowAll           bool
	flagShowPgids         bool
	flagNoPids            bool
	flagThreads           bool
	flagUsername          []string
	flagUTF8              bool
	flagVersion           bool
	flagVT100             bool
	flagWide              bool
	installedMemory       *mem.VirtualMemoryStat
	processes             []pstree.Process
	rainbowString         string = ""
	screenWidth           int
	usageTemplate         string
	username              string
	validAttributes       []string = []string{"age", "cpu", "mem"}
	validAttributesString string   = strings.Join(validAttributes, ", ")
	version               string   = "0.6.2"
	versionString         string
	rootCmd               = &cobra.Command{
		Use:    "pstree",
		Short:  "",
		Long:   "",
		PreRun: pstreePreRunCmd,
		RunE:   pstreeRunCmd,
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	username = util.DetermineUsername()
	colorSupport, colorCount = util.HasColorSupport()
	if colorSupport {
		colorizeString = " [--colorize]"
		rainbowString = " [--rainbow]"
		colorString = " [-C, --color <attr>]"

	}
	usageTemplate = fmt.Sprintf(
		`Usage: pstree [-acUimgtuvw] [--age] [-all]%s%s
          [-s, --contains <pattern>] [-l, --level <level>]
          [--no-pids] [-p, --pid <pid>]%s [--user <user> ...]
   or: pstree -V

Display a tree of processes.

{{.Flags.FlagUsages}}
Process group leaders are marked with '='.
`, colorString, colorizeString, rainbowString)

	GetPersistentFlags(rootCmd, colorSupport, colorCount, username)
	rootCmd.SetUsageTemplate(usageTemplate)
}

func pstreePreRunCmd(cmd *cobra.Command, args []string) {
}

func pstreeRunCmd(cmd *cobra.Command, args []string) error {
	if flagDebug {
		logger.Init(slog.LevelDebug)
	} else {
		logger.Init(slog.LevelInfo)
	}
	installedMemory, _ = util.GetTotalMemory()

	if slices.Contains(flagUsername, "root") && flagExcludeRoot {
		fmt.Fprintln(os.Stdout, "why would you do that?")
		os.Exit(1)
	}

	if (util.BtoI(flagColorize) + util.BtoI(flagRainbow) + util.StoI(flagColor)) > 1 {
		return errors.New("only one of --color, --rainbow, and --color-attr can be used")
	}

	if (util.BtoI(flagIBM850) + util.BtoI(flagUTF8) + util.BtoI(flagVT100)) > 1 {
		return errors.New("only one of --ibm-850, --utf-8, and --vt-100 can be used")
	}

	if flagColor != "" && !util.Contains(validAttributes, flagColor) {
		errorMessage = fmt.Sprintf("valid options for --color-attr are: %s", validAttributesString)
		return errors.New(errorMessage)
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

	screenWidth = util.GetScreenWidth()
	pstree.GetProcesses(logger.Logger, &processes)
	pstree.MakeTree(logger.Logger, &processes)
	pstree.MarkProcs(logger.Logger, &processes, flagContains, flagUsername, flagExcludeRoot, flagPid)
	pstree.DropProcs(logger.Logger, &processes)

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

	if flagLevel == 0 {
		flagLevel = 100
	}

	if flagShowAll {
		flagAge = true
		flagArguments = true
		flagCpu = true
		flagMemory = true
		flagShowPgids = true
		flagThreads = true
	}

	displayOptions = pstree.DisplayOptions{
		ColorAttr:       flagColor,
		ColorizeOutput:  flagColorize,
		GraphicsMode:    flagGraphicsMode,
		IBM850Graphics:  flagIBM850,
		InstalledMemory: installedMemory.Total,
		MaxDepth:        flagLevel,
		NoPids:          flagNoPids,
		RainbowOutput:   flagRainbow,
		ShowArguments:   flagArguments,
		ShowCpuPercent:  flagCpu,
		ShowMemoryUsage: flagMemory,
		ShowNumThreads:  flagThreads,
		ShowPGIDs:       flagShowPgids,
		ShowProcessAge:  flagAge,
		UTF8Graphics:    flagUTF8,
		VT100Graphics:   flagVT100,
		WideDisplay:     flagWide,
	}
	pstree.PrintTree(logger.Logger, processes, 0, "", screenWidth, currentLevel, displayOptions)

	return nil
}
