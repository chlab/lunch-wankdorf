package scraper

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
)

// MenuData contains the scraped content
type MenuData struct {
	Content string
}

// ScrapeMenuContent retrieves only the relevant menu content from the URL
func ScrapeMenuContent(url string) (*MenuData, error) {
	menuData := &MenuData{}
	var menuContent strings.Builder

	// Initialize a Colly collector
	c := colly.NewCollector(
		// Adjust user agent to avoid being blocked
		colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36"),
	)

	// Look for main content divs that might contain the menu
	c.OnHTML("div", func(e *colly.HTMLElement) {
		html, err := e.DOM.Html()
		if err != nil {
			fmt.Printf("Error getting HTML: %v\n", err)
			return
		}
		menuContent.WriteString(html)
	})

	// Error handling
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	// Start scraping
	err := c.Visit(url)
	if err != nil {
		return nil, err
	}

	// If no content was found, try with specific menu-related classes or ids
	if menuContent.Len() == 0 {
		c2 := colly.NewCollector(
			colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36"),
		)

		c2.OnHTML("[class*='menu'], [id*='menu'], [class*='lunch'], [id*='lunch']", func(e *colly.HTMLElement) {
			html, err := e.DOM.Html()
			if err != nil {
				fmt.Printf("Error getting HTML: %v\n", err)
				return
			}
			menuContent.WriteString(html)
		})

		err = c2.Visit(url)
		if err != nil {
			return nil, err
		}
	}

	// If we still don't have content, get the main content area as a fallback
	if menuContent.Len() == 0 {
		c3 := colly.NewCollector(
			colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36"),
		)

		c3.OnHTML("div.container > div, main, article, #content", func(e *colly.HTMLElement) {
			html, err := e.DOM.Html()
			if err != nil {
				fmt.Printf("Error getting HTML: %v\n", err)
				return
			}
			menuContent.WriteString(html)
		})

		err = c3.Visit(url)
		if err != nil {
			return nil, err
		}
	}

	// Last resort: get the entire body content
	if menuContent.Len() == 0 {
		c4 := colly.NewCollector(
			colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36"),
		)

		c4.OnHTML("body", func(e *colly.HTMLElement) {
			html, err := e.DOM.Html()
			if err != nil {
				fmt.Printf("Error getting HTML: %v\n", err)
				return
			}
			menuContent.WriteString(html)
		})

		err = c4.Visit(url)
		if err != nil {
			return nil, err
		}
	}

	// Minify the HTML content
	rawContent := menuContent.String()
	if len(rawContent) > 0 {
		// Initialize minifier
		m := minify.New()
		m.AddFunc("text/html", html.Minify)

		// Create input and output buffers
		input := bytes.NewBufferString(rawContent)
		output := &bytes.Buffer{}

		// Minify the content
		err := m.Minify("text/html", output, input)
		if err != nil {
			fmt.Printf("Error minifying HTML: %v, using original content\n", err)
			menuData.Content = rawContent
		} else {
			minified := output.String()
			fmt.Printf("Minified HTML from %d to %d bytes (%.1f%%)\n", 
				len(rawContent), 
				len(minified), 
				float64(len(minified))/float64(len(rawContent))*100)
			menuData.Content = minified
		}
	} else {
		menuData.Content = rawContent
	}

	return menuData, nil
}