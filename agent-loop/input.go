package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

func getInput() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Fprint(os.Stdout, color.GreenString("Você: "))

	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Erro ao ler entrada: %v\n", err)
		os.Exit(1)
	}

	input = strings.TrimSpace(input)
	return input
}
