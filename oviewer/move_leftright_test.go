package oviewer

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestDocument_optimalCursor(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		fileName        string
		wrap            bool
		columnDelimiter string
		x               int
		width           int
		HeaderColumn    int
		verticalHeader  int
	}
	type args struct {
		cursor int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			name: "noDelimiter",
			fields: fields{
				fileName: filepath.Join(testdata, "normal.txt"),
				wrap:     true,
			},
			args: args{
				cursor: 10,
			},
			want: 10,
		},
		{
			name: "CSVRight",
			fields: fields{
				fileName:        filepath.Join(testdata, "MOCK_DATA.csv"),
				columnDelimiter: ",",
				x:               10,
				width:           40,
			},
			args: args{
				cursor: 4,
			},
			want: 4,
		},
		{
			name: "CSVLeft",
			fields: fields{
				fileName:        filepath.Join(testdata, "MOCK_DATA.csv"),
				columnDelimiter: ",",
				x:               30,
				width:           10,
			},
			args: args{
				cursor: 0,
			},
			want: 5,
		},
		{
			name: "OverColumn",
			fields: fields{
				fileName:        filepath.Join(testdata, "MOCK_DATA.csv"),
				columnDelimiter: ",",
				x:               10,
				width:           40,
			},
			args: args{
				cursor: 20,
			},
			want: 7,
		},
		{
			name: "VerticalHeader1",
			fields: fields{
				fileName:        filepath.Join(testdata, "MOCK_DATA.csv"),
				columnDelimiter: ",",
				x:               10,
				width:           40,
				verticalHeader:  5,
			},
			args: args{
				cursor: 0,
			},
			want: 3,
		},
		{
			name: "VerticalHeader2",
			fields: fields{
				fileName:        filepath.Join(testdata, "MOCK_DATA.csv"),
				columnDelimiter: ",",
				x:               10,
				width:           40,
				verticalHeader:  20,
			},
			args: args{
				cursor: 0,
			},
			want: 5,
		},
		{
			name: "HeaderColumn1",
			fields: fields{
				fileName:        filepath.Join(testdata, "MOCK_DATA.csv"),
				columnDelimiter: ",",
				HeaderColumn:    1,
				x:               10,
				width:           40,
			},
			args: args{
				cursor: 3,
			},
			want: 3,
		},
		{
			name: "HeaderColumn2",
			fields: fields{
				fileName:        filepath.Join(testdata, "MOCK_DATA.csv"),
				columnDelimiter: ",",
				HeaderColumn:    4,
				x:               10,
				width:           30,
			},
			args: args{
				cursor: 6,
			},
			want: 4,
		},
		{
			name: "HeaderColumn3",
			fields: fields{
				fileName:        filepath.Join(testdata, "MOCK_DATA.csv"),
				columnDelimiter: ",",
				HeaderColumn:    4,
				x:               10,
				width:           40,
			},
			args: args{
				cursor: 7,
			},
			want: 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootFileReadHelper(t, tt.fields.fileName)
			root.prepareScreen()
			m := root.Doc
			m.WrapMode = tt.fields.wrap
			m.ColumnMode = true
			m.ColumnDelimiter = tt.fields.columnDelimiter
			m.x = tt.fields.x
			m.width = tt.fields.width
			m.VerticalHeader = tt.fields.verticalHeader
			m.HeaderColumn = tt.fields.HeaderColumn
			ctx := context.Background()
			root.prepareDraw(ctx)
			if got := m.optimalCursor(root.scr, tt.args.cursor); got != tt.want {
				t.Errorf("Document.correctCursor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDocument_optimalX(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		fileName        string
		wrap            bool
		columnDelimiter string
		x               int
		width           int
		HeaderColumn    int
		verticalHeader  int
	}
	type args struct {
		cursor int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantX   int
		wantErr bool
	}{
		{
			name: "cursorAtZero",
			fields: fields{
				fileName: filepath.Join(testdata, "normal.txt"),
				width:    40,
			},
			args: args{
				cursor: 0,
			},
			wantX:   0,
			wantErr: false,
		},
		{
			name: "cursorLessThanHeaderColumn",
			fields: fields{
				fileName:        filepath.Join(testdata, "MOCK_DATA.csv"),
				columnDelimiter: ",",
				width:           40,
				HeaderColumn:    5,
			},
			args: args{
				cursor: 3,
			},
			wantX:   0,
			wantErr: false,
		},
		{
			name: "cursorGreaterThanColumns",
			fields: fields{
				fileName:        filepath.Join(testdata, "MOCK_DATA.csv"),
				columnDelimiter: ",",
				width:           40,
				x:               10,
			},
			args: args{
				cursor: 20,
			},
			wantX:   10,
			wantErr: true,
		},
		{
			name: "cursorWithinScreen",
			fields: fields{
				fileName:        filepath.Join(testdata, "MOCK_DATA.csv"),
				columnDelimiter: ",",
				width:           40,
				x:               10,
			},
			args: args{
				cursor: 4,
			},
			wantX:   10,
			wantErr: false,
		},
		{
			name: "cursorOffScreen",
			fields: fields{
				fileName:        filepath.Join(testdata, "MOCK_DATA.csv"),
				columnDelimiter: ",",
				width:           40,
				x:               10,
			},
			args: args{
				cursor: 6,
			},
			wantX:   20,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootFileReadHelper(t, tt.fields.fileName)
			root.prepareScreen()
			m := root.Doc
			m.WrapMode = tt.fields.wrap
			m.ColumnMode = true
			m.ColumnDelimiter = tt.fields.columnDelimiter
			m.x = tt.fields.x
			m.width = tt.fields.width
			m.VerticalHeader = tt.fields.verticalHeader
			m.HeaderColumn = tt.fields.HeaderColumn
			ctx := context.Background()
			root.prepareDraw(ctx)
			gotX, err := m.optimalX(root.scr, tt.args.cursor)
			if (err != nil) != tt.wantErr {
				t.Errorf("Document.optimalX() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotX != tt.wantX {
				t.Errorf("Document.optimalX() = %v, want %v", gotX, tt.wantX)
			}
		})
	}
}

func TestDocument_moveColumnLeft(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		fileName        string
		wrap            bool
		columnDelimiter string
		x               int
		width           int
		HeaderColumn    int
		verticalHeader  int
		columnCursor    int
	}
	type args struct {
		n     int
		cycle bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		wantX   int
	}{
		{
			name: "moveLeftWithinBounds",
			fields: fields{
				fileName:        filepath.Join(testdata, "MOCK_DATA.csv"),
				wrap:            false,
				columnDelimiter: ",",
				x:               20,
				width:           40,
				columnCursor:    3,
			},
			args: args{
				n:     1,
				cycle: false,
			},
			wantErr: false,
			wantX:   0,
		},
		{
			name: "moveLeftOutOfBoundsNoCycle",
			fields: fields{
				fileName:        filepath.Join(testdata, "MOCK_DATA.csv"),
				wrap:            false,
				columnDelimiter: ",",
				x:               0,
				width:           40,
				columnCursor:    0,
			},
			args: args{
				n:     1,
				cycle: false,
			},
			wantErr: true,
			wantX:   0,
		},
		{
			name: "moveLeftOutOfBoundsWithCycle",
			fields: fields{
				fileName:        filepath.Join(testdata, "MOCK_DATA.csv"),
				wrap:            false,
				columnDelimiter: ",",
				x:               20,
				width:           40,
				columnCursor:    0,
			},
			args: args{
				n:     1,
				cycle: true,
			},
			wantErr: false,
			wantX:   0,
		},
		{
			name: "moveLeftToHeaderColumn",
			fields: fields{
				fileName:        filepath.Join(testdata, "MOCK_DATA.csv"),
				wrap:            false,
				columnDelimiter: ",",
				x:               30,
				width:           40,
				columnCursor:    5,
				HeaderColumn:    4,
			},
			args: args{
				n:     1,
				cycle: false,
			},
			wantErr: false,
			wantX:   25,
		},
		{
			name: "moveLeftWithWrapMode",
			fields: fields{
				fileName:        filepath.Join(testdata, "MOCK_DATA.csv"),
				wrap:            true,
				columnDelimiter: ",",
				x:               20,
				width:           40,
				columnCursor:    3,
			},
			args: args{
				n:     1,
				cycle: false,
			},
			wantErr: false,
			wantX:   0,
		},
		{
			name: "moveLeftColumn2",
			fields: fields{
				fileName:        filepath.Join(testdata, "column2.txt"),
				wrap:            false,
				columnDelimiter: "|",
				x:               80,
				width:           160,
				columnCursor:    2,
			},
			args: args{
				n:     1,
				cycle: false,
			},
			wantErr: false,
			wantX:   44,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootFileReadHelper(t, tt.fields.fileName)
			root.prepareScreen()
			m := root.Doc
			m.WrapMode = tt.fields.wrap
			m.ColumnMode = true
			m.ColumnDelimiter = tt.fields.columnDelimiter
			m.x = tt.fields.x
			m.width = tt.fields.width
			m.VerticalHeader = tt.fields.verticalHeader
			m.HeaderColumn = tt.fields.HeaderColumn
			m.columnCursor = tt.fields.columnCursor
			ctx := context.Background()
			root.prepareDraw(ctx)
			if err := m.moveColumnLeft(tt.args.n, root.scr, tt.args.cycle); (err != nil) != tt.wantErr {
				t.Errorf("Document.moveColumnLeft() error = %v, wantErr %v", err, tt.wantErr)
			}
			if m.x != tt.wantX {
				t.Errorf("Document.moveColumnLeft() = %v, want %v", m.x, tt.wantX)
			}
		})
	}
}

func TestDocument_moveColumnRight(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		fileName        string
		wrap            bool
		columnDelimiter string
		x               int
		width           int
		HeaderColumn    int
		verticalHeader  int
		columnCursor    int
	}
	type args struct {
		n     int
		cycle bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		wantX   int
	}{
		{
			name: "moveRightWithinBounds",
			fields: fields{
				fileName:        filepath.Join(testdata, "MOCK_DATA.csv"),
				wrap:            false,
				columnDelimiter: ",",
				x:               10,
				width:           40,
				columnCursor:    2,
			},
			args: args{
				n:     1,
				cycle: false,
			},
			wantErr: false,
			wantX:   10,
		},
		{
			name: "moveRightOutOfBoundsNoCycle",
			fields: fields{
				fileName:        filepath.Join(testdata, "MOCK_DATA.csv"),
				wrap:            false,
				columnDelimiter: ",",
				x:               39,
				width:           40,
				columnCursor:    8,
			},
			args: args{
				n:     1,
				cycle: false,
			},
			wantErr: true,
			wantX:   39,
		},
		{
			name: "moveRightOutOfBoundsWithCycle",
			fields: fields{
				fileName:        filepath.Join(testdata, "MOCK_DATA.csv"),
				wrap:            false,
				columnDelimiter: ",",
				x:               39,
				width:           40,
				columnCursor:    8,
			},
			args: args{
				n:     1,
				cycle: true,
			},
			wantErr: false,
			wantX:   0,
		},
		{
			name: "moveRightToEndOfLine",
			fields: fields{
				fileName:        filepath.Join(testdata, "MOCK_DATA.csv"),
				wrap:            false,
				columnDelimiter: ",",
				x:               7,
				width:           40,
				columnCursor:    7,
			},
			args: args{
				n:     1,
				cycle: false,
			},
			wantErr: false,
			wantX:   9,
		},
		{
			name: "moveRightWithWrapMode",
			fields: fields{
				fileName:        filepath.Join(testdata, "MOCK_DATA.csv"),
				wrap:            true,
				columnDelimiter: ",",
				x:               0,
				width:           40,
				columnCursor:    2,
			},
			args: args{
				n:     1,
				cycle: false,
			},
			wantErr: false,
			wantX:   0,
		},
		{
			name: "moveRightToHeaderColumn2",
			fields: fields{
				fileName:        filepath.Join(testdata, "column2.txt"),
				wrap:            false,
				columnDelimiter: "|",
				x:               0,
				width:           40,
				columnCursor:    0,
			},
			args: args{
				n:     1,
				cycle: false,
			},
			wantErr: false,
			wantX:   37,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootFileReadHelper(t, tt.fields.fileName)
			root.prepareScreen()
			m := root.Doc
			m.WrapMode = tt.fields.wrap
			m.ColumnMode = true
			m.ColumnDelimiter = tt.fields.columnDelimiter
			m.x = tt.fields.x
			m.width = tt.fields.width
			m.VerticalHeader = tt.fields.verticalHeader
			m.HeaderColumn = tt.fields.HeaderColumn
			m.columnCursor = tt.fields.columnCursor
			ctx := context.Background()
			root.prepareDraw(ctx)
			if err := m.moveColumnRight(tt.args.n, root.scr, tt.args.cycle); (err != nil) != tt.wantErr {
				t.Errorf("Document.moveColumnRight() error = %v, wantErr %v", err, tt.wantErr)
			}
			if m.x != tt.wantX {
				t.Errorf("Document.moveColumnRight() = %v, want %v", m.x, tt.wantX)
			}
		})
	}
}
