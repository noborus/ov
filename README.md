# ov - Oviewer

[![PkgGoDev](https://pkg.go.dev/badge/github.com/noborus/ov)](https://pkg.go.dev/github.com/noborus/ov)
[![Actions Status](https://github.com/noborus/ov/workflows/Go/badge.svg)](https://github.com/noborus/ov/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/noborus/ov)](https://goreportcard.com/report/github.com/noborus/ov)

ov is a feature rich terminal pager.

(The old repository name was oviewer.)

![ov.png](https://raw.githubusercontent.com/noborus/ov/master/docs/ov.png)

## feature

* Better support for unicode and wide width.
* Support for compressed files (gzip, bzip2, zstd, lz4, xz).
* Supports column mode.
* Header rows can be fixed.
* Dynamic wrap / nowrap switchable.
* Background color to alternate rows.
* Columns can be selected with separators.

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

### Homebrew(macOS or Linux)

```console
brew install noborus/tap/ov
```

### binary

You can download the binary from [releases](https://github.com/noborus/ov/releases).

```console
curl -L -O https://github.com/noborus/ov/releases/download/vx.x.x/ov_x.x.x_linux_amd64.zip
unzip ov_x.x.x_linux_amd64.zip
sudo install ov /usr/local/bin
```

### go get(simplified version)

It will be installed in $GOPATH/bin by the following command.

```console
go get -u github.com/noborus/ov
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
  -C, --alternate-rows            color to alternate rows
  -i, --case-sensitive            case-sensitive in search
  -d, --column-delimiter string   column delimiter (default ",")
  -c, --column-mode               column mode
      --config string             config file (default is $HOME/.ov.yaml)
      --debug                     debug mode
  -X, --exit-write                output the current screen when exiting
  -H, --header int                number of header rows to fix
  -h, --help                      help for ov
      --help-key                  display key bind information
  -n, --line-number               line number
  -F, --quit-if-one-screen        quit if the output fits on one screen
  -x, --tab-width int             tab stop width (default 8)
  -v, --version                   display version information
  -w, --wrap                      wrap mode (default true)
```

It can also be changed after startup.
Refer to the [motion image](docs/image.md).

## config

You can set colors and key bindings in the setting file.

Please refer to the sample [ov.yaml](https://github.com/noborus/ov/blob/master/ov.yaml) configuration file.

### psql

Set environment variable `PSQL_PAGER`(PostgreSQL 11 or later).

```sh
export PSQL_PAGER='ov -w=f -H2 -F -C -d "|"'
```

You can also write in `~/.psqlrc` in previous versions.

```filename:~/.psqlrc
\setenv PAGER 'ov -w=f -H2 -F -C -d "|"'
```

### mysql

Use the --pager option with the mysql client.

```console
mysql --pager='ov -w=f -H3 -F -C -d "|"'
```

You can also write in `~/.my.cnf`.

```filename:~/.my.cnf
[client]
pager=ov -w=f -H3 -F -C -d "|"
```

## Key bindings

```
  [Escape], [q], [ctrl+c]   : exit             * quit
  [Q]                       : write_exit       * output screen and quit
  [h], [ctrl+alt+c]         : help             * display help screen
  [ctrl+l]                  : sync             * screen sync

	Moving

  [Enter], [Down], [ctrl+N] : down             * forward by one line
  [Up], [ctrl+p]            : up               * backward by one line
  [Home]                    : top              * go to begin of line
  [End]                     : bottom           * go to end of line
  [PageDown], [ctrl+v]      : page_down        * forward by page
  [PageUp], [ctrl+b]        : page_up          * backward by page
  [ctrl+d]                  : page_half_down   * forward a half page
  [ctrl+u]                  : page_half_up     * backward a half page
  [left]                    : left             * scroll to left
  [right]                   : right            * scroll to right
  [ctrl+left]               : half_left        * scroll left half screen
  [ctrl+right]              : half_right       * scroll right half screen
  [g]                       : goto             * number of go to line

	Mark position

  [m]                       : mark             * mark current position
  [>]                       : next_mark        * move to next marked position
  [<]                       : previous_mark    * move to previous marked position

	Search

  [/]                       : search           * forward search mode
  [?]                       : backsearch       * backward search mode
  [n]                       : next_search      * repeat forward search
  [N]                       : next_backsearch  * repeat backward search

	Change display

  [w], [W]                  : wrap_mode        * wrap/nowrap toggle
  [c]                       : column_mode      * column mode toggle
  [C]                       : alter_rows_mode  * color to alternate rows toggle
  [G]                       : line_number_mode * line number togle

	Change Display with Input

  [d]                       : delimiter        * delimiter string
  [H]                       : header           * number of header lines
  [t]                       : tabwidth         * TAB width

```