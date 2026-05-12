package youtube

type Channel struct {
	ID   string
	Name string
}

type Video struct {
	ID          string
	Title       string
	Description string
}

type ScrapedLink struct {
	URL        string
	VideoID    string
	VideoTitle string
	Context    string // surrounding text from the video description
}

// Quota tracks YouTube Data API v3 quota consumption.
// Daily quota resets at midnight Pacific Time (default 10,000 units).
type Quota struct {
	Used      int // cumulative units consumed
	Remaining int // from response headers, -1 if unknown
}

type apiChannelResponse struct {
	Items []struct {
		ID      string `json:"id"`
		Snippet struct {
			Title string `json:"title"`
		} `json:"snippet"`
	} `json:"items"`
}

type apiSearchResponse struct {
	NextPageToken string `json:"nextPageToken"`
	Items         []struct {
		ID struct {
			VideoID string `json:"videoId"`
		} `json:"id"`
		Snippet struct {
			Title string `json:"title"`
		} `json:"snippet"`
	} `json:"items"`
}

type apiVideosResponse struct {
	Items []struct {
		ID      string `json:"id"`
		Snippet struct {
			Title       string `json:"title"`
			Description string `json:"description"`
		} `json:"snippet"`
	} `json:"items"`
}
