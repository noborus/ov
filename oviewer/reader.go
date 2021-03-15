package oviewer

import (
	"bufio"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"encoding/binary"
	"errors"
	"io"
	"log"
	"os"
	"syscall"
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
	var err error
	buf := [7]byte{}
	n, err := io.ReadAtLeast(reader, buf[:], len(buf))
	if err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			return UNCOMPRESSED, bytes.NewReader(buf[:n])
		}
		return UNCOMPRESSED, bytes.NewReader(nil)
	}

	mr := io.MultiReader(bytes.NewReader(buf[:n]), reader)

	var r io.Reader
	cFormat := compressType(buf[:7])
	switch cFormat {
	case GZIP:
		r, err = gzip.NewReader(mr)
	case BZIP2:
		r = bzip2.NewReader(mr)
	case ZSTD:
		r, err = zstd.NewReader(mr)
	case LZ4:
		r = lz4.NewReader(mr)
	case XZ:
		r, err = xz.NewReader(mr)
	}
	if err != nil || r == nil {
		r = mr
	}
	return cFormat, r
}

func (m *Document) ReadFollow(pos int64) error {
	if !m.BufEOF() {
		return nil
	}
	m.mu.Lock()
	m.eof = false
	m.mu.Unlock()

	r, err := os.Open(m.FileName)
	if err != nil {
		return err
	}
	fd, _ := syscall.InotifyInit()
	_, _ = syscall.InotifyAddWatch(fd, m.FileName, syscall.IN_MODIFY)

	_, err = r.Seek(pos, io.SeekStart)
	if err != nil {
		return err
	}

	go func() {
		reader := bufio.NewReader(r)
		var line bytes.Buffer
		for {
			buf, isPrefix, err := reader.ReadLine()
			if err != nil {
				if errors.Is(err, io.EOF) || errors.Is(err, io.ErrClosedPipe) || errors.Is(err, os.ErrClosed) {
					if err = waitForChange(fd); err != nil {
						return
					}
					continue
				}
				log.Printf("error: %v\n", err)
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
			line.Reset()
		}
	}()
	log.Println("Follow mode end")
	return nil
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

func waitForChange(fd int) error {
	for {
		var buf [syscall.SizeofInotifyEvent]byte
		_, err := syscall.Read(fd, buf[:])
		if err != nil {
			return err
		}
		r := bytes.NewReader(buf[:])
		var ev = syscall.InotifyEvent{}
		_ = binary.Read(r, binary.LittleEndian, &ev)
		if ev.Mask&syscall.IN_MODIFY == syscall.IN_MODIFY {
			return nil
		}
	}
}
