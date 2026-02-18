package oviewer

import (
	"context"
	"fmt"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
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

func TestRoot_markByPattern_event(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	root := rootFileReadHelper(t, filepath.Join(testdata, "test3.txt"))
	root.prepareScreen()
	ctx := context.Background()
	pattern := "9999"
	root.markByPattern(ctx, pattern)

	deadline := time.After(1 * time.Second)
	for {
		ev := root.Screen.PollEvent()
		if addMarksEv, ok := ev.(*eventAddMarks); ok {
			if len(addMarksEv.marks) == 0 {
				t.Errorf("eventAddMarks.marks is empty, want at least 1")
			}
			found := false
			for _, m := range addMarksEv.marks {
				if m.lineNum == 9998 {
					found = true
					break
				}
			}
			if !found {
				fmt.Println("marks:", addMarksEv.marks)
				t.Errorf("eventAddMarks.marks does not contain expected lineNum 1")
			}
			return
		}
		select {
		case <-deadline:
			t.Fatalf("eventAddMarks event not received")
		default:
		}
	}
}
