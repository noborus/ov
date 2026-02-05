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
		if !reflect.DeepEqual(root.Doc.marked, MachedLineList([]MachedLine{{lineNum: 1, contents: root.Doc.getLineC(1).lc}})) {
			t.Errorf("addMark() = %#v, want %#v", root.Doc.marked, MachedLineList([]MachedLine{{lineNum: 1, contents: root.Doc.getLineC(1).lc}}))
		}
		root.Doc.topLN = 10
		root.draw(ctx)
		root.addMark(ctx)
		if !reflect.DeepEqual(root.Doc.marked, MachedLineList([]MachedLine{{lineNum: 1, contents: root.Doc.getLineC(1).lc}, {lineNum: 10, contents: root.Doc.getLineC(10).lc}})) {
			t.Errorf("addMark() = %#v, want %#v", root.Doc.marked, MachedLineList([]MachedLine{{lineNum: 1, contents: root.Doc.getLineC(1).lc}, {lineNum: 10, contents: root.Doc.getLineC(10).lc}}))
		}
		root.Doc.topLN = 1
		root.draw(ctx)
		root.removeMark(ctx)
		if !reflect.DeepEqual(root.Doc.marked, MachedLineList([]MachedLine{{lineNum: 10, contents: root.Doc.getLineC(10).lc}})) {
			t.Errorf("removeAllMark() = %#v, want %#v", root.Doc.marked, MachedLineList([]MachedLine{{lineNum: 10, contents: root.Doc.getLineC(10).lc}}))
		}
		root.removeMark(ctx)
		root.Doc.topLN = 2
		root.draw(ctx)
		root.addMark(ctx)
		if !reflect.DeepEqual(root.Doc.marked, MachedLineList([]MachedLine{{lineNum: 10, contents: root.Doc.getLineC(10).lc}, {lineNum: 2, contents: root.Doc.getLineC(2).lc}})) {
			t.Errorf("addMark() = %#v, want %#v", root.Doc.marked, MachedLineList([]MachedLine{{lineNum: 10, contents: root.Doc.getLineC(10).lc}, {lineNum: 2, contents: root.Doc.getLineC(2).lc}}))
		}
		root.removeAllMark(ctx)
		if !reflect.DeepEqual(root.Doc.marked, MachedLineList(nil)) {
			t.Errorf("removeAllMark() = %#v, want %#v", root.Doc.marked, MachedLineList(nil))
		}
	})

	t.Run("TestAddMarksNoDuplicate", func(t *testing.T) {
		root.prepareScreen()
		ctx := context.Background()
		root.Doc.topLN = 1
		root.draw(ctx)

		root.Doc.marked = MachedLineList{
			{lineNum: 1, contents: root.Doc.getLineC(1).lc},
		}
		root.addMarks(ctx, MachedLineList{
			{lineNum: 1, contents: root.Doc.getLineC(1).lc}, // duplicate of existing
			{lineNum: 2, contents: root.Doc.getLineC(2).lc},
			{lineNum: 2, contents: root.Doc.getLineC(2).lc}, // duplicate in input
		})
		want := MachedLineList{
			{lineNum: 1, contents: root.Doc.getLineC(1).lc},
			{lineNum: 2, contents: root.Doc.getLineC(2).lc},
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
			root.Doc.marked = MachedLineList{
				MachedLine{lineNum: 1, contents: StrToContents("a", 0)},
				MachedLine{lineNum: 3, contents: StrToContents("b", 0)},
				MachedLine{lineNum: 5, contents: StrToContents("c", 0)},
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
			root.Doc.marked = MachedLineList{
				MachedLine{lineNum: 1, contents: StrToContents("a", 0)},
				MachedLine{lineNum: 3, contents: StrToContents("b", 0)},
				MachedLine{lineNum: 5, contents: StrToContents("c", 0)},
			}
			root.Doc.markedPoint = tt.markedPoint
			root.prevMark(context.Background())
			if root.Doc.topLN != tt.wantLine {
				t.Errorf("got line %d, want line %d", root.Doc.topLN, tt.wantLine)
			}
		})
	}
}
