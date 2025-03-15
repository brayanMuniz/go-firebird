package cronjobs

import (
	"context"
	"go-firebird/types"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	language "cloud.google.com/go/language/apiv2"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/robfig/cron/v3"
)

// FetchBlueskyHandler fetches a hydrated feed using the Bluesky API.
type FeedCallParameters struct {
	uri   string
	limit int
}

const (
	feedMethod = "app.bsky.feed.getFeed"
)

// TODO: update this so it returns a list of gotten post and an err if there were any
func callFeed(uri string) (types.FeedResponse, error) {
	client := &xrpc.Client{
		Client:    &http.Client{Timeout: 10 * time.Second},
		Host:      "https://public.api.bsky.app", // public endpoint for unauthenticated requests.
		UserAgent: nil,
	}

	feedAtURI := uri

	// Read query parameters
	limit := 50

	// The limit can be adjusted (min 1, max 100, default 50).
	params := map[string]interface{}{
		"feed":  feedAtURI,
		"limit": limit,
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

	// Fire Feed: Run every 10 minutes at 0 minutes
	_, err := c.AddFunc("*/10 * * * *", func() {
		log.Println("\nCronJob: Fire Feed Running")
		// TODO:
		_, err := callFeed(fireURI)
		if err != nil {
			log.Println("Error getting Fire Feed", err)
		}
	})
	if err != nil {
		log.Println("Error scheduling Fire Feed", err)
	}

	// Earthquake Feed: Run every 10 minutes at 2 minute mark
	_, err = c.AddFunc("2-59/10 * * * *", func() {
		log.Println("\nCronJob: EarthQuake Feed Running")
		callFeed(earthQuakeURI)

	})
	if err != nil {
		log.Println("Error scheduling EarthQuake Feed:", err)
	}

	// Hurricane Feed: Run every 10 minutes at 4 minutes mark
	_, err = c.AddFunc("4-59/10 * * * *", func() {
		log.Println("\nCronJob: Hurricane Feed Running")
		callFeed(hurricaneURI)
	})
	if err != nil {
		log.Println("Error scheduling Hurricane Feed:", err)
	}

	c.Start()
}
