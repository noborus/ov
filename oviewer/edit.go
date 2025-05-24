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
const DefaultEditor = "vi"

// edit suspends the current screen display and runs the edit.
// It will return when you exit the edit.
func (root *Root) edit(context.Context) {
	if err := root.Screen.Suspend(); err != nil {
		root.setMessageLog(err.Error())
		return
	}
	defer func() {
		log.Println("Resume from editor")
		if err := root.Screen.Resume(); err != nil {
			log.Println(err)
		}
	}()
	editor := root.identifyEditor()

	numStr := strconv.Itoa(root.Doc.topLN + root.Doc.firstLine())
	command, args := replaceEditorArgs(editor, numStr, root.Doc.FileName)

	log.Println("Editing with command:", command, "and args:", args)
	c := exec.Command(command, args...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		root.setMessageLog(err.Error())
	}
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

	return "vi"
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
