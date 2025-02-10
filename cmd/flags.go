package cmd

import (
	"github.com/spf13/cobra"
)

func GetPersistentFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&flagFile, "file", "f", "", "Read input from <file> (- is stdin). File format must\nbe the output of \"ps -axwwo user,pid,ppid,pgid,command\".")
	cmd.PersistentFlags().BoolVarP(&flagArguments, "arguments", "a", false, "Show command line arguments.")
	cmd.PersistentFlags().BoolVarP(&flagVersion, "version", "V", false, "Display version information.")
	cmd.PersistentFlags().BoolVarP(&flagWide, "wide", "w", false, "Wide output, not truncated to window width.")
	cmd.PersistentFlags().StringVarP(&flagUsername, "user", "u", "", "Show only branches containing processes of <user>.")
	cmd.PersistentFlags().Int32VarP(&flagStart, "start", "", 0, "Start at PID <start>.")
	cmd.PersistentFlags().IntVarP(&flagLevel, "level", "l", 0, "Print tree to <depth> level deep.")
	cmd.PersistentFlags().StringVarP(&flagContains, "", "s", "", "Show only branches containing process with <string> in commandline.")
}
