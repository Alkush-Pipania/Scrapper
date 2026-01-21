package scraper

type ScrapedData struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
	SiteName    string `json:"site_name"`

	ContentText string `json:"content_text"`
	Author      string `json:"author"`
	PublishedAt string `json:"published_at"`
}
