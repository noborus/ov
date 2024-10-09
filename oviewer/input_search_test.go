package oviewer

import (
	"reflect"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func Test_eventInputSearch_Prompt(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "searchPrompt",
			want: "/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := NewInput()
			e := newSearchEvent(input.SearchCandidate, forward)
			if got := e.Prompt(); got != tt.want {
				t.Errorf("eventInputSearch.Prompt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_eventInputSearch_Confirm(t *testing.T) {
	type fields struct {
		EventTime tcell.EventTime
		clist     *candidate
		value     string
	}
	type args struct {
		str string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "key enter",
			fields: fields{
				EventTime: tcell.EventTime{},
				clist:     blankCandidate(),
				value:     "",
			},
			args: args{
				str: "test",
			},
			want: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &eventInputSearch{
				EventTime: tt.fields.EventTime,
				clist:     tt.fields.clist,
				value:     tt.fields.value,
			}
			if got := e.Confirm(tt.args.str); got.(*eventInputSearch).value != tt.want {
				t.Errorf("eventInputSearch.Confirm() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_eventInputSearch_Up(t *testing.T) {
	type fields struct {
		list []string
	}
	type args struct {
		str string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "key up",
			fields: fields{
				list: []string{"a", "b", "c"},
			},
			args: args{
				str: "testLast",
			},
			want: "testLast",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := NewInput()
			for _, v := range tt.fields.list {
				input.SearchCandidate.toAddLast(v)
			}
			e := newSearchEvent(input.SearchCandidate, forward)
			if got := e.Up(tt.args.str); got != tt.want {
				t.Errorf("eventInputSearch.Up() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_eventInputSearch_Down(t *testing.T) {
	type fields struct {
		list []string
	}
	type args struct {
		str string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "key down",
			fields: fields{
				list: []string{"a", "b", "c"},
			},
			args: args{
				str: "testTop",
			},
			want: "a",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := NewInput()
			for _, v := range tt.fields.list {
				input.SearchCandidate.toAddLast(v)
			}
			e := newSearchEvent(input.SearchCandidate, forward)
			if got := e.Down(tt.args.str); got != tt.want {
				t.Logf("%v\n", e.clist.list)
				t.Errorf("eventInputSearch.Down() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInput_searchCandidates(t *testing.T) {
	type args struct {
		n int
	}
	tests := []struct {
		name   string
		fields struct {
			list []string
		}
		args args
		want []string
	}{
		{
			name: "testSearchCandidates",
			fields: struct {
				list []string
			}{
				list: []string{"a", "b", "c"},
			},
			args: args{
				n: 10,
			},
			want: []string{"a", "b", "c"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := NewInput()
			for _, v := range tt.fields.list {
				input.SearchCandidate.toAddLast(v)
			}
			if got := input.searchCandidates(tt.args.n); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Input.searchCandidates() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoot_setSearchMode(t *testing.T) {
	root := rootHelper(t)
	type fields struct {
		SmartCaseSensitive bool
		CaseSensitive      bool
		Regexp             bool
		nonMatch           bool
	}
	type args struct {
		searchType searchType
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "forwardsearchMode",
			fields: fields{
				SmartCaseSensitive: true,
				CaseSensitive:      false,
				Regexp:             true,
				nonMatch:           false,
			},
			args: args{
				searchType: forward,
			},
			want: "(R)(S)",
		},
		{
			name: "backsearchMode",
			fields: fields{
				SmartCaseSensitive: false,
				CaseSensitive:      false,
				Regexp:             true,
				nonMatch:           false,
			},
			args: args{
				searchType: backward,
			},
			want: "(R)",
		},
		{
			name: "backsearchModeCase",
			fields: fields{
				SmartCaseSensitive: false,
				CaseSensitive:      true,
				Regexp:             false,
				nonMatch:           false,
			},
			args: args{
				searchType: backward,
			},
			want: "(Aa)",
		},
		{
			name: "filterMode",
			fields: fields{
				SmartCaseSensitive: false,
				CaseSensitive:      false,
				Regexp:             true,
				nonMatch:           false,
			},
			args: args{
				searchType: filter,
			},
			want: "(R)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root.Config.SmartCaseSensitive = tt.fields.SmartCaseSensitive
			root.Config.CaseSensitive = tt.fields.CaseSensitive
			root.Config.RegexpSearch = tt.fields.Regexp
			root.setSearchMode(tt.args.searchType)
			if got := root.searchOpt; got != tt.want {
				t.Errorf("Root.setSearchMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_stripBackSlash(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "nonEsacape",
			args: args{
				str: "\test",
			},
			want: "\test",
		},
		{
			name: "nonEsacape2",
			args: args{
				str: "!test",
			},
			want: "!test",
		},
		{
			name: "backSlash1",
			args: args{
				str: `\!test`,
			},
			want: "!test",
		},
		{
			name: "backSlash2",
			args: args{
				str: `te\!st`,
			},
			want: "te!st",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stripBackSlash(tt.args.str); got != tt.want {
				t.Errorf("stripSlash() = %v, want %v", got, tt.want)
			}
		})
	}
}
