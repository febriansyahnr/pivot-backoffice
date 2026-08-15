package statusHistoryModel

type StatusHistoryMetadata struct {
	Label          string `json:"label"`
	Actor          string `json:"actor"`
	Description    string `json:"description"`
	Recommendation string `json:"recommendation,omitempty"`
}
