package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

var EditFileTool = Tool{
	Name:        "edit_file",
	Description: "Apply a patch to a specific file. Provide the path to the file and the patch content. The type of edit can be 'replace', 'append', or 'prepend'.",
	InputSchema: map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "The path to the file to apply the patch to.",
			},
			"patch": map[string]any{
				"type":        "string",
				"description": "The patch content to apply to the file.",
			},
			"from_line": map[string]any{
				"type":        "integer",
				"description": "The starting line number of the patch.",
			},
			"to_line": map[string]any{
				"type":        "integer",
				"description": "The ending line number of the patch.",
			},
			"type": map[string]any{
				"type":        "string",
				"description": "The type of edit to apply. Can be 'replace', 'append', or 'prepend'. The 'replace' type requires 'from_line' and 'to_line' to specify the range of lines to replace. The 'append' and 'prepend' types do not require line numbers.",
			},
		},
	},
	Function: EditFile,
}

type EditFileInput struct {
	Path     string `json:"path"`
	Patch    string `json:"patch"`
	FromLine int    `json:"from_line"`
	ToLine   int    `json:"to_line"`
	Type     string `json:"type"`
}

func EditFile(input json.RawMessage) (string, error) {
	var EditFileInput EditFileInput
	if err := json.Unmarshal(input, &EditFileInput); err != nil {
		return "", err
	}

	path := EditFileInput.Path
	if path == "" {
		path = "."
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(content), "\n")
	patchLines := strings.Split(EditFileInput.Patch, "\n")
	switch EditFileInput.Type {
	case "replace":
		if EditFileInput.FromLine < 1 || EditFileInput.ToLine < EditFileInput.FromLine || EditFileInput.ToLine > len(lines) {
			return "", fmt.Errorf("invalid line range %d-%d for file with %d lines", EditFileInput.FromLine, EditFileInput.ToLine, len(lines))
		}
		lines = append(lines[:EditFileInput.FromLine-1], append(patchLines, lines[EditFileInput.ToLine:]...)...)
	case "append":
		lines = append(lines, patchLines...)
	case "prepend":
		lines = append(patchLines, lines...)
	default:
		return "", fmt.Errorf("unsupported edit type %q", EditFileInput.Type)
	}

	content = []byte(strings.Join(lines, "\n"))
	if err := os.WriteFile(path, content, 0644); err != nil {
		return "", err
	}
	return string(content), nil
}
