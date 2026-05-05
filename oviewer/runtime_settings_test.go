package oviewer

import (
	"reflect"
	"testing"
)

func intPtr(i int) *int {
	return &i
}

func strPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func Test_updateRuntimeSettings(t *testing.T) {
	type args struct {
		runtime       RunTimeSettings
		configGeneral General
	}
	tests := []struct {
		name string
		args args
		want RunTimeSettings
	}{
		{
			name: "test1",
			args: args{
				runtime: RunTimeSettings{
					TabWidth: 4,
					Header:   1,
				},
				configGeneral: General{},
			},
			want: RunTimeSettings{
				TabWidth: 4,
				Header:   1,
			},
		},
		{
			name: "test2",
			args: args{
				runtime: RunTimeSettings{
					TabWidth: 4,
					Header:   1,
				},
				configGeneral: General{
					TabWidth: intPtr(8),
					Header:   intPtr(2),
				},
			},
			want: RunTimeSettings{
				TabWidth: 8,
				Header:   2,
			},
		},
		{
			name: "test3",
			args: args{
				runtime: RunTimeSettings{
					SkipLines: 3,
				},
				configGeneral: General{
					TabWidth: intPtr(8),
					Header:   intPtr(2),
				},
			},
			want: RunTimeSettings{
				TabWidth:  8,
				Header:    2,
				SkipLines: 3,
			},
		},
		{
			name: "test4",
			args: args{
				runtime: RunTimeSettings{
					SkipLines:  5,
					ColumnMode: true,
				},
				configGeneral: General{
					TabWidth:  intPtr(8),
					Header:    intPtr(2),
					SkipLines: intPtr(3),
				},
			},
			want: RunTimeSettings{
				TabWidth:   8,
				Header:     2,
				SkipLines:  3,
				ColumnMode: true,
			},
		},
		{
			name: "test5",
			args: args{
				runtime: RunTimeSettings{
					ColumnWidth: true,
					LineNumMode: true,
				},
				configGeneral: General{
					TabWidth:   intPtr(8),
					Header:     intPtr(2),
					SkipLines:  intPtr(3),
					ColumnMode: boolPtr(true),
				},
			},
			want: RunTimeSettings{
				TabWidth:    8,
				Header:      2,
				SkipLines:   3,
				ColumnMode:  true,
				ColumnWidth: true,
				LineNumMode: true,
			},
		},
		{
			name: "test6",
			args: args{
				runtime: RunTimeSettings{
					WrapMode:   false,
					FollowMode: true,
				},
				configGeneral: General{
					TabWidth:    intPtr(8),
					Header:      intPtr(2),
					SkipLines:   intPtr(3),
					ColumnMode:  boolPtr(true),
					ColumnWidth: boolPtr(true),
					LineNumMode: boolPtr(true),
				},
			},
			want: RunTimeSettings{
				TabWidth:    8,
				Header:      2,
				SkipLines:   3,
				ColumnMode:  true,
				ColumnWidth: true,
				LineNumMode: true,
				WrapMode:    false,
				FollowMode:  true,
			},
		},
		{
			name: "test7",
			args: args{
				runtime: RunTimeSettings{
					FollowAll:     true,
					FollowSection: true,
				},
				configGeneral: General{
					TabWidth:    intPtr(8),
					Header:      intPtr(2),
					SkipLines:   intPtr(3),
					ColumnMode:  boolPtr(true),
					ColumnWidth: boolPtr(true),
					LineNumMode: boolPtr(true),
					WrapMode:    boolPtr(false),
					FollowMode:  boolPtr(true),
				},
			},
			want: RunTimeSettings{
				TabWidth:      8,
				Header:        2,
				SkipLines:     3,
				ColumnMode:    true,
				ColumnWidth:   true,
				LineNumMode:   true,
				WrapMode:      false,
				FollowMode:    true,
				FollowAll:     true,
				FollowSection: true,
			},
		},
		{
			name: "test8",
			args: args{
				runtime: RunTimeSettings{
					FollowName:      true,
					ColumnDelimiter: ",",
				},
				configGeneral: General{
					TabWidth:      intPtr(8),
					Header:        intPtr(2),
					SkipLines:     intPtr(3),
					ColumnMode:    boolPtr(true),
					ColumnWidth:   boolPtr(true),
					LineNumMode:   boolPtr(true),
					WrapMode:      boolPtr(false),
					FollowMode:    boolPtr(true),
					FollowAll:     boolPtr(true),
					FollowSection: boolPtr(true),
				},
			},
			want: RunTimeSettings{
				TabWidth:        8,
				Header:          2,
				SkipLines:       3,
				ColumnMode:      true,
				ColumnWidth:     true,
				LineNumMode:     true,
				WrapMode:        false,
				FollowMode:      true,
				FollowAll:       true,
				FollowSection:   true,
				FollowName:      true,
				ColumnDelimiter: ",",
			},
		},
		{
			name: "test9",
			args: args{
				runtime: RunTimeSettings{
					WatchInterval:  10,
					MarkStyleWidth: 2,
				},
				configGeneral: General{
					TabWidth:        intPtr(8),
					Header:          intPtr(2),
					SkipLines:       intPtr(3),
					ColumnMode:      boolPtr(true),
					ColumnWidth:     boolPtr(true),
					LineNumMode:     boolPtr(true),
					WrapMode:        boolPtr(false),
					FollowMode:      boolPtr(true),
					FollowAll:       boolPtr(true),
					FollowSection:   boolPtr(true),
					FollowName:      boolPtr(true),
					ColumnDelimiter: strPtr(","),
				},
			},
			want: RunTimeSettings{
				TabWidth:        8,
				Header:          2,
				SkipLines:       3,
				ColumnMode:      true,
				ColumnWidth:     true,
				LineNumMode:     true,
				WrapMode:        false,
				FollowMode:      true,
				FollowAll:       true,
				FollowSection:   true,
				FollowName:      true,
				ColumnDelimiter: ",",
				WatchInterval:   10,
				MarkStyleWidth:  2,
			},
		},
		{
			name: "test10",
			args: args{
				runtime: RunTimeSettings{
					SectionDelimiter:     "##",
					SectionStartPosition: 5,
				},
				configGeneral: General{
					TabWidth:        intPtr(8),
					Header:          intPtr(2),
					SkipLines:       intPtr(3),
					ColumnMode:      boolPtr(true),
					ColumnWidth:     boolPtr(true),
					LineNumMode:     boolPtr(true),
					WrapMode:        boolPtr(false),
					FollowMode:      boolPtr(true),
					FollowAll:       boolPtr(true),
					FollowSection:   boolPtr(true),
					FollowName:      boolPtr(true),
					ColumnDelimiter: strPtr(","),
					WatchInterval:   intPtr(10),
					MarkStyleWidth:  intPtr(2),
				},
			},
			want: RunTimeSettings{
				TabWidth:             8,
				Header:               2,
				SkipLines:            3,
				ColumnMode:           true,
				ColumnWidth:          true,
				LineNumMode:          true,
				WrapMode:             false,
				FollowMode:           true,
				FollowAll:            true,
				FollowSection:        true,
				FollowName:           true,
				ColumnDelimiter:      ",",
				WatchInterval:        10,
				MarkStyleWidth:       2,
				SectionDelimiter:     "##",
				SectionStartPosition: 5,
			},
		},
		{
			name: "test11",
			args: args{
				runtime: RunTimeSettings{
					Caption: "test caption",
				},
				configGeneral: General{
					Wrap: strPtr("word"),
				},
			},
			want: RunTimeSettings{
				Caption:   "test caption",
				Converter: convWordWrap,
				WrapMode:  true,
			},
		},
		{
			name: "test12",
			args: args{
				runtime: RunTimeSettings{
					Caption: "test caption",
				},
				configGeneral: General{
					Wrap: strPtr("false"),
				},
			},
			want: RunTimeSettings{
				Caption:  "test caption",
				WrapMode: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := updateRunTimeSettings(tt.args.runtime, tt.args.configGeneral); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("updateRuntimeSettings() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func Test_updateRuntimeSettings_AllFields(t *testing.T) {
	type args struct {
		runtime       RunTimeSettings
		configGeneral General
	}
	tests := []struct {
		name string
		args args
		want RunTimeSettings
	}{
		{
			name: "testEmpty",
			args: args{
				runtime:       RunTimeSettings{},
				configGeneral: General{},
			},
			want: RunTimeSettings{},
		},
		{
			name: "testNoChange",
			args: args{
				runtime: RunTimeSettings{
					TabWidth:             4,
					Header:               1,
					VerticalHeader:       2,
					HeaderColumn:         3,
					SkipLines:            5,
					WatchInterval:        10,
					MarkStyleWidth:       2,
					SectionStartPosition: 6,
					SectionHeaderNum:     1,
					HScrollWidth:         "10%",
					HScrollWidthNum:      10,
					RulerType:            RulerRelative,
					AlternateRows:        true,
					ColumnMode:           true,
					ColumnWidth:          true,
					ColumnRainbow:        true,
					LineNumMode:          true,
					WrapMode:             true,
					FollowMode:           true,
					FollowAll:            true,
					FollowSection:        true,
					FollowName:           true,
					PlainMode:            true,
					SectionHeader:        true,
					HideOtherSection:     true,
					ColumnDelimiter:      ",",
					SectionDelimiter:     "##",
					JumpTarget:           "target",
					MultiColorWords:      []string{"word1", "word2"},
					Caption:              "test caption",
					Converter:            convAlign,
				},
				configGeneral: General{},
			},
			want: RunTimeSettings{
				TabWidth:             4,
				Header:               1,
				VerticalHeader:       2,
				HeaderColumn:         3,
				SkipLines:            5,
				WatchInterval:        10,
				MarkStyleWidth:       2,
				SectionStartPosition: 6,
				SectionHeaderNum:     1,
				HScrollWidth:         "10%",
				HScrollWidthNum:      10,
				RulerType:            RulerRelative,
				AlternateRows:        true,
				ColumnMode:           true,
				ColumnWidth:          true,
				ColumnRainbow:        true,
				LineNumMode:          true,
				WrapMode:             true,
				FollowMode:           true,
				FollowAll:            true,
				FollowSection:        true,
				FollowName:           true,
				PlainMode:            true,
				SectionHeader:        true,
				HideOtherSection:     true,
				ColumnDelimiter:      ",",
				SectionDelimiter:     "##",
				JumpTarget:           "target",
				MultiColorWords:      []string{"word1", "word2"},
				Caption:              "test caption",
				Converter:            convAlign,
			},
		},
		{
			name: "testAllFields",
			args: args{
				runtime: RunTimeSettings{
					TabWidth:             4,
					Header:               1,
					VerticalHeader:       2,
					HeaderColumn:         3,
					SkipLines:            5,
					WatchInterval:        10,
					MarkStyleWidth:       2,
					SectionStartPosition: 6,
					SectionHeaderNum:     1,
					HScrollWidth:         "10%",
					HScrollWidthNum:      10,
					RulerType:            RulerRelative,
					AlternateRows:        true,
					ColumnMode:           true,
					ColumnWidth:          true,
					ColumnRainbow:        true,
					LineNumMode:          true,
					WrapMode:             true,
					FollowMode:           true,
					FollowAll:            true,
					FollowSection:        true,
					FollowName:           true,
					PlainMode:            true,
					SectionHeader:        true,
					HideOtherSection:     true,
					ColumnDelimiter:      ",",
					SectionDelimiter:     "##",
					JumpTarget:           "target",
					MultiColorWords:      []string{"word1", "word2"},
					Caption:              "test caption",
					Converter:            convAlign,
				},
				configGeneral: General{
					TabWidth:             intPtr(8),
					Header:               intPtr(2),
					VerticalHeader:       intPtr(3),
					HeaderColumn:         intPtr(4),
					SkipLines:            intPtr(6),
					WatchInterval:        intPtr(15),
					MarkStyleWidth:       intPtr(3),
					SectionStartPosition: intPtr(7),
					SectionHeaderNum:     intPtr(2),
					HScrollWidth:         strPtr("20%"),
					HScrollWidthNum:      intPtr(20),
					RulerType:            (*RulerType)(intPtr(int(RulerAbsolute))),
					AlternateRows:        boolPtr(false),
					ColumnMode:           boolPtr(false),
					ColumnWidth:          boolPtr(false),
					ColumnRainbow:        boolPtr(false),
					LineNumMode:          boolPtr(false),
					WrapMode:             boolPtr(false),
					FollowMode:           boolPtr(false),
					FollowAll:            boolPtr(false),
					FollowSection:        boolPtr(false),
					FollowName:           boolPtr(false),
					PlainMode:            boolPtr(false),
					SectionHeader:        boolPtr(false),
					HideOtherSection:     boolPtr(false),
					ColumnDelimiter:      strPtr("\t"),
					SectionDelimiter:     strPtr("###"),
					JumpTarget:           strPtr("newTarget"),
					MultiColorWords:      &[]string{"newWord1", "newWord2"},
					Caption:              strPtr("new caption"),
					Converter:            strPtr(convRaw),
				},
			},
			want: RunTimeSettings{
				TabWidth:             8,
				Header:               2,
				VerticalHeader:       3,
				HeaderColumn:         4,
				SkipLines:            6,
				WatchInterval:        15,
				MarkStyleWidth:       3,
				SectionStartPosition: 7,
				SectionHeaderNum:     2,
				HScrollWidth:         "20%",
				HScrollWidthNum:      20,
				RulerType:            RulerAbsolute,
				AlternateRows:        false,
				ColumnMode:           false,
				ColumnWidth:          false,
				ColumnRainbow:        false,
				LineNumMode:          false,
				WrapMode:             false,
				FollowMode:           false,
				FollowAll:            false,
				FollowSection:        false,
				FollowName:           false,
				PlainMode:            false,
				SectionHeader:        false,
				HideOtherSection:     false,
				ColumnDelimiter:      "\t",
				SectionDelimiter:     "###",
				JumpTarget:           "newTarget",
				MultiColorWords:      []string{"newWord1", "newWord2"},
				Caption:              "new caption",
				Converter:            convRaw,
			},
		},
		{
			name: "testAlign",
			args: args{
				runtime: RunTimeSettings{
					Converter: convEscaped,
				},
				configGeneral: General{
					Align: boolPtr(true),
				},
			},
			want: RunTimeSettings{
				Converter: convAlign,
			},
		},
		{
			name: "testRaw",
			args: args{
				runtime: RunTimeSettings{
					Converter: convEscaped,
				},
				configGeneral: General{
					Raw: boolPtr(true),
				},
			},
			want: RunTimeSettings{
				Converter: convRaw,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := updateRunTimeSettings(tt.args.runtime, tt.args.configGeneral); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("updateRuntimeSettings() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func Test_updateRuntimeStyle(t *testing.T) {
	redStyle := OVStyle{Foreground: "red"}
	blueStyle := OVStyle{Foreground: "blue"}
	boldStyle := OVStyle{Bold: true}

	rainbowStyles := []OVStyle{{Foreground: "red"}, {Foreground: "green"}}
	newRainbowStyles := []OVStyle{{Foreground: "yellow"}, {Foreground: "purple"}}

	multiColorStyles := []OVStyle{{Foreground: "cyan"}, {Foreground: "magenta"}}
	newMultiColorStyles := []OVStyle{{Foreground: "white"}, {Foreground: "black"}}

	type args struct {
		base Style
		dst  StyleConfig
	}
	tests := []struct {
		name string
		args args
		want Style
	}{
		{
			name: "empty update",
			args: args{
				base: NewStyle(),
				dst:  StyleConfig{},
			},
			want: NewStyle(),
		},
		{
			name: "update column rainbow",
			args: args{
				base: Style{ColumnRainbow: rainbowStyles},
				dst:  StyleConfig{ColumnRainbow: &newRainbowStyles},
			},
			want: Style{ColumnRainbow: newRainbowStyles},
		},
		{
			name: "update multi color highlight",
			args: args{
				base: Style{MultiColorHighlight: multiColorStyles},
				dst:  StyleConfig{MultiColorHighlight: &newMultiColorStyles},
			},
			want: Style{MultiColorHighlight: newMultiColorStyles},
		},
		{
			name: "update header",
			args: args{
				base: Style{Header: redStyle},
				dst:  StyleConfig{Header: &blueStyle},
			},
			want: Style{Header: blueStyle},
		},
		{
			name: "update body",
			args: args{
				base: Style{Body: redStyle},
				dst:  StyleConfig{Body: &blueStyle},
			},
			want: Style{Body: blueStyle},
		},
		{
			name: "update line number",
			args: args{
				base: Style{LineNumber: redStyle},
				dst:  StyleConfig{LineNumber: &blueStyle},
			},
			want: Style{LineNumber: blueStyle},
		},
		{
			name: "update search highlight",
			args: args{
				base: Style{SearchHighlight: redStyle},
				dst:  StyleConfig{SearchHighlight: &blueStyle},
			},
			want: Style{SearchHighlight: blueStyle},
		},
		{
			name: "update column highlight",
			args: args{
				base: Style{ColumnHighlight: redStyle},
				dst:  StyleConfig{ColumnHighlight: &blueStyle},
			},
			want: Style{ColumnHighlight: blueStyle},
		},
		{
			name: "update mark line",
			args: args{
				base: Style{MarkLine: redStyle},
				dst:  StyleConfig{MarkLine: &blueStyle},
			},
			want: Style{MarkLine: blueStyle},
		},
		{
			name: "update section line",
			args: args{
				base: Style{SectionLine: redStyle},
				dst:  StyleConfig{SectionLine: &blueStyle},
			},
			want: Style{SectionLine: blueStyle},
		},
		{
			name: "update vertical header",
			args: args{
				base: Style{VerticalHeader: redStyle},
				dst:  StyleConfig{VerticalHeader: &blueStyle},
			},
			want: Style{VerticalHeader: blueStyle},
		},
		{
			name: "update jump target line",
			args: args{
				base: Style{JumpTargetLine: redStyle},
				dst:  StyleConfig{JumpTargetLine: &blueStyle},
			},
			want: Style{JumpTargetLine: blueStyle},
		},
		{
			name: "update alternate",
			args: args{
				base: Style{Alternate: redStyle},
				dst:  StyleConfig{Alternate: &blueStyle},
			},
			want: Style{Alternate: blueStyle},
		},
		{
			name: "update ruler",
			args: args{
				base: Style{Ruler: redStyle},
				dst:  StyleConfig{Ruler: &boldStyle},
			},
			want: Style{Ruler: boldStyle},
		},
		{
			name: "update header border",
			args: args{
				base: Style{HeaderBorder: redStyle},
				dst:  StyleConfig{HeaderBorder: &blueStyle},
			},
			want: Style{HeaderBorder: blueStyle},
		},
		{
			name: "update section header border",
			args: args{
				base: Style{SectionHeaderBorder: redStyle},
				dst:  StyleConfig{SectionHeaderBorder: &blueStyle},
			},
			want: Style{SectionHeaderBorder: blueStyle},
		},
		{
			name: "update vertical header border",
			args: args{
				base: Style{VerticalHeaderBorder: redStyle},
				dst:  StyleConfig{VerticalHeaderBorder: &blueStyle},
			},
			want: Style{VerticalHeaderBorder: blueStyle},
		},
		{
			name: "update multiple fields",
			args: args{
				base: Style{
					Header:       redStyle,
					Body:         redStyle,
					LineNumber:   redStyle,
					MarkLine:     redStyle,
					SectionLine:  redStyle,
					Ruler:        redStyle,
					HeaderBorder: redStyle,
				},
				dst: StyleConfig{
					Header:       &blueStyle,
					Body:         &blueStyle,
					Ruler:        &boldStyle,
					HeaderBorder: &boldStyle,
				},
			},
			want: Style{
				Header:       blueStyle,
				Body:         blueStyle,
				LineNumber:   redStyle,
				MarkLine:     redStyle,
				SectionLine:  redStyle,
				Ruler:        boldStyle,
				HeaderBorder: boldStyle,
			},
		},
		{
			name: "complete style update",
			args: args{
				base: NewStyle(), // Starting with default style
				dst: StyleConfig{
					ColumnRainbow:        &newRainbowStyles,
					MultiColorHighlight:  &newMultiColorStyles,
					Header:               &blueStyle,
					Body:                 &blueStyle,
					LineNumber:           &blueStyle,
					SearchHighlight:      &blueStyle,
					ColumnHighlight:      &blueStyle,
					MarkLine:             &blueStyle,
					SectionLine:          &blueStyle,
					VerticalHeader:       &blueStyle,
					JumpTargetLine:       &blueStyle,
					Alternate:            &blueStyle,
					Ruler:                &boldStyle,
					HeaderBorder:         &boldStyle,
					SectionHeaderBorder:  &boldStyle,
					VerticalHeaderBorder: &boldStyle,
					SelectActive:         &blueStyle,
					SelectCopied:         &blueStyle,
					PauseLine:            &blueStyle,
				},
			},
			want: Style{
				ColumnRainbow:        newRainbowStyles,
				MultiColorHighlight:  newMultiColorStyles,
				Header:               blueStyle,
				Body:                 blueStyle,
				LineNumber:           blueStyle,
				SearchHighlight:      blueStyle,
				ColumnHighlight:      blueStyle,
				MarkLine:             blueStyle,
				SectionLine:          blueStyle,
				VerticalHeader:       blueStyle,
				JumpTargetLine:       blueStyle,
				Alternate:            blueStyle,
				Ruler:                boldStyle,
				HeaderBorder:         boldStyle,
				SectionHeaderBorder:  boldStyle,
				VerticalHeaderBorder: boldStyle,
				SelectActive:         blueStyle,
				SelectCopied:         blueStyle,
				PauseLine:            blueStyle,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := updateRuntimeStyle(tt.args.base, tt.args.dst); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("updateRuntimeStyle() = %v, want %v", got, tt.want)
			}
		})
	}
}
