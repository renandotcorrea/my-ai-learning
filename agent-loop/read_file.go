package main

import (
	"encoding/json"
	"os"
)

var ReadFileTool = Tool{
	Name:        "read_file",
	Description: "Reads the contents of the specified file. Provide the full path to the file you want to read.",
	InputSchema: map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "The path to the file to read.",
			},
		},
	},
	Function: ReadFile,
}

type ReadFileInput struct {
	Path string `json:"path"`
}

func ReadFile(input json.RawMessage) (string, error) {
	var ReadFileInput ReadFileInput
	if err := json.Unmarshal(input, &ReadFileInput); err != nil {
		return "", err
	}

	path := ReadFileInput.Path
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
