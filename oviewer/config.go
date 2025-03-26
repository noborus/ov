package oviewer

// Config represents the settings of ov.
type Config struct {
	// KeyBinding
	Keybind map[string][]string
	// Mode represents the operation of the customized mode.
	Mode map[string]General
	// ViewMode represents the view mode.
	// ViewMode sets several settings together and can be easily switched.
	ViewMode string
	// Default keybindings. Disabled if the default keybinding is "disable".
	DefaultKeyBind string

	// Prompt is the prompt setting.
	Prompt OVPromptConfig
	// MinStartX is the minimum value of the start position.
	MinStartX int

	// StyleColumnRainbow  is the style that applies to the column rainbow color highlight.
	StyleColumnRainbow []OVStyle
	// StyleMultiColorHighlight is the style that applies to the multi color highlight.
	StyleMultiColorHighlight []OVStyle
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
	// StyleRuler is a style that applies to the ruler.
	StyleRuler OVStyle
	// StyleHeaderBorder is the style that applies to the boundary line of the header.
	// The boundary line of the header refers to the visual separator between the header and the rest of the content.
	StyleHeaderBorder OVStyle
	// StyleSectionHeaderBorder is the style that applies to the boundary line of the section header.
	// The boundary line of the section header is the line that separates different sections in the header.
	StyleSectionHeaderBorder OVStyle
	// StyleVerticalHeaderBorder is the style that applies to the boundary character of the vertical header.
	// The boundary character of the vertical header refers to the visual separator that delineates the vertical header from the rest of the content.
	StyleVerticalHeaderBorder OVStyle

	// GeneralConfig is the general setting.
	General General
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

	// IsWriteOnExit indicates whether to write the current screen on exit.
	IsWriteOnExit bool
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

// General is the general configuration.
type General struct {
	// Converter is the converter name.
	Converter *string
	// Align is the alignment.
	Align *bool
	// Raw is the raw setting.
	Raw *bool
	// Caption is an additional caption to display after the file name.
	Caption *string
	// ColumnDelimiter is a column delimiter.
	ColumnDelimiter *string
	// SectionDelimiter is a section delimiter.
	SectionDelimiter *string
	// Specified string for jumpTarget.
	JumpTarget *string
	// MultiColorWords specifies words to color separated by spaces.
	MultiColorWords *[]string

	// TabWidth is tab stop num.
	TabWidth *int
	// Header is number of header lines to be fixed.
	Header *int
	// VerticalHeader is the number of vertical header lines.
	VerticalHeader *int
	// HeaderColumn is the number of columns from the left to be fixed.
	// If 0 is specified, no columns are fixed.
	HeaderColumn *int
	// SkipLines is the rows to skip.
	SkipLines *int
	// WatchInterval is the watch interval (seconds).
	WatchInterval *int
	// MarkStyleWidth is width to apply the style of the marked line.
	MarkStyleWidth *int
	// SectionStartPosition is a section start position.
	SectionStartPosition *int
	// SectionHeaderNum is the number of lines in the section header.
	SectionHeaderNum *int
	// HScrollWidth is the horizontal scroll width.
	HScrollWidth *string
	// HScrollWidthNum is the horizontal scroll width.
	HScrollWidthNum *int
	// RulerType is the ruler type (0: none, 1: relative, 2: absolute).
	RulerType *RulerType
	// AlternateRows alternately style rows.
	AlternateRows *bool
	// ColumnMode is column mode.
	ColumnMode *bool
	// ColumnWidth is column width mode.
	ColumnWidth *bool
	// ColumnRainbow is column rainbow.
	ColumnRainbow *bool
	// LineNumMode displays line numbers.
	LineNumMode *bool
	// Wrap is Wrap mode.
	WrapMode *bool
	// FollowMode is the follow mode.
	FollowMode *bool
	// FollowAll is a follow mode for all documents.
	FollowAll *bool
	// FollowSection is a follow mode that uses section instead of line.
	FollowSection *bool
	// FollowName is the mode to follow files by name.
	FollowName *bool
	// PlainMode is whether to enable the original character decoration.
	PlainMode *bool
	// SectionHeader is whether to display the section header.
	SectionHeader *bool
	// HideOtherSection is whether to hide other sections.
	HideOtherSection *bool
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
		StyleVerticalHeader: OVStyle{},
		StyleVerticalHeaderBorder: OVStyle{
			Background: "#c0c0c0",
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
		StyleRuler: OVStyle{
			Background: "#333333",
			Foreground: "#CCCCCC",
			Bold:       true,
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
