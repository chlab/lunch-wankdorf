package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/leuenbergerc/lunch-wankdorf/pkg/ai"
	"github.com/leuenbergerc/lunch-wankdorf/pkg/file"
	"github.com/leuenbergerc/lunch-wankdorf/pkg/scraper"
)

// RestaurantMenu defines a restaurant menu source
type RestaurantMenu struct {
	Name    string
	URL     string
	BaseURL string
}

// Available restaurant menus
var restaurantMenus = map[string]RestaurantMenu{
	"gira": {
		Name:    "Gira",
		URL:     "https://app.food2050.ch/de/sbb-gira/gira/menu/mittagsmenue/weekly",
		BaseURL: "https://app.food2050.ch",
	},
}

// Config holds application configuration settings
type Config struct {
	DebugMode    bool   // If true, debug files will be written
	DryRun       bool   // If true, no API calls will be made
	RestaurantID string // ID of the restaurant to fetch menu from (defaults to "gira")
}

// Run starts the application
func Run(config Config) {
	// Load environment variables from .env file
	loadEnv()

	// Use default restaurant if none specified
	restaurantID := config.RestaurantID
	if restaurantID == "" {
		restaurantID = "gira"
	}

	// Get restaurant menu configuration
	restaurant, exists := restaurantMenus[restaurantID]
	if !exists {
		log.Fatalf("Restaurant with ID '%s' not found", restaurantID)
	}

	// Fetch the restaurant menu content
	fmt.Printf("Scraping menu data for %s from %s\n", restaurant.Name, restaurant.URL)
	htmlContent, err := scraper.ScrapeMenuContent(restaurant.URL)
	if err != nil {
		log.Fatalf("Error scraping menu data: %v", err)
	}

	// Save debug files if debug mode is enabled
	if config.DebugMode {
		// Save raw HTML content to debug file
		htmlDebugFile, err := file.WriteToDebugFile(htmlContent.Content, "raw_html")
		if err != nil {
			log.Printf("Warning: Could not write raw HTML to debug file: %v", err)
		} else {
			fmt.Printf("Saved raw HTML to %s\n", htmlDebugFile)
		}
	}

	// Extract menu content
	menuData := ExtractMenuContent(htmlContent.Content)

	// Save debug files if debug mode is enabled
	if config.DebugMode {
		// Save extracted menu content to debug file
		menuContentDebugFile, err := file.WriteToDebugFile(menuData, "menu_content")
		if err != nil {
			log.Printf("Warning: Could not write menu content to debug file: %v", err)
		} else {
			fmt.Printf("Saved menu content to %s\n", menuContentDebugFile)
		}
	}

	contentLength := len(menuData)
	fmt.Printf("Successfully extracted menu content (%d bytes)\n", contentLength)

	if contentLength == 0 {
		log.Fatalf("No menu content found on the page")
	}

	// Print a sample of the content
	preview := menuData
	if len(preview) > 500 {
		preview = preview[:500] + "..."
	}
	fmt.Println("\nMenu Content Sample:")
	fmt.Println("=============")
	fmt.Println(preview)
	fmt.Println("=============")

	// Abort menu parsing if dry run is enabled
	if config.DryRun {
		fmt.Println("Dry Run, aborting parsing menu...")
		return
	}

	// Parse menu using OpenAI
	fmt.Println("Parsing menu data with OpenAI...")
	parsedMenu, err := ai.ParseRestaurantMenu(menuData)
	if err != nil {
		file.WriteToDebugFile(parsedMenu, "parsed_menu")
		log.Fatalf("Error parsing menu data: %v", err)
	}

	// Process the menu to add full URLs to links
	processedMenu, err := processMenuLinks(parsedMenu, restaurant.BaseURL)
	if err != nil {
		log.Printf("Warning: Could not process menu links: %v", err)
		processedMenu = parsedMenu // Fall back to original parsed menu
	}

	// Format JSON for output
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(processedMenu), "", "  "); err != nil {
		prettyJSON = *bytes.NewBufferString(processedMenu)
	}

	// Save debug files if debug mode is enabled
	if config.DebugMode {
		// Save parsed menu to debug file
		parsedMenuDebugFile, err := file.WriteToDebugFile(prettyJSON.String(), "parsed_menu")
		if err != nil {
			log.Printf("Warning: Could not write parsed menu to debug file: %v", err)
		} else {
			fmt.Printf("Saved parsed menu to %s\n", parsedMenuDebugFile)
		}
	}

	// Output the parsed menu
	fmt.Println("\nWeekly Menu:")
	fmt.Println("===========")
	fmt.Println(prettyJSON.String())
}

// processMenuLinks adds the restaurant's base URL to relative links in the menu
func processMenuLinks(menuJSON string, baseURL string) (string, error) {
	// Parse the JSON string
	var menuData map[string][]map[string]interface{}
	if err := json.Unmarshal([]byte(menuJSON), &menuData); err != nil {
		return "", fmt.Errorf("failed to unmarshal menu JSON: %v", err)
	}

	// Process each day's menu items
	for day, menuItems := range menuData {
		for i, item := range menuItems {
			// Check if the item has a link
			if link, ok := item["link"].(string); ok && link != "" {
				// Check if it's a relative URL
				if strings.HasPrefix(link, "/") {
					// Create the full URL by combining base URL and relative path
					item["link"] = baseURL + link
					menuData[day][i] = item
				}
			}
		}
	}

	// Convert back to JSON
	processedJSON, err := json.Marshal(menuData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal processed menu: %v", err)
	}

	return string(processedJSON), nil
}

// ExtractMenuContent extracts menu-specific content from HTML
func ExtractMenuContent(html string) string {
	return scraper.OptimizeHTML(html)
}

// loadEnv attempts to load environment variables from a .env file
func loadEnv() {
	// Try to find .env file in current directory or parent directories
	dir, err := os.Getwd()
	if err != nil {
		log.Println("Warning: Could not determine current directory:", err)
		return
	}

	// Look for .env in current and parent directories (up to 3 levels)
	for i := 0; i < 3; i++ {
		envFile := filepath.Join(dir, ".env")
		if _, err := os.Stat(envFile); err == nil {
			err = godotenv.Load(envFile)
			if err != nil {
				log.Println("Warning: Error loading .env file:", err)
			} else {
				log.Println("Loaded environment from", envFile)
			}
			return
		}
		// Move up to parent directory
		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			break // Reached root directory
		}
		dir = parentDir
	}

	log.Println("No .env file found, using environment variables if set")
}
