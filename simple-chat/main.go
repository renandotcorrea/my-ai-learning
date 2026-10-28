package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/fatih/color"
	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

func main() {
	loadEnvFile(".env")

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
			fmt.Fprint(os.Stdout, color.GreenString("Você: "))

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
			stream := openaiClient.Responses.NewStreaming(context.Background(), responses.ResponseNewParams{
				Model: openai.ChatModelGPT6Luna,
				Input: responses.ResponseNewParamsInputUnion{
					OfInputItemList: inputList,
				},
			})

			fmt.Fprintln(os.Stdout)
			fmt.Fprint(os.Stdout, color.CyanString("Agente: "))
			stokens := int64(0)
			var finalText strings.Builder
			for stream.Next() {
				resp := stream.Current().Response
				outputText := resp.OutputText()
				finalText.WriteString(outputText)

				fmt.Fprint(os.Stdout, outputText)
				tokenCount += resp.Usage.TotalTokens
				stokens += resp.Usage.TotalTokens
			}

			if stream.Err() != nil {
				fmt.Printf("Erro ao criar resposta: %v\n", stream.Err())
				os.Exit(1)
			}

			inputList = append(inputList, responses.ResponseInputItemUnionParam{
				OfInputMessage: &responses.ResponseInputItemMessageParam{
					Role:    "system",
					Content: responses.ResponseInputMessageContentListParam{{OfInputText: &responses.ResponseInputTextParam{Text: finalText.String()}}},
				},
			})

			fmt.Fprintln(os.Stdout)

			fmt.Fprintln(os.Stdout, color.YellowString("tokens usados: %d", stokens))
			fmt.Fprintln(os.Stdout, color.YellowString("latência: %s", time.Since(start)))
			fmt.Fprintln(os.Stdout)
		}
	}()

	<-signalChan
	fmt.Fprintln(os.Stdout, "\nInterrupção recebida. Encerrando...")
}

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
