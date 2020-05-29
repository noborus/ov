package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/noborus/ov/internal/oviewer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Version represents the version
	Version string = "dev"
	// Revision set "git rev-parse --short HEAD"
	Revision string = "HEAD"
)

var cfgFile string

// Ver is version information.
var Ver bool

// Config structure contains the variables that determine the overall behavior.
type Config struct {
	// Wrap is Wrap mode.
	Wrap bool
	// TabWidth is tab stop num.
	TabWidth int
	// HeaderLen is number of header rows to be fixed.
	Header int
	// AfterWrite writes the current screen on exit.
	AfterWrite bool
	// QuiteSmall Quit if the output fits on one screen.
	QuitSmall bool
	// CaseSensitive is case-sensitive if true
	CaseSensitive bool
	// Color to alternate rows
	AlternateRows bool
	// Column mode
	ColumnMode bool
	// Column Delimiter
	ColumnDelimiter string
	// Line Number
	LineNumMode bool
	// Debug is enable debug display.
	Debug bool

	// Alternating background color.
	ColorAlternate string

	// Header color.
	ColorHeader string

	// OverStrike color.
	ColorOverStrike string

	// OverLine color.
	ColorOverLine string
}

var config Config

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "ov",
	Short: "ov is a feature rich pager",
	Long: `ov is a feature rich pager(such as more/less).
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

		if err := os.Setenv("TCELL_TRUECOLOR", "disable"); err != nil {
			return err
		}

		root := oviewer.New()
		root.Header = config.Header
		root.TabWidth = config.TabWidth
		root.WrapMode = config.Wrap
		root.AfterWrite = config.AfterWrite
		root.QuitSmall = config.QuitSmall
		root.CaseSensitive = config.CaseSensitive
		root.AlternateRows = config.AlternateRows
		root.ColumnMode = config.ColumnMode
		root.LineNumMode = config.LineNumMode
		root.ColumnDelimiter = strings.ReplaceAll(config.ColumnDelimiter, "\\t", "\t")

		if config.ColorAlternate != "" {
			oviewer.ColorAlternate = tcell.GetColor(config.ColorAlternate)
		}
		if config.ColorHeader != "" {
			oviewer.HeaderStyle = oviewer.HeaderStyle.Foreground(tcell.GetColor(config.ColorHeader))
		}
		if config.ColorOverStrike != "" {
			oviewer.OverStrikeStyle = oviewer.OverStrikeStyle.Foreground(tcell.GetColor(config.ColorOverStrike))
		}
		if config.ColorOverLine != "" {
			oviewer.OverLineStyle = oviewer.OverLineStyle.Foreground(tcell.GetColor(config.ColorOverLine))
		}

		if err := root.Run(args); err != nil {
			return err
		}
		if root.AfterWrite {
			root.WriteOriginal()
		}
		return nil
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.oviewer.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&Ver, "version", "v", false, "display version information")

	rootCmd.PersistentFlags().BoolVarP(&config.Wrap, "wrap", "w", true, "wrap mode")
	_ = viper.BindPFlag("Wrap", rootCmd.PersistentFlags().Lookup("wrap"))

	rootCmd.PersistentFlags().IntVarP(&config.TabWidth, "tab-width", "x", 8, "tab stop width")
	_ = viper.BindPFlag("TabWidth", rootCmd.PersistentFlags().Lookup("tab-width"))

	rootCmd.PersistentFlags().IntVarP(&config.Header, "header", "H", 0, "number of header rows to fix")
	_ = viper.BindPFlag("Header", rootCmd.PersistentFlags().Lookup("header"))

	rootCmd.PersistentFlags().BoolVarP(&config.AfterWrite, "exit-write", "X", false, "output the current screen when exiting")
	_ = viper.BindPFlag("ExitWrite", rootCmd.PersistentFlags().Lookup("exit-write"))

	rootCmd.PersistentFlags().BoolVarP(&config.QuitSmall, "quit-if-one-screen", "F", false, "quit if the output fits on one screen")
	_ = viper.BindPFlag("QuitSmall", rootCmd.PersistentFlags().Lookup("quit-if-one-screen"))

	rootCmd.PersistentFlags().BoolVarP(&config.CaseSensitive, "case-sensitive", "i", false, "case-sensitive in search")
	_ = viper.BindPFlag("CaseSensitive", rootCmd.PersistentFlags().Lookup("case-sensitive"))

	rootCmd.PersistentFlags().BoolVarP(&config.AlternateRows, "alternate-rows", "C", false, "color to alternate rows")
	_ = viper.BindPFlag("AlternateRows", rootCmd.PersistentFlags().Lookup("alternate-rows"))

	rootCmd.PersistentFlags().BoolVarP(&config.ColumnMode, "column-mode", "c", false, "column mode")
	_ = viper.BindPFlag("ColumnMode", rootCmd.PersistentFlags().Lookup("column-mode"))

	rootCmd.PersistentFlags().StringVarP(&config.ColumnDelimiter, "column-delimiter", "d", ",", "column delimiter")
	_ = viper.BindPFlag("ColumnDelimiter", rootCmd.PersistentFlags().Lookup("column-delimiter"))

	rootCmd.PersistentFlags().BoolVarP(&config.LineNumMode, "line-number", "n", false, "line number")
	_ = viper.BindPFlag("LineNumMode", rootCmd.PersistentFlags().Lookup("line-number"))

	rootCmd.PersistentFlags().BoolVarP(&config.Debug, "debug", "", false, "debug mode")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".oviewer" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".ov")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	_ = viper.ReadInConfig()

	if err := viper.Unmarshal(&config); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
