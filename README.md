# Oviewer

Oviewer is a feature rich terminal pager.

## feature

* Better support for unicode and wide width.
* Support for compressed files (gzip, bzip2, zstd, lz4, xz).
* Header row can be fixed.
* Dynamic wrap / nowrap switchable.

## install

```console
go get -u github.com/noborus/oviewer...
```

## Usage

The command name of Oviewer is `ov`.

Oviewer supports open file name or standard input.

```console
ov filename
```

```console
cat filename|ov
```

```console
$ ov --help
Oviewer is a feature rich pager(such as more/less).
It supports various compressed files(gzip, bzip2, zstd, lz4, and xz).

Usage:
  ov [flags]

Flags:
  -i, --case-sensitive       case-sensitive
      --config string        config file (default is $HOME/.oviewer.yaml)
      --debug                Debug mode
  -X, --exit-write           Output the current screen when exiting
  -H, --header int           number of header rows to fix
  -h, --help                 help for ov
  -F, --quit-if-one-screen   Quit if the output fits on one screen
  -x, --tab-width int        tab stop (default 8)
  -v, --version              display version information
  -w, --wrap                 wrap mode (default true)
```

### psql

Set environment variable `PSQL_PAGER`(PostgreSQL 11 or later).

```sh
export PSQL_PAGER="ov -w=f -H2 -F"
```

You can also write in `~/.psqlrc` in previous versions.

```filename:~/.psqlrc
\setenv PAGER 'ov -w=f -H2 -F'
```

### mysql

Use the --pager option with the mysql client.

```console
mysql --pager="ov -w=f -H3 -F"
```

You can also write in `~/.my.cnf`.

```filename:~/.my.cnf
[client]
pager=ov -w=f -H3 -F
```

## Key bindings

* <kbd>q</kbd>,<kbd>Esc</kbd> - quit
* <kbd>Q</kbd> - output screen and quit

### Move

* <kbd>HOME</kbd> - go to begin of line
* <kbd>END</kbd> - go to end of line
* <kbd>KEY_UP</kbd> - backward by one line
* <kbd>KEY_DOWN</kbd>, <kbd>Enter</kbd> - forward by one line
* <kbd>PgUP</kbd>, <kbd>Ctrl</kbd>+<kbd>b</kbd> - backward by page
* <kbd>PgDn</kbd>, <kbd>Ctrl</kbd>+<kbd>v</kbd> - forward by page
* <kbd>Ctrl</kbd>+<kbd>d</kbd> - forward a half page
* <kbd>Ctrl</kbd>+<kbd>u</kbd> - backward a half page

* <kbd>KEY_LEFT</kbd> - scroll to left
* <kbd>KEY_RIGHT</kbd> - scroll to right

* <kbd>Ctrl</kbd>+<kbd>KEY_LEFT</kbd> - scroll to left page
* <kbd>Ctrl</kbd>+<kbd>KEY_RIGHT</kbd> - scroll to right page

### Mode

* <kbd>w</kbd> - wrap/nowrap toggle

### Input Mode

* <kbd>/</kbd> - forward search mode
* <kbd>?</kbd> - previous search mode
* <kbd>H</kbd> - number of header lines
* <kbd>g</kbd> - number of go to line

## Keys in search mode

* <kbd>Ctrl</kbd>+<kbd>i</kbd> - case-sensitive/insensitive toggle

### Key after search input mode

* <kbd>n</kbd> - for next match
* <kbd>N</kbd> - for next match in reverse direction
