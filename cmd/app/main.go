package main

import (
	"flag"
	"log"

	"github.com/chlab/lunch-wankdorf/internal/app"
)

func main() {
	// Define command line flags
	debugMode := flag.Bool("debug", false, "Enable debug mode with detailed output files")
	dryRun := flag.Bool("dryRun", false, "When enabled, no API calls will be made")
	restaurantID := flag.String("restaurant", "gira", "ID of the restaurant to fetch menu from")
	uploadToR2 := flag.Bool("upload", false, "Upload parsed menu to Cloudflare R2 storage")
	photosOnly := flag.Bool("photos", false, "Only add newly published dish photos to the menu, without re-parsing it")
	flag.Parse()

	// Create config for the application
	config := app.Config{
		DebugMode:    *debugMode,
		DryRun:       *dryRun,
		RestaurantID: *restaurantID,
		UploadToR2:   *uploadToR2,
	}

	log.Println("Starting Lunch Wankdorf application...")

	run := app.Run
	if *photosOnly {
		run = app.RunPhotoUpdate
	}

	if err := run(config); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
