package oviewer

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/creack/pty"
	"golang.org/x/term"
)

// virtual terminal size.
const (
	COLS = 80
	ROWS = 24
)

// Command is the structure of the command.
type Command struct {
	stdout io.Reader
	stderr io.Reader
	cmd    *exec.Cmd
	docout *Document
	docerr *Document
	args   []string
}

// NewCommand return the structure of Command.
func NewCommand(args ...string) *Command {
	return &Command{
		args: args,
	}
}

// Exec return the structure of oviewer.
func (command *Command) Exec() (*Root, error) {
	docout, docerr, err := newOutErrDocument()
	if err != nil {
		return nil, err
	}
	command.docout = docout
	command.docerr = docerr

	command.cmd = exec.Command(command.args[0], command.args[1:]...)
	so, se, err := commandStart(command.cmd)
	if err != nil {
		return nil, err
	}
	command.stdout = so
	command.stderr = se

	command.docout.Caption = "(" + command.cmd.Args[0] + ")" + command.docout.FileName
	command.docerr.Caption = "(" + command.cmd.Args[0] + ")" + command.docerr.FileName
	atomic.StoreInt32(&command.docout.closed, 0)
	atomic.StoreInt32(&command.docerr.closed, 0)
	command.docout.seekable = false
	command.docerr.seekable = false
	command.docout.store.formfeedTime = true
	command.docerr.store.formfeedTime = true

	if err = command.docout.ControlReader(so, command.Reload); err != nil {
		log.Printf("%s", err)
	}
	if err = command.docerr.ControlReader(se, command.stderrReload); err != nil {
		log.Printf("%s", err)
	}
	return NewOviewer(command.docout, command.docerr)
}

// Wait waits for the command to exit.
// Wait does not actually `wait` because it is Waiting in the goroutine at the start.
func (command *Command) Wait() {
	if command.cmd == nil || command.cmd.Process == nil {
		return
	}

	atomic.StoreInt32(&command.docout.closed, 1)
	atomic.StoreInt32(&command.docerr.closed, 1)

	// Kill the command if it hasn't exited yet.
	if err := command.cmd.Process.Kill(); err != nil {
		log.Println(err)
	}
}

// Reload restarts the command.
func (command *Command) Reload() *bufio.Reader {
	command.Wait()
	if command.docout.WatchMode {
		s := command.docout.store
		s.appendFormFeed(s.chunkForAdd(false, s.size))
	} else {
		command.docout.reset()
	}
	command.cmd = exec.Command(command.args[0], command.args[1:]...)
	so, se, err := commandStart(command.cmd)
	if err != nil {
		log.Println(err)
		str := fmt.Sprintf("command error: %s", err)
		reader := bufio.NewReader(strings.NewReader(str))
		return reader
	}
	command.stdout = so
	command.stderr = se

	command.docerr.requestReload()
	atomic.StoreInt32(&command.docerr.store.readCancel, 0)
	log.Println("stderr receive done")

	return bufio.NewReader(so)
}

// stderrReload is called when the command is restarted.
func (command *Command) stderrReload() *bufio.Reader {
	if !command.docout.WatchMode {
		command.docerr.reset()
	} else {
		s := command.docerr.store
		s.appendFormFeed(s.chunkForAdd(false, s.size))
	}

	return bufio.NewReader(command.stderr)
}

// ExecCommand return the structure of oviewer.
// ExecCommand executes the command and opens stdout/stderr as document.
// obsolete: use NewCommand and Exec instead.
func ExecCommand(cmd *exec.Cmd) (*Root, error) {
	docout, docerr, err := newOutErrDocument()
	if err != nil {
		return nil, err
	}

	so, se, err := commandStart(cmd)
	if err != nil {
		return nil, err
	}

	docout.Caption = "(" + cmd.Args[0] + ")" + docout.FileName
	err = docout.ControlReader(so, nil)
	if err != nil {
		log.Printf("%s", err)
	}
	docerr.Caption = "(" + cmd.Args[0] + ")" + docerr.FileName
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
func commandStart(cmd *exec.Cmd) (io.Reader, io.Reader, error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		cmd.Stdin = os.Stdin
	}
	return cmdOutput(cmd)
}

func cmdOutput(cmd *exec.Cmd) (io.Reader, io.Reader, error) {
	if runtime.GOOS == "windows" {
		return pipeOutput(cmd)
	}
	return ptyOutput(cmd)
}

// pipeOutput returns the stdout and stderr of the command.
// pipeOutput is used on Windows.
func pipeOutput(cmd *exec.Cmd) (io.Reader, io.Reader, error) {
	so, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("stdout pipe error: %w", err)
	}

	se, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("stderr pipe error: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("command start error: %w", err)
	}

	return so, se, nil
}

// ptyOutput returns the stdout and stderr of the command.
func ptyOutput(cmd *exec.Cmd) (io.Reader, io.Reader, error) {
	// STDOUT
	stdout, outReader, err := pty.Open()
	if err != nil {
		return nil, nil, fmt.Errorf("pty open error: %w", err)
	}
	if err := pty.Setsize(stdout, &pty.Winsize{Cols: COLS, Rows: ROWS}); err != nil {
		return nil, nil, fmt.Errorf("pty setsize error: %w", err)
	}
	cmd.Stdout = stdout
	var so io.Reader = outReader
	if STDOUTPIPE != nil {
		so = io.TeeReader(so, STDOUTPIPE)
	}

	// STDERR
	stderr, errReader, err := pty.Open()
	if err != nil {
		return nil, nil, fmt.Errorf("pty open error: %w", err)
	}
	cmd.Stderr = stderr
	var se io.Reader = errReader
	if STDERRPIPE != nil {
		se = io.TeeReader(se, STDERRPIPE)
	}
	if err := cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("command start error: %w", err)
	}

	go func() {
		if err := cmd.Wait(); err != nil {
			log.Printf("wait: %s", err)
		}
		time.Sleep(100 * time.Millisecond)
		stdout.Close()
		stderr.Close()
	}()

	return so, se, nil
}
