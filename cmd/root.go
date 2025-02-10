package cmd

import (
	"fmt"
	"os"

	"github.com/gdanko/pstree/pkg/pstree"
	"github.com/spf13/cobra"
)

var (
	err           error
	flagArguments bool
	flagAscii     bool
	flagContains  string
	flagFile      string
	flagLevel     int
	flagPid       int32
	flagShowPids  bool
	flagStart     int32
	flagUsername  string
	flagVersion   bool
	flagWide      bool
	initialIndent string = ""
	startingPid   int32
	versionString string
	version       string = "0.1.0"
	tree          map[int32][]int32
	rootCmd       = &cobra.Command{
		Use:    "pstree",
		Short:  "",
		Long:   "",
		PreRun: pstreePreRunCmd,
		Run:    pstreeRunCmd,
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	GetPersistentFlags(rootCmd)
	rootCmd.SetUsageTemplate(`Usage: pstree [-aApw] [-f file] [-l n] [--show-pids] [--start n]
	      [-u user] [-c string]
   or: pstree -V

Display a tree of processes.

{{.Flags.FlagUsages}}
Process group leaders are marked with '='.
`)
}

func pstreePreRunCmd(cmd *cobra.Command, args []string) {
}

func pstreeRunCmd(cmd *cobra.Command, args []string) {
	// if flagFile != "" {
	// 	tree, err = pstree.GetTreeDataFromFile(flagFile, flagUsername, flagContains, flagLevel)
	// } else {
	// 	tree = pstree.GetTreeData(flagUsername, flagContains, flagLevel)
	// }

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

	tree, err = pstree.GetTreeDataFromPs(flagUsername, flagContains, flagLevel)

	startingPid = pstree.FindFirstPid(tree)
	if flagStart > 0 {
		startingPid = flagStart
	}

	pstree.GenerateTree(startingPid, tree, "", "", initialIndent, flagArguments, flagWide, flagShowPids, flagAscii)
}
