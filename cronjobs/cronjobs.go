package cronjobs

import (
	"bytes"
	"context"
	"encoding/json"
	"go-firebird/types"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bluesky-social/indigo/xrpc"
	"github.com/robfig/cron/v3"
)

// Example output
// 2025/02/24 08:47:00 Function 3 running every 5 minutes, 2 minutes apart
// 2025/02/24 08:50:00 Function 1 running every 5 minutes
// 2025/02/24 08:51:00 Function 2 running every 5 minutes, 1 minute apart
// 2025/02/24 08:52:00 Function 3 running every 5 minutes, 2 minutes apart

// NOTE: This will later be changed to send to the model, but for testing this works

// sendToClient sends the fetched feed data to the client URL specified in the environment
func sendToClient(data types.FeedResponse) {
	clientBaseURL := os.Getenv("CLIENT_URL")
	if clientBaseURL == "" {
		log.Println("CLIENT_URL is not set. Defaulting to http://localhost:3000")
		clientBaseURL = "http://localhost:3000"
	}

	// Append the API route to the base URL
	clientURL := clientBaseURL + "/api/testing"
	log.Println("cronjobs client_url: ", clientURL)

	// Convert the struct to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return
	}

	// Send a POST request to the client URL
	resp, err := http.Post(clientURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error sending data to %s: %v", clientURL, err)
		return
	}
	defer resp.Body.Close()

	log.Printf("Sent data to %s with response: %v", clientURL, resp.Status)
}

// FetchBlueskyHandler fetches a hydrated feed using the Bluesky API.
type FeedCallParameters struct {
	uri    string
	limit  int
	cursor string
}

const (
	feedMethod = "app.bsky.feed.getFeed"
)

func callFeed(p FeedCallParameters) {
	// Initialize the xrpc client to use the public API endpoint.
	client := &xrpc.Client{
		Client:    &http.Client{Timeout: 10 * time.Second},
		Host:      "https://public.api.bsky.app", // public endpoint for unauthenticated requests.
		UserAgent: nil,                           // can set a custom agent
	}

	// This is the fire URI
	feedAtURI := p.uri

	// Read query parameters
	limit := 10
	if p.limit != 0 {
		limit = p.limit

	}

	// The limit can be adjusted (min 1, max 100, default 50).
	params := map[string]interface{}{
		"feed":  feedAtURI,
		"limit": limit,
	}

	if p.cursor != "" {
		params["cursor"] = p.cursor
	}

	log.Printf("Fetching feed with params: %+v", params)

	var out types.FeedResponse

	// Call the Bluesky API using the xrpc client.
	err := client.Do(context.Background(), xrpc.Query, "json", feedMethod, params, nil, &out)
	if err != nil {
		log.Printf("Error fetching feed via xrpc: %v", err)
		return
	}

	log.Printf("Feed gotten")
	log.Printf("Sending to client")
	sendToClient(out)
}

func InitCronJobs() {
	log.Println("Starting Cron Jobs")
	c := cron.New()

	// Function 1: Run every 5 minutes at 0 minutes
	_, err := c.AddFunc("*/5 * * * *", func() {
		log.Println("Function 1 running every 5 minutes")
		feedAtURI := "at://did:plc:qiknc4t5rq7yngvz7g4aezq7/app.bsky.feed.generator/aaaejsyozb6iq"
		callFeed(FeedCallParameters{uri: feedAtURI, limit: 10, cursor: ""})
	})
	if err != nil {
		log.Println("Error scheduling Function 1:", err)
	}

	// Function 2: Run every 5 minutes at 1 minute mark
	_, err = c.AddFunc("1-59/5 * * * *", func() {
		log.Println("Function 2 running every 5 minutes, 1 minute apart")
	})
	if err != nil {
		log.Println("Error scheduling Function 2:", err)
	}

	// Function 3: Run every 5 minutes at 2 minutes mark
	_, err = c.AddFunc("2-59/5 * * * *", func() {
		log.Println("Function 3 running every 5 minutes, 2 minutes apart")
	})
	if err != nil {
		log.Println("Error scheduling Function 3:", err)
	}

	c.Start()
}
