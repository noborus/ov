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
			m := docHelper(t, "test")
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

func TestDocument_optimalCursor(t *testing.T) {
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
				width:           10,
			},
			args: args{
				cursor: 4,
			},
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
			args: args{
				cursor: 0,
			},
			want: 4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := docFileReadHelper(t, tt.fields.fileName)
			m.WrapMode = tt.fields.wrap
			m.ColumnDelimiter = tt.fields.columnDelimiter
			m.x = tt.fields.x
			m.width = tt.fields.width
			scr := setupSCRHelper(t, m)
			if got := m.optimalCursor(scr, tt.args.cursor); got != tt.want {
				t.Errorf("Document.correctCursor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func moveColumnWidth(t *testing.T) *Document {
	t.Helper()
	m := docFileReadHelper(t, filepath.Join(testdata, "ps.txt"))
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
			wantErr:    true,
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

func MOCKDATADoc(t *testing.T) *Document {
	t.Helper()
	m := docFileReadHelper(t, filepath.Join(testdata, "MOCK_DATA.csv"))
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
			m := MOCKDATADoc(t)
			m.columnCursor = tt.fields.cursor
			lines := make(map[int]LineC, 1)
			lines[0] = LineC{
				valid: true,
				str:   "id,first_name,last_name,email,gender,ip_address,animal,app",
				pos:   []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 5},
			}
			scr := SCR{
				lines: lines,
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

func TestDocument_optimalXWidth(t *testing.T) {
	type args struct {
		cursor int
	}
	type fields struct {
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "csv1",
			args: args{cursor: 1},
			want: 9,
		},
		{
			name: "csvOver",
			args: args{cursor: 11},
			want: 66,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := MOCKDATADoc(t)
			m.columnWidths = []int{9, 16, 21, 26, 33, 39, 48, 53, 59, 66}
			got, err := m.optimalXWidth(tt.args.cursor)
			if (err != nil) != tt.wantErr {
				t.Errorf("Document.optimalXWidth() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Document.optimalXWidth() = %v, want %v", got, tt.want)
			}
		})
	}
}
