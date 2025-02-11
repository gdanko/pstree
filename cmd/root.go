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
	colorCount      int
	colorizeString  string = ""
	colorSupport    bool
	err             error
	flagArguments   bool
	flagAscii       bool
	flagColorize    bool
	flagContains    string
	flagExcludeRoot bool
	flagFile        string
	flagLevel       int
	flagPid         int32
	flagShowPids    bool
	flagUsername    string
	flagVersion     bool
	flagWide        bool
	initialIndent   string = ""
	screenWidth     int
	startingPid     int32
	tree            map[int32][]int32
	usageTemplate   string
	version         string = "0.2.1"
	versionString   string
	rootCmd         = &cobra.Command{
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
		`Usage: pstree [-aAUpw]%s [-f file] [-l n] [--show-pids] 
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

	if flagFile != "" {
		tree, err = pstree.GetTreeDataFromFile(flagFile, flagUsername, flagContains, flagLevel)
	} else {
		tree = pstree.GetTreeData(flagUsername, flagContains, flagLevel, flagExcludeRoot)
		// tree, err = pstree.GetTreeDataFromPs(flagUsername, flagContains, flagLevel)
	}

	screenWidth = util.GetScreenWidth()

	startingPid = pstree.FindFirstPid(tree)
	if flagPid > 0 {
		startingPid = flagPid
	}
	pstree.GenerateTree(startingPid, tree, "", "", initialIndent, flagArguments, flagWide, flagShowPids, flagAscii, flagColorize, screenWidth)

	return nil
}
