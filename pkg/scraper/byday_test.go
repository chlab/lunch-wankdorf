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
	grouped, counts, err := GroupMenuByDay(weeklyGrid)
	if err != nil {
		t.Fatalf("GroupMenuByDay() error = %v", err)
	}

	// Monday's dishes are repeated by the single-day view but must only count once
	if got, want := counts["monday"], 2; got != want {
		t.Errorf("counts[monday] = %d, want %d", got, want)
	}
	if got, want := counts["friday"], 2; got != want {
		t.Errorf("counts[friday] = %d, want %d", got, want)
	}
	if _, ok := counts["saturday"]; ok {
		t.Errorf("counts has saturday, want no entry for a day without dishes")
	}

	// Each dish has to end up under the day its link points at
	mondaySection := section(grouped, "<h2>monday")
	if !strings.Contains(mondaySection, "PASTA SALSICCA") || !strings.Contains(mondaySection, "PIZZA FUNGHI") {
		t.Errorf("monday section is missing dishes:\n%s", mondaySection)
	}
	if strings.Contains(mondaySection, "PASTA PESTO") {
		t.Errorf("monday section contains a friday dish:\n%s", mondaySection)
	}

	fridaySection := section(grouped, "<h2>friday")
	if !strings.Contains(fridaySection, "PASTA PESTO") || !strings.Contains(fridaySection, "PIZZA SICILIANA") {
		t.Errorf("friday section is missing dishes:\n%s", fridaySection)
	}

	// The category is the dish name the page shows above the description
	if !strings.Contains(mondaySection, "<h3>Pasta Del Giorno</h3>") {
		t.Errorf("monday section is missing the dish category:\n%s", mondaySection)
	}

	// Links without a date (navigation, footer) are not dishes
	if strings.Contains(grouped, "Newsletter") {
		t.Errorf("grouped content contains a non-dish link:\n%s", grouped)
	}
}

func TestGroupMenuByDayWithoutDatedLinks(t *testing.T) {
	grouped, counts, err := GroupMenuByDay(`<div><a href="https://x/newsletter">Newsletter</a></div>`)
	if err != nil {
		t.Fatalf("GroupMenuByDay() error = %v", err)
	}

	// The caller relies on this to fall back to the ungrouped content
	if len(counts) != 0 {
		t.Errorf("counts = %v, want empty when the page has no dated dish links", counts)
	}
	if grouped != "" {
		t.Errorf("grouped = %q, want empty", grouped)
	}
}

// section returns the part of grouped starting at heading, up to the next heading.
func section(grouped, heading string) string {
	start := strings.Index(grouped, heading)
	if start == -1 {
		return ""
	}
	rest := grouped[start+len(heading):]
	if end := strings.Index(rest, "<h2>"); end != -1 {
		return rest[:end]
	}
	return rest
}
