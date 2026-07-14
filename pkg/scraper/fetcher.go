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
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/gocolly/colly/v2"
)

const (
	userAgent          = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36"
	httpRequestTimeout = 30 * time.Second
	// Generous: each of Espace's dishes has to be opened to find its link. Must stay
	// under the workflow's step timeout, or the runner kills the scrape before it can
	// report why it gave up.
	chromeTimeout = 4 * time.Minute
	// One day must not be able to spend the whole scrape's budget: without a deadline
	// of its own, a day whose menu never renders leaves nothing for the days after it.
	dayScrapeTimeout   = 60 * time.Second
	cookieClickTimeout = 5 * time.Second
	daySwitchTimeout   = 15 * time.Second
	// Short: a page that already failed to render must not hold up the diagnosis
	pageDescribeTimeout = 10 * time.Second
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
	return scrapeEspace(pageURL, "", debug)
}

// ScrapeEspaceDay scrapes a single weekday, given as a YYYY-MM-DD date. The daily
// photo run only needs the photos of the day being published, and asking for just
// that day keeps it off the other days' tabs, which it has no use for.
//
// A date the page has no tab for - a holiday, a day outside the published week -
// yields no days rather than an error: the caller treats an empty scrape as "nothing
// published yet", which is exactly what it is.
func ScrapeEspaceDay(pageURL, date string, debug bool) (*MenuData, error) {
	return scrapeEspace(pageURL, date, debug)
}

// scrapeEspace scrapes the whole published week, or a single day when onlyDate is
// set to a YYYY-MM-DD date.
func scrapeEspace(pageURL, onlyDate string, debug bool) (*MenuData, error) {
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
		if onlyDate != "" && tab.Date != onlyDate {
			continue
		}

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
		//
		// The day gets a deadline of its own: the waits below poll until their context
		// runs out, so on the scrape's own context a day that never renders would burn
		// the entire budget and leave nothing for the days behind it. Cancelling only
		// the day's context leaves the browser alive for them.
		dayCtx, cancelDay := context.WithTimeout(ctx, dayScrapeTimeout)

		var capture *dayCapture
		err = chromedp.Run(dayCtx,
			chromedp.Navigate(dayURL),
			chromedp.WaitVisible(`app-category`, chromedp.ByQuery),
			waitForDay(tab.Date),
			chromedp.Evaluate(jsCaptureMenu, &capture),
		)
		if err != nil {
			cancelDay()
			log.Printf("Error scraping %s menu: %v\nThe page was showing: %s",
				day, err, describePage(ctx))
			if debug {
				continue // Try next day in debug mode
			}
			return nil, fmt.Errorf("failed to scrape the %s menu: %w", day, err)
		}
		if capture == nil {
			cancelDay()
			return nil, fmt.Errorf("found no menu on the page for %s", day)
		}

		// Opening every dish to find its link is slow and clicks a lot of buttons, so
		// losing them is not worth failing a menu over - the dish just gets no link.
		links := make(map[string]string)
		if err := chromedp.Run(dayCtx, evaluateAsync(jsDishLinks, &links)); err != nil {
			log.Printf("Warning: could not read the %s dish links: %v", day, err)
		}
		cancelDay()

		days = append(days, DayMenu{
			Day:    day,
			Date:   tab.Date,
			HTML:   capture.HTML,
			Dishes: capture.Dishes,
			Photos: capture.Photos,
			Links:  links,
			URL:    dayURL,
		})

		// Kept whole for the debug output; each day is parsed on its own
		fmt.Fprintf(&allMenus, "<h2>%s (%s)</h2>\n%s\n", day, tab.Date, capture.HTML)

		log.Printf("Scraped %s %s: %d dishes, %d photos, %d links (%d bytes)",
			day, tab.Date, capture.Dishes, len(capture.Photos), len(links), len(capture.HTML))
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
// and the dish photos, keyed by the category heading they sit under.
type dayCapture struct {
	HTML   string            `json:"html"`
	Dishes int               `json:"dishes"`
	Photos map[string]string `json:"photos"`
}

// The tab links carry the date they serve, e.g. .../Mittagsmenü/date/2026-07-17
// jsDescribePage is what the page has to say for itself when the menu never showed up.
const jsDescribePage = `(() => {
	const selected = document.querySelector('[mat-tab-link][aria-selected="true"]');
	return {
		url: location.href,
		title: document.title,
		readyState: document.readyState,
		tabs: document.querySelectorAll('[mat-tab-link]').length,
		categories: document.querySelectorAll('app-category').length,
		selected: selected ? selected.getAttribute('href') : '',
		text: (document.body ? document.body.innerText : '').replace(/\s+/g, ' ').trim().slice(0, 400),
	};
})()`

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
//
// The dish photos are CSS background images rather than <img> tags, and a day
// whose photos are not published yet serves a placeholder for every dish, which we
// drop. Photos are keyed by the category heading, the one thing that ties a photo
// to a dish once the model has rewritten the dish's name.
const jsCaptureMenu = `(() => {
	const container = document.querySelector('app-menu-container');
	if (!container) return null;

	const clone = container.cloneNode(true);
	clone.querySelectorAll('nav').forEach((nav) => nav.remove());

	const isOnOffer = (category) => {
		const grid = category.querySelector('app-product-grid');
		const text = grid ? grid.innerText.replace(/\s+/g, ' ').trim() : '';
		// a category with nothing on offer renders its product as "."
		return text !== '' && text !== '.';
	};

	const photoOf = (category) => {
		for (const element of category.querySelectorAll('*')) {
			const background = getComputedStyle(element).backgroundImage;
			if (!background || !background.includes('http')) continue;

			const url = background.replace(/^url\(["']?|["']?\)$/g, '');
			// the photos for a day are only published on the morning of that day
			if (url.includes('` + espaceFallbackPhoto + `')) return '';
			return url;
		}
		return '';
	};

	const onOffer = [...container.querySelectorAll('app-category')].filter(isOnOffer);

	const photos = {};
	onOffer.forEach((category) => {
		const heading = category.innerText.split('\n')[0].trim();
		const photo = photoOf(category);
		if (heading && photo) photos[heading] = photo;
	});

	return { html: clone.outerHTML, dishes: onOffer.length, photos };
})()`

// The dish cards are not links: they are cards with a click handler, and the dish's
// URL only exists once the app has routed to it. So we open each dish, note where
// we ended up, and come back. Best effort - a dish we can't open simply gets no
// link, and keeps the rest of the menu.
const jsDishLinks = `(async () => {
	const dayURL = location.href;
	const links = {};

	const categories = [...document.querySelectorAll('app-category')];
	for (let i = 0; i < categories.length; i++) {
		// The click routes away, so the elements have to be looked up again each time
		const category = [...document.querySelectorAll('app-category')][i];
		if (!category) continue;

		const heading = category.innerText.split('\n')[0].trim();
		const grid = category.querySelector('app-product-grid');
		const dish = grid ? grid.innerText.replace(/\s+/g, ' ').trim() : '';
		// nothing on offer, so nothing to link to
		if (!heading || dish === '' || dish === '.') continue;

		const card = grid.querySelector('mat-card');
		if (!card) continue;

		card.click();
		await new Promise((resolve) => setTimeout(resolve, 800));

		if (location.href !== dayURL) {
			links[heading] = location.href;
			history.back();
			await new Promise((resolve) => setTimeout(resolve, 800));
		}
	}

	return links;
})()`

// evaluateAsync runs an expression that returns a promise and waits for it, which
// chromedp.Evaluate does not do on its own.
func evaluateAsync(expression string, result any) chromedp.Action {
	return chromedp.Evaluate(expression, result, func(p *runtime.EvaluateParams) *runtime.EvaluateParams {
		return p.WithAwaitPromise(true)
	})
}

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

// describePage reports what the browser is actually showing, so a scrape that timed
// out waiting for the menu says whether it was looking at a consent wall, an error
// page or a day other than the one it asked for - none of which the timeout itself
// can tell apart. It runs on the scrape's context rather than the day's, which is
// already spent by the time anything wants a diagnosis.
func describePage(ctx context.Context) string {
	ctx, cancel := context.WithTimeout(ctx, pageDescribeTimeout)
	defer cancel()

	var page struct {
		URL        string `json:"url"`
		Title      string `json:"title"`
		ReadyState string `json:"readyState"`
		Tabs       int    `json:"tabs"`
		Categories int    `json:"categories"`
		Selected   string `json:"selected"`
		Text       string `json:"text"`
	}
	if err := chromedp.Run(ctx, chromedp.Evaluate(jsDescribePage, &page)); err != nil {
		return fmt.Sprintf("(could not be read: %v)", err)
	}

	return fmt.Sprintf(
		"url=%q title=%q readyState=%s tabs=%d categories=%d selectedTab=%q\ntext: %s",
		page.URL, page.Title, page.ReadyState, page.Tabs, page.Categories, page.Selected, page.Text)
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
