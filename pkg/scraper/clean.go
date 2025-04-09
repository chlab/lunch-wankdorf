package scraper

import (
	"regexp"
	"strings"
)

// CleanHTML removes unnecessary parts of HTML like images, SVGs, comments, etc.
func CleanHTML(html string) string {
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
