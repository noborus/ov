package oviewer

import (
	"bufio"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"time"

	"github.com/klauspost/compress/zstd"
	"github.com/pierrec/lz4"
	"github.com/ulikunitz/xz"
)

func uncompressedReader(reader io.Reader) io.ReadCloser {
	var err error
	buf := [7]byte{}
	n, err := io.ReadAtLeast(reader, buf[:], len(buf))
	if err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			return ioutil.NopCloser(bytes.NewReader(buf[:n]))
		}
		return ioutil.NopCloser(bytes.NewReader(nil))
	}

	var r io.ReadCloser
	rd := io.MultiReader(bytes.NewReader(buf[:n]), reader)

	switch {
	case bytes.Equal(buf[:3], []byte{0x1f, 0x8b, 0x8}):
		r, err = gzip.NewReader(rd)
	case bytes.Equal(buf[:3], []byte{0x42, 0x5A, 0x68}):
		r = ioutil.NopCloser(bzip2.NewReader(rd))
	case bytes.Equal(buf[:4], []byte{0x28, 0xb5, 0x2f, 0xfd}):
		var zr *zstd.Decoder
		zr, err = zstd.NewReader(rd)
		r = ioutil.NopCloser(zr)
	case bytes.Equal(buf[:4], []byte{0x04, 0x22, 0x4d, 0x18}):
		r = ioutil.NopCloser(lz4.NewReader(rd))
	case bytes.Equal(buf[:7], []byte{0xfd, 0x37, 0x7a, 0x58, 0x5a, 0x0, 0x0}):
		var zr *xz.Reader
		zr, err = xz.NewReader(rd)
		r = ioutil.NopCloser(zr)
	}
	if err != nil || r == nil {
		r = ioutil.NopCloser(rd)
	}
	return r
}

// ReadAll reads all from the reader to the buffer.
// It returns if beforeSize is accumulated in buffer
// before the end of read.
func (m *Document) ReadAll(r io.ReadCloser) error {
	reader := bufio.NewReader(r)
	ch := make(chan struct{})
	go func() {
		defer close(ch)
		defer r.Close()

		var line bytes.Buffer

		for {
			buf, isPrefix, err := reader.ReadLine()
			if err != nil {
				if errors.Is(err, io.EOF) || errors.Is(err, io.ErrClosedPipe) {
					break
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
		m.eof = true
	}()

	select {
	case <-ch:
		return nil
	case <-time.After(500 * time.Millisecond):
		return nil
	}
}
