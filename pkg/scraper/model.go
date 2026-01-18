package scraper

type ScrapedData struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
	SiteName    string `json:"site_name"`

	// The Data
	ContentText string `json:"content_text"`
	ContentHTML string `json:"content_html"`
	Author      string `json:"author"`
	PublishedAt string `json:"published_at"`
}
