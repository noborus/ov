package oviewer

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/gdamore/tcell/v3"
)

func TestRoot_Mark(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	root := rootFileReadHelper(t, filepath.Join(testdata, "test3.txt"))
	t.Run("TestMark", func(t *testing.T) {
		root.prepareScreen()
		ctx := context.Background()
		root.Doc.topLN = 1
		root.draw(ctx)
		root.addMark(ctx)
		if !reflect.DeepEqual(root.Doc.marked, MatchedLineList([]MatchedLine{{lineNum: 1, line: []byte(root.Doc.getLineC(1).str)}})) {
			t.Errorf("addMark() = %#v, want %#v", root.Doc.marked, MatchedLineList([]MatchedLine{{lineNum: 1, line: []byte(root.Doc.getLineC(1).str)}}))
		}
		root.Doc.topLN = 10
		root.draw(ctx)
		root.addMark(ctx)
		if !reflect.DeepEqual(root.Doc.marked, MatchedLineList([]MatchedLine{{lineNum: 1, line: []byte(root.Doc.getLineC(1).str)}, {lineNum: 10, line: []byte(root.Doc.getLineC(10).str)}})) {
			t.Errorf("addMark() = %#v, want %#v", root.Doc.marked, MatchedLineList([]MatchedLine{{lineNum: 1, line: []byte(root.Doc.getLineC(1).str)}, {lineNum: 10, line: []byte(root.Doc.getLineC(10).str)}}))
		}
		root.Doc.topLN = 1
		root.draw(ctx)
		root.removeMark(ctx)
		if !reflect.DeepEqual(root.Doc.marked, MatchedLineList([]MatchedLine{{lineNum: 10, line: []byte(root.Doc.getLineC(10).str)}})) {
			t.Errorf("removeAllMark() = %#v, want %#v", root.Doc.marked, MatchedLineList([]MatchedLine{{lineNum: 10, line: []byte(root.Doc.getLineC(10).str)}}))
		}
		root.removeMark(ctx)
		root.Doc.topLN = 2
		root.draw(ctx)
		root.addMark(ctx)
		if !reflect.DeepEqual(root.Doc.marked, MatchedLineList([]MatchedLine{{lineNum: 10, line: []byte(root.Doc.getLineC(10).str)}, {lineNum: 2, line: []byte(root.Doc.getLineC(2).str)}})) {
			t.Errorf("addMark() = %#v, want %#v", root.Doc.marked, MatchedLineList([]MatchedLine{{lineNum: 10, line: []byte(root.Doc.getLineC(10).str)}, {lineNum: 2, line: []byte(root.Doc.getLineC(2).str)}}))
		}
		root.removeAllMark(ctx)
		if !reflect.DeepEqual(root.Doc.marked, MatchedLineList(nil)) {
			t.Errorf("removeAllMark() = %#v, want %#v", root.Doc.marked, MatchedLineList(nil))
		}
	})

	t.Run("TestAddMarksNoDuplicate", func(t *testing.T) {
		root.prepareScreen()
		ctx := context.Background()
		root.Doc.topLN = 1
		root.draw(ctx)

		root.Doc.marked = MatchedLineList{
			{lineNum: 1, line: []byte(root.Doc.getLineC(1).str)},
		}
		root.addMarks(ctx, MatchedLineList{
			{lineNum: 1, line: []byte(root.Doc.getLineC(1).str)}, // duplicate of existing
			{lineNum: 2, line: []byte(root.Doc.getLineC(2).str)},
			{lineNum: 2, line: []byte(root.Doc.getLineC(2).str)}, // duplicate in input
		})
		want := MatchedLineList{
			{lineNum: 1, line: []byte(root.Doc.getLineC(1).str)},
			{lineNum: 2, line: []byte(root.Doc.getLineC(2).str)},
		}
		if !reflect.DeepEqual(root.Doc.marked, want) {
			t.Errorf("addMarks() = %#v, want %#v", root.Doc.marked, want)
		}
	})
}

func TestRoot_nextMark(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	root := rootFileReadHelper(t, filepath.Join(testdata, "test3.txt"))
	root.prepareScreen()
	tests := []struct {
		name        string
		markedPoint int
		wantLine    int
	}{
		{
			name:        "testMarkNext1",
			markedPoint: 0,
			wantLine:    3,
		},
		{
			name:        "testMarkNext2",
			markedPoint: 1,
			wantLine:    5,
		},
		{
			name:        "testMarkNext3",
			markedPoint: 2,
			wantLine:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root.nextMark(context.Background()) // no marked
			root.Doc.marked = MatchedLineList{
				MatchedLine{lineNum: 1, line: []byte("a")},
				MatchedLine{lineNum: 3, line: []byte("b")},
				MatchedLine{lineNum: 5, line: []byte("c")},
			}
			root.Doc.markedPoint = tt.markedPoint
			root.nextMark(context.Background())
			if root.Doc.topLN != tt.wantLine {
				t.Errorf("got line %d, want line %d", root.Doc.topLN, tt.wantLine)
			}
		})
	}
}

func TestRoot_prevMark(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	root := rootFileReadHelper(t, filepath.Join(testdata, "test3.txt"))
	root.prepareScreen()
	tests := []struct {
		name        string
		markedPoint int
		wantLine    int
	}{
		{
			name:        "testMarkPrev1",
			markedPoint: 2,
			wantLine:    3,
		},
		{
			name:        "testMarkPrev2",
			markedPoint: 1,
			wantLine:    1,
		},
		{
			name:        "testMarkPrev3",
			markedPoint: 0,
			wantLine:    5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root.prevMark(context.Background()) // no marked
			root.Doc.marked = MatchedLineList{
				MatchedLine{lineNum: 1, line: []byte("a")},
				MatchedLine{lineNum: 3, line: []byte("b")},
				MatchedLine{lineNum: 5, line: []byte("c")},
			}
			root.Doc.markedPoint = tt.markedPoint
			root.prevMark(context.Background())
			if root.Doc.topLN != tt.wantLine {
				t.Errorf("got line %d, want line %d", root.Doc.topLN, tt.wantLine)
			}
		})
	}
}

func TestRoot_goMarkNumber(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	root := rootHelper(t)
	root.Doc.marked = MatchedLineList{
		{lineNum: 10, line: []byte("a")},
		{lineNum: 20, line: []byte("b")},
		{lineNum: 30, line: []byte("c")},
	}
	root.Doc.markedPoint = 1 // Start at index 1

	tests := []struct {
		name        string
		input       string
		startPoint  int
		wantPoint   int
		wantLineNum int
	}{
		{
			name:        "absolute index",
			input:       "2",
			startPoint:  0,
			wantPoint:   2,
			wantLineNum: 30,
		},
		{
			name:        "relative positive index",
			input:       "+1",
			startPoint:  1,
			wantPoint:   2,
			wantLineNum: 30,
		},
		{
			name:        "relative negative index",
			input:       "-1",
			startPoint:  2,
			wantPoint:   1,
			wantLineNum: 20,
		},
		{
			name:        "out of range high",
			input:       "10",
			startPoint:  0,
			wantPoint:   2,
			wantLineNum: 30,
		},
		{
			name:        "out of range low",
			input:       "-10",
			startPoint:  2,
			wantPoint:   0,
			wantLineNum: 10,
		},
		{
			name:        "invalid input",
			input:       "abc",
			startPoint:  1,
			wantPoint:   1,
			wantLineNum: 20,
		},
		{
			name:        "empty input",
			input:       "",
			startPoint:  1,
			wantPoint:   1,
			wantLineNum: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root.Doc.markedPoint = tt.startPoint
			root.goMarkNumber(tt.input)
			if root.Doc.markedPoint != tt.wantPoint {
				t.Errorf("goMarkNumber(%q) markedPoint = %d, want %d", tt.input, root.Doc.markedPoint, tt.wantPoint)
			}
			if len(root.Doc.marked) > 0 && root.Doc.markedPoint < len(root.Doc.marked) {
				if root.Doc.marked[root.Doc.markedPoint].lineNum != tt.wantLineNum {
					t.Errorf("goMarkNumber(%q) lineNum = %d, want %d", tt.input, root.Doc.marked[root.Doc.markedPoint].lineNum, tt.wantLineNum)
				}
			}
		})
	}
}

func Test_calcMarkIndex(t *testing.T) {
	type args struct {
		input   string
		current int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "positive absolute index",
			args: args{
				input:   "5",
				current: 2,
			},
			want: 5,
		},
		{
			name: "negative relative index",
			args: args{
				input:   "-2",
				current: 5,
			},
			want: 3,
		},
		{
			name: "positive relative index",
			args: args{
				input:   "+3",
				current: 4,
			},
			want: 7,
		},
		{
			name: "invalid input",
			args: args{
				input:   "abc",
				current: 10,
			},
			want: 10,
		},
		{
			name: "invalid relative input",
			args: args{
				input:   "+abc",
				current: 8,
			},
			want: 8,
		},
		{
			name: "zero index",
			args: args{
				input:   "0",
				current: 5,
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calcMarkIndex(tt.args.input, tt.args.current)
			if got != tt.want {
				t.Errorf("calcMarkIndex(%q, %d) = %d, want %d", tt.args.input, tt.args.current, got, tt.want)
			}
		})
	}
}

func TestRoot_markByPattern(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	root := rootFileReadHelper(t, filepath.Join(testdata, "test3.txt"))
	root.prepareScreen()
	ctx := context.Background()
	searcher := NewSearcher("9999", regexpCompile("9999", false), false, false)
	marks := root.Doc.allMatchedLines(ctx, searcher, 0)
	if len(marks) == 0 {
		t.Fatalf("allMatchedLines() returned no marks")
	}
	found := false
	for _, m := range marks {
		if m.lineNum == 9998 {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("allMatchedLines() does not contain expected lineNum 9998: %#v", marks)
	}
}
