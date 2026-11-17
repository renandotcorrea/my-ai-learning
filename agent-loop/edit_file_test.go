package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestEditFileReplacesInclusiveRange(t *testing.T) {
	path := filepath.Join(t.TempDir(), "file.txt")
	original := "one\ntwo\nthree\nfour"
	if err := os.WriteFile(path, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}

	input, err := json.Marshal(EditFileInput{
		Path:     path,
		Patch:    "replacement",
		FromLine: 2,
		ToLine:   3,
		Type:     "replace",
	})
	if err != nil {
		t.Fatal(err)
	}

	got, err := EditFile(input)
	if err != nil {
		t.Fatal(err)
	}
	if want := "one\nreplacement\nfour"; got != want {
		t.Fatalf("EditFile() = %q, want %q", got, want)
	}
}

func TestEditFileRejectsInvalidReplacementRangeWithoutWriting(t *testing.T) {
	tests := []struct {
		name     string
		fromLine int
		toLine   int
	}{
		{name: "zero start", fromLine: 0, toLine: 1},
		{name: "reversed range", fromLine: 2, toLine: 1},
		{name: "end past file", fromLine: 1, toLine: 4},
		{name: "start past file", fromLine: 4, toLine: 4},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "file.txt")
			original := "one\ntwo\nthree"
			if err := os.WriteFile(path, []byte(original), 0600); err != nil {
				t.Fatal(err)
			}

			input, err := json.Marshal(EditFileInput{
				Path:     path,
				Patch:    "replacement",
				FromLine: test.fromLine,
				ToLine:   test.toLine,
				Type:     "replace",
			})
			if err != nil {
				t.Fatal(err)
			}

			if _, err := EditFile(input); err == nil {
				t.Fatal("EditFile() expected an error for an invalid range")
			}

			got, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != original {
				t.Fatalf("file content = %q, want unchanged content %q", got, original)
			}
		})
	}
}
