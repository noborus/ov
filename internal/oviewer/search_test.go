package oviewer

import (
	"reflect"
	"regexp"
	"testing"
)

func Test_rangePosition(t *testing.T) {
	type args struct {
		s      string
		substr string
		number int
	}
	tests := []struct {
		name string
		args args
		want rangePos
	}{
		{
			name: "testNil",
			args: args{},
			want: rangePos{
				start: 0,
				end:   0,
			},
		},
		{
			name: "testNil2",
			args: args{
				s:      "test",
				substr: "t",
				number: 0,
			},
			want: rangePos{
				start: 0,
				end:   0,
			},
		},
		{
			name: "test",
			args: args{
				s:      "test",
				substr: "t",
				number: 1,
			},
			want: rangePos{
				start: 1,
				end:   3,
			},
		},
		{
			name: "testComma",
			args: args{
				s:      "a,b,c",
				substr: ",",
				number: 1,
			},
			want: rangePos{
				start: 2,
				end:   3,
			},
		},
		{
			name: "testVerticalBar",
			args: args{
				s:      "a|b|c",
				substr: "|",
				number: 2,
			},
			want: rangePos{
				start: 4,
				end:   5,
			},
		},
		{
			name: "testUnicodeBar",
			args: args{
				s:      "a│b│c",
				substr: "│",
				number: 1,
			},
			want: rangePos{
				start: 4,
				end:   5,
			},
		},
		{
			name: "testUnicodeBar2",
			args: args{
				s:      "a│b│c",
				substr: "│",
				number: 2,
			},
			want: rangePos{
				start: 8,
				end:   9,
			},
		},
		{
			name: "testUnicodeBar3",
			args: args{
				s:      "a│b│c",
				substr: "│",
				number: 3,
			},
			want: rangePos{
				start: -1,
				end:   -1,
			},
		},
		{
			name: "testNone",
			args: args{
				s:      "a│b│c",
				substr: "│",
				number: 9,
			},
			want: rangePos{
				start: -1,
				end:   -1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := rangePosition(tt.args.s, tt.args.substr, tt.args.number); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("rangePosition() = %v, want %v", got, tt.want)
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
		want []rangePos
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
			want: []rangePos{
				{
					start: 0,
					end:   1,
				},
				{
					start: 3,
					end:   4,
				},
			},
		},
		{
			name: "testNone",
			args: args{
				s:  "testtest",
				re: regexp.MustCompile("a"),
			},
			want: nil,
		},
		{
			name: "testInCaseSensitive:",
			args: args{
				s:  "TEST",
				re: regexp.MustCompile("(?i)e"),
			},
			want: []rangePos{
				{
					start: 1,
					end:   2,
				},
			},
		},
		{
			name: "testCaseSensitive:",
			args: args{
				s:  "TEST",
				re: regexp.MustCompile("e"),
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := searchPosition(tt.args.s, tt.args.re); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("searchPosition() = %v, want %v", got, tt.want)
			}
		})
	}
}
