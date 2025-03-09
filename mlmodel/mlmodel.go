package mlmodel

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

type MLRequest map[string]string
type MLResponse map[string][]float64

const mlURL = "https://firebirdmodel-165032778338.us-central1.run.app/tweets/"

func CallModel(inputs MLRequest) (MLResponse, error) {
	payloadBytes, err := json.Marshal(inputs)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, mlURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("ML model returned status: " + resp.Status)
	}

	var mlResp MLResponse
	if err := json.NewDecoder(resp.Body).Decode(&mlResp); err != nil {
		return nil, err
	}

	return mlResp, nil
}
