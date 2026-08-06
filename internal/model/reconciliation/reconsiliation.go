package reconciliation

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
)

// UUIDGenerator is a function type for generating UUIDs
type UUIDGenerator func() (uuid.UUID, error)

// defaultUUIDGenerator is the default implementation using uuid.NewV7
var defaultUUIDGenerator UUIDGenerator = uuid.NewV7

type Reconciliation struct {
	UUID         string `db:"uuid" json:"uuid"`
	OriginalName string `db:"original_name" json:"originalName"`
	// recon file path
	FilePath string `db:"file_path" json:"filePath"`
	// recon result file path
	ResultFilePath string `db:"result_file_path" json:"resultFilePath"`
	// PENDING, SUCCESS, FAILED
	Status          string         `db:"status" json:"status"`
	TransactionType string         `db:"transaction_type" json:"transactionType"`
	Reasons         sql.NullString `db:"reasons" json:"reasons"`
	CreatedBy       string         `db:"created_by" json:"createdBy"`
	CreatedAt       time.Time      `db:"created_at" json:"createdAt"`
	UpdatedAt       time.Time      `db:"updated_at" json:"updatedAt"`
}

type ReconciliationResponse struct {
	UUID            string    `json:"uuid"`
	OriginalName    string    `json:"originalName"`
	FilePath        string    `json:"filePath"`
	ResultFilePath  string    `json:"resultFilePath"`
	Status          string    `json:"status"`
	TransactionType string    `json:"transactionType"`
	Reasons         string    `json:"reasons"`
	CreatedBy       string    `json:"createdBy"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

func (r *Reconciliation) ToResponse() *ReconciliationResponse {
	return &ReconciliationResponse{
		UUID:            r.UUID,
		OriginalName:    r.OriginalName,
		FilePath:        r.FilePath,
		ResultFilePath:  r.ResultFilePath,
		Status:          r.Status,
		Reasons:         r.Reasons.String,
		TransactionType: r.TransactionType,
		CreatedBy:       r.CreatedBy,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
	}
}

func NewReconciliation(transactionType, createdBy, filePath string) (*Reconciliation, error) {
	id, err := defaultUUIDGenerator()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	return &Reconciliation{
		UUID:            id.String(),
		Status:          constant.StatusPending,
		FilePath:        filePath,
		CreatedBy:       createdBy,
		TransactionType: transactionType,
		ResultFilePath:  "",
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

func (r *Reconciliation) GetTransactionReferenceByTransactionType() string {
	switch r.TransactionType {
	case constant.TypeDisbursement:
		return constant.TypeDisbursement
	}

	// Type PAYMENT, REFUND, WITHDRAWAL (Manual) will be PAYMENT reference
	return constant.ReferencePayment
}
