package oviewer

import (
	"context"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"codeberg.org/tslocum/cbind"
	"github.com/gdamore/tcell/v3"
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

func TestRoot_BackSearch(t *testing.T) {
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
			root.BackSearch(tt.args.str)
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
			if got := substr.MatchString(tt.args.s); got != tt.want {
				t.Errorf("searchWord.Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensitiveWord_Match(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
			substr := sensitiveWord{
				word: tt.fields.word,
			}
			if got := substr.MatchString(tt.args.s); got != tt.want {
				t.Errorf("sensitiveWord.Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_regexpWord_Match(t *testing.T) {
	t.Parallel()
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
			name: "testBlankLine",
			fields: fields{
				word: regexpCompile("^$", false),
			},
			args: args{
				"",
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
			t.Parallel()
			substr := regexpWord{
				regexp: tt.fields.word,
			}
			if got := substr.MatchString(tt.args.s); got != tt.want {
				t.Errorf("regexpWord.match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getSearchMatch(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
			searcher := NewSearcher(tt.args.searchWord, tt.args.searchReg, tt.args.caseSensitive, tt.args.regexpSearch)
			if got := searcher.MatchString(tt.word); got != tt.want {
				t.Errorf("getSearchMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_regexpCompile(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
			if got := regexpCompile(tt.args.r, tt.args.caseSensitive); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("regexpCompile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_searchPositionReg(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
			if got := searchPositionReg(tt.args.s, tt.args.re); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("searchPosition() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoot_setSearch(t *testing.T) {
	t.Parallel()
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
		config Config
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
		{
			name: "testSmartCaseSensitiveTrue",
			config: Config{
				SmartCaseSensitive: true,
			},
			fields: fields{
				input: &Input{},
			},
			args: args{
				word:          "Test",
				caseSensitive: false,
			},
			want: sensitiveWord{
				word: "Test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := &Root{
				input:  tt.fields.input,
				Config: tt.config,
			}
			if got := root.setSearcher(tt.args.word, tt.args.caseSensitive); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Root.setSearch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_multiRegexpCompile(t *testing.T) {
	t.Parallel()
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
		{
			name: "test2",
			args: args{
				[]string{"a", "b"},
			},
			want: []*regexp.Regexp{
				regexp.MustCompile("a"),
				regexp.MustCompile("b"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := multiRegexpCompile(tt.args.words); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("multiRegexpCompile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_condRegexpCompile(t *testing.T) {
	t.Parallel()
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
		{
			name: "testunclosed bracket",
			args: args{
				in: `/[abc/`,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := condRegexpCompile(tt.args.in); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("condRegexpCompile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDocument_searchChunk(t *testing.T) {
	t.Parallel()
	type args struct {
		chunkNum int
		searcher Searcher
	}
	tests := []struct {
		name     string
		fileName string
		args     args
		want     int
		wantErr  bool
	}{
		{
			name:     "testNotFound",
			fileName: filepath.Join(testdata, "ct.log"),
			args: args{
				chunkNum: 0,
				searcher: NewSearcher("test", regexpCompile("test", false), false, false),
			},
			want:    0,
			wantErr: true,
		},
		{
			name:     "testFound",
			fileName: filepath.Join(testdata, "ct.log"),
			args: args{
				chunkNum: 0,
				searcher: NewSearcher("error", regexpCompile("error", false), true, false),
			},
			want:    3,
			wantErr: false,
		},
		{
			name:     "testCaseSensitive",
			fileName: filepath.Join(testdata, "ct.log"),
			args: args{
				chunkNum: 0,
				searcher: NewSearcher("EXCEPTION", regexpCompile("EXCEPTION", false), true, false),
			},
			want:    0,
			wantErr: true,
		},
		{
			name:     "testRegexp",
			fileName: filepath.Join(testdata, "ct.log"),
			args: args{
				chunkNum: 0,
				searcher: NewSearcher("error", regexpCompile("error", true), true, true),
			},
			want:    3,
			wantErr: false,
		},
		{
			name:     "testEnd",
			fileName: filepath.Join(testdata, "ct.log"),
			args: args{
				chunkNum: 0,
				searcher: NewSearcher("\\.$", regexpCompile("\\.$", true), true, true),
			},
			want:    0,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := docFileReadHelper(t, tt.fileName)
			got, err := m.searchChunk(tt.args.chunkNum, tt.args.searcher)
			if (err != nil) != tt.wantErr {
				t.Errorf("Document.searchChunk() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Document.searchChunk() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensitive_FindAll(t *testing.T) {
	type fields struct {
		searchWord    string
		searchReg     *regexp.Regexp
		caseSensitive bool
		regexpSearch  bool
	}
	type args struct {
		s string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   [][]int
	}{
		{
			name: "testFound",
			fields: fields{
				searchWord:    "error",
				searchReg:     regexpCompile("error", false),
				caseSensitive: true,
				regexpSearch:  false,
			},
			args: args{
				s: "error",
			},
			want: [][]int{{0, 5}},
		},
		{
			name: "testCSFound",
			fields: fields{
				searchWord:    "error",
				searchReg:     regexpCompile("error", false),
				caseSensitive: false,
				regexpSearch:  false,
			},
			args: args{
				s: "error",
			},
			want: [][]int{{0, 5}},
		},
		{
			name: "testRegexpFound",
			fields: fields{
				searchWord:    "err*",
				searchReg:     regexpCompile("err*", false),
				caseSensitive: false,
				regexpSearch:  true,
			},
			args: args{
				s: "error",
			},
			want: [][]int{{0, 3}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			substr := NewSearcher(tt.fields.searchWord, tt.fields.searchReg, tt.fields.caseSensitive, tt.fields.regexpSearch)
			if got := substr.FindAll(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sensitiveWord.FindAll() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_search_String(t *testing.T) {
	type fields struct {
		searchWord    string
		searchReg     *regexp.Regexp
		caseSensitive bool
		regexpSearch  bool
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "testString",
			fields: fields{
				searchWord:    "error",
				searchReg:     regexpCompile("error", false),
				caseSensitive: false,
				regexpSearch:  false,
			},
			want: "error",
		},
		{
			name: "testCaseSensitive",
			fields: fields{
				searchWord:    "ERROR",
				searchReg:     regexpCompile("ERROR", true),
				caseSensitive: true,
				regexpSearch:  false,
			},
			want: "ERROR",
		},
		{
			name: "testRegexp",
			fields: fields{
				searchWord:    "err*",
				searchReg:     regexpCompile("err*", false),
				caseSensitive: false,
				regexpSearch:  true,
			},
			want: "err*",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			substr := NewSearcher(tt.fields.searchWord, tt.fields.searchReg, tt.fields.caseSensitive, tt.fields.regexpSearch)
			if got := substr.String(); got != tt.want {
				t.Errorf("searchWord.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoot_startSearchLN(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		fileName         string
		topLN            int
		lastSearchLN     int
		sectionHeaderNum int
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name: "test1",
			fields: fields{
				fileName:         filepath.Join(testdata, "test3.txt"),
				topLN:            10,
				lastSearchLN:     -1,
				sectionHeaderNum: 0,
			},
			want: 10,
		},
		{
			name: "test lastSearchLN",
			fields: fields{
				fileName:         filepath.Join(testdata, "test3.txt"),
				topLN:            10,
				lastSearchLN:     11,
				sectionHeaderNum: 3,
			},
			want: 11,
		},
		{
			name: "test topLN",
			fields: fields{
				fileName:         filepath.Join(testdata, "test3.txt"),
				topLN:            10,
				lastSearchLN:     9,
				sectionHeaderNum: 3,
			},
			want: 10,
		},
		{
			name: "test topLN2",
			fields: fields{
				fileName:         filepath.Join(testdata, "test3.txt"),
				topLN:            10,
				lastSearchLN:     14,
				sectionHeaderNum: 3,
			},
			want: 10,
		},
		{
			name: "test topLN3",
			fields: fields{
				fileName:         filepath.Join(testdata, "test3.txt"),
				topLN:            10,
				lastSearchLN:     8,
				sectionHeaderNum: 3,
			},
			want: 10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootFileReadHelper(t, tt.fields.fileName)
			root.prepareScreen()
			root.Doc.topLN = tt.fields.topLN
			root.Doc.lastSearchLN = tt.fields.lastSearchLN
			root.Doc.SectionHeaderNum = tt.fields.sectionHeaderNum
			ctx := context.Background()
			root.draw(ctx)
			if got := root.startSearchLN(); got != tt.want {
				t.Errorf("Root.startSearchLN() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoot_searchMove(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		fileName string
	}
	type args struct {
		forward  bool
		lineNum  int
		searcher Searcher
		nonMatch bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want1  bool
		want2  bool
	}{
		{
			name: "searcher nil",
			fields: fields{
				fileName: filepath.Join(testdata, "test3.txt"),
			},
			args: args{
				forward:  true,
				lineNum:  0,
				searcher: nil,
				nonMatch: false,
			},
			want1: false,
			want2: false,
		},
		{
			name: "test3 forward",
			fields: fields{
				fileName: filepath.Join(testdata, "test3.txt"),
			},
			args: args{
				forward:  true,
				lineNum:  0,
				searcher: NewSearcher("10", regexpCompile("10", false), false, false),
				nonMatch: false,
			},
			want1: true,
			want2: true,
		},
		{
			name: "test3 backForward",
			fields: fields{
				fileName: filepath.Join(testdata, "test3.txt"),
			},
			args: args{
				forward:  false,
				lineNum:  9000,
				searcher: NewSearcher("1000", regexpCompile("1000", false), false, false),
				nonMatch: false,
			},
			want1: true,
			want2: true,
		},
		{
			name: "test3 not found",
			fields: fields{
				fileName: filepath.Join(testdata, "test3.txt"),
			},
			args: args{
				forward:  true,
				lineNum:  0,
				searcher: NewSearcher("test", regexpCompile("test", false), false, false),
				nonMatch: false,
			},
			want1: false,
			want2: false,
		},
		{
			name: "test3 nonMatch",
			fields: fields{
				fileName: filepath.Join(testdata, "test3.txt"),
			},
			args: args{
				forward:  true,
				lineNum:  0,
				searcher: NewSearcher("test", regexpCompile("test", false), false, false),
				nonMatch: true,
			},
			want1: true,
			want2: true,
		},
		{
			name: "test3 backSearch nonMatch",
			fields: fields{
				fileName: filepath.Join(testdata, "test3.txt"),
			},
			args: args{
				forward:  false,
				lineNum:  1000,
				searcher: NewSearcher("test", regexpCompile("test", false), false, false),
				nonMatch: true,
			},
			want1: true,
			want2: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootFileReadHelper(t, tt.fields.fileName)
			root.prepareScreen()
			ctx := context.Background()
			root.everyUpdate(ctx)
			root.draw(ctx)
			root.Doc.nonMatch = tt.args.nonMatch
			root.searcher = tt.args.searcher
			if got := root.searchMove(ctx, tt.args.forward, tt.args.lineNum, tt.args.searcher); got != tt.want1 {
				t.Errorf("Root.searchMove() = %v, want %v", got, tt.want1)
			}
			/*
				if eventF := root.Screen.HasPendingEvent(); eventF != tt.want2 {
					t.Errorf("Root.searchMove() HasPendingEvent() = %v, want %v", eventF, tt.want2)
				}
			*/
		})
	}
}

func TestRoot_incSearch(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		fileName string
		topLN    int
		word     string
	}
	type args struct {
		forward bool
		lineNum int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "test3 forward",

			fields: fields{
				fileName: filepath.Join(testdata, "test3.txt"),
				word:     "10",
			},
			args: args{
				forward: true,
				lineNum: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootFileReadHelper(t, tt.fields.fileName)
			root.Doc.topLN = tt.fields.topLN
			root.input.value = tt.fields.word
			root.prepareScreen()
			ctx := context.Background()
			root.incSearch(ctx, tt.args.forward, tt.args.lineNum)
			root.returnStartPosition()
			if got := root.Doc.topLN; got != tt.fields.topLN {
				t.Errorf("Root.returnStartPosition() = %v, want %v", got, tt.fields.topLN)
			}
		})
	}
}

func Test_cancelKeys(t *testing.T) {
	type args struct {
		c          *cbind.Configuration
		cancelKeys []string
		cancelApp  func(_ *tcell.EventKey) *tcell.EventKey
	}
	tests := []struct {
		name    string
		args    args
		want    *cbind.Configuration
		wantErr bool
	}{
		{
			name: "test1",

			args: args{
				c:          cbind.NewConfiguration(),
				cancelKeys: []string{"ctrl+c", "c"},
				cancelApp:  nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := cancelKeys(tt.args.c, tt.args.cancelKeys, tt.args.cancelApp)
			if (err != nil) != tt.wantErr {
				t.Errorf("cancelKeys() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
