package scraper

import (
	"strings"
	"testing"
)

// Mirrors the food2050 layout: a category row ("Pasta Del Giorno") with one dated
// link per weekday, followed by the single-day view that repeats the current day.
const weeklyGrid = `
<div>
  <div>
    <div><p>Pasta Del Giorno</p></div>
    <div><a href="https://x/menu,pasta-del-giorno/2026-07-13"><div><p>PASTA SALSICCA, Tomatensauce</p></div></a></div>
    <div><a href="https://x/menu,pasta-del-giorno/2026-07-17"><div><p>PASTA PESTO, Basilikum</p></div></a></div>
  </div>
  <div>
    <div><p>Pizza Del Giorno</p></div>
    <div><a href="https://x/menu,pizza-del-giorno/2026-07-13"><div><p>PIZZA FUNGHI, Champignons</p></div></a></div>
    <div><a href="https://x/menu,pizza-del-giorno/2026-07-17"><div><p>PIZZA SICILIANA, Kapern</p></div></a></div>
  </div>
</div>
<div>
  <div><p>Montag, 13.07.2026</p></div>
  <div><p>Pasta Del Giorno</p><a href="https://x/menu,pasta-del-giorno/2026-07-13"><div><p>PASTA SALSICCA, Tomatensauce</p></div></a></div>
</div>
<div><a href="https://x/newsletter">Newsletter</a></div>
`

func TestGroupMenuByDayAssignsDishesToTheDayInTheirLink(t *testing.T) {
	days, err := GroupMenuByDay(weeklyGrid)
	if err != nil {
		t.Fatalf("GroupMenuByDay() error = %v", err)
	}

	// Only the days that actually have dishes, in order
	if got, want := len(days), 2; got != want {
		t.Fatalf("got %d days, want %d: %+v", got, want, days)
	}

	monday, friday := days[0], days[1]
	if monday.Day != "monday" || friday.Day != "friday" {
		t.Fatalf("days are %q and %q, want monday and friday", monday.Day, friday.Day)
	}

	// Monday's dishes are repeated by the single-day view but must only count once
	if got, want := monday.Dishes, 2; got != want {
		t.Errorf("monday.Dishes = %d, want %d (the repeated dish must be de-duplicated)", got, want)
	}
	if got, want := friday.Dishes, 2; got != want {
		t.Errorf("friday.Dishes = %d, want %d", got, want)
	}

	// Each dish has to end up under the day its link points at
	if !strings.Contains(monday.HTML, "PASTA SALSICCA") || !strings.Contains(monday.HTML, "PIZZA FUNGHI") {
		t.Errorf("monday is missing dishes:\n%s", monday.HTML)
	}
	if strings.Contains(monday.HTML, "PASTA PESTO") {
		t.Errorf("monday contains a friday dish:\n%s", monday.HTML)
	}
	if !strings.Contains(friday.HTML, "PASTA PESTO") || !strings.Contains(friday.HTML, "PIZZA SICILIANA") {
		t.Errorf("friday is missing dishes:\n%s", friday.HTML)
	}

	// The category is the dish name the page shows above the description
	if !strings.Contains(monday.HTML, "<h3>Pasta Del Giorno</h3>") {
		t.Errorf("monday is missing the dish category:\n%s", monday.HTML)
	}

	// Links without a date (navigation, footer) are not dishes
	if strings.Contains(monday.HTML+friday.HTML, "Newsletter") {
		t.Error("grouped content contains a non-dish link")
	}
}

func TestGroupMenuByDayWithoutDatedLinks(t *testing.T) {
	days, err := GroupMenuByDay(`<div><a href="https://x/newsletter">Newsletter</a></div>`)
	if err != nil {
		t.Fatalf("GroupMenuByDay() error = %v", err)
	}

	// The caller turns this into a loud failure rather than parsing a menu it can't
	// attribute to days
	if len(days) != 0 {
		t.Errorf("got %+v, want no days when the page has no dated dish links", days)
	}
}
