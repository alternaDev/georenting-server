package gcm

import "os"

var (
	googleAPIKey    = os.Getenv("GOOGLE_API_KEY")
	googleProjectID = os.Getenv("GOOGLE_PROJECT_ID")
)
