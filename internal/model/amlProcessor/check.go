package amlcommon

type CheckRequest struct {
	Name              string `json:"name" validate:"required"`
	ReferenceID       string `json:"referenceId"`
	SubjectType       string `json:"subjectType,omitempty"`
	Score             string `json:"score,omitempty"`
	Nationality       string `json:"nationality,omitempty"`
	DOB               string `json:"dob,omitempty"`
	Gender            string `json:"gender,omitempty"`
	CountryLocation   string `json:"countryLocation,omitempty"`
	PlaceOfBirth      string `json:"placeOfBirth,omitempty"`
	RegisteredCountry string `json:"registeredCountry,omitempty"`
}

type CheckResponse struct {
	Code            string            `json:"code"`
	Message         string            `json:"message"`
	Extra           *any              `json:"extra"`
	TransactionID   string            `json:"transactionId"`
	PricingStrategy string            `json:"pricingStrategy"`
	Data            CheckResponseData `json:"data"`
}

type CheckResponseData struct {
	TransID string `json:"transId"`
	Status  string `json:"status"`
}
