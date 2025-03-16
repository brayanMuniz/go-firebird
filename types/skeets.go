package types

// Skeet represents a post stored in Firestore
type Skeet struct {
	Avatar         string    `firestore:"avatar"`
	Content        string    `firestore:"content"`
	Timestamp      string    `firestore:"timestamp"`
	Handle         string    `firestore:"handle"`
	DisplayName    string    `firestore:"displayName"`
	UID            string    `firestore:"uid"`
	Classification []float64 `firestore:"classification" json:"classification"`
	Sentiment      Sentiment `firestore:"sentiment" json:"sentiment"`
}

// This was made out of necessity because of how long the parameters would have been lol
type SaveCompleteSkeetType struct {
	NewSkeet       Skeet
	Classification []float64
	Entities       []Entity
	Sentiment      Sentiment
}
