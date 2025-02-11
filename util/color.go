package util

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/fatih/color"
)

var (
	colorize   string
	ansiEscape = regexp.MustCompile(`\x1B(?:[@-Z\\-_]|\[[0-?]*[ -/]*[@-~])`)
)

func visibleLength(text string) int {
	return len(ansiEscape.ReplaceAllString(text, ""))
}

// truncateANSI truncates an ANSI-colored string to fit within a max width.
func TruncateANSI(text string, maxWidth int) string {
	if visibleLength(text) <= maxWidth {
		return text // No need to truncate
	}

	var result strings.Builder
	visibleChars := 0
	i := 0

	for i < len(text) {
		if loc := ansiEscape.FindStringIndex(text[i:]); loc != nil && loc[0] == 0 {
			// If an ANSI sequence is found at the current position, keep it.
			result.WriteString(text[i : i+loc[1]])
			i += loc[1]
		} else {
			// Otherwise, process visible characters
			if visibleChars >= maxWidth {
				break
			}
			result.WriteByte(text[i])
			visibleChars++
			i++
		}
	}

	// Append reset code to prevent ANSI bleed
	return result.String() + "\x1b[0m"
}

func ColorBlack(text string) string {
	red := color.New(color.FgBlack).SprintFunc()
	return red(text)
}

func ColorBlue(text string) string {
	red := color.New(color.FgBlue).SprintFunc()
	return red(text)
}

func ColorCyan(text string) string {
	red := color.New(color.FgCyan).SprintFunc()
	return red(text)
}

func ColorGreen(text string) string {
	red := color.New(color.FgGreen).SprintFunc()
	return red(text)
}

func ColorPurple(text string) string {
	red := color.New(color.FgMagenta).SprintFunc()
	return red(text)
}

func ColorRed(text string) string {
	red := color.New(color.FgRed).SprintFunc()
	return red(text)
}

func ColorWhite(text string) string {
	red := color.New(color.FgWhite).SprintFunc()
	return red(text)
}

func ColorYellow(text string) string {
	red := color.New(color.FgYellow).SprintFunc()
	return red(text)
}

func Color8() string {
	return fmt.Sprintf(
		"%s%s%s%s%s",
		ColorRed("c"),
		ColorYellow("o"),
		ColorGreen("l"),
		ColorBlue("o"),
		ColorPurple("r"),
	)
}
