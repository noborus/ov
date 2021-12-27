package oviewer

import (
	"reflect"
	"testing"
)

func Test_max(t *testing.T) {
	type args struct {
		a int
		b int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "test1",
			args: args{a: 1, b: 0},
			want: 1,
		},
		{
			name: "test2",
			args: args{a: 1, b: 2},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := max(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("max() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_min(t *testing.T) {
	type args struct {
		a int
		b int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "test1",
			args: args{a: 1, b: 0},
			want: 0,
		},
		{
			name: "test2",
			args: args{a: 1, b: 2},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := min(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("min() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_removeStr(t *testing.T) {
	type args struct {
		list []string
		s    string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "test1",
			args: args{
				list: []string{"a", "b", "c"},
				s:    "c",
			},
			want: []string{"a", "b"},
		},
		{
			name: "testZero",
			args: args{
				list: []string{},
				s:    "c",
			},
			want: []string{},
		},
		{
			name: "noRemove",
			args: args{
				list: []string{"a", "b", "c"},
				s:    "",
			},
			want: []string{"a", "b", "c"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := removeStr(tt.args.list, tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("removeStr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_removeInt(t *testing.T) {
	type args struct {
		list []int
		c    int
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			name: "test1",
			args: args{
				list: []int{1, 2, 3},
				c:    3,
			},
			want: []int{1, 2},
		},
		{
			name: "testZero",
			args: args{
				list: []int{},
				c:    3,
			},
			want: []int{},
		},
		{
			name: "noRemove",
			args: args{
				list: []int{1, 2, 3},
				c:    4,
			},
			want: []int{1, 2, 3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := removeInt(tt.args.list, tt.args.c); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("removeInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_containsInt(t *testing.T) {
	type args struct {
		list []int
		e    int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "testTrue",
			args: args{
				list: []int{1, 2, 3},
				e:    3,
			},
			want: true,
		},
		{
			name: "testFalse",
			args: args{
				list: []int{1, 2, 3},
				e:    4,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := containsInt(tt.args.list, tt.args.e); got != tt.want {
				t.Errorf("containsInt() = %v, want %v", got, tt.want)
			}
		})
	}
}
