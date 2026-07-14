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
	// gpt-4.1-mini used to lose about a fifth of the dishes it was given, even one
	// day at a time. gpt-5.4-mini returned every dish on every run of the same
	// input, and did it faster. See the model notes in the README.
	DefaultModel      = "gpt-5.4-mini"
	completionTimeout = 3 * time.Minute
)

// Model returns the model to parse menus with, overridable via OPENAI_MODEL.
func Model() string {
	if model := os.Getenv("OPENAI_MODEL"); model != "" {
		return model
	}
	return DefaultModel
}

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

// IconsList describes each icon plus an optional disambiguation hint, for use
// in the prompt. The schema enum uses just the bare icon names (see iconNames).
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

func iconNames() []string {
	out := make([]string, len(IconsList))
	for i, item := range IconsList {
		if idx := strings.Index(item, " ("); idx > 0 {
			out[i] = item[:idx]
		} else {
			out[i] = item
		}
	}
	return out
}

func menuItemSchema(includeLink bool) map[string]any {
	properties := map[string]any{
		"name":        map[string]any{"type": "string"},
		"description": map[string]any{"type": "string"},
		"type":        map[string]any{"type": "string"},
		"icon":        map[string]any{"type": "string", "enum": iconNames()},
	}
	required := []string{"name", "description", "type", "icon"}
	if includeLink {
		properties["link"] = map[string]any{"type": "string"}
		required = append(required, "link")
	}
	return map[string]any{
		"type":                 "object",
		"properties":           properties,
		"required":             required,
		"additionalProperties": false,
	}
}

func itemsSchema(includeLink bool) json.RawMessage {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"items": map[string]any{
				"type":  "array",
				"items": menuItemSchema(includeLink),
			},
		},
		"required":             []string{"items"},
		"additionalProperties": false,
	}
	b, _ := json.Marshal(schema)
	return b
}

func createCompletion(prompt string, schema json.RawMessage, schemaName string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", errors.New("OPENAI_API_KEY environment variable not set")
	}

	client := openai.NewClient(apiKey)
	ctx, cancel := context.WithTimeout(context.Background(), completionTimeout)
	defer cancel()

	req := openai.ChatCompletionRequest{
		Model: Model(),
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
			JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
				Name:   schemaName,
				Schema: schema,
				Strict: true,
			},
		},
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", errors.New("no response from API")
	}

	return resp.Choices[0].Message.Content, nil
}

// ParseDayMenu sends a single day's HTML to OpenAI to extract that day's dishes.
//
// One call per day, rather than one call for the whole week: the model reliably
// lost interest towards the end of a week-long document and returned the last days
// empty. A day is small enough to parse in full, the day itself is never in doubt,
// and a day that does come back short can be retried on its own.
func ParseDayMenu(day string, dayHTML string) ([]MenuItem, error) {
	prompt := `Parse the following HTML extracted from a restaurant's menu page. The text is in German.
It contains the dishes for a single day (` + day + `). Return every dish on offer that day.
A category with no dish (its content is just ".") is closed — skip it, do not invent a dish for it.
Ignore prices, allergen information and climate labels.
For each menu item provide:
- name: dish name
- description: dish description (remove double commas and other formatting noise but keep the content)
- type: dish type (vegetarian, meat, etc.)
- icon: the icon that best fits the dish — use the name first, description second
- link: link to the dish on the restaurant's website, or an empty string if none
Icon hints (the parenthetical is a hint, not part of the icon name): ` + strings.Join(IconsList, ", ") + `

HTML:
` + dayHTML

	result, err := createCompletion(prompt, itemsSchema(true), "restaurant_day_menu")
	if err != nil {
		return nil, fmt.Errorf("failed to parse the %s menu: %w", day, err)
	}

	var parsed struct {
		Items []MenuItem `json:"items"`
	}
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse the %s menu JSON: %w", day, err)
	}

	return parsed.Items, nil
}

// ParseRestaurantPdfMenu sends extracted text from a PDF to OpenAI to extract menu information.
func ParseRestaurantPdfMenu(extractedText string, restaurantName string, pdfURL string) (*WeeklyMenu, error) {
	prompt := `Parse the following extracted text from a restaurant's menu PDF.
For each menu item provide:
- name: dish name
- description: dish description
- type: dish type (vegetarian, meat, etc.)
- icon: the icon that best fits the dish — use the name first, description second
Icon hints (the parenthetical is a hint, not part of the icon name): ` + strings.Join(IconsList, ", ") + `
Only include food, ignore drinks. If not specified otherwise, assume Turbolama are vegan bowls.

Extracted PDF content:
` + extractedText

	result, err := createCompletion(prompt, itemsSchema(false), "restaurant_pdf_menu")
	if err != nil {
		return nil, fmt.Errorf("failed to parse PDF menu: %w", err)
	}

	var parsed struct {
		Items []MenuItem `json:"items"`
	}
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse menu items from JSON: %w", err)
	}

	for i := range parsed.Items {
		parsed.Items[i].Restaurant = restaurantName
		parsed.Items[i].Link = pdfURL
	}

	return &WeeklyMenu{Type: "weekly", Menu: parsed.Items}, nil
}
