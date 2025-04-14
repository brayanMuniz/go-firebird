package summarization

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"go-firebird/db"
	"go-firebird/types"
	"log"
	"strings"
	"sync"
)

const maxSkeetsForSummary = 69
const maxPromptLength = 15000 // Rough character limit for prompt

// fetches skeets for each disaster and calls OpenAI for summarization.
// It modifies the input slice directly.
func GenerateSummaries(
	ctx context.Context,
	disasters []types.DisasterData,
	firestoreClient *firestore.Client,
	openaiClient *openai.Client,
) error {
	log.Printf("Starting summary generation for %d disasters...", len(disasters))

	var wg sync.WaitGroup

	for i := range disasters {
		wg.Add(1)
		go func(disasterIndex int) {
			defer wg.Done()
			disaster := &disasters[disasterIndex]

			log.Printf("Fetching skeets for disaster ID %s (%s)", disaster.ID, disaster.DisasterType)

			// 1. Fetch relevant skeets for all locations in the cluster
			combinedSkeetText, err := fetchSkeetsForDisaster(ctx, disaster, firestoreClient)
			if err != nil {
				log.Printf("Error fetching skeets for disaster %s: %v. Skipping summary.", disaster.ID, err)
				return
			}

			if combinedSkeetText == "" {
				log.Printf("No relevant skeets found for disaster %s within the timeframe. Skipping summary.", disaster.ID)
				return
			}

			// 2. Call OpenAI for summary
			log.Printf("Requesting summary from OpenAI for disaster %s...", disaster.ID)
			summary, err := callOpenAISummary(ctx, combinedSkeetText, disaster.DisasterType, openaiClient)
			if err != nil {
				log.Printf("Error getting summary from OpenAI for disaster %s: %v. Skipping summary.", disaster.ID, err)
				return
			}

			// 3. Update the disaster object
			log.Printf("Received summary for disaster %s.", disaster.ID)
			disaster.Summary = summary

		}(i) // Pass index to the goroutine
	}

	wg.Wait() // Wait for all goroutines to finish

	log.Println("Summary generation finished.")
	return nil
}

// fetchSkeetsForDisaster retrieves skeets for all locations within a disaster's timeframe.
func fetchSkeetsForDisaster(
	ctx context.Context,
	disaster *types.DisasterData,
	firestoreClient *firestore.Client,
) (string, error) {
	var allSkeetsContent []string
	totalSkeetsFetched := 0

	// Use ReportedDate and LastUpdate from the disaster object
	startDate := disaster.ReportedDate
	endDate := disaster.LastUpdate

	if startDate == "" || endDate == "" {
		return "", fmt.Errorf("disaster %s is missing ReportedDate or LastUpdate", disaster.ID)
	}

	for _, locID := range disaster.LocationIDs {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
			// Continue fetching
		}

		// Limit skeets per location to avoid overwhelming the summary
		if totalSkeetsFetched >= maxSkeetsForSummary {
			log.Printf("Reached max skeet limit (%d) for disaster %s summary.", maxSkeetsForSummary, disaster.ID)
			break
		}

		skeets, err := db.GetSkeetsSubCollection(firestoreClient, locID, startDate, endDate)
		if err != nil {
			log.Printf("Warning: Failed to get skeets for location %s in disaster %s: %v", locID, disaster.ID, err)
			continue
		}

		for _, skeetDoc := range skeets {
			if totalSkeetsFetched < maxSkeetsForSummary {
				// Ensure content exists and is not empty
				if skeetDoc.SkeetData.Content != "" {
					content := skeetDoc.SkeetData.Content
					allSkeetsContent = append(allSkeetsContent, content)
					totalSkeetsFetched++
				}
			} else {
				break // Stop adding skeets from this location if limit reached
			}
		}
	}

	if len(allSkeetsContent) == 0 {
		return "", nil // No skeets found
	}

	combined := strings.Join(allSkeetsContent, "\n---\n")

	if len(combined) > maxPromptLength {
		log.Printf("Warning: Combined skeet text for disaster %s exceeds max length (%d), truncating.", disaster.ID, maxPromptLength)
		combined = combined[:maxPromptLength]
	}

	return combined, nil
}

// callOpenAISummary sends text to OpenAI and requests a summary.
func callOpenAISummary(
	ctx context.Context,
	skeetText string,
	disasterType types.Category,
	client *openai.Client,
) (string, error) {
	prompt := fmt.Sprintf("Summarize the following collection of social media posts related to a potential %s event. Focus on the key impacts, locations mentioned, and overall situation described. If a tweet feels incongruent to the disaster type or location, disregard the tweet from the summary. Provide a concise summary (2-3 sentences maximum):\n\n---\n%s\n---\n\nSummary:", disasterType, skeetText)

	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4oMini,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are an assistant that summarizes social media posts about potential disaster events concisely.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			MaxTokens:   150,
			N:           1,
			Stop:        nil, // Let the model decide when to stop (fuck it we [BALL])
			Temperature: 0.5, // Lower temperature for more focused summary
		},
	)

	if err != nil {
		return "", fmt.Errorf("openai chat completion error: %w", err)
	}

	if len(resp.Choices) == 0 || resp.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("openai returned empty response or choices")
	}

	// Return the content of the first choice
	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}
