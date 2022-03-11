package oviewer

import (
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func Test_searchWord_Match(t *testing.T) {
	type fields struct {
		word string
	}
	type args struct {
		s string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "testTrue",
			fields: fields{
				word: "t",
			},
			args: args{
				"test",
			},
			want: true,
		},
		{
			name: "testFalse",
			fields: fields{
				word: "f",
			},
			args: args{
				"test",
			},
			want: false,
		},
		{
			name: "testEscapeSequences",
			fields: fields{
				word: "test",
			},
			args: args{
				"\x1B[31mtest\x1B[0m",
			},
			want: true,
		},
		{
			name: "testEscapeSequences2",
			fields: fields{
				word: "m",
			},
			args: args{
				"\x1B[31mtest\x1B[0m",
			},
			want: false,
		},
		{
			name: "testEscapeSequences3",
			fields: fields{
				word: "test",
			},
			args: args{
				"tes\x1B[31mt\x1B[0m",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			substr := searchWord{
				word: tt.fields.word,
			}
			if got := substr.Match(tt.args.s); got != tt.want {
				t.Errorf("searchWord.Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensitiveWord_Match(t *testing.T) {
	type fields struct {
		word string
	}
	type args struct {
		s string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "testTrue",
			fields: fields{
				word: "t",
			},
			args: args{
				"test",
			},
			want: true,
		},
		{
			name: "testFalse",
			fields: fields{
				word: "t",
			},
			args: args{
				"TEST",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			substr := sensitiveWord{
				word: tt.fields.word,
			}
			if got := substr.Match(tt.args.s); got != tt.want {
				t.Errorf("sensitiveWord.Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_regexpWord_Match(t *testing.T) {
	type fields struct {
		word *regexp.Regexp
	}
	type args struct {
		s string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "testTrue",
			fields: fields{
				word: regexpCompile("t", false),
			},
			args: args{
				"test",
			},
			want: true,
		},
		{
			name: "testFalse",
			fields: fields{
				word: regexpCompile("t", true),
			},
			args: args{
				"TEST",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			substr := regexpWord{
				word: tt.fields.word,
			}
			if got := substr.Match(tt.args.s); got != tt.want {
				t.Errorf("regexpWord.match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getSearchMatch(t *testing.T) {
	type args struct {
		searchWord    string
		searchReg     *regexp.Regexp
		caseSensitive bool
		regexpSearch  bool
	}
	tests := []struct {
		name string
		args args
		word string
		want bool
	}{
		{
			name: "testInSensitive",
			args: args{
				searchWord:    "t",
				searchReg:     nil,
				caseSensitive: false,
				regexpSearch:  false,
			},
			word: "test",
			want: true,
		},
		{
			name: "testSensitive",
			args: args{
				searchWord:    "t",
				searchReg:     nil,
				caseSensitive: true,
				regexpSearch:  false,
			},
			word: "test",
			want: true,
		},
		{
			name: "testRegexpInSensitive",
			args: args{
				searchWord:    "t.",
				searchReg:     regexpCompile("t.", false),
				caseSensitive: false,
				regexpSearch:  true,
			},
			word: "test",
			want: true,
		},
		{
			name: "testRegexpSensitive",
			args: args{
				searchWord:    "T.",
				searchReg:     regexpCompile("T", true),
				caseSensitive: true,
				regexpSearch:  true,
			},
			word: "test",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			searcher := NewSearcher(tt.args.searchWord, tt.args.searchReg, tt.args.caseSensitive, tt.args.regexpSearch)
			if got := searcher.Match(tt.word); got != tt.want {
				t.Errorf("getSearchMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_regexpCompile(t *testing.T) {
	type args struct {
		r             string
		caseSensitive bool
	}
	tests := []struct {
		name string
		args args
		want *regexp.Regexp
	}{
		{
			name: "regexpTrue",
			args: args{
				r:             "t.",
				caseSensitive: true,
			},
			want: regexp.MustCompile("t."),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := regexpCompile(tt.args.r, tt.args.caseSensitive); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("regexpCompile() = %v, want %v", got, tt.want)
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
			name: "test1",
			args: args{
				s:      "test",
				substr: "t",
				number: 0,
			},
			wantS: 1,
			wantE: 3,
		},
		{
			name: "test2",
			args: args{
				s:      "test",
				substr: "t",
				number: 1,
			},
			wantS: 4,
			wantE: 4,
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

func Test_searchPositionReg(t *testing.T) {
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
				re: nil,
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
				re: regexpCompile("a", false),
			},
			want: nil,
		},
		{
			name: "testInCaseSensitive",
			args: args{
				s:  "TEST",
				re: regexpCompile("e", false),
			},
			want: [][]int{
				{1, 2},
			},
		},
		{
			name: "testCaseSensitive",
			args: args{
				s:  "TEST",
				re: regexpCompile("e", true),
			},
			want: nil,
		},
		{
			name: "testMeta",
			args: args{
				s:  "test",
				re: regexpCompile("+", false),
			},
			want: nil,
		},
		{
			name: "testMeta2",
			args: args{
				s:  "test",
				re: regexpCompile("t+", false),
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
				re: regexpCompile("man", false),
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

func Test_searchPosition(t *testing.T) {
	type args struct {
		caseSensitive bool
		s             string
		substr        string
	}
	tests := []struct {
		name string
		args args
		want [][]int
	}{
		{
			name: "testNil",
			args: args{
				caseSensitive: false,
				s:             "t",
				substr:        "",
			},
			want: nil,
		},
		{
			name: "testTest",
			args: args{
				caseSensitive: false,
				s:             "test",
				substr:        "t",
			},
			want: [][]int{
				{0, 1},
				{3, 4},
			},
		},
		{
			name: "testNone",
			args: args{
				caseSensitive: false,
				s:             "",
				substr:        "test",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := searchPositionStr(tt.args.caseSensitive, tt.args.s, tt.args.substr); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("searchPosition() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoot_setSearch(t *testing.T) {
	type fields struct {
		input *Input
	}
	type args struct {
		word          string
		caseSensitive bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   Searcher
	}{
		{
			name: "testNil",
			fields: fields{
				input: &Input{},
			},
			args: args{
				word:          "",
				caseSensitive: false,
			},
			want: nil,
		},
		{
			name: "test1",
			fields: fields{
				input: &Input{},
			},
			args: args{
				word:          "test",
				caseSensitive: false,
			},
			want: searchWord{
				word: strings.ToLower("test"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := &Root{
				input: tt.fields.input,
			}
			if got := root.setSearcher(tt.args.word, tt.args.caseSensitive); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Root.setSearch() = %v, want %v", got, tt.want)
			}
		})
	}
}
