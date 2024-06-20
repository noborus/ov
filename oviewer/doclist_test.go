package oviewer

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestRoot_DocumentList(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		fileNames []string
	}
	tests := []struct {
		name     string
		fields   fields
		wantLen  int
		wantLen2 int
	}{
		{
			name: "DocumentLen",
			fields: fields{
				fileNames: []string{filepath.Join(testdata, "test.txt")},
			},
			wantLen:  1,
			wantLen2: 1,
		},
		{
			name: "DocumentLen2",
			fields: fields{
				fileNames: []string{
					filepath.Join(testdata, "test.txt"),
					filepath.Join(testdata, "test2.txt"),
				},
			},
			wantLen:  2,
			wantLen2: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := Open(tt.fields.fileNames...)
			if err != nil {
				t.Fatalf("NewOviewer error = %v", err)
			}
			for !root.Doc.BufEOF() {
			}
			if got := root.DocumentLen(); got != tt.wantLen {
				t.Errorf("Root.DocumentLen() = %v, want %v", got, tt.wantLen)
			}
			if got := root.getDocument(0).FileName; got != tt.fields.fileNames[0] {
				t.Errorf("Root.getDocument() = %v, want %v", got, tt.fields.fileNames[0])
			}
			ctx := context.Background()
			root.closeDocument(ctx)
			if got := root.DocumentLen(); got != tt.wantLen2 {
				t.Errorf("Root.DocumentLen() = %v, want %v", got, tt.wantLen2)
			}
		})
	}
}

func TestRoot_SwitchingDoc(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	fileName1 := filepath.Join(testdata, "test.txt")
	fileName2 := filepath.Join(testdata, "test2.txt")
	root, err := Open(fileName1, fileName2)
	if err != nil {
		t.Fatalf("NewOviewer error = %v", err)
	}
	ctx := context.Background()
	keyBind, err := root.setKeyConfig(ctx)
	if err != nil {
		t.Fatal(err)
	}
	help, err := NewHelp(keyBind)
	if err != nil {
		t.Fatal(err)
	}
	root.helpDoc = help
	for !root.Doc.BufEOF() {
	}
	logDoc, err := NewLogDoc()
	if err != nil {
		t.Fatal(err)
	}
	root.logDoc = logDoc

	root.nextDoc(ctx)
	if root.Doc.FileName != fileName2 {
		t.Errorf("Root.nextDoc() = %v, want %v", root.Doc.FileName, fileName2)
	}
	root.previousDoc(ctx)
	if root.Doc.FileName != fileName1 {
		t.Errorf("Root.previousDoc() = %v, want %v", root.Doc.FileName, fileName1)
	}
	root.switchDocument(ctx, 2)
	if root.Doc.FileName != fileName2 {
		t.Errorf("Root.switchDocument() = %v, want %v", root.Doc.FileName, fileName2)
	}
	root.helpDisplay(ctx)
	if root.Doc.Caption != "Help" {
		t.Errorf("Root.helpDisplay() = %v, want %v", root.Doc.Caption, "Help")
	}
	root.helpDisplay(ctx)
	root.logDisplay(ctx)
	if root.Doc.Caption != "Log" {
		t.Errorf("Root.logDisplay() = %v, want %v", root.Doc.Caption, "Log")
	}
	root.logDisplay(ctx)
	root.toNormal(ctx)
	if root.Doc.FileName != fileName2 {
		t.Errorf("Root.switchDocument() = %v, want %v", root.Doc.FileName, fileName2)
	}
}
