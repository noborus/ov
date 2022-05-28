package main

import (
	"testing"

	"github.com/noborus/ov/oviewer"
	"github.com/spf13/viper"
)

var tconfig oviewer.Config

// This test is a yaml file check.
func Test_Config(t *testing.T) {
	type args struct {
		cfgFile string
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		wantErr2 bool
	}{
		{
			name: "test ov.yaml",
			args: args{
				cfgFile: "ov.yaml",
			},
			wantErr:  false,
			wantErr2: false,
		},
		{
			name: "test ov-less.yaml",
			args: args{
				cfgFile: "ov-less.yaml",
			},
			wantErr:  false,
			wantErr2: false,
		},
		{
			name: "test error",
			args: args{
				cfgFile: "config-fail.yaml",
			},
			wantErr:  true,
			wantErr2: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.SetConfigFile(tt.args.cfgFile)
			if err := viper.ReadInConfig(); (err != nil) != tt.wantErr {
				t.Errorf("Config error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err := viper.Unmarshal(&tconfig); (err != nil) != tt.wantErr2 {
				t.Errorf("Config error = %v, wantErr2 %v", err, tt.wantErr2)
				return
			}
		})
	}
}
