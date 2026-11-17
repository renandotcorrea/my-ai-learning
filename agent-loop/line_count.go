package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

var LinesCountTool = Tool{
	Name:        "lines_count",
	Description: "Counts the number of lines in the specified file. Provide the full path to the file you want to count lines for.",
	InputSchema: map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "The path to the file to count lines for.",
			},
		},
	},
	Function: LinesCount,
}

type LinesCountInput struct {
	Path string `json:"path"`
}

func LinesCount(input json.RawMessage) (string, error) {
	var LinesCountInput LinesCountInput
	if err := json.Unmarshal(input, &LinesCountInput); err != nil {
		return "", err
	}

	path := LinesCountInput.Path
	if path == "" {
		path = "."
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(content), "\n")
	return fmt.Sprintf("%d", len(lines)), nil
}
