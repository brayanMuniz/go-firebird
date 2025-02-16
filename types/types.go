package types

import ()

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
