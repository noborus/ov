package main

import (
	"testing"

	"github.com/spf13/viper"
)

// This test is a yaml file check.
func Test_Config(t *testing.T) {
	type args struct {
		cfgFile string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test ov.yaml",
			args: args{
				cfgFile: "ov.yaml",
			},
			wantErr: false,
		},
		{
			name: "test ov-less.yaml",
			args: args{
				cfgFile: "ov-less.yaml",
			},
			wantErr: false,
		},
		{
			name: "test error",
			args: args{
				cfgFile: "config-fail.yaml",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.SetConfigFile(tt.args.cfgFile)
			err := viper.ReadInConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
