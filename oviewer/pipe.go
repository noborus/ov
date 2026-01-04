package oviewer

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"golang.org/x/term"
)

// pipeBuffer pipes the buffer content to the specified command using the terminal.
func (root *Root) pipeBuffer(input string) {
	cmdStr := strings.TrimSpace(input)
	if cmdStr == "" {
		root.setMessage("pipe command is empty")
		return
	}

	log.Printf("Pipe buffer to command: %s\n", cmdStr)
	if err := root.Screen.Suspend(); err != nil {
		root.setMessageLog(err.Error())
		return
	}
	defer func() {
		log.Println("Resume from pipe")
		if err := root.Screen.Resume(); err != nil {
			log.Println(err)
		}
	}()

	if err := root.runPipeCommand(cmdStr); err != nil {
		fmt.Fprintf(os.Stderr, "pipe command error: %s\n", err)
	}

	fmt.Println("press any key to continue...")
	if err := waitForKey(); err != nil {
		log.Printf("waitForKey error: %s\n", err)
	}
}

// pipeBufferDoc pipes the buffer content to the specified command and opens the output as a new document.
func (root *Root) pipeBufferDoc(ctx context.Context, input string) {
	cmdStr := strings.TrimSpace(input)
	if cmdStr == "" {
		root.setMessage("pipe command is empty")
		return
	}

	sourceDoc := root.Doc
	cmd := pipeCommand(cmdStr)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		root.setMessageLogf("pipe command error: %s", err)
		return
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		root.setMessageLogf("pipe command error: %s", err)
		return
	}
	if root.logDoc != nil {
		cmd.Stderr = root.logDoc
	} else {
		cmd.Stderr = io.Discard
	}

	if err := cmd.Start(); err != nil {
		root.setMessageLogf("pipe command error: %s", err)
		return
	}

	render, err := renderDoc(sourceDoc, stdout)
	if err != nil {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		_ = cmd.Wait()
		root.setMessageLogf("pipe command error: %s", err)
		return
	}

	render.documentType = DocPipe
	render.Caption = "pipe:" + cmdStr
	render.RunTimeSettings = sourceDoc.RunTimeSettings
	render.regexpCompile()
	render.conv = render.converterType(render.Converter)
	root.insertDocument(ctx, root.CurrentDoc, render)
	root.setMessageLogf("pipe buffer: %s", cmdStr)

	go func() {
		if err := exportBuffer(sourceDoc, stdin); err != nil {
			log.Printf("pipe buffer export error: %s\n", err)
		}
		if err := stdin.Close(); err != nil {
			log.Printf("pipe buffer stdin close error: %s\n", err)
		}
		if err := cmd.Wait(); err != nil {
			log.Printf("pipe command error: %s\n", err)
		}
	}()
}

// runPipeCommand runs the command with buffer content as stdin.
func (root *Root) runPipeCommand(cmdStr string) error {
	cmd := pipeCommand(cmdStr)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	if err := exportBuffer(root.Doc, stdin); err != nil {
		_ = stdin.Close()
		return fmt.Errorf("failed to export buffer: %w", err)
	}
	if err := stdin.Close(); err != nil {
		return fmt.Errorf("failed to close stdin: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("command failed: %w", err)
	}

	return nil
}

func pipeCommand(cmdStr string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.Command("cmd", "/c", cmdStr)
	}
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = getShell()
	}
	return exec.Command(shell, "-c", cmdStr)
}

func exportBuffer(doc *Document, w io.Writer) error {
	start := doc.BufStartNum()
	end := doc.BufEndNum()
	if doc.PlainMode {
		return doc.ExportPlain(w, start, end)
	}
	return doc.Export(w, start, end)
}

// waitForKey waits for the user to press any key.
func waitForKey() error {
	tty, err := getTTY()
	if err != nil {
		return err
	}
	defer tty.Close()

	oldState, err := term.MakeRaw(int(tty.Fd()))
	if err != nil {
		return err
	}
	defer term.Restore(int(tty.Fd()), oldState)

	buf := make([]byte, 1)
	_, err = tty.Read(buf)
	return err
}
