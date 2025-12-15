package oviewer

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/gdamore/tcell/v3"
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
			root := rootFileReadHelper(t, tt.fields.fileNames...)
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
	root := rootFileReadHelper(t, fileName1, fileName2)
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
	root.closeDocument(ctx)
	if root.Doc.FileName != fileName1 {
		t.Errorf("Root.switchDocument() = %v, want %v", root.Doc.FileName, fileName1)
	}
}

func TestRoot_addDocument(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		fileNames []string
	}
	type args struct {
		fileName string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			name: "addDocument",
			fields: fields{
				fileNames: []string{filepath.Join(testdata, "test.txt")},
			},
			args: args{
				fileName: filepath.Join(testdata, "test2.txt"),
			},
			want: 2,
		},
	}
	{
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				root := rootFileReadHelper(t, tt.fields.fileNames...)
				ctx := context.Background()
				addDoc, err := OpenDocument(tt.args.fileName)
				if err != nil {
					t.Fatal(err)
				}
				root.addDocument(ctx, addDoc)
				if got := root.DocumentLen(); got != tt.want {
					t.Errorf("Root.addDocument() = %v, want %v", got, tt.want)
				}
				root.closeDocument(ctx)
				if got := root.DocumentLen(); got != tt.want-1 {
					t.Errorf("Root.closeDocument() = %v, want %v", got, tt.want-1)
				}
			})
		}
	}
}

func TestRoot_insertDocument(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		fileNames []string
	}
	type args struct {
		num int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			name: "insertDocument",
			fields: fields{
				fileNames: []string{filepath.Join(testdata, "test.txt")},
			},
			args: args{
				num: 0,
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootFileReadHelper(t, tt.fields.fileNames...)
			ctx := context.Background()
			m, err := NewDocument()
			if err != nil {
				t.Fatal(err)
			}
			m.documentType = DocFilter
			root.insertDocument(ctx, tt.args.num, m)
			if got := root.DocumentLen(); got != tt.want {
				t.Errorf("Root.insertDocument() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoot_closeAllDocumentsOfType(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		fileNames []string
		count     int
	}
	tests := []struct {
		name   string
		fields fields
		want   int
		want1  []string
	}{
		{
			name: "closeAllDocumentsOfTypeNon",
			fields: fields{
				fileNames: []string{filepath.Join(testdata, "test.txt")},
				count:     0,
			},
			want:  0,
			want1: []string{},
		},
		{
			name: "closeAllDocumentsOfType1",
			fields: fields{
				fileNames: []string{filepath.Join(testdata, "test.txt")},
				count:     1,
			},
			want:  0,
			want1: []string{"filter:test"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootFileReadHelper(t, tt.fields.fileNames...)
			searcher := NewSearcher("test", nil, false, false)
			ctx := context.Background()
			for range tt.fields.count {
				root.filterDocument(ctx, searcher)
			}
			got, got1 := root.closeAllDocumentsOfType(DocFilter)
			if got != tt.want {
				t.Errorf("Root.closeAllDocumentsOfType() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Root.closeAllDocumentsOfType() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
