package main

import (
	"fmt"
	"os"
	"strings"
)

func loadEnvFile(filename string) {
	b, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Erro ao ler arquivo .env: %v\n", err)
		os.Exit(1)
	}

	lines := strings.SplitSeq(string(b), "\n")
	for line := range lines {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			os.Setenv(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}
}
