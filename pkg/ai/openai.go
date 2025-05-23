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
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	req := openai.ChatCompletionRequest{
		Model: ModelGPT4oMini,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	}

	// Standard text-only request
	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", errors.New("no response from API")
	}

	return resp.Choices[0].Message.Content, nil
}

// validateJSON validates that a string is valid JSON
// and attempts to extract valid JSON if it's embedded in markdown or text
func validateJSON(result string) (string, error) {
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

// ParseRestaurantHtmlMenu sends HTML content to OpenAI to extract menu information
func ParseRestaurantHtmlMenu(htmlContent string) (string, error) {
	prompt := `Parse the following HTML extracted from a restaurant's weekly menu page. The text is in German.
Be aware that a day may be empty due to a holiday or other reason. Important: The week starts on Monday and so does the menu.
Return a JSON structure where the key is the day of the week in English  and the value is an array of menu options
for that day. Each menu option should have these keys:
- name: The name of the dish
- description: A description of the dish
- type: The type of dish (vegetarian, meat, etc.)
- link: A link to the dish on the restaurant's website
Format your response as clean, properly formatted JSON only, with no explanations or additional text.
Remove any double commas or other formatting issues from the description but don't change the content.
Here is the extracted HTML of the menu:
` + htmlContent

	result, err := CreateCompletion(prompt)
	if err != nil {
		return "", fmt.Errorf("failed to parse menu: %w", err)
	}

	// Validate the JSON
	cleanedJSON, err := validateJSON(result)
	if err != nil {
		return "", err
	}

	// Parse the days of the week structure
	var dailyMenu map[string]interface{}
	if err := json.Unmarshal([]byte(cleanedJSON), &dailyMenu); err != nil {
		return "", fmt.Errorf("failed to parse menu JSON: %w", err)
	}

	// Create the new structure
	finalMenu := struct {
		Type string                 `json:"type"`
		Menu map[string]interface{} `json:"menu"`
	}{
		Type: "daily",
		Menu: dailyMenu,
	}

	// Convert back to JSON
	// TODO: don't convert back to JSON
	finalJSON, err := json.Marshal(finalMenu)
	if err != nil {
		return "", fmt.Errorf("failed to marshal final menu: %w", err)
	}

	return string(finalJSON), nil
}

// ParseRestaurantPdfMenu sends extracted text from a PDF to OpenAI to extract menu information
func ParseRestaurantPdfMenu(extractedText string, restaurantName string, pdfURL string) (string, error) {
	prompt := `Parse the following extracted text from a restaurant's menu PDF.
Return a JSON structure with an array of menu options. Each menu option should have these keys:
- name: The name of the dish
- description: A description of the dish
- type: The type of dish (vegetarian, meat, etc.)
Only include food, ignore drinks.
Format your response as clean, properly formatted JSON only, with no explanations or additional text.

Extracted PDF content:
` + extractedText

	result, err := CreateCompletion(prompt)
	if err != nil {
		return "", fmt.Errorf("failed to parse PDF menu: %w", err)
	}

	// First validate and clean the JSON
	cleanedJSON, err := validateJSON(result)
	if err != nil {
		return "", err
	}

	// Try to parse the JSON - it might be in different formats
	var parsedItems []map[string]interface{}

	// First try parsing as array of items directly
	if err := json.Unmarshal([]byte(cleanedJSON), &parsedItems); err != nil {
		// Check for common JSON structures

		// Try menuItems field
		var menuItemsObject struct {
			MenuItems []map[string]interface{} `json:"menuItems"`
		}
		if jsonErr := json.Unmarshal([]byte(cleanedJSON), &menuItemsObject); jsonErr == nil && len(menuItemsObject.MenuItems) > 0 {
			parsedItems = menuItemsObject.MenuItems
		} else {
			// Try menuOptions field (what OpenAI seems to be returning)
			var menuOptionsObject struct {
				MenuOptions []map[string]interface{} `json:"menuOptions"`
			}
			if jsonErr := json.Unmarshal([]byte(cleanedJSON), &menuOptionsObject); jsonErr == nil && len(menuOptionsObject.MenuOptions) > 0 {
				parsedItems = menuOptionsObject.MenuOptions
			} else {
				return "", fmt.Errorf("failed to parse menu items from JSON: %w", err)
			}
		}
	}

	// Add restaurant and link fields to each menu item
	for i := range parsedItems {
		parsedItems[i]["restaurant"] = restaurantName
		parsedItems[i]["link"] = pdfURL
	}

	// Create the final structure
	finalMenu := struct {
		Type string                   `json:"type"`
		Menu []map[string]interface{} `json:"menu"`
	}{
		Type: "weekly",
		Menu: parsedItems,
	}

	// Convert back to JSON
	finalJSON, err := json.Marshal(finalMenu)
	if err != nil {
		return "", fmt.Errorf("failed to marshal final menu: %w", err)
	}

	return string(finalJSON), nil
}
