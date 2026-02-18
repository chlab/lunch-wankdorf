# AGENTS.md

This file is for coding agents working in this repository.

## What this project does

- Backend (`Go`): scrapes lunch menus from multiple restaurant sources (HTML/PDF), parses them with OpenAI, and uploads weekly JSON files to Cloudflare R2.
- Frontend (`Vue 3 + Vite + Tailwind`): fetches the current week JSON files from R2 and renders a daily lunch view.

## Repo map

- `cmd/app/main.go`: CLI entrypoint.
- `internal/app/app.go`: orchestration (restaurant config, scraping mode, OpenAI parse, R2 upload).
- `pkg/scraper/*`: HTML scraping (`colly`), custom scraping for Espace-restaurant (`chromedp`), PDF URL/download/text extraction, HTML cleanup.
- `pkg/ai/openai.go`: OpenAI prompting and JSON normalization.
- `pkg/file/writer.go`: debug/output file writers.
- `web/`: frontend app.
- `scripts/`: helper scripts for upload/prune/publish.
- `.github/workflows/`: weekly fetch, weekly prune, and frontend deploy.

## Local runbook

- Backend, single restaurant:
  - `go run ./cmd/app/main.go -restaurant gira -dryRun`
  - `go run ./cmd/app/main.go -restaurant espace -debug`
  - `go run ./cmd/app/main.go -restaurant turbolama -upload`
- Backend, all restaurants:
  - `bash ./scripts/upload-all-menus.sh`
- Frontend:
  - `cd web && npm ci`
  - `cd web && npm run dev`
  - `cd web && npm run build`

## Required environment variables

- `OPENAI_API_KEY`
- `CLOUDFLARE_ACCOUNT_ID`
- `CLOUDFLARE_ACCESS_KEY_ID`
- `CLOUDFLARE_SECRET_ACCESS_KEY`
- `CLOUDFLARE_BUCKET_NAME`

Notes:
- `.env` is loaded by the backend (`internal/app/loadEnv`), searching up to 3 parent levels.
- Never commit `.env` or secrets.

## Data contracts and naming

- Uploaded files use: `<restaurant>_<iso_week>_<year>.json` (for example `gira_8_2026.json`).
- JSON shape expected by frontend:
  - Daily menus: `{ "type": "daily", "menu": { "Monday": [ ... ] } }`
  - Weekly menus: `{ "type": "weekly", "menu": [ ... ] }`
- Frontend computes current week/year and fetches each configured restaurant file from:
  - `https://pub-201cbf927f0b4c8991d32485a57b9d40.r2.dev`
- Frontend Vite base path is `/lunch-wankdorf` (GitHub Pages deploy target).

## Current frontend toggles/behavior

- Static foodtrucks data is in `web/src/foodtrucks.json`.
- Foodtrucks are currently gated by `foodtrucksEnabled` in `web/src/App.vue` (set to `false` right now).

## Change guidelines

- Keep restaurant IDs stable (`gira`, `luna`, `sole`, `espace`, `turbolama`, `freibank`) unless a migration is intentional.
- If you add/remove a restaurant:
  - Update `restaurantMenus` in `internal/app/app.go`.
  - Update frontend `getMenuFiles()` in `web/src/App.vue`.
  - Update scripts/workflows matrix if needed.
- Keep JSON output backward-compatible with frontend parsing unless both sides are updated in one change.
- Prefer minimal, targeted edits; avoid broad refactors in this repo unless explicitly requested.

## Validation checklist before handoff

- Backend: `go test ./...` (or at least `go run ./cmd/app/main.go -restaurant <id> -dryRun`).
- Frontend: `cd web && npm run build`.
- If touching scraping logic, validate with `-debug` and inspect `debug/` artifacts.

## CI/CD summary

- `weekly-menu-fetch.yml`: runs Mondays 05:00 UTC, fetches/upload menus (matrix of restaurants, currently excluding `freibank`).
- `weekly-menu-prune.yml`: runs Sundays 23:00 UTC, removes old R2 menu files.
- `deploy.yml`: builds `web/` and deploys to GitHub Pages on push to `main`.
