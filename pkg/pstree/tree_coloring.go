package pstree

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
)

//------------------------------------------------------------------------------
// DISPLAY FORMATTING AND STYLING
//------------------------------------------------------------------------------
// Functions in this section handle the visual styling of the process tree,
// including colorization, width calculation, and text truncation.

// ColorizeField applies appropriate color formatting to a specific field in the process tree output.
//
// This method enhances the visual representation of the process tree by applying colors
// to different elements based on the current display options. It supports two main coloring modes:
//
// 1. Standard colorization (--colorize flag): Each field type gets a predefined color
//
//   - Username: Blue
//
//   - Command: Blue
//
//   - Arguments: Red
//
//   - PID/PGID: Purple
//
//   - CPU: Yellow
//
//   - Memory: Orange
//
//   - Age: Bold Green
//
//   - Threads: Bold White
//
//   - Tree characters: Green
//
//     2. Attribute-based colorization (--color flag): Colors are applied based on process attributes
//     like CPU or memory usage, with thresholds determining the color (green/yellow/red)
//
// The colors help to quickly identify important information in the tree, such as high
// resource usage processes or specific elements of interest.
//
// Parameters:
//   - fieldName: String identifying which field is being colored (e.g., "cpu", "memory", "command")
//   - value: Pointer to the string value to be colored (modified in-place)
//   - pidIndex: Index of the process to be colored
//
// The method uses the hybrid approach data (combining gopsutil with direct ps command calls)
// when applying attribute-based coloring, ensuring accurate thresholds for CPU and memory usage.
// colorizeField applies appropriate color formatting to a specific field in the process tree output.
//
// This method enhances the visual representation of the process tree by applying colors
// to different elements based on the current display options. It supports two main coloring modes:
//
//  1. Standard colorization (--colorize flag): Each field type gets a predefined color
//  2. Attribute-based colorization (--color flag): Colors are applied based on process attributes
//     like CPU or memory usage, with thresholds determining the color (green/yellow/red)
//
// Parameters:
//   - fieldName: String identifying which field is being colored (e.g., "cpu", "memory", "command")
//   - value: Pointer to the string value to be colored (modified in-place)
//   - pidIndex: Index of the process to be colored
//
// Refactoring opportunity: This function could be split into:
// - applyStandardColors: Apply standard color scheme
// - applyAttributeBasedColors: Apply colors based on attribute thresholds
func (processTree *ProcessTree) colorizeField(fieldName string, value *string, pidIndex int) {
	var (
		process *Process
	)
	// Only apply colors if the terminal supports them
	if processTree.DisplayOptions.ColorSupport {
		// Standard colorization mode (--colorize flag)
		if processTree.DisplayOptions.ColorizeOutput {
			// Apply specific colors based on the field type
			switch fieldName {
			case "age":
				processTree.Colorizer.Age(processTree.ColorScheme, value)
			case "args":
				processTree.Colorizer.Args(processTree.ColorScheme, value)
			case "connector":
				processTree.Colorizer.Connector(processTree.ColorScheme, value)
			case "command":
				processTree.Colorizer.Command(processTree.ColorScheme, value)
			case "compactStr":
				processTree.Colorizer.CompactStr(processTree.ColorScheme, value)
			case "cpu":
				processTree.Colorizer.CPU(processTree.ColorScheme, value)
			case "memory":
				processTree.Colorizer.Memory(processTree.ColorScheme, value)
			case "owner":
				processTree.Colorizer.Owner(processTree.ColorScheme, value)
			case "ownerTransition":
				processTree.Colorizer.OwnerTransition(processTree.ColorScheme, value)
			case "pidPgid":
				processTree.Colorizer.PIDPGID(processTree.ColorScheme, value)
			// case "prefix":
			// 	processTree.Colorizer.Prefix(processTree.ColorScheme, value)
			case "threads":
				processTree.Colorizer.NumThreads(processTree.ColorScheme, value)
			}
		} else if processTree.DisplayOptions.ColorAttr != "" {
			// Attribute-based colorization mode (--color flag)
			// Don't apply attribute-based coloring to the tree prefix
			if fieldName != "prefix" {
				process = &processTree.Nodes[pidIndex]
				switch processTree.DisplayOptions.ColorAttr {
				case "age":
					// Ensure process age is shown when coloring by age
					processTree.DisplayOptions.ShowProcessAge = true

					// Apply color based on process age thresholds in seconds
					if process.Age < 60 {
						// Low age (< 1 minute)
						processTree.Colorizer.ProcessAgeLow(processTree.ColorScheme, value)
					} else if process.Age >= 60 && process.Age < 3600 {
						// Medium age (< 1 hour)
						processTree.Colorizer.ProcessAgeMedium(processTree.ColorScheme, value)
					} else if process.Age >= 3600 && process.Age < 86400 {
						// High age (> 1 hour and < 1 day)
						processTree.Colorizer.ProcessAgeHigh(processTree.ColorScheme, value)
					} else if process.Age >= 86400 {
						// Very high age (> 1 day)
						processTree.Colorizer.ProcessAgeVeryHigh(processTree.ColorScheme, value)
					}
				case "cpu":
					// Ensure CPU percentage is shown when coloring by CPU
					processTree.DisplayOptions.ShowCpuPercent = true

					// Apply color based on CPU usage thresholds in percentage
					if process.CPUPercent < 5 {
						// Low CPU usage (< 5%)
						processTree.Colorizer.CPULow(processTree.ColorScheme, value)
					} else if process.CPUPercent >= 5 && process.CPUPercent < 15 {
						// Medium CPU usage (5-15%)
						processTree.Colorizer.CPUMedium(processTree.ColorScheme, value)
					} else if process.CPUPercent >= 15 {
						// High CPU usage (> 15%)
						processTree.Colorizer.CPUHigh(processTree.ColorScheme, value)
					}
				case "mem":
					// Ensure memory usage is shown when coloring by memory
					processTree.DisplayOptions.ShowMemoryUsage = true

					// Calculate memory usage as percentage of total system memory
					percent := (process.MemoryInfo.RSS / processTree.DisplayOptions.InstalledMemory) * 100

					// Apply color based on memory usage thresholds in percentage
					if percent < 10 {
						// Low memory usage (< 10%)
						processTree.Colorizer.MemoryLow(processTree.ColorScheme, value)
					} else if percent >= 10 && percent < 20 {
						// Medium memory usage (10-20%)
						processTree.Colorizer.MemoryMedium(processTree.ColorScheme, value)
					} else if percent >= 20 {
						// High memory usage (> 20%)
						processTree.Colorizer.MemoryHigh(processTree.ColorScheme, value)
					}
				}
				// } else {
				// 	processTree.Colorizer.Default(processTree.ColorScheme, value)
			}
		}
	}
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
//
// truncateANSI truncates a string containing ANSI escape sequences to fit within a specified screen width.
// It preserves ANSI color and formatting codes while only counting visible characters toward the width limit.
//
// Parameters:
//   - input: The string to truncate, which may contain ANSI escape sequences
//
// The function handles multi-byte Unicode characters correctly by using utf8.DecodeRuneInString
// and accounts for characters with different display widths using the runewidth package.
// If truncation occurs, "..." is appended to the result.
//
// Returns:
//   - A string that fits within screenWidth, with ANSI sequences preserved.
func (processTree *ProcessTree) truncateANSI(input string) string {
	dots := "..."

	if processTree.DisplayOptions.ScreenWidth <= 3 {
		return dots
	}

	// First, check actual display width
	if processTree.visibleWidth(input) <= processTree.DisplayOptions.ScreenWidth {
		return input // No truncation needed
	}

	targetWidth := processTree.DisplayOptions.ScreenWidth - len(dots)
	var output strings.Builder
	width := 0

	for len(input) > 0 {
		if loc := AnsiEscape.FindStringIndex(input); loc != nil && loc[0] == 0 {
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
	return output.String() + "\x1b[0m" // Prevent ANSI bleed
}

func (processTree *ProcessTree) stripANSI(input string) string {
	var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	return ansiRegex.ReplaceAllString(input, "")
}

func (processTree *ProcessTree) truncatePlain(input string) string {
	visibleWidth := processTree.visibleWidth(input)

	if visibleWidth <= processTree.DisplayOptions.ScreenWidth {
		return input
	}

	// If the string is longer than the screen width, truncate it
	var (
		builder   strings.Builder
		currWidth int
		truncated bool
		byteIndex int
		charWidth int
		r         rune
		size      int
	)

	builder.Grow(processTree.DisplayOptions.ScreenWidth + 3) // +3 for "..."

	for byteIndex = 0; byteIndex < len(input); {
		r, size = utf8.DecodeRuneInString(input[byteIndex:])
		charWidth = runewidth.RuneWidth(r)

		if currWidth+charWidth > processTree.DisplayOptions.ScreenWidth-3 { // -3 for "..."
			truncated = true
			break
		}

		builder.WriteRune(r)
		currWidth += charWidth
		byteIndex += size
	}

	if truncated {
		builder.WriteString("...")
	}

	return builder.String()
}

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
//
// visibleWidth calculates the display width of a string containing ANSI escape sequences.
// It ignores ANSI escape sequences and counts only the visible characters' width.
// The function properly handles multi-byte Unicode characters and characters with
// different display widths (like CJK characters that take up 2 columns).
//
// Parameters:
//   - input: The string to calculate the width for, which may contain ANSI escape sequences
//
// Returns:
//   - The display width of the string, excluding ANSI escape sequences
func (processTree *ProcessTree) visibleWidth(input string) int {
	width := 0
	for len(input) > 0 {
		if loc := AnsiEscape.FindStringIndex(input); loc != nil && loc[0] == 0 {
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
