# Lunch Wankdorf

A Go application for generating lunch options using the OpenAI API.

## Project Structure

This project follows the [Standard Go Project Layout](https://github.com/golang-standards/project-layout).

- `/cmd/app`: Main application entry point
- `/internal/app`: Application-specific code not meant to be used by external applications
- `/pkg`: Library code that's ok to use by external applications
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

Note: The `.env` file is git-ignored to prevent accidentally committing your API key.