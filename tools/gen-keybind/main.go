// gen-keybind generates the KeyBind section of ov.yaml from DefaultKeyBinds().
// Usage: go run ./tools/gen-keybind
package main

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/noborus/ov/oviewer"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	const configFile = "ov.yaml"

	keyBind := oviewer.DefaultKeyBinds()

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
	generated := sb.String()

	// Read ov.yaml and replace the KeyBind section.
	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("read %s: %w", configFile, err)
	}

	result, err := replaceKeyBindSection(string(data), generated)
	if err != nil {
		return err
	}

	if err := os.WriteFile(configFile, []byte(result), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", configFile, err)
	}

	fmt.Printf("Updated KeyBind section in %s (%d actions)\n", configFile, len(keys))
	return nil
}

// replaceKeyBindSection replaces the content between "KeyBind:" and the next
// top-level key (or EOF) with the generated YAML.
func replaceKeyBindSection(content, generated string) (string, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))

	var before, after strings.Builder
	inKeybind := false
	found := false

	for scanner.Scan() {
		line := scanner.Text()
		if !inKeybind && !found {
			if line == "KeyBind:" {
				inKeybind = true
				found = true
				continue
			}
			before.WriteString(line)
			before.WriteByte('\n')
			continue
		}
		if inKeybind {
			// A top-level key starts with a non-space, non-comment character.
			if len(line) > 0 && line[0] != ' ' && line[0] != '#' && line[0] != '\n' {
				inKeybind = false
				after.WriteString(line)
				after.WriteByte('\n')
			}
			continue
		}
		after.WriteString(line)
		after.WriteByte('\n')
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("scanning config: %w", err)
	}
	if !found {
		return "", fmt.Errorf("KeyBind: section not found in ov.yaml")
	}

	return before.String() + generated + "\n" + after.String(), nil
}
