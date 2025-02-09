package cmd

import (
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "pstree",
		Short: "pstree",
		Long:  "pstree",
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {

}
