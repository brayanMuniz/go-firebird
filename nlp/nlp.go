package nlp

import (
	language "cloud.google.com/go/language/apiv2"
	"cloud.google.com/go/language/apiv2/languagepb"
	"context"
	"encoding/base64"
	"fmt"
	"google.golang.org/api/option"
	"log"
	"os"
	"sync"
)

// languageClient a singleton languageClient instance.
var (
	languageClient *language.Client
	clientOnce     sync.Once
)

// Sentiment
type Sentiment struct {
	Magnitude float32 `json:"magnitude"`
	Score     float32 `json:"score"`
}

// Entity represents a named entity detected in the text.
type Entity struct {
	Name     string            `json:"name"`
	Type     string            `json:"type"`
	Metadata map[string]string `json:"metadata"`
	Mentions []EntityMention   `json:"mentions"`
}

// EntityMention holds details about an entity mention.
type EntityMention struct {
	Content     string  `json:"content"`
	BeginOffset int32   `json:"begin_offset"`
	Probability float32 `json:"probability"`
}

func AnalyzeSentiment(client *language.Client, text string) (Sentiment, error) {
	var sentiment Sentiment
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
func AnalyzeEntities(client *language.Client, text string) ([]Entity, error) {
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

	var entities []Entity
	for _, e := range resp.Entities {
		var mentions []EntityMention
		for _, m := range e.Mentions {
			mentions = append(mentions, EntityMention{
				Content:     m.Text.Content,
				BeginOffset: m.Text.BeginOffset,
				Probability: m.Probability,
			})
		}
		md := make(map[string]string)
		for k, v := range e.Metadata {
			md[k] = v
		}
		entities = append(entities, Entity{
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
