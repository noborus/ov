package oviewer

import "regexp"

// RunTimeSettings structure contains the RunTimeSettings of the display.
// RunTimeSettings contains values that determine the behavior of each document.
type RunTimeSettings struct {
	// Name is the name of the view mode.
	Name string
	// Converter is the converter name.
	Converter string
	// Caption is an additional caption to display after the file name.
	Caption string
	// ColumnDelimiterReg is a compiled regular expression of ColumnDelimiter.
	ColumnDelimiterReg *regexp.Regexp
	// ColumnDelimiter is a column delimiter.
	ColumnDelimiter string
	// SectionDelimiterReg is a section delimiter.
	SectionDelimiterReg *regexp.Regexp
	// SectionDelimiter is a section delimiter.
	SectionDelimiter string
	// Specified string for jumpTarget.
	JumpTarget string
	// MultiColorWords specifies words to color separated by spaces.
	MultiColorWords []string

	// TabWidth is tab stop num.
	TabWidth int
	// Header is number of header lines to be fixed.
	Header int
	// VerticalHeader is the number of vertical header lines.
	VerticalHeader int
	// HeaderColumn is the number of columns from the left to be fixed.
	// If 0 is specified, no columns are fixed.
	HeaderColumn int
	// SkipLines is the rows to skip.
	SkipLines int
	// WatchInterval is the watch interval (seconds).
	WatchInterval int
	// MarkStyleWidth is width to apply the style of the marked line.
	MarkStyleWidth int
	// SectionStartPosition is a section start position.
	SectionStartPosition int
	// SectionHeaderNum is the number of lines in the section header.
	SectionHeaderNum int
	// HScrollWidth is the horizontal scroll width.
	HScrollWidth string
	// HScrollWidthNum is the horizontal scroll width.
	HScrollWidthNum int
	// VScrollLines is the number of lines to scroll with the mouse wheel.
	VScrollLines int
	// RulerType is the ruler type (0: none, 1: relative, 2: absolute).
	RulerType RulerType
	// AlternateRows alternately style rows.
	AlternateRows bool
	// ColumnMode is column mode.
	ColumnMode bool
	// ColumnWidth is column width mode.
	ColumnWidth bool
	// ColumnRainbow is column rainbow.
	ColumnRainbow bool
	// LineNumMode displays line numbers.
	LineNumMode bool
	// WrapMode is wrap mode.
	WrapMode bool
	// FollowMode is the follow mode.
	FollowMode bool
	// FollowAll is a follow mode for all documents.
	FollowAll bool
	// FollowSection is a follow mode that uses section instead of line.
	FollowSection bool
	// FollowName is the mode to follow files by name.
	FollowName bool
	// PlainMode is whether to enable the original character decoration.
	PlainMode bool
	// SectionHeader is whether to display the section header.
	SectionHeader bool
	// HideOtherSection is whether to hide other sections.
	HideOtherSection bool
	// StatusLine is whether to display the status line.
	StatusLine bool

	// PromptConfig is the prompt configuration.
	OVPromptConfig
	// Style is the style of the document.
	Style Style
}

// Style structure contains the style settings of the display.
type Style struct {
	// ColumnRainbow is the style that applies to the column rainbow color highlight.
	ColumnRainbow []OVStyle
	// MultiColorHighlight is the style that applies to the multi color highlight.
	MultiColorHighlight []OVStyle
	// Header is the style that applies to the header.
	Header OVStyle
	// Body is the style that applies to the body.
	Body OVStyle
	// LineNumber is a style that applies line number.
	LineNumber OVStyle
	// SearchHighlight is the style that applies to the search highlight.
	SearchHighlight OVStyle
	// ColumnHighlight is the style that applies to the column highlight.
	ColumnHighlight OVStyle
	// MarkLine is a style that marked line.
	MarkLine OVStyle
	// SectionLine is a style that section delimiter line.
	SectionLine OVStyle
	// VerticalHeader is a style that applies to the vertical header.
	VerticalHeader OVStyle
	// JumpTargetLine is the line that displays the search results.
	JumpTargetLine OVStyle
	// Alternate is a style that applies line by line.
	Alternate OVStyle
	// Ruler is a style that applies to the ruler.
	Ruler OVStyle
	// HeaderBorder is the style that applies to the boundary line of the header.
	// The boundary line of the header refers to the visual separator between the header and the rest of the content.
	HeaderBorder OVStyle
	// SectionHeaderBorder is the style that applies to the boundary line of the section header.
	// The boundary line of the section header is the line that separates different sections in the header.
	SectionHeaderBorder OVStyle
	// VerticalHeaderBorder is the style that applies to the boundary character of the vertical header.
	// The boundary character of the vertical header refers to the visual separator that delineates the vertical header from the rest of the content.
	VerticalHeaderBorder OVStyle
	// LeftStatus is the style that applies to the left status line.
	LeftStatus OVStyle
	// RightStatus is the style that applies to the right status line.
	RightStatus OVStyle
	// SelectActive is the style that applies to the text being selected (during mouse drag).
	SelectActive OVStyle
	// SelectCopied is the style that applies to the text that has been copied to clipboard.
	SelectCopied OVStyle
	// PauseLine is the style that applies to the line where follow mode is paused.
	PauseLine OVStyle
}

// The name of the converter that can be specified.
const (
	convEscaped  string = "es"       // convEscaped processes escape sequence(default).
	convRaw      string = "raw"      // convRaw is displayed without processing escape sequences as they are.
	convAlign    string = "align"    // convAlign is aligned in each column.
	convWordWrap string = "wordwrap" // convWordWrap is wrapped at word boundaries.
)

const (
	nameGeneral string = "general"
)

// NewRunTimeSettings returns the structure of RunTimeSettings with default values.
func NewRunTimeSettings() RunTimeSettings {
	return RunTimeSettings{
		TabWidth:       8,
		MarkStyleWidth: 1,
		Converter:      convEscaped,
		OVPromptConfig: NewOVPromptConfig(),
		Style:          NewStyle(),
		StatusLine:     true,
		VScrollLines:   2,
	}
}

// NewStyle returns the structure of Style with default values.
func NewStyle() Style {
	return Style{
		Header: OVStyle{
			Bold: true,
		},
		Alternate: OVStyle{
			Background: "gray",
		},
		LineNumber: OVStyle{
			Bold: true,
		},
		SearchHighlight: OVStyle{
			Reverse: true,
		},
		ColumnHighlight: OVStyle{
			Reverse: true,
		},
		MarkLine: OVStyle{
			Background: "darkgoldenrod",
		},
		SectionLine: OVStyle{
			Background: "slateblue",
		},
		VerticalHeader: OVStyle{},
		VerticalHeaderBorder: OVStyle{
			Background: "#c0c0c0",
		},
		MultiColorHighlight: []OVStyle{
			{Foreground: "red"},
			{Foreground: "aqua"},
			{Foreground: "yellow"},
			{Foreground: "fuchsia"},
			{Foreground: "lime"},
			{Foreground: "blue"},
			{Foreground: "gray"},
		},
		ColumnRainbow: []OVStyle{
			{Foreground: "white"},
			{Foreground: "crimson"},
			{Foreground: "aqua"},
			{Foreground: "lightsalmon"},
			{Foreground: "lime"},
			{Foreground: "blue"},
			{Foreground: "yellowgreen"},
		},
		JumpTargetLine: OVStyle{
			Underline: true,
		},
		Ruler: OVStyle{
			Background: "#333333",
			Foreground: "#CCCCCC",
			Bold:       true,
		},
		SelectActive: OVStyle{
			Reverse: true,
		},
		SelectCopied: OVStyle{
			Background: "slategray",
		},
		PauseLine: OVStyle{
			Background: "#663333",
		},
	}
}

// setOldStyle applies deprecated style settings for backward compatibility.
//
// Deprecated: This function is planned to be removed in future versions.
// It reads and applies old style settings to maintain compatibility with older configurations.
// Use the new style configuration methods instead.
func setOldStyle(base RunTimeSettings, config Config) RunTimeSettings {
	blank := OVStyle{}
	if config.StyleBody != blank {
		base.Style.Body = config.StyleBody
	}
	if config.StyleHeader != blank {
		base.Style.Header = config.StyleHeader
	}
	if config.StyleLineNumber != blank {
		base.Style.LineNumber = config.StyleLineNumber
	}
	if config.StyleSearchHighlight != blank {
		base.Style.SearchHighlight = config.StyleSearchHighlight
	}
	if config.StyleColumnHighlight != blank {
		base.Style.ColumnHighlight = config.StyleColumnHighlight
	}
	if config.StyleMarkLine != blank {
		base.Style.MarkLine = config.StyleMarkLine
	}
	if config.StyleSectionLine != blank {
		base.Style.SectionLine = config.StyleSectionLine
	}
	if config.StyleVerticalHeader != blank {
		base.Style.VerticalHeader = config.StyleVerticalHeader
	}
	if config.StyleJumpTargetLine != blank {
		base.Style.JumpTargetLine = config.StyleJumpTargetLine
	}
	if config.StyleAlternate != blank {
		base.Style.Alternate = config.StyleAlternate
	}
	if config.StyleRuler != blank {
		base.Style.Ruler = config.StyleRuler
	}
	if config.StyleHeaderBorder != blank {
		base.Style.HeaderBorder = config.StyleHeaderBorder
	}
	if config.StyleSectionHeaderBorder != blank {
		base.Style.SectionHeaderBorder = config.StyleSectionHeaderBorder
	}
	if config.StyleVerticalHeaderBorder != blank {
		base.Style.VerticalHeaderBorder = config.StyleVerticalHeaderBorder
	}
	return base
}

// setOldPrompt applies deprecated prompt settings for backward compatibility.
//
// Deprecated: This function is planned to be removed in future versions.
func setOldPrompt(base RunTimeSettings, config Config) RunTimeSettings {
	prompt := config.Prompt
	// Old PromptConfig settings are loaded with lower priority.
	if prompt.Normal.ShowFilename != nil {
		base.OVPromptConfig.Normal.ShowFilename = *prompt.Normal.ShowFilename
	}
	if prompt.Normal.InvertColor != nil {
		base.OVPromptConfig.Normal.InvertColor = *prompt.Normal.InvertColor
	}
	if prompt.Normal.ProcessOfCount != nil {
		base.OVPromptConfig.Normal.ProcessOfCount = *prompt.Normal.ProcessOfCount
	}
	return base
}

// applyIfSet copies the dereferenced value of override into base if override is non-nil.
func applyIfSet[T any](base *T, override *T) {
	if override != nil {
		*base = *override
	}
}

// updateRunTimeSettings updates the RunTimeSettings.
func updateRunTimeSettings(base RunTimeSettings, override General) RunTimeSettings {
	applyIfSet(&base.TabWidth, override.TabWidth)
	applyIfSet(&base.Header, override.Header)
	applyIfSet(&base.VerticalHeader, override.VerticalHeader)
	applyIfSet(&base.HeaderColumn, override.HeaderColumn)
	applyIfSet(&base.SkipLines, override.SkipLines)
	applyIfSet(&base.WatchInterval, override.WatchInterval)
	applyIfSet(&base.MarkStyleWidth, override.MarkStyleWidth)
	applyIfSet(&base.SectionStartPosition, override.SectionStartPosition)
	applyIfSet(&base.SectionHeaderNum, override.SectionHeaderNum)
	applyIfSet(&base.HScrollWidth, override.HScrollWidth)
	applyIfSet(&base.HScrollWidthNum, override.HScrollWidthNum)
	applyIfSet(&base.VScrollLines, override.VScrollLines)
	applyIfSet(&base.RulerType, override.RulerType)
	applyIfSet(&base.AlternateRows, override.AlternateRows)
	applyIfSet(&base.ColumnMode, override.ColumnMode)
	applyIfSet(&base.ColumnWidth, override.ColumnWidth)
	applyIfSet(&base.ColumnRainbow, override.ColumnRainbow)
	applyIfSet(&base.LineNumMode, override.LineNumMode)
	applyIfSet(&base.WrapMode, override.WrapMode)
	applyIfSet(&base.FollowMode, override.FollowMode)
	applyIfSet(&base.FollowAll, override.FollowAll)
	applyIfSet(&base.FollowSection, override.FollowSection)
	applyIfSet(&base.FollowName, override.FollowName)
	applyIfSet(&base.PlainMode, override.PlainMode)
	applyIfSet(&base.SectionHeader, override.SectionHeader)
	applyIfSet(&base.HideOtherSection, override.HideOtherSection)
	applyIfSet(&base.StatusLine, override.StatusLine)
	applyIfSet(&base.ColumnDelimiter, override.ColumnDelimiter)
	applyIfSet(&base.SectionDelimiter, override.SectionDelimiter)
	applyIfSet(&base.JumpTarget, override.JumpTarget)
	applyIfSet(&base.MultiColorWords, override.MultiColorWords)
	applyIfSet(&base.Caption, override.Caption)
	applyIfSet(&base.Converter, override.Converter)
	if override.Align != nil && *override.Align {
		base.Converter = convAlign
	}
	if override.Raw != nil && *override.Raw {
		base.Converter = convRaw
	}
	if override.Wrap != nil {
		// Normalize wrap mode: support short forms (c, w)
		wrapMode := *override.Wrap
		switch wrapMode {
		case "w", "word":
			base.Converter = convWordWrap
			base.WrapMode = true
		case "f", "false", "no", "n", "0", "FALSE", "False":
			base.WrapMode = false
		default:
			base.WrapMode = true // Default to true for any other value, including "c", "char", "true", "yes", etc.
		}
	}
	base.OVPromptConfig = updatePromptConfig(base.OVPromptConfig, override.Prompt)
	base.Style = updateRuntimeStyle(base.Style, override.Style)
	return base
}

// updatePromptConfig updates the prompt configuration.
func updatePromptConfig(base OVPromptConfig, override PromptConfig) OVPromptConfig {
	applyIfSet(&base.Normal.InvertColor, override.Normal.InvertColor)
	applyIfSet(&base.Normal.ShowFilename, override.Normal.ShowFilename)
	applyIfSet(&base.Normal.ProcessOfCount, override.Normal.ProcessOfCount)
	applyIfSet(&base.Normal.CursorType, override.Normal.CursorType)
	applyIfSet(&base.Input.CursorType, override.Input.CursorType)
	return base
}

// updateRuntimeStyle updates the style.
func updateRuntimeStyle(base Style, override StyleConfig) Style {
	applyIfSet(&base.ColumnRainbow, override.ColumnRainbow)
	applyIfSet(&base.MultiColorHighlight, override.MultiColorHighlight)
	applyIfSet(&base.Header, override.Header)
	applyIfSet(&base.Body, override.Body)
	applyIfSet(&base.LineNumber, override.LineNumber)
	applyIfSet(&base.SearchHighlight, override.SearchHighlight)
	applyIfSet(&base.ColumnHighlight, override.ColumnHighlight)
	applyIfSet(&base.MarkLine, override.MarkLine)
	applyIfSet(&base.SectionLine, override.SectionLine)
	applyIfSet(&base.VerticalHeader, override.VerticalHeader)
	applyIfSet(&base.JumpTargetLine, override.JumpTargetLine)
	applyIfSet(&base.Alternate, override.Alternate)
	applyIfSet(&base.Ruler, override.Ruler)
	applyIfSet(&base.HeaderBorder, override.HeaderBorder)
	applyIfSet(&base.SectionHeaderBorder, override.SectionHeaderBorder)
	applyIfSet(&base.VerticalHeaderBorder, override.VerticalHeaderBorder)
	applyIfSet(&base.LeftStatus, override.LeftStatus)
	applyIfSet(&base.RightStatus, override.RightStatus)
	applyIfSet(&base.SelectActive, override.SelectActive)
	applyIfSet(&base.SelectCopied, override.SelectCopied)
	applyIfSet(&base.PauseLine, override.PauseLine)
	return base
}
