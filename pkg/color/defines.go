package color

const (
	// Reset
	AnsiReset = "\033[0m"

	// Regular Colors
	AnsiBlack   = "\033[30m"
	AnsiRed     = "\033[31m"
	AnsiGreen   = "\033[32m"
	AnsiYellow  = "\033[33m"
	AnsiBlue    = "\033[34m"
	AnsiMagenta = "\033[35m"
	AnsiCyan    = "\033[36m"
	AnsiWhite   = "\033[37m"

	// Bold Colors (technically "bright", but often shown as bold in terminals)
	AnsiBlackBold   = "\033[1;30m"
	AnsiRedBold     = "\033[1;31m"
	AnsiGreenBold   = "\033[1;32m"
	AnsiYellowBold  = "\033[1;33m"
	AnsiBlueBold    = "\033[1;34m"
	AnsiMagentaBold = "\033[1;35m"
	AnsiCyanBold    = "\033[1;36m"
	AnsiWhiteBold   = "\033[1;37m"
)

type ColorFunc func(cs ColorScheme, text *string)

var Colorizers = map[string]Colorizer{
	"8color": {
		Age:                Color8GreenBold,
		Args:               Color8Red,
		Command:            Color8BlueBold,
		CompactedThread:    Color8BlackBold,
		CompactStr:         Color8BlackBold,
		Connector:          Color8BlackBold,
		CPU:                Color8YellowBold,
		CPUHigh:            Color8Red,
		CPULow:             Color8Green,
		CPUMedium:          Color8Yellow,
		Default:            Color8Green,
		Memory:             Color8RedBold,
		MemoryHigh:         Color8Red,
		MemoryLow:          Color8Green,
		MemoryMedium:       Color8Yellow,
		NumThreads:         Color8WhiteBold,
		Owner:              Color8CyanBold,
		OwnerTransition:    Color8BlackBold,
		PIDPGID:            Color8MagentaBold,
		Prefix:             Color8Green,
		ProcessAgeHigh:     Color8Cyan,
		ProcessAgeLow:      Color8Red,
		ProcessAgeMedium:   Color8Yellow,
		ProcessAgeVeryHigh: Color8Green,
	},
	"256color": {
		Age:                Color256Green,
		Args:               Color256Red,
		Command:            Color256Blue,
		CompactedThread:    Color256BlackBold,
		CompactStr:         Color256BlackBold,
		Connector:          Color256BlackBold,
		CPU:                Color256Yellow,
		CPUHigh:            Color256Red,
		CPULow:             Color256Green,
		CPUMedium:          Color256Yellow,
		Default:            Color256Green,
		Memory:             Color256Orange,
		MemoryHigh:         Color256Red,
		MemoryLow:          Color256Green,
		MemoryMedium:       Color256Yellow,
		NumThreads:         Color256White,
		Owner:              Color256Cyan,
		OwnerTransition:    Color256BlackBold,
		PIDPGID:            Color256Magenta,
		Prefix:             Color256Green,
		ProcessAgeHigh:     Color256Cyan,
		ProcessAgeLow:      Color256Red,
		ProcessAgeMedium:   Color256Yellow,
		ProcessAgeVeryHigh: Color256Green,
	},
}

type Colorizer struct {
	Age                ColorFunc
	Args               ColorFunc
	Command            ColorFunc
	CompactedThread    ColorFunc
	CompactStr         ColorFunc
	Connector          ColorFunc
	CPU                ColorFunc
	CPUHigh            ColorFunc
	CPULow             ColorFunc
	CPUMedium          ColorFunc
	Default            ColorFunc
	Memory             ColorFunc
	MemoryHigh         ColorFunc
	MemoryLow          ColorFunc
	MemoryMedium       ColorFunc
	NumThreads         ColorFunc
	Owner              ColorFunc
	OwnerTransition    ColorFunc
	PIDPGID            ColorFunc
	Prefix             ColorFunc
	ProcessAgeHigh     ColorFunc
	ProcessAgeLow      ColorFunc
	ProcessAgeMedium   ColorFunc
	ProcessAgeVeryHigh ColorFunc
}

type ColorMap struct {
	R    int
	G    int
	B    int
	Ansi string
}

type ColorScheme struct {
	Black       ColorMap
	BlackBold   ColorMap
	Blue        ColorMap
	BlueBold    ColorMap
	Cyan        ColorMap
	CyanBold    ColorMap
	Green       ColorMap
	GreenBold   ColorMap
	Orange      ColorMap
	OrangeBold  ColorMap
	Magenta     ColorMap
	MagentaBold ColorMap
	Red         ColorMap
	RedBold     ColorMap
	White       ColorMap
	WhiteBold   ColorMap
	Yellow      ColorMap
	YellowBold  ColorMap
}

// https://en.wikipedia.org/wiki/ANSI_escape_code#Colors
// https://www.ditig.com/256-colors-cheat-sheet
var ColorSchemes map[string]ColorScheme = map[string]ColorScheme{
	"windows10": {
		Black:       ColorMap{R: 12, G: 12, B: 12},
		BlackBold:   ColorMap{R: 118, G: 118, B: 118},
		Blue:        ColorMap{R: 0, G: 255, B: 218},
		BlueBold:    ColorMap{R: 59, G: 120, B: 255},
		Cyan:        ColorMap{R: 58, G: 150, B: 221},
		CyanBold:    ColorMap{R: 97, G: 214, B: 214},
		Green:       ColorMap{R: 19, G: 161, B: 14},
		GreenBold:   ColorMap{R: 22, G: 198, B: 12},
		Magenta:     ColorMap{R: 136, G: 23, B: 152},
		MagentaBold: ColorMap{R: 180, G: 0, B: 158},
		Red:         ColorMap{R: 197, G: 15, B: 31},
		RedBold:     ColorMap{R: 231, G: 72, B: 86},
		White:       ColorMap{R: 204, G: 204, B: 204},
		WhiteBold:   ColorMap{R: 242, G: 242, B: 242},
		Yellow:      ColorMap{R: 193, G: 156, B: 0},
		YellowBold:  ColorMap{R: 249, G: 241, B: 165},
		// Not part of the standard 16 colors
		Orange:     ColorMap{R: 215, G: 95, B: 0},
		OrangeBold: ColorMap{R: 255, G: 135, B: 0},
	},
	"powershell": {
		Black:       ColorMap{R: 0, G: 0, B: 0},
		BlackBold:   ColorMap{R: 128, G: 128, B: 128},
		Blue:        ColorMap{R: 0, G: 0, B: 128},
		BlueBold:    ColorMap{R: 0, G: 0, B: 255},
		Cyan:        ColorMap{R: 0, G: 128, B: 128},
		CyanBold:    ColorMap{R: 0, G: 255, B: 255},
		Green:       ColorMap{R: 0, G: 128, B: 0},
		GreenBold:   ColorMap{R: 0, G: 255, B: 0},
		Magenta:     ColorMap{R: 1, G: 36, B: 86},
		MagentaBold: ColorMap{R: 255, G: 0, B: 255},
		Red:         ColorMap{R: 128, G: 0, B: 0},
		RedBold:     ColorMap{R: 255, G: 0, B: 0},
		White:       ColorMap{R: 192, G: 192, B: 192},
		WhiteBold:   ColorMap{R: 255, G: 255, B: 255},
		Yellow:      ColorMap{R: 237, G: 237, B: 240},
		YellowBold:  ColorMap{R: 255, G: 255, B: 0},
		// Not part of the standard 16 colors
		Orange:     ColorMap{R: 215, G: 95, B: 0},
		OrangeBold: ColorMap{R: 255, G: 135, B: 0},
	},
	"darwin": {
		Black:       ColorMap{R: 0, G: 0, B: 0},
		BlackBold:   ColorMap{R: 102, G: 102, B: 102},
		Blue:        ColorMap{R: 0, G: 0, B: 178},
		BlueBold:    ColorMap{R: 0, G: 0, B: 255},
		Cyan:        ColorMap{R: 0, G: 166, B: 178},
		CyanBold:    ColorMap{R: 0, G: 230, B: 230},
		Green:       ColorMap{R: 0, G: 166, B: 0},
		GreenBold:   ColorMap{R: 0, G: 217, B: 0},
		Magenta:     ColorMap{R: 178, G: 0, B: 178},
		MagentaBold: ColorMap{R: 230, G: 0, B: 230},
		Red:         ColorMap{R: 153, G: 0, B: 0},
		RedBold:     ColorMap{R: 230, G: 0, B: 0},
		White:       ColorMap{R: 191, G: 191, B: 191},
		WhiteBold:   ColorMap{R: 230, G: 230, B: 230},
		Yellow:      ColorMap{R: 153, G: 153, B: 0},
		YellowBold:  ColorMap{R: 230, G: 230, B: 0},
		// Not part of the standard 16 colors
		Orange:     ColorMap{R: 215, G: 95, B: 0},
		OrangeBold: ColorMap{R: 255, G: 135, B: 0},
	},
	"linux": {
		Black:       ColorMap{R: 1, G: 1, B: 1},
		BlackBold:   ColorMap{R: 128, G: 128, B: 128},
		Blue:        ColorMap{R: 0, G: 111, B: 184},
		BlueBold:    ColorMap{R: 0, G: 0, B: 255},
		Cyan:        ColorMap{R: 41, G: 181, B: 233},
		CyanBold:    ColorMap{R: 0, G: 255, B: 255},
		Green:       ColorMap{R: 57, G: 181, B: 74},
		GreenBold:   ColorMap{R: 0, G: 255, B: 0},
		Magenta:     ColorMap{R: 118, G: 38, B: 113},
		MagentaBold: ColorMap{R: 255, G: 0, B: 255},
		Red:         ColorMap{R: 222, G: 56, B: 43},
		RedBold:     ColorMap{R: 255, G: 0, B: 0},
		White:       ColorMap{R: 204, G: 204, B: 204},
		WhiteBold:   ColorMap{R: 255, G: 255, B: 255},
		Yellow:      ColorMap{R: 255, G: 199, B: 6},
		YellowBold:  ColorMap{R: 255, G: 255, B: 0},
		// Not part of the standard 16 colors
		Orange:     ColorMap{R: 215, G: 95, B: 0},
		OrangeBold: ColorMap{R: 255, G: 135, B: 0},
	},
	"xterm": {
		Black:       ColorMap{R: 0, G: 0, B: 0},
		BlackBold:   ColorMap{R: 127, G: 127, B: 127},
		Blue:        ColorMap{R: 0, G: 0, B: 238},
		BlueBold:    ColorMap{R: 92, G: 92, B: 255},
		Cyan:        ColorMap{R: 0, G: 205, B: 205},
		CyanBold:    ColorMap{R: 0, G: 255, B: 255},
		Green:       ColorMap{R: 0, G: 205, B: 0},
		GreenBold:   ColorMap{R: 0, G: 255, B: 0},
		Magenta:     ColorMap{R: 205, G: 0, B: 205},
		MagentaBold: ColorMap{R: 255, G: 0, B: 255},
		Red:         ColorMap{R: 205, G: 0, B: 0},
		RedBold:     ColorMap{R: 255, G: 0, B: 0},
		White:       ColorMap{R: 229, G: 229, B: 229},
		WhiteBold:   ColorMap{R: 255, G: 255, B: 255},
		Yellow:      ColorMap{R: 205, G: 205, B: 0},
		YellowBold:  ColorMap{R: 255, G: 255, B: 0},
		// Not part of the standard 16 colors
		Orange:     ColorMap{R: 215, G: 95, B: 0},
		OrangeBold: ColorMap{R: 255, G: 135, B: 0},
	},
	"ansi8": {
		Black:       ColorMap{Ansi: AnsiBlack},
		BlackBold:   ColorMap{Ansi: AnsiBlackBold},
		Blue:        ColorMap{Ansi: AnsiBlue},
		BlueBold:    ColorMap{Ansi: AnsiBlueBold},
		Cyan:        ColorMap{Ansi: AnsiCyan},
		CyanBold:    ColorMap{Ansi: AnsiCyanBold},
		Green:       ColorMap{Ansi: AnsiGreen},
		GreenBold:   ColorMap{Ansi: AnsiGreenBold},
		Magenta:     ColorMap{Ansi: AnsiMagenta},
		MagentaBold: ColorMap{Ansi: AnsiMagentaBold},
		Red:         ColorMap{Ansi: AnsiRed},
		RedBold:     ColorMap{Ansi: AnsiRedBold},
		White:       ColorMap{Ansi: AnsiWhite},
		WhiteBold:   ColorMap{Ansi: AnsiWhiteBold},
		Yellow:      ColorMap{Ansi: AnsiYellow},
		YellowBold:  ColorMap{Ansi: AnsiYellowBold},
	},
}
