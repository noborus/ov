package oviewer

import (
	"path/filepath"
	"reflect"
	"regexp"
	"testing"
)

func TestDocument_moveBeginLeft(t *testing.T) {
	tests := []struct {
		name  string
		wantX int
	}{
		{
			name:  "test1",
			wantX: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewDocument()
			if err != nil {
				t.Fatal(err)
			}
			m.moveBeginLeft()
			if m.x != tt.wantX {
				t.Errorf("screenAdjustX() gotX = %v, wantX %v", m.x, tt.wantX)
			}
		})
	}
}

func Test_screenAdjustX(t *testing.T) {
	type args struct {
		left   int
		right  int
		cl     int
		cr     int
		widths []int
		cursor int
	}
	tests := []struct {
		name       string
		args       args
		wantX      int
		wantCursor int
		wantErr    bool
	}{
		{
			name: "leftEdge",
			args: args{
				left:   0,
				right:  80,
				cl:     0,
				cr:     10,
				widths: []int{0, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
				cursor: 0,
			},
			wantX:      0,
			wantCursor: 0,
			wantErr:    false,
		},
		{
			name: "right1",
			args: args{
				left:   0,
				right:  80,
				cl:     10,
				cr:     20,
				widths: []int{0, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
				cursor: 1,
			},
			wantX:      0,
			wantCursor: 1,
			wantErr:    false,
		},
		{
			name: "rightScroll",
			args: args{
				left:   0,
				right:  80,
				cl:     80,
				cr:     90,
				widths: []int{0, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
				cursor: 8,
			},
			wantX:      12,
			wantCursor: 8,
			wantErr:    false,
		},
		{
			name: "rightEdgeNotMove",
			args: args{
				left:   130,
				right:  180,
				cl:     100,
				cr:     140,
				widths: []int{0, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
				cursor: 11,
			},
			wantX:      100,
			wantCursor: 10,
			wantErr:    true,
		},
		{
			name: "rightEdgeNotScroll",
			args: args{
				left:   80,
				right:  140,
				cl:     100,
				cr:     110,
				widths: []int{0, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
				cursor: 11,
			},
			wantX:      80,
			wantCursor: 10,
			wantErr:    true,
		},
		{
			name: "rightEdgeOver",
			args: args{
				left:   10,
				right:  90,
				cl:     100,
				cr:     140,
				widths: []int{0, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
				cursor: 11,
			},
			wantX:      62,
			wantCursor: 10,
			wantErr:    false,
		},
		{
			name: "left1",
			args: args{
				left:   60,
				right:  140,
				cl:     50,
				cr:     60,
				widths: []int{0, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
				cursor: 5,
			},
			wantX:      48,
			wantCursor: 5,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := screenAdjustX(tt.args.left, tt.args.right, tt.args.cl, tt.args.cr, tt.args.widths, tt.args.cursor)
			if (err != nil) != tt.wantErr {
				t.Errorf("screenAdjustX() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantX {
				t.Errorf("screenAdjustX() gotX = %v, wantX %v", got, tt.wantX)
			}
			if got1 != tt.wantCursor {
				t.Errorf("screenAdjustX() gotCursor = %v, wantCursor %v", got1, tt.wantCursor)
			}
		})
	}
}

func Test_screenAdjustX2(t *testing.T) {
	type args struct {
		left   int
		right  int
		cl     int
		cr     int
		widths []int
		cursor int
	}
	tests := []struct {
		name       string
		args       args
		wantX      int
		wantCursor int
		wantErr    bool
	}{
		{
			name: "leftOne",
			args: args{
				left:   20,
				right:  40,
				cl:     10,
				cr:     20,
				widths: []int{0, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
				cursor: 1,
			},
			wantX:      8,
			wantCursor: 1,
			wantErr:    false,
		},
		{
			name: "leftNoMove",
			args: args{
				left:   60,
				right:  80,
				cl:     40,
				cr:     50,
				widths: []int{0, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
				cursor: 4,
			},
			wantX:      40,
			wantCursor: 5,
			wantErr:    false,
		},
		{
			name: "leftMoveLargeColumn",
			args: args{
				left:   100,
				right:  140,
				cl:     10,
				cr:     100,
				widths: []int{0, 10, 100, 110},
				cursor: 1,
			},
			wantX:      60,
			wantCursor: 1,
			wantErr:    false,
		},
		{
			name: "RightOne",
			args: args{
				left:   0,
				right:  30,
				cl:     30,
				cr:     40,
				widths: []int{0, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
				cursor: 3,
			},
			wantX:      12,
			wantCursor: 3,
			wantErr:    false,
		},
		{
			name: "RightScroll",
			args: args{
				left:   0,
				right:  35,
				cl:     20,
				cr:     100,
				widths: []int{0, 10, 20, 100},
				cursor: 2,
			},
			wantX:      18,
			wantCursor: 2,
			wantErr:    false,
		},
		{
			name: "RightNotCursor",
			args: args{
				left:   0,
				right:  35,
				cl:     40,
				cr:     50,
				widths: []int{0, 10, 40, 50, 100},
				cursor: 2,
			},
			wantX:      33,
			wantCursor: 1,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotX, gotCursor, err := screenAdjustX(tt.args.left, tt.args.right, tt.args.cl, tt.args.cr, tt.args.widths, tt.args.cursor)
			if (err != nil) != tt.wantErr {
				t.Errorf("screenAdjustX() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotX != tt.wantX {
				t.Errorf("screenAdjustX() gotX = %v, wantX %v", gotX, tt.wantX)
			}
			if gotCursor != tt.wantCursor {
				t.Errorf("screenAdjustX() gotCursor = %v, wantCursor %v", gotCursor, tt.wantCursor)
			}
		})
	}
}

func Test_splitDelimiter(t *testing.T) {
	type args struct {
		str          string
		delimiter    string
		delimiterReg *regexp.Regexp
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			name: "blank",
			args: args{
				str:          "",
				delimiter:    ",",
				delimiterReg: nil,
			},
			want: nil,
		},
		{
			name: "abc",
			args: args{
				str:          "a,b,c",
				delimiter:    ",",
				delimiterReg: nil,
			},
			want: []int{0, 3, 5},
		},
		{
			name: "abc2",
			args: args{
				str:          "|aa|bb|cc|",
				delimiter:    "|",
				delimiterReg: nil,
			},
			want: []int{2, 5, 8}, // 23 | 56 | 89
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := splitByDelimiter(tt.args.str, tt.args.delimiter, tt.args.delimiterReg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitDelimiter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDocument_correctCursor(t *testing.T) {
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
			args: args{cursor: 10},
			want: 10,
		},
		{
			name: "CSVRight",
			fields: fields{
				fileName:        filepath.Join(testdata, "MOCK_DATA.csv"),
				columnDelimiter: ",",
				x:               10,
				width:           10,
			},
			args: args{cursor: 4},
			want: 2,
		},
		{
			name: "CSVLeft",
			fields: fields{
				fileName:        filepath.Join(testdata, "MOCK_DATA.csv"),
				columnDelimiter: ",",
				x:               30,
				width:           10,
			},
			args: args{cursor: 0},
			want: 4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := OpenDocument(tt.fields.fileName)
			if err != nil {
				t.Fatal(err)
			}
			m.WrapMode = tt.fields.wrap
			m.ColumnDelimiter = tt.fields.columnDelimiter
			m.x = tt.fields.x
			m.width = tt.fields.width
			for !m.BufEOF() {
			}
			if got := m.correctCursor(tt.args.cursor); got != tt.want {
				t.Errorf("Document.correctCursor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func moveColumnWidth(t *testing.T) *Document {
	t.Helper()
	m, err := OpenDocument(filepath.Join(testdata, "ps.txt"))
	if err != nil {
		t.Fatal(err)
	}
	for !m.BufEOF() {
	}
	m.ColumnWidth = true
	m.width = 80
	m.height = 23
	return m
}

func TestDocument_moveColumnWithLeft(t *testing.T) {
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
			m := moveColumnWidth(t)
			m.columnCursor = tt.fields.cursor
			numbers := make([]LineNumber, m.height)
			for i := 0; i < m.height; i++ {
				numbers[i] = LineNumber{
					number: i,
					wrap:   1,
				}
			}
			scr := SCR{
				numbers: numbers,
			}
			m.setColumnWidths()
			if err := m.moveColumnLeft(tt.args.n, scr, tt.args.cycle); (err != nil) != tt.wantErr {
				t.Errorf("Document.moveColumnWidthLeft() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got := m.columnCursor; got != tt.wantCursor {
				t.Errorf("Document.moveColumnWidthLeft() = %v, want %v", got, tt.wantCursor)
			}
		})
	}
}

func TestDocument_moveColumnWidthRight(t *testing.T) {
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
			wantErr:    false,
			wantCursor: 10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := moveColumnWidth(t)
			m.columnCursor = tt.fields.cursor
			numbers := make([]LineNumber, m.height)
			for i := 0; i < m.height; i++ {
				numbers[i] = LineNumber{
					number: i,
					wrap:   1,
				}
			}
			scr := SCR{
				numbers: numbers,
			}
			m.setColumnWidths()
			m.x = tt.fields.x
			if err := m.moveColumnRight(tt.args.n, scr, tt.args.cycle); (err != nil) != tt.wantErr {
				t.Errorf("Document.moveColumnWidthRight() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got := m.columnCursor; got != tt.wantCursor {
				t.Errorf("Document.moveColumnWidthRight() = %v, want %v", got, tt.wantCursor)
			}
		})
	}
}

func moveColumnDelimiter(t *testing.T) *Document {
	t.Helper()
	m, err := OpenDocument(filepath.Join(testdata, "MOCK_DATA.csv"))
	if err != nil {
		t.Fatal(err)
	}
	for !m.BufEOF() {
	}
	m.width = 80
	m.height = 23
	m.setDelimiter(",")
	return m
}

func TestDocument_moveColumnDelimiterLeft(t *testing.T) {
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
			wantErr:    false,
			wantCursor: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := moveColumnDelimiter(t)
			m.columnCursor = tt.fields.cursor
			numbers := make([]LineNumber, m.height)
			for i := 0; i < m.height; i++ {
				numbers[i] = LineNumber{
					number: i,
					wrap:   1,
				}
			}
			scr := SCR{
				numbers: numbers,
			}
			if err := m.moveColumnLeft(tt.args.n, scr, tt.args.cycle); (err != nil) != tt.wantErr {
				t.Errorf("Document.moveColumnDelimiterLeft() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got := m.columnCursor; got != tt.wantCursor {
				t.Errorf("Document.moveColumnDelimiterLeft() = %v, want %v", got, tt.wantCursor)
			}
		})
	}
}
