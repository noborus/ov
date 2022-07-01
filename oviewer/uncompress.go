package oviewer

import (
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"errors"
	"io"

	"github.com/klauspost/compress/zstd"
	"github.com/pierrec/lz4"
	"github.com/ulikunitz/xz"
)

// Compressed represents the type of compression.
type Compressed int

const (
	// UNCOMPRESSED is an uncompressed format.
	UNCOMPRESSED Compressed = iota
	// GZIP is gzip compressed format.
	GZIP
	// BZIP2 is bzip2 compressed format.
	BZIP2
	// ZSTD is zstd compressed format.
	ZSTD
	// LZ4 is lz4 compressed format.
	LZ4
	// XZ is xz compressed format.
	XZ
)

func compressType(header []byte) Compressed {
	switch {
	case bytes.Equal(header[:3], []byte{0x1f, 0x8b, 0x8}):
		return GZIP
	case bytes.Equal(header[:3], []byte{0x42, 0x5A, 0x68}):
		return BZIP2
	case bytes.Equal(header[:4], []byte{0x28, 0xb5, 0x2f, 0xfd}):
		return ZSTD
	case bytes.Equal(header[:4], []byte{0x04, 0x22, 0x4d, 0x18}):
		return LZ4
	case bytes.Equal(header[:7], []byte{0xfd, 0x37, 0x7a, 0x58, 0x5a, 0x0, 0x0}):
		return XZ
	}
	return UNCOMPRESSED
}

func (c Compressed) String() string {
	switch c {
	case GZIP:
		return "GZIP"
	case BZIP2:
		return "BZIP2"
	case ZSTD:
		return "ZSTD"
	case LZ4:
		return "LZ4"
	case XZ:
		return "XZ"
	}
	return "UNCOMPRESSED"
}

func uncompressedReader(reader io.Reader) (Compressed, io.Reader) {
	buf := [7]byte{}
	n, err := io.ReadAtLeast(reader, buf[:], len(buf))
	if err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			return UNCOMPRESSED, bytes.NewReader(buf[:n])
		}
		return UNCOMPRESSED, bytes.NewReader(nil)
	}

	mr := io.MultiReader(bytes.NewReader(buf[:n]), reader)
	cFormat := compressType(buf[:7])
	r := compressedFormatReader(cFormat, mr)

	return cFormat, r
}

func compressedFormatReader(cFormat Compressed, reader io.Reader) io.Reader {
	var r io.Reader
	var err error
	switch cFormat {
	case GZIP:
		r, err = gzip.NewReader(reader)
	case BZIP2:
		r = bzip2.NewReader(reader)
	case ZSTD:
		r, err = zstd.NewReader(reader)
	case LZ4:
		r = lz4.NewReader(reader)
	case XZ:
		r, err = xz.NewReader(reader)
	}
	if err != nil || r == nil {
		r = reader
	}
	return r
}
