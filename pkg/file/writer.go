package file

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WriteString writes a string to a file
func WriteString(content, fileName string) error {
	return WriteBytes([]byte(content), fileName)
}

// WriteBytes writes bytes to a file
func WriteBytes(content []byte, fileName string) error {
	fileName = strings.ToLower(fileName)
	// Create the directory if it doesn't exist
	dir := filepath.Dir(fileName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write the file
	if err := os.WriteFile(fileName, content, 0o644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", fileName, err)
	}

	return nil
}

// WriteToDebugFile writes content to a debug file in the specified directory
// The file will be named using the restaurant name and the specified filetype
func WriteToDebugFile(content []byte, label string, restaurantName string, fileType string) (string, error) {
	// Create debug directory if it doesn't exist
	debugDir := "debug"
	if err := os.MkdirAll(debugDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create debug directory: %w", err)
	}

	// Use default filetype if not specified
	if fileType == "" {
		fileType = "txt"
	}

	// Generate filename using restaurant name (overwrite existing files)
	fileName := filepath.Join(debugDir, fmt.Sprintf("%s_%s.%s", restaurantName, label, fileType))

	// Write the content to the file
	if err := WriteBytes(content, fileName); err != nil {
		return "", err
	}

	return fileName, nil
}
