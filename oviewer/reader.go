package oviewer

import (
	"bufio"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"errors"
	"io"
	"log"
	"os"
	"time"

	"github.com/klauspost/compress/zstd"
	"github.com/pierrec/lz4"
	"github.com/ulikunitz/xz"
)

type Compressed int

const (
	UNCOMPRESSED Compressed = iota
	GZIP
	BZIP2
	ZSTD
	LZ4
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

func (m *Document) ReadFollow(reader *bufio.Reader) error {
	var line bytes.Buffer

	for {
		buf, isPrefix, err := reader.ReadLine()
		if err != nil {
			return err
		}
		line.Write(buf)
		if isPrefix {
			continue
		}

		m.mu.Lock()
		m.lines = append(m.lines, line.String())
		m.endNum++
		m.mu.Unlock()
		line.Reset()
	}
}

// ReadAll reads all from the reader to the buffer.
// It returns if beforeSize is accumulated in buffer
// before the end of read.
func (m *Document) ReadAll(r io.Reader) error {
	reader := bufio.NewReader(r)
	ch := make(chan struct{})
	go func() {
		defer close(ch)

		var line bytes.Buffer
		for {
			buf, isPrefix, err := reader.ReadLine()
			if err != nil {
				m.mu.Lock()
				defer m.mu.Unlock()
				if errors.Is(err, io.EOF) || errors.Is(err, io.ErrClosedPipe) || errors.Is(err, os.ErrClosed) {
					m.eof = true
					return
				}
				log.Printf("error: %v\n", err)
				m.eof = false
				return
			}

			line.Write(buf)
			if isPrefix {
				continue
			}

			m.mu.Lock()
			m.lines = append(m.lines, line.String())
			m.endNum++
			m.mu.Unlock()
			if m.endNum == m.beforeSize {
				ch <- struct{}{}
			}
			line.Reset()
		}
	}()

	select {
	case <-ch:
		return nil
	case <-time.After(500 * time.Millisecond):
		go func() {
			<-ch
		}()
		return nil
	}
}

func (m *Document) ReadAllChan(cl chan bool, r io.Reader) error {
	reader := bufio.NewReader(r)
	ch := make(chan struct{})
	go func() {
		defer close(ch)

		var line bytes.Buffer
		for {
			buf, isPrefix, err := reader.ReadLine()
			if err != nil {
				m.mu.Lock()
				defer m.mu.Unlock()
				if errors.Is(err, io.EOF) || errors.Is(err, io.ErrClosedPipe) || errors.Is(err, os.ErrClosed) {
					m.eof = true
					cl <- true
					return
				}
				log.Printf("error: %v\n", err)
				m.eof = false
				return
			}

			line.Write(buf)
			if isPrefix {
				continue
			}

			m.mu.Lock()
			m.lines = append(m.lines, line.String())
			m.endNum++
			m.mu.Unlock()
			if m.endNum == m.beforeSize {
				ch <- struct{}{}
			}
			line.Reset()
		}
	}()

	select {
	case <-ch:
		return nil
	case <-time.After(500 * time.Millisecond):
		go func() {
			<-ch
		}()
		return nil
	}
}
