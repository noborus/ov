package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/noborus/ov/oviewer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Version represents the version.
	Version = "dev"
	// Revision set "git rev-parse --short HEAD".
	Revision = "HEAD"
)

var (
	cfgFile string

	config oviewer.Config

	// ver is version information.
	ver bool
	// helpKey is key bind information.
	helpKey bool
	// completion is the generation of shell completion.
	completion bool
)

// ErrCompletion indicates that the completion argument was invalid.
var ErrCompletion = errors.New("requires one of the arguments bash/zsh/fish/powershell")

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "ov",
	Short: "ov is a feature rich pager",
	Long: `ov is a feature rich pager(such as more/less).
It supports various compressed files(gzip, bzip2, zstd, lz4, and xz).
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if ver {
			fmt.Printf("ov version %s rev:%s\n", Version, Revision)
			return nil
		}
		if helpKey {
			HelpKey(cmd, args)
			return nil
		}

		if completion {
			return Completion(cmd, args)
		}

		if config.Debug {
			fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}

		ov, err := oviewer.Open(args...)
		if err != nil {
			return err
		}
		ov.SetConfig(config)

		if err := ov.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			return nil
		}

		if ov.AfterWrite {
			ov.WriteOriginal()
		}
		if ov.Debug {
			ov.WriteLog()
		}
		return nil
	},
}

// HelpKey displays key bindings and exits.
func HelpKey(cmd *cobra.Command, _ []string) {
	fmt.Println(cmd.Short)
	keyBind := oviewer.GetKeyBinds(config.Keybind)
	fmt.Println(oviewer.KeyBindString(keyBind))
}

// Completion is shell completion.
func Completion(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return ErrCompletion
	}

	switch args[0] {
	case "bash":
		return cmd.Root().GenBashCompletion(os.Stdout)
	case "zsh":
		return cmd.Root().GenZshCompletion(os.Stdout)
	case "fish":
		return cmd.Root().GenFishCompletion(os.Stdout, true)
	case "powershell":
		return cmd.Root().GenPowerShellCompletion(os.Stdout)
	}

	return ErrCompletion
}

func init() {
	config = oviewer.NewConfig()
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ov.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&ver, "version", "v", false, "display version information")
	rootCmd.PersistentFlags().BoolVarP(&helpKey, "help-key", "", false, "display key bind information")
	rootCmd.PersistentFlags().BoolVarP(&completion, "completion", "", false, "Generate completion script [bash|zsh|fish|powershell]")

	// Config.General
	rootCmd.PersistentFlags().IntP("tab-width", "x", 8, "tab stop width")
	_ = viper.BindPFlag("general.TabWidth", rootCmd.PersistentFlags().Lookup("tab-width"))

	rootCmd.PersistentFlags().IntP("header", "H", 0, "number of header rows to fix")
	_ = viper.BindPFlag("general.Header", rootCmd.PersistentFlags().Lookup("header"))

	rootCmd.PersistentFlags().BoolP("alternate-rows", "C", false, "color to alternate rows")
	_ = viper.BindPFlag("general.AlternateRows", rootCmd.PersistentFlags().Lookup("alternate-rows"))

	rootCmd.PersistentFlags().BoolP("column-mode", "c", false, "column mode")
	_ = viper.BindPFlag("general.ColumnMode", rootCmd.PersistentFlags().Lookup("column-mode"))

	rootCmd.PersistentFlags().BoolP("line-number", "n", false, "line number")
	_ = viper.BindPFlag("general.LineNumMode", rootCmd.PersistentFlags().Lookup("line-number"))

	rootCmd.PersistentFlags().BoolP("wrap", "w", true, "wrap mode")
	_ = viper.BindPFlag("general.Wrap", rootCmd.PersistentFlags().Lookup("wrap"))

	rootCmd.PersistentFlags().StringP("column-delimiter", "d", ",", "column delimiter")
	_ = viper.BindPFlag("general.ColumnDelimiter", rootCmd.PersistentFlags().Lookup("column-delimiter"))

	// Config
	rootCmd.PersistentFlags().BoolP("disable-mouse", "", false, "disable mouse support")
	_ = viper.BindPFlag("DisableMouse", rootCmd.PersistentFlags().Lookup("disable-mouse"))

	rootCmd.PersistentFlags().BoolP("exit-write", "X", false, "output the current screen when exiting")
	_ = viper.BindPFlag("AfterWrite", rootCmd.PersistentFlags().Lookup("exit-write"))

	rootCmd.PersistentFlags().BoolP("quit-if-one-screen", "F", false, "quit if the output fits on one screen")
	_ = viper.BindPFlag("QuitSmall", rootCmd.PersistentFlags().Lookup("quit-if-one-screen"))

	rootCmd.PersistentFlags().BoolP("case-sensitive", "i", false, "case-sensitive in search")
	_ = viper.BindPFlag("CaseSensitive", rootCmd.PersistentFlags().Lookup("case-sensitive"))

	rootCmd.PersistentFlags().BoolP("debug", "", false, "debug mode")
	_ = viper.BindPFlag("Debug", rootCmd.PersistentFlags().Lookup("debug"))
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

		// Search config in home directory with name ".ov" (without extension).
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
