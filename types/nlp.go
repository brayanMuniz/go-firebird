package types

type Sentiment struct {
	Magnitude float32 `firestore:"magnitude" json:"magnitude"`
	Score     float32 `firestore:"score" json:"score"`
}

// Entity represents a named entity detected in the text.
type Entity struct {
	Name     string            `json:"name"`
	Type     string            `json:"type"`
	Metadata map[string]string `json:"metadata"`
	Mentions []EntityMention   `json:"mentions"`
}

// EntityMention holds details about an entity mention.
type EntityMention struct {
	Content     string  `json:"content"`
	BeginOffset int32   `json:"begin_offset"`
	Probability float32 `json:"probability"`
}
