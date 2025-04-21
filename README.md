# Lunch Wankdorf

A Go application for fetching and parsing the weekly lunch menu from the Wankdorf restaurants using OpenAI.
The result is published as a [little web app](https://chlab.github.io/lunch-wankdorf/).

The project was hacked together in two evenings of vibe-coding, so don't judge my code too harshly.

## Todo

- [x] Add support for Post Espace restaurant. This one is bit trickier since it's an Angular app and Colly doesn't render JS. Also, each day is a different route, so the renderer would actually need to click around on the page.
- [x] Run the scraper in a nightly GitHub Action
- [ ] Gzip the menus before uploading them to R2
- [x] Add all known foodtrucks in the frontend
- [ ] Add Turbo Lama and maybe Freibank
- [ ] Maybe generate icons for each menu item
- [x] Change the JSON structure of the menus to allow for daily and weekly menus
- [ ] Consider adding the restaurant name to the menu instead of adding it in the frontend
- [ ] Add food_type field (healthy, fast food, streetfood, etc.)

## Project Structure

This project follows the [Standard Go Project Layout](https://github.com/golang-standards/project-layout).

- `/cmd/app`: Main application entry point
- `/internal/app`: Application-specific code not meant to be used by external applications
- `/pkg`: Library code that's ok to use by external applications
  - `/pkg/ai`: OpenAI API integration
  - `/pkg/scraper`: Web scraping functionality using Colly
- `/scripts`: Scripts to perform various build, install, analysis, etc operations
- `/web`: Vuejs frontend

## Requirements

- Go 1.22 or later
- OpenAI API key
- Cloudflare R2 bucket with a `lunch-wankdorf` bucket and access key

## Running the application

1. Create a `.env` file in the project root with your OpenAI API key:
   ```
   cp .env.example .env
   ```
   Then edit the `.env` file to add your actual API key.

2. Run the application:
   ```bash
   ./scripts/run.sh
   # or
   go run ./cmd/app/main.go -h
   ```

The application will:
1. Scrape the weekly menu from the three SBB restaurants
   - Uses Colly to extract only the relevant menu HTML content
2. Send the targeted HTML content to OpenAI for parsing
3. Parse the structured menu data in JSON format with days of the week and menu options
4. Upload the structured menu data to a Cloudflare R2 bucket

The frontend app retrieves the structured menu data from the Cloudflare R2 bucket and displays it.
