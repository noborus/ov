# ov - feature rich terminal pager

[![PkgGoDev](https://pkg.go.dev/badge/github.com/noborus/ov)](https://pkg.go.dev/github.com/noborus/ov)
[![Actions Status](https://github.com/noborus/ov/workflows/Go/badge.svg)](https://github.com/noborus/ov/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/noborus/ov)](https://goreportcard.com/report/github.com/noborus/ov)

ov is a terminal pager.

* Can be used instead of `less` or `more`.  It can also be used instead of `tail -f`.
* `ov` also has an effective function for tabular text.

![ov1.png](https://raw.githubusercontent.com/noborus/ov/master/docs/ov1.png)

<!-- vscode-markdown-toc -->
* 1. [feature](#feature)
* 2. [install](#install)
	* 2.1. [deb package](#debpackage)
	* 2.2. [rpm package](#rpmpackage)
	* 2.3. [MacPorts (macOS)](#MacPortsmacOS)
	* 2.4. [Homebrew(macOS or Linux)](#HomebrewmacOSorLinux)
	* 2.5. [pkg (FreeBSD)](#pkgFreeBSD)
	* 2.6. [binary](#binary)
	* 2.7. [go install](#goinstall)
	* 2.8. [go get(details or developer version)](#gogetdetailsordeveloperversion)
* 3. [Usage](#Usage)
	* 3.1. [basic usage](#basicusage)
	* 3.2. [follow mode](#followmode)
	* 3.3. [follow all mode](#followallmode)
	* 3.4. [exec mode](#execmode)
	* 3.5. [Search](#Search)
	* 3.6. [Mark](#Mark)
	* 3.7. [Mouse support](#Mousesupport)
* 4. [command option](#commandoption)
* 5. [Key bindings](#Keybindings)
* 6. [config](#config)
	* 6.1. [psql](#psql)
	* 6.2. [mysql](#mysql)
* 7. [Customize](#Customize)
	* 7.1. [Style customization](#Stylecustomization)
	* 7.2. [Key binding customization](#Keybindingcustomization)

<!-- vscode-markdown-toc-config
	numbering=true
	autoSave=true
	/vscode-markdown-toc-config -->
<!-- /vscode-markdown-toc -->

##  1. <a name='feature'></a>feature

* Better support for Unicode and East Asian Width.
* Support for compressed files (gzip, bzip2, zstd, lz4, xz).
* Columns support column mode that can be selected by delimiter.
* The header row can always be displayed.
* Dynamic wrap/nowrap switchable.
* Supports alternating row style changes.
* Shortcut keys are customizable.
* The style of the effect is customizable.
* Supports follow-mode (like tail -f).
* Supports following multiple files and switching when updated.
* Supports the execution of commands that toggle both stdout and stder for display.
* Supports incremental search and regular expression search.

##  2. <a name='install'></a>install

###  2.1. <a name='debpackage'></a>deb package

You can download the package from [releases](https://github.com/noborus/ov/releases).

```console
curl -L -O https://github.com/noborus/ov/releases/download/vx.x.x/ov_x.x.x-1_amd64.deb
sudo dpkg -i ov_x.x.x-1_amd64.deb
```

###  2.2. <a name='rpmpackage'></a>rpm package

You can download the package from [releases](https://github.com/noborus/ov/releases).

```console
sudo rpm -ivh https://github.com/noborus/ov/releases/download/vx.x.x/ov_x.x.x-1_amd64.rpm
```

###  2.3. <a name='MacPortsmacOS'></a>MacPorts (macOS)

```console
sudo port install ov
```

###  2.4. <a name='HomebrewmacOSorLinux'></a>Homebrew(macOS or Linux)

```console
brew install noborus/tap/ov
```

###  2.5. <a name='pkgFreeBSD'></a>pkg (FreeBSD)

```console
pkg install ov
```

###  2.6. <a name='binary'></a>binary

You can download the binary from [releases](https://github.com/noborus/ov/releases).

```console
curl -L -O https://github.com/noborus/ov/releases/download/vx.x.x/ov_x.x.x_linux_amd64.zip
unzip ov_x.x.x_linux_amd64.zip
sudo install ov /usr/local/bin
```

###  2.7. <a name='goinstall'></a>go install

It will be installed in $GOPATH/bin by the following command.

```console
go install github.com/noborus/ov@latest
```

###  2.8. <a name='gogetdetailsordeveloperversion'></a>go get(details or developer version)

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

##  3. <a name='Usage'></a>Usage

###  3.1. <a name='basicusage'></a>basic usage

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

###  3.2. <a name='followmode'></a>follow mode

Output appended data and move it to the bottom line (like `tail -f`).

```console
ov --follow-mode /var/log/syslog
```

```console
(while :; do echo random-$RANDOM; sleep 0.1; done;)|./ov  --follow-mode
```

![ov-tail.gif](https://raw.githubusercontent.com/noborus/ov/master/docs/ov-tail.gif)

###  3.3. <a name='followallmode'></a>follow all mode

Same as follow-mode, and switches to the last updated file when there are multiple files.

```console
ov --follow-all /var/log/nginx/access.log /var/log/nginx/error.log
```

###  3.4. <a name='execmode'></a>exec mode

Execute the command to display stdout / stderr.
Arguments after (`--`) are interpreted as command arguments.

```console
ov --follow-all --exec -- make
```

![ov-exec.gif](https://raw.githubusercontent.com/noborus/ov/master/docs/ov-exec.gif)


###  3.5. <a name='Search'></a>Search

Search by forward search `/` key(default) or the backward search `?` key(defualt).
Search can be toggled between incremental search, regular expression search, and case sensitivity.
Displayed when the following are enabled in the search input prompt:

| Function | display | (Default)key |command option |
|:---------|:--------|:----|:--------------|
| Incremental search | (I) | alt+i | --incremental |
| Regular expression search | (R) | alt+r | --regexp-search  |
| Case sensitive | (Aa) | alt+c |  -i, --case-sensitive |

###  3.6. <a name='Mark'></a>Mark

Mark the display position with the `m` key(default).
The mark is decorated with `StyleMarkLine` and `MarkStyleWidth`.

Marks can be erased individually with the `M` key(default).
It is also possible to delete all marks with the `ctrl + delete` key(default).

Use the `>`next and `<`previous (default) key to move to the marked position.

###  3.7. <a name='Mousesupport'></a>Mouse support

The ov makes the mouse support its control.
This can be disabled with the option `--disable-mouse`.

If mouse support is enabled, tabs and line breaks will be interpreted correctly when copying.

Copying to the clipboard uses [atotto/clipboard](https://github.com/atotto/clipboard).
For this reason, the 'xclip' or 'xsel' command is required in Linux/Unix environments.

Selecting the range with the mouse and then left-clicking will copy it to the clipboard.

Pasting in ov is done with the middle button.
In other applications, it is pasted from the clipboard (often by pressing the right-click).

##  4. <a name='commandoption'></a>command option

```console
$ ov --help
ov is a feature rich pager(such as more/less).
It supports various compressed files(gzip, bzip2, zstd, lz4, and xz).

Usage:
  ov [flags]

Flags:
  -C, --alternate-rows            alternately change the line color
  -i, --case-sensitive            case-sensitive in search
  -d, --column-delimiter string   column delimiter (default ",")
  -c, --column-mode               column mode
      --completion                generate completion script [bash|zsh|fish|powershell]
      --config string             config file (default is $HOME/.ov.yaml)
      --debug                     debug mode
      --disable-mouse             disable mouse support
  -e, --exec                      exec command
  -X, --exit-write                output the current screen when exiting
  -A, --follow-all                follow all
  -f, --follow-mode               follow mode
  -H, --header int                number of header rows to fix
  -h, --help                      help for ov
      --help-key                  display key bind information
      --incsearch                 incremental search (default true)
  -n, --line-number               line number mode
  -F, --quit-if-one-screen        quit if the output fits on one screen
      --regexp-search             regular expression search
      --skip-lines int            skip the number of lines
  -x, --tab-width int             tab stop width (default 8)
  -v, --version                   display version information
  -w, --wrap                      wrap mode (default true)
```

It can also be changed after startup.
Refer to the [motion image](docs/image.md).

##  5. <a name='Keybindings'></a>Key bindings

```console
 [Escape], [q]                * quit
 [ctrl+c]                     * cancel
 [Q]                          * output screen and quit
 [h], [ctrl+f1], [ctrl+alt+c] * display help screen
 [ctrl+f2], [ctrl+alt+e]      * display log screen
 [ctrl+l]                     * screen sync
 [ctrl+f]                     * follow mode toggle
 [ctrl+a]                     * follow all mode toggle
 [ctrl+alt+r]                 * enable/disable mouse

	Moving

 [Enter], [Down], [ctrl+N]    * forward by one line
 [Up], [ctrl+p]               * backward by one line
 [Home]                       * go to begin of line
 [End]                        * go to end of line
 [PageDown], [ctrl+v]         * forward by page
 [PageUp], [ctrl+b]           * backward by page
 [ctrl+d]                     * forward a half page
 [ctrl+u]                     * backward a half page
 [left]                       * scroll to left
 [right]                      * scroll to right
 [ctrl+left]                  * scroll left half screen
 [ctrl+right]                 * scroll right half screen
 [g]                          * number of go to line

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
 [C]                          * color to alternate rows toggle
 [G]                          * line number toggle

	Change Display with Input

 [p], [P]                     * view mode selection
 [d]                          * delimiter string
 [H]                          * number of header lines
 [ctrl+s]                     * number of skip lines
 [t]                          * TAB width

	Key binding when typing

 [alt+c]                      * case-sensitive toggle
 [alt+r]                      * regular expression search toggle
 [alt+i]                      * incremental search toggle
```

##  6. <a name='config'></a>config

You can set style and key bindings in the setting file.

Please refer to the sample [ov.yaml](https://github.com/noborus/ov/blob/master/ov.yaml) configuration file.

###  6.1. <a name='psql'></a>psql

Set environment variable `PSQL_PAGER`(PostgreSQL 11 or later).

```console
export PSQL_PAGER='ov -w=f -H2 -F -C -d "|"'
```

You can also write in `~/.psqlrc` in previous versions.

```text
\setenv PAGER 'ov -w=f -H2 -F -C -d "|"'
```

###  6.2. <a name='mysql'></a>mysql

Use the --pager option with the mysql client.

```console
mysql --pager='ov -w=f -H3 -F -C -d "|"'
```

You can also write in `~/.my.cnf`.

```ini
[client]
pager=ov -w=f --skip-lines 1 -H1 -F -C -d "|"
```

##  7. <a name='Customize'></a>Customize

###  7.1. <a name='Stylecustomization'></a>Style customization

You can customize the following items.

* StyleAlternate
* StyleHeader
* StyleOverStrike
* StyleOverLine
* StyleLineNumber
* StyleSearchHighlight
* StyleColumnHighlight
* StyleMarkLine

Specifies the color name for the foreground and background [colors](https://pkg.go.dev/github.com/gdamore/tcell/v2#pkg-constants).
Specify bool values for Bold, Blink, Shaded, Italic, and Underline.

[Example]

```yaml
StyleAlternate:
  Background: "gray"
  Bold: true
  Underline: true
```

###  7.2. <a name='Keybindingcustomization'></a>Key binding customization

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

See [ov.yaml](https://github.com/noborus/ov/blob/master/ov.yaml) for more information..
