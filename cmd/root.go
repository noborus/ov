package cmd

import (
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/noborus/oviewer"
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
	// AfterWrite writes the current screen on exit.
	AfterWrite bool
	// Debug is enable debug display.
	Debug bool
}

var config Config

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ov",
	Short: "Oviewer is a feature rich pager",
	Long: `Oviewer is a feature rich pager(such as more/less).
It supports various compressed files(gzip, bzip2, zstd, lz4, and xz).
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if Ver {
			fmt.Printf("ov version %s rev:%s\n", Version, Revision)
			return nil
		}

		if config.Debug {
			fmt.Println("Using config file:", viper.ConfigFileUsed())
		}
		oviewer.Debug = config.Debug
		err := os.Setenv("TCELL_TRUECOLOR", "disable")
		if err != nil {
			return err
		}

		root := oviewer.New()
		root.Header = config.Header
		root.TabWidth = config.TabWidth
		root.WrapMode = config.Wrap
		root.AfterWrite = config.AfterWrite

		err = root.Run(args)
		if err != nil {
			return err
		}
		if root.AfterWrite {
			root.WriteOriginal()
		}
		return nil
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
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.oviewer.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&Ver, "version", "v", false, "display version information")
	rootCmd.PersistentFlags().BoolVarP(&config.Wrap, "wrap", "w", true, "wrap mode")
	rootCmd.PersistentFlags().IntVarP(&config.TabWidth, "tab-width", "x", 8, "tab stop")
	rootCmd.PersistentFlags().IntVarP(&config.Header, "header", "H", 0, "number of header rows to fix")
	rootCmd.PersistentFlags().BoolVarP(&config.AfterWrite, "exit-write", "X", false, "Output the current screen when exiting")
	rootCmd.PersistentFlags().BoolVarP(&config.Debug, "debug", "", false, "Debug mode")

	viper.BindPFlag("Wrap", rootCmd.PersistentFlags().Lookup("wrap"))
	viper.BindPFlag("TabWidth", rootCmd.PersistentFlags().Lookup("tab-width"))
	viper.BindPFlag("Header", rootCmd.PersistentFlags().Lookup("header"))
	viper.BindPFlag("ExitWrite", rootCmd.PersistentFlags().Lookup("exit-write"))
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

		// Search config in home directory with name ".oviewer" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".oviewer")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	viper.ReadInConfig()

	if err := viper.Unmarshal(&config); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
