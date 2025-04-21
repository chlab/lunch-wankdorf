package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
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
			// MaxTokens: 2000,
		},
	)
	if err != nil {
		return "", fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", errors.New("no response from API")
	}

	return resp.Choices[0].Message.Content, nil
}

// ParseRestaurantMenu sends HTML content to OpenAI to extract menu information
func ParseRestaurantMenu(htmlContent string) (string, error) {
	prompt := `Parse the following HTML extracted from a restaurant's weekly menu page. The text is in German.
Be aware that a day may be empty due to a holiday or other reason. Important: The week starts on Monday and so does the menu.
Return a JSON structure where the key is the day of the week in English  and the value is an array of menu options 
for that day. Each menu option should have these keys:
- name: The name of the dish
- description: A description of the dish
- type: The type of dish (vegetarian, meat, etc.)
- link: A link to the dish on the restaurant's website
Format your response as clean, properly formatted JSON only, with no explanations or additional text.
Here is the extracted HTML of the menu:
` + htmlContent

	result, err := CreateCompletion(prompt)
	if err != nil {
		return "", fmt.Errorf("failed to parse menu: %w", err)
	}

	// Validate that the result is valid JSON
	var jsonData interface{}
	if err := json.Unmarshal([]byte(result), &jsonData); err != nil {
		// If not valid JSON, try to extract JSON from the response
		// Sometimes the model might include markdown backticks or explanations
		jsonStartIdx := strings.Index(result, "{")
		jsonEndIdx := strings.LastIndex(result, "}")

		if jsonStartIdx >= 0 && jsonEndIdx > jsonStartIdx {
			jsonStr := result[jsonStartIdx : jsonEndIdx+1]
			if err := json.Unmarshal([]byte(jsonStr), &jsonData); err == nil {
				return jsonStr, nil
			}
		}

		return result, fmt.Errorf("API returned invalid JSON: %v", err)
	}

	return result, nil
}
