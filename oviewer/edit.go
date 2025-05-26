package oviewer

import (
	"context"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/google/shlex"
)

// DefaultEditor is the fallback editor to use if no other editor is specified.
const DefaultEditor = "vim +%d %f"

// edit suspends the current screen display and runs the edit.
// It will return when you exit the edit.
func (root *Root) edit(context.Context) {
	fileName, err := root.editFileName()
	if err != nil {
		root.setMessageLog(err.Error())
		return
	}

	if err := root.Screen.Suspend(); err != nil {
		root.setMessageLog(err.Error())
		return
	}
	num := max(root.Doc.topLN+root.Doc.firstLine(), 0)
	log.Println("num", num)
	defer func() {
		log.Println("Resume from editor")
		if err := root.Screen.Resume(); err != nil {
			log.Println(err)
		}
		// Reload the document after editing.
		if root.Doc.seekable {
			root.reload(root.Doc)
			root.sendGoto(num + 1)
		}
	}()
	editor := root.identifyEditor()

	numStr := strconv.Itoa(num)
	command, args := replaceEditorArgs(editor, numStr, fileName)

	log.Println("Editing with command:", command, "and args:", args)
	c := exec.Command(command, args...)
	c.Stdin = os.Stdout
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		root.setMessageLog(err.Error())
	}
}

// editFileName determines the file name to edit based on the document's properties.
// If the document is seekable, it uses the original file name.
// If not, it saves the document to a temporary file and returns that file name.
func (root *Root) editFileName() (string, error) {
	fileName := root.Doc.FileName
	if root.Doc.seekable {
		return fileName, nil
	}

	log.Println("Editing non-seekable document")
	saveTempFile, err := root.saveTempFile()
	if err != nil {
		return "", err
	}
	log.Println("Editing non-seekable document, using temporary file:", saveTempFile)
	return saveTempFile, nil
}

// saveTempFile saves the current document to a temporary file and returns its name.
func (root *Root) saveTempFile() (string, error) {
	tempFile, err := os.CreateTemp("", "ov-edit-*.txt")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	if err := root.Doc.Export(tempFile, root.Doc.BufStartNum(), root.Doc.BufEndNum()); err != nil {
		log.Printf("Failed to export document to temporary file: %v", err)
		return "", err
	}
	fileName := tempFile.Name()
	os.Chmod(fileName, 0o400) // Read-only permission

	return fileName, nil
}

// identifyEditor determines the editor to use based on environment variables and configuration.
func (root *Root) identifyEditor() string {
	if ovedit := os.Getenv("OVEDIT"); ovedit != "" {
		return ovedit
	}

	if editor := root.Config.Editor; editor != "" {
		return editor
	}

	if visual := os.Getenv("VISUAL"); visual != "" {
		return visual
	}

	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}

	if sensibleEditor, _ := exec.LookPath("sensible-editor"); sensibleEditor != "" {
		return sensibleEditor
	}

	return DefaultEditor
}

// replaceEditorArgs replaces %d with numStr and %f with fileName in the editor command string.
// If %f is not present, fileName is appended at the end.
func replaceEditorArgs(editorCmd, numStr, fileName string) (string, []string) {
	args, err := shlex.Split(editorCmd)
	if err != nil {
		log.Printf("Failed to parse editor command '%s': %v", editorCmd, err)
		return DefaultEditor, []string{fileName}
	}

	if len(args) == 0 {
		log.Printf("Editor command '%s' resulted in an empty argument list", editorCmd)
		return DefaultEditor, []string{fileName}
	}

	hasFile := false
	command := args[0]
	if len(args) <= 1 {
		return command, []string{fileName}
	}

	args = args[1:]
	for i, arg := range args {
		// temporarily replace %d and %f to avoid confusion with %%d and %%f.
		arg = strings.ReplaceAll(arg, "%%d", "<<PERCENT_D>>")
		arg = strings.ReplaceAll(arg, "%%f", "<<PERCENT_F>>")
		// replace %d and %f with numStr and fileName.
		arg = strings.ReplaceAll(arg, "%d", numStr)
		if strings.Contains(arg, "%f") {
			arg = strings.ReplaceAll(arg, "%f", fileName)
			hasFile = true
		}
		// replace the temporary placeholders back to %d and %f.
		arg = strings.ReplaceAll(arg, "<<PERCENT_D>>", "%d")
		arg = strings.ReplaceAll(arg, "<<PERCENT_F>>", "%f")
		args[i] = arg
	}
	if !hasFile {
		args = append(args, fileName)
	}
	return command, args
}
