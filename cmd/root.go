package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/gdanko/pstree/pkg/pstree"
	"github.com/gdanko/pstree/util"
	"github.com/spf13/cobra"
)

var (
	colorCount       int
	colorString      string = ""
	colorSupport     bool
	currentLevel     int = 0
	displayOptions   pstree.DisplayOptions
	err              error
	flagArguments    bool
	flagColor        bool
	flagContains     string
	flagCpu          bool
	flagExcludeRoot  bool
	flagFile         string
	flagGraphicsMode int
	flagLevel        int
	flagMemory       bool
	flagNoPids       bool
	flagPid          int32
	flagRainbow      bool
	flagThreads      bool
	flagUsername     string
	flagVersion      bool
	flagWide         bool
	initialIndent    string = ""
	processes        []pstree.Process
	rainbowString    string = ""
	screenWidth      int
	startingPidIndex int
	usageTemplate    string
	version          string = "0.4.3"
	versionString    string
	rootCmd          = &cobra.Command{
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
	colorSupport, colorCount := util.HasColorSupport()
	if colorSupport {
		colorString = " [--color]"
		rainbowString = " [--rainbow]"
	}
	usageTemplate = fmt.Sprintf(
		`Usage: pstree [-acUmntw]%s [-s, --contains <str>] [-l, --level <int>]
	      [-g, --mode <int>] [-p, --pid <int>]%s [-u, --user <str>] 
   or: pstree -V

Display a tree of processes.

{{.Flags.FlagUsages}}
Process group leaders are marked with '='.
`, colorString, rainbowString)

	GetPersistentFlags(rootCmd, colorSupport, colorCount)
	rootCmd.SetUsageTemplate(usageTemplate)
}

func pstreePreRunCmd(cmd *cobra.Command, args []string) {
}

func pstreeRunCmd(cmd *cobra.Command, args []string) error {
	if flagUsername != "" && flagExcludeRoot {
		return errors.New("flags --user and --exclude-root are mutually exclusive")
	}

	if flagColor && flagRainbow {
		return errors.New("flags --color and --rainbow are mutually exclusive")
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

	displayOptions = pstree.DisplayOptions{
		ColorizeOutput:  flagColor,
		GraphicsMode:    flagGraphicsMode,
		HidePids:        flagNoPids,
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
