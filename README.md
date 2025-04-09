# Lunch Wankdorf

A Go application for fetching and parsing the weekly lunch menu from the Wankdorf restaurant using OpenAI.

## Project Structure

This project follows the [Standard Go Project Layout](https://github.com/golang-standards/project-layout).

- `/cmd/app`: Main application entry point
- `/internal/app`: Application-specific code not meant to be used by external applications
- `/pkg`: Library code that's ok to use by external applications
  - `/pkg/ai`: OpenAI API integration
  - `/pkg/scraper`: Web scraping functionality using Colly
- `/scripts`: Scripts to perform various build, install, analysis, etc operations

## Requirements

- Go 1.22 or later
- OpenAI API key

## Running the application

1. Create a `.env` file in the project root with your OpenAI API key:
   ```
   cp .env.example .env
   ```
   Then edit the `.env` file to add your actual API key.

2. Run the application:
   ```bash
   ./scripts/run.sh
   ```

The application will:
1. Scrape the weekly menu from https://app.food2050.ch/de/sbb-gira/gira/menu/mittagsmenue/weekly
   - Uses Colly to extract only the relevant menu HTML content
2. Send the targeted HTML content to OpenAI for parsing
3. Display the structured menu data in JSON format with days of the week and menu options

Note: The `.env` file is git-ignored to prevent accidentally committing your API key.