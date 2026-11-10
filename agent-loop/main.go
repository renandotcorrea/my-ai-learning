package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"

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

	toolMap := make(map[string]Tool)
	for _, tool := range []Tool{ListFilesTool} {
		toolMap[tool.Name] = tool
	}

	agent := NewAgent(&openaiClient, toolMap)

	if err := agent.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Erro ao executar o agente:", err)
	}

}

type Agent struct {
	Client       *openai.Client
	History      []responses.ResponseInputItemUnionParam
	Tools        map[string]Tool
	Instructions string
	TokenCount   int64
}

func NewAgent(client *openai.Client, tools map[string]Tool) *Agent {
	return &Agent{
		Client:  client,
		History: []responses.ResponseInputItemUnionParam{},
		Tools:   tools,
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

	tools := []responses.ToolUnionParam{}
	for _, tool := range a.Tools {
		tools = append(tools, tool.ToOpenAI())
	}

	resp, err := a.Client.Responses.New(ctx, responses.ResponseNewParams{
		Model:        openai.ChatModelGPT6Luna,
		Instructions: openai.String(a.Instructions),
		Input: responses.ResponseNewParamsInputUnion{
			OfInputItemList: a.History,
		},
		Tools: tools,
	})
	if err != nil {
		return err
	}
	a.IncrementTokenCount(resp.Usage.TotalTokens)

	var outputText strings.Builder
	for _, v := range resp.Output {
		fmt.Println("Processing output item of type:", v.Type)
		switch v.Type {
		case "message":
			for _, c := range v.AsMessage().Content {
				outputText.WriteString(c.Text)
			}
		case "function_call":
			outputText.WriteString(fmt.Sprintf("Function called: %s\n", v.AsFunctionCall().Name))
			if tool, ok := a.Tools[v.AsFunctionCall().Name]; ok {
				val, err := tool.Function(json.RawMessage(v.AsFunctionCall().Arguments))
				if err != nil {
					fmt.Fprintf(&outputText, "Error executing tool: %s\n", err)
				} else {
					fmt.Fprintf(&outputText, "Tool result: %s\n", val)
				}
			}
		}
	}

	a.addToHistory(outputText.String(), responses.EasyInputMessageRoleAssistant)
	fmt.Fprintln(os.Stdout, color.CyanString("Agente: "), outputText.String())
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
