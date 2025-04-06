package handlers

import (
	"encoding/csv"
	"fmt"
	"go-firebird/types"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type CsvRecord struct {
	Text       string `json:"text"`
	CreatedAt  string `json:"createdAt"`
	Prediction string `json:"prediction"`
}

// --- Placeholders ---
const (
	placeholderDid        = "did:placeholder:12345abcdef"
	placeholderHandle     = "placeholder.bsky.social"
	placeholderAvatar     = ""
	placeholderRecordType = "app.bsky.feed.post"
	placeholderLabelSrc   = placeholderDid
	fireURI               = "at://did:plc:qiknc4t5rq7yngvz7g4aezq7/app.bsky.feed.generator/aaaejsyozb6iq"
	earthQuakeURI         = "at://did:plc:qiknc4t5rq7yngvz7g4aezq7/app.bsky.feed.generator/aaaejxlobe474"
	hurricaneURI          = "at://did:plc:qiknc4t5rq7yngvz7g4aezq7/app.bsky.feed.generator/aaaejwgffwqky"
)

var placeholderAuthor = types.Author{
	DID:         placeholderDid,
	Handle:      placeholderHandle,
	DisplayName: "Placeholder Account",
	Avatar:      placeholderAvatar,
	CreatedAt:   time.Now().Format(time.RFC3339),
	Labels:      []types.Label{},
}

// Made an LLM write this
func convertToFeedResponse(records []CsvRecord) types.FeedResponse {
	feedEntries := make([]types.FeedEntry, 0, len(records))

	for i, rec := range records {
		// Generate unique-ish placeholders for URI and CID for this demo
		// In a real system, these would come from the database or be generated properly
		postUri := fireURI                                // default
		postCid := fmt.Sprintf("bafyreiplaceholder%d", i) // Example CID format
		if rec.Prediction == "earthquake" {
			postUri = earthQuakeURI
		}
		if rec.Prediction == "hurricane" {
			postUri = hurricaneURI
		}

		// Create the Label based on the prediction
		postLabels := []types.Label{
			{
				Src: placeholderLabelSrc,                         // Who created the label
				URI: postUri,                                     // Label applies to this post URI
				Val: rec.Prediction,                              // The actual prediction value
				CTS: rec.CreatedAt,                               // Timestamp for the label (use post creation time)
				CID: fmt.Sprintf("bafyreiplaceholderlabel%d", i), // Placeholder label CID
			},
		}

		// Create the main Record part of the post
		postRecord := types.Record{
			Type:      placeholderRecordType,
			CreatedAt: rec.CreatedAt,  // Use the timestamp from the CSV
			Text:      rec.Text,       // Use the text from the CSV
			Langs:     []string{"en"}, // Assume English for now
			Embed:     nil,            // No embeds from CSV
			Facets:    nil,            // No facets parsed from CSV
		}

		// Assemble the Post
		post := types.Post{
			Author:      placeholderAuthor, // Use the predefined placeholder author
			CID:         postCid,
			Record:      postRecord,
			URI:         postUri,
			IndexedAt:   rec.CreatedAt, // Use CSV creation time as indexed time
			Labels:      postLabels,
			LikeCount:   0, // Default counts
			ReplyCount:  0,
			RepostCount: 0,
			QuoteCount:  0,
			Embed:       nil, // No top-level embed
		}

		// Create the FeedEntry
		feedEntry := types.FeedEntry{
			Post: post,
		}

		feedEntries = append(feedEntries, feedEntry)
	}

	// Assemble the final FeedResponse
	response := types.FeedResponse{
		Feed:   feedEntries,
		Cursor: "", // No real cursor for this demo data
	}

	return response
}

func AddDisasterDemoData(c *gin.Context) {
	fmt.Println("Making Data Good")

	// Read csv file
	f, err := os.Open("./demo_data.csv")
	if err != nil {
		fmt.Println("Error reading file:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}
	defer f.Close()

	var disasterRecords []CsvRecord
	csvReader := csv.NewReader(f)
	rec, err := csvReader.Read()

	// skip the header in the csv file
	if err == io.EOF {
		fmt.Printf("end of file. This should NOT happen")
	}
	if err != nil {
		fmt.Printf("%+v\n", rec)
	}

	for {
		rec, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("%+v\n", rec)
		}

		if len(rec) != 3 {
			fmt.Println("Skipping record with incorrect number of fields: expected 3, got %d. Record: %v", len(rec), rec)
			continue // skip bad
		}

		d := CsvRecord{
			Text:       rec[0],
			CreatedAt:  rec[1],
			Prediction: rec[2],
		}

		disasterRecords = append(disasterRecords, d)
	}

	feedData := convertToFeedResponse(disasterRecords)

	// TODO: extract processor.feed  logic, extract the logic so it does NOT save it

	c.JSON(http.StatusOK, gin.H{"uwu": feedData})

}
