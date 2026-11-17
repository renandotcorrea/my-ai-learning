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
	for _, tool := range []Tool{ListFilesTool, ReadFileTool, LinesCountTool} {
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
		Instructions: `Você é um assistente de inteligência artificial que ajuda os usuários a interagir com arquivos e diretórios no sistema.
		Sempre forneça respostas claras e detalhadas sobre as operações realizadas.`,
		TokenCount: 0,
	}
}

func (a *Agent) addToHistory(input string, role responses.EasyInputMessageRole) {
	a.History = append(a.History, responses.ResponseInputItemParamOfMessage(input, role))
}

func (a *Agent) addRawToHistory(item responses.ResponseInputItemUnionParam) {
	a.History = append(a.History, item)
}

func (a *Agent) IncrementTokenCount(count int64) {
	a.TokenCount += count
}

func (a *Agent) Call(ctx context.Context, input string, readUserInput *bool) error {
	if *readUserInput {
		a.addToHistory(input, responses.EasyInputMessageRoleUser)
	}

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
	toolResult := []responses.ResponseInputItemUnionParam{}
	for _, v := range resp.Output {
		fmt.Println("Processing output item of type:", v.Type)
		switch v.Type {
		case "message":
			for _, c := range v.AsMessage().Content {
				outputText.WriteString(c.Text)
			}
		case "function_call":
			call := v.AsFunctionCall()
			fmt.Fprintf(&outputText, "Function called: %s\n", call.Name)

			a.addRawToHistory(responses.ResponseInputItemParamOfFunctionCall(call.Arguments, call.CallID, call.Name))

			toolRes := a.executeTool(call.Name, json.RawMessage(call.Arguments))
			if toolRes.OfFunctionCallOutput != nil {
				toolRes.OfFunctionCallOutput.CallID = openai.String(call.CallID)
			}
			toolResult = append(toolResult, toolRes)
		case "reasoning":
			// Handle reasoning output if needed
			// For now, just print it
			r := v.AsReasoning()
			fmt.Fprintf(os.Stdout, "%s", color.CyanString("Reasoning: "))
			for _, c := range r.Content {
				fmt.Fprintf(os.Stdout, "%s\n", c.Text)
			}
			fmt.Fprintln(os.Stdout)
		}
	}

	if len(toolResult) == 0 {
		a.addToHistory(outputText.String(), responses.EasyInputMessageRoleAssistant)
		fmt.Fprintln(os.Stdout, color.CyanString("Agente: "), outputText.String())
		*readUserInput = true
	} else {
		for _, tr := range toolResult {
			a.addRawToHistory(tr)
			//fmt.Fprintf(os.Stdout, "%s: %#v\n", color.CyanString("Tool result: "), tr)
		}
		*readUserInput = false
	}
	return nil
}

func (a *Agent) Run() error {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		readUserInput := true
		for {
			input := ""
			if readUserInput {
				input = getInput()
			}
			if err := a.Call(context.Background(), input, &readUserInput); err != nil {
				fmt.Fprintln(os.Stderr, "Erro ao chamar o agente:", err)
				break
			}
		}
	}()
	<-signalChan
	fmt.Fprintln(os.Stdout, "Interrupção recebida. Encerrando...")
	return nil
}

func (a *Agent) executeTool(name string, args json.RawMessage) responses.ResponseInputItemUnionParam {
	if tool, ok := a.Tools[name]; ok {
		fmt.Fprintf(os.Stdout, "%s: %s(%s)\n", color.CyanString("Tool: "), name, string(args))

		v, err := tool.Function(args)
		if err != nil {
			return responses.ResponseInputItemParamOfFunctionCallOutput(err.Error())
		}
		return responses.ResponseInputItemParamOfFunctionCallOutput(v)
	}

	return responses.ResponseInputItemParamOfFunctionCallOutput(fmt.Sprintf("tool not found: %s", name))
}
