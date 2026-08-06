package snapCoreModel

import (
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/shopspring/decimal"
)

type TransactionType string

const TypeVA TransactionType = "VA"
const TypeQRIS TransactionType = "QRIS"
const TypeBankTransfer TransactionType = "BANK_TRANSFER"

type AutoReconTrxRequest struct {
	// transaction uuid
	UUID string `json:"uuid"`
	// type of request, VA or QRIS
	Type        TransactionType `json:"type" validate:"required,oneof=VA QRIS"`
	ReferenceNo string          `json:"referenceNo"`
	//secondary reference no
	ReferenceNo2 string          `json:"referenceNo2"`
	TrxTimestamp *time.Time      `json:"trxTimestamp"`
	Amount       decimal.Decimal `json:"amount"`
	Bank         string          `json:"bank"`
}

type AutoReconTrxResponse struct {
	// transaction uuid
	UUID                 string `json:"uuid"`
	ProcessorReferenceID string `json:"processor_reference_id"`
	// type of request, VA or QRIS
	Type string              `json:"type"`
	Code constant.TReconCode `json:"code"`
	// status of request, success or failed
	Status  string `json:"status"`
	Message string `json:"message"`
}

type AutoReconResponse struct {
	Data    AutoReconTrxResponse `json:"data"`
	Code    string               `json:"code"`
	Message string               `json:"message"`
	Error   interface{}          `json:"error,omitempty"`
}
