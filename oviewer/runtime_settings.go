package oviewer

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

// updateRunTimeSettings updates the RunTimeSettings.
func updateRunTimeSettings(src RunTimeSettings, dst General) RunTimeSettings {
	if dst.TabWidth != nil {
		src.TabWidth = *dst.TabWidth
	}
	if dst.Header != nil {
		src.Header = *dst.Header
	}
	if dst.VerticalHeader != nil {
		src.VerticalHeader = *dst.VerticalHeader
	}
	if dst.HeaderColumn != nil {
		src.HeaderColumn = *dst.HeaderColumn
	}
	if dst.SkipLines != nil {
		src.SkipLines = *dst.SkipLines
	}
	if dst.WatchInterval != nil {
		src.WatchInterval = *dst.WatchInterval
	}
	if dst.MarkStyleWidth != nil {
		src.MarkStyleWidth = *dst.MarkStyleWidth
	}
	if dst.SectionStartPosition != nil {
		src.SectionStartPosition = *dst.SectionStartPosition
	}
	if dst.SectionHeaderNum != nil {
		src.SectionHeaderNum = *dst.SectionHeaderNum
	}
	if dst.HScrollWidth != nil {
		src.HScrollWidth = *dst.HScrollWidth
	}
	if dst.HScrollWidthNum != nil {
		src.HScrollWidthNum = *dst.HScrollWidthNum
	}
	if dst.VScrollLines != nil {
		src.VScrollLines = *dst.VScrollLines
	}
	if dst.RulerType != nil {
		src.RulerType = *dst.RulerType
	}
	if dst.AlternateRows != nil {
		src.AlternateRows = *dst.AlternateRows
	}
	if dst.ColumnMode != nil {
		src.ColumnMode = *dst.ColumnMode
	}
	if dst.ColumnWidth != nil {
		src.ColumnWidth = *dst.ColumnWidth
	}
	if dst.ColumnRainbow != nil {
		src.ColumnRainbow = *dst.ColumnRainbow
	}
	if dst.LineNumMode != nil {
		src.LineNumMode = *dst.LineNumMode
	}
	if dst.WrapMode != nil {
		src.WrapMode = *dst.WrapMode
	}
	if dst.FollowMode != nil {
		src.FollowMode = *dst.FollowMode
	}
	if dst.FollowAll != nil {
		src.FollowAll = *dst.FollowAll
	}
	if dst.FollowSection != nil {
		src.FollowSection = *dst.FollowSection
	}
	if dst.FollowName != nil {
		src.FollowName = *dst.FollowName
	}
	if dst.PlainMode != nil {
		src.PlainMode = *dst.PlainMode
	}
	if dst.SectionHeader != nil {
		src.SectionHeader = *dst.SectionHeader
	}
	if dst.HideOtherSection != nil {
		src.HideOtherSection = *dst.HideOtherSection
	}
	if dst.StatusLine != nil {
		src.StatusLine = *dst.StatusLine
	}
	if dst.ColumnDelimiter != nil {
		src.ColumnDelimiter = *dst.ColumnDelimiter
	}
	if dst.SectionDelimiter != nil {
		src.SectionDelimiter = *dst.SectionDelimiter
	}
	if dst.JumpTarget != nil {
		src.JumpTarget = *dst.JumpTarget
	}
	if dst.MultiColorWords != nil {
		src.MultiColorWords = *dst.MultiColorWords
	}
	if dst.Caption != nil {
		src.Caption = *dst.Caption
	}
	if dst.Converter != nil {
		src.Converter = *dst.Converter
	}
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
	if dst.Normal.InvertColor != nil {
		src.Normal.InvertColor = *dst.Normal.InvertColor
	}
	if dst.Normal.ShowFilename != nil {
		src.Normal.ShowFilename = *dst.Normal.ShowFilename
	}
	if dst.Normal.ProcessOfCount != nil {
		src.Normal.ProcessOfCount = *dst.Normal.ProcessOfCount
	}
	if dst.Normal.CursorType != nil {
		src.Normal.CursorType = *dst.Normal.CursorType
	}
	if dst.Input.CursorType != nil {
		src.Input.CursorType = *dst.Input.CursorType
	}
	return src
}

// updateRuntimeStyle updates the style.
func updateRuntimeStyle(src Style, dst StyleConfig) Style {
	if dst.ColumnRainbow != nil {
		src.ColumnRainbow = *dst.ColumnRainbow
	}
	if dst.MultiColorHighlight != nil {
		src.MultiColorHighlight = *dst.MultiColorHighlight
	}
	if dst.Header != nil {
		src.Header = *dst.Header
	}
	if dst.Body != nil {
		src.Body = *dst.Body
	}
	if dst.LineNumber != nil {
		src.LineNumber = *dst.LineNumber
	}
	if dst.SearchHighlight != nil {
		src.SearchHighlight = *dst.SearchHighlight
	}
	if dst.ColumnHighlight != nil {
		src.ColumnHighlight = *dst.ColumnHighlight
	}
	if dst.MarkLine != nil {
		src.MarkLine = *dst.MarkLine
	}
	if dst.SectionLine != nil {
		src.SectionLine = *dst.SectionLine
	}
	if dst.VerticalHeader != nil {
		src.VerticalHeader = *dst.VerticalHeader
	}
	if dst.JumpTargetLine != nil {
		src.JumpTargetLine = *dst.JumpTargetLine
	}
	if dst.Alternate != nil {
		src.Alternate = *dst.Alternate
	}
	if dst.Ruler != nil {
		src.Ruler = *dst.Ruler
	}
	if dst.HeaderBorder != nil {
		src.HeaderBorder = *dst.HeaderBorder
	}
	if dst.SectionHeaderBorder != nil {
		src.SectionHeaderBorder = *dst.SectionHeaderBorder
	}
	if dst.VerticalHeaderBorder != nil {
		src.VerticalHeaderBorder = *dst.VerticalHeaderBorder
	}
	if dst.LeftStatus != nil {
		src.LeftStatus = *dst.LeftStatus
	}
	if dst.RightStatus != nil {
		src.RightStatus = *dst.RightStatus
	}
	if dst.SelectActive != nil {
		src.SelectActive = *dst.SelectActive
	}
	if dst.SelectCopied != nil {
		src.SelectCopied = *dst.SelectCopied
	}
	if dst.PauseLine != nil {
		src.PauseLine = *dst.PauseLine
	}
	return src
}
