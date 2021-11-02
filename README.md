# ov - Oviewer

[![PkgGoDev](https://pkg.go.dev/badge/github.com/noborus/ov)](https://pkg.go.dev/github.com/noborus/ov)
[![Actions Status](https://github.com/noborus/ov/workflows/Go/badge.svg)](https://github.com/noborus/ov/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/noborus/ov)](https://goreportcard.com/report/github.com/noborus/ov)

ov is a feature rich terminal pager.
It has an effective function for tabular text.

(The old repository name was oviewer.)

![ov.gif](https://raw.githubusercontent.com/noborus/ov/master/docs/ov.gif)

## feature

* Better support for Unicode and East Asian Width.
* Support for compressed files (gzip, bzip2, zstd, lz4, xz).
* Columns support column mode that can be selected by delimiter.
* Header rows can be fixed.
* Dynamic wrap / nowrap switchable.
* Supports alternating row style changes.
* Shortcut keys are customizable.
* The style of the effect is customizable.
* Supports follow-mode (like tail -f).
* Supports following multiple files and switching when updated.
* Run the command you can target the stdout/stderr.
* Supports incremental search and regular expression search.

## install

### deb package

You can download the package from [releases](https://github.com/noborus/ov/releases).

```console
curl -L -O https://github.com/noborus/ov/releases/download/vx.x.x/ov_x.x.x-1_amd64.deb
sudo dpkg -i ov_x.x.x-1_amd64.deb
```

### rpm package

You can download the package from [releases](https://github.com/noborus/ov/releases).

```console
sudo rpm -ivh https://github.com/noborus/ov/releases/download/vx.x.x/ov_x.x.x-1_amd64.rpm
```

### MacPorts (macOS)

```console
sudo port install ov
```

### Homebrew(macOS or Linux)

```console
brew install noborus/tap/ov
```

### pkg (FreeBSD)

```console
pkg install ov
```

### binary

You can download the binary from [releases](https://github.com/noborus/ov/releases).

```console
curl -L -O https://github.com/noborus/ov/releases/download/vx.x.x/ov_x.x.x_linux_amd64.zip
unzip ov_x.x.x_linux_amd64.zip
sudo install ov /usr/local/bin
```

### go install

It will be installed in $GOPATH/bin by the following command.

```console
go install github.com/noborus/ov@latest
```

### go get(details or developer version)

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

## Usage

ov supports open file name or standard input.

```console
ov filename
```

```console
cat filename|ov
```

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

### Search

Search by forward search `/` key(default) or the backward search `?` key(defualt).
Search can be toggled between incremental search, regular expression search, and case sensitivity.
Displayed when the following are enabled in the search input prompt:

| Function | display | (Default)key |command option |
|:---------|:--------|:----|:--------------|
| Incremental search | (I) | alt+i | --incremental |
| Regular expression search | (R) | alt+r | --regexp-search  |
| Case sensitive | (Aa) | alt+c |  -i, --case-sensitive |

### Mark

Mark the display position with the `m` key(default).
The mark is decorated with `StyleMarkLine` and `MarkStyleWidth`.

Marks can be erased individually with the `M` key(default).
It is also possible to delete all marks with the `ctrl + delete` key(default).

Use the `>`next and `<`previous (default) key to move to the marked position.

### Mouse support

The ov makes the mouse support its control.
This can be disabled with the option `--disable-mouse`.

If mouse support is enabled, tabs and line breaks will be interpreted correctly when copying.

Copying to the clipboard uses [atotto/clipboard](https://github.com/atotto/clipboard).
For this reason, the 'xclip' or 'xsel' command is required in Linux/Unix environments.

Selecting the range with the mouse and then left-clicking will copy it to the clipboard.

Pasting in ov is done with the middle button.
In other applications, it is pasted from the clipboard (often by pressing the right-click).

## config

You can set style and key bindings in the setting file.

Please refer to the sample [ov.yaml](https://github.com/noborus/ov/blob/master/ov.yaml) configuration file.

### follow mode

Output appended data and move it to the bottom line (like tail -f).

```sh
ov --follow-mode /var/log/syslog
```

```sh
(while :; do echo random-$RANDOM; sleep 0.1; done;)|./ov  --follow-mode
```

![ov-tail.gif](https://raw.githubusercontent.com/noborus/ov/master/docs/ov-tail.gif)

### follow all mode

Same as follow-mode, and switches to the last updated file when there are multiple files.

```sh
ov --follow-all /var/log/nginx/access.log /var/log/nginx/error.log
```

### exec mode

Execute the command to display stdout / stderr.
Arguments after (`--`) are interpreted as command arguments.

```sh
ov --follow-all --exec -- make
```

![ov-exec.gif](https://raw.githubusercontent.com/noborus/ov/master/docs/ov-exec.gif)

### psql

Set environment variable `PSQL_PAGER`(PostgreSQL 11 or later).

```sh
export PSQL_PAGER='ov -w=f -H2 -F -C -d "|"'
```

You can also write in `~/.psqlrc` in previous versions.

```sh
\setenv PAGER 'ov -w=f -H2 -F -C -d "|"'
```

### mysql

Use the --pager option with the mysql client.

```console
mysql --pager='ov -w=f -H3 -F -C -d "|"'
```

You can also write in `~/.my.cnf`.

```console
[client]
pager=ov -w=f -H3 -F -C -d "|"
```

## Key bindings

```console
  [Escape], [q]              * quit
  [ctrl+c]                   * cancel
  [Q]                        * output screen and quit
  [h], [ctrl+alt+c]          * display help screen
  [ctrl+alt+e]               * display log screen
  [ctrl+l]                   * screen sync
  [ctrl+f]                   * follow mode toggle
  [ctrl+a]                   * follow all mode toggle
  [ctrl+alt+r]               * enable/disable mouse
  [ctrl+k]                   * close current document

	Moving

  [Enter], [Down], [ctrl+N]  * forward by one line
  [Up], [ctrl+p]             * backward by one line
  [Home]                     * go to begin of line
  [End]                      * go to end of line
  [PageDown], [ctrl+v]       * forward by page
  [PageUp], [ctrl+b]         * backward by page
  [ctrl+d]                   * forward a half page
  [ctrl+u]                   * backward a half page
  [left]                     * scroll to left
  [right]                    * scroll to right
  [ctrl+left]                * scroll left half screen
  [ctrl+right]               * scroll right half screen
  [g]                        * number of go to line
  []]                        * next document
  [[]                        * previous document

	Mark position

  [m]                        * mark current position
  [M]                        * remove mark current position
  [ctrl+delete]              * remove all mark
  [>]                        * move to next marked position
  [<]                        * move to previous marked position

	Search

  [/]                        * forward search mode
  [?]                        * backward search mode
  [n]                        * repeat forward search
  [N]                        * repeat backward search

	Change display

  [p], [P]                   * view mode selection
  [w], [W]                   * wrap/nowrap toggle
  [c]                        * column mode toggle
  [C]                        * color to alternate rows toggle
  [G]                        * line number toggle

	Change Display with Input

  [d]                        * delimiter string
  [H]                        * number of header lines
  [ctrl+s]                   * number of skip lines
  [t]                        * TAB width

	Key binding when typing

  [alt+c]                    * case-sensitive toggle
  [alt+r]                    * regular expression search toggle
  [alt+i]                    * incremental search toggle
```

## Customize

### Style customization

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

### Key binding customization

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
