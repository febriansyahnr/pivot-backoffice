package unifiedPaymentModel

import (
	"mime/multipart"
	"time"
)

type UploadProofOfPaymentRequest struct {
	PaymentID         string
	MerchantID        string
	ProofOfPayment    *multipart.FileHeader
	Reason            string
	FileExtension     string
	OriginalFileName  string
}

type UploadProofOfPaymentResponse struct {
	PaymentID           string    `json:"paymentId"`
	Status              string    `json:"status"`
	InvestigationStatus string    `json:"investigationStatus"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

type InvestigationPoPMetadata struct {
	Bucket        string `json:"bucket"`
	Path          string `json:"path"`
	MerchantNotes string `json:"merchantNotes"`
}
