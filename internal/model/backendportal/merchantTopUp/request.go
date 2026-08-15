package merchantTopUp

import (
	"time"

	common "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/common"
)

type CreateMerchantTopUpRequest struct {
	MerchantID      string    `json:"merchantId"`
	AccountName     string    `json:"accountName"`
	PaymentMethodID string    `json:"paymentMethodId"`
	ReferenceNumber string    `json:"referenceNumber"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type MerchantTopUpRequest struct {
	AccountName     string `json:"-" validate:"required,oneof=DISBURSEMENT WALLET"` // Filled in when mapping a route
	PaymentMethodID string `json:"paymentMethodId" validate:"required"`
}

type MerchantTopUpCallbackRequest struct {
	UUID                 string                                          `json:"uuid"`
	MerchantID           string                                          `json:"merchantId"`
	MerchantName         string                                          `json:"merchantName"`
	AccountName          string                                          `json:"accountName"`
	Amount               common.Amount                                   `json:"amount"`
	BalanceBefore        common.Amount                                   `json:"balanceBefore"`
	BalanceAfter         common.Amount                                   `json:"balanceAfter"`
	PaymentMethod        MerchantTopUpCallbackPaymentMethodObject        `json:"paymentMethod"`
	PaymentMethodOptions MerchantTopUpCallbackPaymentMethodOptionsObject `json:"paymentMethodOptions"`
	TransactionTime      time.Time                                       `json:"transactionTime"`

	ParentMerchantID string `json:"-"`
}

type MerchantTopUpCallbackPaymentMethodObject struct {
	Type string `json:"type"`
}

type MerchantTopUpCallbackPaymentMethodOptionsObject struct {
	VirtualAccount *MerchantTopUpCallbackPaymentMethodOptionVAObject `json:"virtualAccount,omitempty"`
}

type MerchantTopUpCallbackPaymentMethodOptionVAObject struct {
	Channel              string `json:"channel"`
	VirtualAccountNumber string `json:"virtualAccountNumber"`
	VirtualAccountName   string `json:"virtualAccountName"`
}

type GetMerchantTopUpReferencesRequest struct {
	MerchantID string
}
