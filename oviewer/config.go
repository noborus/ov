package oviewer

// Config represents the settings of ov.
type Config struct {
	// KeyBinding
	Keybind map[string][]string
	// Mode represents the operation of the customized mode.
	Mode map[string]general
	// ViewMode represents the view mode.
	// ViewMode sets several settings together and can be easily switched.
	ViewMode string
	// Default keybindings. Disabled if the default keybinding is "disable".
	DefaultKeyBind string
	// StyleColumnRainbow  is the style that applies to the column rainbow color highlight.
	StyleColumnRainbow []OVStyle
	// StyleMultiColorHighlight is the style that applies to the multi color highlight.
	StyleMultiColorHighlight []OVStyle

	// Prompt is the prompt setting.
	Prompt OVPromptConfig

	// StyleHeader is the style that applies to the header.
	StyleHeader OVStyle
	// StyleBody is the style that applies to the body.
	StyleBody OVStyle
	// StyleLineNumber is a style that applies line number.
	StyleLineNumber OVStyle
	// StyleSearchHighlight is the style that applies to the search highlight.
	StyleSearchHighlight OVStyle
	// StyleColumnHighlight is the style that applies to the column highlight.
	StyleColumnHighlight OVStyle
	// StyleMarkLine is a style that marked line.
	StyleMarkLine OVStyle
	// StyleSectionLine is a style that section delimiter line.
	StyleSectionLine OVStyle
	// StyleVerticalHeader is a style that applies to the vertical header.
	StyleVerticalHeader OVStyle
	// StyleJumpTargetLine is the line that displays the search results.
	StyleJumpTargetLine OVStyle
	// StyleAlternate is a style that applies line by line.
	StyleAlternate OVStyle
	// StyleOverStrike is a style that applies to overstrike.
	StyleOverStrike OVStyle
	// StyleOverLine is a style that applies to overstrike underlines.
	StyleOverLine OVStyle
	// General represents the general behavior.
	General general
	// BeforeWriteOriginal specifies the number of lines before the current position.
	// 0 is the top of the current screen
	BeforeWriteOriginal int
	// AfterWriteOriginal specifies the number of lines after the current position.
	// 0 specifies the bottom of the screen.
	AfterWriteOriginal int
	// MemoryLimit is a number that limits chunk loading.
	MemoryLimit int
	// MemoryLimitFile is a number that limits the chunks loading a file into memory.
	MemoryLimitFile int
	// DisableMouse indicates whether mouse support is disabled.
	DisableMouse bool
	// IsWriteOriginal indicates whether the current screen should be written on quit.
	IsWriteOriginal bool
	// QuitSmall indicates whether to quit if the output fits on one screen.
	QuitSmall bool
	// QuitSmallFilter indicates whether to quit if the output fits on one screen and a filter is applied.
	QuitSmallFilter bool
	// CaseSensitive is case-sensitive if true.
	CaseSensitive bool
	// SmartCaseSensitive indicates whether lowercase search should ignore case.
	SmartCaseSensitive bool
	// RegexpSearch indicates whether to use regular expression search.
	RegexpSearch bool
	// Incsearch indicates whether to use incremental search.
	Incsearch bool
	// NotifyEOF specifies the number of times to notify EOF.
	NotifyEOF int

	// ShrinkChar specifies the character to display when the column is shrunk.
	ShrinkChar string
	// DisableColumnCycle indicates whether to disable column cycling.
	DisableColumnCycle bool
	// Debug indicates whether to enable debug output.
	Debug bool
}

// OVPromptConfigNormal is the normal prompt setting.
type OVPromptConfigNormal struct {
	// ShowFilename controls whether to display filename.
	ShowFilename bool
	// InvertColor controls whether the text is colored and inverted.
	InvertColor bool
	// ProcessOfCount controls whether to display the progress of the count.
	ProcessOfCount bool
}

// OVPromptConfig is the prompt setting.
type OVPromptConfig struct {
	// Normal is the normal prompt setting.
	Normal OVPromptConfigNormal
}

// NewConfig return the structure of Config with default values.
func NewConfig() Config {
	return Config{
		MemoryLimit:     -1,
		MemoryLimitFile: 100,
		StyleHeader: OVStyle{
			Bold: true,
		},
		StyleAlternate: OVStyle{
			Background: "gray",
		},
		StyleOverStrike: OVStyle{
			Bold: true,
		},
		StyleOverLine: OVStyle{
			Underline: true,
		},
		StyleLineNumber: OVStyle{
			Bold: true,
		},
		StyleSearchHighlight: OVStyle{
			Reverse: true,
		},
		StyleColumnHighlight: OVStyle{
			Reverse: true,
		},
		StyleMarkLine: OVStyle{
			Background: "darkgoldenrod",
		},
		StyleSectionLine: OVStyle{
			Background: "slateblue",
		},
		StyleVerticalHeader: OVStyle{
			Background: "darkgray",
		},
		StyleMultiColorHighlight: []OVStyle{
			{Foreground: "red"},
			{Foreground: "aqua"},
			{Foreground: "yellow"},
			{Foreground: "fuchsia"},
			{Foreground: "lime"},
			{Foreground: "blue"},
			{Foreground: "grey"},
		},
		StyleColumnRainbow: []OVStyle{
			{Foreground: "white"},
			{Foreground: "crimson"},
			{Foreground: "aqua"},
			{Foreground: "lightsalmon"},
			{Foreground: "lime"},
			{Foreground: "blue"},
			{Foreground: "yellowgreen"},
		},
		StyleJumpTargetLine: OVStyle{
			Underline: true,
		},
		General: general{
			TabWidth:       8,
			MarkStyleWidth: 1,
		},
		Prompt: OVPromptConfig{
			Normal: OVPromptConfigNormal{
				ShowFilename:   true,
				InvertColor:    true,
				ProcessOfCount: true,
			},
		},
	}
}
