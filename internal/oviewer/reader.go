package oviewer

import (
	"bufio"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"io"
	"io/ioutil"
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
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return ioutil.NopCloser(bytes.NewReader(buf[:n]))
		}
		return ioutil.NopCloser(bytes.NewReader(nil))
	}

	rd := io.MultiReader(bytes.NewReader(buf[:n]), reader)
	var r io.ReadCloser
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
func (m *Model) ReadAll(r io.Reader) error {
	reader := bufio.NewReader(r)
	ch := make(chan struct{})
	go func() {
		var buf bytes.Buffer
		defer close(ch)
		for {
			l, isPrefix, err := reader.ReadLine()
			buf.Write(l)
			if err != nil {
				if err == io.EOF || err == io.ErrClosedPipe {
					break
				}
				m.eof = false
				return
			}
			if isPrefix {
				continue
			}
			m.mu.Lock()
			m.buffer = append(m.buffer, buf.String())
			buf.Reset()
			m.endNum++
			m.mu.Unlock()
			if m.endNum == m.beforeSize {
				ch <- struct{}{}
			}
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
