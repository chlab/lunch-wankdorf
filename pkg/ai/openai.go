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
	ModelGPT41Mini    = "gpt-4.1-mini"
	completionTimeout = 3 * time.Minute
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

var weekdays = []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}

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

func htmlMenuSchema() json.RawMessage {
	item := menuItemSchema(true)
	properties := make(map[string]any, len(weekdays))
	for _, d := range weekdays {
		properties[d] = map[string]any{"type": "array", "items": item}
	}
	schema := map[string]any{
		"type":                 "object",
		"properties":           properties,
		"required":             weekdays,
		"additionalProperties": false,
	}
	b, _ := json.Marshal(schema)
	return b
}

func pdfMenuSchema() json.RawMessage {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"items": map[string]any{
				"type":  "array",
				"items": menuItemSchema(false),
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
		Model: ModelGPT41Mini,
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

// ParseRestaurantHtmlMenu sends HTML content to OpenAI to extract menu information.
func ParseRestaurantHtmlMenu(htmlContent string) (*DailyMenu, error) {
	prompt := `Parse the following HTML extracted from a restaurant's weekly menu page. The text is in German.
The week starts on Monday. A day may have no menu (holiday, closed) — return an empty array for that day.
Dishes are listed under a heading naming the day they are served on. Put every dish under
that day, and return every dish you are given — do not skip or summarize the later days.
For each menu item provide:
- name: dish name
- description: dish description (remove double commas and other formatting noise but keep the content)
- type: dish type (vegetarian, meat, etc.)
- icon: the icon that best fits the dish — use the name first, description second
- link: link to the dish on the restaurant's website, or an empty string if none
Icon hints (the parenthetical is a hint, not part of the icon name): ` + strings.Join(IconsList, ", ") + `

HTML:
` + htmlContent

	result, err := createCompletion(prompt, htmlMenuSchema(), "restaurant_html_menu")
	if err != nil {
		return nil, fmt.Errorf("failed to parse menu: %w", err)
	}

	var parsed struct {
		Monday    []MenuItem `json:"monday"`
		Tuesday   []MenuItem `json:"tuesday"`
		Wednesday []MenuItem `json:"wednesday"`
		Thursday  []MenuItem `json:"thursday"`
		Friday    []MenuItem `json:"friday"`
		Saturday  []MenuItem `json:"saturday"`
		Sunday    []MenuItem `json:"sunday"`
	}
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse menu JSON: %w", err)
	}

	return &DailyMenu{
		Type: "daily",
		Menu: map[string][]MenuItem{
			"Monday":    parsed.Monday,
			"Tuesday":   parsed.Tuesday,
			"Wednesday": parsed.Wednesday,
			"Thursday":  parsed.Thursday,
			"Friday":    parsed.Friday,
			"Saturday":  parsed.Saturday,
			"Sunday":    parsed.Sunday,
		},
	}, nil
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

	result, err := createCompletion(prompt, pdfMenuSchema(), "restaurant_pdf_menu")
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
