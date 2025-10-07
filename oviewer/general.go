package oviewer

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
	// HScrollWidth is the horizontal scroll width as a string (e.g., "80%", "40").
	HScrollWidth *string
	// HScrollWidthNum is the horizontal scroll width as an integer (number of columns).
	HScrollWidthNum *int
	// VScrollLines is the number of lines to scroll with the mouse wheel.
	VScrollLines *int
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
	// PromptConfig holds settings related to the command prompt.
	Prompt PromptConfig
	// Style is the style setting for general appearance and formatting.
	Style StyleConfig
}

// SetHeader sets the number of header lines to be fixed.
func (g *General) SetHeader(header int) {
	g.Header = &header
}

// SetTabWidth sets the tab stop number.
func (g *General) SetTabWidth(tabWidth int) {
	g.TabWidth = &tabWidth
}

// SetFollowMode sets the follow mode.
func (g *General) SetFollowMode(follow bool) {
	g.FollowMode = &follow
}

// SetCaption sets the caption.
func (g *General) SetCaption(caption string) {
	g.Caption = &caption
}

// SetMultiColorWords sets the multi-color words.
func (g *General) SetMultiColorWords(words []string) {
	copied := make([]string, len(words))
	copy(copied, words)
	g.MultiColorWords = &copied
}

// SetColumnMode sets the column mode.
func (g *General) SetColumnMode(mode bool) {
	g.ColumnMode = &mode
}

// SetStatusLine sets the status line visibility.
func (g *General) SetStatusLine(status bool) {
	g.StatusLine = &status
}

// SetConverter sets the converter name.
func (g *General) SetConverter(converter string) {
	g.Converter = &converter
}

// SetAlign sets the alignment.
func (g *General) SetAlign(align bool) {
	g.Align = &align
}

// SetRaw sets the raw mode.
func (g *General) SetRaw(raw bool) {
	g.Raw = &raw
}

// SetColumnDelimiter sets the column delimiter.
func (g *General) SetColumnDelimiter(delim string) {
	g.ColumnDelimiter = &delim
}

// SetSectionDelimiter sets the section delimiter.
func (g *General) SetSectionDelimiter(delim string) {
	g.SectionDelimiter = &delim
}

// SetJumpTarget sets the jump target string.
func (g *General) SetJumpTarget(target string) {
	g.JumpTarget = &target
}

// SetVerticalHeader sets the number of vertical header lines.
func (g *General) SetVerticalHeader(verticalHeader int) {
	g.VerticalHeader = &verticalHeader
}

// SetHeaderColumn sets the number of header columns to be fixed.
func (g *General) SetHeaderColumn(headerColumn int) {
	g.HeaderColumn = &headerColumn
}

// SetSkipLines sets the number of lines to skip.
func (g *General) SetSkipLines(skipLines int) {
	g.SkipLines = &skipLines
}

// SetWatchInterval sets the watch interval in seconds.
func (g *General) SetWatchInterval(interval int) {
	g.WatchInterval = &interval
}

// SetMarkStyleWidth sets the width for the marked line style.
func (g *General) SetMarkStyleWidth(width int) {
	g.MarkStyleWidth = &width
}

// SetSectionStartPosition sets the section start position.
func (g *General) SetSectionStartPosition(pos int) {
	g.SectionStartPosition = &pos
}

// SetSectionHeaderNum sets the number of lines in the section header.
func (g *General) SetSectionHeaderNum(num int) {
	g.SectionHeaderNum = &num
}

// SetHScrollWidth sets the horizontal scroll width as a string.
func (g *General) SetHScrollWidth(width string) {
	g.HScrollWidth = &width
}

// SetHScrollWidthNum sets the horizontal scroll width as an integer.
func (g *General) SetHScrollWidthNum(num int) {
	g.HScrollWidthNum = &num
}

func (g *General) SetVScrollLines(num int) {
	g.VScrollLines = &num
}

// SetRulerType sets the ruler type.
func (g *General) SetRulerType(rtype RulerType) {
	g.RulerType = &rtype
}

// SetAlternateRows sets the alternate row style.
func (g *General) SetAlternateRows(alternate bool) {
	g.AlternateRows = &alternate
}

// SetColumnWidth sets the column width mode.
func (g *General) SetColumnWidth(width bool) {
	g.ColumnWidth = &width
}

// SetColumnRainbow sets the column rainbow mode.
func (g *General) SetColumnRainbow(rainbow bool) {
	g.ColumnRainbow = &rainbow
}

// SetLineNumMode sets the line number display mode.
func (g *General) SetLineNumMode(lineNum bool) {
	g.LineNumMode = &lineNum
}

// SetWrapMode sets the wrap mode.
func (g *General) SetWrapMode(wrap bool) {
	g.WrapMode = &wrap
}

// SetFollowAll sets the follow mode for all documents.
func (g *General) SetFollowAll(followAll bool) {
	g.FollowAll = &followAll
}

// SetFollowSection sets the follow mode by section.
func (g *General) SetFollowSection(followSection bool) {
	g.FollowSection = &followSection
}

// SetFollowName sets the follow mode by file name.
func (g *General) SetFollowName(followName bool) {
	g.FollowName = &followName
}

// SetPlainMode sets the plain mode for character decoration.
func (g *General) SetPlainMode(plain bool) {
	g.PlainMode = &plain
}

// SetSectionHeader sets whether to display the section header.
func (g *General) SetSectionHeader(sectionHeader bool) {
	g.SectionHeader = &sectionHeader
}

// SetHideOtherSection sets whether to hide other sections.
func (g *General) SetHideOtherSection(hide bool) {
	g.HideOtherSection = &hide
}
