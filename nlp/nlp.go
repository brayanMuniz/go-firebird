package nlp

import (
	"context"
	"encoding/base64"
	"fmt"
	"go-firebird/types"
	"log"
	"os"
	"sync"

	language "cloud.google.com/go/language/apiv2"
	"cloud.google.com/go/language/apiv2/languagepb"
	"google.golang.org/api/option"
)

// languageClient a singleton languageClient instance.
var (
	languageClient *language.Client
	clientOnce     sync.Once
)

func AnalyzeSentiment(client *language.Client, text string) (types.Sentiment, error) {
	var sentiment types.Sentiment
	ctx := context.Background()
	req := &languagepb.AnalyzeSentimentRequest{
		Document: &languagepb.Document{
			Source: &languagepb.Document_Content{
				Content: text,
			},
			Type: languagepb.Document_PLAIN_TEXT,
		},
		EncodingType: languagepb.EncodingType_UTF8,
	}

	resp, err := client.AnalyzeSentiment(ctx, req)
	if err != nil {
		return sentiment, fmt.Errorf("AnalyzeEntities Requesterror: %w", err)
	}

	sentiment.Score = resp.DocumentSentiment.Score
	sentiment.Magnitude = resp.DocumentSentiment.Magnitude

	return sentiment, nil
}

// sends text to the Cloud Natural Language API to extract named entities
// and returns a slice of Entity structs along with any error encountered
func AnalyzeEntities(client *language.Client, text string) ([]types.Entity, error) {
	ctx := context.Background()
	req := &languagepb.AnalyzeEntitiesRequest{
		Document: &languagepb.Document{
			Source: &languagepb.Document_Content{
				Content: text,
			},
			Type: languagepb.Document_PLAIN_TEXT,
		},
		EncodingType: languagepb.EncodingType_UTF8,
	}

	resp, err := client.AnalyzeEntities(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("AnalyzeEntities error: %w", err)
	}

	var entities []types.Entity
	for _, e := range resp.Entities {
		var mentions []types.EntityMention
		for _, m := range e.Mentions {
			mentions = append(mentions, types.EntityMention{
				Content:     m.Text.Content,
				BeginOffset: m.Text.BeginOffset,
				Probability: m.Probability,
			})
		}
		md := make(map[string]string)
		for k, v := range e.Metadata {
			md[k] = v
		}
		entities = append(entities, types.Entity{
			Name:     e.Name,
			Type:     e.Type.String(),
			Metadata: md,
			Mentions: mentions,
		})
	}
	return entities, nil
}

// initializes and returns a language client.
func InitLanguageClient() (*language.Client, error) {
	var err error

	clientOnce.Do(func() {
		// Decode credentials
		encodedCreds := os.Getenv("NATURAL_LANGUAGE_CREDENTIALS")
		creds, err := base64.StdEncoding.DecodeString(encodedCreds)
		if err != nil {
			log.Fatalf("Failed to decode Natural language credentials credentials: %v", err)

		}

		// Create the Natural Language API client using the decoded credentials
		opt := option.WithCredentialsJSON(creds)
		languageClient, err = language.NewClient(context.Background(), opt)
		if err != nil {
			log.Fatalf("Failed to create Natural Language client: %v", err)
		}

	})

	return languageClient, err
}

func CloseLanguageClient() {
	if languageClient != nil {
		languageClient.Close()
	}
}

func ComputeSimpleAverageSentiment(skeets []types.SkeetSubDoc) float32 {
	var totalScore float32
	count := 0

	for _, skeet := range skeets {
		totalScore += skeet.SkeetData.Sentiment.Score
		count++
	}

	if count == 0 {
		return 0
	}

	return totalScore / float32(count)
}
