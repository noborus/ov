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

// controlSpecifier represents a control request.
type controlSpecifier struct {
	searcher Searcher
	request  request
	chunkNum int
	done     chan bool
}

type request string

const (
	requestStart    = "start"
	requestContinue = "read"
	requestFollow   = "follow"
	requestClose    = "close"
	requestReload   = "reload"
	requestLoad     = "load"
	requestSearch   = "search"
)

// ControlFile controls file read and loads in chunks.
// ControlFile can be reloaded by file name.
func (m *Document) ControlFile(file *os.File) error {
	if m.seekable {
		m.loadedChunks.Resize(FileLoadChunksLimit + 1)
	} else {
		m.loadedChunks.Resize(LoadChunksLimit + 1)
	}

	go func() {
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
		log.Println("close m.ctlCh")
	}()

	m.requestStart()
	return nil
}

func (m *Document) control(sc controlSpecifier, reader *bufio.Reader) (*bufio.Reader, error) {
	if atomic.LoadInt32(&m.closed) == 1 && sc.request != requestReload {
		return nil, fmt.Errorf("%w %s", ErrAlreadyClose, sc.request)
	}
	var err error
	switch sc.request {
	case requestStart:
		reader, err = m.firstRead(reader)
		if !m.BufEOF() {
			m.requestContinue()
		}
		return reader, err
	case requestContinue:
		if !m.seekable {
			if LoadChunksLimit > 0 && m.loadedChunks.Len() >= LoadChunksLimit {
				log.Println("stop read", m.loadedChunks.Len(), LoadChunksLimit)
				return reader, nil
			}
			chunkNum := m.lastChunkNum()
			if chunkNum != 0 {
				m.loadedChunks.PeekOrAdd(chunkNum, struct{}{})
			}
		}
		reader, err = m.continueRead(reader)
		if !m.BufEOF() {
			m.requestContinue()
		}
		return reader, err
	case requestFollow:
		reader, err = m.followRead(reader)
		if !m.BufEOF() {
			m.requestContinue()
		}
		return reader, err
	case requestLoad:
		m.currentChunk = sc.chunkNum
		if m.seekable {
			if err := m.managesChunksFile(sc.chunkNum); err != nil {
				return reader, nil
			}
			if sc.chunkNum != 0 {
				m.loadedChunks.Add(sc.chunkNum, struct{}{})
			}
			return m.readChunk(reader, sc.chunkNum)
		}
		if !m.BufEOF() {
			if sc.chunkNum < m.lastChunkNum() {
				return reader, fmt.Errorf("%w %d", ErrAlreadyLoaded, sc.chunkNum)
			}
			m.managesChunksMem(sc.chunkNum)
			m.requestContinue()
		}
		return reader, nil
	case requestSearch:
		_, err = m.searchChunk(sc.chunkNum, sc.searcher)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return reader, nil
			}
			return reader, err
		}
		if err := m.managesChunksFile(sc.chunkNum); err != nil {
			return reader, nil
		}
		return m.readChunk(reader, sc.chunkNum)
	case requestReload:
		if !m.WatchMode {
			m.loadedChunks.Purge()
		}
		reader, err = m.reloadRead(reader)
		m.requestStart()
		return reader, err
	case requestClose:
		err = m.close()
		log.Println(err)
		return reader, err
	default:
		panic(fmt.Sprintf("unexpected %s", sc.request))
	}
}

// unloadChunk unloads the chunk from memory.
func (m *Document) unloadChunk(chunkNum int) {
	m.loadedChunks.Remove(chunkNum)
	m.chunks[chunkNum].lines = nil
}

// managesChunksFile manages Chunks of regular files.
// manage chunk eviction.
func (m *Document) managesChunksFile(chunkNum int) error {
	if chunkNum == 0 {
		return nil
	}
	for m.loadedChunks.Len() > FileLoadChunksLimit {
		k, _, _ := m.loadedChunks.GetOldest()
		if chunkNum != k {
			m.unloadChunk(k)
		}
	}

	chunk := m.chunks[chunkNum]
	if len(chunk.lines) != 0 || atomic.LoadInt32(&m.closed) != 0 {
		return fmt.Errorf("%w %d", ErrAlreadyLoaded, chunkNum)
	}
	return nil
}

// managesChunksMem manages Chunks other than regular files.
// The specified chunk is already in memory, so only the first chunk is unloaded.
// Change the start position after unloading.
func (m *Document) managesChunksMem(chunkNum int) {
	if chunkNum == 0 {
		return
	}
	if (LoadChunksLimit < 0) || (m.loadedChunks.Len() < LoadChunksLimit) {
		return
	}
	k, _, _ := m.loadedChunks.GetOldest()
	m.unloadChunk(k)
	m.mu.Lock()
	m.startNum = (k + 1) * ChunkSize
	m.mu.Unlock()
}

// ControlLog controls log.
// ControlLog is only supported reload.
func (m *Document) ControlLog() error {
	go func() {
		for sc := range m.ctlCh {
			switch sc.request {
			case requestLoad:
			case requestReload:
				m.reset()
			default:
				panic(fmt.Sprintf("unexpected %s", sc.request))
			}
			if sc.done != nil {
				close(sc.done)
			}
		}
		log.Println("close m.ctlCh")
	}()
	return nil
}

// ControlReader is the controller for io.Reader.
// Assuming call from Exec. reload executes the argument function.
func (m *Document) ControlReader(r io.Reader, reload func() *bufio.Reader) error {
	m.seekable = false
	reader := bufio.NewReader(r)
	go func() {
		var err error
		for sc := range m.ctlCh {
			switch sc.request {
			case requestStart:
				// controlReader is the same for first and continue.
				fallthrough
			case requestContinue:
				chunkNum := m.lastChunkNum()
				if chunkNum != 0 {
					m.loadedChunks.PeekOrAdd(chunkNum, struct{}{})
				}
				reader, err = m.continueRead(reader)
				if !m.BufEOF() {
					m.requestContinue()
				}
			case requestLoad:
				m.currentChunk = sc.chunkNum
				m.managesChunksMem(sc.chunkNum)
			case requestReload:
				if reload != nil {
					log.Println("reset")
					reader = reload()
					m.requestStart()
				}
			default:
				panic(fmt.Sprintf("unexpected %s", sc.request))
			}
			if err != nil {
				log.Printf("%v %s", sc.request, err)
			}
			if sc.done != nil {
				close(sc.done)
			}
		}
		log.Println("close ctlCh")
	}()

	m.requestStart()
	return nil
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
func (m *Document) requestClose() {
	atomic.StoreInt32(&m.readCancel, 1)
	sc := controlSpecifier{
		request: requestClose,
		done:    make(chan bool),
	}

	log.Println("close send")
	m.ctlCh <- sc
	<-sc.done
	atomic.StoreInt32(&m.readCancel, 0)
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
