package app

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/leuenbergerc/lunch-wankdorf/pkg/ai"
)

// Run starts the application
func Run() {
	// Load environment variables from .env file
	loadEnv()

	// Sample prompt to get lunch suggestions
	prompt := `Generate 3 lunch options for a team in Bern, Switzerland. 
For each option, provide:
1. Name of the dish
2. Type of cuisine 
3. Estimated preparation time
4. A brief description
Format as a simple list.`

	fmt.Println("Sending prompt to OpenAI API...")
	response, err := ai.CreateCompletion(prompt)
	if err != nil {
		log.Fatalf("Error getting AI response: %v", err)
	}

	fmt.Println("\nLunch Options:")
	fmt.Println("==============")
	fmt.Println(response)
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