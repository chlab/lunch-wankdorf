#!/bin/bash
# Fetch this week's menu for every restaurant and publish it to R2.
#
# Needs OPENAI_API_KEY and the CLOUDFLARE_* credentials (a .env in the project root
# is picked up automatically).

set -e

# Turbolama and Freibank are disabled: they kept moving their menu around, so
# nothing is scraped for them.
RESTAURANTS=("gira" "luna" "sole" "espace")

cd "$(dirname "$(dirname "$0")")"

echo "Fetching and uploading menus for: ${RESTAURANTS[*]}"

for restaurant in "${RESTAURANTS[@]}"; do
  echo "========================================"
  echo "Processing restaurant: $restaurant"
  echo "========================================"

  go run ./cmd/app -restaurant="$restaurant" -upload

  echo "Completed processing for $restaurant"
done

echo "All restaurants processed successfully!"
