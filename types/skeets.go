package types

type SaveSkeetResult struct {
	SavedSkeetID         string    `json:"savedSkeetId"`
	Content              string    `json:"content"`
	NewLocationNames     []string  `json:"newLocationNames"`
	ProcessedEntityCount int       `json:"processedEntityCount"`
	Classification       []float64 `json:"classification"`
	Sentiment            Sentiment `json:"sentiment"`
	AlreadyExist         bool      `json:"alreadyExist"`
	ErrorSaving          bool      `json:"errorSaving"`
}

type Category string

const (
	Wildfire    Category = "wildfire"
	Hurricane   Category = "hurricane"
	Earthquake  Category = "earthquake"
	NonDisaster Category = "non-disaster"
)

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
