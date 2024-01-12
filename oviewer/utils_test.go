package oviewer

import (
	"reflect"
	"regexp"
	"testing"
)

func Test_max(t *testing.T) {
	t.Parallel()
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := max(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("max() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_min(t *testing.T) {
	t.Parallel()
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := min(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("min() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_removeStr(t *testing.T) {
	t.Parallel()
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := remove(tt.args.list, tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("remove() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_removeInt(t *testing.T) {
	t.Parallel()
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := remove(tt.args.list, tt.args.c); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("remove() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_containsInt(t *testing.T) {
	t.Parallel()
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := contains(tt.args.list, tt.args.e); got != tt.want {
				t.Errorf("containsInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_toAddTop(t *testing.T) {
	t.Parallel()
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
				s:    "f",
			},
			want: []string{"f", "a", "b", "c"},
		},
		{
			name: "test2",
			args: args{
				list: []string{},
				s:    "f",
			},
			want: []string{"f"},
		},
		{
			name: "testNoAdd",
			args: args{
				list: []string{"a", "b", "c"},
				s:    "",
			},
			want: []string{"a", "b", "c"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := toAddTop(tt.args.list, tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toTop() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_toAddLast(t *testing.T) {
	t.Parallel()
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
				list: []string{
					"a",
				},
				s: "b",
			},
			want: []string{
				"a",
				"b",
			},
		},
		{
			name: "test2",
			args: args{
				list: []string{
					"a",
					"b",
				},
				s: "a",
			},
			want: []string{
				"a",
				"b",
			},
		},
		{
			name: "testNoAdd",
			args: args{
				list: []string{
					"a",
					"b",
				},
				s: "",
			},
			want: []string{
				"a",
				"b",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := toAddLast(tt.args.list, tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toAddLast() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_toLast(t *testing.T) {
	t.Parallel()
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
				s:    "a",
			},
			want: []string{"b", "c", "a"},
		},
		{
			name: "testAdd",
			args: args{
				list: []string{"a", "b", "c"},
				s:    "x",
			},
			want: []string{"a", "b", "c", "x"},
		},
		{
			name: "testNoAdd",
			args: args{
				list: []string{"a", "b", "c"},
				s:    "",
			},
			want: []string{"a", "b", "c"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := toLast(tt.args.list, tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toLast() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_remove(t *testing.T) {
	t.Parallel()
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
				list: []string{
					"a",
				},
				s: "b",
			},
			want: []string{
				"a",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := remove(tt.args.list, tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("remove() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_allIndex(t *testing.T) {
	t.Parallel()
	type args struct {
		s      string
		substr string
		reg    *regexp.Regexp
	}
	tests := []struct {
		name string
		args args
		want [][]int
	}{
		{
			name: "test1",
			args: args{
				s:      "a,b,c",
				substr: ",",
				reg:    nil,
			},
			want: [][]int{
				{1, 2},
				{3, 4},
			},
		},
		{
			name: "test2",
			args: args{
				s:      "a|b|c",
				substr: "|",
				reg:    nil,
			},
			want: [][]int{
				{1, 2},
				{3, 4},
			},
		},
		{
			name: "testTab",
			args: args{
				s:      "a	b	c",
				substr: "	",
				reg:    nil,
			},
			want: [][]int{
				{1, 2},
				{3, 4},
			},
		},
		{
			name: "testUnicode",
			args: args{
				s:      "a│b│c",
				substr: "│",
				reg:    nil,
			},
			want: [][]int{
				{1, 2},
				{5, 6},
			},
		},
		{
			name: "testRegex",
			args: args{
				s:      "a  b c",
				substr: `/\s+/`,
				reg:    regexp.MustCompile(`\s+`),
			},
			want: [][]int{
				{1, 3},
				{4, 5},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := allIndex(tt.args.s, tt.args.substr, tt.args.reg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("allIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_allStringIndex(t *testing.T) {
	t.Parallel()
	type args struct {
		s      string
		substr string
	}
	tests := []struct {
		name string
		args args
		want [][]int
	}{
		{
			name: "test1",
			args: args{
				s:      "a,b,c",
				substr: ",",
			},
			want: [][]int{
				{1, 2},
				{3, 4},
			},
		},
		{
			name: "testNone",
			args: args{
				s:      "a,b,c",
				substr: "@",
			},
			want: nil,
		},
		{
			name: "testNoSubstr",
			args: args{
				s:      "a,b,c",
				substr: "",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := allStringIndex(tt.args.s, tt.args.substr); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("allIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}
