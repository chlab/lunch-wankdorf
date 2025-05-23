#!/bin/bash
# Script to run main.go for all available restaurants and upload them to R2

set -e

echo "Starting script to fetch and upload menus for all restaurants..."

RESTAURANTS=("gira" "luna" "sole" "espace" "turbolama" "freibank")

MAIN_DIR="$(dirname "$(dirname "$0")")/cmd/app"
cd "$(dirname "$(dirname "$0")")"

for restaurant in "${RESTAURANTS[@]}"; do
  echo "========================================"
  echo "Processing restaurant: $restaurant"
  echo "========================================"
  
  go run "$MAIN_DIR/main.go" -restaurant="$restaurant" -upload
  
  echo "Completed processing for $restaurant"
  
  # Add a short delay between runs to avoid API rate limits
  sleep 2
done

echo "All restaurants processed successfully!"