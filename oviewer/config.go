package oviewer

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
			Reverse: true,
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
