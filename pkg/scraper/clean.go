package scraper

import (
	"bytes"
	gohtml "html"
	"log"
	"regexp"
	"strings"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
)

// Pre-compiled regexes for HTML cleaning
var (
	reComment       = regexp.MustCompile(`<!--[\s\S]*?-->`)
	reStyle         = regexp.MustCompile(`<style[^>]*>[\s\S]*?</style>`)
	reScript        = regexp.MustCompile(`<script[^>]*>[\s\S]*?</script>`)
	reSVG           = regexp.MustCompile(`<svg[^>]*>[\s\S]*?</svg>`)
	reImg           = regexp.MustCompile(`<img[^>]*>`)
	reIframe        = regexp.MustCompile(`<iframe[^>]*>[\s\S]*?</iframe>`)
	reVideo         = regexp.MustCompile(`<video[^>]*>[\s\S]*?</video>`)
	reAudio         = regexp.MustCompile(`<audio[^>]*>[\s\S]*?</audio>`)
	reCanvas        = regexp.MustCompile(`<canvas[^>]*>[\s\S]*?</canvas>`)
	rePath          = regexp.MustCompile(`<path[^>]*>[\s\S]*?</path>`)
	reObject        = regexp.MustCompile(`<object[^>]*>[\s\S]*?</object>`)
	reHidden        = regexp.MustCompile(`<[^>]* hidden[^>]*>[\s\S]*?</[^>]*>`)
	reDisplayNone   = regexp.MustCompile(`<[^>]* style="[^"]*display:\s*none[^"]*"[^>]*>[\s\S]*?</[^>]*>`)
	reVisHidden     = regexp.MustCompile(`<[^>]* style="[^"]*visibility:\s*hidden[^"]*"[^>]*>[\s\S]*?</[^>]*>`)
	reClassDouble   = regexp.MustCompile(` class="[^"]*"`)
	reClassSingle   = regexp.MustCompile(` class='[^']*'`)
	reStyleDouble   = regexp.MustCompile(` style="[^"]*"`)
	reStyleSingle   = regexp.MustCompile(` style='[^']*'`)
	reTargetDouble  = regexp.MustCompile(` target="[^"]*"`)
	reTargetSingle  = regexp.MustCompile(` target='[^']*'`)
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
		log.Printf("Error minifying HTML: %v, using original content", err)
		minifiedContent = htmlContent
	} else {
		minified := output.String()
		log.Printf("Minified HTML from %d to %d bytes (%.1f%%)",
			len(htmlContent),
			len(minified),
			float64(len(minified))/float64(len(htmlContent))*100)
		minifiedContent = minified
	}
	return minifiedContent
}

// cleanHTML removes unnecessary parts of HTML like images, SVGs, comments, etc.
func cleanHTML(htmlContent string) string {
	htmlContent = reComment.ReplaceAllString(htmlContent, "")
	htmlContent = reStyle.ReplaceAllString(htmlContent, "")
	htmlContent = reScript.ReplaceAllString(htmlContent, "")
	htmlContent = reSVG.ReplaceAllString(htmlContent, "")
	htmlContent = reImg.ReplaceAllString(htmlContent, "")
	htmlContent = reIframe.ReplaceAllString(htmlContent, "")
	htmlContent = reVideo.ReplaceAllString(htmlContent, "")
	htmlContent = reAudio.ReplaceAllString(htmlContent, "")
	htmlContent = reCanvas.ReplaceAllString(htmlContent, "")
	htmlContent = rePath.ReplaceAllString(htmlContent, "")
	htmlContent = reObject.ReplaceAllString(htmlContent, "")
	htmlContent = reHidden.ReplaceAllString(htmlContent, "")
	htmlContent = reDisplayNone.ReplaceAllString(htmlContent, "")
	htmlContent = reVisHidden.ReplaceAllString(htmlContent, "")

	// Decode all HTML entities (covers &nbsp;, &amp;, &lt;, &gt;, numeric entities, etc.)
	htmlContent = gohtml.UnescapeString(htmlContent)

	return htmlContent
}

// stripTags removes class and style attributes from HTML tags
func stripTags(htmlContent string) string {
	htmlContent = reClassDouble.ReplaceAllString(htmlContent, "")
	htmlContent = reClassSingle.ReplaceAllString(htmlContent, "")
	htmlContent = reStyleDouble.ReplaceAllString(htmlContent, "")
	htmlContent = reStyleSingle.ReplaceAllString(htmlContent, "")
	htmlContent = reTargetDouble.ReplaceAllString(htmlContent, "")
	htmlContent = reTargetSingle.ReplaceAllString(htmlContent, "")
	return htmlContent
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
