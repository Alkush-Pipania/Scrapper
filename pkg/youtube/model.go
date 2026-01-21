package youtube

type VideoData struct {
	VideoID      string `json:"video_id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	ChannelTitle string `json:"channel_title"`
	ChannelID    string `json:"channel_id"`
	PublishedAt  string `json:"published_at"`
	ThumbnailURL string `json:"thumbnail_url"`
	Duration     string `json:"duration"`
	ViewCount    string `json:"view_count"`
	LikeCount    string `json:"like_count"`
	Transcript   string `json:"transcript,omitempty"`
}

type videoListResponse struct {
	Items []videoItem `json:"items"`
}

type videoItem struct {
	ID      string         `json:"id"`
	Snippet snippetData    `json:"snippet"`
	Stats   statisticsData `json:"statistics"`
	Details contentDetails `json:"contentDetails"`
}

type snippetData struct {
	Title        string         `json:"title"`
	Description  string         `json:"description"`
	ChannelTitle string         `json:"channelTitle"`
	ChannelID    string         `json:"channelId"`
	PublishedAt  string         `json:"publishedAt"`
	Thumbnails   thumbnailsData `json:"thumbnails"`
}

type thumbnailsData struct {
	MaxRes  thumbnailInfo `json:"maxres"`
	High    thumbnailInfo `json:"high"`
	Medium  thumbnailInfo `json:"medium"`
	Default thumbnailInfo `json:"default"`
}

type thumbnailInfo struct {
	URL string `json:"url"`
}

type statisticsData struct {
	ViewCount string `json:"viewCount"`
	LikeCount string `json:"likeCount"`
}

type contentDetails struct {
	Duration string `json:"duration"`
}

type captionListResponse struct {
	Items []captionItem `json:"items"`
}

type captionItem struct {
	ID      string             `json:"id"`
	Snippet captionSnippetData `json:"snippet"`
}

type captionSnippetData struct {
	Language string `json:"language"`
	Name     string `json:"name"`
}
