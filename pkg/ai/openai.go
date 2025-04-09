package ai

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/sashabaranov/go-openai"
)

// Model constants
const (
	ModelGPT4oMini = "gpt-4o-mini"
)

// CreateCompletion sends a prompt to the OpenAI API and returns the response
func CreateCompletion(prompt string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", errors.New("OPENAI_API_KEY environment variable not set")
	}

	client := openai.NewClient(apiKey)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: ModelGPT4oMini,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			MaxTokens: 2000,
		},
	)

	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", errors.New("no response from API")
	}

	return resp.Choices[0].Message.Content, nil
}

// ParseRestaurantMenu sends HTML content to OpenAI to extract menu information
func ParseRestaurantMenu(htmlContent string) (string, error) {
	prompt := `Parse the following HTML of a restaurant's menu options for the week. The text is in German. 
Return a JSON structure where the key is the day of the week in English (Monday, Tuesday, etc.) 
and the value is an array of menu options for that day. Each menu option should have these keys:
- name: The name of the dish
- description: A description of the dish
- type: The type of dish (vegetarian, meat, etc.)

Format your response as clean, properly formatted JSON only, with no explanations or additional text.

Here is the menu HTML content to parse:
` + htmlContent

	return CreateCompletion(prompt)
}