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
		tt := tt
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

func TestRoot_inputMode(t *testing.T) {
	root := rootHelper(t)
	ctx := context.Background()
	root.setDelimiterMode(ctx)
	if root.inputPrompt() != "Delimiter:" {
		t.Errorf("Root.inputMode() = %v, want %v", root.inputPrompt(), "Delimiter:")
	}
	root.setHeaderMode(ctx)
	if root.inputPrompt() != "Header length:" {
		t.Errorf("Root.inputMode() = %v, want %v", root.inputPrompt(), "Header length:")
	}
	root.setSkipLinesMode(ctx)
	if root.inputPrompt() != "Skip lines:" {
		t.Errorf("Root.inputMode() = %v, want %v", root.inputPrompt(), "Skip lines:")
	}
	root.setGoLineMode(ctx)
	if root.inputPrompt() != "Goto line:" {
		t.Errorf("Root.inputMode() = %v, want %v", root.inputPrompt(), "Goto line:")
	}
	root.setMultiColorMode(ctx)
	if root.inputPrompt() != "Multi color:" {
		t.Errorf("Root.inputMode() = %v, want %v", root.inputPrompt(), "Multi color:")
	}
	root.setForwardSearchMode(ctx)
	if root.inputPrompt() != "/" {
		t.Errorf("Root.inputMode() = %v, want %v", root.inputPrompt(), "/")
	}
	root.setBackSearchMode(ctx)
	if root.inputPrompt() != "?" {
		t.Errorf("Root.inputMode() = %v, want %v", root.inputPrompt(), "?")
	}
	root.setSearchFilterMode(ctx)
	if root.inputPrompt() != "&" {
		t.Errorf("Root.inputMode() = %v, want %v", root.inputPrompt(), "&")
	}
	root.setJumpTargetMode(ctx)
	if root.inputPrompt() != "Jump Target line:" {
		t.Errorf("Root.inputMode() = %v, want %v", root.inputPrompt(), "Jump Target line:")
	}
	root.setSaveBufferMode(ctx)
	if root.inputPrompt() != "(Save)file:" {
		t.Errorf("Root.inputMode() = %v, want %v", root.inputPrompt(), "(Save)file:")
	}
	root.setViewInputMode(ctx)
	if root.inputPrompt() != "Mode:" {
		t.Errorf("Root.inputMode() = %v, want %v", root.inputPrompt(), "Mode:")
	}
	root.setWatchIntervalMode(ctx)
	if root.inputPrompt() != "Watch interval:" {
		t.Errorf("Root.inputMode() = %v, want %v", root.inputPrompt(), "Watch interval:")
	}
	root.setWriteBAMode(ctx)
	if root.inputPrompt() != "WriteAndQuit Before:After:" {
		t.Errorf("Root.inputMode() = %v, want %v", root.inputPrompt(), "WriteAndQuit Before:After:")
	}
	root.setTabWidthMode(ctx)
	if root.inputPrompt() != "TAB width:" {
		t.Errorf("Root.inputMode() = %v, want %v", root.inputPrompt(), "TAB width:")
	}
	root.setSectionDelimiterMode(ctx)
	if root.inputPrompt() != "Section delimiter:" {
		t.Errorf("Root.inputMode() = %v, want %v", root.inputPrompt(), "Section delimiter:")
	}
	root.setSectionStartMode(ctx)
	if root.inputPrompt() != "Section start:" {
		t.Errorf("Root.inputMode() = %v, want %v", root.inputPrompt(), "Section start:")
	}
	root.setSectionNumMode(ctx)
	if root.inputPrompt() != "Section Num:" {
		t.Errorf("Root.inputMode() = %v, want %v", root.inputPrompt(), "Section Num:")
	}
	root.setSectionStartMode(ctx)
	if root.inputPrompt() != "Section start:" {
		t.Errorf("Root.inputMode() = %v, want %v", root.inputPrompt(), "Section start:")
	}
}

func TestRoot_inputState(t *testing.T) {
	root := rootHelper(t)
	ctx := context.Background()
	root.setForwardSearchMode(ctx)
	root.setSearcher("test", false)
	root.inputCaseSensitive(ctx)
	if root.searchOpt != "(Aa)" {
		t.Errorf("Root.inputState() = %v, want %v", root.searchOpt, "(Aa)")
	}
	root.inputSmartCaseSensitive(ctx)
	if root.searchOpt != "(S)" {
		t.Errorf("Root.inputState() = %v, want %v", root.searchOpt, "(S)")
	}
	root.inputIncSearch(ctx)
	if root.searchOpt != "(I)(S)" {
		t.Errorf("Root.inputState() = %v, want %v", root.searchOpt, "(I)(S)")
	}
	root.inputRegexpSearch(ctx)
	if root.searchOpt != "(R)(I)(S)" {
		t.Errorf("Root.inputState() = %v, want %v", root.searchOpt, "(R)(I)(S)")
	}
	if root.searchOpt != "(R)(I)(S)" {
		t.Errorf("Root.inputState() = %v, want %v", root.searchOpt, "(R)(I)(S)")
	}
	root.inputNonMatch(ctx)
	root.setBackSearchMode(ctx)
	if root.searchOpt != "(R)(I)(S)" {
		t.Errorf("Root.inputState() = %v, want %v", root.searchOpt, "(R)(I)(S)")
	}
	root.setSearchFilterMode(ctx)
	root.inputNonMatch(ctx)
	if root.searchOpt != "Non-match(R)(S)" {
		t.Errorf("Root.inputState() = %v, want %v", root.searchOpt, "Non-match(R)(S)")
	}
	root.inputPrevious(ctx)
	if root.searchOpt != "Non-match(R)(S)" {
		t.Errorf("Root.inputState() = %v, want %v", root.searchOpt, "Non-match(R)(S)")
	}
	root.inputNext(ctx)
	if root.searchOpt != "Non-match(R)(S)" {
		t.Errorf("Root.inputState() = %v, want %v", root.searchOpt, "Non-match(R)(S)")
	}
	root.setHeaderMode(ctx)
	root.setPromptOpt()
	if root.searchOpt != "" {
		t.Errorf("Root.inputState() = %v, want %v", root.searchOpt, "")
	}
}
