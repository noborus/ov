package oviewer

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"sync/atomic"
)

// controlSpecifier represents a control request.
type controlSpecifier struct {
	searcher Searcher
	done     chan bool
	request  request
	chunkNum int
}

// request represents a control request.
type request string

// control requests.
const (
	requestStart    request = "start"
	requestContinue request = "continue"
	requestFollow   request = "follow"
	requestClose    request = "close"
	requestReload   request = "reload"
	requestLoad     request = "load"
	requestSearch   request = "search"
)

// ControlFile controls file read and loads in chunks.
// ControlFile can be reloaded by file name.
func (m *Document) ControlFile(file *os.File) error {
	m.memoryLimit = loadChunksCapacity(m.seekable)
	m.store.setNewLoadChunks(m.memoryLimit)
	atomic.StoreInt32(&m.closed, 0)
	r, err := m.fileReader(file)
	if err != nil {
		atomic.StoreInt32(&m.closed, 1)
		log.Println(err)
	}
	atomic.StoreInt32(&m.store.eof, 0)

	go func() {
		reader := bufio.NewReader(r)
		for sc := range m.ctlCh {
			reader, err = m.controlFile(sc, reader)
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
	}()

	m.requestStart()
	return nil
}

// ControlReader is the controller for io.Reader.
// Assuming call from Exec. reload executes the argument function.
func (m *Document) ControlReader(r io.Reader, reload func() *bufio.Reader) error {
	m.memoryLimit = loadChunksCapacity(false)
	m.store.setNewLoadChunks(m.memoryLimit)
	m.seekable = false
	reader := bufio.NewReader(r)

	go func() {
		var err error
		for sc := range m.ctlCh {
			reader, err = m.controlReader(sc, reader, reload)
			if err != nil {
				log.Println(sc.request, err)
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
		log.Println("close ctlCh")
	}()

	m.requestStart()
	return nil
}

// ControlLog controls log.
// ControlLog is only supported reload.
func (m *Document) ControlLog() error {
	go func() {
		for sc := range m.ctlCh {
			m.controlLog(sc)
			if sc.done != nil {
				sc.done <- true
				close(sc.done)
			}
		}
		log.Println("close m.ctlCh")
	}()
	return nil
}

// controlFile controls file read and loads in chunks.
// controlFile receives and executes request.
func (m *Document) controlFile(sc controlSpecifier, reader *bufio.Reader) (*bufio.Reader, error) {
	if atomic.LoadInt32(&m.closed) == 1 && sc.request != requestReload {
		return nil, fmt.Errorf("%w %s", ErrAlreadyClose, sc.request)
	}
	var err error
	switch sc.request {
	case requestStart:
		return m.firstRead(reader)
	case requestContinue:
		if !m.store.isContinueRead(m.memoryLimit) {
			return reader, nil
		}
		if m.seekable && (m.FollowMode || m.FollowAll) {
			if atomic.LoadInt32(&m.store.eof) == 0 && atomic.LoadInt32(&m.tmpFollow) == 0 {
				return m.tmpRead(reader)
			}
		}
		return m.continueRead(reader)
	case requestFollow:
		// Remove the last line from the cache as it may be appended.
		m.cache.Remove(m.BufEndNum() - 1)
		return m.followRead(reader)
	case requestLoad:
		return m.loadRead(reader, sc.chunkNum)
	case requestSearch:
		return m.searchRead(reader, sc.chunkNum, sc.searcher)
	case requestReload:
		reader, err = m.reloadRead(reader)
		m.requestStart()
		return reader, err
	case requestClose:
		err = m.close()
		return reader, err
	default:
		panic(fmt.Sprintf("unexpected %s", sc.request))
	}
}

// controlReader controls io.Reader.
// controlReader receives and executes request.
func (m *Document) controlReader(sc controlSpecifier, reader *bufio.Reader, reload func() *bufio.Reader) (*bufio.Reader, error) {
	switch sc.request {
	case requestStart:
		// controlReader is the same for first and continue.
		return m.continueRead(reader)
	case requestContinue:
		return m.continueRead(reader)
	case requestLoad:
		// Since controlReader is loaded outside, it only evicts.
		m.store.evictChunksMem(sc.chunkNum)
	case requestReload:
		if reload != nil {
			log.Println("reload")
			reader = reload()
			m.requestStart()
		}
	default:
		panic(fmt.Sprintf("unexpected %s", sc.request))
	}
	return reader, nil
}

// controlLog controls log.
// controlLog receives and executes request.
func (m *Document) controlLog(sc controlSpecifier) {
	switch sc.request {
	case requestLoad:
	case requestFollow:
	case requestReload:
		m.reset()
	default:
		panic(fmt.Sprintf("unexpected %s", sc.request))
	}
}

// requestStart send instructions to start reading.
func (m *Document) requestStart() {
	go func() {
		m.ctlCh <- controlSpecifier{
			request: requestStart,
		}
	}()
}

// requestContinue sends instructions to continue reading.
func (m *Document) requestContinue() {
	go func() {
		m.ctlCh <- controlSpecifier{
			request: requestContinue,
		}
	}()
}

// requestLoad sends instructions to load chunks into memory.
func (m *Document) requestLoad(chunkNum int) {
	sc := controlSpecifier{
		request:  requestLoad,
		chunkNum: chunkNum,
		done:     make(chan bool),
	}
	m.ctlCh <- sc
	<-sc.done
}

// requestSearch sends instructions to load chunks into memory.
func (m *Document) requestSearch(chunkNum int, searcher Searcher) bool {
	sc := controlSpecifier{
		request:  requestSearch,
		searcher: searcher,
		chunkNum: chunkNum,
		done:     make(chan bool),
	}
	m.ctlCh <- sc
	return <-sc.done
}

// requestClose sends instructions to close the file.
func (m *Document) requestClose() bool {
	atomic.StoreInt32(&m.store.readCancel, 1)
	defer atomic.StoreInt32(&m.store.readCancel, 0)
	sc := controlSpecifier{
		request: requestClose,
		done:    make(chan bool),
	}

	m.ctlCh <- sc
	return <-sc.done
}

// requestReload sends instructions to reload the file.
func (m *Document) requestReload() {
	sc := controlSpecifier{
		request: requestReload,
		done:    make(chan bool),
	}
	m.ctlCh <- sc
	<-sc.done
}
