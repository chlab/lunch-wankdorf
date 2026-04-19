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

const (
	ModelGPT4oMini       = "gpt-4o-mini"
	completionTimeout    = 3 * time.Minute
)

// MenuItem represents a single dish on a restaurant menu.
type MenuItem struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Icon        string `json:"icon"`
	Link        string `json:"link,omitempty"`
	Restaurant  string `json:"restaurant,omitempty"`
}

// DailyMenu wraps a per-day menu (HTML restaurants).
type DailyMenu struct {
	Type string                `json:"type"`
	Menu map[string][]MenuItem `json:"menu"`
}

// WeeklyMenu wraps a flat list of items (PDF restaurants).
type WeeklyMenu struct {
	Type string     `json:"type"`
	Menu []MenuItem `json:"menu"`
}

// Icons list for menu items
var IconsList = []string{
	"bento",
	"curry (only curries)",
	"dumplings (ravioli, gnocchi, tortellini, asian dumplings)",
	"french-fries",
	"fried-chicken (only chicken)",
	"hamburger (any type of burger)",
	"hot-dog",
	"korean-rice-cake (spring or summer rolls)",
	"lasagna-sheets",
	"miso-soup (asian-style soup)",
	"nachos",
	"noodles (asian, not pasta)",
	"paella",
	"pizza",
	"rack-of-lamb",
	"rice-bowl (rice dishes, risotto, bowl)",
	"salad",
	"sandwich",
	"sausage",
	"seafood",
	"spaghetti (pasta)",
	"porridge (mac n cheese)",
	"steak (grilled meats, bbq)",
	"steak-rare (meat)",
	"sushi",
	"taco",
	"vegan-food (vegetarian bowls)",
	"wrap",
}

// CreateCompletion sends a prompt to the OpenAI API and returns the response
func CreateCompletion(prompt string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", errors.New("OPENAI_API_KEY environment variable not set")
	}

	client := openai.NewClient(apiKey)
	ctx, cancel := context.WithTimeout(context.Background(), completionTimeout)
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

// extractJSON attempts to extract a JSON string from an API response that may
// be wrapped in markdown code blocks or contain explanatory text.
func extractJSON(result string) string {
	result = strings.TrimSpace(result)

	// Already looks like valid JSON
	if strings.HasPrefix(result, "{") || strings.HasPrefix(result, "[") {
		return result
	}

	// Try to extract a JSON object
	if start := strings.Index(result, "{"); start >= 0 {
		if end := strings.LastIndex(result, "}"); end > start {
			return result[start : end+1]
		}
	}

	// Try to extract a JSON array
	if start := strings.Index(result, "["); start >= 0 {
		if end := strings.LastIndex(result, "]"); end > start {
			return result[start : end+1]
		}
	}

	return result
}

// ParseRestaurantHtmlMenu sends HTML content to OpenAI to extract menu information
func ParseRestaurantHtmlMenu(htmlContent string) (*DailyMenu, error) {
	prompt := `Parse the following HTML extracted from a restaurant's weekly menu page. The text is in German.
Be aware that a day may be empty due to a holiday or other reason. Important: The week starts on Monday and so does the menu.
Return a JSON structure where the key is the day of the week in English  and the value is an array of menu options
for that day. Each menu option should have these keys:
- name: The name of the dish
- description: A description of the dish
- type: The type of dish (vegetarian, meat, etc.)
- icon: One of the icons in the list below that fits the dish best. The list is comma-separated in the format: icon-name (optional hints).
        Use the menu item name first and the description second to determine the best suited icon.
		Very important: must be an exact match of the icon-name. Do not invent any new names or abbreviations.
- link: A link to the dish on the restaurant's website
List of icons: ` + strings.Join(IconsList, ", ") + `
Format your response as clean, properly formatted JSON only, with no explanations or additional text.
Remove any double commas or other formatting issues from the description but don't change the content.
Here is the extracted HTML of the menu:
` + htmlContent

	result, err := CreateCompletion(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse menu: %w", err)
	}

	cleanedJSON := extractJSON(result)

	var dailyMenu map[string][]MenuItem
	if err := json.Unmarshal([]byte(cleanedJSON), &dailyMenu); err != nil {
		return nil, fmt.Errorf("failed to parse menu JSON: %w", err)
	}

	return &DailyMenu{Type: "daily", Menu: dailyMenu}, nil
}

// ParseRestaurantPdfMenu sends extracted text from a PDF to OpenAI to extract menu information
func ParseRestaurantPdfMenu(extractedText string, restaurantName string, pdfURL string) (*WeeklyMenu, error) {
	prompt := `Parse the following extracted text from a restaurant's menu PDF.
Return a JSON structure with an array of menu options. Each menu option should have these keys:
- name: The name of the dish
- description: A description of the dish
- type: The type of dish (vegetarian, meat, etc.)
- icon: One of the icons in the list below that fits the dish best. The list is comma-separated in the format: icon-name (optional hints).
        Use the menu item name first and the description second to determine the best suited icon.
		Very important: must be an exact match of the icon-name. Do not invent any new names or abbreviations.
List of icons: ` + strings.Join(IconsList, ", ") + `
Only include food, ignore drinks. If not specified otherwise, assume Turbolama are vegan bowls.
Format your response as clean, properly formatted JSON only, with no explanations or additional text.

Extracted PDF content:
` + extractedText

	result, err := CreateCompletion(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PDF menu: %w", err)
	}

	cleanedJSON := extractJSON(result)

	// Try to parse the JSON - it might be in different formats
	var parsedItems []MenuItem

	// First try parsing as array of items directly
	if err := json.Unmarshal([]byte(cleanedJSON), &parsedItems); err != nil {
		// Try menuItems field
		var menuItemsObject struct {
			MenuItems []MenuItem `json:"menuItems"`
		}
		if jsonErr := json.Unmarshal([]byte(cleanedJSON), &menuItemsObject); jsonErr == nil && len(menuItemsObject.MenuItems) > 0 {
			parsedItems = menuItemsObject.MenuItems
		} else {
			// Try menuOptions field (what OpenAI seems to be returning)
			var menuOptionsObject struct {
				MenuOptions []MenuItem `json:"menuOptions"`
			}
			if jsonErr := json.Unmarshal([]byte(cleanedJSON), &menuOptionsObject); jsonErr == nil && len(menuOptionsObject.MenuOptions) > 0 {
				parsedItems = menuOptionsObject.MenuOptions
			} else {
				return nil, fmt.Errorf("failed to parse menu items from JSON: %w", err)
			}
		}
	}

	// Add restaurant and link fields to each menu item
	for i := range parsedItems {
		parsedItems[i].Restaurant = restaurantName
		parsedItems[i].Link = pdfURL
	}

	return &WeeklyMenu{Type: "weekly", Menu: parsedItems}, nil
}
