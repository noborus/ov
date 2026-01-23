package oviewer

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/gdamore/tcell/v3"
)

func moveText(t *testing.T) *Document {
	t.Helper()
	m := docFileReadHelper(t, filepath.Join(testdata, "MOCK_DATA.csv"))
	m.bodyWidth = 80
	m.height = 23
	return m
}

func TestDocument_moveLine(t *testing.T) {
	type args struct {
		lN int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "moveNoLine",
			args: args{
				lN: 10,
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewDocument()
			if err != nil {
				t.Fatal(err)
			}
			if got := m.moveLine(tt.args.lN); got != tt.want {
				t.Errorf("Document.moveLine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDocument_moveLineFile(t *testing.T) {
	type fields struct {
		fileName string
	}
	type args struct {
		lN int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			name: "moveLine",
			fields: fields{
				fileName: filepath.Join(testdata, "move.txt"),
			},
			args: args{
				lN: 10,
			},
			want: 10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := docFileReadHelper(t, tt.fields.fileName)
			if got := m.moveLine(tt.args.lN); got != tt.want {
				t.Errorf("Document.moveLine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDocument_moveTopFile(t *testing.T) {
	type fields struct {
		fileName string
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name: "moveTop",
			fields: fields{
				fileName: filepath.Join(testdata, "move.txt"),
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := docFileReadHelper(t, tt.fields.fileName)
			m.moveTop()
			if got := m.topLN; got != tt.want {
				t.Errorf("Document.moveLine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDocument_moveLineNth(t *testing.T) {
	type fields struct {
		fileName string
		width    int
		wrap     bool
	}
	type args struct {
		lN  int
		nTh int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
		want1  int
	}{
		{
			name: "moveNoLineNth",
			fields: fields{
				fileName: filepath.Join(testdata, "move.txt"),
				width:    10,
				wrap:     true,
			},
			args: args{
				lN:  1,
				nTh: 6,
			},
			want:  1,
			want1: 6,
		},
		{
			name: "moveNoLineNthNone",
			fields: fields{
				fileName: filepath.Join(testdata, "move.txt"),
				width:    10,
				wrap:     true,
			},
			args: args{
				lN:  3,
				nTh: 6,
			},
			want:  3,
			want1: 0,
		},
		{
			name: "moveNoLineNthNoWrap",
			fields: fields{
				fileName: filepath.Join(testdata, "move.txt"),
				width:    10,
				wrap:     false,
			},
			args: args{
				lN:  1,
				nTh: 6,
			},
			want:  1,
			want1: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := docFileReadHelper(t, tt.fields.fileName)
			m.bodyWidth = tt.fields.width
			m.WrapMode = tt.fields.wrap
			got, got1 := m.moveLineNth(tt.args.lN, tt.args.nTh)
			if got != tt.want {
				t.Errorf("Document.moveLineNth() LN = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Document.moveLineNth() nTh = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestDocument_moveBottomFile(t *testing.T) {
	type fields struct {
		fileName string
		width    int
		height   int
		wrap     bool
	}
	tests := []struct {
		name   string
		fields fields
		want   int
		want1  int
	}{
		{
			name: "moveBottomNoWrap",
			fields: fields{
				fileName: filepath.Join(testdata, "move.txt"),
				width:    10,
				height:   10,
				wrap:     false,
			},
			want:  0,
			want1: 33,
		},
		{
			name: "moveBottomWrap",
			fields: fields{
				fileName: filepath.Join(testdata, "move.txt"),
				width:    10,
				height:   10,
				wrap:     true,
			},
			want:  70,
			want1: 40,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := docFileReadHelper(t, tt.fields.fileName)
			m.bodyWidth = tt.fields.width
			m.height = tt.fields.height
			m.WrapMode = tt.fields.wrap
			m.moveBottom()
			if got := m.topLX; got != tt.want {
				t.Errorf("Document.moveLine() LX = %v, want %v", got, tt.want)
			}
			if got := m.topLN; got != tt.want1 {
				t.Errorf("Document.moveLine() LN = %v, want %v", got, tt.want1)
			}
		})
	}
}

func Test_leftMostX(t *testing.T) {
	type args struct {
		width int
		lc    contents
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			name: "oneLine",
			args: args{
				width: 10,
				lc:    StrToContents("0123456789", 8),
			},
			want: []int{0},
		},
		{
			name: "twoLine",
			args: args{
				width: 10,
				lc:    StrToContents("000000000111111111", 8),
			},
			want: []int{0, 10},
		},
		{
			name: "fullWidth",
			args: args{
				width: 10,
				// 0あいうえ
				// おかきくけ
				// こ
				lc: StrToContents("0あいうえおかきくけこ", 8),
			},
			want: []int{0, 9, 19},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := leftX(tt.args.width, tt.args.lc); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("leftMostX() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDocument_moveYUp(t *testing.T) {
	type fields struct {
		fileName string
		lX       int
		lN       int
		width    int
		height   int
		wrap     bool
	}
	type args struct {
		moveY int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
		want1  int
	}{
		{
			name: "moveYUp",
			fields: fields{
				fileName: filepath.Join(testdata, "move.txt"),
				lX:       0,
				lN:       3,
				width:    10,
				height:   10,
				wrap:     false,
			},
			args: args{
				moveY: 1,
			},
			want:  0,
			want1: 2,
		},
		{
			name: "moveYUpWrap",
			fields: fields{
				fileName: filepath.Join(testdata, "move.txt"),
				lX:       40,
				lN:       1,
				width:    10,
				height:   10,
				wrap:     true,
			},
			args: args{
				moveY: 1,
			},
			want:  30,
			want1: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := docFileReadHelper(t, tt.fields.fileName)
			m.bodyWidth = tt.fields.width
			m.height = tt.fields.height
			m.WrapMode = tt.fields.wrap
			m.topLN = tt.fields.lN
			m.topLX = tt.fields.lX
			m.moveLimitYUp(tt.args.moveY)
			if m.topLX != tt.want {
				t.Errorf("Document.moveYUp() LX = %v, want %v", m.topLX, tt.want)
			}
			if m.topLN != tt.want1 {
				t.Errorf("Document.moveYUp() LN = %v, want %v", m.topLN, tt.want1)
			}
		})
	}
}

func Test_leftX(t *testing.T) {
	type args struct {
		width int
		lc    contents
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			name: "oneLine",
			args: args{
				width: 10,
				lc:    StrToContents("0123456789", 8),
			},
			want: []int{0},
		},
		{
			name: "MixedWidth",
			args: args{
				width: 80,
				lc:    StrToContents(`47382,"90718","9071800","オキナワケン","ヤエヤマグンヨナグニチョウ","イカニケイサイガナイバアイ","沖縄県","八重山郡与那国町","以下に掲載がない場合",0,0,0,0,0,0`, 8),
			},
			want: []int{0, 79},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := leftX(tt.args.width, tt.args.lc); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("leftX() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDocument_movePgUpDown(t *testing.T) {
	type fields struct {
		down int
		up   int
		wrap bool
	}
	tests := []struct {
		name   string
		fields fields
		wantLX int
		wantLN int
	}{
		{
			name: "movePgUp",
			fields: fields{
				down: 2,
				up:   1,
				wrap: false,
			},
			wantLX: 0,
			wantLN: 23,
		},
		{
			name: "moveWrapPgUpDown",
			fields: fields{
				down: 2,
				up:   1,
				wrap: true,
			},
			wantLX: 0,
			wantLN: 13,
		},
		{
			name: "moveWrapPgUpDown2",
			fields: fields{
				down: 2,
				up:   3,
				wrap: true,
			},
			wantLX: 0,
			wantLN: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := moveText(t)
			m.WrapMode = tt.fields.wrap
			for range tt.fields.down {
				m.movePgDn()
			}
			for range tt.fields.up {
				m.movePgUp()
			}
			if m.topLX != tt.wantLX {
				t.Errorf("Document.movePgUpDown() LX = %v, wantLX %v", m.topLX, tt.wantLX)
			}
			if m.topLN != tt.wantLN {
				t.Errorf("Document.movePgUpDown() LN = %v, wantLN %v", m.topLN, tt.wantLN)
			}
		})
	}
}

func TestDocument_moveHfUpDown(t *testing.T) {
	type fields struct {
		down int
		up   int
		wrap bool
	}
	tests := []struct {
		name   string
		fields fields
		wantLX int
		wantLN int
	}{
		{
			name: "moveHfUpDown",
			fields: fields{
				down: 2,
				up:   1,
				wrap: false,
			},
			wantLX: 0,
			wantLN: 11,
		},
		{
			name: "moveWrapHfUpDown",
			fields: fields{
				down: 2,
				up:   1,
				wrap: true,
			},
			wantLX: 0,
			wantLN: 6,
		},
		{
			name: "moveHfUpDown2",
			fields: fields{
				down: 1,
				up:   3,
				wrap: false,
			},
			wantLX: 0,
			wantLN: 0,
		},
		{
			name: "moveWrapHfUpDown2",
			fields: fields{
				down: 2,
				up:   3,
				wrap: true,
			},
			wantLX: 0,
			wantLN: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := moveText(t)
			m.WrapMode = tt.fields.wrap
			for range tt.fields.down {
				m.moveHfDn()
			}
			for range tt.fields.up {
				m.moveHfUp()
			}
			if m.topLX != tt.wantLX {
				t.Errorf("Document.moveHfUpDown() LX = %v, wantLX %v", m.topLX, tt.wantLX)
			}
			if m.topLN != tt.wantLN {
				t.Errorf("Document.moveHfUpDown() LN = %v, wantLN %v", m.topLN, tt.wantLN)
			}
		})
	}
}

func TestDocument_limitMoveDown(t *testing.T) {
	type fields struct {
		wrap bool
	}
	type args struct {
		lX int
		lN int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		wantLX int
		wantLN int
	}{
		{
			name: "NormalNoWrap",
			fields: fields{
				wrap: false,
			},
			args: args{
				lX: 0,
				lN: 10,
			},
			wantLX: 0,
			wantLN: 10,
		},
		{
			name: "NormalWrap",
			fields: fields{
				wrap: true,
			},
			args: args{
				lX: 0,
				lN: 10,
			},
			wantLX: 0,
			wantLN: 10,
		},
		{
			name: "overNoWrap",
			fields: fields{
				wrap: false,
			},
			args: args{
				lX: 0,
				lN: 2000,
			},
			wantLX: 0,
			wantLN: 979,
		},
		{
			name: "boundary",
			fields: fields{
				wrap: true,
			},
			args: args{
				lX: 0,
				lN: 988,
			},
			wantLX: 0,
			wantLN: 988,
		},
		{
			name: "overWrap",
			fields: fields{
				wrap: true,
			},
			args: args{
				lX: 0,
				lN: 2000,
			},
			wantLX: 80,
			wantLN: 988,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := moveText(t)
			m.WrapMode = tt.fields.wrap
			m.limitMoveDown(tt.args.lX, tt.args.lN)
			if m.topLX != tt.wantLX {
				t.Errorf("Document.limitMoveDown() LX = %v, wantLX %v", m.topLX, tt.wantLX)
			}
			if m.topLN != tt.wantLN {
				t.Errorf("Document.limitMoveDown() LN = %v, wantLN %v", m.topLN, tt.wantLN)
			}
		})
	}
}

func TestDocument_nextSection(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		sectionDelimiter string
	}
	type args struct {
		ctx context.Context
		lN  int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "nextSection1",
			fields: fields{
				sectionDelimiter: "^#",
			},
			args: args{
				ctx: context.Background(),
				lN:  0,
			},
			want:    2,
			wantErr: false,
		},
		{
			name: "nextSection2",
			fields: fields{
				sectionDelimiter: "^#",
			},
			args: args{
				ctx: context.Background(),
				lN:  2,
			},
			want:    4,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := docFileReadHelper(t, filepath.Join(testdata, "section.txt"))
			m.setSectionDelimiter(tt.fields.sectionDelimiter)
			got, err := m.nextSection(tt.args.ctx, tt.args.lN)
			if (err != nil) != tt.wantErr {
				t.Errorf("Document.nextSection() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Document.nextSection() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDocument_prevSection(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		sectionDelimiter string
	}
	type args struct {
		ctx context.Context
		lN  int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "prevSection1",
			fields: fields{
				sectionDelimiter: "^#",
			},
			args: args{
				ctx: context.Background(),
				lN:  4,
			},
			want:    2,
			wantErr: false,
		},
		{
			name: "prevSection2",
			fields: fields{
				sectionDelimiter: "^#",
			},
			args: args{
				ctx: context.Background(),
				lN:  2,
			},
			want:    0,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := docFileReadHelper(t, filepath.Join(testdata, "section.txt"))
			m.setSectionDelimiter(tt.fields.sectionDelimiter)
			got, err := m.prevSection(tt.args.ctx, tt.args.lN)
			if (err != nil) != tt.wantErr {
				t.Errorf("Document.prevSection() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Document.prevSection() = %v, want %v", got, tt.want)
			}
		})
	}
}
