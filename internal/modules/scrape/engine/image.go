package engine

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// extractImage tries multiple strategies to find the best image URL
func (s *Scraper) extractImage(doc *goquery.Document, baseURL string) string {
	// Priority 1: OG and Twitter meta tags
	imgURL := s.findMeta(doc, "og:image", "og:image:url", "og:image:secure_url", "twitter:image", "twitter:image:src")
	if imgURL != "" {
		return s.resolveURL(baseURL, imgURL)
	}

	// Priority 2: link rel="image_src" (some older sites)
	if href, exists := doc.Find(`link[rel="image_src"]`).Attr("href"); exists && href != "" {
		return s.resolveURL(baseURL, href)
	}

	// Priority 3: First large image in article/main content
	// Check common content containers for images
	selectors := []string{
		"article img",
		"main img",
		".post-content img",
		".entry-content img",
		".content img",
		"#content img",
	}

	for _, selector := range selectors {
		var foundImg string
		doc.Find(selector).EachWithBreak(func(i int, sel *goquery.Selection) bool {
			src := s.getImageSrc(sel)
			if src != "" && !s.isIconOrLogo(src) {
				foundImg = s.resolveURL(baseURL, src)
				return false // break
			}
			return true
		})
		if foundImg != "" {
			return foundImg
		}
	}

	// Priority 4: Any large image on the page (fallback)
	var fallbackImg string
	doc.Find("img").EachWithBreak(func(i int, sel *goquery.Selection) bool {
		src := s.getImageSrc(sel)
		if src != "" && !s.isIconOrLogo(src) {
			// Skip tiny images (likely icons/logos)
			width, _ := sel.Attr("width")
			height, _ := sel.Attr("height")
			if width != "" && (width == "1" || width == "16" || width == "32") {
				return true
			}
			if height != "" && (height == "1" || height == "16" || height == "32") {
				return true
			}
			fallbackImg = s.resolveURL(baseURL, src)
			return false
		}
		return true
	})

	return fallbackImg
}

// getImageSrc extracts the image source, handling lazy loading
func (s *Scraper) getImageSrc(sel *goquery.Selection) string {
	// Check standard src first
	if src, exists := sel.Attr("src"); exists && src != "" && !strings.HasPrefix(src, "data:") {
		return src
	}

	// Check lazy loading attributes
	lazyAttrs := []string{"data-src", "data-lazy-src", "data-original", "data-lazy", "data-srcset", "srcset"}
	for _, attr := range lazyAttrs {
		if val, exists := sel.Attr(attr); exists && val != "" {
			// For srcset, get the first URL
			if attr == "srcset" || attr == "data-srcset" {
				parts := strings.Split(val, ",")
				if len(parts) > 0 {
					firstSrc := strings.TrimSpace(strings.Split(parts[0], " ")[0])
					if firstSrc != "" {
						return firstSrc
					}
				}
			} else {
				return val
			}
		}
	}

	return ""
}

// isIconOrLogo tries to detect if an image URL is likely an icon or logo
func (s *Scraper) isIconOrLogo(src string) bool {
	lowered := strings.ToLower(src)
	iconPatterns := []string{
		"favicon", "icon", "logo", "sprite", "avatar", "badge",
		"spacer", "pixel", "tracking", "analytics", "1x1", "blank",
	}
	for _, pattern := range iconPatterns {
		if strings.Contains(lowered, pattern) {
			return true
		}
	}
	return false
}

// resolveURL converts relative URLs to absolute URLs
func (s *Scraper) resolveURL(baseURL, relativeURL string) string {
	if relativeURL == "" {
		return ""
	}

	// Already absolute
	if strings.HasPrefix(relativeURL, "http://") || strings.HasPrefix(relativeURL, "https://") {
		return relativeURL
	}

	// Protocol-relative URL
	if strings.HasPrefix(relativeURL, "//") {
		return "https:" + relativeURL
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return relativeURL
	}

	ref, err := url.Parse(relativeURL)
	if err != nil {
		return relativeURL
	}

	return base.ResolveReference(ref).String()
}

func (s *Scraper) findMeta(doc *goquery.Document, tags ...string) string {
	for _, tag := range tags {
		val := doc.Find(fmt.Sprintf("meta[property='%s']", tag)).AttrOr("content", "")
		if val != "" {
			return val
		}
		val = doc.Find(fmt.Sprintf("meta[name='%s']", tag)).AttrOr("content", "")
		if val != "" {
			return val
		}
	}
	// Special case for <title> tag
	for _, tag := range tags {
		if tag == "title" {
			return strings.TrimSpace(doc.Find("title").Text())
		}
	}
	return ""
}
