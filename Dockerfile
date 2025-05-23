FROM chromedp/headless-shell:stable

RUN apt-get update && apt-get install -y ca-certificates \
    && apt-get clean && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Create a non-root user and switch to it
RUN useradd -m lunch
USER lunch
