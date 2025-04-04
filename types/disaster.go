package types

type Severity string

const (
	Low      Severity = "low"
	Medium   Severity = "medium"
	High     Severity = "high"
	Critical Severity = "critical"
)

type Status string

const (
	Not_Active Status = "not_active"
	Active     Status = "active"
	Recovery   Status = "recovery"
)

type DisasterData struct {
	// Fields representing the *cluster*
	ID            string      `firestore:"-"`
	Lat           float64     `firestore:"lat"`         // Centroid Latitude
	Long          float64     `firestore:"long"`        // Centroid Longitude
	LocationIDs   []string    `firestore:"locationIDs"` // IDs of locations included in this disaster
	LocationCount int         `firestore:"locationCount"`
	BoundingBox   BoundingBox `firestore:"boundingBox"`

	// Aggregated/Derived Disaster Info
	DisasterType Category `firestore:"disasterType"` // Dominant or seed type
	Severity     Severity `firestore:"severity,omitempty"`
	Status       Status   `firestore:"status,omitempty"`
	ReportedDate string   `firestore:"reportedDate"`      // Earliest firstSkeetTimestamp in the cluster
	LastUpdate   string   `firestore:"lastUpdate"`        // Latest lastSkeetTimestamp in the cluster
	Summary      string   `firestore:"summary,omitempty"` // To be filled later by LLM

	// Aggregated counts/sentiment for the cluster
	TotalSkeetsAmount int           `firestore:"totalSkeetsAmount"`
	ClusterSentiment  float32       `firestore:"clusterSentiment"`
	ClusterCounts     DisasterCount `firestore:"clusterCounts"`
}

type BoundingBox struct {
	MinLat float64 `firestore:"minLat"`
	MaxLat float64 `firestore:"maxLat"`
	MinLon float64 `firestore:"minLon"`
	MaxLon float64 `firestore:"maxLon"`
}

