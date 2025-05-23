#!/bin/bash

# Check if AWS CLI is installed
if ! command -v aws &> /dev/null; then
    echo "Error: AWS CLI is not installed."
    echo "Please install it first:"
    echo "  - macOS: brew install awscli"
    echo "  - Ubuntu/Debian: sudo apt-get install awscli"
    echo "  - Or visit: https://aws.amazon.com/cli/"
    exit 1
fi

# Load environment variables if .env file exists
ENV_FILE="$(dirname "$0")/../.env"
if [ -f "$ENV_FILE" ]; then
    set -a
    source "$ENV_FILE"
    set +a
fi

# Calculate the cutoff date (one week ago)
# Handle different date command syntax (macOS vs Linux)
if date -v-1w +%Y-%m-%d >/dev/null 2>&1; then
    # macOS
    CUTOFF_DATE=$(date -v-1w +%Y-%m-%d)
    CUTOFF_WEEK=$(date -v-1w +%V)
    CUTOFF_YEAR=$(date -v-1w +%Y)
else
    # Linux
    CUTOFF_DATE=$(date -d '1 week ago' +%Y-%m-%d)
    CUTOFF_WEEK=$(date -d '1 week ago' +%V)
    CUTOFF_YEAR=$(date -d '1 week ago' +%Y)
fi

echo "Pruning menu files older than week $CUTOFF_WEEK of $CUTOFF_YEAR (date: $CUTOFF_DATE)"

# Ask for confirmation
if [[ "$1" == "--yes" ]]; then
    CONFIRM="y"
else
    read -p "Do you want to continue? (y/N): " -n 1 -r CONFIRM
    echo
fi

if [[ ! $CONFIRM =~ ^[Yy]$ ]]; then
    echo "Operation aborted."
    exit 0
fi

# Configure AWS CLI for Cloudflare R2
export AWS_ACCESS_KEY_ID="$CLOUDFLARE_ACCESS_KEY_ID"
export AWS_SECRET_ACCESS_KEY="$CLOUDFLARE_SECRET_ACCESS_KEY"
export AWS_DEFAULT_REGION="auto"
R2_ENDPOINT="https://$CLOUDFLARE_ACCOUNT_ID.eu.r2.cloudflarestorage.com"

# List all objects in the bucket
aws s3api list-objects-v2 \
    --endpoint-url "$R2_ENDPOINT" \
    --bucket "$CLOUDFLARE_BUCKET_NAME" \
    --query 'Contents[].Key' \
    --output text | tr '\t' '\n' | while read -r file; do
    
    # Skip empty lines
    if [ -z "$file" ]; then
        continue
    fi
    
    # Extract week and year from filename (format: restaurantname_weeknumber_year.json)
    if [[ $file =~ _([0-9]+)_([0-9]{4})\.json$ ]]; then
        FILE_WEEK="${BASH_REMATCH[1]}"
        FILE_YEAR="${BASH_REMATCH[2]}"
        
        # Convert week numbers to comparable format (remove leading zeros)
        FILE_WEEK_NUM=$((10#$FILE_WEEK))
        CUTOFF_WEEK_NUM=$((10#$CUTOFF_WEEK))
        
        # Check if file is older than cutoff
        if [ "$FILE_YEAR" -lt "$CUTOFF_YEAR" ] || \
           ([ "$FILE_YEAR" -eq "$CUTOFF_YEAR" ] && [ "$FILE_WEEK_NUM" -lt "$CUTOFF_WEEK_NUM" ]); then
            echo "Deleting old file: $file (week $FILE_WEEK, year $FILE_YEAR)"
            aws s3 rm "s3://$CLOUDFLARE_BUCKET_NAME/$file" --endpoint-url "$R2_ENDPOINT"
        else
            echo "Keeping current file: $file (week $FILE_WEEK, year $FILE_YEAR)"
        fi
    else
        echo "Skipping file with unexpected format: $file"
    fi
done

echo "Pruning completed."