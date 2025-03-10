package oviewer

import (
	"context"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func candidateHelpter() *candidate {
	return &candidate{
		list: []string{"a", "b", "c"},
	}
}

func TestInput_keyEvent(t *testing.T) {
	t.Parallel()
	type fields struct {
		value   string
		cursorX int
	}
	type args struct {
		evKey *tcell.EventKey
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      bool
		wantValue string
	}{
		{
			name: "key enter",
			fields: fields{
				value:   "test",
				cursorX: 4,
			},
			args: args{
				evKey: tcell.NewEventKey(tcell.KeyEnter, 0, 0),
			},
			want:      true,
			wantValue: "test",
		},
		{
			name: "key cancel",
			fields: fields{
				value:   "test",
				cursorX: 4,
			},
			args: args{
				evKey: tcell.NewEventKey(tcell.KeyEscape, 0, 0),
			},
			want:      false,
			wantValue: "",
		},
		{
			name: "key rune",
			fields: fields{
				value:   "test",
				cursorX: 4,
			},
			args: args{
				evKey: tcell.NewEventKey(tcell.KeyRune, 'a', 0),
			},
			want:      false,
			wantValue: "testa",
		},
		{
			name: "key tab",
			fields: fields{
				value:   "test",
				cursorX: 4,
			},
			args: args{
				evKey: tcell.NewEventKey(tcell.KeyTAB, 0, 0),
			},
			want:      false,
			wantValue: "test\t",
		},
		{
			name: "key backspace",
			fields: fields{
				value:   "test",
				cursorX: 4,
			},
			args: args{
				evKey: tcell.NewEventKey(tcell.KeyBackspace, 0, 0),
			},
			want:      false,
			wantValue: "tes",
		},
		{
			name: "key delete",
			fields: fields{
				value:   "test",
				cursorX: 0,
			},
			args: args{
				evKey: tcell.NewEventKey(tcell.KeyDelete, 0, 0),
			},
			want:      false,
			wantValue: "est",
		},
		{
			name: "key left",
			fields: fields{
				value:   "test",
				cursorX: 4,
			},
			args: args{
				evKey: tcell.NewEventKey(tcell.KeyLeft, 0, 0),
			},
			want:      false,
			wantValue: "test",
		},
		{
			name: "key right",
			fields: fields{
				value:   "test",
				cursorX: 0,
			},
			args: args{
				evKey: tcell.NewEventKey(tcell.KeyRight, 0, 0),
			},
			want:      false,
			wantValue: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := &Input{
				value:   tt.fields.value,
				cursorX: tt.fields.cursorX,
			}
			got := input.keyEvent(tt.args.evKey)
			if got != tt.want {
				t.Errorf("Input.keyEvent() = %v, want %v", got, tt.want)
			}
			if tt.wantValue != input.value {
				t.Errorf("Input.value = %v, want %v", input.value, tt.wantValue)
			}
		})
	}
}

func Test_candidate_up(t *testing.T) {
	type fields struct {
		list []string
		p    int
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "testUp1",
			fields: fields{
				list: []string{"a", "b", "c"},
				p:    2,
			},
			want: "b",
		},
		{
			name: "testUp2",
			fields: fields{
				list: []string{"a", "b", "c"},
				p:    0,
			},
			want: "c",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := blankCandidate()
			c.list = tt.fields.list
			c.p = tt.fields.p
			if got := c.up(); got != tt.want {
				t.Errorf("candidate.up() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_candidate_down(t *testing.T) {
	type fields struct {
		list []string
		p    int
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "testDown1",
			fields: fields{
				list: []string{"a", "b", "c"},
				p:    0,
			},
			want: "b",
		},
		{
			name: "testDown2",
			fields: fields{
				list: []string{"a", "b", "c"},
				p:    2,
			},
			want: "a",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := blankCandidate()
			c.list = tt.fields.list
			c.p = tt.fields.p
			if got := c.down(); got != tt.want {
				t.Errorf("candidate.down() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInput_previous(t *testing.T) {
	type fields struct {
		Event   Eventer
		value   string
		cursorX int
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "testPrevious",
			fields: fields{
				Event:   newSearchEvent(candidateHelpter(), forward),
				value:   "test",
				cursorX: 4,
			},
			want: "c",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &Input{
				Event:   tt.fields.Event,
				value:   tt.fields.value,
				cursorX: tt.fields.cursorX,
			}
			input.previous()
			input.previous()
			if input.value != tt.want {
				t.Errorf("Input.value = %v, want %v", input.value, tt.want)
			}
		})
	}
}

func TestInput_next(t *testing.T) {
	type fields struct {
		Event   Eventer
		value   string
		cursorX int
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "testNext",
			fields: fields{
				Event:   newSearchEvent(candidateHelpter(), forward),
				value:   "test",
				cursorX: 4,
			},
			want: "a",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &Input{
				Event:   tt.fields.Event,
				value:   tt.fields.value,
				cursorX: tt.fields.cursorX,
			}
			input.next()
			if input.value != tt.want {
				t.Errorf("Input.value = %v, want %v", input.value, tt.want)
			}
		})
	}
}

func TestRoot_inputPrompt(t *testing.T) {
	root := rootHelper(t)
	ctx := context.Background()
	root.inputDelimiter(ctx)
	if root.inputPrompt() != "Delimiter:" {
		t.Errorf("Root.inputMode() = %v, want %v", root.inputPrompt(), "Delimiter:")
	}
	root.inputHeader(ctx)
	if root.inputPrompt() != "Header length:" {
		t.Errorf("Root.inputMode() = %v, want %v", root.inputPrompt(), "Header length:")
	}
	root.inputSkipLines(ctx)
	if root.inputPrompt() != "Skip lines:" {
		t.Errorf("Root.inputMode() = %v, want %v", root.inputPrompt(), "Skip lines:")
	}
	root.inputGoLine(ctx)
	if root.inputPrompt() != "Goto line:" {
		t.Errorf("Root.inputMode() = %v, want %v", root.inputPrompt(), "Goto line:")
	}
	root.inputMultiColor(ctx)
	if root.inputPrompt() != "Multi color:" {
		t.Errorf("Root.inputMode() = %v, want %v", root.inputPrompt(), "Multi color:")
	}
	root.input.Event = normal()
	if root.inputPrompt() != "" {
		t.Errorf("Root.inputMode() = %v, want %v", root.inputPrompt(), "")
	}
	root.inputForwardSearch(ctx)
	if root.inputPrompt() != "/" {
		t.Errorf("Root.inputMode() = %v, want %v", root.inputPrompt(), "/")
	}
	root.inputBackSearch(ctx)
	if root.inputPrompt() != "?" {
		t.Errorf("Root.inputMode() = %v, want %v", root.inputPrompt(), "?")
	}
	root.inputSearchFilter(ctx)
	if root.inputPrompt() != "&" {
		t.Errorf("Root.inputMode() = %v, want %v", root.inputPrompt(), "&")
	}
	root.inputJumpTarget(ctx)
	if root.inputPrompt() != "Jump target line:" {
		t.Errorf("Root.inputMode() = %v, want %v", root.inputPrompt(), "Jump target line:")
	}
	root.inputSaveBuffer(ctx)
	if root.inputPrompt() != "(Save)file:" {
		t.Errorf("Root.inputMode() = %v, want %v", root.inputPrompt(), "(Save)file:")
	}
	root.inputViewMode(ctx)
	if root.inputPrompt() != "Mode:" {
		t.Errorf("Root.inputMode() = %v, want %v", root.inputPrompt(), "Mode:")
	}
	root.inputWatchInterval(ctx)
	if root.inputPrompt() != "Watch interval:" {
		t.Errorf("Root.inputMode() = %v, want %v", root.inputPrompt(), "Watch interval:")
	}
	root.inputWriteBA(ctx)
	if root.inputPrompt() != "WriteAndQuit Before:After:" {
		t.Errorf("Root.inputMode() = %v, want %v", root.inputPrompt(), "WriteAndQuit Before:After:")
	}
	root.inputTabWidth(ctx)
	if root.inputPrompt() != "TAB width:" {
		t.Errorf("Root.inputMode() = %v, want %v", root.inputPrompt(), "TAB width:")
	}
	root.inputSectionDelimiter(ctx)
	if root.inputPrompt() != "Section delimiter:" {
		t.Errorf("Root.inputMode() = %v, want %v", root.inputPrompt(), "Section delimiter:")
	}
	root.inputSectionStart(ctx)
	if root.inputPrompt() != "Section start:" {
		t.Errorf("Root.inputMode() = %v, want %v", root.inputPrompt(), "Section start:")
	}
	root.inputSectionNum(ctx)
	if root.inputPrompt() != "Section Num:" {
		t.Errorf("Root.inputMode() = %v, want %v", root.inputPrompt(), "Section Num:")
	}
	root.inputSectionStart(ctx)
	if root.inputPrompt() != "Section start:" {
		t.Errorf("Root.inputMode() = %v, want %v", root.inputPrompt(), "Section start:")
	}
	root.inputConvert(ctx)
	if root.inputPrompt() != "Convert:" {
		t.Errorf("Root.inputMode() = %v, want %v", root.inputPrompt(), "Convert:")
	}
	root.inputVerticalHeader(ctx)
	if root.inputPrompt() != "Vertical header length:" {
		t.Errorf("Root.inputMode() = %v, want %v", root.inputPrompt(), "Vertical header length:")
	}
	root.inputHeaderColumn(ctx)
	if root.inputPrompt() != "Header column:" {
		t.Errorf("Root.inputMode() = %v, want %v", root.inputPrompt(), "Header column:")
	}
}

func TestRoot_inputState(t *testing.T) {
	root := rootHelper(t)
	ctx := context.Background()
	root.inputForwardSearch(ctx)
	root.setSearcher("test", false)
	root.toggleCaseSensitive(ctx)
	if root.searchOpt != "(Aa)" {
		t.Errorf("Root.inputState() = %v, want %v", root.searchOpt, "(Aa)")
	}
	root.toggleSmartCaseSensitive(ctx)
	if root.searchOpt != "(S)" {
		t.Errorf("Root.inputState() = %v, want %v", root.searchOpt, "(S)")
	}
	root.toggleIncSearch(ctx)
	if root.searchOpt != "(I)(S)" {
		t.Errorf("Root.inputState() = %v, want %v", root.searchOpt, "(I)(S)")
	}
	root.toggleRegexpSearch(ctx)
	if root.searchOpt != "(R)(I)(S)" {
		t.Errorf("Root.inputState() = %v, want %v", root.searchOpt, "(R)(I)(S)")
	}
	if root.searchOpt != "(R)(I)(S)" {
		t.Errorf("Root.inputState() = %v, want %v", root.searchOpt, "(R)(I)(S)")
	}
	root.toggleNonMatch(ctx)
	root.inputBackSearch(ctx)
	if root.searchOpt != "(R)(I)(S)" {
		t.Errorf("Root.inputState() = %v, want %v", root.searchOpt, "(R)(I)(S)")
	}
	root.inputSearchFilter(ctx)
	root.toggleNonMatch(ctx)
	if root.searchOpt != "Non-match(R)(S)" {
		t.Errorf("Root.inputState() = %v, want %v", root.searchOpt, "Non-match(R)(S)")
	}
	root.candidatePrevious(ctx)
	if root.searchOpt != "Non-match(R)(S)" {
		t.Errorf("Root.inputState() = %v, want %v", root.searchOpt, "Non-match(R)(S)")
	}
	root.candidateNext(ctx)
	if root.searchOpt != "Non-match(R)(S)" {
		t.Errorf("Root.inputState() = %v, want %v", root.searchOpt, "Non-match(R)(S)")
	}
	root.inputHeader(ctx)
	root.setPromptOpt()
	if root.searchOpt != "" {
		t.Errorf("Root.inputState() = %v, want %v", root.searchOpt, "")
	}
}

func TestRoot_inputUp(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name  string
		args  args
		Event Eventer
		want  string
	}{
		{
			name: "testConvertType",
			args: args{str: "raw"},
			Event: newConvertTypeEvent(
				&candidate{
					list: []string{convEscaped, convRaw, convAlign},
					p:    2,
				},
			),
			want: "raw",
		},
		{
			name: "testDelimiter",
			args: args{str: ","},
			Event: newDelimiterEvent(
				&candidate{
					list: []string{",", "|", "\t"},
					p:    2,
				},
			),
			want: "|",
		},
		{
			name: "testGoto",
			args: args{str: "1000"},
			Event: newGotoEvent(
				&candidate{
					list: []string{"1", "10", "100"},
					p:    2,
				},
			),
			want: "10",
		},
		{
			name:  "testHeader",
			args:  args{str: "1"},
			Event: newHeaderEvent(),
			want:  "2",
		},
		{
			name:  "testHeaderColumn",
			args:  args{str: "1"},
			Event: newHeaderColumnEvent(),
			want:  "2",
		},
		{
			name: "testJumpTarget",
			args: args{str: "1000"},
			Event: newJumpTargetEvent(
				&candidate{
					list: []string{"1", "10", "100"},
					p:    2,
				},
			),
			want: "10",
		},
		{
			name: "testMultiColor",
			args: args{str: "1"},
			Event: newMultiColorEvent(
				&candidate{
					list: []string{"1", "2", "3"},
					p:    2,
				},
			),
			want: "2",
		},
		{
			name:  "testNormal",
			args:  args{str: "test"},
			Event: normal(),
			want:  "",
		},
		{
			name: "testSaveBuffer",
			args: args{str: "test"},
			Event: newSaveBufferEvent(
				&candidate{
					list: []string{"test", "test2", "test3"},
					p:    2,
				},
			),
			want: "test2",
		},
		{
			name: "testSectionDelimiter",
			Event: newSectionDelimiterEvent(
				&candidate{
					list: []string{",", "|", "\t"},
					p:    2,
				},
			),
			want: "|",
		},
		{
			name:  "testSectionNum",
			args:  args{str: "1"},
			Event: newSectionNumEvent(),
			want:  "2",
		},
		{
			name: "testSectionStart",
			Event: newSectionStartEvent(
				&candidate{
					list: []string{"1", "10", "100"},
					p:    2,
				},
			),
			want: "10",
		},
		{
			name:  "testSkipLines",
			args:  args{str: "1"},
			Event: newSkipLinesEvent(),
			want:  "2",
		},
		{
			name: "testTabWidth",
			Event: newTabWidthEvent(
				&candidate{
					list: []string{"4", "8", "3"},
					p:    2,
				},
			),
			want: "8",
		},
		{
			name:  "testVerticalHeader",
			args:  args{str: "1"},
			Event: newVerticalHeaderEvent(),
			want:  "2",
		},
		{
			name: "testViewMode",
			args: args{str: "test"},
			Event: newViewModeEvent(
				&candidate{
					list: []string{"test", "test2", "test3"},
					p:    2,
				},
			),
			want: "test2",
		},
		{
			name: "testWatchInterval",
			args: args{str: "1"},
			Event: newWatchIntervalEvent(
				&candidate{
					list: []string{"1", "2", "3"},
					p:    2,
				},
			),
			want: "2",
		},
		{
			name: "testWriteBAMode",
			args: args{str: "test"},
			Event: newWriteBAEvent(
				&candidate{
					list: []string{"test", "test2", "test3"},
					p:    2,
				},
			),
			want: "test2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputEvent := tt.Event
			got := inputEvent.Up(tt.args.str)
			if got != tt.want {
				t.Errorf("Root.inputUp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoot_inputDown(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name  string
		args  args
		Event Eventer
		want  string
	}{
		{
			name: "testConvertType",
			args: args{str: "raw"},
			Event: newConvertTypeEvent(
				&candidate{
					list: []string{convEscaped, convRaw, convAlign},
					p:    0,
				},
			),
			want: "raw",
		},
		{
			name: "testDelimiter",
			args: args{str: ","},
			Event: newDelimiterEvent(
				&candidate{
					list: []string{",", "|", "\t"},
					p:    0,
				},
			),
			want: "|",
		},
		{
			name: "testGoto",
			args: args{str: "1000"},
			Event: newGotoEvent(
				&candidate{
					list: []string{"1", "10", "100"},
					p:    0,
				},
			),
			want: "10",
		},
		{
			name:  "testHeaderColumn",
			args:  args{str: "1"},
			Event: newHeaderColumnEvent(),
			want:  "0",
		},
		{
			name:  "testHeader",
			args:  args{str: "3"},
			Event: newHeaderEvent(),
			want:  "2",
		},
		{
			name: "testJumpTarget",
			args: args{str: "1000"},
			Event: newJumpTargetEvent(
				&candidate{
					list: []string{"1", "10", "100"},
					p:    0,
				},
			),
			want: "10",
		},
		{
			name: "testMultiColor",
			args: args{str: "1"},
			Event: newMultiColorEvent(
				&candidate{
					list: []string{"1", "2", "3"},
					p:    0,
				},
			),
			want: "2",
		},
		{
			name:  "testNormal",
			args:  args{str: "test"},
			Event: normal(),
			want:  "",
		},
		{
			name: "testSaveBuffer",
			args: args{str: "test"},
			Event: newSaveBufferEvent(
				&candidate{
					list: []string{"test", "test2", "test3"},
					p:    0,
				},
			),
			want: "test2",
		},
		{
			name: "testSectionDelimiter",
			Event: newSectionDelimiterEvent(
				&candidate{
					list: []string{",", "|", "\t"},
					p:    0,
				},
			),
			want: "|",
		},
		{
			name:  "testSectionNum",
			args:  args{str: "1"},
			Event: newSectionNumEvent(),
			want:  "0",
		},
		{
			name: "testSectionStart",
			Event: newSectionStartEvent(
				&candidate{
					list: []string{"1", "10", "100"},
					p:    0,
				},
			),
			want: "10",
		},
		{
			name:  "testSkipLines",
			args:  args{str: "1"},
			Event: newSkipLinesEvent(),
			want:  "0",
		},
		{
			name: "testTabWidth",
			args: args{str: "4"},
			Event: newTabWidthEvent(
				&candidate{
					list: []string{"4", "8", "3"},
					p:    2,
				},
			),
			want: "4",
		},
		{
			name:  "testVerticalHeader",
			args:  args{str: "4"},
			Event: newVerticalHeaderEvent(),
			want:  "3",
		},
		{
			name: "testViewMode",
			args: args{str: "test"},
			Event: newViewModeEvent(
				&candidate{
					list: []string{"test", "test2", "test3"},
					p:    0,
				},
			),
			want: "test2",
		},
		{
			name: "testWatchInterval",
			args: args{str: "1"},
			Event: newWatchIntervalEvent(
				&candidate{
					list: []string{"1", "2", "3"},
					p:    0,
				},
			),
			want: "2",
		},
		{
			name: "testWriteBAMode",
			args: args{str: "test"},
			Event: newWriteBAEvent(
				&candidate{
					list: []string{"test", "test2", "test3"},
					p:    0,
				},
			),
			want: "test2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.Event.Down(tt.args.str)
			if got != tt.want {
				t.Errorf("Root.inputDown() = %v, want %v", got, tt.want)
			}
		})
	}
}
