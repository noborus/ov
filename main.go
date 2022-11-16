package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/noborus/ov/oviewer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

var (
	// Version represents the version.
	Version = "dev"
	// Revision set "git rev-parse --short HEAD".
	Revision = "HEAD"
)

var (
	// cfgFile is the configuration file name.
	cfgFile string
	// config is the oviewer setting.
	config oviewer.Config

	// ver is version information.
	ver bool
	// helpKey is key bind information.
	helpKey bool
	// completion is the generation of shell completion.
	completion string
	// execCommand targets the output of executing the command.
	execCommand bool
)

var (
	// ErrCompletion indicates that the completion argument was invalid.
	ErrCompletion = errors.New("requires one of the arguments bash/zsh/fish/powershell")
	// ErrNoArgument indicating that there are no arguments to execute.
	ErrNoArgument = errors.New("no arguments to execute")
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "ov",
	Short: "ov is a feature rich pager",
	Long: `ov is a feature rich pager(such as more/less).
It supports various compressed files(gzip, bzip2, zstd, lz4, and xz).
`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if config.Debug {
			cfg := viper.ConfigFileUsed()
			if cfg != "" {
				fmt.Fprintln(os.Stderr, "Using config file:", cfg)
			} else {
				fmt.Fprintln(os.Stderr, "config file not found")
			}
		}

		if ver {
			fmt.Printf("ov version %s rev:%s\n", Version, Revision)
			return nil
		}
		if helpKey {
			HelpKey(cmd, args)
			return nil
		}

		if completion != "" {
			return Completion(cmd, completion)
		}
		// Actually tabs when "\t" is specified as an option.
		if config.General.ColumnDelimiter == "\\t" {
			config.General.ColumnDelimiter = "\t"
		}

		// Set a global variable to convert to a style before opening the file.
		oviewer.OverStrikeStyle = oviewer.ToTcellStyle(config.StyleOverStrike)
		oviewer.OverLineStyle = oviewer.ToTcellStyle(config.StyleOverLine)

		SetRedirect()

		if execCommand {
			return ExecCommand(args)
		}
		return RunOviewer(args)
	},
}

// HelpKey displays key bindings and exits.
func HelpKey(cmd *cobra.Command, _ []string) {
	fmt.Println(cmd.Short)
	keyBind := oviewer.GetKeyBinds(config)
	fmt.Println(oviewer.KeyBindString(keyBind))
}

// Completion is shell completion.
func Completion(cmd *cobra.Command, arg string) error {
	switch arg {
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

// RunOviewer displays the argument file.
func RunOviewer(args []string) error {
	ov, err := oviewer.Open(args...)
	if err != nil {
		return err
	}

	ov.SetConfig(config)

	if err := ov.Run(); err != nil {
		return err
	}

	if ov.IsWriteOriginal {
		ov.WriteOriginal()
	}
	if ov.Debug {
		ov.WriteLog()
	}
	return nil
}

// ExecCommand targets the output of command execution (stdout/stderr).
func ExecCommand(args []string) error {
	if len(args) == 0 {
		return ErrNoArgument
	}

	command := exec.Command(args[0], args[1:]...)
	ov, err := oviewer.ExecCommand(command)
	if err != nil {
		return err
	}

	defer func() {
		if command == nil || command.Process == nil {
			return
		}
		if err := command.Process.Kill(); err != nil {
			log.Println(err)
		}
		if err := command.Wait(); err != nil {
			log.Println(err)
		}
	}()

	ov.SetConfig(config)

	if err := ov.Run(); err != nil {
		return err
	}

	if ov.IsWriteOriginal {
		ov.WriteOriginal()
	}
	if ov.Debug {
		ov.WriteLog()
	}
	return nil
}

// SetRedirect saves stdout and stderr during execution.
func SetRedirect() {
	// Suppress the output of os.Stdout and os.Stderr
	// because the screen collapses.
	if term.IsTerminal(int(os.Stdout.Fd())) {
		tmpStdout := os.Stdout
		os.Stdout = nil
		defer func() {
			os.Stdout = tmpStdout
		}()
	} else {
		oviewer.STDOUTPIPE = os.Stdout
	}

	if term.IsTerminal(int(os.Stderr.Fd())) {
		tmpStderr := os.Stderr
		os.Stderr = nil
		defer func() {
			os.Stderr = tmpStderr
		}()
	} else {
		oviewer.STDERRPIPE = os.Stderr
	}
}

func init() {
	config = oviewer.NewConfig()
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ov.yaml)")
	_ = rootCmd.RegisterFlagCompletionFunc("config", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"yaml"}, cobra.ShellCompDirectiveFilterFileExt
	})
	rootCmd.PersistentFlags().BoolVarP(&ver, "version", "v", false, "display version information")
	rootCmd.PersistentFlags().BoolVarP(&helpKey, "help-key", "", false, "display key bind information")
	rootCmd.PersistentFlags().BoolVarP(&execCommand, "exec", "e", false, "exec command")

	rootCmd.PersistentFlags().StringVarP(&completion, "completion", "", "", "generate completion script [bash|zsh|fish|powershell]")
	_ = rootCmd.RegisterFlagCompletionFunc("completion", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"bash", "zsh", "fish", "powershell"}, cobra.ShellCompDirectiveNoFileComp
	})

	// Config.General
	rootCmd.PersistentFlags().IntP("tab-width", "x", 8, "tab stop width")
	_ = viper.BindPFlag("general.TabWidth", rootCmd.PersistentFlags().Lookup("tab-width"))
	_ = rootCmd.RegisterFlagCompletionFunc("tab-width", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"3\ttab width", "2\ttab width", "4\ttab width", "8\ttab width"}, cobra.ShellCompDirectiveNoFileComp
	})

	rootCmd.PersistentFlags().IntP("header", "H", 0, "number of header rows to fix")
	_ = viper.BindPFlag("general.Header", rootCmd.PersistentFlags().Lookup("header"))
	_ = rootCmd.RegisterFlagCompletionFunc("header", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"1"}, cobra.ShellCompDirectiveNoFileComp
	})

	rootCmd.PersistentFlags().IntP("skip-lines", "", 0, "skip the number of lines")
	_ = viper.BindPFlag("general.SkipLines", rootCmd.PersistentFlags().Lookup("skip-lines"))
	_ = rootCmd.RegisterFlagCompletionFunc("skip-lines", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"1"}, cobra.ShellCompDirectiveNoFileComp
	})

	rootCmd.PersistentFlags().BoolP("alternate-rows", "C", false, "alternately change the line color")
	_ = viper.BindPFlag("general.AlternateRows", rootCmd.PersistentFlags().Lookup("alternate-rows"))

	rootCmd.PersistentFlags().BoolP("column-mode", "c", false, "column mode")
	_ = viper.BindPFlag("general.ColumnMode", rootCmd.PersistentFlags().Lookup("column-mode"))

	rootCmd.PersistentFlags().BoolP("line-number", "n", false, "line number mode")
	_ = viper.BindPFlag("general.LineNumMode", rootCmd.PersistentFlags().Lookup("line-number"))

	rootCmd.PersistentFlags().BoolP("wrap", "w", true, "wrap mode")
	_ = viper.BindPFlag("general.WrapMode", rootCmd.PersistentFlags().Lookup("wrap"))

	rootCmd.PersistentFlags().BoolP("plain", "p", false, "disable original decoration")
	_ = viper.BindPFlag("general.PlainMode", rootCmd.PersistentFlags().Lookup("plain"))

	rootCmd.PersistentFlags().StringP("column-delimiter", "d", ",", "column delimiter")
	_ = viper.BindPFlag("general.ColumnDelimiter", rootCmd.PersistentFlags().Lookup("column-delimiter"))
	_ = rootCmd.RegisterFlagCompletionFunc("column-delimiter", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{",\tcomma", "|\tvertical line", "\\\\t\ttab", "│\tbox"}, cobra.ShellCompDirectiveNoFileComp
	})

	rootCmd.PersistentFlags().StringP("section-delimiter", "", "", "section delimiter")
	_ = viper.BindPFlag("general.SectionDelimiter", rootCmd.PersistentFlags().Lookup("section-delimiter"))

	rootCmd.PersistentFlags().IntP("section-start", "", 0, "section start position")
	_ = viper.BindPFlag("general.SectionStartPosition", rootCmd.PersistentFlags().Lookup("section-start"))

	rootCmd.PersistentFlags().BoolP("follow-mode", "f", false, "follow mode")
	_ = viper.BindPFlag("general.FollowMode", rootCmd.PersistentFlags().Lookup("follow-mode"))

	rootCmd.PersistentFlags().BoolP("follow-all", "A", false, "follow all")
	_ = viper.BindPFlag("general.FollowAll", rootCmd.PersistentFlags().Lookup("follow-all"))

	rootCmd.PersistentFlags().BoolP("follow-section", "", false, "follow section")
	_ = viper.BindPFlag("general.FollowSection", rootCmd.PersistentFlags().Lookup("follow-section"))

	rootCmd.PersistentFlags().IntP("watch", "T", 0, "watch mode interval")
	_ = viper.BindPFlag("general.WatchInterval", rootCmd.PersistentFlags().Lookup("watch"))

	// Config
	rootCmd.PersistentFlags().BoolP("disable-mouse", "", false, "disable mouse support")
	_ = viper.BindPFlag("DisableMouse", rootCmd.PersistentFlags().Lookup("disable-mouse"))

	rootCmd.PersistentFlags().BoolP("exit-write", "X", false, "output the current screen when exiting")
	_ = viper.BindPFlag("IsWriteOriginal", rootCmd.PersistentFlags().Lookup("exit-write"))

	rootCmd.PersistentFlags().IntP("exit-write-before", "b", 0, "NUM before the current lines when exiting")
	_ = viper.BindPFlag("BeforeWriteOriginal", rootCmd.PersistentFlags().Lookup("exit-write-before"))

	rootCmd.PersistentFlags().IntP("exit-write-after", "a", 0, "NUM after the current lines when exiting")
	_ = viper.BindPFlag("AfterWriteOriginal", rootCmd.PersistentFlags().Lookup("exit-write-after"))

	rootCmd.PersistentFlags().BoolP("quit-if-one-screen", "F", false, "quit if the output fits on one screen")
	_ = viper.BindPFlag("QuitSmall", rootCmd.PersistentFlags().Lookup("quit-if-one-screen"))

	rootCmd.PersistentFlags().BoolP("case-sensitive", "i", false, "case-sensitive in search")
	_ = viper.BindPFlag("CaseSensitive", rootCmd.PersistentFlags().Lookup("case-sensitive"))

	rootCmd.PersistentFlags().BoolP("regexp-search", "", false, "regular expression search")
	_ = viper.BindPFlag("RegexpSearch", rootCmd.PersistentFlags().Lookup("regexp-search"))

	rootCmd.PersistentFlags().BoolP("incsearch", "", true, "incremental search")
	_ = viper.BindPFlag("Incsearch", rootCmd.PersistentFlags().Lookup("incsearch"))

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
			fmt.Fprintln(os.Stderr, err)
			return
		}

		// Search config in home directory with name ".ov" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".ov")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	_ = viper.ReadInConfig()
	if err := viper.Unmarshal(&config); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
