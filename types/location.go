package types

type LocationData struct {
	ID                  string                 `firestore:"-"` // tell firestore to ignore
	LocationName        string                 `firestore:"locationName"`
	FormattedAddress    string                 `firestore:"formattedAddress"`
	Lat                 float64                `firestore:"lat"`
	Long                float64                `firestore:"long"`
	Type                string                 `firestore:"type"`
	AvgSentimentList    []AvgLocationSentiment `firestore:"avgSentimentList"`
	LatestSkeetsAmount  int                    `firestore:"latestSkeetsAmount"`
	LatestDisasterCount DisasterCount          `firestore:"latestDisasterCount"`
	LatestSentiment     float32                `firestore:"latestSentiment"`
	FirstSkeetTimestamp string                 `firestore:"firstSkeetTimestamp,omitempty"`
	LastSkeetTimestamp  string                 `firestore:"lastSkeetTimestamp,omitempty"`
}

type DisasterCount struct {
	FireCount        int `firestore:"fireCount"`
	HurricaneCount   int `firestore:"hurricaneCount"`
	EarthquakeCount  int `firestore:"earthquakeCount"`
	NonDisasterCount int `firestore:"nonDisasterCount"`
}

type AvgLocationSentiment struct {
	TimeStamp        string        `firestore:"timeStamp"`
	SkeetsAmount     int           `firestore:"skeetsAmount"`
	AverageSentiment float32       `firestore:"averageSentiment"`
	DisasterCount    DisasterCount `firestore:"disasterCount"`
}

type NewLocationMetaData struct {
	LocationName string
	Type         string
	NewLocation  bool
}

// skeet stored under a location's subcollection.
type SkeetSubDoc struct {
	Mentions     []EntityMention `firestore:"mentions" json:"mentions"`
	LocationName string          `firestore:"locationName" json:"locationName"`
	Type         string          `firestore:"type" json:"type"`
	SkeetData    Skeet           `firestore:"skeetData" json:"skeetData"`
}
