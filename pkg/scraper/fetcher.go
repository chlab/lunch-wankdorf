package scraper

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/chromedp/cdproto/browser"
	"github.com/chromedp/chromedp"
	"github.com/gocolly/colly/v2"
)

const (
	userAgent          = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36"
	httpRequestTimeout = 30 * time.Second
	chromeTimeout      = 60 * time.Second
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
		colly.UserAgent(userAgent),
	)

	// Look for main content divs that might contain the menu
	c.OnHTML("div", func(e *colly.HTMLElement) {
		html, err := e.DOM.Html()
		if err != nil {
			log.Printf("Error getting HTML: %v", err)
			return
		}
		menuContent.WriteString(html)
	})

	// Error handling
	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Request URL: %s failed with response: %v, Error: %v", r.Request.URL, r, err)
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
		colly.UserAgent(userAgent),
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
				log.Printf("Found menu PDF URL: %s", pdfURL)
			}
		}
	})

	// Handle errors
	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Request URL: %s failed with error: %v", r.Request.URL, err)
	})

	// Set timeout for the request
	c.SetRequestTimeout(httpRequestTimeout)

	// Start scraping
	err := c.Visit(url)
	if err != nil {
		return "", fmt.Errorf("error visiting %s: %w", url, err)
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
		return fmt.Errorf("error creating file %s: %w", outputPath, err)
	}
	defer file.Close()

	// Download the PDF
	log.Printf("Downloading PDF from %s to %s...", pdfURL, outputPath)
	client := &http.Client{Timeout: httpRequestTimeout}
	resp, err := client.Get(pdfURL)
	if err != nil {
		return fmt.Errorf("error downloading PDF: %w", err)
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad server response: %s", resp.Status)
	}

	// Copy PDF data to file
	bytesWritten, err := io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("error saving PDF data: %w", err)
	}

	log.Printf("Download complete! %d bytes written to %s", bytesWritten, outputPath)
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
		log.Println("Debug mode enabled: Chrome browser will stay open for inspection")
		opts = append(opts, chromedp.Flag("headless", false))          // Disable headless mode
		opts = append(opts, chromedp.Flag("enable-automation", false)) // Hide automation banner
		log.Println("Configured Chrome to run in visible mode")
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
				log.Printf("ChromeDP: "+format, args...)
			}
		}),
	)

	ctx, cancelTimeout := context.WithTimeout(ctx, chromeTimeout)

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
		// Decline cookies if the banner appears (ignore if not present)
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Try to click the cookie reject button, but don't fail if it's not there
			err := chromedp.Click(`#cookiescript_reject`, chromedp.ByQuery).Do(ctx)
			if err != nil {
				log.Printf("Cookie banner not found or not clickable, continuing: %v", err)
			}
			return nil
		}),
		chromedp.Sleep(2*time.Second),
	)
	if err != nil {
		if debug {
			log.Printf("Error during initial page setup: %v", err)
			waitForInterrupt()
		}
		return nil, fmt.Errorf("failed to setup page: %w", err)
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
				log.Printf("Error scraping %s menu: %v", dayName, err)
				continue // Try next day in debug mode
			}
			return nil, fmt.Errorf("failed to scrape %s menu: %w", dayName, err)
		}

		// Add day header and append to combined content
		allMenus.WriteString(fmt.Sprintf("\n<!-- %s Menu -->\n<h2>%s</h2>\n%s\n", dayName, dayName, dayMenu))

		log.Printf("Scraped %s menu (length: %d bytes)", dayName, len(dayMenu))
	}

	// Get the combined content
	htmlContent := allMenus.String()

	if debug {
		log.Printf("Successfully scraped all weekly menus (total: %d bytes)", len(htmlContent))
		waitForInterrupt()
	}

	return &MenuData{Content: htmlContent}, nil
}

// waitForInterrupt blocks until an interrupt signal (Ctrl+C) is received,
// allowing the debug browser to stay open for inspection.
func waitForInterrupt() {
	log.Println("Keeping browser open for inspection. Press Ctrl+C to exit.")
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	log.Println("Received interrupt, shutting down.")
}
