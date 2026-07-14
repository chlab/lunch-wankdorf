package scraper

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	neturl "net/url"
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
	chromeTimeout      = 2 * time.Minute
	cookieClickTimeout = 5 * time.Second
	daySwitchTimeout   = 15 * time.Second
)

// MenuData contains the scraped content
type MenuData struct {
	Content string
	// Days holds the menu split into one section per weekday, for scrapers that can
	// do the split themselves. Nil when the content still has to be split (see
	// GroupMenuByDay).
	Days []DayMenu
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

// ScrapeEspaceWebsite is a custom scraper for the SV Espace restaurant website.
// It loads each weekday by its own dated URL and combines the menus into a single
// HTML document with one labelled section per day.
func ScrapeEspaceWebsite(pageURL string, debug bool) (*MenuData, error) {
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

	// Navigate to website and handle initial setup
	err := chromedp.Run(ctx,
		chromedp.Navigate(pageURL),
		// Do not grant any permissions to avoid the geolocation prompt
		browser.GrantPermissions([]browser.PermissionType{}),
		// Wait for the weekday navigation to load
		chromedp.WaitVisible(`[mat-tab-link]`, chromedp.ByQuery),
		rejectCookies(),
	)
	if err != nil {
		if debug {
			log.Printf("Error during initial page setup: %v", err)
			waitForInterrupt()
		}
		return nil, fmt.Errorf("failed to setup page: %w", err)
	}

	// Each weekday tab links to its own dated URL, so the day never has to be
	// inferred from the tab's position.
	var tabs []menuTab
	if err := chromedp.Run(ctx, chromedp.Evaluate(jsWeekdayTabs, &tabs)); err != nil {
		return nil, fmt.Errorf("failed to read the weekday tabs: %w", err)
	}
	if len(tabs) == 0 {
		return nil, fmt.Errorf("found no dated weekday tabs, the page markup has probably changed")
	}

	base, err := neturl.Parse(pageURL)
	if err != nil {
		return nil, fmt.Errorf("invalid menu URL %q: %w", pageURL, err)
	}

	var allMenus strings.Builder
	days := make([]DayMenu, 0, len(tabs))

	for _, tab := range tabs {
		date, err := time.Parse(time.DateOnly, tab.Date)
		if err != nil {
			return nil, fmt.Errorf("unexpected tab date %q: %w", tab.Date, err)
		}
		day := strings.ToLower(date.Weekday().String())

		href, err := neturl.Parse(tab.Href)
		if err != nil {
			return nil, fmt.Errorf("invalid tab link %q: %w", tab.Href, err)
		}
		dayURL := base.ResolveReference(href).String()

		// Loading each day by its own URL rebuilds the DOM, so we can never capture
		// the previous day's dishes because the page had not re-rendered yet.
		var capture *dayCapture
		err = chromedp.Run(ctx,
			chromedp.Navigate(dayURL),
			chromedp.WaitVisible(`app-category`, chromedp.ByQuery),
			waitForDay(tab.Date),
			chromedp.Evaluate(jsCaptureMenu, &capture),
		)
		if err != nil {
			if debug {
				log.Printf("Error scraping %s menu: %v", day, err)
				continue // Try next day in debug mode
			}
			return nil, fmt.Errorf("failed to scrape the %s menu: %w", day, err)
		}
		if capture == nil {
			return nil, fmt.Errorf("found no menu on the page for %s", day)
		}

		days = append(days, DayMenu{
			Day:    day,
			Date:   tab.Date,
			HTML:   capture.HTML,
			Dishes: capture.Dishes,
		})

		// Kept whole for the debug output; each day is parsed on its own
		fmt.Fprintf(&allMenus, "<h2>%s (%s)</h2>\n%s\n", day, tab.Date, capture.HTML)

		log.Printf("Scraped %s %s: %d dishes (%d bytes)", day, tab.Date, capture.Dishes, len(capture.HTML))
	}

	htmlContent := allMenus.String()

	if debug {
		log.Printf("Successfully scraped all weekly menus (total: %d bytes)", len(htmlContent))
		waitForInterrupt()
	}

	return &MenuData{Content: htmlContent, Days: days}, nil
}

// menuTab is a weekday tab on the SV menu page, e.g. "Fri. 17.07."
type menuTab struct {
	Href string `json:"href"`
	Date string `json:"date"`
}

// dayCapture is one day's menu, with the dish count we check the model against
type dayCapture struct {
	HTML   string `json:"html"`
	Dishes int    `json:"dishes"`
}

// The tab links carry the date they serve, e.g. .../Mittagsmenü/date/2026-07-17
const jsWeekdayTabs = `[...document.querySelectorAll('[mat-tab-link]')]
	.map((tab) => {
		const href = tab.getAttribute('href') || '';
		const date = (href.match(/(\d{4}-\d{2}-\d{2})$/) || [])[1] || '';
		return { href, date };
	})
	.filter((tab) => tab.date)`

// The menu container holds the weekday tab bar as well, which names every day of
// the week. Handing that to the model alongside a "this is monday" heading is the
// same ambiguity that made it drop days elsewhere, so the tab bar is cut out.
const jsCaptureMenu = `(() => {
	const container = document.querySelector('app-menu-container');
	if (!container) return null;

	const clone = container.cloneNode(true);
	clone.querySelectorAll('nav').forEach((nav) => nav.remove());

	const dishes = [...container.querySelectorAll('app-category')].filter((category) => {
		const grid = category.querySelector('app-product-grid');
		const text = grid ? grid.innerText.replace(/\s+/g, ' ').trim() : '';
		// a category with nothing on offer renders its product as "."
		return text !== '' && text !== '.';
	}).length;

	return { html: clone.outerHTML, dishes };
})()`

// rejectCookies declines the cookie banner, tolerating its absence.
//
// chromedp.Click polls for the node until its context runs out, so running it on
// the scrape's own context meant a missing banner burned the entire timeout and
// left the context dead for every step after it. Giving the click its own short
// deadline is what actually makes the banner optional.
func rejectCookies() chromedp.ActionFunc {
	return func(ctx context.Context) error {
		clickCtx, cancel := context.WithTimeout(ctx, cookieClickTimeout)
		defer cancel()

		if err := chromedp.Click(`#cookiescript_reject`, chromedp.ByQuery).Do(clickCtx); err != nil {
			log.Printf("Cookie banner not found or not clickable, continuing: %v", err)
		}
		return nil
	}
}

// waitForDay blocks until the page shows the day we asked for, so a slow render
// can't be mistaken for the menu of the day we requested.
func waitForDay(date string) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		expr := fmt.Sprintf(
			`!!document.querySelector('[mat-tab-link][aria-selected="true"]')?.getAttribute('href')?.endsWith(%q)`,
			date)

		deadline := time.Now().Add(daySwitchTimeout)
		for time.Now().Before(deadline) {
			var showing bool
			if err := chromedp.Evaluate(expr, &showing).Do(ctx); err != nil {
				return err
			}
			if showing {
				return nil
			}
			if err := chromedp.Sleep(250 * time.Millisecond).Do(ctx); err != nil {
				return err
			}
		}

		return fmt.Errorf("page never switched to %s", date)
	}
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
