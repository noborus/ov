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
func setOldStyle(src RunTimeSettings, config Config) RunTimeSettings {
	blank := OVStyle{}
	if config.StyleBody != blank {
		src.Style.Body = config.StyleBody
	}
	if config.StyleHeader != blank {
		src.Style.Header = config.StyleHeader
	}
	if config.StyleLineNumber != blank {
		src.Style.LineNumber = config.StyleLineNumber
	}
	if config.StyleSearchHighlight != blank {
		src.Style.SearchHighlight = config.StyleSearchHighlight
	}
	if config.StyleColumnHighlight != blank {
		src.Style.ColumnHighlight = config.StyleColumnHighlight
	}
	if config.StyleMarkLine != blank {
		src.Style.MarkLine = config.StyleMarkLine
	}
	if config.StyleSectionLine != blank {
		src.Style.SectionLine = config.StyleSectionLine
	}
	if config.StyleVerticalHeader != blank {
		src.Style.VerticalHeader = config.StyleVerticalHeader
	}
	if config.StyleJumpTargetLine != blank {
		src.Style.JumpTargetLine = config.StyleJumpTargetLine
	}
	if config.StyleAlternate != blank {
		src.Style.Alternate = config.StyleAlternate
	}
	if config.StyleRuler != blank {
		src.Style.Ruler = config.StyleRuler
	}
	if config.StyleHeaderBorder != blank {
		src.Style.HeaderBorder = config.StyleHeaderBorder
	}
	if config.StyleSectionHeaderBorder != blank {
		src.Style.SectionHeaderBorder = config.StyleSectionHeaderBorder
	}
	if config.StyleVerticalHeaderBorder != blank {
		src.Style.VerticalHeaderBorder = config.StyleVerticalHeaderBorder
	}
	return src
}

// setOldPrompt applies deprecated prompt settings for backward compatibility.
//
// Deprecated: This function is planned to be removed in future versions.
func setOldPrompt(src RunTimeSettings, config Config) RunTimeSettings {
	prompt := config.Prompt
	// Old PromptConfig settings are loaded with lower priority.
	if prompt.Normal.ShowFilename != nil {
		src.OVPromptConfig.Normal.ShowFilename = *prompt.Normal.ShowFilename
	}
	if prompt.Normal.InvertColor != nil {
		src.OVPromptConfig.Normal.InvertColor = *prompt.Normal.InvertColor
	}
	if prompt.Normal.ProcessOfCount != nil {
		src.OVPromptConfig.Normal.ProcessOfCount = *prompt.Normal.ProcessOfCount
	}
	return src
}

// applyIfSet copies the dereferenced value of src into dst if src is non-nil.
func applyIfSet[T any](dst *T, src *T) {
	if src != nil {
		*dst = *src
	}
}

// updateRunTimeSettings updates the RunTimeSettings.
func updateRunTimeSettings(src RunTimeSettings, dst General) RunTimeSettings {
	applyIfSet(&src.TabWidth, dst.TabWidth)
	applyIfSet(&src.Header, dst.Header)
	applyIfSet(&src.VerticalHeader, dst.VerticalHeader)
	applyIfSet(&src.HeaderColumn, dst.HeaderColumn)
	applyIfSet(&src.SkipLines, dst.SkipLines)
	applyIfSet(&src.WatchInterval, dst.WatchInterval)
	applyIfSet(&src.MarkStyleWidth, dst.MarkStyleWidth)
	applyIfSet(&src.SectionStartPosition, dst.SectionStartPosition)
	applyIfSet(&src.SectionHeaderNum, dst.SectionHeaderNum)
	applyIfSet(&src.HScrollWidth, dst.HScrollWidth)
	applyIfSet(&src.HScrollWidthNum, dst.HScrollWidthNum)
	applyIfSet(&src.VScrollLines, dst.VScrollLines)
	applyIfSet(&src.RulerType, dst.RulerType)
	applyIfSet(&src.AlternateRows, dst.AlternateRows)
	applyIfSet(&src.ColumnMode, dst.ColumnMode)
	applyIfSet(&src.ColumnWidth, dst.ColumnWidth)
	applyIfSet(&src.ColumnRainbow, dst.ColumnRainbow)
	applyIfSet(&src.LineNumMode, dst.LineNumMode)
	applyIfSet(&src.WrapMode, dst.WrapMode)
	applyIfSet(&src.FollowMode, dst.FollowMode)
	applyIfSet(&src.FollowAll, dst.FollowAll)
	applyIfSet(&src.FollowSection, dst.FollowSection)
	applyIfSet(&src.FollowName, dst.FollowName)
	applyIfSet(&src.PlainMode, dst.PlainMode)
	applyIfSet(&src.SectionHeader, dst.SectionHeader)
	applyIfSet(&src.HideOtherSection, dst.HideOtherSection)
	applyIfSet(&src.StatusLine, dst.StatusLine)
	applyIfSet(&src.ColumnDelimiter, dst.ColumnDelimiter)
	applyIfSet(&src.SectionDelimiter, dst.SectionDelimiter)
	applyIfSet(&src.JumpTarget, dst.JumpTarget)
	applyIfSet(&src.MultiColorWords, dst.MultiColorWords)
	applyIfSet(&src.Caption, dst.Caption)
	applyIfSet(&src.Converter, dst.Converter)
	if dst.Align != nil && *dst.Align {
		src.Converter = convAlign
	}
	if dst.Raw != nil && *dst.Raw {
		src.Converter = convRaw
	}
	if dst.Wrap != nil {
		// Normalize wrap mode: support short forms (c, w)
		wrapMode := *dst.Wrap
		switch wrapMode {
		case "w", "word":
			src.Converter = convWordWrap
			src.WrapMode = true
		case "f", "false", "no", "n", "0", "FALSE", "False":
			src.WrapMode = false
		default:
			src.WrapMode = true // Default to true for any other value, including "c", "char", "true", "yes", etc.
		}
	}
	src.OVPromptConfig = updatePromptConfig(src.OVPromptConfig, dst.Prompt)
	src.Style = updateRuntimeStyle(src.Style, dst.Style)
	return src
}

// updatePromptConfig updates the prompt configuration.
func updatePromptConfig(src OVPromptConfig, dst PromptConfig) OVPromptConfig {
	applyIfSet(&src.Normal.InvertColor, dst.Normal.InvertColor)
	applyIfSet(&src.Normal.ShowFilename, dst.Normal.ShowFilename)
	applyIfSet(&src.Normal.ProcessOfCount, dst.Normal.ProcessOfCount)
	applyIfSet(&src.Normal.CursorType, dst.Normal.CursorType)
	applyIfSet(&src.Input.CursorType, dst.Input.CursorType)
	return src
}

// updateRuntimeStyle updates the style.
func updateRuntimeStyle(src Style, dst StyleConfig) Style {
	applyIfSet(&src.ColumnRainbow, dst.ColumnRainbow)
	applyIfSet(&src.MultiColorHighlight, dst.MultiColorHighlight)
	applyIfSet(&src.Header, dst.Header)
	applyIfSet(&src.Body, dst.Body)
	applyIfSet(&src.LineNumber, dst.LineNumber)
	applyIfSet(&src.SearchHighlight, dst.SearchHighlight)
	applyIfSet(&src.ColumnHighlight, dst.ColumnHighlight)
	applyIfSet(&src.MarkLine, dst.MarkLine)
	applyIfSet(&src.SectionLine, dst.SectionLine)
	applyIfSet(&src.VerticalHeader, dst.VerticalHeader)
	applyIfSet(&src.JumpTargetLine, dst.JumpTargetLine)
	applyIfSet(&src.Alternate, dst.Alternate)
	applyIfSet(&src.Ruler, dst.Ruler)
	applyIfSet(&src.HeaderBorder, dst.HeaderBorder)
	applyIfSet(&src.SectionHeaderBorder, dst.SectionHeaderBorder)
	applyIfSet(&src.VerticalHeaderBorder, dst.VerticalHeaderBorder)
	applyIfSet(&src.LeftStatus, dst.LeftStatus)
	applyIfSet(&src.RightStatus, dst.RightStatus)
	applyIfSet(&src.SelectActive, dst.SelectActive)
	applyIfSet(&src.SelectCopied, dst.SelectCopied)
	applyIfSet(&src.PauseLine, dst.PauseLine)
	return src
}
