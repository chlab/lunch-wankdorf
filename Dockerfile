# Build stage
FROM golang:1.24.3 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY pkg/ ./pkg/

RUN go build -o lunch-app ./cmd/app/main.go

# Final stage
FROM chromedp/headless-shell:stable

# Install git for any additional dependencies
RUN apt-get update && apt-get install -y git \
    && apt-get clean && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/lunch-app .

# Create a non-root user and switch to it
RUN useradd -m lunch
USER lunch

ENV OPENAI_API_KEY=""
ENV CLOUDFLARE_ACCOUNT_ID=""
ENV CLOUDFLARE_ACCESS_KEY_ID=""
ENV CLOUDFLARE_SECRET_ACCESS_KEY=""
ENV CLOUDFLARE_BUCKET_NAME=""

ENTRYPOINT ["./lunch-app"]
