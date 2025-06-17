package oviewer

import (
	"time"
)

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
	Prompt PromptConfig
	// MinStartX is the minimum value of the start position.
	MinStartX int

	// StyleOverStrike is a style that applies to overstrike.
	StyleOverStrike OVStyle
	// StyleOverLine is a style that applies to overstrike underlines.
	StyleOverLine OVStyle

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

	// ReadWaitTime is the time to wait for reading before starting a search.
	ReadWaitTime time.Duration

	// ClipboardMethod specifies the method to use for copying to the clipboard.
	// Supported values:
	// - "OSC52": Uses the OSC52 escape sequence for clipboard operations. This requires terminal support.
	// - "default": Uses the default clipboard method provided by the system or application.
	// In fact, all other settings are default except for OSC52. In the future, “auto” will be added.
	ClipboardMethod string

	// Editor is the editor command to use for editing files.
	Editor string
	// ShrinkChar specifies the character to display when the column is shrunk.
	ShrinkChar string
	// DisableColumnCycle indicates whether to disable column cycling.
	DisableColumnCycle bool
	// Debug indicates whether to enable debug output.
	Debug bool
	// deprecatedStyleConfig is the old style setting.
	//
	// Deprecated: This setting is planned to be removed in future versions.
	deprecatedStyleConfig `yaml:",inline" mapstructure:",squash"`
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
	// StatusLine indicates whether to hide the status line.
	StatusLine *bool
	// Prompt is the prompt configuration.
	Prompt PromptConfig
	// Style is the style setting.
	Style StyleConfig
}

// PromptConfigNormal is the normal prompt setting.
type PromptConfigNormal struct {
	// ShowFilename controls whether to display filename.
	ShowFilename *bool
	// InvertColor controls whether the text is colored and inverted.
	InvertColor *bool
	// ProcessOfCount controls whether to display the progress of the count.
	ProcessOfCount *bool
	// CursorType controls the type of cursor to display.
	CursorType *int
}

// PromptConfigInput is the input prompt setting.
type PromptConfigInput struct {
	CursorType *int
}

// PromptConfig is the prompt setting.
type PromptConfig struct {
	// Normal is the normal prompt setting.
	Normal PromptConfigNormal
	Input  PromptConfigInput
}

// OVPromptConfigNormal is the normal prompt setting.
type OVPromptConfigNormal struct {
	// ShowFilename controls whether to display filename.
	ShowFilename bool
	// InvertColor controls whether the text is colored and inverted.
	InvertColor bool
	// ProcessOfCount controls whether to display the progress of the count.
	ProcessOfCount bool
	// CursorType controls the type of cursor to display.
	CursorType int
}

// OVPromptConfigInput is the input prompt setting.
type OVPromptConfigInput struct {
	CursorType int
}

// OVPromptConfig is the prompt setting.
type OVPromptConfig struct {
	// Normal is the normal prompt setting.
	Normal OVPromptConfigNormal
	Input  OVPromptConfigInput
}

// NewOVPromptConfig returns the structure of OVPromptConfig with default values.
func NewOVPromptConfig() OVPromptConfig {
	return OVPromptConfig{
		Normal: OVPromptConfigNormal{
			ShowFilename:   true,
			InvertColor:    true,
			ProcessOfCount: true,
		},
		Input: OVPromptConfigInput{},
	}
}

// NewConfig return the structure of Config with default values.
func NewConfig() Config {
	return Config{
		MemoryLimit:     -1,
		MemoryLimitFile: 100,
		ReadWaitTime:    1000 * time.Millisecond,
	}
}

// StyleConfig is the style setting.
type StyleConfig struct {
	// ColumnRainbow is the style that applies to the column rainbow color highlight.
	ColumnRainbow *[]OVStyle
	// MultiColorHighlight is the style that applies to the multi color highlight.
	MultiColorHighlight *[]OVStyle
	// Header is the style that applies to the header.
	Header *OVStyle
	// Body is the style that applies to the body.
	Body *OVStyle
	// LineNumber is a style that applies line number.
	LineNumber *OVStyle
	// SearchHighlight is the style that applies to the search highlight.
	SearchHighlight *OVStyle
	// ColumnHighlight is the style that applies to the column highlight.
	ColumnHighlight *OVStyle
	// MarkLine is a style that marked line.
	MarkLine *OVStyle
	// SectionLine is a style that section delimiter line.
	SectionLine *OVStyle
	// VerticalHeader is a style that applies to the vertical header.
	VerticalHeader *OVStyle
	// JumpTargetLine is the line that displays the search results.
	JumpTargetLine *OVStyle
	// Alternate is a style that applies line by line.
	Alternate *OVStyle
	// Ruler is a style that applies to the ruler.
	Ruler *OVStyle
	// HeaderBorder is the style that applies to the boundary line of the header.
	// The boundary line of the header refers to the visual separator between the header and the rest of the content.
	HeaderBorder *OVStyle
	// SectionHeaderBorder is the style that applies to the boundary line of the section header.
	// The boundary line of the section header is the line that separates different sections in the header.
	SectionHeaderBorder *OVStyle
	// VerticalHeaderBorder is the style that applies to the boundary character of the vertical header.
	// The boundary character of the vertical header refers to the visual separator that delineates the vertical header from the rest of the content.
	VerticalHeaderBorder *OVStyle
	// LeftStatus is the style that applies to the left side of the status line.
	LeftStatus *OVStyle
	// RightStatus is the style that applies to the right side of the status line.
	RightStatus *OVStyle
}

// deprecatedStyleConfig is the old style setting.
//
// Deprecated: This setting is planned to be removed in future versions.
type deprecatedStyleConfig struct {
	// Deprecated: This setting is planned to be removed in future versions.
	// Use General.Style.ColumnRainbow instead.
	StyleColumnRainbow []OVStyle
	// Deprecated: This setting is planned to be removed in future versions.
	// Use General.Style.MultiColorHighlight instead.
	StyleMultiColorHighlight []OVStyle
	// Deprecated: This setting is planned to be removed in future versions.
	// Use General.Style.Header instead.
	StyleHeader OVStyle
	// Deprecated: This setting is planned to be removed in future versions.
	// Use General.Style.Body instead.
	StyleBody OVStyle
	// Deprecated: This setting is planned to be removed in future versions.
	// Use General.Style.LineNumber instead.
	StyleLineNumber OVStyle
	// Deprecated: This setting is planned to be removed in future versions.
	// Use General.Style.SearchHighlight instead.
	StyleSearchHighlight OVStyle
	// Deprecated: This setting is planned to be removed in future versions.
	// Use General.Style.ColumnHighlight instead.
	StyleColumnHighlight OVStyle
	// Deprecated: This setting is planned to be removed in future versions.
	// Use General.Style.MarkLine instead.
	StyleMarkLine OVStyle
	// Deprecated: This setting is planned to be removed in future versions.
	// Use General.Style.SectionLine instead.
	StyleSectionLine OVStyle
	// Deprecated: This setting is planned to be removed in future versions.
	// Use General.Style.VerticalHeader instead.
	StyleVerticalHeader OVStyle
	// Deprecated: This setting is planned to be removed in future versions.
	// Use General.Style.JumpTargetLine instead.
	StyleJumpTargetLine OVStyle
	// Deprecated: This setting is planned to be removed in future versions.
	// Use General.Style.Alternate instead.
	StyleAlternate OVStyle
	// Deprecated: This setting is planned to be removed in future versions.
	// Use General.Style.Ruler instead.
	StyleRuler OVStyle
	// Deprecated: This setting is planned to be removed in future versions.
	// Use General.Style.HeaderBorder instead.
	StyleHeaderBorder OVStyle
	// Deprecated: This setting is planned to be removed in future versions.
	// Use General.Style.SectionHeaderBorder instead.
	StyleSectionHeaderBorder OVStyle
	// Deprecated: This setting is planned to be removed in future versions.
	// Use General.Style.VerticalHeaderBorder instead.
	StyleVerticalHeaderBorder OVStyle
}
