package oviewer

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync/atomic"

	"golang.org/x/term"
)

type Command struct {
	args    []string
	command *exec.Cmd
	stdout  io.Reader
	stderr  io.Reader
	docout  *Document
	docerr  *Document
}

func NewCommand(args ...string) *Command {
	return &Command{
		args: args,
	}
}

func (cmd *Command) Exec() (*Root, error) {
	cmd.command = exec.Command(cmd.args[0], cmd.args[1:]...)
	var err error
	cmd.docout, cmd.docerr, err = newOutErrDocument()
	if err != nil {
		return nil, err
	}

	so, se, err := commandStart(cmd.command)
	if err != nil {
		return nil, err
	}

	cmd.stdout = so
	cmd.stderr = se

	cmd.docout.Caption = "(" + cmd.command.Args[0] + ")" + cmd.docout.FileName
	atomic.StoreInt32(&cmd.docout.closed, 0)
	atomic.StoreInt32(&cmd.docerr.closed, 0)
	cmd.docout.seekable = false
	cmd.docerr.seekable = false
	cmd.docout.formfeedTime = true
	cmd.docerr.formfeedTime = true
	err = cmd.docout.ControlReader(so, cmd.Reload)
	if err != nil {
		log.Printf("%s", err)
	}
	cmd.docerr.Caption = "(" + cmd.command.Args[0] + ")" + cmd.docerr.FileName
	err = cmd.docerr.ControlReader(se, cmd.stderrReload)
	if err != nil {
		log.Printf("%s", err)
	}
	return NewOviewer(cmd.docout, cmd.docerr)
}

func (cmd *Command) Wait() {
	if cmd.command == nil || cmd.command.Process == nil {
		return
	}
	atomic.StoreInt32(&cmd.docout.closed, 1)
	atomic.StoreInt32(&cmd.docerr.closed, 1)
	if err := cmd.command.Process.Kill(); err != nil {
		log.Println(err)
	}
	if err := cmd.command.Wait(); err != nil {
		log.Println(err)
	}
}

func (cmd *Command) Reload() *bufio.Reader {
	cmd.Wait()
	if cmd.docout.WatchMode {
		cmd.docout.appendFormFeed(cmd.docout.lastChunk())
	} else {
		cmd.docout.reset()
	}
	cmd.command = exec.Command(cmd.args[0], cmd.args[1:]...)
	so, se, err := commandStart(cmd.command)
	if err != nil {
		log.Println(err)
		str := fmt.Sprintf("command error: %s", err)
		reader := bufio.NewReader(strings.NewReader(str))
		return reader
	}
	cmd.stdout = so
	cmd.stderr = se

	sc := controlSpecifier{
		request: requestReload,
		done:    make(chan bool),
	}
	log.Println("stderr reload send")
	cmd.docerr.ctlCh <- sc
	<-sc.done
	atomic.StoreInt32(&cmd.docerr.readCancel, 0)
	log.Println("stderr receive done")

	return bufio.NewReader(so)
}

func (cmd *Command) stderrReload() *bufio.Reader {
	if !cmd.docout.WatchMode {
		cmd.docerr.reset()
	} else {
		cmd.docerr.appendFormFeed(cmd.docerr.lastChunk())
	}

	return bufio.NewReader(cmd.stderr)
}

// ExecCommand return the structure of oviewer.
// ExecCommand executes the command and opens stdout/stderr as document.
func ExecCommand(command *exec.Cmd) (*Root, error) {
	docout, docerr, err := newOutErrDocument()
	if err != nil {
		return nil, err
	}

	so, se, err := commandStart(command)
	if err != nil {
		return nil, err
	}

	docout.Caption = "(" + command.Args[0] + ")" + docout.FileName
	err = docout.ControlReader(so, nil)
	if err != nil {
		log.Printf("%s", err)
	}
	docerr.Caption = "(" + command.Args[0] + ")" + docerr.FileName
	err = docerr.ControlReader(se, nil)
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

	docerr, err := NewDocument()
	if err != nil {
		return nil, nil, err
	}
	docerr.FileName = "STDERR"

	return docout, docerr, nil
}

func commandStart(command *exec.Cmd) (io.Reader, io.Reader, error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		command.Stdin = os.Stdin
	}

	// STDOUT
	outReader, err := command.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("stdout pipe error: %w", err)
	}
	var so io.Reader = outReader
	if STDOUTPIPE != nil {
		so = io.TeeReader(so, STDOUTPIPE)
	}

	// STDERR
	errReader, err := command.StderrPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("stderr pipe error: %w", err)
	}
	var se io.Reader = errReader
	if STDERRPIPE != nil {
		se = io.TeeReader(se, STDERRPIPE)
	}

	if err := command.Start(); err != nil {
		return nil, nil, fmt.Errorf("command start error: %w", err)
	}
	return so, se, nil
}
