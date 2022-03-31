package oviewer

import (
	"path/filepath"
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
