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

func ProcessLocationAvgSentiment(firestoreClient *firestore.Client, locationID string, locationData types.LocationData) error {
	// NOTE: this right here is my religion
	var logBuilder strings.Builder
	addLog := func(format string, args ...interface{}) {
		logBuilder.WriteString(fmt.Sprintf(format, args...))
		logBuilder.WriteString("\n")
	}

	start := "1970-01-01T00:00:00Z" // big bang of computers
	end := time.Now().UTC().Format(time.RFC3339)

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

		// compute the total count and add it to the field
		dCount := types.DisasterCount{}
		for _, v := range skeets {
			switch GetCategory(v.SkeetData) {
			case types.Wildfire:
				dCount.FireCount += 1
			case types.Earthquake:
				dCount.EarthquakeCount += 1
			case types.Hurricane:
				dCount.HurricaneCount += 1
			case types.NonDisaster:
				dCount.NonDisasterCount += 1
			}
		}

		addLog("FireCount: %v", dCount.FireCount)
		addLog("EarthquakeCount: %v", dCount.EarthquakeCount)
		addLog("HurricaneCount: %v", dCount.HurricaneCount)
		addLog("NonDisasterCount: %v", dCount.NonDisasterCount)

		avg := nlp.ComputeSimpleAverageSentiment(skeets)
		newSentiment := types.AvgLocationSentiment{
			TimeStamp:        end,
			SkeetsAmount:     len(skeets),
			AverageSentiment: avg,
			DisasterCount:    dCount,
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

		// Update latestSkeetsAmount after successful init
		errUpdate := db.UpdateLocationFields(firestoreClient, locationID, map[string]interface{}{
			"latestSkeetsAmount": newSentiment.SkeetsAmount,
		})
		if errUpdate != nil {
			addLog("Warning: Failed to update latestSkeetsAmount after init: %v", errUpdate)
		} else {
			addLog("Updated latestSkeetsAmount after init.")
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

		// since this is a new field, you would have to check if the latestSentiment has the newfield, if it does not, fetch all the skeets.
		newDisasterCount := latestSentiment.DisasterCount
		if newDisasterCount.FireCount == 0 && newDisasterCount.EarthquakeCount == 0 && newDisasterCount.HurricaneCount == 0 && newDisasterCount.NonDisasterCount == 0 {
			addLog("This location does not have count field. Fetching all skeets. ")
			allSkeets, err := db.GetSkeetsSubCollection(firestoreClient, locationID, start, end)
			if err != nil {
				addLog("Failed to get all skeets subcollection: %v", err)
				return err
			}

			// compute the total count and add it to the field
			dCount := types.DisasterCount{}
			for _, v := range allSkeets {
				switch GetCategory(v.SkeetData) {
				case types.Wildfire:
					dCount.FireCount += 1
				case types.Earthquake:
					dCount.EarthquakeCount += 1
				case types.Hurricane:
					dCount.HurricaneCount += 1
				case types.NonDisaster:
					dCount.NonDisasterCount += 1
				}
			}

			newDisasterCount = dCount

			// update the document
			latestSentiment.DisasterCount = newDisasterCount
			locationData.AvgSentimentList[len(locationData.AvgSentimentList)-1] = latestSentiment
			e := db.UpdateLocationSentimentList(firestoreClient, locationID, locationData.AvgSentimentList)
			if e != nil {
				addLog("Failed to update location sentiment list with new count: %v", e)
				log.Println(logBuilder.String())
				return e
			}
			addLog("Successfully updated latest sentiment with backfilled counts.")

		} else {
			addLog("This location does have count field. Caculating new count")
			for _, v := range newSkeets {
				switch GetCategory(v.SkeetData) {
				case types.Wildfire:
					newDisasterCount.FireCount += 1
				case types.Earthquake:
					newDisasterCount.EarthquakeCount += 1
				case types.Hurricane:
					newDisasterCount.HurricaneCount += 1
				case types.NonDisaster:
					newDisasterCount.NonDisasterCount += 1
				}
			}
		}

		addLog("FireCount: %v", newDisasterCount.FireCount)
		addLog("EarthquakeCount: %v", newDisasterCount.EarthquakeCount)
		addLog("HurricaneCount: %v", newDisasterCount.HurricaneCount)
		addLog("NonDisasterCount: %v", newDisasterCount.NonDisasterCount)

		if len(newSkeets) > 0 {
			addLog("New skeets! Adding new entry")
			newSkeetsAvg := nlp.ComputeSimpleAverageSentiment(newSkeets)
			newAvg := ((latestSentiment.AverageSentiment * float32(latestSentiment.SkeetsAmount)) + newSkeetsAvg) / ((float32(latestSentiment.SkeetsAmount)) + float32(len(newSkeets)))

			newSentiment := types.AvgLocationSentiment{
				TimeStamp:        end,
				SkeetsAmount:     latestSentiment.SkeetsAmount + len(newSkeets),
				AverageSentiment: newAvg,
				DisasterCount:    newDisasterCount,
			}
			totalDisasterCount := newDisasterCount.FireCount + newDisasterCount.HurricaneCount + newDisasterCount.EarthquakeCount + newDisasterCount.NonDisasterCount
			if newSentiment.SkeetsAmount < (totalDisasterCount) {
				newSentiment.SkeetsAmount = totalDisasterCount
				addLog("Skeets Amount is behind totalDisaster. Updating")
			}
			addLog("TimeStamp is: %v", newSentiment.TimeStamp)
			addLog("Skeets amount is: %v", newSentiment.SkeetsAmount)
			addLog("Average is: %v", newSentiment.AverageSentiment)
			addLog("Total Disaster count is: %v", totalDisasterCount)

			locationData.AvgSentimentList = append(locationData.AvgSentimentList, newSentiment)
			e := db.AddNewLocSentimentAvg(firestoreClient, locationID, locationData.AvgSentimentList)
			if e != nil {
				addLog("Failed to add new avgLocation: %v", e)
				log.Println(logBuilder.String())
				return err
			}
			addLog("Added new average to list")

			errUpdate := db.UpdateLocationFields(firestoreClient, locationID, map[string]interface{}{
				"latestSkeetsAmount": newSentiment.SkeetsAmount,
			})
			if errUpdate != nil {
				addLog("Failed to update location fields: %v", errUpdate)
				log.Println(logBuilder.String())
				return errUpdate
			}

		} else {
			addLog("No new skeets")
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
