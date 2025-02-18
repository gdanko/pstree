package util

import (
	"fmt"
	"regexp"
	"strings"
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
func colorMyText(red, green, blue int, text *string) {
	coloredText := fmt.Sprintf("\033[38;2;%d;%d;%dm%s\033[0m", red, green, blue, *text)
	*text = coloredText
}

func colorMyTextBold(red, green, blue int, text *string) {
	coloredText := fmt.Sprintf("\033[1;38;2;%d;%d;%dm%s\033[0m", red, green, blue, *text)
	*text = coloredText
}

func ColorBlack(text *string) {
	colorMyText(0, 0, 0, text)
}

func ColorBoldBlack(text *string) {
	colorMyTextBold(0, 0, 0, text)
}

func ColorBlue(text *string) {
	colorMyText(0, 0, 238, text)
}

func ColorBoldBlue(text *string) {
	colorMyTextBold(0, 0, 238, text)
}

func ColorCyan(text *string) {
	colorMyText(0, 205, 205, text)
}

func ColorBoldCyan(text *string) {
	colorMyTextBold(0, 205, 205, text)
}

func ColorGreen(text *string) {
	colorMyText(0, 205, 0, text)
}

func ColorBoldGreen(text *string) {
	colorMyTextBold(0, 205, 0, text)
}

func ColorOrange(text *string) {
	colorMyText(255, 128, 0, text)
}

func ColorBoldOrange(text *string) {
	colorMyTextBold(255, 128, 0, text)
}

func ColorPurple(text *string) {
	colorMyText(205, 0, 205, text)
}

func ColorBoldPurple(text *string) {
	colorMyTextBold(205, 0, 205, text)
}

func ColorRed(text *string) {
	colorMyText(205, 0, 0, text)
}

func ColoBoldRed(text *string) {
	colorMyTextBold(205, 0, 0, text)
}

func ColorWhite(text *string) {
	colorMyText(229, 229, 229, text)
}

func ColorBoldWhite(text *string) {
	colorMyTextBold(229, 229, 229, text)
}

func ColorYellow(text *string) {
	colorMyText(205, 205, 0, text)
}

func ColorBoldYellow(text *string) {
	colorMyTextBold(205, 205, 0, text)
}

func Color8() string {
	l1 := "c"
	l2 := "o"
	l3 := "l"
	l4 := "o"
	l5 := "r"
	ColorRed(&l1)
	ColorYellow(&l2)
	ColorGreen(&l3)
	ColorBlue(&l4)
	ColorPurple((&l5))

	return fmt.Sprintf("%s%s%s%s%s", l1, l2, l3, l4, l5)
}
