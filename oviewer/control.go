package oviewer

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sync/atomic"
)

type controlSpecifier struct {
	searcher Searcher
	control  control
	chunkNum int
	done     chan bool
}

type control string

const (
	firstControl    = "first"
	continueControl = "read"
	followControl   = "follow"
	closeControl    = "close"
	reloadControl   = "reload"
	loadControl     = "load"
	searchControl   = "search"
)

// ControlFile controls file read and loads in chunks.
// ControlFile can be reloaded by file name.
func (m *Document) ControlFile(file *os.File) error {
	go func() error {
		atomic.StoreInt32(&m.closed, 0)
		r, err := m.fileReader(file)
		if err != nil {
			atomic.StoreInt32(&m.closed, 1)
			log.Println(err)
		}
		atomic.StoreInt32(&m.eof, 0)
		reader := bufio.NewReader(r)
		for sc := range m.ctlCh {
			reader, err = m.control(sc, reader)
			if err != nil {
				log.Println(sc.control, err)
			}
			if sc.done != nil {
				if err != nil {
					sc.done <- false
				} else {
					sc.done <- true
				}
				close(sc.done)
			}
		}
		log.Println("close m.ctlCh")
		return nil
	}()

	m.startControl()
	return nil
}

func (m *Document) control(sc controlSpecifier, reader *bufio.Reader) (*bufio.Reader, error) {
	if atomic.LoadInt32(&m.closed) == 1 && sc.control != reloadControl {
		return nil, fmt.Errorf("closed %s", sc.control)
	}

	var err error
	switch sc.control {
	case firstControl:
		reader, err = m.firstRead(reader)
		if !m.BufEOF() {
			m.continueControl()
		}
	case continueControl:
		reader, err = m.continueRead(reader)
		if !m.BufEOF() {
			m.continueControl()
		}
	case followControl:
		reader, err = m.followRead(reader)
	case loadControl:
		reader, err = m.readChunk(reader, sc.chunkNum)
	case searchControl:
		_, err = m.searchChunk(sc.chunkNum, sc.searcher)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return reader, nil
			}
			return reader, err
		}
		return m.readChunk(reader, sc.chunkNum)
	case reloadControl:
		reader, err = m.reloadRead(reader)
		m.startControl()
	case closeControl:
		err = m.close()
		log.Println(err)
		err = nil
	default:
		panic(fmt.Sprintf("unexpected %s", sc.control))
	}
	return reader, nil
}

// ControlLog controls log.
// ControlLog is only supported reload.
func (m *Document) ControlLog() error {
	go func() error {
		for sc := range m.ctlCh {
			switch sc.control {
			case reloadControl:
				m.reset()
			default:
				panic(fmt.Sprintf("unexpected %s", sc.control))
			}
			if sc.done != nil {
				close(sc.done)
			}
		}
		log.Println("close m.ctlCh")
		return nil
	}()
	return nil
}

// ControlReader is the controller for io.Reader.
// Assuming call from Exec. reload executes the argument function.
func (m *Document) ControlReader(r io.Reader, reload func() *bufio.Reader) error {
	reader := bufio.NewReader(r)
	go func() error {
		var err error
		for sc := range m.ctlCh {
			switch sc.control {
			case firstControl:
				reader, err = m.firstRead(reader)
				if !m.BufEOF() {
					m.continueControl()
				}
			case continueControl:
				reader, err = m.continueRead(reader)
				if !m.BufEOF() {
					m.continueControl()
				}
			case reloadControl:
				log.Println("reset")
				reader = reload()
				m.startControl()
			default:
				panic(fmt.Sprintf("unexpected %s", sc.control))
			}
			if err != nil {
				log.Println(err)
			}
			if sc.done != nil {
				close(sc.done)
			}
		}
		log.Println("close ctlCh")
		return nil
	}()

	m.startControl()
	return nil
}

func (m *Document) startControl() {
	go func() {
		m.ctlCh <- controlSpecifier{
			control: firstControl,
		}
	}()
}

func (m *Document) continueControl() {
	go func() {
		m.ctlCh <- controlSpecifier{
			control: continueControl,
		}
	}()
}
