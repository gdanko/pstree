package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

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
	flagArguments         bool
	flagColor             string
	flagColorize          bool
	flagContains          string
	flagCpu               bool
	flagExcludeRoot       bool
	flagGraphicsMode      int
	flagLevel             int
	flagMemory            bool
	flagNoPids            bool
	flagPid               int32
	flagRainbow           bool
	flagShowAll           bool
	flagThreads           bool
	flagUsername          string
	flagVersion           bool
	flagWide              bool
	installedMemory       *mem.VirtualMemoryStat
	processes             []pstree.Process
	rainbowString         string = ""
	screenWidth           int
	startingPidIndex      int
	usageTemplate         string
	validAttributes       []string = []string{"age", "cpu", "mem"}
	validAttributesString string   = strings.Join(validAttributes, ", ")
	version               string   = "0.5.3"
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
	colorSupport, colorCount = util.HasColorSupport()
	if colorSupport {
		colorizeString = " [--colorize]"
		rainbowString = " [--rainbow]"
		colorString = "[-C, --color-attr <str>]"

	}
	usageTemplate = fmt.Sprintf(
		`Usage: pstree [-acUmntw] [-all]%s [-s, --contains <str>] [-l, --level <int>]
	      [-g, --mode <int>] [-p, --pid <int>]%s [-u, --user <str>] 
	      %s
   or: pstree -V

Display a tree of processes.

{{.Flags.FlagUsages}}
Process group leaders are marked with '='.
`, colorizeString, rainbowString, colorString)

	GetPersistentFlags(rootCmd, colorSupport, colorCount)
	rootCmd.SetUsageTemplate(usageTemplate)
}

func pstreePreRunCmd(cmd *cobra.Command, args []string) {
}

func pstreeRunCmd(cmd *cobra.Command, args []string) error {
	installedMemory, _ = util.GetTotalMemory()

	if flagUsername != "" && flagExcludeRoot {
		return errors.New("flags --user and --exclude-root are mutually exclusive")
	}

	if (util.BtoI(flagColorize) + util.BtoI(flagRainbow) + util.StoI(flagColor)) > 1 {
		return errors.New("only one of --color, --rainbow, and --color-attr can be used")
	}

	if flagColor != "" && !util.Contains(validAttributes, flagColor) {
		errorMessage = fmt.Sprintf("valid options for --color-attr are: %s", validAttributesString)
		return errors.New(errorMessage)
	}

	if flagVersion {
		versionString = fmt.Sprintf(`pstree %s
Copyright (C) 2024 Gary Danko

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
	pstree.GetProcesses(&processes)
	pstree.MakeTree(&processes)
	pstree.MarkProcs(&processes, flagContains, flagUsername, flagExcludeRoot, flagPid)
	pstree.DropProcs(&processes)

	if flagPid > 1 {
		startingPidIndex = pstree.GetPIDIndex(processes, flagPid)
		if startingPidIndex == -1 {
			fmt.Printf("PID %d does not exist.\n", flagPid)
			os.Exit(1)
		}
	}

	if flagUsername != "" {
		if !util.UserExists(flagUsername) {
			fmt.Printf("User '%s' does not exist.\n", flagUsername)
			os.Exit(1)
		}
	}

	if flagLevel == 0 {
		flagLevel = 100
	}

	if flagShowAll {
		flagArguments = true
		flagCpu = true
		flagMemory = true
		flagThreads = true
	}

	displayOptions = pstree.DisplayOptions{
		ColorAttr:       flagColor,
		ColorizeOutput:  flagColorize,
		GraphicsMode:    flagGraphicsMode,
		HidePids:        flagNoPids,
		InstalledMemory: installedMemory.Total,
		MaxDepth:        flagLevel,
		RainbowOutput:   flagRainbow,
		ShowArguments:   flagArguments,
		ShowCpuPercent:  flagCpu,
		ShowMemoryUsage: flagMemory,
		ShowNumThreads:  flagThreads,
		WideDisplay:     flagWide,
	}
	pstree.PrintTree(processes, startingPidIndex, "", screenWidth, currentLevel, displayOptions)

	return nil
}
