package oviewer

import (
	"reflect"
	"regexp"
	"testing"
)

func TestRoot_contains(t *testing.T) {
	type fields struct {
		input *Input
	}
	type args struct {
		s          string
		searchType SearchType
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "test1",
			fields: fields{
				input: &Input{
					value: "t",
				},
			},
			args: args{
				s:          "test",
				searchType: searchInsensitive,
			},
			want: true,
		},
		{
			name: "testEscapeSequences",
			fields: fields{
				input: &Input{
					value: "test",
				},
			},
			args: args{
				s:          "\x1B[31mtest\x1B[0m",
				searchType: searchRegexp,
			},
			want: true,
		},
		{
			name: "testEscapeSequences2",
			fields: fields{
				input: &Input{
					value: "m",
				},
			},
			args: args{
				s:          "\x1B[31mtest\x1B[0m",
				searchType: searchRegexp,
			},
			want: false,
		},
		{
			name: "testEscapeSequences3",
			fields: fields{
				input: &Input{
					value: "test",
				},
			},
			args: args{
				s:          "tes\x1B[31mt\x1B[0m",
				searchType: searchRegexp,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := &Root{
				input: tt.fields.input,
			}
			root.searchReg = regexp.MustCompile(root.input.value)
			if got := root.contains(tt.args.s, tt.args.searchType); got != tt.want {
				t.Errorf("Root.contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_rangePosition(t *testing.T) {
	type args struct {
		s      string
		substr string
		number int
	}
	tests := []struct {
		name  string
		args  args
		wantS int
		wantE int
	}{
		{
			name:  "testNil",
			args:  args{},
			wantS: 0,
			wantE: 0,
		},
		{
			name: "testNil2",
			args: args{
				s:      "test",
				substr: "t",
				number: 0,
			},
			wantS: 0,
			wantE: 0,
		},
		{
			name: "test",
			args: args{
				s:      "test",
				substr: "t",
				number: 1,
			},
			wantS: 1,
			wantE: 3,
		},
		{
			name: "testComma",
			args: args{
				s:      "a,b,c",
				substr: ",",
				number: 1,
			},
			wantS: 2,
			wantE: 3,
		},
		{
			name: "testVerticalBar",
			args: args{
				s:      "a|b|c",
				substr: "|",
				number: 2,
			},
			wantS: 4,
			wantE: 5,
		},
		{
			name: "testUnicodeBar",
			args: args{
				s:      "a│b│c",
				substr: "│",
				number: 1,
			},
			wantS: 4,
			wantE: 5,
		},
		{
			name: "testUnicodeBar2",
			args: args{
				s:      "a│b│c",
				substr: "│",
				number: 2,
			},
			wantS: 8,
			wantE: 9,
		},
		{
			name: "testUnicodeBar3",
			args: args{
				s:      "a│b│c",
				substr: "│",
				number: 3,
			},
			wantS: -1,
			wantE: -1,
		},
		{
			name: "testNone",
			args: args{
				s:      "a│b│c",
				substr: "│",
				number: 9,
			},
			wantS: -1,
			wantE: -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotS, gotE := rangePosition(tt.args.s, tt.args.substr, tt.args.number)
			if gotS != tt.wantS {
				t.Errorf("rangePosition() got = %v, want %v", gotS, tt.wantS)
			}
			if gotE != tt.wantE {
				t.Errorf("rangePosition() got1 = %v, want %v", gotE, tt.wantE)
			}
		})
	}
}

func Test_searchPosition(t *testing.T) {
	type args struct {
		s  string
		re *regexp.Regexp
	}
	tests := []struct {
		name string
		args args
		want [][]int
	}{
		{
			name: "testNil",
			args: args{
				s:  "",
				re: regexp.MustCompile("t"),
			},
			want: nil,
		},
		{
			name: "testTest",
			args: args{
				s:  "test",
				re: regexp.MustCompile("t"),
			},
			want: [][]int{
				{0, 1},
				{3, 4},
			},
		},
		{
			name: "testNone",
			args: args{
				s:  "testtest",
				re: regexpComple("a", false),
			},
			want: nil,
		},
		{
			name: "testInCaseSensitive",
			args: args{
				s:  "TEST",
				re: regexpComple("e", false),
			},
			want: [][]int{
				{1, 2},
			},
		},
		{
			name: "testCaseSensitive",
			args: args{
				s:  "TEST",
				re: regexpComple("e", true),
			},
			want: nil,
		},
		{
			name: "testMeta",
			args: args{
				s:  "test",
				re: regexpComple("+", false),
			},
			want: nil,
		},
		{
			name: "testMeta2",
			args: args{
				s:  "test",
				re: regexpComple("t+", false),
			},
			want: [][]int{
				{0, 1},
				{3, 4},
			},
		},
		{
			name: "testM",
			args: args{
				s:  "man",
				re: regexpComple("man", false),
			},
			want: [][]int{
				{0, 3},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := searchPositionReg(tt.args.s, tt.args.re); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("searchPosition() = %v, want %v", got, tt.want)
			}
		})
	}
}
