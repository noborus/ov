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
	docout, docerr, err := newOutErrDocument()
	if err != nil {
		return nil, err
	}

	go finishCommand(docout, docerr)

	so, se, err := commandStart(command)
	if err != nil {
		return nil, err
	}

	docout.Caption = "(" + command.Args[0] + ")" + docout.FileName
	err = docout.ReadAll(so)
	if err != nil {
		log.Printf("%s", err)
	}
	docerr.Caption = "(" + command.Args[0] + ")" + docerr.FileName
	err = docerr.ReadAll(se)
	if err != nil {
		log.Printf("%s", err)
	}
	return NewOviewer(docout, docerr)
}

func newOutErrDocument() (*Document, *Document, error) {
	docout, err := NewDocument()
	if err != nil {
		return nil, nil, err
	}
	docout.FileName = "STDOUT"
	docout.preventReload = true

	docerr, err := NewDocument()
	if err != nil {
		return nil, nil, err
	}
	docerr.FileName = "STDERR"
	docerr.preventReload = true

	return docout, docerr, nil
}

func commandStart(command *exec.Cmd) (io.Reader, io.Reader, error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		command.Stdin = os.Stdin
	}

	// STDOUT
	outReader, err := command.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}
	var so io.Reader = outReader
	if STDOUTPIPE != nil {
		so = io.TeeReader(so, STDOUTPIPE)
	}

	// STDERR
	errReader, err := command.StderrPipe()
	if err != nil {
		return nil, nil, err
	}
	var se io.Reader = errReader
	if STDERRPIPE != nil {
		se = io.TeeReader(se, STDERRPIPE)
	}

	if err := command.Start(); err != nil {
		return nil, nil, err
	}
	return so, se, nil
}

func finishCommand(docout *Document, docerr *Document) {
	<-docout.eofCh
	<-docerr.eofCh
	atomic.StoreInt32(&docout.changed, 1)
	atomic.StoreInt32(&docerr.changed, 1)
	atomic.StoreInt32(&docout.closed, 1)
	atomic.StoreInt32(&docerr.closed, 1)
}
