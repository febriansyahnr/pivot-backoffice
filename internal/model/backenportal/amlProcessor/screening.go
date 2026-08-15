package amlcommon

const (
	NodeTypeAMLNameScreening = "Aml Name Screening"
	DetailStatusPending      = "PENDING"
	DetailStatusDismiss      = "DISMISS"
	DetailStatusOnMonitor    = "ON_MONITOR"
)

type ScreeningResponse struct {
	Status        string           `json:"status"`
	TransactionID string           `json:"transactionId"`
	ReferenceID   string           `json:"referenceId"`
	Result        *ScreeningResult `json:"result,omitempty"`
}

type ScreeningResult struct {
	ID            string                `json:"id"`
	CompletedAt   string                `json:"completedAt"`
	TransactionID string                `json:"transactionId"`
	Detail        []ScreeningDetailItem `json:"detail"`
	Summary       NodeSummary           `json:"summary"`
	MatchedCount  int                   `json:"matchedCount"`
	Attributes    ScreeningAttributes   `json:"attributes"`
}

type ScreeningDetailItem struct {
	NodeDetail
	Status string `json:"status"`
}

type ScreeningAttributes struct {
	DOB               string   `json:"dob"`
	Name              string   `json:"name"`
	Score             int      `json:"score"`
	Gender            string   `json:"gender"`
	EntityType        string   `json:"entityType"`
	HitCategory       []string `json:"hitCategory"`
	ReferenceID       string   `json:"referenceId"`
	PlaceOfBirth      string   `json:"placeOfBirth"`
	CountryLocation   string   `json:"countryLocation"`
	RegisteredCountry string   `json:"registeredCountry"`
}

type UpdateDetailStatusRequest struct {
	Name   string `json:"name" validate:"required"`
	DOB    string `json:"dob" validate:"required"`
	Status string `json:"status" validate:"required,oneof=PENDING DISMISS ON_MONITOR"`
}
