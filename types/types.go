package types

// TweetAnalysis wraps multiple tweets
type TweetAnalysis struct {
	Tweets []Tweet `json:"tweets"`
}

type Tweet struct {
	OriginalTweetData struct {
		TweetText string `json:"tweet_text" jsonschema:"default="`
		Timestamp string `json:"timestamp" jsonschema:"default="`
		User      struct {
			Username       string `json:"username" jsonschema:"default="`
			Verified       bool   `json:"verified" jsonschema:"default=false"`
			FollowersCount int    `json:"followers_count" jsonschema:"default=0"`
		} `json:"user"`
		Hashtags []string `json:"hashtags" jsonschema:"default=[]"`
	} `json:"original_tweet_data"`
	Analysis []struct {
		// Only allowed disasters
		DisasterName       string  `json:"disaster_name" jsonschema:"enum=Wildfire,Flood,Earthquake,Hurricane,Tornado,Tsunami,Volcanic Eruption,Landslide,Blizzard,Drought;default="`
		RelevantToDisaster float64 `json:"relevant_to_disaster" jsonschema:"default=0"`
		Sentiment          struct {
			Emotion  string  `json:"emotion" jsonschema:"default="`
			Tone     string  `json:"tone" jsonschema:"default="`
			Polarity float64 `json:"polarity" jsonschema:"default=0"`
			Details  struct {
				Urgency         float64 `json:"urgency" jsonschema:"default=0"`
				EmotionalImpact float64 `json:"emotional_impact" jsonschema:"default=0"`
			} `json:"details"`
		} `json:"sentiment"`
		SarcasmConfidence float64 `json:"sarcasm_confidence" jsonschema:"default=0"`
	} `json:"analysis"`
}

// FeedResponse represents the root structure of the API response.
type FeedResponse struct {
	Cursor string      `json:"cursor"`
	Feed   []FeedEntry `json:"feed"`
}

// FeedEntry represents each post in the feed.
type FeedEntry struct {
	Post Post `json:"post"`
}

// Post represents the structure of an individual post.
type Post struct {
	Author      Author  `json:"author"`
	CID         string  `json:"cid"`
	Embed       *Embed  `json:"embed,omitempty"` // Nullable field
	IndexedAt   string  `json:"indexedAt"`
	Labels      []Label `json:"labels"`
	LikeCount   int     `json:"likeCount"`
	QuoteCount  int     `json:"quoteCount"`
	Record      Record  `json:"record"`
	ReplyCount  int     `json:"replyCount"`
	RepostCount int     `json:"repostCount"`
	URI         string  `json:"uri"`
}

// Author represents the author of a post.
type Author struct {
	Avatar      string  `json:"avatar"`
	CreatedAt   string  `json:"createdAt"`
	DID         string  `json:"did"`
	DisplayName string  `json:"displayName"`
	Handle      string  `json:"handle"`
	Labels      []Label `json:"labels,omitempty"`
}

// Label represents labels associated with a post or author.
type Label struct {
	CID string `json:"cid"`
	CTS string `json:"cts"`
	Src string `json:"src"`
	URI string `json:"uri"`
	Val string `json:"val"`
	Ver int    `json:"ver,omitempty"`
}

// Embed represents an embedded image, video, or record in a post.
type Embed struct {
	Type   string       `json:"$type"`
	Images []ImageEmbed `json:"images,omitempty"`
	Media  *MediaEmbed  `json:"media,omitempty"`
	Record *RecordEmbed `json:"record,omitempty"`
}

// ImageEmbed represents an image embedded in a post.
type ImageEmbed struct {
	Alt         string      `json:"alt"`
	AspectRatio AspectRatio `json:"aspectRatio"`
	Fullsize    string      `json:"fullsize"`
	Thumb       string      `json:"thumb"`
}

// AspectRatio represents the width and height of an image.
type AspectRatio struct {
	Height int `json:"height"`
	Width  int `json:"width"`
}

// MediaEmbed represents video or external media in an embed.
type MediaEmbed struct {
	Type        string      `json:"$type"`
	Alt         string      `json:"alt"`
	AspectRatio AspectRatio `json:"aspectRatio"`
	Thumbnail   string      `json:"thumbnail,omitempty"`
	Playlist    string      `json:"playlist,omitempty"`
	Video       *Blob       `json:"video,omitempty"`
}

// RecordEmbed represents an embedded record within a post.
type RecordEmbed struct {
	Type   string      `json:"$type"`
	Record EmbedRecord `json:"record"`
}

// EmbedRecord represents metadata about an embedded post.
type EmbedRecord struct {
	CID string `json:"cid"`
	URI string `json:"uri"`
}

// Record represents the content of a post.
type Record struct {
	Type      string   `json:"$type"`
	CreatedAt string   `json:"createdAt"`
	Embed     *Embed   `json:"embed,omitempty"`
	Facets    []Facet  `json:"facets,omitempty"`
	Langs     []string `json:"langs"`
	Text      string   `json:"text"`
}

// Facet represents features like tags or links in a post.
type Facet struct {
	Features []Feature `json:"features"`
	Index    Index     `json:"index"`
}

// Feature represents a tag or link feature in text.
type Feature struct {
	Type string `json:"$type"`
	Tag  string `json:"tag,omitempty"`
	URI  string `json:"uri,omitempty"`
}

// Index represents the position of a facet in text.
type Index struct {
	ByteStart int `json:"byteStart"`
	ByteEnd   int `json:"byteEnd"`
}

// Blob represents binary data such as images or videos.
type Blob struct {
	Type     string  `json:"$type"`
	MimeType string  `json:"mimeType"`
	Ref      RefLink `json:"ref"`
	Size     int     `json:"size"`
}

// RefLink represents a reference to a file.
type RefLink struct {
	Link string `json:"$link"`
}
