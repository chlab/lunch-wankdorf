package app

import (
	"strings"
	"testing"

	"github.com/chlab/lunch-wankdorf/pkg/ai"
)

func menuWith(items ...ai.MenuItem) *ai.DailyMenu {
	return &ai.DailyMenu{Type: "daily", Menu: map[string][]ai.MenuItem{"Friday": items}}
}

func TestFillMissingPhotosMatchesOnCategory(t *testing.T) {
	menu := menuWith(
		ai.MenuItem{Name: "Poulet Cordon Bleu", Category: "Postino"},
		ai.MenuItem{Name: "Glasierter Kohl", Category: "Green"},
	)

	result := fillMissingPhotos(menu, map[string]map[string]string{
		"friday": {"Postino": "https://cmnkftzpna.cloudimg.io/postino.jpg"},
	})

	if result.added != 1 {
		t.Errorf("added = %d, want 1", result.added)
	}
	if result.unmatched != 0 {
		t.Errorf("unmatched = %d, want 0", result.unmatched)
	}

	postino := menu.Menu["Friday"][0]
	if postino.Photo == "" || postino.PhotoLarge == "" {
		t.Errorf("Postino has no photo: %+v", postino)
	}
	// The dish the restaurant published no photo for keeps its icon
	if green := menu.Menu["Friday"][1]; green.Photo != "" {
		t.Errorf("Green got a photo it was not given: %q", green.Photo)
	}
}

func TestFillMissingPhotosKeepsPhotosItAlreadyHas(t *testing.T) {
	menu := menuWith(ai.MenuItem{
		Name:     "Poulet Cordon Bleu",
		Category: "Postino",
		Photo:    "https://example.test/yesterday.jpg",
	})

	result := fillMissingPhotos(menu, map[string]map[string]string{
		"friday": {"Postino": "https://cmnkftzpna.cloudimg.io/postino.jpg"},
	})

	if result.added != 0 || result.alreadySet != 1 {
		t.Errorf("got added=%d alreadySet=%d, want 0 and 1", result.added, result.alreadySet)
	}
	if got := menu.Menu["Friday"][0].Photo; got != "https://example.test/yesterday.jpg" {
		t.Errorf("photo = %q, want the published one to be left alone", got)
	}
}

// The categories are the only thing tying a photo to a dish. If they stop lining
// up, the run has to say so rather than look like an uneventful success.
func TestFillMissingPhotosReportsPhotosThatMatchNoDish(t *testing.T) {
	menu := menuWith(ai.MenuItem{Name: "Poulet Cordon Bleu", Category: ""})

	result := fillMissingPhotos(menu, map[string]map[string]string{
		"friday": {"Postino": "https://cmnkftzpna.cloudimg.io/postino.jpg"},
	})

	if result.added != 0 {
		t.Errorf("added = %d, want 0", result.added)
	}
	if result.unmatched != 1 {
		t.Fatalf("unmatched = %d, want 1", result.unmatched)
	}
	if got := strings.Join(result.unmatchedCategories, ","); got != "friday/Postino" {
		t.Errorf("unmatchedCategories = %q, want the day and category that found no dish", got)
	}
}
