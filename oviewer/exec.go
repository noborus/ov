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
	"time"

	"github.com/kr/pty"
	"golang.org/x/term"
)

// Command is the structure of the command.
type Command struct {
	stdout  io.Reader
	stderr  io.Reader
	command *exec.Cmd
	docout  *Document
	docerr  *Document
	args    []string
}

// NewCommand return the structure of Command.
func NewCommand(args ...string) *Command {
	return &Command{
		args: args,
	}
}

// Exec return the structure of oviewer.
func (cmd *Command) Exec() (*Root, error) {
	docout, docerr, err := newOutErrDocument()
	if err != nil {
		return nil, err
	}
	cmd.docout = docout
	cmd.docerr = docerr

	cmd.command = exec.Command(cmd.args[0], cmd.args[1:]...)
	so, se, err := commandStart(cmd.command)
	if err != nil {
		return nil, err
	}
	cmd.stdout = so
	cmd.stderr = se

	cmd.docout.Caption = "(" + cmd.command.Args[0] + ")" + cmd.docout.FileName
	cmd.docerr.Caption = "(" + cmd.command.Args[0] + ")" + cmd.docerr.FileName
	atomic.StoreInt32(&cmd.docout.closed, 0)
	atomic.StoreInt32(&cmd.docerr.closed, 0)
	cmd.docout.seekable = false
	cmd.docerr.seekable = false
	cmd.docout.store.formfeedTime = true
	cmd.docerr.store.formfeedTime = true

	if err = cmd.docout.ControlReader(so, cmd.Reload); err != nil {
		log.Printf("%s", err)
	}
	if err = cmd.docerr.ControlReader(se, cmd.stderrReload); err != nil {
		log.Printf("%s", err)
	}
	return NewOviewer(cmd.docout, cmd.docerr)
}

// Wait waits for the command to exit.
func (cmd *Command) Wait() {
	if cmd.command == nil || cmd.command.Process == nil {
		return
	}

	atomic.StoreInt32(&cmd.docout.closed, 1)
	atomic.StoreInt32(&cmd.docerr.closed, 1)

	// Kill the command if it hasn't exited yet.
	if err := cmd.command.Process.Kill(); err != nil {
		log.Println(err)
	}
	// Wait for the command to exit.
	if err := cmd.command.Wait(); err != nil {
		log.Println(err)
	}
}

// Reload restarts the command.
func (cmd *Command) Reload() *bufio.Reader {
	cmd.Wait()
	if cmd.docout.WatchMode {
		s := cmd.docout.store
		s.appendFormFeed(s.chunkForAdd(false, s.size))
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

	cmd.docerr.requestReload()
	atomic.StoreInt32(&cmd.docerr.store.readCancel, 0)
	log.Println("stderr receive done")

	return bufio.NewReader(so)
}

// stderrReload is called when the command is restarted.
func (cmd *Command) stderrReload() *bufio.Reader {
	if !cmd.docout.WatchMode {
		cmd.docerr.reset()
	} else {
		s := cmd.docerr.store
		s.appendFormFeed(s.chunkForAdd(false, s.size))
	}

	return bufio.NewReader(cmd.stderr)
}

// ExecCommand return the structure of oviewer.
// ExecCommand executes the command and opens stdout/stderr as document.
// obsolete: use NewCommand and Exec instead.
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

// newOutErrDocument returns the structure of Document.
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

// commandStart starts the command.
func commandStart(command *exec.Cmd) (io.Reader, io.Reader, error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		command.Stdin = os.Stdin
	}

	// STDOUT
	stdout, outReader, err := pty.Open()
	if err != nil {
		return nil, nil, fmt.Errorf("pty open error: %w", err)
	}
	command.Stdout = stdout
	var so io.Reader = outReader
	if STDOUTPIPE != nil {
		so = io.TeeReader(so, STDOUTPIPE)
	}

	// STDERR
	stderr, errReader, err := pty.Open()
	if err != nil {
		return nil, nil, fmt.Errorf("pty open error: %w", err)
	}
	command.Stderr = stderr
	var se io.Reader = errReader
	if STDERRPIPE != nil {
		se = io.TeeReader(se, STDERRPIPE)
	}

	if err := command.Start(); err != nil {
		return nil, nil, fmt.Errorf("command start error: %w", err)
	}
	go func() {
		if err := command.Wait(); err != nil {
			log.Println(err)
		}
		time.Sleep(100 * time.Millisecond)
		log.Println("command done close")
		stdout.Close()
		stderr.Close()
	}()

	return so, se, nil
}
