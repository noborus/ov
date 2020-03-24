# Zpager

Zpager is a feature rich pager.

## install

```console
go get -u github.com/noborus/zpager...
```

## Usage

Zpager supports open file name or standard input.

```console
Feature rich pager(such as more/less).
It supports various compressed files(gzip, bzip2, zstd, lz4, and xz).

Usage:
  zpager [flags]

Flags:
      --config string   config file (default is $HOME/.zpager.yaml)
  -H, --header int      number of header rows to fix
  -h, --help            help for zpager
  -X, --no-init         Output the current screen when exiting
  -x, --tab-width int   tab stop (default 8)
  -v, --version         display version information
  -w, --wrap            wrap mode (default true)
```
