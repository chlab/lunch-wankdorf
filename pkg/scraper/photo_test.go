package scraper

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPhotoURLsResizesEspacePhotos(t *testing.T) {
	// The page's own URL is signed for the size it wants; we drop the query and ask
	// for our own sizes unsigned.
	photo := "https://cmnkftzpna.cloudimg.io/https%3A%2F%2Fqnips.blob.core.windows.net%2Freleaseproductpics%2F23237897.jpg" +
		"?ci_url_encoded=1&w=420&h=190&func=boundmin&ci_sign=4d9ad04c"

	thumb, large := PhotoURLs(photo)

	if strings.Contains(thumb, "ci_sign") || strings.Contains(large, "ci_sign") {
		t.Errorf("kept the page's signature, which only covers its own size:\n%s\n%s", thumb, large)
	}
	if !strings.Contains(thumb, "w=96") {
		t.Errorf("thumb = %q, want a 96px wide image", thumb)
	}
	if !strings.Contains(large, "w=900") {
		t.Errorf("large = %q, want a 900px wide image", large)
	}
	// Without this cloudimg does not understand the percent-encoded origin
	if !strings.Contains(thumb, "ci_url_encoded=1") {
		t.Errorf("thumb = %q, want ci_url_encoded=1", thumb)
	}
}

func TestPhotoURLsResizesFood2050Photos(t *testing.T) {
	// These sit on plain storage with no resizing, so they go through the site's
	// own image optimizer
	photo := "https://storage.googleapis.com/dish-images-prod/dish-media%2Fzfv%2F86f3d0e8%2Fpizza.jpg"

	thumb, large := PhotoURLs(photo)

	for _, url := range []string{thumb, large} {
		if !strings.HasPrefix(url, "https://app.food2050.ch/_next/image?url=") {
			t.Errorf("url = %q, want it to go through the image optimizer", url)
		}
		// The optimizer takes the photo as a query parameter, so it has to be escaped
		if strings.Contains(url, "url=https://") {
			t.Errorf("url = %q, want the photo URL to be escaped", url)
		}
	}
	if !strings.Contains(thumb, "w=128") {
		t.Errorf("thumb = %q, want a small image", thumb)
	}
	if !strings.Contains(large, "w=1080") {
		t.Errorf("large = %q, want a large image", large)
	}
}

func TestPhotoURLsWithoutAPhoto(t *testing.T) {
	thumb, large := PhotoURLs("")
	if thumb != "" || large != "" {
		t.Errorf("got (%q, %q), want two empty strings", thumb, large)
	}
}

const dishPage = `<html><body>
<script id="__NEXT_DATA__" type="application/json">
{"props":{"pageProps":{"organisation":{"outlet":{"menuCategory":{"menuItem":{"dish":
{"name":"Pizza","imageUrl":"https://storage.googleapis.com/dish-images-prod/pizza.jpg"}}}}}}}}
</script>
</body></html>`

// A dish that has no photo yet: the field is there but empty
const dishPageWithoutPhoto = `<html><body>
<script id="__NEXT_DATA__" type="application/json">
{"props":{"pageProps":{"organisation":{"outlet":{"menuCategory":{"menuItem":{"dish":
{"name":"Pasta","imageUrl":""}}}}}}}}
</script>
</body></html>`

func TestFetchDishPhoto(t *testing.T) {
	tests := []struct {
		name string
		page string
		want string
	}{
		{"dish with a photo", dishPage, "https://storage.googleapis.com/dish-images-prod/pizza.jpg"},
		{"dish without a photo", dishPageWithoutPhoto, ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Write([]byte(test.page))
			}))
			defer server.Close()

			photo, err := FetchDishPhoto(server.Client(), server.URL)
			if err != nil {
				t.Fatalf("FetchDishPhoto() error = %v", err)
			}
			if photo != test.want {
				t.Errorf("photo = %q, want %q", photo, test.want)
			}
		})
	}
}

func TestFetchDishPhotoOnAPageWithoutData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("<html><body>no next data here</body></html>"))
	}))
	defer server.Close()

	// The caller logs this and carries on with the icon, rather than failing the menu
	if _, err := FetchDishPhoto(server.Client(), server.URL); err == nil {
		t.Error("want an error when the page carries no embedded data")
	}
}
