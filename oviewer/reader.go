package oviewer

import (
	"bufio"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync/atomic"

	"github.com/klauspost/compress/zstd"
	"github.com/pierrec/lz4"
	"github.com/ulikunitz/xz"
	"golang.org/x/term"
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
		if m.seekable {
			if err := m.close(); err != nil {
				log.Printf("ReadFile: %s", err)
			}
		}
		atomic.StoreInt32(&m.changed, 1)
		m.followCh <- struct{}{}
	}()
	if STDOUTPIPE != nil {
		reader = io.TeeReader(reader, STDOUTPIPE)
	}

	return m.ReadAll(reader)
}

// onceFollowMode opens the follow mode only once.
func (m *Document) onceFollowMode() {
	if atomic.SwapInt32(&m.openFollow, 1) == 1 {
		return
	}
	if m.file == nil {
		return
	}

	ctx := context.Background()
	ctx, m.cancel = context.WithCancel(ctx)
	go m.startFollowMode(ctx)
}

// startFollowMode opens the file in follow mode.
// Seek to the position where the file was closed, and then read.
func (m *Document) startFollowMode(ctx context.Context) {
	defer m.cancel()
	<-m.followCh
	if m.seekable {
		// Wait for the file to open until it changes.
		select {
		case <-ctx.Done():
			log.Println("follow mode cancel")
			return
		case <-m.changCh:
		}
		m.file = m.openFollowFile()
	}

	r := compressedFormatReader(m.CFormat, m.file)
	if err := m.ContinueReadAll(ctx, r); err != nil {
		log.Printf("%s follow mode read %v", m.FileName, err)
	}
}

func (m *Document) openFollowFile() *os.File {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, err := os.Open(m.FileName)
	if err != nil {
		log.Printf("openFollowFile: %s", err)
		return m.file
	}
	atomic.StoreInt32(&m.closed, 0)
	atomic.StoreInt32(&m.eof, 0)
	if _, err := r.Seek(m.offset, io.SeekStart); err != nil {
		log.Printf("openFollowMode: %s", err)
	}
	return r
}

// Close closes the File.
// Record the last read position.
func (m *Document) close() error {
	if m.checkClose() {
		return nil
	}

	if m.seekable {
		pos, err := m.file.Seek(0, io.SeekCurrent)
		if err != nil {
			return fmt.Errorf("close: %w", err)
		}
		m.offset = pos
	}
	if err := m.file.Close(); err != nil {
		return fmt.Errorf("close: %w", err)
	}
	atomic.StoreInt32(&m.openFollow, 0)
	atomic.StoreInt32(&m.closed, 1)
	atomic.StoreInt32(&m.changed, 1)
	return nil
}

// ReadAll reads all from the reader to the buffer.
// It returns if beforeSize is accumulated in buffer
// before the end of read.
func (m *Document) ReadAll(r io.Reader) error {
	reader := bufio.NewReader(r)
	go func() {
		if m.checkClose() {
			return
		}

		if err := m.readAll(reader); err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrClosedPipe) || errors.Is(err, os.ErrClosed) {
				m.eofCh <- struct{}{}
				atomic.StoreInt32(&m.eof, 1)
				return
			}
			log.Printf("error: %v\n", err)
			atomic.StoreInt32(&m.eof, 0)
			return
		}
	}()

	// Named pipes for continuous read.
	if !m.seekable {
		m.onceFollowMode()
	}
	return nil
}

// ReadReader reads reader.
// A wrapper for ReadAll, used when eofCh notifications are not needed.
func (m *Document) ReadReader(r io.Reader) error {
	go func() {
		<-m.eofCh
	}()

	return m.ReadAll(r)
}

// ContinueReadAll continues to read even if it reaches EOF.
func (m *Document) ContinueReadAll(ctx context.Context, r io.Reader) error {
	reader := bufio.NewReader(r)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if m.checkClose() {
			return nil
		}

		if err := m.readAll(reader); err != nil {
			if errors.Is(err, io.EOF) {
				<-m.changCh
				continue
			}
			return err
		}
	}
}

func (m *Document) readAll(reader *bufio.Reader) error {
	var line strings.Builder

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

func (m *Document) reload() error {
	if (m.file == os.Stdin && m.BufEOF()) || !m.seekable && m.checkClose() {
		return fmt.Errorf("%w %s", ErrAlreadyClose, m.FileName)
	}

	if m.seekable {
		if m.cancel != nil {
			m.cancel()
		}
		if !m.checkClose() && m.file != nil {
			if err := m.close(); err != nil {
				log.Println(err)
			}
		}
	}

	m.reset()

	if !m.seekable {
		return nil
	}

	atomic.StoreInt32(&m.closed, 0)
	return m.ReadFile(m.FileName)
}

func (m *Document) append(lines ...string) {
	m.mu.Lock()
	for _, line := range lines {
		m.lines = append(m.lines, line)
		m.endNum++
	}
	m.mu.Unlock()
	atomic.StoreInt32(&m.changed, 1)
}

func (m *Document) reset() {
	m.mu.Lock()
	m.endNum = 0
	m.lines = m.lines[:0]
	m.mu.Unlock()
	atomic.StoreInt32(&m.changed, 1)
	m.ClearCache()
}

func (m *Document) checkClose() bool {
	return atomic.LoadInt32(&m.closed) == 1
}
