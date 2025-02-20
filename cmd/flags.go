package cmd

import (
	"fmt"
	"strings"

	"github.com/gdanko/pstree/util"
	"github.com/giancarlosio/gorainbow"
	"github.com/spf13/cobra"
)

func GetPersistentFlags(cmd *cobra.Command, colorSupport bool, colorCount int, username string) {
	cmd.PersistentFlags().BoolVarP(&flagArguments, "arguments", "a", false, "show command line arguments")
	cmd.PersistentFlags().BoolVarP(&flagIBM850, "ibm-850", "i", false, "use IBM-850 line drawing characters")
	cmd.PersistentFlags().BoolVarP(&flagUTF8, "utf-8", "u", false, "use UTF-8 (Unicode) line drawing characters")
	cmd.PersistentFlags().BoolVarP(&flagVT100, "vt-100", "v", false, "use VT-100 line drawing characters")
	cmd.PersistentFlags().BoolVarP(&flagShowAll, "all", "", false, "equivalent to -a --age -c -g -m -t")
	cmd.PersistentFlags().IntVarP(&flagLevel, "level", "l", 0, "print tree to <level> level deep")
	cmd.PersistentFlags().BoolVarP(&flagShowPgids, "pgid", "g", false, "show process group IDs")
	cmd.PersistentFlags().StringVarP(&flagUsername, "user", "", "", "show only branches containing processes of <user>; cannot be used with --exclude-root")
	cmd.PersistentFlags().BoolVarP(&flagCpu, "cpu", "c", false, "show CPU utilization percentage with each process, e.g., (c:0.00%)")
	cmd.PersistentFlags().BoolVarP(&flagAge, "age", "", false, "show the age of the process using the format (dd:hh:mm:ss)")
	cmd.PersistentFlags().BoolVarP(&flagThreads, "threads", "t", false, "show the number of threads with each process, e.g., (t:xx)")
	cmd.PersistentFlags().BoolVarP(&flagMemory, "memory", "m", false, "show the memory usage with each process, e.g., (m:x.y MiB)")
	cmd.PersistentFlags().BoolVarP(&flagExcludeRoot, "exclude-root", "U", false, "don't show branches containing only root processes; cannot be used with --user")
	cmd.PersistentFlags().StringVarP(&flagContains, "contains", "s", "", "show only branches containing processes with <pattern> in the command line")
	cmd.PersistentFlags().Int32VarP(&flagPid, "pid", "p", 0, "show only branches containing process <pid>")
	cmd.PersistentFlags().BoolVarP(&flagNoPids, "no-pids", "", false, "do not show process IDs")
	cmd.PersistentFlags().BoolVarP(&flagWide, "wide", "w", false, "wide output, not truncated to window width")
	if colorSupport {
		cmd.PersistentFlags().StringVarP(&flagColor, "color", "C", "", fmt.Sprintf("color the process name by given attribute; valid options are: %s;\ncannot be used with --color or --rainbow", strings.Join(validAttributes, ", ")))
		if colorCount >= 8 && colorCount < 256 {
			cmd.PersistentFlags().BoolVarP(&flagColorize, "colorize", "", false, fmt.Sprintf("add some %s to the output; cannot be used with --color-attr or --rainbow", util.Color8()))
		} else if colorCount >= 256 {
			cmd.PersistentFlags().BoolVarP(&flagColorize, "colorize", "", false, gorainbow.Rainbow("add some beautiful color to the pstree output; cannot be used with --color-attr or --rainbow"))
			cmd.PersistentFlags().BoolVarP(&flagRainbow, "rainbow", "", false, "please don't; cannot be used with --color or --color-attr")
		}
	}
	if username == "gdanko" || username == "gary.danko" {
		cmd.PersistentFlags().BoolVarP(&flagDebug, "debug", "d", false, "show debugging data")
	}
	cmd.PersistentFlags().BoolVarP(&flagVersion, "version", "V", false, "display version information")
}
