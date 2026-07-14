# Lunch Wankdorf

A Go application for fetching and parsing the weekly lunch menu from the Wankdorf restaurants using OpenAI.
The result is published as a [little web app](https://chlab.github.io/lunch-wankdorf/).

The project was hacked together in two evenings of vibe-coding, so don't judge my code too harshly.

## Todo

- [x] Add support for Post Espace restaurant. This one is bit trickier since it's an Angular app and Colly doesn't render JS. Also, each day is a different route, so the renderer would actually need to click around on the page.
- [x] Run the scraper in a nightly GitHub Action
- [ ] Gzip the menus before uploading them to R2
- [x] Add all known foodtrucks in the frontend
- [x] Add Turbo Lama and maybe Freibank
- [x] Maybe generate icons for each menu item
- [x] Change the JSON structure of the menus to allow for daily and weekly menus
- [ ] Consider adding the restaurant name to the menu instead of adding it in the frontend
- [ ] Add food_type field (healthy, fast food, streetfood, etc.)
- [x] Strip quotes from the menu titles, also correct turbolama titles to capitalized instead of all caps
- [ ] Fix the menu links by adding a base URL to the restaurants

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

- Go 1.24 or later
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

Useful flags: `-restaurant <id>` picks the restaurant, `-dryRun` scrapes without
calling OpenAI, `-debug` writes the intermediate HTML to `debug/`, and `-upload`
publishes to R2. A weekly GitHub Action runs the whole set every Monday morning.

The application will:
1. Scrape the weekly menu for one restaurant
2. Split the week into one section per day (see below)
3. Send each day to OpenAI separately and check the result against the dishes the page offered
4. Add the dish photos (see below)
5. Upload the structured menu data to a Cloudflare R2 bucket

The frontend app retrieves the structured menu data from the Cloudflare R2 bucket and displays it.

## Dish photos

Where a restaurant has a photo of a dish, the frontend shows it instead of the
icon. Only about 60% of the SBB dishes have one, so the icon is not going away.

The two restaurant types publish photos differently, which is why there are two
jobs:

- **Gira, Luna, Sole** keep the photo on each dish's own page. The menu items
  already carry that link, so the weekly run follows it and reads the photo out of
  the page's embedded data. The whole week is available up front.
- **Espace** has the photo on the menu page itself, but only publishes a day's
  photos on the morning of that day (they appear around 07:30) — the rest of the
  week serves a placeholder. `Daily Photo Fetch` therefore runs every weekday and
  fills them in as they appear:

  ```bash
  go run ./cmd/app -restaurant espace -photos -upload
  ```

  It deliberately does **not** re-parse the menu. It scrapes the photos, fills the
  blanks in the published JSON and puts it back. Re-running the model daily would
  pay it to rewrite text we already have, and risk regressing a menu that was
  already correct.

Photos are matched to dishes on the **category** heading ("Chefs Choice", "Pizza
Del Giorno"), which the model copies verbatim — the dish name is no good as a key,
because the model rewrites it. If the categories ever stop lining up, the job says
so rather than quietly adding nothing.

No image is ever downloaded or stored: both restaurants' CDNs resize on request, so
the menu carries a small URL for the list and a larger one for the lightbox.

## How the parsing works, and why

The naive version of this — hand OpenAI the whole week and ask for a menu — quietly
loses dishes, and it always loses them at the *end* of the week. That is where the
Friday menus kept disappearing to. Two things prevent it, and both are needed:

**Every dish is assigned to a day by us, never by the model.**
Gira, Luna and Sole (food2050) render the week as a transposed grid: dishes are
grouped by category ("Pasta Del Giorno") with one link per weekday, and the only
thing tying a dish to a day is the date at the end of its link. `GroupMenuByDay`
reads that date and splits the page into days. Espace (SV) is a different site
entirely — an Angular app where each day is its own dated route — so its scraper
loads each day by its own URL and waits for the page to actually show that day
before capturing it.

**Each day is parsed in its own request.**
A day is small enough for the model to read in full, and the day it belongs to is
never in question. Days are parsed in parallel, and because we know how many dishes
the page offered, a day that comes back short is retried on its own. If a page ever
stops exposing dates, the run fails loudly rather than uploading a menu with days
silently missing.

This is not a model problem you can buy your way out of: asked to parse a whole
week in one call, even `gpt-5.4-mini` still drops the tail (Espace lost a Friday
dish in 3 of 3 runs).

## Choosing a model

`OPENAI_MODEL` overrides the model; the default is in `pkg/ai/openai.go`.

The model is doing extraction, not reasoning, but it still has to *not get bored*.
Parsing Gira one day at a time — 3 runs of 5 days, counting the returned dishes
against what the page offered:

| Model | Days complete | Dishes lost | Time |
|---|---|---|---|
| gpt-4.1-mini | 9/15 | 18 of 90 | 136s |
| gpt-4.1 | 15/15 | 0 | 73s |
| gpt-5-mini | 15/15 | 0 | 206s |
| **gpt-5.4-mini** (default) | **15/15** | **0** | **54s** |
| gpt-5.4-nano | 15/15 | 0 | 76s |
| gpt-5.4 | 15/15 | 0 | 65s |

Cost is irrelevant at this volume: a weekly run is 4 restaurants × 5 days = 20 calls
of roughly 860 input and 400 output tokens, which is about 5 cents a week on
gpt-5.4-mini — a couple of euros a year. Pick for reliability, not price.

## Frontend

```bash
cd web
npm install
npm run dev     # the R2 bucket sends no CORS header for localhost, see below
npm run lint
npm run format
```

The bucket can't be read from `localhost`, so `npm run dev` on its own shows the
error state. Point it at a local copy of the menus to work on the UI:

```bash
VITE_MENU_BASE_URL=http://localhost:8099 npm run dev
```

## Credits

* Icons from Plasticine [Icons8](https://icons8.com/icons/set/food--style-plasticine/)