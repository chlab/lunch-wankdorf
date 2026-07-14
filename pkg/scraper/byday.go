package scraper

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Dish links end in the date the dish is served on, e.g.
// .../mittagsverpflegung,hauptspeisen,pizza-del-giorno/2026-07-17
var reDishDate = regexp.MustCompile(`/(\d{4}-\d{2}-\d{2})$`)

// GroupMenuByDay restructures a food2050 weekly menu page (Gira, Luna, Sole) into
// explicit per-day sections.
//
// The page is a transposed grid: dishes are grouped by category ("Pasta Del
// Giorno") with one link per weekday, and the only thing tying a dish to a day is
// the date at the end of its link. Asking the model to invert that mapping is what
// made it silently drop the last days of the week. Grouping the dishes here means
// the model only has to read dishes out of a labelled section.
//
// The returned counts are per lowercase weekday and let the caller check that the
// model brought back everything the page offered.
func GroupMenuByDay(htmlContent string) (string, map[string]int, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse menu HTML: %w", err)
	}

	type dish struct {
		category    string
		description string
		link        string
	}

	dishesByDate := make(map[string][]dish)
	// The current day is rendered twice (once in the weekly grid, once in the
	// single-day view below it), so the link doubles as a de-duplication key.
	seen := make(map[string]bool)

	doc.Find("a[href]").Each(func(_ int, link *goquery.Selection) {
		href, _ := link.Attr("href")
		match := reDishDate.FindStringSubmatch(href)
		if match == nil || seen[href] {
			return
		}

		description := normalizeSpace(link.Text())
		if description == "" {
			return
		}

		seen[href] = true
		date := match[1]
		dishesByDate[date] = append(dishesByDate[date], dish{
			category:    categoryOf(link),
			description: description,
			link:        href,
		})
	})

	dates := make([]string, 0, len(dishesByDate))
	for date := range dishesByDate {
		dates = append(dates, date)
	}
	sort.Strings(dates)

	var grouped strings.Builder
	counts := make(map[string]int, len(dates))

	for _, date := range dates {
		parsed, err := time.Parse(time.DateOnly, date)
		if err != nil {
			return "", nil, fmt.Errorf("unexpected dish date %q: %w", date, err)
		}

		day := strings.ToLower(parsed.Weekday().String())
		counts[day] = len(dishesByDate[date])

		fmt.Fprintf(&grouped, "<h2>%s (%s)</h2>\n", day, date)
		for _, d := range dishesByDate[date] {
			fmt.Fprintf(&grouped,
				"<div><h3>%s</h3><p>%s</p><a href=\"%s\">Details</a></div>\n",
				d.category, d.description, d.link)
		}
	}

	return grouped.String(), counts, nil
}

// categoryOf finds the dish's category heading ("Pizza Del Giorno"). In the weekly
// grid the heading is a cell of the row the dish sits in; in the single-day view it
// is a sibling of the dish link.
func categoryOf(link *goquery.Selection) string {
	if heading := link.Parent().ChildrenFiltered("p").First(); heading.Length() > 0 {
		return normalizeSpace(heading.Text())
	}

	var category string
	link.Parent().Parent().Children().EachWithBreak(func(_ int, cell *goquery.Selection) bool {
		// Cells holding a link are dishes, the one without is the row's heading
		if cell.Find("a").Length() > 0 {
			return true
		}
		if heading := cell.Find("p").First(); heading.Length() > 0 {
			category = normalizeSpace(heading.Text())
			return false
		}
		return true
	})

	return category
}

func normalizeSpace(text string) string {
	return strings.Join(strings.Fields(text), " ")
}
