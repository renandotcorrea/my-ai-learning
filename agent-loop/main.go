package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/fatih/color"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

func main() {
	loadEnvFile(".env")
	openaiClient := openai.NewClient(openai.DefaultClientOptions()...)

	fmt.Fprintln(os.Stdout, "Olá, como posso ajudá-lo?")

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	agent := NewAgent(&openaiClient)

	if err := agent.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Erro ao executar o agente:", err)
	}

}

type Agent struct {
	Client       *openai.Client
	History      []responses.ResponseInputItemUnionParam
	Instructions string
	TokenCount   int64
}

func NewAgent(client *openai.Client) *Agent {
	return &Agent{
		Client:  client,
		History: []responses.ResponseInputItemUnionParam{},
		Instructions: `Você é um assistente de inteligência artificial que ajuda os usuários a decidir qual item comprar. 
					Você compara diferentes opções com base em suas características, preços e benefícios. Sempre forneça recomendações imparciais e detalhadas.
					Sempre tente entender completamente as necessidades do usuário antes de fornecer uma recomendação. Ao recomendar um item, explique claramente os motivos da escolha e compare com outras opções disponíveis.
					Trafa também os prós e contras de cada opção.`,
		TokenCount: 0,
	}
}

func (a *Agent) addToHistory(input string, role responses.EasyInputMessageRole) {
	a.History = append(a.History, responses.ResponseInputItemParamOfMessage(input, role))
}

func (a *Agent) IncrementTokenCount(count int64) {
	a.TokenCount += count
}

func (a *Agent) Call(ctx context.Context, input string) error {
	a.addToHistory(input, responses.EasyInputMessageRoleUser)

	resp, err := a.Client.Responses.New(ctx, responses.ResponseNewParams{
		Model:        openai.ChatModelGPT6Luna,
		Instructions: openai.String(a.Instructions),
		Input: responses.ResponseNewParamsInputUnion{
			OfInputItemList: a.History,
		},
	})
	if err != nil {
		return err
	}
	a.IncrementTokenCount(resp.Usage.TotalTokens)

	a.addToHistory(resp.OutputText(), responses.EasyInputMessageRoleAssistant)
	fmt.Fprintln(os.Stdout, color.CyanString("Agente: "), resp.OutputText())
	return nil
}

func (a *Agent) Run() error {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		for {
			input := getInput()
			if err := a.Call(context.Background(), input); err != nil {
				fmt.Fprintln(os.Stderr, "Erro ao chamar o agente:", err)
				break
			}
		}
	}()
	<-signalChan
	fmt.Fprintln(os.Stdout, "Interrupção recebida. Encerrando...")
	return nil
}
