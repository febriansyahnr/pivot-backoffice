package creditcardCoreProcessorModel

import (
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/shopspring/decimal"
)

// AutoReconTrxRequest represents the request structure for credit card reconciliation
type AutoReconTrxRequest struct {
	Type         string          `json:"type"`
	ReferenceNo  string          `json:"reference_no"`
	ReferenceNo2 string          `json:"reference_no_2,omitempty"`
	TrxTimestamp *time.Time      `json:"trx_timestamp"`
	Amount       decimal.Decimal `json:"amount"`
	Bank         string          `json:"bank"`
	MerchantID   string          `json:"merchant_id"`
}

// AutoReconTrxResponse represents the response structure for credit card reconciliation
type AutoReconTrxResponse struct {
	UUID                 string              `json:"uuid"`
	Type                 string              `json:"type"`
	Status               string              `json:"status"`
	Code                 constant.TReconCode `json:"code,omitempty"`
	ProcessorReferenceID string              `json:"processor_reference_id,omitempty"`
}

// AutoReconResponse wraps the reconciliation response with standard API structure
type AutoReconResponse struct {
	Code    string               `json:"code"`
	Message string               `json:"message"`
	Data    AutoReconTrxResponse `json:"data"`
	Error   interface{}          `json:"error,omitempty"`
}
