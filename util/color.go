package util

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/fatih/color"
)

var (
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
	out := color.New(color.FgBlack).SprintFunc()
	return out(text)
}

func ColorBoldBlack(text string) string {
	out := color.New(color.FgBlack).Add(color.Bold).SprintFunc()
	return out(text)
}

func ColorBlue(text string) string {
	out := color.New(color.FgBlue).SprintFunc()
	return out(text)
}

func ColorBoldBlue(text string) string {
	out := color.New(color.FgBlue).Add(color.Bold).SprintFunc()
	return out(text)
}

func ColorCyan(text string) string {
	out := color.New(color.FgCyan).SprintFunc()
	return out(text)
}

func ColorBoldCyan(text string) string {
	out := color.New(color.FgCyan).Add(color.Bold).SprintFunc()
	return out(text)
}

func ColorGreen(text string) string {
	out := color.New(color.FgGreen).SprintFunc()
	return out(text)
}

func ColorBoldGreen(text string) string {
	out := color.New(color.FgGreen).Add(color.Bold).SprintFunc()
	return out(text)
}

func ColorOrange(text string) string {
	out := color.RGB(255, 128, 0).SprintFunc()
	return out(text)
}

func ColorBoldOrange(text string) string {
	out := color.RGB(255, 128, 0).Add(color.Bold).SprintFunc()
	return out(text)
}

func ColorPurple(text string) string {
	out := color.New(color.FgMagenta).SprintFunc()
	return out(text)
}

func ColorBoldPurple(text string) string {
	out := color.New(color.FgMagenta).Add(color.Bold).SprintFunc()
	return out(text)
}

func ColorRed(text string) string {
	out := color.New(color.FgRed).SprintFunc()
	return out(text)
}

func ColorBoldRed(text string) string {
	out := color.New(color.FgRed).Add(color.Bold).SprintFunc()
	return out(text)
}

func ColorWhite(text string) string {
	out := color.New(color.FgWhite).SprintFunc()
	return out(text)
}

func ColorBoldWhite(text string) string {
	out := color.New(color.FgWhite).Add(color.Bold).SprintFunc()
	return out(text)
}

func ColorYellow(text string) string {
	out := color.New(color.FgYellow).SprintFunc()
	return out(text)
}

func ColorBoldYellow(text string) string {
	out := color.New(color.FgYellow).Add(color.Bold).SprintFunc()
	return out(text)
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
