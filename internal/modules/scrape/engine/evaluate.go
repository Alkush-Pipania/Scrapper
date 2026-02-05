package engine

import (
	"strings"

	domain "github.com/Alkush-Pipania/Scrapper/internal/domain/scrape"
)

const (
	minContentLen     = 200
	minDescriptionLen = 40
)

type staticEval struct {
	ok              bool
	needsScreenshot bool
	reason          string
}

func (s *Scraper) evaluateStatic(d *domain.ScrapedData) staticEval {
	if d == nil {
		return staticEval{ok: false, needsScreenshot: true, reason: "no data"}
	}

	titleOK := strings.TrimSpace(d.Title) != ""
	contentLen := len(strings.TrimSpace(d.ContentText))
	contentOK := contentLen >= minContentLen
	descLen := len(strings.TrimSpace(d.Description))
	descOK := descLen >= minDescriptionLen

	dataOK := titleOK && (contentOK || descOK)

	imgOK := d.ImageURL != "" && !s.isIconOrLogo(d.ImageURL)
	needsScreenshot := !imgOK

	var reasons []string
	if !titleOK {
		reasons = append(reasons, "missing title")
	}
	if !contentOK && !descOK {
		reasons = append(reasons, "missing content/description")
	}
	if !imgOK {
		reasons = append(reasons, "missing/weak image")
	}

	return staticEval{
		ok:              dataOK,
		needsScreenshot: needsScreenshot,
		reason:          strings.Join(reasons, ", "),
	}
}
