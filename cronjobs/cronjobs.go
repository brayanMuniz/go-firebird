package cronjobs

import (
	"context"
	"go-firebird/skeetprocessor"
	"go-firebird/types"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	language "cloud.google.com/go/language/apiv2"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/robfig/cron/v3"
)

const (
	feedMethod = "app.bsky.feed.getFeed"
)

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
	// Fire Feed: Run every 3 hours starting at 0:00.
	_, err := c.AddFunc("0 0-23/3 * * *", func() {
		log.Println("\nCronJob: Fire Feed Running")
		out, err := callFeed(fireURI)
		if err != nil {
			log.Println("Error getting Fire Feed", err)
		} else {
			skeetprocessor.SaveFeed(out, firestoreClient, nlpClient)
		}
	})
	if err != nil {
		log.Println("Error scheduling Fire Feed", err)
	}

	// Earthquake Feed: Run every 3 hours starting at 1:00.
	_, err = c.AddFunc("0 1-23/3 * * *", func() {
		log.Println("\nCronJob: EarthQuake Feed Running")
		callFeed(earthQuakeURI)
		out, err := callFeed(fireURI)
		if err != nil {
			log.Println("Error getting Earthquake Feed", err)
		} else {
			skeetprocessor.SaveFeed(out, firestoreClient, nlpClient)

		}

	})
	if err != nil {
		log.Println("Error scheduling EarthQuake Feed:", err)
	}

	// Hurricane Feed: Run every 3 hours starting at 2:00.
	_, err = c.AddFunc("0 2-23/3 * * *", func() {
		log.Println("\nCronJob: Hurricane Feed Running")
		out, err := callFeed(hurricaneURI)
		if err != nil {
			log.Println("Error getting Hurricane Feed", err)
		} else {
			skeetprocessor.SaveFeed(out, firestoreClient, nlpClient)
		}

	})
	if err != nil {
		log.Println("Error scheduling Hurricane Feed:", err)
	}

	c.Start()
}
