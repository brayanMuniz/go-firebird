package cronjobs

import (
	"context"
	"go-firebird/db"
	"go-firebird/processor"
	"go-firebird/types"
	"log"
	"net/http"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	language "cloud.google.com/go/language/apiv2"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/robfig/cron/v3"
)

const (
	feedMethod = "app.bsky.feed.getFeed"
)

// scheduleLocationSentimentUpdate retrieves all valid locations, processes each in parallel,
// and logs a summary of successes and failures.
func scheduleLocationSentimentUpdate(firestoreClient *firestore.Client) {
	// Fetch all valid locations.
	validLocations, err := db.GetValidLocations(firestoreClient)
	if err != nil {
		log.Printf("Error fetching valid locations: %v", err)
		return
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	goodSaving := make([]string, 0)
	failureSaving := make([]string, 0)
	// My little variables cant be this cute!

	// Process each location concurrently.
	for _, locData := range validLocations {
		wg.Add(1)
		go func(location types.LocationData) {
			defer wg.Done()

			// Use a hash of the location name as the document ID.
			docId := db.HashString(location.LocationName)
			if err := processor.ProcessLocationAvgSentiment(firestoreClient, docId, locData); err != nil {
				log.Printf("Error processing location %s: %v", docId, err)
				mu.Lock()
				failureSaving = append(failureSaving, docId)
				mu.Unlock()
			} else {
				mu.Lock()
				goodSaving = append(goodSaving, docId)
				mu.Unlock()
			}
		}(locData)
	}

	wg.Wait()

	log.Printf("Location Sentiment Update Completed at %s", time.Now().UTC().Format(time.RFC3339))
	log.Printf("Successful updates: %v", goodSaving)
	log.Printf("Failed updates: %v", failureSaving)
}

func callFeed(uri string) (types.FeedResponse, error) {
	client := &xrpc.Client{
		Client:    &http.Client{Timeout: 10 * time.Second},
		Host:      "https://public.api.bsky.app", // public endpoint for unauthenticated requests.
		UserAgent: nil,
	}

	feedAtURI := uri

	// The limit can be adjusted (min 1, max 100, default 50).
	params := map[string]interface{}{
		"feed":  feedAtURI,
		"limit": 50,
	}

	log.Printf("Fetching feed with params: %+v", params)

	var out types.FeedResponse

	// Call the Bluesky API using the xrpc client.
	err := client.Do(context.Background(), xrpc.Query, "json", feedMethod, params, nil, &out)
	if err != nil {
		log.Printf("Error fetching feed via xrpc: %v", err)
		return out, err
	}
	return out, nil

}

func InitCronJobs(firestoreClient *firestore.Client, nlpClient *language.Client) {
	log.Println("\nStarting Cron Jobs -------------------------------------------------------")

	fireURI := "at://did:plc:qiknc4t5rq7yngvz7g4aezq7/app.bsky.feed.generator/aaaejsyozb6iq"
	earthQuakeURI := "at://did:plc:qiknc4t5rq7yngvz7g4aezq7/app.bsky.feed.generator/aaaejxlobe474"
	hurricaneURI := "at://did:plc:qiknc4t5rq7yngvz7g4aezq7/app.bsky.feed.generator/aaaejwgffwqky"

	c := cron.New()
	// Fire Feed: Run every 4 hours starting at 0:00.
	_, err := c.AddFunc("0 0-23/4 * * *", func() {
		log.Println("\nCronJob: Fire Feed Running")
		out, err := callFeed(fireURI)
		if err != nil {
			log.Println("Error getting Fire Feed", err)
		} else {
			processor.SaveFeed(out, firestoreClient, nlpClient)
		}
	})
	if err != nil {
		log.Println("Error scheduling Fire Feed", err)
	}

	// Earthquake Feed: Run every 4 hours starting at 1:00.
	_, err = c.AddFunc("0 1-23/4 * * *", func() {
		log.Println("\nCronJob: EarthQuake Feed Running")
		callFeed(earthQuakeURI)
		out, err := callFeed(fireURI)
		if err != nil {
			log.Println("Error getting Earthquake Feed", err)
		} else {
			processor.SaveFeed(out, firestoreClient, nlpClient)

		}

	})
	if err != nil {
		log.Println("Error scheduling EarthQuake Feed:", err)
	}

	// Hurricane Feed: Run every 4 hours starting at 2:00.
	_, err = c.AddFunc("0 2-23/4 * * *", func() {
		log.Println("\nCronJob: Hurricane Feed Running")
		out, err := callFeed(hurricaneURI)
		if err != nil {
			log.Println("Error getting Hurricane Feed", err)
		} else {
			processor.SaveFeed(out, firestoreClient, nlpClient)
		}

	})
	if err != nil {
		log.Println("Error scheduling Hurricane Feed:", err)
	}

	// Update location average sentiment every 12 hours.
	_, locErr := c.AddFunc("0 0,12 * * *", func() {
		log.Println("\nCronJob: Updating average sentiment for all locations")
		scheduleLocationSentimentUpdate(firestoreClient)
	})
	if locErr != nil {
		log.Printf("Error scheduling Location Sentiment CronJob: %v", err)
	}

	c.Start()
}
