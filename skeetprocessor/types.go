package skeetprocessor

import "go-firebird/nlp"

type SaveSkeetResult struct {
	SavedSkeetID         string        `json:"savedSkeetId"`
	Content              string        `json:"content"`
	NewLocationNames     []string      `json:"newLocationNames"`
	ProcessedEntityCount int           `json:"processedEntityCount"`
	Classification       []float64     `json:"classification"`
	Sentiment            nlp.Sentiment `json:"sentiment"`
	AlreadyExist         bool          `json:"alreadyExist"`
	ErrorSaving          bool          `json:"errorSaving"`
}
