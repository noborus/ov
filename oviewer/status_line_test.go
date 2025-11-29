package oviewer

import (
	"testing"
)

func TestRoot_statusMode(t *testing.T) {
	type fields struct {
		FollowSection bool
		FollowAll     bool
		FollowMode    bool
		FollowName    bool
		WatchMode     bool
		pauseFollow   bool
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "test watch",
			fields: fields{
				WatchMode: true,
			},
			want: "(Watch)",
		},
		{
			name: "test follow section",
			fields: fields{
				FollowSection: true,
			},
			want: "(Follow Section)",
		},
		{
			name: "test follow all",
			fields: fields{
				FollowAll: true,
			},
			want: "(Follow All)",
		},
		{
			name: "test follow name",
			fields: fields{
				FollowMode: true,
				FollowName: true,
			},
			want: "(Follow Name)",
		},
		{
			name: "test follow mode",
			fields: fields{
				FollowMode: true,
			},
			want: "(Follow Mode)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootHelper(t)
			root.Doc.WatchMode = tt.fields.WatchMode
			root.Doc.FollowSection = tt.fields.FollowSection
			root.FollowAll = tt.fields.FollowAll
			root.Doc.FollowName = tt.fields.FollowName
			root.Doc.FollowMode = tt.fields.FollowMode
			root.Doc.pauseFollow = tt.fields.pauseFollow
			if got := root.statusMode(); got != tt.want {
				t.Errorf("Root.statusDisplay() = %v, want %v", got, tt.want)
			}
		})
	}
}
