package disbursementModel

import (
	"time"
)

type GetBulkDisbursementFilterRequest struct {
	MerchantID     string     `json:"-"`
	StartCreatedAt *time.Time `json:"startCreatedAt"`
	EndCreatedAt   *time.Time `json:"endCreatedAt"`
	Status         string     `json:"status"`
	Sort           string     `json:"sort"`
	SortBy         string     `json:"sortBy"`
	ReferenceID    string     `json:"referenceId"`
}

type CreateBulkDisbursementRequest struct {
	MerchantID string `json:"merchantId"`
	File       string `json:"file"`
	Status     string `json:"status"`
	CreatedBy  string `json:"createdBy"`
}

type BatchCreateDisbursementRequest struct {
	BulkID       string                `json:"bulkId"`
	MerchantID   string                `json:"merchantId"`
	MerchantName string                `json:"merchantName"`
	CreatedBy    string                `json:"createdBy"`
	CreatedFrom  string                `json:"createdFrom"`
	TotalTrx     int                   `json:"totalTrx"`
	AutoApprove  bool                  `json:"autoApprove"`
	Data         []CreateSingleRequest `json:"data"`
}

type BatchProcessDisbursementRequest struct {
	BulkID          string   `json:"bulkId"`
	DisbursementIDs []string `json:"disbursementIDs"`
}
