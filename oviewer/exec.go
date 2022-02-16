package oviewer

import (
	"io"
	"log"
	"os"
	"os/exec"
	"sync/atomic"

	"golang.org/x/term"
)

// ExecCommand return the structure of oviewer.
// ExecCommand executes the command and opens stdout/stderr as document.
func ExecCommand(command *exec.Cmd) (*Root, error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		command.Stdin = os.Stdin
	}

	docout, err := NewDocument()
	if err != nil {
		return nil, err
	}
	docout.FileName = "STDOUT"
	outReader, err := command.StdoutPipe()
	if err != nil {
		return nil, err
	}
	docout.preventReload = true

	docerr, err := NewDocument()
	if err != nil {
		return nil, err
	}
	docerr.FileName = "STDERR"
	errReader, err := command.StderrPipe()
	if err != nil {
		return nil, err
	}
	docerr.preventReload = true

	if err := command.Start(); err != nil {
		return nil, err
	}

	go func() {
		<-docout.eofCh
		atomic.StoreInt32(&docout.changed, 1)
		atomic.StoreInt32(&docerr.changed, 1)
		docout.FileName = "STDOUT(done)"
		docerr.FileName = "STDERR(done)"
		atomic.StoreInt32(&docout.closed, 1)
		atomic.StoreInt32(&docerr.closed, 1)
	}()

	var reader io.Reader
	reader = outReader
	if STDOUTPIPE != nil {
		reader = io.TeeReader(reader, STDOUTPIPE)
	}

	err = docout.ReadAll(reader)
	if err != nil {
		log.Printf("%s", err)
	}

	var errreader io.Reader
	errreader = errReader
	if STDERRPIPE != nil {
		errreader = io.TeeReader(errreader, STDERRPIPE)
	}
	err = docerr.ReadAll(errreader)
	if err != nil {
		log.Printf("%s", err)
	}

	return NewOviewer(docout, docerr)
}
