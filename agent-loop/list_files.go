package main

import (
	"encoding/json"
	"os"
	"strings"
)

var ListFilesTool = Tool{
	Name:        "list_files",
	Description: "Lists all files in the specified directory. Defaults to the current directory if no path is provided. Use it to explore the contents of any directory on your system.",
	InputSchema: map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "The path to the directory to list files from. Defaults to the current directory if empty.",
			},
		},
	},
	Function: ListFiles,
}

type ListFilesInput struct {
	Path string `json:"path"`
}

func ListFiles(input json.RawMessage) (string, error) {
	var listFilesInput ListFilesInput
	if err := json.Unmarshal(input, &listFilesInput); err != nil {
		return "", err
	}

	path := listFilesInput.Path
	if path == "" {
		path = "."
	}

	files, err := os.ReadDir(path)
	if err != nil {
		return "", err
	}
	var fileNames []string
	for _, file := range files {
		fileNames = append(fileNames, file.Name())
	}
	return strings.Join(fileNames, "\n"), nil
}
