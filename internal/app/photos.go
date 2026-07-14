package app

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/chlab/lunch-wankdorf/pkg/ai"
	"github.com/chlab/lunch-wankdorf/pkg/scraper"
)

// Photos are a nice-to-have: a dish without one falls back to its icon, so a
// restaurant that has no photos, or a fetch that fails, must never fail the run.
const (
	photoFetchWorkers = 8
	photoFetchTimeout = 20 * time.Second
)

// addPhotos fills in the dish photos for a parsed menu and reports how many it
// found. Where a photo comes from depends on the restaurant: Espace puts them on
// the menu page (keyed by category), food2050 only on the dish pages we already
// link to.
func addPhotos(menu *ai.DailyMenu, days []scraper.DayMenu) int {
	photosByDay := make(map[string]map[string]string, len(days))
	for _, day := range days {
		photosByDay[day.Day] = day.Photos
	}

	fromPage := fillMissingPhotos(menu, photosByDay)

	// Whatever the page didn't give us, look up on the dish's own page
	var toFetch []*ai.MenuItem
	for day, items := range menu.Menu {
		for i := range items {
			item := &menu.Menu[day][i]
			if item.Photo == "" && item.Link != "" {
				toFetch = append(toFetch, item)
			}
		}
	}

	found := fromPage.added + fetchDishPhotos(toFetch)
	log.Printf("Photos: %d of %d dishes have one", found, countItems(menu))

	return found
}

// fetchDishPhotos looks up each dish's photo on its own page, in parallel.
func fetchDishPhotos(items []*ai.MenuItem) int {
	if len(items) == 0 {
		return 0
	}

	client := &http.Client{Timeout: photoFetchTimeout}
	queue := make(chan *ai.MenuItem)

	var mu sync.Mutex
	var found int

	var wg sync.WaitGroup
	for range photoFetchWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range queue {
				photo, err := scraper.FetchDishPhoto(client, item.Link)
				if err != nil {
					// A missing photo is not worth failing the menu over
					log.Printf("Warning: could not fetch the photo for %q: %v", item.Name, err)
					continue
				}
				if photo == "" {
					continue
				}

				thumb, large := scraper.PhotoURLs(photo)

				mu.Lock()
				item.Photo, item.PhotoLarge = thumb, large
				found++
				mu.Unlock()
			}
		}()
	}

	for _, item := range items {
		queue <- item
	}
	close(queue)
	wg.Wait()

	return found
}

// addDishLinks gives the dishes the link to their page on the restaurant's site,
// for scrapers that had to work it out themselves. A dish whose page we could not
// reach links to the day's menu instead, which is better than no link at all.
func addDishLinks(menu *ai.DailyMenu, days []scraper.DayMenu) {
	for _, day := range days {
		if len(day.Links) == 0 && day.URL == "" {
			continue
		}

		items := menu.Menu[capitalize(day.Day)]
		for i := range items {
			item := &items[i]
			// food2050 puts the link in the markup, so the model already has it
			if item.Link != "" {
				continue
			}

			if link := day.Links[item.Category]; link != "" {
				item.Link = link
			} else {
				item.Link = day.URL
			}
		}
	}
}

// photoMerge is what filling in the photos did, so the caller can tell an
// uneventful run ("no new photos yet") from a broken one ("photos found, but they
// belong to no dish on the menu").
type photoMerge struct {
	added               int
	alreadySet          int
	unmatched           int
	unmatchedCategories []string
}

// fillMissingPhotos adds the photos to the menu items that have none, matching them
// on the category the dish is listed under. Items that already carry a photo are
// left alone, so a menu never loses a photo it once had.
func fillMissingPhotos(menu *ai.DailyMenu, photosByDay map[string]map[string]string) photoMerge {
	var result photoMerge

	for day, photos := range photosByDay {
		items := menu.Menu[capitalize(day)]

		for category, photo := range photos {
			var matched bool

			for i := range items {
				item := &items[i]
				if item.Category != category {
					continue
				}
				matched = true

				if item.Photo != "" {
					result.alreadySet++
					continue
				}
				item.Photo, item.PhotoLarge = scraper.PhotoURLs(photo)
				result.added++
			}

			if !matched {
				result.unmatched++
				result.unmatchedCategories = append(result.unmatchedCategories, day+"/"+category)
			}
		}
	}

	sort.Strings(result.unmatchedCategories)
	return result
}

func countItems(menu *ai.DailyMenu) int {
	var count int
	for _, items := range menu.Menu {
		count += len(items)
	}
	return count
}

// RunPhotoUpdate adds the dish photos that have been published since the menu was
// parsed, without touching the menu itself.
//
// Espace only publishes a day's photos on the morning of that day, so on Monday
// the rest of the week has none. Rather than re-parse the whole week daily - which
// would pay a model to rewrite text we already have, and risk regressing it - this
// scrapes the photos, fills in the blanks in the published menu and puts it back.
func RunPhotoUpdate(config Config) error {
	loadEnv()

	restaurant, exists := restaurantMenus[config.RestaurantID]
	if !exists {
		return fmt.Errorf("restaurant with ID '%s' not found", config.RestaurantID)
	}
	if !restaurant.HasCustomScraper {
		return fmt.Errorf("%s publishes its photos on the dish pages, which the weekly run already fetches", restaurant.Name)
	}

	// Only today's photos can be new: Espace publishes a day's photos on the morning
	// of that day, and every earlier day of the week was picked up by the run on its
	// own morning. Asking for the whole week would re-scrape days we already have.
	today := time.Now().Format(time.DateOnly)

	log.Printf("Looking for new %s photos for %s", restaurant.Name, today)

	scraped, err := scraper.ScrapeEspaceDay(restaurant.URL, today, config.DebugMode)
	if err != nil {
		return fmt.Errorf("error scraping menu data: %w", err)
	}

	photosByDay := make(map[string]map[string]string, len(scraped.Days))
	var scrapedPhotos int
	for _, day := range scraped.Days {
		photosByDay[day.Day] = day.Photos
		scrapedPhotos += len(day.Photos)
	}
	if scrapedPhotos == 0 {
		log.Println("No photos published yet, nothing to do")
		return nil
	}

	bucket, err := openMenuBucket()
	if err != nil {
		return err
	}

	menuJSON, err := bucket.get(restaurant.Name)
	if err != nil {
		return err
	}

	var menu ai.DailyMenu
	if err := json.Unmarshal(menuJSON, &menu); err != nil {
		return fmt.Errorf("failed to parse the published menu: %w", err)
	}

	result := fillMissingPhotos(&menu, photosByDay)

	// A photo the menu has no home for means the categories no longer line up - the
	// one thing that ties a photo to a dish. Say so, rather than reporting "nothing
	// to do" and looking like a quiet success.
	if result.unmatched > 0 {
		log.Printf("Warning: %d photos matched no menu item (categories: %s). "+
			"Has the menu been parsed since categories were added?",
			result.unmatched, strings.Join(result.unmatchedCategories, ", "))
	}

	if result.added == 0 {
		log.Printf("Found %d photos, %d already published, nothing to add",
			scrapedPhotos, result.alreadySet)
		return nil
	}
	added := result.added

	updated, err := json.MarshalIndent(&menu, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal menu: %w", err)
	}

	if !config.UploadToR2 {
		fmt.Println(string(updated))
		log.Printf("Would add %d photos (pass -upload to publish them)", added)
		return nil
	}

	if err := bucket.put(restaurant.Name, updated); err != nil {
		return err
	}

	log.Printf("Added %d photos to the %s menu", added, restaurant.Name)
	return nil
}
