package paymentCaptureModel

import "time"

type CaptureHistoryResponse struct {
	ID             string    `json:"captureId"`
	Currency       string    `json:"currency"`
	CapturedAmount float64   `json:"capturedAmount"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
}
