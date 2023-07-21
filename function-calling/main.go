package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	openai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

func main() {
	// Read text from file. Filepath is given as first argument
	filepath := os.Args[1]
	text, err := getTextFromFile(filepath)
	if err != nil {
		log.Fatal(err)
	}

	config := openai.DefaultAzureConfig(os.Getenv("AOAI_KEY"), os.Getenv("AOAI_ENDPOINT"))
	config.APIVersion = "2023-07-01-preview"
	config.AzureModelMapperFunc = func(model string) string {
		modelmapper := map[string]string{
			"gpt-3.5-turbo-16k-0613": "tsunomur-gpt-35-turb-16k",
		}
		if val, ok := modelmapper[model]; ok {
			return val
		}
		return model
	}

	client := openai.NewClientWithConfig(config)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo16K0613,
			Messages: []openai.ChatCompletionMessage{
				{
					Role: openai.ChatMessageRoleUser,
					Content: fmt.Sprintf(`
					Please provide as Japanese, don't use English.必ず日本語にしてください。
					Provide two sentence summary of the following text, emphasizing the most impactful new feature and main service, product if this text including new feature release.
					Keep the summary extremely brief, ideally within 200 characters. Please translate into Japanese.
					"%s"
					Output:`,
						text),
				},
			},
			Functions: []openai.FunctionDefinition{
				{
					Name: "summarize",
					Parameters: &jsonschema.Definition{
						Type: jsonschema.Object,
						Properties: map[string]jsonschema.Definition{
							"description": {
								Type:        jsonschema.String,
								Description: "summary of contents",
							},
							"product": {
								Type:        jsonschema.Array,
								Description: "list of service names",
								Items: &jsonschema.Definition{
									Type: jsonschema.String,
								},
							},
						},
						Required: []string{"description", "product"},
					},
				},
			},
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	type SummarizedArgs struct {
		Description string   `json:"description"`
		Product     []string `json:"product"`
	}
	log.Println(resp.Choices[0].FinishReason)
	fmt.Printf("%#v\n", resp)
	var summarizedArgs SummarizedArgs
	err = json.Unmarshal([]byte(resp.Choices[0].Message.FunctionCall.Arguments), &summarizedArgs)
	if err != nil {
		log.Println(err)
	}
	// Join products with comma
	var p string
	for _, v := range summarizedArgs.Product {
		p += v + ","
	}
	fmt.Printf("* Summary: %s\n* Products: %s\n", summarizedArgs.Description, p)
}

// getTextFromFile reads text from file
func getTextFromFile(filename string) (string, error) {
	// ファイルを読み込む
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}

	// ファイルの内容を文字列に変換する
	str := string(b)

	return str, nil
}
