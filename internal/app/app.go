package app

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/leuenbergerc/lunch-wankdorf/pkg/ai"
	"github.com/leuenbergerc/lunch-wankdorf/pkg/scraper"
)

const (
	menuURL = "https://app.food2050.ch/de/sbb-gira/gira/menu/mittagsmenue/weekly"
)

// Run starts the application
func Run() {
	// Load environment variables from .env file
	loadEnv()

	// Fetch the restaurant menu content
	fmt.Println("Scraping menu data from", menuURL)
	menuData, err := scraper.ScrapeMenuContent(menuURL)
	if err != nil {
		log.Fatalf("Error scraping menu data: %v", err)
	}

	contentLength := len(menuData.Content)
	fmt.Printf("Successfully scraped menu content (%d bytes)\n", contentLength)

	if contentLength == 0 {
		log.Fatalf("No menu content found on the page")
	}

	// Print the full HTML content
	fmt.Println("\nHTML Content:")
	fmt.Println("=============")
	fmt.Println(menuData.Content)
	fmt.Println("=============\n")

	// Parse menu using OpenAI
	fmt.Println("Parsing menu data with OpenAI...")
	parsedMenu, err := ai.ParseRestaurantMenu(menuData.Content)
	if err != nil {
		log.Fatalf("Error parsing menu data: %v", err)
	}

	// Output the parsed menu
	fmt.Println("\nWeekly Menu:")
	fmt.Println("===========")
	fmt.Println(parsedMenu)
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