# Oviewer

Oviewer is a feature rich pager.

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
Oviewer is a feature rich pager(such as more/less).
It supports various compressed files(gzip, bzip2, zstd, lz4, and xz).

Usage:
  ov [flags]

Flags:
      --config string   config file (default is $HOME/.oviewer.yaml)
  -H, --header int      number of header rows to fix
  -h, --help            help for ov
  -X, --post-write      Output the current screen when exiting
  -x, --tab-width int   tab stop (default 8)
  -v, --version         display version information
  -w, --wrap            wrap mode (default true)
```
