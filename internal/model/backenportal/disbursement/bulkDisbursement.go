package disbursementModel

import "time"

type BulkDisbursement struct {
	UUID         string     `json:"uuid" db:"uuid" example:"b7b95825-6ec3-486a-a3cf-62d4ed815757"`
	MerchantID   string     `json:"merchantId" db:"merchant_id" example:"50f95963-373a-4a6e-a7c9-79b6cf891f9c"`
	File         string     `json:"file" db:"file" example:"https://storage.googleapis.com/bucket-name/object-path"`
	FileFailed   *string    `json:"fileFailed" db:"file_failed" example:"https://storage.googleapis.com/bucket-name/object-path"`
	FileRejected *string    `json:"fileRejected" db:"file_rejected" example:"https://storage.googleapis.com/bucket-name/object-path"`
	Status       string     `json:"status" db:"status" example:"DONE"`
	CreatedBy    *string    `json:"createdBy" db:"created_by" example:"John Doe"`
	CreatedAt    time.Time  `json:"createdAt" db:"created_at" example:"2021-01-01T00:00:00Z"`
	UpdatedAt    time.Time  `json:"updatedAt" db:"updated_at" example:"2021-01-01T00:00:00Z"`
	DeletedAt    *time.Time `json:"deletedAt" db:"deleted_at" example:"2021-01-01T00:00:00Z"`
}

type BulkDisbursementWithAggregate struct {
	UUID           string     `json:"uuid" db:"uuid" example:"b7b95825-6ec3-486a-a3cf-62d4ed815757"`
	MerchantID     string     `json:"merchantId" db:"merchant_id" example:"50f95963-373a-4a6e-a7c9-79b6cf891f9c"`
	File           string     `json:"file" db:"file" example:"https://storage.googleapis.com/bucket-name/object-path"`
	FileFailed     *string    `json:"fileFailed" db:"file_failed" example:"https://storage.googleapis.com/bucket-name/object-path"`
	FileRejected   *string    `json:"fileRejected" db:"file_rejected" example:"https://storage.googleapis.com/bucket-name/object-path"`
	Status         string     `json:"status" db:"status" example:"DONE"`
	TotalAmount    float64    `json:"totalAmount" db:"total_amount" example:"1000000.00"`
	TotalTrx       int        `json:"totalTrx" db:"total_trx" example:"10"`
	TotalApproved  int        `json:"totalApproved" db:"total_approved" example:"8"`
	TotalRejected  int        `json:"totalRejected" db:"total_rejected" example:"2"`
	TotalSuccess   int        `json:"totalSuccess" db:"total_success" example:"6"`
	TotalFailed    int        `json:"totalFailed" db:"total_failed" example:"2"`
	TotalCancelled int        `json:"totalCancelled" db:"total_cancelled" example:"2"`
	TotalPending   int        `json:"totalPending" db:"total_pending" example:"3"`
	CreatedBy      *string    `json:"createdBy" db:"created_by" example:"John Doe"`
	CreatedAt      time.Time  `json:"createdAt" db:"created_at" example:"2021-01-01T00:00:00Z"`
	UpdatedAt      time.Time  `json:"updatedAt" db:"updated_at" example:"2021-01-01T00:00:00Z"`
	DeletedAt      *time.Time `json:"deletedAt" db:"deleted_at" example:"2021-01-01T00:00:00Z"`
}

type BulkDisbursementDetail struct {
	UUID           string     `json:"uuid" db:"uuid" example:"b7b95825-6ec3-486a-a3cf-62d4ed815757"`
	MerchantID     string     `json:"merchantId" db:"merchant_id" example:"50f95963-373a-4a6e-a7c9-79b6cf891f9c"`
	Status         string     `json:"status" db:"status" example:"DONE"`
	TotalAmount    float64    `json:"totalAmount" db:"total_amount" example:"1000000.00"`
	TotalTrx       int        `json:"totalTrx" db:"total_trx" example:"10"`
	TotalApproved  int        `json:"totalApproved" db:"total_approved" example:"8"`
	TotalRejected  int        `json:"totalRejected" db:"total_rejected" example:"2"`
	TotalSuccess   float64    `json:"totalSuccess" db:"total_success" example:"6"`
	TotalFailed    float64    `json:"totalFailed" db:"total_failed" example:"2"`
	TotalCancelled float64    `json:"totalCancelled" db:"total_cancelled" example:"2"`
	TotalPending   float64    `json:"totalPending" db:"total_pending" example:"3"`
	CreatedBy      *string    `json:"createdBy" db:"created_by" example:"John Doe"`
	CreatedAt      time.Time  `json:"createdAt" db:"created_at" example:"2021-01-01T00:00:00Z"`
	UpdatedAt      time.Time  `json:"updatedAt" db:"updated_at" example:"2021-01-01T00:00:00Z"`
	DeletedAt      *time.Time `json:"deletedAt" db:"deleted_at" example:"2021-01-01T00:00:00Z"`
}
