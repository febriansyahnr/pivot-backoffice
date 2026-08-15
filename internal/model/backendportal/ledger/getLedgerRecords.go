package ledger_model

import (
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/orchestrator"
)

type GetLedgerTransactionRequest struct {
	AccountID     uuid.UUID
	Status        string
	ReferenceType string
	StartDate     time.Time
	EndDate       time.Time
}

type GetLedgerTransactionData struct {
	ReferenceID          string
	Debit                float64
	Credit               float64
	Type                 string
	Channel              string
	Status               string
	Remarks              string
	ReasonType           string
	ReasonDescription    string
	TransactionTimestamp time.Time
}

type GetLedgerTransactionDetail struct {
	ReferenceID          string    `db:"reference_id"`
	MerchantID           string    `db:"merchant_id"`
	AccountID            string    `db:"account_id"`
	Currency             string    `db:"currency"`
	Credit               float64   `db:"credit"`
	Debit                float64   `db:"debit"`
	Type                 string    `db:"type"`
	Channel              string    `db:"channel"`
	Status               string    `db:"status"`
	ReasonType           *string   `db:"reason_type"`
	ReasonDescription    *string   `db:"reason_description"`
	Remarks              string    `db:"remarks"`
	TransactionTimestamp time.Time `db:"transaction_timestamp"`
	CreatedAt            time.Time `db:"created_at"`
	UpdatedAt            time.Time `db:"updated_at"`
	Reference            *string   `db:"reference"`
}

type GetLedgerDetailResponse struct {
	ReferenceID          string    `json:"reference_id"`
	MerchantID           string    `json:"merchant_id"`
	AccountID            string    `json:"account_id"`
	Currency             string    `json:"currency"`
	Credit               float64   `json:"credit"`
	Debit                float64   `json:"debit"`
	Type                 string    `json:"type"`
	Channel              string    `json:"channel"`
	Status               string    `json:"status"`
	ReasonType           string    `json:"reason_type,omitempty"`
	ReasonDescription    string    `json:"reason_description,omitempty"`
	Remarks              string    `json:"remarks"`
	TransactionTimestamp time.Time `json:"transaction_timestamp"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
	Reference            *string   `json:"reference,omitempty"`
}

func ToGetLedgerDetailResponse(data *orchestratorModel.AccountTransaction) *GetLedgerDetailResponse {
	return &GetLedgerDetailResponse{
		ReferenceID:          data.ReferenceID,
		MerchantID:           data.MerchantID.String(),
		AccountID:            data.AccountID.String(),
		Currency:             data.Currency,
		Credit:               data.Credit,
		Debit:                data.Debit,
		Type:                 data.Type,
		Channel:              data.Channel,
		Status:               data.Status,
		ReasonType:           data.ReasonType.String,
		ReasonDescription:    data.ReasonDescription.String,
		Remarks:              data.Remarks,
		TransactionTimestamp: data.TransactionTimestamp,
		CreatedAt:            data.CreatedAt,
		UpdatedAt:            data.UpdatedAt,
		Reference:            &data.Reference,
	}
}

func (req *GetLedgerTransactionRequest) AdjustDateTime() error {
	if req.StartDate.IsZero() && req.EndDate.IsZero() {
		req.StartDate = time.Now().UTC()
		req.EndDate = time.Now().UTC().AddDate(0, 0, constant.DefaultDateRange)
	}
	if req.EndDate.IsZero() {
		req.EndDate = req.StartDate.AddDate(0, 0, constant.DefaultDateRange)
	}
	if req.StartDate.IsZero() {
		req.StartDate = req.EndDate.AddDate(0, 0, -constant.DefaultDateRange)
	}
	if !req.EndDate.After(req.StartDate) {
		return constant.ErrInvalidDateRange
	}

	return nil
}
