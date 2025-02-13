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
	colorizeString   string = ""
	colorSupport     bool
	err              error
	flagArguments    bool
	flagColorize     bool
	flagContains     string
	flagExcludeRoot  bool
	flagFile         string
	flagGraphicsMode int
	flagLevel        int
	flagPid          int32
	flagShowPids     bool
	flagUsername     string
	flagVersion      bool
	flagWide         bool
	initialIndent    string = ""
	processes        []pstree.Process
	screenWidth      int
	startingPidIndex int
	usageTemplate    string
	version          string = "0.2.1"
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
		colorizeString = " [--colorize]"
	}
	usageTemplate = fmt.Sprintf(
		`Usage: pstree [-aUpw] [-g n]%s [-l n] [--show-pids] 
	      [--pid n] [-u user] [-c string]
   or: pstree -V

Display a tree of processes.

{{.Flags.FlagUsages}}
Process group leaders are marked with '='.
`, colorizeString)

	GetPersistentFlags(rootCmd, colorSupport, colorCount)
	rootCmd.SetUsageTemplate(usageTemplate)
}

func pstreePreRunCmd(cmd *cobra.Command, args []string) {
}

func pstreeRunCmd(cmd *cobra.Command, args []string) error {
	if flagUsername != "" && flagExcludeRoot {
		return errors.New("flags --user and --exclude-root are mutually exclusive")
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

	// when --pid is used, we want to print `-+= 1 root /sbin/launchd``

	if flagPid > 1 {
		startingPidIndex = pstree.GetPIDIndex(processes, flagPid)
		if startingPidIndex <= 1 {
			startingPidIndex = pstree.GetPIDIndex(processes, 1)
		}
	}

	pstree.PrintTree(processes, startingPidIndex, "", screenWidth, flagArguments, flagShowPids, flagGraphicsMode, flagWide, flagColorize)

	return nil
}
