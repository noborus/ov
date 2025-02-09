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
			ctx := context.Background()
			root.prepareDraw(ctx)
			if got := m.optimalCursor(root.scr, tt.args.cursor); got != tt.want {
				t.Errorf("Document.correctCursor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoot_moveColumnWithLeft(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		cursor int
	}
	type args struct {
		n     int
		cycle bool
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantErr    bool
		wantCursor int
	}{
		{
			name: "ps1",
			fields: fields{
				cursor: 1,
			},
			args: args{
				n:     1,
				cycle: true,
			},
			wantErr:    false,
			wantCursor: 0,
		},
		{
			name: "ps2",
			fields: fields{
				cursor: 4,
			},
			args: args{
				n:     2,
				cycle: true,
			},
			wantErr:    false,
			wantCursor: 2,
		},
		{
			name: "psCycle",
			fields: fields{
				cursor: 0,
			},
			args: args{
				n:     1,
				cycle: true,
			},
			wantErr:    false,
			wantCursor: 10,
		},
		{
			name: "psNoCycle",
			fields: fields{
				cursor: 0,
			},
			args: args{
				n:     1,
				cycle: false,
			},
			wantErr:    true,
			wantCursor: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootFileReadHelper(t, filepath.Join(testdata, "ps.txt"))
			root.prepareScreen()
			m := root.Doc
			m.ColumnWidth = true
			m.ColumnMode = true
			m.columnCursor = tt.fields.cursor
			root.prepareDraw(context.Background())
			if err := m.moveColumnLeft(tt.args.n, root.scr, tt.args.cycle); (err != nil) != tt.wantErr {
				t.Errorf("Document.moveColumnWidthLeft() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got := m.columnCursor; got != tt.wantCursor {
				t.Errorf("Document.moveColumnWidthLeft() = %v, want %v", got, tt.wantCursor)
			}
		})
	}
}

func TestDocument_moveColumnWidthRight(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		cursor int
		x      int
	}
	type args struct {
		n     int
		cycle bool
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantErr    bool
		wantCursor int
	}{
		{
			name: "ps1",
			fields: fields{
				cursor: 1,
			},
			args: args{
				n:     1,
				cycle: true,
			},
			wantErr:    false,
			wantCursor: 2,
		},
		{
			name: "ps2",
			fields: fields{
				cursor: 4,
			},
			args: args{
				n:     2,
				cycle: true,
			},
			wantErr:    false,
			wantCursor: 6,
		},
		{
			name: "psCycle",
			fields: fields{
				cursor: 10,
				x:      180,
			},
			args: args{
				n:     1,
				cycle: true,
			},
			wantErr:    false,
			wantCursor: 0,
		},
		{
			name: "psNoCycle",
			fields: fields{
				cursor: 10,
				x:      180,
			},
			args: args{
				n:     1,
				cycle: false,
			},
			wantErr:    true,
			wantCursor: 10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootFileReadHelper(t, filepath.Join(testdata, "ps.txt"))
			root.prepareScreen()
			m := root.Doc
			m.ColumnMode = true
			m.ColumnWidth = true
			m.columnCursor = tt.fields.cursor
			m.setColumnWidths()
			m.x = tt.fields.x
			root.prepareDraw(context.Background())
			if err := m.moveColumnRight(tt.args.n, root.scr, tt.args.cycle); (err != nil) != tt.wantErr {
				t.Errorf("Document.moveColumnWidthRight() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got := m.columnCursor; got != tt.wantCursor {
				t.Errorf("Document.moveColumnWidthRight() = %v, want %v", got, tt.wantCursor)
			}
		})
	}
}

func TestDocument_moveColumnDelimiterLeft(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		cursor int
	}
	type args struct {
		n     int
		cycle bool
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantErr    bool
		wantCursor int
	}{
		{
			name: "csv1",
			fields: fields{
				cursor: 1,
			},
			args: args{
				n:     1,
				cycle: true,
			},
			wantErr:    false,
			wantCursor: 0,
		},
		{
			name: "csv2",
			fields: fields{
				cursor: 4,
			},
			args: args{
				n:     1,
				cycle: true,
			},
			wantErr:    false,
			wantCursor: 3,
		},
		{
			name: "csvCycle",
			fields: fields{
				cursor: 0,
			},
			args: args{
				n:     1,
				cycle: true,
			},
			wantErr:    false,
			wantCursor: 7,
		},
		{
			name: "csvNoCycle",
			fields: fields{
				cursor: 0,
			},
			args: args{
				n:     1,
				cycle: false,
			},
			wantErr:    true,
			wantCursor: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootFileReadHelper(t, filepath.Join(testdata, "MOCK_DATA.csv"))
			root.prepareScreen()
			m := root.Doc
			m.ColumnMode = true
			m.columnCursor = tt.fields.cursor
			m.setDelimiter(",")
			root.prepareDraw(context.Background())
			if err := m.moveColumnLeft(tt.args.n, root.scr, tt.args.cycle); (err != nil) != tt.wantErr {
				t.Errorf("Document.moveColumnDelimiterLeft() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got := m.columnCursor; got != tt.wantCursor {
				t.Errorf("Document.moveColumnDelimiterLeft() = %v, want %v", got, tt.wantCursor)
			}
		})
	}
}
