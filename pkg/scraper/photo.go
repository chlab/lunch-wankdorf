package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"regexp"
	"strings"
)

// The restaurants serve their dish photos far larger than we need (Espace's are
// 3000px, several megabytes). Both origins can resize on the fly, so we hand the
// frontend a thumbnail to sit where the icon goes and a bigger one for the
// lightbox, rather than shipping the originals.
const (
	// The thumbnail is shown at 64px, asked for at twice that so it stays sharp on
	// the phone screens this is mostly read on
	thumbWidth = 128
	largeWidth = 900
)

// The photo sits in the page's embedded Next.js state, at
// props.pageProps.organisation.outlet.menuCategory.menuItem.dish.imageUrl
var reNextData = regexp.MustCompile(`(?s)<script id="__NEXT_DATA__" type="application/json">(.*?)</script>`)

// Espace serves this placeholder for days whose photos aren't published yet
const espaceFallbackPhoto = "product-fallback.jpg"

type nextData struct {
	Props struct {
		PageProps struct {
			Organisation struct {
				Outlet struct {
					MenuCategory struct {
						MenuItem struct {
							Dish struct {
								ImageURL string `json:"imageUrl"`
							} `json:"dish"`
						} `json:"menuItem"`
					} `json:"menuCategory"`
				} `json:"outlet"`
			} `json:"organisation"`
		} `json:"pageProps"`
	} `json:"props"`
}

// FetchDishPhoto returns the photo for a food2050 dish, given the link the menu
// item already carries. Not every dish has one; an empty string means no photo,
// which is not an error.
func FetchDishPhoto(client *http.Client, dishURL string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, dishURL, nil)
	if err != nil {
		return "", fmt.Errorf("invalid dish URL %q: %w", dishURL, err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch %s: %w", dishURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetching %s: %s", dishURL, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read %s: %w", dishURL, err)
	}

	match := reNextData.FindSubmatch(body)
	if match == nil {
		return "", fmt.Errorf("no embedded page data found on %s", dishURL)
	}

	var data nextData
	if err := json.Unmarshal(match[1], &data); err != nil {
		return "", fmt.Errorf("failed to parse the page data of %s: %w", dishURL, err)
	}

	return data.Props.PageProps.Organisation.Outlet.MenuCategory.MenuItem.Dish.ImageURL, nil
}

// PhotoURLs turns a photo URL from either restaurant into the thumbnail and the
// larger version the lightbox shows. Both origins resize on request, so no image
// is downloaded or stored by us. An empty photo yields two empty strings.
func PhotoURLs(photo string) (thumb string, large string) {
	if photo == "" {
		return "", ""
	}

	switch {
	// Espace (SV) serves its photos through cloudimg, which takes the size as
	// query parameters. The page's own URL is signed for its size, so we drop the
	// query and ask unsigned for the sizes we want.
	case strings.Contains(photo, "cloudimg.io"):
		base, _, _ := strings.Cut(photo, "?")
		return fmt.Sprintf("%s?ci_url_encoded=1&w=%d&h=%d&func=crop", base, thumbWidth, thumbWidth),
			fmt.Sprintf("%s?ci_url_encoded=1&w=%d", base, largeWidth)

	// food2050's photos sit on plain storage with no resizing, but the site's own
	// image optimizer will serve any width from them, in webp.
	default:
		optimize := func(width int) string {
			return fmt.Sprintf("https://app.food2050.ch/_next/image?url=%s&w=%d&q=75",
				neturl.QueryEscape(photo), width)
		}
		// The optimizer only serves widths it has been configured for
		return optimize(128), optimize(1080)
	}
}
