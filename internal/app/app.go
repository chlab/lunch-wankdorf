package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/chlab/lunch-wankdorf/pkg/ai"
	"github.com/chlab/lunch-wankdorf/pkg/file"
	"github.com/chlab/lunch-wankdorf/pkg/scraper"
	"github.com/joho/godotenv"
)

// RestaurantMenu defines a restaurant menu source
type RestaurantMenu struct {
	Name             string
	URL              string
	BaseURL          string
	HasCustomScraper bool   // Indicates if a custom scraping function should be used
	MenuType         string // Type of menu: "html" or "pdf"
	MenuSelector     string // CSS selector to find the menu link (for PDF menus)
}

// Available restaurant menus
var restaurantMenus = map[string]RestaurantMenu{
	"gira": {
		Name:             "Gira",
		URL:              "https://app.food2050.ch/de/v2/zfv/sbb/gira/mittagsverpflegung/menu/weekly",
		BaseURL:          "https://app.food2050.ch",
		HasCustomScraper: false,
		MenuType:         "html",
	},
	"luna": {
		Name:             "Luna",
		URL:              "https://app.food2050.ch/de/v2/zfv/sbb/restaurant-luna/mittagsverpflegung/menu/weekly",
		BaseURL:          "https://app.food2050.ch",
		HasCustomScraper: false,
		MenuType:         "html",
	},
	"sole": {
		Name:             "Sole",
		URL:              "https://app.food2050.ch/de/v2/zfv/sbb/sole/mittagsverpflegung/menu/weekly",
		BaseURL:          "https://app.food2050.ch",
		HasCustomScraper: false,
		MenuType:         "html",
	},
	"espace": {
		Name:             "Espace",
		URL:              "https://web.sv-restaurant.ch/menu/Post,%20Restaurant%20Espace,%20Bern/Mittagsmen%C3%BC",
		BaseURL:          "https://web.sv-restaurant.ch/menu/Post,%20Restaurant%20Espace,%20Bern/Mittagsmen%C3%BC",
		HasCustomScraper: true,
		MenuType:         "html",
	},
	"turbolama": {
		Name:             "Turbolama",
		URL:              "https://www.turbolama.ch/",
		BaseURL:          "https://www.turbolama.ch/",
		HasCustomScraper: false,
		MenuType:         "pdf",
		MenuSelector:     "a[aria-label=\"FOOD MENU\"]",
	},
	"freibank": {
		Name: "Freibank",
		// would need selector: div[data-hook=\"app.container\
		URL:              "https://www.freibank.ch/speisundtrankangebot",
		BaseURL:          "https://www.freibank.ch/",
		HasCustomScraper: false,
		MenuType:         "html",
	},
}

// Config holds application configuration settings
type Config struct {
	DebugMode    bool   // If true, debug files will be written
	DryRun       bool   // If true, no API calls will be made
	RestaurantID string // ID of the restaurant to fetch menu from (defaults to "gira")
	UploadToR2   bool   // If true, upload parsed menu to R2 storage
}

// Run starts the application
func Run(config Config) error {
	// Load environment variables from .env file
	loadEnv()

	restaurantID := config.RestaurantID
	if restaurantID == "" {
		return fmt.Errorf("restaurant not defined")
	}

	// Get restaurant menu configuration
	restaurant, exists := restaurantMenus[restaurantID]
	if !exists {
		return fmt.Errorf("restaurant with ID '%s' not found", restaurantID)
	}

	log.Printf("Processing menu for %s from %s", restaurant.Name, restaurant.URL)

	// Handle different menu types (HTML or PDF)
	switch restaurant.MenuType {
	case "html":
		return processHTMLMenu(restaurant, config)
	case "pdf":
		return processPDFMenu(restaurant, config)
	default:
		return fmt.Errorf("unsupported menu type: %s", restaurant.MenuType)
	}
}

// processHTMLMenu handles HTML-based menus
func processHTMLMenu(restaurant RestaurantMenu, config Config) error {
	// Fetch the restaurant menu content
	log.Printf("Scraping HTML menu data for %s", restaurant.Name)
	var htmlContent *scraper.MenuData
	var err error

	if restaurant.HasCustomScraper {
		// Use custom scraper based on restaurant name
		switch strings.ToLower(restaurant.Name) {
		case "espace":
			htmlContent, err = scraper.ScrapeEspaceWebsite(restaurant.URL, config.DebugMode)
		default:
			err = fmt.Errorf("no custom scraper found for restaurant %s", restaurant.Name)
		}
	} else {
		// Use standard scraper
		htmlContent, err = scraper.ScrapeMenuContent(restaurant.URL, config.DebugMode)
	}

	if err != nil {
		return fmt.Errorf("error scraping menu data: %w", err)
	}

	// Save debug files if debug mode is enabled
	if config.DebugMode {
		htmlDebugFile, err := file.WriteToDebugFile([]byte(htmlContent.Content), "raw_html", restaurant.Name, "html")
		if err != nil {
			log.Printf("Warning: Could not write raw HTML to debug file: %v", err)
		} else {
			log.Printf("Saved raw HTML to %s", htmlDebugFile)
		}
	}

	// Extract menu content
	menuData := scraper.OptimizeHTML(htmlContent.Content)

	// Save debug files if debug mode is enabled
	if config.DebugMode {
		menuContentDebugFile, err := file.WriteToDebugFile([]byte(menuData), "menu_content", restaurant.Name, "html")
		if err != nil {
			log.Printf("Warning: Could not write menu content to debug file: %v", err)
		} else {
			log.Printf("Saved menu content to %s", menuContentDebugFile)
		}
	}

	contentLength := len(menuData)
	log.Printf("Successfully extracted menu content (%d bytes)", contentLength)

	if contentLength == 0 {
		return fmt.Errorf("no menu content found on the page")
	}

	// Print a sample of the content
	preview := menuData
	if len(preview) > 500 {
		preview = preview[:500] + "..."
	}
	log.Println("Menu Content Sample:")
	log.Println("=============")
	log.Println(preview)
	log.Println("=============")

	// Abort menu parsing if dry run is enabled
	if config.DryRun {
		log.Println("Dry Run, aborting parsing menu...")
		return nil
	}

	// Parse menu using OpenAI
	log.Println("Parsing menu data with OpenAI...")
	parsedMenu, err := ai.ParseRestaurantHtmlMenu(menuData)
	if err != nil {
		file.WriteToDebugFile(parsedMenu, "parsed_menu", restaurant.Name, "json")
		return fmt.Errorf("error parsing menu data: %w", err)
	}

	// Process the menu to add full URLs to links
	processedMenu, err := processMenuLinks(parsedMenu, restaurant.BaseURL)
	if err != nil {
		log.Printf("Warning: Could not process menu links: %v", err)
		processedMenu = parsedMenu // Fall back to original parsed menu
	}

	// Format JSON for output
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, processedMenu, "", "  "); err != nil {
		prettyJSON = *bytes.NewBuffer(processedMenu)
	}

	// Save debug files if debug mode is enabled
	if config.DebugMode {
		parsedMenuDebugFile, err := file.WriteToDebugFile(prettyJSON.Bytes(), "parsed_menu", restaurant.Name, "json")
		if err != nil {
			log.Printf("Warning: Could not write parsed menu to debug file: %v", err)
		} else {
			log.Printf("Saved parsed menu to %s", parsedMenuDebugFile)
		}
	}

	// Output the parsed menu
	fmt.Println(prettyJSON.String())

	// Upload to R2 if enabled
	if config.UploadToR2 {
		if err := uploadMenuToR2(processedMenu, restaurant.Name); err != nil {
			log.Printf("Warning: Failed to upload menu to R2: %v", err)
		} else {
			log.Println("Successfully uploaded menu to R2 storage")
		}
	}

	return nil
}

// processPDFMenu handles PDF-based menus
func processPDFMenu(restaurant RestaurantMenu, config Config) error {
	if restaurant.MenuSelector == "" {
		return fmt.Errorf("MenuSelector is required for PDF menu restaurants but not configured for %s", restaurant.Name)
	}

	log.Printf("Fetching PDF menu for %s", restaurant.Name)
	log.Printf("Looking for menu link with selector: %s", restaurant.MenuSelector)

	// Fetch the PDF menu URL using the selector
	pdfURL, err := scraper.FetchPDFMenuURL(restaurant.URL, restaurant.MenuSelector)
	if err != nil {
		return fmt.Errorf("error fetching PDF URL: %w", err)
	}

	var pdfFilePath string

	if config.DebugMode {
		// In debug mode, save the PDF to the debug directory
		pdfFilePath = filepath.Join("debug", fmt.Sprintf("%s_menu.pdf",
			strings.ToLower(restaurant.Name)))
		log.Printf("Debug mode: Saving PDF to %s", pdfFilePath)
	} else {
		// In production mode, save to a temporary directory
		tempDir, err := os.MkdirTemp("", "menu-pdf")
		if err != nil {
			return fmt.Errorf("error creating temporary directory: %w", err)
		}
		defer os.RemoveAll(tempDir)

		pdfFilePath = filepath.Join(tempDir, fmt.Sprintf("%s_menu.pdf",
			strings.ToLower(restaurant.Name)))
	}

	// Download the PDF file
	if err := scraper.DownloadPDF(pdfURL, pdfFilePath); err != nil {
		return fmt.Errorf("error downloading PDF: %w", err)
	}

	log.Printf("Successfully downloaded PDF menu for %s", restaurant.Name)

	// Abort menu parsing if dry run is enabled
	if config.DryRun {
		log.Println("Dry Run, aborting parsing menu...")
		return nil
	}

	// Extract text from PDF
	log.Println("Extracting text from PDF...")
	pdfText, err := scraper.ExtractTextFromPDF(pdfFilePath, 1) // Extract only first page
	if err != nil {
		return fmt.Errorf("error extracting text from PDF: %w", err)
	}

	// Save extracted text to debug file if debug mode is enabled
	if config.DebugMode {
		textDebugFile, err := file.WriteToDebugFile([]byte(pdfText), "extracted_text", restaurant.Name, "txt")
		if err != nil {
			log.Printf("Warning: Could not write extracted PDF text to debug file: %v", err)
		} else {
			log.Printf("Saved extracted PDF text to %s", textDebugFile)
		}
	}

	// Parse PDF menu using OpenAI
	log.Println("Parsing PDF menu data with OpenAI...")
	parsedMenu, err := ai.ParseRestaurantPdfMenu(pdfText, restaurant.Name, pdfURL)
	if err != nil {
		if config.DebugMode {
			file.WriteToDebugFile(parsedMenu, "parsed_menu", restaurant.Name, "json")
		}
		return fmt.Errorf("error parsing PDF menu data: %w", err)
	}

	// Format JSON for output
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, parsedMenu, "", "  "); err != nil {
		prettyJSON = *bytes.NewBuffer(parsedMenu)
	}

	// Save debug files if debug mode is enabled
	if config.DebugMode {
		parsedMenuDebugFile, err := file.WriteToDebugFile(prettyJSON.Bytes(), "parsed_menu", restaurant.Name, "json")
		if err != nil {
			log.Printf("Warning: Could not write parsed menu to debug file: %v", err)
		} else {
			log.Printf("Saved parsed menu to %s", parsedMenuDebugFile)
		}
	}

	// Output the parsed menu
	fmt.Println(prettyJSON.String())

	// Upload to R2 if enabled
	if config.UploadToR2 {
		if err := uploadMenuToR2(parsedMenu, restaurant.Name); err != nil {
			log.Printf("Warning: Failed to upload menu to R2: %v", err)
		} else {
			log.Println("Successfully uploaded menu to R2 storage")
		}
	}

	return nil
}

// processMenuLinks adds the restaurant's base URL to relative links in the menu
func processMenuLinks(menuJSON []byte, baseURL string) ([]byte, error) {
	var dailyMenu ai.DailyMenu
	if err := json.Unmarshal(menuJSON, &dailyMenu); err != nil {
		return nil, fmt.Errorf("failed to unmarshal menu JSON: %w", err)
	}

	// Process each day's menu items
	for day, items := range dailyMenu.Menu {
		for i, item := range items {
			if item.Link != "" && strings.HasPrefix(item.Link, "/") {
				dailyMenu.Menu[day][i].Link = baseURL + item.Link
			}
		}
	}

	processedJSON, err := json.Marshal(dailyMenu)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal processed menu: %w", err)
	}

	return processedJSON, nil
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

// uploadMenuToR2 uploads the menu JSON to Cloudflare R2 storage
func uploadMenuToR2(menuJSON []byte, restaurantName string) error {
	// Check for required environment variables
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	accessKeyID := os.Getenv("CLOUDFLARE_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("CLOUDFLARE_SECRET_ACCESS_KEY")
	bucketName := os.Getenv("CLOUDFLARE_BUCKET_NAME")

	if accountID == "" || accessKeyID == "" || secretAccessKey == "" || bucketName == "" {
		return fmt.Errorf("missing required Cloudflare R2 credentials in environment variables")
	}

	// Create an S3 client configured for Cloudflare R2
	endpoint := fmt.Sprintf("https://%s.eu.r2.cloudflarestorage.com", accountID)
	svc := s3.New(s3.Options{
		BaseEndpoint: aws.String(endpoint),
		Region:       "auto",
		Credentials:  credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, ""),
		UsePathStyle: true,
	})

	// Generate a filename with the restaurant name and current week number
	// Format: <restaurantname>_<weeknumber>_<year>.json
	year, week := time.Now().ISOWeek()
	lowercaseRestaurantName := strings.ToLower(restaurantName)
	filename := fmt.Sprintf("%s_%d_%d.json", lowercaseRestaurantName, week, year)

	// Upload the file to R2
	contentType := "application/json"
	_, err := svc.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(filename),
		Body:        bytes.NewReader(menuJSON),
		ContentType: &contentType,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file to R2: %w", err)
	}

	return nil
}
