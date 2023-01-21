package oviewer

import (
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestRoot_Search(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		fileNames []string
	}
	type args struct {
		str string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test1",
			fields: fields{
				fileNames: []string{
					filepath.Join(testdata, "test.txt"),
				},
			},
			args: args{
				str: "test",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := openFiles(tt.fields.fileNames)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewOviewer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			root.Search(tt.args.str)
		})
	}
}

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

func Test_multiRegexpCompile(t *testing.T) {
	type args struct {
		words []string
	}
	tests := []struct {
		name string
		args args
		want []*regexp.Regexp
	}{
		{
			name: "test1",
			args: args{
				[]string{".", "["},
			},
			want: []*regexp.Regexp{
				regexp.MustCompile("."),
				regexp.MustCompile(`\[`),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := multiRegexpCompile(tt.args.words); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("multiRegexpCompile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_condRegexpCompile(t *testing.T) {
	type args struct {
		in string
	}
	tests := []struct {
		name string
		args args
		want *regexp.Regexp
	}{
		{
			name: "testRegexpCompile",
			args: args{
				in: "/[,|;]/",
			},
			want: regexp.MustCompile(`[,|;]`),
		},
		{
			name: "testString1",
			args: args{
				in: ",",
			},
			want: nil,
		},
		{
			name: "testString2",
			args: args{
				in: "/",
			},
			want: nil,
		},
		{
			name: "testString4",
			args: args{
				in: "/end",
			},
			want: nil,
		},
		{
			name: "testString5",
			args: args{
				in: "s/e",
			},
			want: nil,
		},
		{
			name: "testString6",
			args: args{
				in: `\/\/`,
			},
			want: nil,
		},
		{
			name: "testString7",
			args: args{
				in: `//`,
			},
			want: regexp.MustCompile(``),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := condRegexpCompile(tt.args.in); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("condRegexpCompile() = %v, want %v", got, tt.want)
			}
		})
	}
}
