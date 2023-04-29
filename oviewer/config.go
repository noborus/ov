package oviewer

// NewConfig return the structure of Config with default values.
func NewConfig() Config {
	return Config{
		LoadChunksLimit:     -1,
		FileLoadChunksLimit: 100,
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
			TabWidth:             8,
			MarkStyleWidth:       1,
			SectionStartPosition: 0,
			JumpTarget:           0,
		},
	}
}
