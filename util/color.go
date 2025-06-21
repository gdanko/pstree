package util

import (
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
)

var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// visibleWidth calculates the display width of a string containing ANSI escape sequences.
// It ignores ANSI escape sequences and counts only the visible characters' width.
// The function properly handles multi-byte Unicode characters and characters with
// different display widths (like CJK characters that take up 2 columns).
//
// Parameters:
//   - input: The string to calculate the width for, which may contain ANSI escape sequences
//
// Returns:
//
//	The display width of the string, excluding ANSI escape sequences
func visibleWidth(input string) int {
	width := 0
	for len(input) > 0 {
		if loc := ansiEscape.FindStringIndex(input); loc != nil && loc[0] == 0 {
			// Skip ANSI
			input = input[loc[1]:]
			continue
		}
		r, size := utf8.DecodeRuneInString(input)
		width += runewidth.RuneWidth(r)
		input = input[size:]
	}
	return width
}

// TruncateANSI truncates a string containing ANSI escape sequences to fit within a specified screen width.
// It preserves ANSI color and formatting codes while only counting visible characters toward the width limit.
//
// Parameters:
//   - logger: A structured logger for debug output
//   - input: The string to truncate, which may contain ANSI escape sequences
//   - screenWidth: The maximum width (in visible characters) the output string should occupy
//
// The function handles multi-byte Unicode characters correctly by using utf8.DecodeRuneInString
// and accounts for characters with different display widths using the runewidth package.
// If truncation occurs, "..." is appended to the result.
//
// Returns:
//
//	A string that fits within screenWidth, with ANSI sequences preserved.
func TruncateANSI(logger *slog.Logger, input string, screenWidth int) string {
	dots := "..."

	if screenWidth <= 3 {
		return dots
	}

	// First, check actual display width
	if visibleWidth(input) <= screenWidth {
		return input // No truncation needed
	}

	targetWidth := screenWidth - len(dots)
	var output strings.Builder
	width := 0

	for len(input) > 0 {
		if loc := ansiEscape.FindStringIndex(input); loc != nil && loc[0] == 0 {
			esc := input[loc[0]:loc[1]]
			output.WriteString(esc)
			input = input[loc[1]:]
			continue
		}

		r, size := utf8.DecodeRuneInString(input)
		rw := runewidth.RuneWidth(r)

		if width+rw > targetWidth {
			break
		}

		output.WriteRune(r)
		width += rw
		input = input[size:]
	}

	output.WriteString(dots)
	return output.String()
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

func ColorGray(text *string) {
	colorMyText(128, 128, 128, text)
}

func ColorBoldGray(text *string) {
	colorMyTextBold(128, 128, 128, text)
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

func ColorMagenta(text *string) {
	colorMyText(205, 0, 205, text)
}

func ColorBoldMagenta(text *string) {
	colorMyTextBold(205, 0, 205, text)
}

func ColorRed(text *string) {
	colorMyText(205, 0, 0, text)
}

func ColorBoldRed(text *string) {
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
