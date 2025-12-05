package oviewer

import (
	"reflect"
	"testing"
)

func Test_rawConverter_convert(t *testing.T) {
	tests := []struct {
		name    string
		r       rawConverter
		st      *parseState
		str     string
		want    bool
		wantStr string
	}{
		{
			name: "convertABC",
			r:    *newRawConverter(),
			str:  "abc",
			st: &parseState{
				str: "\n",
			},
			want:    false,
			wantStr: "abc",
		},
		{
			name: "convertEscape",
			r:    *newRawConverter(),
			str:  "abc",
			st: &parseState{
				str: "\x1b",
			},
			want:    true,
			wantStr: "abc^[",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.st.lc = StrToContents(tt.str, tt.st.tabWidth)
			if got := tt.r.convert(tt.st); got != tt.want {
				t.Errorf("rawConverter.convert() = %v, want %v", got, tt.want)
			}
			gotStr, _ := ContentsToStr(tt.st.lc)
			if !reflect.DeepEqual(gotStr, tt.wantStr) {
				t.Errorf("rawConverter.convert() = %v, want %v", gotStr, tt.wantStr)
			}
		})
	}
}
