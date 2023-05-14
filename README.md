# ov - feature rich terminal pager

[![PkgGoDev](https://pkg.go.dev/badge/github.com/noborus/ov)](https://pkg.go.dev/github.com/noborus/ov)
[![Actions Status](https://github.com/noborus/ov/workflows/Go/badge.svg)](https://github.com/noborus/ov/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/noborus/ov)](https://goreportcard.com/report/github.com/noborus/ov)

ov is a terminal pager.

* `ov` can be used instead of `less` or `more` or `tail -f`.
* `ov` also has an effective function for tabular text.

![ov1.png](https://raw.githubusercontent.com/noborus/ov/master/docs/ov1.png)

<!-- vscode-markdown-toc -->
* 1. [Feature](#feature)
  * 1.1. [Not supported](#not-supported)
  * 1.2. [TODO(Implemented in v0.20.0)](#todo(implemented-in-v0.20.0))
* 2. [Install](#install)
  * 2.1. [deb package](#deb-package)
  * 2.2. [rpm package](#rpm-package)
  * 2.3. [MacPorts (macOS)](#macports-(macos))
  * 2.4. [Homebrew(macOS or Linux)](#homebrew(macos-or-linux))
  * 2.5. [pkg (FreeBSD)](#pkg-(freebsd))
  * 2.6. [Arch Linux](#arch-linux)
  * 2.7. [nix (nixOS, Linux, or macOS)](#nix-(nixos,-linux,-or-macos))
  * 2.8. [Binary](#binary)
  * 2.9. [go install](#go-install)
  * 2.10. [go get(details or developer version)](#go-get(details-or-developer-version))
* 3. [Usage](#usage)
  * 3.1. [Basic usage](#basic-usage)
  * 3.2. [Config](#config)
  * 3.3. [Header](#header)
    * 3.3.1. [Skip](#skip)
  * 3.4. [Column mode](#column-mode)
  * 3.5. [Column rainbow mode](#column-rainbow-mode)
  * 3.6. [Wrap/NoWrap](#wrap/nowrap)
  * 3.7. [Alternate-Rows](#alternate-rows)
  * 3.8. [Section](#section)
  * 3.9. [Follow mode](#follow-mode)
  * 3.10. [Follow all mode](#follow-all-mode)
  * 3.11. [Follow section mode](#follow-section-mode)
  * 3.12. [Exec mode](#exec-mode)
  * 3.13. [Search](#search)
  * 3.14. [Mark](#mark)
  * 3.15. [Watch](#watch)
  * 3.16. [Mouse support](#mouse-support)
  * 3.17. [Multi color highlight](#multi-color-highlight)
  * 3.18. [Plain](#plain)
  * 3.19. [Jump target](#jump-target)
  * 3.20. [View mode](#view-mode)
  * 3.21. [Output on exit](#output-on-exit)
* 4. [How to reduce memory usage](#how-to-reduce-memory-usage)
  * 4.1. [Regular file (seekable)](#regular-file-(seekable))
  * 4.2. [Other files, pipes(Non-seekable)](#other-files,-pipes(non-seekable))
* 5. [Command option](#command-option)
* 6. [Key bindings](#key-bindings)
* 7. [Customize](#customize)
  * 7.1. [Style customization](#style-customization)
  * 7.2. [Key binding customization](#key-binding-customization)

<!-- vscode-markdown-toc-config
	numbering=true
	autoSave=true
	/vscode-markdown-toc-config -->
<!-- /vscode-markdown-toc -->

##  1. <a name='feature'></a>Feature

* Supports fixed [header](#header) line display (both wrap/nowrap).
* Supports [column mode](#column-mode), which recognizes columns by delimiter.
* Also, in column mode, there is a [column-rainbow](#column-rainbow-mode) mode that colors each column.
* Supports section-by-section movement, splitting [sections](#section) by delimiter.
* Dynamic [wrap/nowrap](#wrap/nowrap) switchable.
* Supports alternating row styling.
* Shortcut keys are [customizable](#key-binding-customization).
* The style of the effect is [customizable](#style-customization).
* Supports [follow-mode](#follow-mode) (like tail -f).
* Supports [follow-section](#follow-section-mode), which is displayed when the section is updated.
* Supports following multiple files and switching when updated([follow-all](#follow-all-mode)).
* Supports the [execution](#exec-mode) of commands that toggle both stdout and stder for display.
* Supports [watch](#watch) mode, which reads files on a regular basis.
* Supports incremental [search](#search) and regular expression search.
* Supports [multi-color](#multi-color-highlight) to highlight multiple words individually.
* Better support for Unicode and East Asian Width.
* Supports compressed files (gzip, bzip2, zstd, lz4, xz).
* Suitable for tabular text. [psql](https://noborus.github.io/ov/psql), [mysql](https://noborus.github.io/ov/mysql/), [csv](https://noborus.github.io/ov/csv/), [etc...](https://noborus.github.io/ov/)
  
###  1.1. <a name='not-supported'></a>Not supported

* Does not support syntax highlighting for file types (source code, markdown, etc.)
* Does not support Filter function (`&pattern` equivalent of `less`)

###  1.2. <a name='todo(implemented-in-v0.20.0)'></a>TODO(Implemented in v0.20.0)

* Allow opening files larger than memory
* Enable follow by file name (equivalent to `tail -F`)
* Support columns with fixed widths instead of delimiters
* Support watch in exec mode (equivalent to `watch` command)

##  2. <a name='install'></a>Install

###  2.1. <a name='deb-package'></a>deb package

You can download the package from [releases](https://github.com/noborus/ov/releases).

```console
curl -L -O https://github.com/noborus/ov/releases/download/vx.x.x/ov_x.x.x-1_amd64.deb
sudo dpkg -i ov_x.x.x-1_amd64.deb
```

###  2.2. <a name='rpm-package'></a>rpm package

You can download the package from [releases](https://github.com/noborus/ov/releases).

```console
sudo rpm -ivh https://github.com/noborus/ov/releases/download/vx.x.x/ov_x.x.x-1_amd64.rpm
```

###  2.3. <a name='macports-(macos)'></a>MacPorts (macOS)

```console
sudo port install ov
```

###  2.4. <a name='homebrew(macos-or-linux)'></a>Homebrew(macOS or Linux)

```console
brew install noborus/tap/ov
```

###  2.5. <a name='pkg-(freebsd)'></a>pkg (FreeBSD)

```console
pkg install ov
```

###  2.6. <a name='arch-linux'></a>Arch Linux

You can install ov using an [AUR helper](https://wiki.archlinux.org/title/AUR_helpers).

AUR package: [https://aur.archlinux.org/packages/ov-bin](https://aur.archlinux.org/packages/ov-bin)

###  2.7. <a name='nix-(nixos,-linux,-or-macos)'></a>nix (nixOS, Linux, or macOS)

ov is available as a nix package. You can install it with

```console
nix profile install nixpkgs#ov
```

if you use flakes, or using nix-env otherwise:

```console
nix-env -iA nixpkgs.ov
```

###  2.8. <a name='binary'></a>Binary

You can download the binary from [releases](https://github.com/noborus/ov/releases).

```console
curl -L -O https://github.com/noborus/ov/releases/download/vx.x.x/ov_x.x.x_linux_amd64.zip
unzip ov_x.x.x_linux_amd64.zip
sudo install ov /usr/local/bin
```

###  2.9. <a name='go-install'></a>go install

It will be installed in $GOPATH/bin by the following command.

```console
go install github.com/noborus/ov@latest
```

###  2.10. <a name='go-get(details-or-developer-version)'></a>go get(details or developer version)

First of all, download only with the following command without installing it.

```console
go get -d github.com/noborus/ov
cd $GOPATH/src/github.com/noborus/ov
```

Next, to install to $GOPATH/bin, run the make install command.

```console
make install
```

Or, install it in a PATH location for other users to use
(For example, in /usr/local/bin).

```console
make
sudo install ov /usr/local/bin
```

##  3. <a name='usage'></a>Usage

(default key `key`) indicates the key that can be specified even after starting the same function as the command line option.

###  3.1. <a name='basic-usage'></a>Basic usage

ov supports open file name or standard input.

```console
ov filename
```

```console
cat filename|ov
```

Used by other commands by setting the environment variable **PAGER**.

```console
export PAGER=ov
```

See the [ov site](https://noborus.github.io/ov/) for more use cases.

###  3.2. <a name='config'></a>Config

You can set style and key bindings in the configuration file.

ov will look for a configuration file in the following paths in descending order:

```filepath
$XDG_CONFIG_HOME/ov/config.yaml
$HOME/.config/ov/config.yaml
$HOME/.ov.yaml
```

On Windows:

```filepath
%USERPROFILE%/.config/ov/config.yaml
%USERPROFILE%/.ov.yaml
```

Create a `config.yaml` file in one of the above directories. If the file is in the user home directory, it should be named `.ov.yaml`.

Please refer to the sample [ov.yaml](https://raw.githubusercontent.com/noborus/ov/master/ov.yaml) configuration file.

> **Note**
>
> If you like `less` key bindings, copy  [ov-less.yaml](https://raw.githubusercontent.com/noborus/ov/master/ov-less.yaml) and use it.

###  3.3. <a name='header'></a>Header

The `--header` (`-H`) (default key `H`) option fixedly displays the specified number of lines.

```console
ov --header 1 README.md
```

####  3.3.1. <a name='skip'></a>Skip

When used with the `--skip-lines` (default key `ctrl+s`) option, it hides the number of lines specified by skip and then displays the header.

```console
ov --skip-lines 1 --header 1 README.md
```

###  3.4. <a name='column-mode'></a>Column mode

Specify the delimiter with `--column-delimiter`(default key is `d`) and set it to `--column-mode`(default key is `c`) to highlight the column.

```console
ov --column-delimiter "," --column-mode test.csv
```

Regular expressions can be used for the `--column-delimiter`.
Enclose in '/' when using regular expressions.

```console
ps aux | ov -H1 --column-delimiter "/\s+/" --column-rainbow --column-mode
```

###  3.5. <a name='column-rainbow-mode'></a>Column rainbow mode

You can also color each column individually in column mode.
Specify `--column-rainbow`(default key is `ctrl+r`) in addition to the `--column-mode` option.

Color customization is possible. Please specify 7 or more colors in `config.yaml`.

```yaml
StyleColumnRainbow:
  - Foreground: "white"
  - Foreground: "aqua"
  - Foreground: "lightsalmon"
  - Foreground: "lime"
  - Foreground: "blue"
  - Foreground: "yellowgreen"
  - Foreground: "red"
```

###  3.6. <a name='wrap/nowrap'></a>Wrap/NoWrap

Supports switching between wrapping and not wrapping lines.

The option is `--wrap`, specify `--wrap=false` (default key `w`, `W`) if you do not want to wrap.

###  3.7. <a name='alternate-rows'></a>Alternate-Rows

Alternate row styles with the `--alternate-rows`(`-C`) (default key `C`) option
The style can be set with [Style customization](#style-customization).

```console
ov --alternate-rows test.csv
```

###  3.8. <a name='section'></a>Section

You specify `--section-delimiter`(default key `alt+d`), you can move up and down in section units.
The start of the section can be adjusted with `--section-start`(default key `ctrl+F3`, `alt+s`).

![section.png](docs/section.png)

The section-delimiter is written in a regular expression (for example: "^#").
(Line breaks are not included in matching lines).

For example, if you specify "^diff" for a diff that contains multiple files,
you can move the diff for each file.

###  3.9. <a name='follow-mode'></a>Follow mode

`--follow`(`-f`)(default key `ctrl+f`) prints appended data and moves to the bottom line (like `tail -f`).

```console
ov --follow-mode /var/log/syslog
```

```console
(while :; do echo random-$RANDOM; sleep 0.1; done;)|./ov  --follow-mode
```

###  3.10. <a name='follow-all-mode'></a>Follow all mode

`--follow-all`(`-A`)(default key `ctrl+a`) is the same as follow mode, it switches to the last updated file if there are multiple files.

```console
ov --follow-all /var/log/nginx/access.log /var/log/nginx/error.log
```

###  3.11. <a name='follow-section-mode'></a>Follow section mode

Use the `--follow-section`(default key `F2`) option to follow by section.
Follow mode is line-by-line, while follow section mode is section-by-section.
Follow section mode displays the bottom section.
The following example is displayed from the header (#) at the bottom.

```console
ov --section-delimiter "^#" --follow-section README.md
```

> **Note**
>
> [Watch](#watch) mode is a mode in which `--follow-section` and `--section-delimiter "^\f"` are automatically set.

###  3.12. <a name='exec-mode'></a>Exec mode

Use the `--exec` (`-e`) option to run the command and display stdout/stderr separately.
Arguments after (`--`) are interpreted as command arguments.

Shows the stderr screen as soon as an error occurs, when used with `--follow-all`.

```console
ov --follow-all --exec -- make
```

###  3.13. <a name='search'></a>Search

Search by forward search `/` key(default) or the backward search `?` key(defualt).
Search can be toggled between incremental search, regular expression search, and case sensitivity.
Displayed when the following are enabled in the search input prompt:

| Function | display | (Default)key |command option |
|:---------|:--------|:----|:--------------|
| Incremental search | (I) | alt+i | --incremental |
| Regular expression search | (R) | alt+r | --regexp-search  |
| Case sensitive | (Aa) | alt+c |  -i, --case-sensitive |

###  3.14. <a name='mark'></a>Mark

Mark the display position with the `m` key(default).
The mark is decorated with `StyleMarkLine` and `MarkStyleWidth`.

Marks can be erased individually with the `M` key(default).
It is also possible to delete all marks with the `ctrl + delete` key(default).

Use the `>`next and `<`previous (default) key to move to the marked position.

###  3.15. <a name='watch'></a>Watch

`ov` has a watch mode that reads the file every N seconds and adds it to the end.
When you reach EOF, add '\f' instead.
Use the `--watch`(`-T`) option.
Go further to the last section.
The default is'section-delimiter', so the last loaded content is displayed.

for example.

```console
ov --watch 1 /proc/meminfo
```

###  3.16. <a name='mouse-support'></a>Mouse support

The ov makes the mouse support its control.
This can be disabled with the option `--disable-mouse`(default key `ctrl+F3`, `contrl+alt+r`).

If mouse support is enabled, tabs and line breaks will be interpreted correctly when copying.

Copying to the clipboard uses [atotto/clipboard](https://github.com/atotto/clipboard).
For this reason, the 'xclip' or 'xsel' command is required in Linux/Unix environments.

Selecting the range with the mouse and then left-clicking will copy it to the clipboard.

Pasting in ov is done with the middle button.
In other applications, it is pasted from the clipboard (often by pressing the right-click).

Also, if mouse support is enabled, horizontal scrolling is possible with `shift+wheel`.

###  3.17. <a name='multi-color-highlight'></a>Multi color highlight

This feature styles multiple words individually.
`.`key(defualt) enters multi-word input mode.
Enter multiple words (regular expressions) separated by spaces.

For example, `error info warn debug` will color errors red, info cyan, warn yellow, and debug magenta.

It can also be specified with the command line option `--multi-color`(`-M`)(default key `.`).
For command line options, pass them separated by ,(comma).

For example:

```console
ov --multi-color "ERROR,WARN,INFO,DEBUG,not,^.{24}" access.log
```

![multi-color.png](https://raw.githubusercontent.com/noborus/ov/master/docs/multi-color.png)

Color customization is possible. Please specify 7 or more colors in config.yaml.

```yaml
StyleMultiColorHighlight:
  - Foreground: "red"
    Reverse: true
  - Foreground: "aqua"
  - Foreground: "yellow"
  - Foreground: "fuchsia"
  - Foreground: "lime"
  - Foreground: "blue"
  - Foreground: "grey"
```

###  3.18. <a name='plain'></a>Plain

Supports undecorating ANSI escape sequences.
The option is `--plain` (or `-p`) (default key `ctrl+e`).

###  3.19. <a name='jump-target'></a>Jump target

You can specify the lines to be displayed in the search results.
This function is similar to `--jump-target` of `less`.
Positive numbers are displayed downwards by the number of lines from the top(1).
Negative numbers are displayed up by the number of lines from the bottom(-1).
. (dot) can be used to specify a percentage. .5 is the middle of the screen(.5).
You can also specify a percentage, such as (50%).

This option can be specified with `--jump-target`(or `-j`) (default key `j`).

###  3.20. <a name='view-mode'></a>View mode

You can also use a combination of modes using the `--view-mode`(default key `p`) option.
In that case, you can set it in advance and specify the combined mode at once.

For example, if you write the following settings in ov.yaml,
 the csv mode will be set with `--view-mode csv`.

```console
ov --view-mode csv test.csv
```

```ov.yaml
Mode:
  p:
    Header: 2
    AlternateRows: true
    ColumnMode: true
    LineNumMode: false
    WrapMode: true
    ColumnDelimiter: "|"
    ColumnRainbow: true
  m:
    Header: 3
    AlternateRows: true
    ColumnMode: true
    LineNumMode: false
    WrapMode: true
    ColumnDelimiter: "|"
  csv:
    Header: 1
    AlternateRows: true
    ColumnMode: true
    LineNumMode: false
    WrapMode: true
    ColumnDelimiter: ","
    ColumnRainbow: true
```

###  3.21. <a name='output-on-exit'></a>Output on exit

`--exit-write` `-X`(default key `Q`) option prints the current screen on exit.
This looks like the display remains on the console after the ov is over.

By default, it outputs the amount of the displayed screen and exits.

```console
ov -X README.md
```

You can change how much is written using `--exit-write-before` and `--exit-write-after`(default key `ctrl+q`).
`--exit-write-before`

`--exit-write-before` specifies the number of lines before the current position(top of screen).
`--exit-write-before 3` will output from 3 lines before.

`--exit-write-after` specifies the number of lines after the current position (top of screen).

`--exit-write-before 3 --exit-write-after 3` outputs 6 lines.

##  4. <a name='how-to-reduce-memory-usage'></a>How to reduce memory usage

Since **v0.22.0** it no longer loads everything into memory.
The first chunk from the beginning to the 10,000th line is loaded into memory
and never freed.
Therefore, files with less than 10,000 lines do not change behavior.

The `--memory-limit` option can be used to limit the chunks loaded into memory.
Memory limits vary by file type.

Also, go may use a lot of memory until the memory is freed by GC.
Also consider setting the environment variable `GOMEMLIMIT`.

```console
export GOMEMLIMIT=100MiB
```

###  4.1. <a name='regular-file-(seekable)'></a>Regular file (seekable)

Normally large (10,000+ lines) files are loaded in chunks when needed. It also frees chunks that are no longer needed.
If `--memory-limit` is not specified, it will be limited to 100.

```console
ov --memory-limit-file 3 /var/log/syslog
```

Specify `MemoryLimit` in the configuration file.

```yaml
MemoryLimitFile: 3
```

###  4.2. <a name='other-files,-pipes(non-seekable)'></a>Other files, pipes(Non-seekable)

Non-seekable files and pipes cannot be read again, so they must exist in memory.

If you specify the upper limit of chunks with `--memory-limit` or `MemoryLimit`,
it will read up to the upper limit first, but after that,
when the displayed position advances, the old chunks will be released.
Unlimited if `--memory-limit` is not specified.

```console
cat /var/log/syslog | ov --memory-limit 10
```

It is recommended to put a limit in the config file as you may receive output larger than memory.

```yaml
MemoryLimit: 1000
```

You can also use the `--memory-limit-file` option and the `MemoryLimitFile` setting for those who think regular files are good memory saving.

##  5. <a name='command-option'></a>Command option

```console
$ ov --help
ov is a feature rich pager(such as more/less).
It supports various compressed files(gzip, bzip2, zstd, lz4, and xz).

Usage:
  ov [flags]

Flags:
  -C, --alternate-rows             alternately change the line color
  -i, --case-sensitive             case-sensitive in search
  -d, --column-delimiter string    column delimiter (default ",")
  -c, --column-mode                column mode
      --column-rainbow             column rainbow
      --column-width               column width mode
      --completion string          generate completion script [bash|zsh|fish|powershell]
      --config string              config file (default is $XDG_CONFIG_HOME/ov/config.yaml)
      --debug                      debug mode
      --disable-mouse              disable mouse support
  -e, --exec                       exec command
  -X, --exit-write                 output the current screen when exiting
  -a, --exit-write-after int       NUM after the current lines when exiting
  -b, --exit-write-before int      NUM before the current lines when exiting
  -A, --follow-all                 follow all
  -f, --follow-mode                follow mode
      --follow-name                follow name mode
      --follow-section             follow section
  -H, --header int                 number of header rows to fix
  -h, --help                       help for ov
      --help-key                   display key bind information
      --incsearch                  incremental search (default true)
  -j, --jump-target string         jump-target
  -n, --line-number                line number mode
      --memory-limit int           Number of chunks to limit in memory (default -1)                       v0.22.0
      --memory-limit-file int      The number of chunks to limit in memory for the file (default 100)     v0.22.0
  -M, --multi-color strings        multi-color
  -p, --plain                      disable original decoration
  -F, --quit-if-one-screen         quit if the output fits on one screen
      --regexp-search              regular expression search
      --section-delimiter string   section delimiter
      --section-start int          section start position
      --skip-lines int             skip the number of lines
  -x, --tab-width int              tab stop width (default 8)
  -v, --version                    display version information
      --view-mode string           view mode
  -T, --watch int                  watch mode interval
  -w, --wrap                       wrap mode (default true)
```

It can also be changed after startup.

##  6. <a name='key-bindings'></a>Key bindings

```console
 [Escape], [q]                * quit
 [ctrl+c]                     * cancel
 [Q]                          * output screen and quit
 [ctrl+q]                     * set output screen and quit
 [ctrl+z]                     * suspend
 [h], [ctrl+alt+c], [ctrl+f1] * display help screen
 [ctrl+f2], [ctrl+alt+e]      * display log screen
 [ctrl+l]                     * screen sync
 [ctrl+f]                     * follow mode toggle
 [ctrl+a]                     * follow all mode toggle
 [ctrl+f3], [ctrl+alt+r]      * enable/disable mouse

	Moving

 [Enter], [Down], [ctrl+N]    * forward by one line
 [Up], [ctrl+p]               * backward by one line
 [Home]                       * go to top of document
 [End]                        * go to end of document
 [PageDown], [ctrl+v]         * forward by page
 [PageUp], [ctrl+b]           * backward by page
 [ctrl+d]                     * forward a half page
 [ctrl+u]                     * backward a half page
 [left]                       * scroll to left
 [right]                      * scroll to right
 [ctrl+left]                  * scroll left half screen
 [ctrl+right]                 * scroll right half screen
 [shift+Home]                 * go to beginning of line
 [shift+End]                  * go to end of line
 [g]                          * go to line(input number and `.n` and `n%` allowed)

	Move document

 []]                          * next document
 [[]                          * previous document
 [ctrl+k]                     * close current document

	Mark position

 [m]                          * mark current position
 [M]                          * remove mark current position
 [ctrl+delete]                * remove all mark
 [>]                          * move to next marked position
 [<]                          * move to previous marked position

	Search

 [/]                          * forward search mode
 [?]                          * backward search mode
 [n]                          * repeat forward search
 [N]                          * repeat backward search

	Change display

 [w], [W]                     * wrap/nowrap toggle
 [c]                          * column mode toggle
 [alt+o]                      * column width toggle                               v0.20.0
 [ctrl+r]                     * column rainbow toggle
 [C]                          * alternate rows of style toggle
 [G]                          * line number toggle
 [ctrl+e]                     * original decoration toggle(plain)

	Change Display with Input

 [p], [P]                     * view mode selection
 [d]                          * column delimiter string
 [H]                          * number of header lines
 [ctrl+s]                     * number of skip lines
 [t]                          * TAB width
 [.]                          * multi color highlight
 [j]                          * jump target(`.n` and `n%` allowed)

	Section

 [alt+d]                      * section delimiter regular expression
 [ctrl+F3], [alt+s]           * section start position
 [space], [ctrl+down]         * next section
 [^], [ctrl+up]               * previous section
 [9]                          * last section
 [F2]                         * follow section mode toggle

	Close and reload

 [ctrl+F9], [ctrl+alt+s]      * close file
 [ctrl+alt+l], [F5]           * reload file
 [ctrl+alt+w], [F4]           * watch mode
 [ctrl+w]                     * set watch interval

	Key binding when typing

 [alt+c]                      * case-sensitive toggle
 [alt+r]                      * regular expression search toggle
 [alt+i]                      * incremental search toggle
 [Up]                         * previous candidate
 [Down]                       * next candidate
 [ctrl+c]                     * copy to clipboard.
 [ctrl+v]                     * paste from clipboard
```

##  7. <a name='customize'></a>Customize

###  7.1. <a name='style-customization'></a>Style customization

You can customize the following items.

* StyleAlternate
* StyleHeader
* StyleOverStrike
* StyleOverLine
* StyleLineNumber
* StyleSearchHighlight
* StyleColumnHighlight
* StyleMarkLine
* StyleSectionLine
* StyleMultiColorHighlight
* StyleColumnRainbow
* StyleJumpTargetLine

Specifies the color name for the foreground and background [colors](https://pkg.go.dev/github.com/gdamore/tcell/v2#pkg-constants).
Specify bool values for Reverse, Bold, Blink, Dim, Italic, and Underline.

[Example]

```yaml
StyleAlternate:
  Background: "gray"
  Bold: true
  Underline: true
```

| item name | value | example |
|:----------|:------|:--------|
| Foreground | "color name" or "rgb" | "red" |
| Background | "color name" or "rgb" | "#2a2a2a" |
| Reverse | true/false | true |
| Bold | true/false | true |
| Blink | true/false | true |
| Dim | true/false | false |
| Italic | true/false | false |
| Underline | true/false | false |

###  7.2. <a name='key-binding-customization'></a>Key binding customization

You can customize key bindings.

[Example]

```yaml
    down:
        - "Enter"
        - "Down"
        - "ctrl+N"
    up:
        - "Up"
        - "ctrl+p"
```

See [ov.yaml](https://github.com/noborus/ov/blob/master/ov.yaml) for more information.
