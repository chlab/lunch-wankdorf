package scraper

import (
	"fmt"
	"strings"

	"rsc.io/pdf"
)

// ExtractTextFromPDF extracts text content from a PDF file
// maxPages: the maximum number of pages to extract (0 for all)
func ExtractTextFromPDF(pdfPath string, maxPages int) (string, error) {
	// Open and read the PDF file
	file, err := pdf.Open(pdfPath)
	if err != nil {
		return "", fmt.Errorf("error opening PDF: %w", err)
	}

	// Get the total number of pages
	numPages := file.NumPage()

	// Determine how many pages to extract
	if maxPages <= 0 || maxPages > numPages {
		maxPages = numPages
	}

	// Build the extracted text
	var allText strings.Builder
	for pageNum := 1; pageNum <= maxPages; pageNum++ {
		page := file.Page(pageNum)
		if page.V.IsNull() {
			continue // Skip invalid pages
		}

		// Extract text content from the page
		content := page.Content()
		if len(content.Text) == 0 {
			allText.WriteString(fmt.Sprintf("--- Page %d [No text content found] ---\n\n", pageNum))
		} else {
			allText.WriteString(fmt.Sprintf("--- Page %d ---\n", pageNum))
			// Process all text elements on the page
			for _, text := range content.Text {
				allText.WriteString(text.S)
				allText.WriteString(" ")
			}
			allText.WriteString("\n\n")
		}
	}

	return allText.String(), nil
}