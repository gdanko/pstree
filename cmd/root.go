package cmd

import (
	"github.com/gdanko/pstree/pkg/pstree"
	"github.com/kr/pretty"
	"github.com/spf13/cobra"
)

var (
	err             error
	flagArguments   bool
	flagContains    string
	flagFile        string
	flagLevel       int
	flagStart       int32
	flagUsername    string
	flagVersion     bool
	flagVersionFull string
	flagWide        bool
	tree            map[int32][]int32
	rootCmd         = &cobra.Command{
		Use:    "pstree",
		Short:  "Display a tree of processes",
		Long:   "Display a tree of processes",
		PreRun: pstreePreRunCmd,
		Run:    pstreeRunCmd,
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	GetPersistentFlags(rootCmd)
}

func pstreePreRunCmd(cmd *cobra.Command, args []string) {
}

func pstreeRunCmd(cmd *cobra.Command, args []string) {
	if flagFile != "" {
		tree, err = pstree.GetTreeDataFromFile(flagFile, flagUsername, flagContains, flagLevel)
	} else {
		tree = pstree.GetTreeData(flagUsername, flagContains, flagLevel)
	}
	pretty.Println(tree)

}
