package oviewer

import (
	"bytes"
	"reflect"
	"testing"
)

func TestDocument_lineToContents(t *testing.T) {
	t.Parallel()
	type args struct {
		lN       int
		tabWidth int
	}
	tests := []struct {
		name    string
		str     string
		args    args
		want    lineContents
		wantErr bool
	}{
		{
			name: "testEmpty",
			str:  ``,
			args: args{
				lN:       0,
				tabWidth: 4,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test1",
			str:  "test\n",
			args: args{
				lN:       0,
				tabWidth: 4,
			},
			want:    StrToContents("test", 4),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewDocument()
			if err != nil {
				t.Fatal(err)
			}
			if err := m.ReadAll(bytes.NewBufferString(tt.str)); err != nil {
				t.Fatal(err)
			}
			<-m.eofCh
			t.Logf("num:%d", m.BufEndNum())
			got, err := m.contentsLN(tt.args.lN, tt.args.tabWidth)
			if (err != nil) != tt.wantErr {
				t.Errorf("Document.lineToContents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Document.lineToContents() = %v, want %v", got, tt.want)
			}
		})
	}
}
