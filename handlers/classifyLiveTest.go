package handlers

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/bluesky-social/indigo/xrpc"
	"github.com/gin-gonic/gin"

	"go-firebird/mlmodel"
	"go-firebird/types"
)

type PostClassification struct {
	UID            string    `json:"uid"`
	Content        string    `json:"content"`
	Classification []float64 `json:"classification"`
}

func ClassifyLiveTest(c *gin.Context) {
	client := &xrpc.Client{
		Client:    &http.Client{Timeout: 10 * time.Second},
		Host:      "https://public.api.bsky.app",
		UserAgent: nil,
	}
	const feedMethod = "app.bsky.feed.getFeed"

	// fire feed
	feedAtURI := "at://did:plc:qiknc4t5rq7yngvz7g4aezq7/app.bsky.feed.generator/aaaejsyozb6iq"

	cursor := c.Query("cursor")
	if cursor != "" {
		cursor = strings.ReplaceAll(cursor, " ", "+")
	}

	// Gets 5 from fire
	params := map[string]interface{}{
		"feed":   feedAtURI,
		"limit":  5,
		"cursor": cursor,
	}

	var out types.FeedResponse
	if err := client.Do(context.Background(), xrpc.Query, "json", feedMethod, params, nil, &out); err != nil {
		log.Printf("Error fetching feed via xrpc: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(out.Feed) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No feed items returned"})
		return
	}

	mlInputs := mlmodel.MLRequest{}
	var posts []PostClassification

	for _, feedItem := range out.Feed {
		if feedItem.Post.URI == "" {
			continue
		}
		mlInputs[feedItem.Post.URI] = feedItem.Post.Record.Text
		posts = append(posts, PostClassification{
			UID:     feedItem.Post.URI,
			Content: feedItem.Post.Record.Text,
		})
	}

	// Call the ML model
	mlResp, err := mlmodel.CallModel(mlInputs)
	if err != nil {
		log.Printf("Error calling ML model: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to contact ML model"})
		return
	}

	// Now, update each post with its classification result.
	for i, post := range posts {
		if classification, ok := mlResp[post.UID]; ok {
			posts[i].Classification = classification
		} else {
			posts[i].Classification = []float64{}
		}
	}

	// Return JSON as an array of posts.
	c.JSON(http.StatusOK, posts)
}
