package cmd

import (
	"fmt"

	"github.com/gdanko/pstree/util"
	"github.com/giancarlosio/gorainbow"
	"github.com/spf13/cobra"
)

func GetPersistentFlags(cmd *cobra.Command, colorSupport bool, colorCount int) {
	cmd.PersistentFlags().BoolVarP(&flagArguments, "arguments", "a", false, "show command line arguments")
	cmd.PersistentFlags().IntVarP(&flagGraphicsMode, "mode", "g", 0, "use graphics chars for tree. n=1: IBM-850, n=2: VT100, n=3: UTF-8")
	cmd.PersistentFlags().IntVarP(&flagLevel, "level", "l", 0, "print tree to <depth> level deep")
	cmd.PersistentFlags().StringVarP(&flagUsername, "user", "u", "", "show only branches containing processes of <user>; cannot be used with --exclude-root")
	cmd.PersistentFlags().BoolVarP(&flagExcludeRoot, "exclude-root", "U", false, "don't show branches containing only root processes; cannot be used with --user")
	cmd.PersistentFlags().StringVarP(&flagContains, "contains", "c", "", "show only branches containing process with <string> in commandline")
	cmd.PersistentFlags().Int32VarP(&flagPid, "pid", "", 0, "show only branches containing process <pid>")
	cmd.PersistentFlags().BoolVarP(&flagShowPids, "show-pids", "p", false, "show PIDs")
	cmd.PersistentFlags().BoolVarP(&flagWide, "wide", "w", false, "wide output, not truncated to window width")
	if colorSupport {
		if colorCount >= 8 && colorCount < 256 {
			cmd.PersistentFlags().BoolVarP(&flagColorize, "colorize", "", false, fmt.Sprintf("add some %s to the output", util.Color8()))
		} else if colorCount >= 256 {
			cmd.PersistentFlags().BoolVarP(&flagColorize, "colorize", "", false, gorainbow.Rainbow("add some beautiful color to the pstree output"))
		}
	}
	cmd.PersistentFlags().BoolVarP(&flagVersion, "version", "V", false, "display version information")
}
