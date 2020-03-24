// Copyright © 2020 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/noborus/zpager"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// Ver is version information.
var Ver bool

type Config struct {
	// Wrap is Wrap mode.
	Wrap bool
	// TabWidth is tab stop num.
	TabWidth int
	// HeaderLen is number of header rows to be fixed.
	Header int
	// NoInit writes the current screen on exit.
	NoInit bool
}

var config Config

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "zpager",
	Short: "Pager for various compressed files",
	Long: `Pager(such as more/less) for various compressed files.
You can view files that are compressed in gzip, bzip 2, zstd, lz 4, and xz.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if Ver {
			fmt.Printf("zpager version %s rev:%s\n", Version, Revision)
			return nil
		}
		m := zpager.NewModel()
		m.TabWidth = config.TabWidth
		m.WrapMode = config.Wrap
		m.HeaderLen = config.Header
		m.PostWrite = config.NoInit
		fmt.Printf("%#v\n", config)
		return zpager.Run(m, args)
	},
}

var (
	// Version represents the version
	Version string
	// Revision set "git rev-parse --short HEAD"
	Revision string
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(version string, revision string) {
	Version = version
	Revision = revision
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.zpager.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&Ver, "version", "v", false, "display version information")
	rootCmd.PersistentFlags().BoolVarP(&config.Wrap, "wrap", "w", true, "wrap mode")
	rootCmd.PersistentFlags().IntVarP(&config.TabWidth, "tab-width", "x", 8, "tab stop")
	rootCmd.PersistentFlags().IntVarP(&config.Header, "header", "H", 0, "number of header rows to fix")
	rootCmd.PersistentFlags().BoolVarP(&config.NoInit, "no-init", "X", false, "Output the current screen when exiting")

	viper.BindPFlags(rootCmd.PersistentFlags())
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".zpager" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".zpager")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	if err := viper.Unmarshal(&config); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
