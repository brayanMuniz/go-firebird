package detection

import (
	"fmt"
	"github.com/google/uuid"
	"go-firebird/types"
	"math"
	"sort"
	"time"
)

const (
	sentimentThreshold  float32 = -0.05
	minDisasterCount    int     = 3    // Min count in *any* disaster category to be a seed
	distanceThresholdKM         = 50.0 // Max distance (km) to cluster locations
	earthRadiusKM               = 6371.0
	// --- Severity Thresholds  ---

	// Skeet Counts
	mediumSkeetThreshold = 30
	highSkeetThreshold   = 100
	critSkeetThreshold   = 150

	// Sentiment
	mediumSentThreshold float32 = -.43
	highSentThreshold   float32 = -.53
	critSentThreshold   float32 = -.63

	// Location Count (Geographic Spread)
	mediumLocCountThreshold = 5
	highLocCountThreshold   = 10
	critLocCountThreshold   = 20
)

func DetectDisastersFromList(locations []types.LocationData) ([]types.DisasterData, error) {
	var disasters []types.DisasterData
	processedLocationIDs := make(map[string]bool)

	// 1. Identify potential disaster "seeds"
	var seeds []*types.LocationData
	for i := range locations {
		loc := &locations[i]

		// --- Seed Criteria ---
		// Location must have a ID to be processed
		if loc.ID == "" {
			fmt.Printf("Warning: Location at index %d (%s) missing ID, skipping.\n", i, loc.LocationName)
			continue
		}

		// Check if sentiment is low AND at least one disaster count is significant
		counts := loc.LatestDisasterCount
		hasSignificantDisasterCount := counts.FireCount >= minDisasterCount ||
			counts.HurricaneCount >= minDisasterCount ||
			counts.EarthquakeCount >= minDisasterCount
		// NOTE: exclude NonDisasterCount from seeding criteria

		isPotentialSeed := loc.LatestSentiment <= sentimentThreshold && hasSignificantDisasterCount
		if isPotentialSeed {
			seeds = append(seeds, loc)
		}
	}

	// 2. Cluster locations around seeds
	for _, seed := range seeds {
		if processedLocationIDs[seed.ID] {
			continue
		}

		fmt.Printf("Starting cluster analysis for seed: %s (%s)\n", seed.LocationName, seed.ID)
		clusterLocations := []*types.LocationData{}
		queue := []*types.LocationData{seed}
		clusterProcessedIDs := make(map[string]bool)
		clusterProcessedIDs[seed.ID] = true

		for len(queue) > 0 {
			currentLoc := queue[0]
			queue = queue[1:]

			if !processedLocationIDs[currentLoc.ID] {
				clusterLocations = append(clusterLocations, currentLoc)
				processedLocationIDs[currentLoc.ID] = true
			}

			for i := range locations {
				neighbor := &locations[i]
				if neighbor.ID == "" || neighbor.ID == currentLoc.ID || clusterProcessedIDs[neighbor.ID] {
					continue
				}
				dist := haversineDistance(currentLoc.Lat, currentLoc.Long, neighbor.Lat, neighbor.Long)
				if dist <= distanceThresholdKM {
					clusterProcessedIDs[neighbor.ID] = true
					queue = append(queue, neighbor)
				}
			}
		}

		// 3. If cluster found, create DisasterData object
		if len(clusterLocations) > 0 {
			fmt.Printf("  Cluster found with %d locations.\n", len(clusterLocations))
			clusterDisasterType := determineClusterDisasterType(clusterLocations)
			if clusterDisasterType != types.NonDisaster {
				disaster := createDisasterFromCluster(clusterLocations, clusterDisasterType)
				disasters = append(disasters, disaster)
			} else {
				fmt.Printf("  Cluster around seed %s classified as NonDisaster based on counts, skipping disaster creation.\n", seed.ID)
			}
		}
	}

	fmt.Printf("Found %d potential disaster seeds out of %d locations.\n", len(seeds), len(locations))
	fmt.Printf("Detection complete. Identified %d distinct disaster clusters.\n", len(disasters))
	return disasters, nil
}

// analyzes the aggregated counts to find the dominant disaster type.
func determineClusterDisasterType(cluster []*types.LocationData) types.Category {
	var totalCounts types.DisasterCount
	for _, loc := range cluster {
		totalCounts.FireCount += loc.LatestDisasterCount.FireCount
		totalCounts.HurricaneCount += loc.LatestDisasterCount.HurricaneCount
		totalCounts.EarthquakeCount += loc.LatestDisasterCount.EarthquakeCount
		// NOTE: don't need NonDisasterCount for determining the *disaster* type
	}

	// Find the category with the highest count
	maxCount := 0
	dominantType := types.NonDisaster // Default to non-disaster

	if totalCounts.FireCount > maxCount {
		maxCount = totalCounts.FireCount
		dominantType = types.Wildfire
	}
	if totalCounts.HurricaneCount > maxCount {
		maxCount = totalCounts.HurricaneCount
		dominantType = types.Hurricane
	}
	if totalCounts.EarthquakeCount > maxCount {
		dominantType = types.Earthquake
	}

	return dominantType
}

// aggregates data from clustered locations into a DisasterData object.
func createDisasterFromCluster(cluster []*types.LocationData, clusterType types.Category) types.DisasterData {
	if len(cluster) == 0 {
		return types.DisasterData{}
	}

	disaster := types.DisasterData{
		ID: uuid.NewString(),
		// Firestore Document ID
		LocationIDs:   make([]string, 0, len(cluster)),
		LocationCount: len(cluster),
		DisasterType:  clusterType,
		Status:        types.Active,
		Severity:      types.Low,
		BoundingBox: types.BoundingBox{
			MinLat: cluster[0].Lat, MaxLat: cluster[0].Lat,
			MinLon: cluster[0].Long, MaxLon: cluster[0].Long,
		},
		TotalSkeetsAmount: 0,
		ClusterSentiment:  0,
		ClusterCounts:     types.DisasterCount{},
	}

	var sumLat, sumLon, totalSentiment float64
	var earliestFirstTime, latestLastTime time.Time
	var firstTimestampSet, lastTimestampSet bool

	for _, loc := range cluster {
		if loc.ID == "" {
			// This should ideally not happen if GetLocationsForDisasterCheck works correctly
			fmt.Printf("Warning: Location %s is missing Firestore ID during cluster creation. Skipping ID.\n", loc.FormattedAddress)
		} else {
			// Add the Firestore Document ID
			disaster.LocationIDs = append(disaster.LocationIDs, loc.ID)
		}

		// Update Bounding Box
		if loc.Lat < disaster.BoundingBox.MinLat {
			disaster.BoundingBox.MinLat = loc.Lat
		}
		if loc.Lat > disaster.BoundingBox.MaxLat {
			disaster.BoundingBox.MaxLat = loc.Lat
		}
		if loc.Long < disaster.BoundingBox.MinLon {
			disaster.BoundingBox.MinLon = loc.Long
		}
		if loc.Long > disaster.BoundingBox.MaxLon {
			disaster.BoundingBox.MaxLon = loc.Long
		}

		// Sum for Centroid & Avg Sentiment
		sumLat += loc.Lat
		sumLon += loc.Long
		totalSentiment += float64(loc.LatestSentiment)

		// Aggregate Counts
		disaster.TotalSkeetsAmount += loc.LatestSkeetsAmount
		disaster.ClusterCounts.FireCount += loc.LatestDisasterCount.FireCount
		disaster.ClusterCounts.HurricaneCount += loc.LatestDisasterCount.HurricaneCount
		disaster.ClusterCounts.EarthquakeCount += loc.LatestDisasterCount.EarthquakeCount
		disaster.ClusterCounts.NonDisasterCount += loc.LatestDisasterCount.NonDisasterCount // Still aggregate this for info

		// Find earliest first timestamp and latest last timestamp (using robust parsing)
		parseAndUpdateTimestamps(loc.FirstSkeetTimestamp, loc.LastSkeetTimestamp, loc.ID,
			&earliestFirstTime, &latestLastTime, &firstTimestampSet, &lastTimestampSet)
	}

	// Calculate Centroid
	count := float64(len(cluster))
	disaster.Lat = sumLat / count
	disaster.Long = sumLon / count

	// Average Cluster Sentiment
	if count > 0 {
		disaster.ClusterSentiment = float32(totalSentiment / count)
	}

	// Set aggregated dates (ISO 8601 format)
	if firstTimestampSet {
		disaster.ReportedDate = earliestFirstTime.UTC().Format(time.RFC3339)
	}
	if lastTimestampSet {
		disaster.LastUpdate = latestLastTime.UTC().Format(time.RFC3339)
	}

	// set sev type
	if disaster.ClusterSentiment < mediumSentThreshold {
		disaster.Severity = types.Medium
	}

	if disaster.ClusterSentiment < highSentThreshold {
		disaster.Severity = types.High
	}

	if disaster.ClusterSentiment < critSentThreshold {
		disaster.Severity = types.Critical
	}

	// Sort LocationIDs for consistency
	sort.Strings(disaster.LocationIDs)

	return disaster
}

// Helper function to parse timestamps robustly
func parseAndUpdateTimestamps(firstTSStr, lastTSStr, locIdentifier string,
	earliestFirstTime, latestLastTime *time.Time,
	firstTimestampSet, lastTimestampSet *bool) {

	parseTime := func(tsStr string) (time.Time, bool) {
		if tsStr == "" {
			return time.Time{}, false
		}
		t, err := time.Parse(time.RFC3339Nano, tsStr)
		if err == nil {
			return t, true
		}
		t, err = time.Parse("2006-01-02T15:04:05Z", tsStr) // without fractional seconds
		if err == nil {
			return t, true
		}
		fmt.Printf("Warning: Could not parse timestamp '%s' for loc %s\n", tsStr, locIdentifier)
		return time.Time{}, false
	}

	if t, ok := parseTime(firstTSStr); ok {
		if !*firstTimestampSet || t.Before(*earliestFirstTime) {
			*earliestFirstTime = t
			*firstTimestampSet = true
		}
	}
	if t, ok := parseTime(lastTSStr); ok {
		if !*lastTimestampSet || t.After(*latestLastTime) {
			*latestLastTime = t
			*lastTimestampSet = true
		}
	}
}

// haversineDistance calculates the great-circle distance between two points
// on the earth (specified in decimal degrees).
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	radLat1 := lat1 * math.Pi / 180
	radLon1 := lon1 * math.Pi / 180
	radLat2 := lat2 * math.Pi / 180
	radLon2 := lon2 * math.Pi / 180

	deltaLat := radLat2 - radLat1
	deltaLon := radLon2 - radLon1

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(radLat1)*math.Cos(radLat2)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKM * c
}
