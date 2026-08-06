package bigquery

type Config struct {
	ProjectID           string
	Location            string // e.g., "asia-southeast2"
	QueryTimeoutSeconds int
	MaxRetries          int
}
