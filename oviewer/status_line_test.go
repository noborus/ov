package oviewer

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestRoot_inputPrompt(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		Eventer
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "test delimiter",
			fields: fields{
				Eventer: newDelimiterEvent(blankCandidate()),
			},
			want: "Delimiter:",
		},
		{
			name: "test go to line",
			fields: fields{
				Eventer: newGotoEvent(blankCandidate()),
			},
			want: "Goto line:",
		},
		{
			name: "test header",
			fields: fields{
				Eventer: newHeaderEvent(),
			},
			want: "Header length:",
		},
		{
			name: "test multi color",
			fields: fields{
				Eventer: newMultiColorEvent(blankCandidate()),
			},
			want: "multicolor:",
		},
		{
			name: "test save buffer",
			fields: fields{
				Eventer: newSaveBufferEvent(blankCandidate()),
			},
			want: "(Save)file:",
		},
		{
			name: "test section delimiter",
			fields: fields{
				Eventer: newSectionDelimiterEvent(blankCandidate()),
			},
			want: "Section delimiter:",
		},
		{
			name: "test section start",
			fields: fields{
				Eventer: newSectionStartEvent(blankCandidate()),
			},
			want: "Section start:",
		},
		{
			name: "test skip lines",
			fields: fields{
				Eventer: newSkipLinesEvent(),
			},
			want: "Skip lines:",
		},
		{
			name: "test tab width",
			fields: fields{
				Eventer: newTabWidthEvent(blankCandidate()),
			},
			want: "TAB width:",
		},
		{
			name: "test watch",
			fields: fields{
				Eventer: newWatchIntervalEvent(blankCandidate()),
			},
			want: "Watch interval:",
		},
		{
			name: "test write buffer",
			fields: fields{
				Eventer: newWriteBAEvent(blankCandidate()),
			},
			want: "WriteAndQuit Before:After:",
		},
		{
			name: "test jump target",
			fields: fields{
				Eventer: newJumpTargetEvent(blankCandidate()),
			},
			want: "Jump Target line:",
		},
		{
			name: "test view mode",
			fields: fields{
				Eventer: newViewModeEvent(blankCandidate()),
			},
			want: "Mode:",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := NewRoot(bytes.NewBufferString("test"))
			if err != nil {
				t.Fatal(err)
			}
			input := root.input
			input.Event = tt.fields.Eventer
			if got := root.inputPrompt(); got != tt.want {
				t.Errorf("Root.inputPrompt() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestRoot_inputPromptSearch(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		Eventer
		RegexpSearch       bool
		Incsearch          bool
		SmartCaseSensitive bool
		CaseSensitive      bool
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "test search",
			fields: fields{
				Eventer:            newSearchEvent(blankCandidate(), forward),
				RegexpSearch:       true,
				Incsearch:          true,
				SmartCaseSensitive: true,
				CaseSensitive:      false,
			},
			want: "(R)(I)(S)/",
		},
		{
			name: "test back search",
			fields: fields{
				Eventer:            newSearchEvent(blankCandidate(), backward),
				RegexpSearch:       true,
				Incsearch:          false,
				SmartCaseSensitive: true,
				CaseSensitive:      false,
			},
			want: "(R)(S)?",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := NewRoot(bytes.NewBufferString("test"))
			if err != nil {
				t.Fatal(err)
			}
			input := root.input
			input.Event = tt.fields.Eventer
			root.Config.RegexpSearch = tt.fields.RegexpSearch
			root.Config.Incsearch = tt.fields.Incsearch
			root.Config.SmartCaseSensitive = tt.fields.SmartCaseSensitive
			root.Config.CaseSensitive = tt.fields.CaseSensitive
			root.setPromptOpt()
			if got := root.inputPrompt(); got != tt.want {
				t.Errorf("Root.inputPrompt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoot_leftStatus(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		eventer   Eventer
		caption   string
		inputMode InputMode
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "test normal",
			fields: fields{
				caption:   "test",
				eventer:   normal(),
				inputMode: Normal,
			},
			want: "test:",
		},
		{
			name: "test search prompt",
			fields: fields{
				caption:   "test",
				eventer:   newSearchEvent(blankCandidate(), forward),
				inputMode: Search,
			},
			want: "/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := NewRoot(bytes.NewBufferString("test"))
			if err != nil {
				t.Fatal(err)
			}
			root.Doc.Caption = tt.fields.caption
			root.input.Event = tt.fields.eventer
			got, _ := root.leftStatus()
			gotStr, _ := ContentsToStr(got)
			if gotStr != tt.want {
				t.Errorf("Root.leftStatus() got = %v, want %v", gotStr, tt.want)
			}
		})
	}
}

func TestRoot_leftStatus2(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		fileNames []string
		eventer   Eventer
		inputMode InputMode
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "test normal",
			fields: fields{
				fileNames: []string{filepath.Join(testdata, "test.txt"), filepath.Join(testdata, "section2.txt")},
				eventer:   normal(),
				inputMode: Normal,
			},
			want: "[0]" + filepath.Join(testdata, "test.txt") + ":",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := Open(tt.fields.fileNames...)
			if err != nil {
				t.Fatal(err)
			}
			root.input.Event = tt.fields.eventer
			got, _ := root.leftStatus()
			gotStr, _ := ContentsToStr(got)
			if gotStr != tt.want {
				t.Errorf("Root.leftStatus() got = %v, want %v", gotStr, tt.want)
			}
		})
	}
}

func TestRoot_statusDisplay(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		FollowSection bool
		FollowAll     bool
		FollowMode    bool
		FollowName    bool
		WatchMode     bool
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "test watch",
			fields: fields{
				WatchMode: true,
			},
			want: "(Watch)",
		},
		{
			name: "test follow section",
			fields: fields{
				FollowSection: true,
			},
			want: "(Follow Section)",
		},
		{
			name: "test follow all",
			fields: fields{
				FollowAll: true,
			},
			want: "(Follow All)",
		},
		{
			name: "test follow name",
			fields: fields{
				FollowMode: true,
				FollowName: true,
			},
			want: "(Follow Name)",
		},
		{
			name: "test follow mode",
			fields: fields{
				FollowMode: true,
			},
			want: "(Follow Mode)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := NewRoot(bytes.NewBufferString("test"))
			if err != nil {
				t.Fatal(err)
			}
			root.Doc.WatchMode = tt.fields.WatchMode
			root.Doc.FollowSection = tt.fields.FollowSection
			root.General.FollowAll = tt.fields.FollowAll
			root.Doc.FollowName = tt.fields.FollowName
			root.Doc.FollowMode = tt.fields.FollowMode
			if got := root.statusDisplay(); got != tt.want {
				t.Errorf("Root.statusDisplay() = %v, want %v", got, tt.want)
			}
		})
	}
}
