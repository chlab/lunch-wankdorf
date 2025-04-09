package file

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// WriteString writes a string to a file
func WriteString(content, fileName string) error {
	// Create the directory if it doesn't exist
	dir := filepath.Dir(fileName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write the file
	if err := os.WriteFile(fileName, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", fileName, err)
	}

	return nil
}

// WriteToDebugFile writes content to a debug file in the specified directory
func WriteToDebugFile(content, label string) (string, error) {
	// Create debug directory if it doesn't exist
	debugDir := "debug"
	if err := os.MkdirAll(debugDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create debug directory: %w", err)
	}

	// Generate filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	fileName := filepath.Join(debugDir, fmt.Sprintf("%s_%s.txt", label, timestamp))

	// Write the content to the file
	if err := WriteString(content, fileName); err != nil {
		return "", err
	}

	return fileName, nil
}
