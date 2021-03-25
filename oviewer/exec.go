package oviewer

import (
	"log"
	"os"
	"os/exec"
	"sync/atomic"
)

// ExecCommand return the structure of oviewer.
// ExecCommand executes the command and opens stdout/stderr as document.
func ExecCommand(command *exec.Cmd) (*Root, error) {
	command.Stdin = os.Stdin
	docout, err := NewDocument()
	if err != nil {
		log.Fatal(err)
	}
	docout.FileName = "STDOUT"
	outReader, err := command.StdoutPipe()
	if err != nil {
		return nil, err
	}

	docerr, err := NewDocument()
	if err != nil {
		return nil, err
	}
	docerr.FileName = "STDERR"
	errReader, err := command.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := command.Start(); err != nil {
		return nil, err
	}

	go func() {
		<-docout.eofCh
		atomic.StoreInt32(&docout.changed, 1)
		atomic.StoreInt32(&docerr.changed, 1)
	}()

	err = docout.ReadAll(outReader)
	if err != nil {
		log.Printf("%s", err)
	}

	err = docerr.ReadAll(errReader)
	if err != nil {
		log.Printf("%s", err)
	}

	return NewOviewer(docout, docerr)
}
