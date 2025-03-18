package processor

import (
	"fmt"
	"go-firebird/db"
	"go-firebird/nlp"
	"go-firebird/types"
	"log"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
)

func ProcessLocationAvgSentiment(firestoreClient *firestore.Client, locationID string) error {
	// Helper function to append a formatted log message.
	// NOTE: this right here is my religion
	var logBuilder strings.Builder
	addLog := func(format string, args ...interface{}) {
		logBuilder.WriteString(fmt.Sprintf(format, args...))
		logBuilder.WriteString("\n")
	}

	start := "1970-01-01T00:00:00Z" // big bang of computers
	end := time.Now().UTC().Format(time.RFC3339)

	locationData, err := db.GetValidLocation(firestoreClient, locationID)
	if err != nil {
		addLog("Error fetching valid location: %v", err)
		log.Println(logBuilder.String())
		return err
	}

	addLog("Running avg sentiment on docId %v | Location name: %v", locationID, locationData.FormattedAddress)

	if len(locationData.AvgSentimentList) == 0 {
		addLog("Sentiment list is empty. Will INIT")
		skeets, err := db.GetSkeetsSubCollection(firestoreClient, locationID, start, end)
		if err != nil {
			addLog("Error fetching skeets subcollection: %v", err)
			log.Println(logBuilder.String())
			return err
		}
		addLog("Fetched %d skeets", len(skeets))

		avg := nlp.ComputeSimpleAverageSentiment(skeets)
		newSentiment := types.AvgLocationSentiment{
			TimeStamp:        end,
			SkeetsAmount:     len(skeets),
			AverageSentiment: avg,
		}

		addLog("TimeStamp is: %v", newSentiment.TimeStamp)
		addLog("Skeets amount is: %v", newSentiment.SkeetsAmount)
		addLog("Average is: %v", newSentiment.AverageSentiment)

		newLocationErr := db.InitLocationSentiment(firestoreClient, locationID, newSentiment)
		if newLocationErr != nil {
			addLog("Error adding new location sentiment: %v", newLocationErr)
			log.Println(logBuilder.String())
			return err
		}

	} else {
		addLog("Sentiment list is not empty")
		latestSentiment := locationData.AvgSentimentList[len(locationData.AvgSentimentList)-1]
		newSkeets, err := db.GetSkeetsSubCollection(firestoreClient, locationID, latestSentiment.TimeStamp, end)
		if err != nil {
			addLog("Failed to get skeets subcollection: %v", err)
			log.Println(logBuilder.String())
			return err
		}

		if len(newSkeets) > 0 {
			addLog("New skeets to average. Adding new average")
			newSkeetsAvg := nlp.ComputeSimpleAverageSentiment(newSkeets)
			newAvg := ((latestSentiment.AverageSentiment * float32(latestSentiment.SkeetsAmount)) + newSkeetsAvg) / ((float32(latestSentiment.SkeetsAmount)) + float32(len(newSkeets)))

			newSentiment := types.AvgLocationSentiment{
				TimeStamp:        end,
				SkeetsAmount:     latestSentiment.SkeetsAmount + len(newSkeets),
				AverageSentiment: newAvg,
			}
			addLog("TimeStamp is: %v", newSentiment.TimeStamp)
			addLog("Skeets amount is: %v", newSentiment.SkeetsAmount)
			addLog("Average is: %v", newSentiment.AverageSentiment)

			locationData.AvgSentimentList = append(locationData.AvgSentimentList, newSentiment)
			e := db.AddNewLocSentimentAvg(firestoreClient, locationID, locationData.AvgSentimentList)
			if e != nil {
				addLog("Failed to add new avgLocation: %v", e)
				log.Println(logBuilder.String())
				return err
			}
			addLog("Added new average to list")

		} else {
			addLog("No new skeets to average")
			addLog("Updating timestamp of latest sentiment")
			e := db.UpdateLocSentimentTimestampWithData(firestoreClient, locationData, locationID, end)
			if e != nil {
				addLog("Failed to update avgLocation field with new timestamp: %v", e)
				log.Println(logBuilder.String())
				return err
			}
		}
	}
	log.Println(logBuilder.String())
	return nil

}
