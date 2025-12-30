package oviewer

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"golang.org/x/term"
)

// pipeBuffer pipes the buffer content to the specified command.
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

// runPipeCommand runs the command with buffer content as stdin.
func (root *Root) runPipeCommand(cmdStr string) error {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = getShell()
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", cmdStr)
	} else {
		cmd = exec.Command(shell, "-c", cmdStr)
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	// Export buffer content to the command's stdin.
	if root.Doc.PlainMode {
		if err := root.Doc.ExportPlain(stdin, root.Doc.BufStartNum(), root.Doc.BufEndNum()); err != nil {
			stdin.Close()
			return fmt.Errorf("failed to export buffer: %w", err)
		}
	} else {
		if err := root.Doc.Export(stdin, root.Doc.BufStartNum(), root.Doc.BufEndNum()); err != nil {
			stdin.Close()
			return fmt.Errorf("failed to export buffer: %w", err)
		}
	}
	stdin.Close()

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("command failed: %w", err)
	}

	return nil
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
