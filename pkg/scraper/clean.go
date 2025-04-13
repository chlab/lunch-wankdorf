package scraper

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
)

func OptimizeHTML(html string) string {
	html = minimizeHTML(html)
	html = cleanHTML(html)
	html = stripTags(html)
	html = removeKlimawirkung(html)
	return html
}

func minimizeHTML(htmlContent string) string {
	// Initialize minifier
	m := minify.New()
	m.AddFunc("text/html", html.Minify)
	m.Add("text/html", &html.Minifier{
		KeepQuotes: true,
	})

	// Create input and output buffers
	input := bytes.NewBufferString(htmlContent)
	output := &bytes.Buffer{}

	// Minify the content
	var minifiedContent string
	err := m.Minify("text/html", output, input)
	if err != nil {
		fmt.Printf("Error minifying HTML: %v, using original content\n", err)
		minifiedContent = htmlContent
	} else {
		minified := output.String()
		fmt.Printf("Minified HTML from %d to %d bytes (%.1f%%)\n",
			len(htmlContent),
			len(minified),
			float64(len(minified))/float64(len(htmlContent))*100)
		minifiedContent = minified
	}
	return minifiedContent
}

// CleanHTML removes unnecessary parts of HTML like images, SVGs, comments, etc.
func cleanHTML(html string) string {
	// Remove HTML comments
	html = regexp.MustCompile(`<!--[\s\S]*?-->`).ReplaceAllString(html, "")

	// Remove CSS
	html = regexp.MustCompile(`<style[^>]*>[\s\S]*?</style>`).ReplaceAllString(html, "")

	// Remove JS
	html = regexp.MustCompile(`<script[^>]*>[\s\S]*?</script>`).ReplaceAllString(html, "")

	// Remove SVG elements
	html = regexp.MustCompile(`<svg[^>]*>[\s\S]*?</svg>`).ReplaceAllString(html, "")

	// Remove img tags
	html = regexp.MustCompile(`<img[^>]*>`).ReplaceAllString(html, "")

	// Remove iframe tags
	html = regexp.MustCompile(`<iframe[^>]*>[\s\S]*?</iframe>`).ReplaceAllString(html, "")

	// Remove video tags
	html = regexp.MustCompile(`<video[^>]*>[\s\S]*?</video>`).ReplaceAllString(html, "")

	// Remove audio tags
	html = regexp.MustCompile(`<audio[^>]*>[\s\S]*?</audio>`).ReplaceAllString(html, "")

	// Remove canvas tags
	html = regexp.MustCompile(`<canvas[^>]*>[\s\S]*?</canvas>`).ReplaceAllString(html, "")

	// Remove path tags
	html = regexp.MustCompile(`<path[^>]*>[\s\S]*?</path>`).ReplaceAllString(html, "")

	// Remove object tags
	html = regexp.MustCompile(`<object[^>]*>[\s\S]*?</object>`).ReplaceAllString(html, "")

	// Remove hidden elements
	html = regexp.MustCompile(`<[^>]* hidden[^>]*>[\s\S]*?</[^>]*>`).ReplaceAllString(html, "")
	html = regexp.MustCompile(`<[^>]* style="[^"]*display:\s*none[^"]*"[^>]*>[\s\S]*?</[^>]*>`).ReplaceAllString(html, "")
	html = regexp.MustCompile(`<[^>]* style="[^"]*visibility:\s*hidden[^"]*"[^>]*>[\s\S]*?</[^>]*>`).ReplaceAllString(html, "")

	// Replace common entities
	html = strings.ReplaceAll(html, "&nbsp;", " ")
	html = strings.ReplaceAll(html, "&amp;", "&")
	html = strings.ReplaceAll(html, "&lt;", "<")
	html = strings.ReplaceAll(html, "&gt;", ">")

	return html
}

// stripTags removes class and style attributes from HTML tags
func stripTags(html string) string {
	// Remove class attributes
	html = regexp.MustCompile(` class="[^"]*"`).ReplaceAllString(html, "")
	html = regexp.MustCompile(` class='[^']*'`).ReplaceAllString(html, "")

	// Remove style attributes
	html = regexp.MustCompile(` style="[^"]*"`).ReplaceAllString(html, "")
	html = regexp.MustCompile(` style='[^']*'`).ReplaceAllString(html, "")

	// Remove target attributes
	html = regexp.MustCompile(` target="[^"]*"`).ReplaceAllString(html, "")
	html = regexp.MustCompile(` target='[^']*'`).ReplaceAllString(html, "")

	return html
}

// all the SBB restaurants have a lot of redundant markdown, signaled by the string "Klimawirkung"
// if we find it, we remove everything after it, otherwise we return the original text
func removeKlimawirkung(text string) string {
	klimaIndex := strings.Index(text, "Klimawirkung")
	if klimaIndex == -1 {
		// "Klimawirkung" not found, return the original text
		return text
	}

	// Return only the text before "Klimawirkung"
	return text[:klimaIndex]
}
