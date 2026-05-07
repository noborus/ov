// gen-keybind generates config YAML files from a shared template and keybind presets.
// Usage: go run ./tools/gen-keybind
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/noborus/ov/oviewer"
)

const (
	templateFileName = "ov.yaml.template"
	keybindTag       = "{{KEYBIND}}"
)

type targetConfig struct {
	path    string
	keyBind oviewer.KeyBind
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	repoRoot, err := findRepoRoot()
	if err != nil {
		return err
	}

	templatePath := filepath.Join(repoRoot, templateFileName)
	templateData, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("read %s: %w", templatePath, err)
	}

	targets := []targetConfig{
		{path: filepath.Join(repoRoot, "ov.yaml"), keyBind: oviewer.DefaultKeyBinds()},
		{path: filepath.Join(repoRoot, "ov-less.yaml"), keyBind: oviewer.LessKeyBinds()},
	}

	for _, target := range targets {
		generated := generateKeyBindSection(target.keyBind)
		result, err := applyTemplate(string(templateData), generated)
		if err != nil {
			return fmt.Errorf("render %s: %w", target.path, err)
		}

		if err := os.WriteFile(target.path, []byte(result), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", target.path, err)
		}

		fmt.Printf("Updated %s (%d actions)\n", target.path, len(target.keyBind))
	}

	return nil
}

func findRepoRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getwd: %w", err)
	}

	dir := wd
	for {
		goMod := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goMod); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found from %s", wd)
		}
		dir = parent
	}
}

func generateKeyBindSection(keyBind oviewer.KeyBind) string {
	// Sort keys for deterministic output.
	keys := make([]string, 0, len(keyBind))
	for k := range keyBind {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	var sb strings.Builder
	sb.WriteString("KeyBind:\n")
	for _, action := range keys {
		vals := keyBind[action]
		sb.WriteString(fmt.Sprintf("    %s:\n", action))
		for _, v := range vals {
			sb.WriteString(fmt.Sprintf("        - %q\n", v))
		}
	}
	return sb.String()
}

func applyTemplate(content, generated string) (string, error) {
	if !strings.Contains(content, keybindTag) {
		return "", fmt.Errorf("template tag %q not found", keybindTag)
	}

	return strings.Replace(content, keybindTag, generated, 1), nil
}
