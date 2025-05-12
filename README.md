# Firebird Core (go-firebird)

**go-firebird** is the backend service for the Firebird project. It's responsible for 
ingesting data from the Bluesky social platform, processing skeets for sentiment and location, 
detecting potential disaster events (wildfires, earthquakes, hurricanes), 
summarizing them using LLMs, and storing all relevant information in Firestore. 
This service provides the analytical backbone for the Firebird UI.

## Key Features

*   **Bluesky Data Ingestion:** Fetches skeets from specific disaster-related feeds on Bluesky.
*   **Sentiment Analysis:** Utilizes Google Cloud Natural Language API for sentiment scoring of skeets.
*   **Location Processing:** Extracts and aggregates location data from skeets.
*   **Disaster Detection:** Implements logic to identify potential disaster events based on 
sentiment velocity, skeet volume, and keyword analysis.
*   **LLM Summarization:** Generates concise summaries of detected disasters using OpenAI's API.
*   **Firestore Integration:** Persists processed skeets, location aggregates, and confirmed disaster data.
*   **API Endpoints:** Provides API endpoints for data fetching, 
disaster detection triggers, and test data management.

## Tech Stack

*   **Language:** Go (Golang)
*   **Web Framework:** Gin
*   **Database:** Google Cloud Firestore
*   **NLP:** Google Cloud Natural Language API
*   **LLM:** OpenAI API
*   **Bluesky Interaction:** `github.com/bluesky-social/indigo/xrpc`
