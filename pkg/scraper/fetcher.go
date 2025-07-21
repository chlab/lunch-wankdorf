package scraper

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/chromedp/cdproto/browser"
	"github.com/chromedp/chromedp"
	"github.com/gocolly/colly/v2"
)

// MenuData contains the scraped content
type MenuData struct {
	Content string
}

// ScrapeMenuContent retrieves only the relevant menu content from the URL
func ScrapeMenuContent(url string, debugMode ...bool) (*MenuData, error) {
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

	menuData.Content = menuContent.String()

	return menuData, nil
}

// FetchPDFMenuURL retrieves a PDF menu URL from a website using a CSS selector
func FetchPDFMenuURL(url string, menuSelector string) (string, error) {
	// Create a new collector
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36"),
	)

	var pdfURL string
	var pdfFound bool
	// Keep track of whether we've already found a link
	firstItem := true

	// Find the menu link using the provided selector
	c.OnHTML(menuSelector, func(e *colly.HTMLElement) {
		// Only process the first match
		if firstItem {
			pdfURL = e.Attr("href")
			if pdfURL != "" {
				pdfFound = true
				firstItem = false
				fmt.Printf("Found menu PDF URL: %s\n", pdfURL)
			}
		}
	})

	// Handle errors
	c.OnError(func(r *colly.Response, err error) {
		fmt.Printf("Request URL: %s failed with error: %v\n", r.Request.URL, err)
	})

	// Set timeout for the request
	c.SetRequestTimeout(30 * time.Second)

	// Start scraping
	err := c.Visit(url)
	if err != nil {
		return "", fmt.Errorf("error visiting %s: %v", url, err)
	}

	// Check if a PDF URL was found
	if !pdfFound {
		return "", fmt.Errorf("no menu PDF link found on the page using selector: %s", menuSelector)
	}

	return pdfURL, nil
}

// DownloadPDF downloads a PDF from the given URL and saves it to the specified file path
func DownloadPDF(pdfURL, outputPath string) error {
	// Create output file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("error creating file %s: %v", outputPath, err)
	}
	defer file.Close()

	// Download the PDF
	fmt.Printf("Downloading PDF from %s to %s...\n", pdfURL, outputPath)
	resp, err := http.Get(pdfURL)
	if err != nil {
		return fmt.Errorf("error downloading PDF: %v", err)
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad server response: %s", resp.Status)
	}

	// Copy PDF data to file
	bytesWritten, err := io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("error saving PDF data: %v", err)
	}

	fmt.Printf("Download complete! %d bytes written to %s\n", bytesWritten, outputPath)
	return nil
}

// ScrapeEspaceWebsite is a custom scraper for the SV Espace restaurant website
// It navigates through all weekday tabs and combines the menus into a single HTML content
func ScrapeEspaceWebsite(url string, debug bool) (*MenuData, error) {
	opts := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.DisableGPU,
		chromedp.WindowSize(1280, 800),
		chromedp.NoSandbox,
	}

	// Don't run in headless mode if debug mode is enabled
	if debug {
		fmt.Println("Debug mode enabled: Chrome browser will stay open for inspection")
		opts = append(opts, chromedp.Flag("headless", false))          // Disable headless mode
		opts = append(opts, chromedp.Flag("enable-automation", false)) // Hide automation banner
		fmt.Println("✓ Configured Chrome to run in visible mode")
	} else {
		opts = append(opts, chromedp.Flag("headless", true)) // Enable headless mode
	}

	allocCtx, cancelAllocator := chromedp.NewExecAllocator(context.Background(), opts...)
	if !debug {
		defer cancelAllocator()
	}

	ctx, cancel := chromedp.NewContext(allocCtx,
		chromedp.WithLogf(func(format string, args ...interface{}) {
			if debug {
				fmt.Printf("ChromeDP: "+format+"\n", args...)
			}
		}),
	)

	ctx, cancelTimeout := context.WithTimeout(ctx, 60*time.Second)

	if !debug {
		defer cancel()
	}
	defer cancelTimeout()

	// Combined menu content from all weekdays
	var allMenus strings.Builder

	// Navigate to website and handle initial setup
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		// Do not grant any permissions to avoid the geolocation prompt
		browser.GrantPermissions([]browser.PermissionType{}),
		// Wait for the weekday navigation to load
		chromedp.WaitVisible(`[mat-tab-link]`, chromedp.ByQuery),
		// Decline cookies if the banner appears
		chromedp.Click(`#cookiescript_reject`, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
	)
	if err != nil {
		if debug {
			fmt.Printf("Error during initial page setup: %v\n", err)
			fmt.Println("Keeping browser open for inspection. Press Ctrl+C to exit.")
			select {} // Block indefinitely in debug mode
		}
		return nil, fmt.Errorf("failed to setup page: %v", err)
	}

	// Loop through each weekday (Monday to Friday)
	weekdays := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday"}
	for day := 1; day <= 5; day++ {
		var dayMenu string
		dayName := weekdays[day-1]

		err := chromedp.Run(ctx,
			// Click on the tab for the current weekday
			chromedp.Click(fmt.Sprintf(`[mat-tab-link]:nth-child(%d)`, day), chromedp.ByQuery),
			chromedp.Sleep(1*time.Second),
			// Extract the menu HTML
			chromedp.OuterHTML(`app-menu-container`, &dayMenu),
		)
		if err != nil {
			if debug {
				fmt.Printf("Error scraping %s menu: %v\n", dayName, err)
				continue // Try next day in debug mode
			}
			return nil, fmt.Errorf("failed to scrape %s menu: %v", dayName, err)
		}

		// Add day header and append to combined content
		allMenus.WriteString(fmt.Sprintf("\n<!-- %s Menu -->\n<h2>%s</h2>\n%s\n", dayName, dayName, dayMenu))

		fmt.Printf("Scraped %s menu (length: %d bytes)\n", dayName, len(dayMenu))
	}

	// Get the combined content
	htmlContent := allMenus.String()

	if debug {
		fmt.Printf("Successfully scraped all weekly menus (total: %d bytes)\n", len(htmlContent))
		fmt.Println("Keeping browser open for inspection. Press Ctrl+C to exit.")
		select {} // Block indefinitely in debug mode
	}

	return &MenuData{Content: htmlContent}, nil
}
