package processor

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"go-firebird/db"
	"go-firebird/mlmodel"
	"go-firebird/nlp"
	"go-firebird/types"
	"log"
	"sync"

	"cloud.google.com/go/firestore"
	language "cloud.google.com/go/language/apiv2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HashString hashes a given string using SHA-256.
func HashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func SaveFeed(out types.FeedResponse, firestoreClient *firestore.Client, nlpClient *language.Client) []types.SaveSkeetResult {
	resultsChan := make(chan types.SaveSkeetResult, len(out.Feed))
	var wg sync.WaitGroup

	for _, v := range out.Feed {
		if v.Post.URI != "" {
			wg.Add(1)
			feedItem := v // capture variable for goroutine
			go func() {
				defer wg.Done()
				newSkeet := types.Skeet{
					Avatar:      feedItem.Post.Author.Avatar,
					Content:     feedItem.Post.Record.Text,
					Handle:      feedItem.Post.Author.Handle,
					DisplayName: feedItem.Post.Author.DisplayName,
					UID:         feedItem.Post.URI,
					Timestamp:   feedItem.Post.Record.CreatedAt,
				}
				savedSkeetResult, err := SaveSkeet(newSkeet, firestoreClient, nlpClient)
				if err != nil {
					savedSkeetResult = types.SaveSkeetResult{
						SavedSkeetID:         feedItem.Post.URI,
						NewLocationNames:     nil,
						ProcessedEntityCount: 0,
						Classification:       nil,
						Sentiment:            types.Sentiment{},
						AlreadyExist:         false,
						Content:              feedItem.Post.Record.Text,
						ErrorSaving:          true,
					}
				}
				resultsChan <- savedSkeetResult
			}()
		}
	}

	wg.Wait()
	close(resultsChan)

	resultsList := make([]types.SaveSkeetResult, 0, len(out.Feed))
	for result := range resultsChan {
		resultsList = append(resultsList, result)
	}

	return resultsList

}

func SaveSkeet(newSkeet types.Skeet, firestoreClient *firestore.Client, nlpClient *language.Client) (types.SaveSkeetResult, error) {
	ctx := context.Background()
	hashedSkeetID := db.HashString(newSkeet.UID)

	var result types.SaveSkeetResult
	result.SavedSkeetID = hashedSkeetID
	result.Content = newSkeet.Content
	result.AlreadyExist = false
	result.ErrorSaving = false

	// Check if the skeet already exists.
	_, err := firestoreClient.Collection("skeets").Doc(hashedSkeetID).Get(ctx)
	if err == nil {
		fmt.Printf("Doc already exists for: %s. Hash: %s\n", newSkeet.UID, hashedSkeetID)
		result.AlreadyExist = true
		result.SavedSkeetID = hashedSkeetID
		return result, nil
	} else if status.Code(err) != codes.NotFound {
		result.ErrorSaving = true
		return result, err
	}

	// Prepare the ML input.
	mlInputs := mlmodel.MLRequest{
		newSkeet.UID: newSkeet.Content,
	}

	// Run ML model call, entity extraction, and sentiment analysis concurrently.
	var (
		classification         []float64
		nlpEntities            []types.Entity
		sentiment              types.Sentiment
		mlErr, nlpErr, sentErr error
	)
	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		mlResp, err := mlmodel.CallModel(mlInputs)
		if err != nil {
			mlErr = err
			return
		}
		classification = mlResp[newSkeet.UID]
	}()

	go func() {
		defer wg.Done()
		var err error
		nlpEntities, err = nlp.AnalyzeEntities(nlpClient, newSkeet.Content)
		if err != nil {
			log.Printf("Error analyzing entities: %v", err)
			nlpEntities = []types.Entity{}
			nlpErr = err
		}
	}()

	go func() {
		defer wg.Done()
		var err error
		sentiment, err = nlp.AnalyzeSentiment(nlpClient, newSkeet.Content)
		if err != nil {
			log.Printf("Error analyzing sentiment: %v", err)
			sentErr = err
		}
	}()

	wg.Wait()
	_ = nlpErr
	_ = sentErr

	if mlErr != nil {
		fmt.Printf("Error processing ML classification for hashedSkeetId: %s", hashedSkeetID)
	}

	result.Classification = classification
	result.Sentiment = sentiment

	data := types.SaveCompleteSkeetType{
		NewSkeet:       newSkeet,
		Classification: classification,
		Entities:       nlpEntities,
		Sentiment:      sentiment,
	}

	newLocations, err := db.SaveCompleteSkeet(firestoreClient, data)
	if err != nil {
		result.ErrorSaving = true
		return result, err
	}
	result.NewLocationNames = newLocations

	count := 0
	for _, entity := range data.Entities {
		if entity.Type == "LOCATION" || entity.Type == "ADDRESS" {
			count++
		}
	}
	result.ProcessedEntityCount = count

	var geoWg sync.WaitGroup
	for _, locationName := range newLocations {
		geoWg.Add(1)
		go func(loc string) {
			defer geoWg.Done()
			db.UpdateLocationGeocoding(firestoreClient, loc)
		}(locationName)
	}
	geoWg.Wait()

	return result, nil
}
