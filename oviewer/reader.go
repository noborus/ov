package oviewer

import (
	"bufio"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sync/atomic"

	"github.com/klauspost/compress/zstd"
	"github.com/pierrec/lz4"
	"github.com/ulikunitz/xz"
	"golang.org/x/term"
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

// ReadAll reads all from the reader to the buffer.
// It returns if beforeSize is accumulated in buffer
// before the end of read.
func (m *Document) ReadAll(r io.Reader) error {
	reader := bufio.NewReader(r)
	go func() {
		select {
		case <-m.closeCh:
			return
		default:
		}

		err := m.readAll(reader)
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrClosedPipe) || errors.Is(err, os.ErrClosed) {
				close(m.eofCh)
				atomic.StoreInt32(&m.eof, 1)
				return
			}
			log.Printf("error: %v\n", err)
			atomic.StoreInt32(&m.eof, 0)
			return
		}
	}()
	return nil
}

// ReadFile reads file.
func (m *Document) ReadFile(fileName string) error {
	if fileName == "" {
		if term.IsTerminal(0) {
			return ErrMissingFile
		}
		m.file = os.Stdin
		m.FileName = "(STDIN)"
	} else {
		m.FileName = fileName
		r, err := os.Open(fileName)
		if err != nil {
			return err
		}
		m.file = r
	}

	cFormat, reader := uncompressedReader(m.file)
	m.CFormat = cFormat

	go func() {
		<-m.eofCh
		m.close()
		close(m.reOpenCh)
	}()
	if err := m.ReadAll(reader); err != nil {
		return err
	}
	return nil
}

// Close closes the File.
// Record the last read position.
func (m *Document) close() error {
	pos, err := m.file.Seek(0, io.SeekCurrent)
	if err != nil {
		return fmt.Errorf("seek: %w", err)
	}
	m.offset = pos
	if err := m.file.Close(); err != nil {
		return fmt.Errorf("close(): %w", err)
	}
	return nil
}

// reOpenRead reopens and reads the file.
// Seek to the position where the file was closed, and then read.
func (m *Document) reOpenRead() error {
	r, err := os.Open(m.FileName)
	if err != nil {
		return err
	}
	m.mu.Lock()
	m.file = r
	m.mu.Unlock()
	atomic.StoreInt32(&m.eof, 0)

	_, err = r.Seek(m.offset, io.SeekStart)
	if err != nil {
		return err
	}

	rr := compressedFormatReader(m.CFormat, r)
	reader := bufio.NewReader(rr)
	for {
		select {
		case <-m.closeCh:
			return nil
		default:
		}
		err := m.readAll(reader)
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrClosedPipe) || errors.Is(err, os.ErrClosed) {
				<-m.changCh
				continue
			}
			log.Println(err)
			break
		}
	}
	return nil
}

func (m *Document) readAll(reader *bufio.Reader) error {
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

		m.append(line.String())
		line.Reset()
	}
}

func (m *Document) append(line string) {
	m.mu.Lock()
	m.lines = append(m.lines, line)
	m.endNum++
	m.mu.Unlock()
	atomic.StoreInt32(&m.changed, 1)
}
