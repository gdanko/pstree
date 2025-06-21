package cmd

import (
	"fmt"
	"strings"

	"github.com/gdanko/pstree/util"
	"github.com/giancarlosio/gorainbow"
	"github.com/spf13/cobra"
)

// GetPersistentFlags configures all command-line flags for the pstree application.
//
// This function sets up all available command-line options including display options,
// filtering options, and formatting options. It handles conditional flags based on
// terminal capabilities (like color support) and user privileges.
//
// Parameters:
//   - cmd: The cobra command to which flags will be added
//   - colorSupport: Boolean indicating if the terminal supports color output
//   - colorCount: Integer representing the number of colors supported by the terminal
//   - username: String containing the current user's username for privilege-based flags
//
// GetPersistentFlags configures all command-line flags for the pstree application.
//
// This function sets up all available command-line options including display options,
// filtering options, and formatting options. It handles conditional flags based on
// terminal capabilities (like color support) and user privileges.
//
// Parameters:
//   - cmd: The cobra command to which flags will be added
//   - colorSupport: Boolean indicating if the terminal supports color
//   - colorCount: Integer representing the number of colors supported by the terminal
//   - username: String containing the current user's username for privilege-based flags
func GetPersistentFlags(cmd *cobra.Command, colorSupport bool, colorCount int, username string) {
	// Drawing characters
	cmd.PersistentFlags().BoolVarP(&flagIBM850, "ibm-850", "i", false, "use IBM-850 line drawing characters")
	cmd.PersistentFlags().BoolVarP(&flagUTF8, "utf-8", "u", false, "use UTF-8 (Unicode) line drawing characters")
	cmd.PersistentFlags().BoolVarP(&flagVT100, "vt-100", "v", false, "use VT-100 line drawing characters")

	// Depth
	cmd.PersistentFlags().IntVarP(&flagLevel, "level", "l", 0, "print tree to <level> level deep")

	// Width
	cmd.PersistentFlags().BoolVarP(&flagWide, "wide", "w", false, "wide output, not truncated to window width")

	// Color options
	if colorSupport {
		cmd.PersistentFlags().StringVarP(&flagColor, "color", "k", "", fmt.Sprintf("color the process name by given attribute; implies --compact-not; valid options are: %s;\ncannot be used with --colorize or --rainbow", strings.Join(validAttributes, ", ")))
		if colorCount >= 8 && colorCount < 256 {
			cmd.PersistentFlags().BoolVarP(&flagColorize, "colorize", "C", false, fmt.Sprintf("add some %s to the output; cannot be used with --color or --rainbow", util.Color8()))
		} else if colorCount >= 256 {
			cmd.PersistentFlags().BoolVarP(&flagColorize, "colorize", "C", false, gorainbow.Rainbow("add some beautiful color to the pstree output; cannot be used with --color or --rainbow"))
			cmd.PersistentFlags().BoolVarP(&flagRainbow, "rainbow", "r", false, "please don't; cannot be used with --color or --colorize")
		}
	}

	// Optional information
	cmd.PersistentFlags().BoolVarP(&flagShowAll, "all", "A", false, "equivalent to -a -c -g -G -m -O -p -t -I; cannot be used with --uid-transitions or --user-transitions")
	cmd.PersistentFlags().BoolVarP(&flagCompactNot, "compact-not", "n", false, "do not compact identical subtrees in output")
	cmd.PersistentFlags().BoolVarP(&flagCpu, "cpu", "c", false, "show CPU utilization percentage with each process, e.g., (c:0.00%); implies --compact-not")
	cmd.PersistentFlags().BoolVarP(&flagMemory, "memory", "m", false, "show the memory usage with each process, e.g., (m:x.y MiB); implies --compact-not")
	cmd.PersistentFlags().BoolVarP(&flagShowOwner, "show-owner", "O", false, "show the owner of the process")
	cmd.PersistentFlags().BoolVarP(&flagShowPgids, "pgid", "g", false, "show process group IDs")
	cmd.PersistentFlags().BoolVarP(&flagShowPids, "show-pids", "p", false, "show PIDs")
	cmd.PersistentFlags().BoolVarP(&flagShowUIDTransitions, "uid-transitions", "I", false, "show processes where the user ID changes from the parent process, e.g., (uid→uid); cannot be used with --user-transitions")
	cmd.PersistentFlags().BoolVarP(&flagShowUserTransitions, "user-transitions", "U", false, "show processes where the user changes from the parent process, e.g., (user→user); cannot be used with --uid-transitions or --all")
	cmd.PersistentFlags().BoolVarP(&flagThreads, "threads", "t", false, "show the number of threads with each process, e.g., (t:xx)")
	cmd.PersistentFlags().BoolVarP(&flagHideThreads, "hide-threads", "", false, "hide threads, show only processes")

	// Filtering and sorting
	cmd.PersistentFlags().BoolVarP(&flagAge, "age", "G", false, "show the age of the process using the format (dd:hh:mm:ss)")
	cmd.PersistentFlags().BoolVarP(&flagArguments, "arguments", "a", false, "show command line arguments")
	cmd.PersistentFlags().BoolVarP(&flagExcludeRoot, "exclude-root", "X", false, "don't show branches containing only root processes; cannot be used with --user")
	cmd.PersistentFlags().Int32VarP(&flagPid, "pid", "P", 0, "show only branches containing process <pid>")
	cmd.PersistentFlags().StringSliceVarP(&flagUsername, "user", "", []string{}, "show only branches containing processes of <user>; this option can be used more than and cannot be used with --exclude-root")
	cmd.PersistentFlags().StringVarP(&flagContains, "contains", "s", "", "show only branches containing processes with <pattern> in the command line; implies --compact-not")
	cmd.PersistentFlags().StringVarP(&flagOrderBy, "order-by", "o", "", fmt.Sprintf("sort the results by <field>; valid options are: %s", strings.Join(validOrderBy, ", ")))

	// Debugging
	if username == "gdanko" || username == "gary.danko" {
		cmd.PersistentFlags().BoolVarP(&flagDebug, "debug", "d", false, "show debugging data")
	}

	// Miscellaneous
	cmd.PersistentFlags().BoolVarP(&flagVersion, "version", "V", false, "display version information")
}
