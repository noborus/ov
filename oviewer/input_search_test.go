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
