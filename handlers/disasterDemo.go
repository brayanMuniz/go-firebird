package handlers

import (
	"encoding/csv"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
)

type Csv_type struct {
	Text       string `json:"text"`
	CreatedAt  string `json:"createdAt"`
	Prediction string `json:"prediction"`
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

	var disasterRecords []Csv_type
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

		d := Csv_type{
			Text:       rec[0],
			CreatedAt:  rec[1],
			Prediction: rec[2],
		}

		disasterRecords = append(disasterRecords, d)
	}

	// TODO: convert all the disaster records to FeedResponse type
	// replace Feed entry with the correct uri route. Make it map correctly
	// pass that onto processor.feed  logic, extract the logic so it does not save it

	c.JSON(http.StatusOK, gin.H{"uwu": disasterRecords})

}
