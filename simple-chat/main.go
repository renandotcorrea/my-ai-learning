package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

func main() {
	b, err := os.ReadFile(".env")
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

	openaiClient := openai.NewClient(openai.DefaultClientOptions()...)

	fmt.Fprintln(os.Stdout, "Olá, como posso ajudá-lo?")

	inputList := []responses.ResponseInputItemUnionParam{}
	tokenCount := int64(0)

	defer func() {
		fmt.Fprintln(os.Stdout, "Total de tokens usados no chat:", tokenCount)
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	go func() {
		for {
			reader := bufio.NewReader(os.Stdin)
			fmt.Fprint(os.Stdout, "Você: ")

			input, err := reader.ReadString('\n')
			if err != nil {
				fmt.Printf("Erro ao ler entrada: %v\n", err)
				os.Exit(1)
			}

			input = strings.TrimSpace(input)

			inputList = append(inputList, responses.ResponseInputItemUnionParam{
				OfInputMessage: &responses.ResponseInputItemMessageParam{
					Role:    "user",
					Content: responses.ResponseInputMessageContentListParam{{OfInputText: &responses.ResponseInputTextParam{Text: input}}},
				},
			})
			start := time.Now()
			resp, err := openaiClient.Responses.New(context.Background(), responses.ResponseNewParams{
				Model: openai.ChatModelGPT6Luna,
				Input: responses.ResponseNewParamsInputUnion{
					OfInputItemList: inputList,
				},
			})
			latency := time.Since(start)

			if err != nil {
				fmt.Printf("Erro ao criar resposta: %v\n", err)
				os.Exit(1)
			}

			outputText := resp.OutputText()

			fmt.Fprintln(os.Stdout, "Agente:", outputText)
			fmt.Fprintln(os.Stdout, "tokens usados:", resp.Usage.TotalTokens)
			fmt.Fprintln(os.Stdout, "latência:", latency)
			tokenCount += resp.Usage.TotalTokens

			inputList = append(inputList, responses.ResponseInputItemUnionParam{
				OfInputMessage: &responses.ResponseInputItemMessageParam{
					Role:    "system",
					Content: responses.ResponseInputMessageContentListParam{{OfInputText: &responses.ResponseInputTextParam{Text: outputText}}},
				},
			})
		}
	}()

	<-signalChan
	fmt.Fprintln(os.Stdout, "\nInterrupção recebida. Encerrando...")
}
