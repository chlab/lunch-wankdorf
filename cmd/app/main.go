package main

import (
	"flag"
	"fmt"

	"github.com/leuenbergerc/lunch-wankdorf/internal/app"
)

func main() {
	// Define command line flags
	debugMode := flag.Bool("debug", false, "Enable debug mode with detailed output files")
	dryRun := flag.Bool("dryRun", false, "When enabled, no API calls will be made")
	flag.Parse()

	// Create config for the application
	config := app.Config{
		DebugMode: *debugMode,
		DryRun:    *dryRun,
	}

	fmt.Println("Starting Lunch Wankdorf application...")
	app.Run(config)
}
