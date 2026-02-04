package oviewer

import (
	"bytes"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v3"
)

func TestRoot_prepareSidebarItems(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()

	tests := []struct {
		name         string
		sidebarMode  SidebarMode
		sidebarWidth int
		setup        func(root *Root)
		wantItems    func(root *Root) []SidebarItem
	}{
		{
			name:         "sidebarWidth zero does nothing",
			sidebarMode:  SidebarModeMark,
			sidebarWidth: 0,
			setup:        func(root *Root) {},
			wantItems:    func(root *Root) []SidebarItem { return nil },
		},
		{
			name:         "SidebarModeMark sets SidebarItems",
			sidebarMode:  SidebarModeMark,
			sidebarWidth: 20,
			setup: func(root *Root) {
				root.Doc.marked = []Mark{
					{lineNum: 0, contents: StrToContents("First", 0)},
					{lineNum: 1, contents: StrToContents("Second", 0)},
				}
				root.Doc.markedPoint = 0
			},
			wantItems: func(root *Root) []SidebarItem {
				return []SidebarItem{
					{Contents: StrToContents("0 First         ", 0), IsCurrent: true},
					{Contents: StrToContents("1 Second        ", 0), IsCurrent: false},
				}
			},
		},
		{
			name:         "SidebarModeDocList sets SidebarItems",
			sidebarMode:  SidebarModeDocList,
			sidebarWidth: 20,
			setup: func(root *Root) {
				root.DocList = []*Document{
					{FileName: "file1.txt"},
					{FileName: "file2.txt"},
				}
				root.CurrentDoc = 1
			},
			wantItems: func(root *Root) []SidebarItem {
				length := 16 // sidebarWidth - 4
				displayName1 := StrToContents("file1.txt", 0)
				displayName2 := StrToContents("file2.txt", 0)
				if len(displayName1) < length {
					displayName1 = append(displayName1, StrToContents(strings.Repeat(" ", length-len(displayName1)), 0)...)
				}
				if len(displayName2) < length {
					displayName2 = append(displayName2, StrToContents(strings.Repeat(" ", length-len(displayName2)), 0)...)
				}
				return []SidebarItem{
					{Contents: StrToContents("file1.txt      ", 0), IsCurrent: false},
					{Contents: StrToContents("file2.txt      ", 0), IsCurrent: true},
				}
			},
		},
		{
			name:         "SidebarModeHelp sets SidebarItems",
			sidebarMode:  SidebarModeHelp,
			sidebarWidth: 20,
			setup:        func(root *Root) {},
			wantItems: func(root *Root) []SidebarItem {
				return []SidebarItem{
					{Contents: StrToContents("[Escape, q]", 0), IsCurrent: false},
					{Contents: StrToContents("  quit", 0), IsCurrent: false},
					{Contents: StrToContents("[ctrl+c]", 0), IsCurrent: false},
					{Contents: StrToContents("  cancel", 0), IsCurrent: false},
					{Contents: StrToContents("[Q]", 0), IsCurrent: false},
					{Contents: StrToContents("  output screen and quit", 0), IsCurrent: false},
					{Contents: StrToContents("[ctrl+q]", 0), IsCurrent: false},
					{Contents: StrToContents("  set output screen and quit", 0), IsCurrent: false},
					{Contents: StrToContents("[alt+shift+F8]", 0), IsCurrent: false},
					{Contents: StrToContents("  set output original screen and quit", 0), IsCurrent: false},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := NewRoot(bytes.NewBufferString("First\nSecond\nThird"))
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}
			root.sidebarMode = tt.sidebarMode
			root.sidebarWidth = tt.sidebarWidth
			tt.setup(root)
			root.scr.vHeight = 10 // arbitrary value for testing
			root.prepareSidebarItems()
			want := tt.wantItems(root)
			got := root.SidebarItems
			if len(got) != len(want) {
				t.Errorf("SidebarItems length = %d, want %d", len(got), len(want))
			}
			for i := range got {
				if !reflect.DeepEqual(got[i].Contents, want[i].Contents) {
					t.Errorf("SidebarItems[%d].Contents = %q, want %q", i, got[i].Contents.String(), want[i].Contents.String())
				}
				if got[i].IsCurrent != want[i].IsCurrent {
					t.Errorf("SidebarItems[%d].IsCurrent = %v, want %v", i, got[i].IsCurrent, want[i].IsCurrent)
				}
			}
		})
	}
}

func TestRoot_sidebarItemsForMark(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		r     io.Reader
		marks []int
		want  []SidebarItem
	}{
		{
			name:  "marks1",
			r:     bytes.NewBufferString("Line 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6\nLine 7\nLine 8\nLine 9\nLine 10\nLine 11\nLine 12\nLine 13\nLine 14\nLine 15\nLine 16\nLine 17\nLine 18\nLine 19\nLine 20"),
			marks: []int{0, 1},
			want: []SidebarItem{
				{Contents: StrToContents("0 Line 1        ", 0), IsCurrent: true},
				{Contents: StrToContents("1 Line 2        ", 0), IsCurrent: false},
			},
		},
		{
			name:  "single mark at start",
			r:     bytes.NewBufferString("First\nSecond\nThird"),
			marks: []int{0},
			want: []SidebarItem{
				{Contents: StrToContents("0 First         ", 0), IsCurrent: true},
			},
		},
		{
			name:  "single mark at end",
			r:     bytes.NewBufferString("Alpha\nBeta\nGamma"),
			marks: []int{2},
			want: []SidebarItem{
				{Contents: StrToContents("2 Gamma         ", 0), IsCurrent: true},
			},
		},
		{
			name:  "multiple marks, current is last",
			r:     bytes.NewBufferString("A\nB\nC\nD"),
			marks: []int{1, 2, 3},
			want: []SidebarItem{
				{Contents: StrToContents("1 B             ", 0), IsCurrent: true},
				{Contents: StrToContents("2 C             ", 0), IsCurrent: false},
				{Contents: StrToContents("3 D             ", 0), IsCurrent: false},
			},
		},
		{
			name:  "long line truncated",
			r:     bytes.NewBufferString("Short\nThis is a very long line that should be truncated"),
			marks: []int{1},
			want: []SidebarItem{
				{Contents: StrToContents("1 This is a very long line that should be truncated", 0), IsCurrent: true},
			},
		},
		{
			name:  "no marks",
			r:     bytes.NewBufferString("One\nTwo\nThree"),
			marks: []int{},
			want:  []SidebarItem{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tcellNewScreen = fakeScreen
			defer func() {
				tcellNewScreen = tcell.NewScreen
			}()
			root, err := NewRoot(tt.r)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}
			root.scr.vHeight = 20 // arbitrary value for testing
			root.sidebarWidth = 20
			root.sidebarMode = SidebarModeMark
			for _, c := range tt.marks {
				root.Doc.topLN = c
				root.Doc.marked = append(root.Doc.marked, Mark{lineNum: c, contents: root.lineContent(c).lc})
			}
			got := root.sidebarItemsForMark()
			if len(got) != len(tt.want) {
				t.Errorf("sidebarItemsForMark() length = %d, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if !reflect.DeepEqual(got[i].Contents, tt.want[i].Contents) {
					t.Errorf("  got[%d]: %#v\n want[%d]: %#v", i, got[i].Contents.String(), i, tt.want[i].Contents.String())
				}
				if got[i].IsCurrent != tt.want[i].IsCurrent {
					t.Errorf("  got[%d].IsCurrent: %#v\n want[%d].IsCurrent: %#v", i, got[i].IsCurrent, i, tt.want[i].IsCurrent)
				}
			}
		})
	}
}

func TestRoot_sidebarItemsForHelp(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "basic help items",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootHelper(t)
			root.scr.vHeight = 20
			root.sidebarWidth = 20
			root.sidebarMode = SidebarModeHelp
			if len(root.sidebarItemsForHelp()) == 0 {
				t.Errorf("sidebarItemsForHelp() returned no items")
			}
		})
	}
}
