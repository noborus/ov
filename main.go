package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/noborus/ov/oviewer"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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

	// pattern is search pattern.
	pattern string

	// filter is filter pattern.
	filter string

	// non match filter pattern.
	nonMatchFilter string

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
		// Set the caption from the environment variable.
		if config.General.Caption == "" {
			config.General.Caption = viper.GetString("CAPTION")
		}

		// Actually tabs when "\t" is specified as an option.
		if config.General.ColumnDelimiter == "\\t" {
			config.General.ColumnDelimiter = "\t"
		}

		// SectionHeader is enabled if SectionHeaderNum is greater than 0.
		if config.General.SectionHeaderNum > 0 {
			config.General.SectionHeader = true
		}

		// Set a global variable to convert to a style before opening the file.
		oviewer.OverStrikeStyle = oviewer.ToTcellStyle(config.StyleOverStrike)
		oviewer.OverLineStyle = oviewer.ToTcellStyle(config.StyleOverLine)
		oviewer.MemoryLimit = config.MemoryLimit
		oviewer.MemoryLimitFile = config.MemoryLimitFile
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

// argsToFiles returns the file name from the argument.
func argsToFiles(args []string) []string {
	var files []string
	for _, arg := range args {
		argFiles, err := filepath.Glob(arg)
		if err != nil {
			continue
		}
		files = append(files, argFiles...)
	}
	// If filePath.Glob does not match,
	// return argument to return correct error
	if len(files) == 0 {
		return args
	}
	return files
}

// RunOviewer displays the argument file.
func RunOviewer(args []string) error {
	files := argsToFiles(args)
	ov, err := oviewer.Open(files...)
	if err != nil {
		return err
	}

	ov.SetConfig(config)

	if pattern != "" {
		ov.Search(pattern)
	}
	if filter != "" {
		ov.Filter(filter, false)
	}
	if nonMatchFilter != "" {
		ov.Filter(nonMatchFilter, true)
	}

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
	cmd := oviewer.NewCommand(args...)
	ov, err := cmd.Exec()
	if err != nil {
		return err
	}

	defer func() {
		cmd.Wait()
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

// flagUsage returns the usage of the flags.
func flagUsage(f *pflag.FlagSet) string {
	buf := new(bytes.Buffer)

	lines := make([]string, 0)

	lines = append(lines, "Short   Long\x00Purpose")

	maxlen := 0
	f.VisitAll(func(flag *pflag.Flag) {
		if flag.Hidden {
			return
		}

		line := ""
		if flag.Shorthand != "" && flag.ShorthandDeprecated == "" {
			line = fmt.Sprintf("  -%s, --%s", flag.Shorthand, flag.Name)
		} else {
			line = fmt.Sprintf("      --%s", flag.Name)
		}

		varname, usage := pflag.UnquoteUsage(flag)
		if varname != "" {
			line += " " + varname
		}

		switch flag.Value.Type() {
		case "string":
		case "bool":
			if flag.DefValue == "true" {
				line += "[=true|false]"
			}
		case "count":
		default:
		}

		// This special character will be replaced with spacing once the
		// correct alignment is calculated
		line += "\x00"
		if len(line) > maxlen {
			maxlen = len(line)
		}

		line += usage

		switch flag.Value.Type() {
		case "int", "int8", "int16", "int32", "int64":
			if flag.DefValue != "0" {
				line += fmt.Sprintf(" (default %s)", flag.DefValue)
			}
		case "string":
			if flag.DefValue != "" {
				line += fmt.Sprintf(" (default %q)", flag.DefValue)
			}
		case "bool":
			if flag.DefValue == "true" {
				line += " (default true)"
			}
		case "count":
		default:
		}
		if len(flag.Deprecated) != 0 {
			line += fmt.Sprintf(" (DEPRECATED: %s)", flag.Deprecated)
		}

		lines = append(lines, line)
	})

	for _, line := range lines {
		sidx := strings.Index(line, "\x00")
		spacing := strings.Repeat(" ", maxlen-sidx)
		fmt.Fprintln(buf, line[:sidx], spacing, line[sidx+1:])
	}

	return buf.String()
}

// UsageFunc returns a function that returns the usage.
// This is for overwriting cobra usage.
func UsageFunc() (cmd func(*cobra.Command) error) {
	return func(c *cobra.Command) error {
		fmt.Fprintf(os.Stdout, "Usage:\n")
		fmt.Fprintf(os.Stdout, "  %s [flags] [FILE]...\n", c.CommandPath())
		fmt.Fprintf(os.Stdout, "\n")
		fmt.Fprintf(os.Stdout, "Flags:\n")
		fmt.Print(flagUsage(c.Flags()))
		return nil
	}
}

func init() {
	config = oviewer.NewConfig()
	cobra.OnInitialize(initConfig)
	rootCmd.SetUsageFunc(UsageFunc())
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config `file` (default is $XDG_CONFIG_HOME/ov/config.yaml)")
	_ = rootCmd.RegisterFlagCompletionFunc("config", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"yaml"}, cobra.ShellCompDirectiveFilterFileExt
	})
	rootCmd.PersistentFlags().BoolVarP(&ver, "version", "v", false, "display version information")
	rootCmd.PersistentFlags().BoolVarP(&helpKey, "help-key", "", false, "display key bind information")
	rootCmd.PersistentFlags().BoolVarP(&execCommand, "exec", "e", false, "command execution result instead of file")

	rootCmd.PersistentFlags().StringVarP(&completion, "completion", "", "", "generate completion script [bash|zsh|fish|powershell]")
	_ = rootCmd.RegisterFlagCompletionFunc("completion", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"bash", "zsh", "fish", "powershell"}, cobra.ShellCompDirectiveNoFileComp
	})

	rootCmd.PersistentFlags().StringVarP(&pattern, "pattern", "", "", "search pattern")
	rootCmd.PersistentFlags().StringVarP(&filter, "filter", "", "", "filter search pattern")
	rootCmd.PersistentFlags().StringVarP(&nonMatchFilter, "non-match-filter", "", "", "filter non match search pattern")
	rootCmd.PersistentFlags().BoolVarP(&oviewer.SkipExtract, "skip-extract", "", false, "skip extracting compressed files")

	// Config.General
	rootCmd.PersistentFlags().IntP("tab-width", "x", 8, "tab stop width")
	_ = viper.BindPFlag("general.TabWidth", rootCmd.PersistentFlags().Lookup("tab-width"))
	_ = rootCmd.RegisterFlagCompletionFunc("tab-width", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"3\ttab width", "2\ttab width", "4\ttab width", "8\ttab width"}, cobra.ShellCompDirectiveNoFileComp
	})

	rootCmd.PersistentFlags().IntP("header", "H", 0, "number of header lines to be displayed constantly")
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

	rootCmd.PersistentFlags().BoolP("column-width", "", false, "column mode for width")
	_ = viper.BindPFlag("general.ColumnWidth", rootCmd.PersistentFlags().Lookup("column-width"))

	rootCmd.PersistentFlags().BoolP("column-rainbow", "", false, "column mode to rainbow")
	_ = viper.BindPFlag("general.ColumnRainbow", rootCmd.PersistentFlags().Lookup("column-rainbow"))

	rootCmd.PersistentFlags().BoolP("line-number", "n", false, "line number mode")
	_ = viper.BindPFlag("general.LineNumMode", rootCmd.PersistentFlags().Lookup("line-number"))

	rootCmd.PersistentFlags().BoolP("wrap", "w", true, "wrap mode")
	_ = viper.BindPFlag("general.WrapMode", rootCmd.PersistentFlags().Lookup("wrap"))

	rootCmd.PersistentFlags().BoolP("plain", "p", false, "disable original decoration")
	_ = viper.BindPFlag("general.PlainMode", rootCmd.PersistentFlags().Lookup("plain"))

	rootCmd.PersistentFlags().StringP("column-delimiter", "d", ",", "column delimiter `character`")
	_ = viper.BindPFlag("general.ColumnDelimiter", rootCmd.PersistentFlags().Lookup("column-delimiter"))
	_ = rootCmd.RegisterFlagCompletionFunc("column-delimiter", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{",\tcomma", "|\tvertical line", "\\\\t\ttab", "│\tbox"}, cobra.ShellCompDirectiveNoFileComp
	})

	rootCmd.PersistentFlags().StringP("section-delimiter", "", "", "`regexp` for section delimiter .e.g. \"^#\"")
	_ = viper.BindPFlag("general.SectionDelimiter", rootCmd.PersistentFlags().Lookup("section-delimiter"))

	rootCmd.PersistentFlags().IntP("section-start", "", 0, "section start position")
	_ = viper.BindPFlag("general.SectionStartPosition", rootCmd.PersistentFlags().Lookup("section-start"))

	rootCmd.PersistentFlags().BoolP("section-header", "", false, "enable section-delimiter line as Header")
	_ = viper.BindPFlag("general.SectionHeader", rootCmd.PersistentFlags().Lookup("section-header"))

	rootCmd.PersistentFlags().IntP("section-header-num", "", 1, "number of header lines")
	_ = viper.BindPFlag("general.SectionHeaderNum", rootCmd.PersistentFlags().Lookup("section-header-num"))

	rootCmd.PersistentFlags().BoolP("follow-mode", "f", false, "monitor file and display new content as it is written")
	_ = viper.BindPFlag("general.FollowMode", rootCmd.PersistentFlags().Lookup("follow-mode"))

	rootCmd.PersistentFlags().BoolP("follow-all", "A", false, "follow multiple files and show the most recently updated one")
	_ = viper.BindPFlag("general.FollowAll", rootCmd.PersistentFlags().Lookup("follow-all"))

	rootCmd.PersistentFlags().BoolP("follow-section", "", false, "section-by-section follow mode")
	_ = viper.BindPFlag("general.FollowSection", rootCmd.PersistentFlags().Lookup("follow-section"))

	rootCmd.PersistentFlags().BoolP("follow-name", "", false, "follow mode to monitor by file name")
	_ = viper.BindPFlag("general.FollowName", rootCmd.PersistentFlags().Lookup("follow-name"))

	rootCmd.PersistentFlags().IntP("watch", "T", 0, "watch mode interval(`seconds`)")
	_ = viper.BindPFlag("general.WatchInterval", rootCmd.PersistentFlags().Lookup("watch"))

	rootCmd.PersistentFlags().StringSliceP("multi-color", "M", nil, "comma separated words(regexp) to color .e.g. \"ERROR,WARNING\"")
	_ = viper.BindPFlag("general.MultiColorWords", rootCmd.PersistentFlags().Lookup("multi-color"))

	rootCmd.PersistentFlags().StringP("jump-target", "j", "", "jump target `[int|int%|.int|'section']`")
	_ = viper.BindPFlag("general.JumpTarget", rootCmd.PersistentFlags().Lookup("jump-target"))

	rootCmd.PersistentFlags().StringP("hscroll-width", "", "10%", "width to scroll horizontally `[int|int%|.int]`")
	_ = viper.BindPFlag("general.HScrollWidth", rootCmd.PersistentFlags().Lookup("hscroll-width"))

	rootCmd.PersistentFlags().StringP("caption", "", "", "custom caption")
	_ = viper.BindPFlag("general.Caption", rootCmd.PersistentFlags().Lookup("caption"))

	rootCmd.PersistentFlags().BoolP("hide-other-section", "", false, "hide other section")
	_ = viper.BindPFlag("general.HideOtherSection", rootCmd.PersistentFlags().Lookup("hide-other-section"))

	// Config
	rootCmd.PersistentFlags().BoolP("quit-if-one-screen", "F", false, "quit if the output fits on one screen")
	_ = viper.BindPFlag("QuitSmall", rootCmd.PersistentFlags().Lookup("quit-if-one-screen"))

	rootCmd.PersistentFlags().BoolP("exit-write", "X", false, "output the current screen when exiting")
	_ = viper.BindPFlag("IsWriteOriginal", rootCmd.PersistentFlags().Lookup("exit-write"))

	rootCmd.PersistentFlags().IntP("exit-write-before", "b", 0, "number before the current lines when exiting")
	_ = viper.BindPFlag("BeforeWriteOriginal", rootCmd.PersistentFlags().Lookup("exit-write-before"))

	rootCmd.PersistentFlags().IntP("exit-write-after", "a", 0, "number after the current lines when exiting")
	_ = viper.BindPFlag("AfterWriteOriginal", rootCmd.PersistentFlags().Lookup("exit-write-after"))

	rootCmd.PersistentFlags().BoolP("case-sensitive", "i", false, "case-sensitive in search")
	_ = viper.BindPFlag("CaseSensitive", rootCmd.PersistentFlags().Lookup("case-sensitive"))

	rootCmd.PersistentFlags().BoolP("smart-case-sensitive", "", false, "smart case-sensitive in search")
	_ = viper.BindPFlag("SmartCaseSensitive", rootCmd.PersistentFlags().Lookup("smart-case-sensitive"))

	rootCmd.PersistentFlags().BoolP("regexp-search", "", false, "regular expression search")
	_ = viper.BindPFlag("RegexpSearch", rootCmd.PersistentFlags().Lookup("regexp-search"))

	rootCmd.PersistentFlags().BoolP("incsearch", "", true, "incremental search")
	_ = viper.BindPFlag("Incsearch", rootCmd.PersistentFlags().Lookup("incsearch"))

	rootCmd.PersistentFlags().IntP("memory-limit", "", -1, "number of chunks to limit in memory")
	_ = viper.BindPFlag("MemoryLimit", rootCmd.PersistentFlags().Lookup("memory-limit"))

	rootCmd.PersistentFlags().IntP("memory-limit-file", "", 100, "number of chunks to limit in memory for the file")
	_ = viper.BindPFlag("MemoryLimitFile", rootCmd.PersistentFlags().Lookup("memory-limit-file"))

	rootCmd.PersistentFlags().BoolP("disable-mouse", "", false, "disable mouse support")
	_ = viper.BindPFlag("DisableMouse", rootCmd.PersistentFlags().Lookup("disable-mouse"))

	rootCmd.PersistentFlags().BoolP("disable-column-cycle", "", false, "disable column cycling")
	_ = viper.BindPFlag("DisableColumnCycle", rootCmd.PersistentFlags().Lookup("disable-column-cycle"))

	rootCmd.PersistentFlags().StringP("view-mode", "", "", "apply predefined settings for a specific mode")
	_ = viper.BindPFlag("ViewMode", rootCmd.PersistentFlags().Lookup("view-mode"))

	rootCmd.PersistentFlags().BoolP("debug", "", false, "debug mode")
	_ = viper.BindPFlag("Debug", rootCmd.PersistentFlags().Lookup("debug"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Get home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}

		// Get XDG config directory.
		xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")

		var defaultConfigPath string

		if xdgConfigHome != "" {
			// If set, use it for the default configuration path.
			defaultConfigPath = filepath.Join(xdgConfigHome, "ov")
		} else {
			// If not set, use the default `$HOME/.config/ov`.
			defaultConfigPath = filepath.Join(home, ".config", "ov")
		}

		// Set the default configuration path and file name.
		viper.AddConfigPath(defaultConfigPath)
		viper.SetConfigName("config")

		// If the default config file does not exist but the legacy config file does exist,
		// then fallback to the legacy config path and file name.
		if !fileExists(filepath.Join(defaultConfigPath, "config.yaml")) &&
			fileExists(filepath.Join(home, ".ov.yaml")) {
			viper.AddConfigPath(home)
			viper.SetConfigName(".ov")
		}
	}

	viper.SetEnvPrefix("ov")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		var configNotFoundError *viper.ConfigFileNotFoundError
		if !errors.As(err, &configNotFoundError) {
			fmt.Fprintln(os.Stderr, "failed to read config file:", err)
			// If the config file is not found, it is not an error and will continue.
		}
	}

	if err := viper.Unmarshal(&config); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
}

// fileExists returns true if the file exists.
func fileExists(path string) bool {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
