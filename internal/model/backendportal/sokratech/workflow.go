package sokratech

import (
	fdscommon "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/fdsProcessor/fdsCommon"
)

// Create an alias for structs imported from the fdscommon package.
// This avoids duplicating structs with the same attributes and preserves model decoupling between common FDS and specific FDS implementations.
type (
	Merchant              fdscommon.Merchant
	Transaction           fdscommon.Transaction
	PaymentMethodTypeCard fdscommon.PaymentMethodTypeCard
)

type PayoutDestination struct {
	BankCode                string `json:"bankCode"`
	AccountNumber           string `json:"accountNumber"`
	AccountName             string `json:"accountName"`
	AccountNumberTypeNumber int64  `json:"accountNumberTypeNumber"`
}

type Customer struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Email       string `json:"email,omitempty"`
	PhoneNumber string `json:"phoneNumber,omitempty"`
	Address     string `json:"address,omitempty"`
}

type PaymentMethod struct {
	Type string                 `json:"type"` // CARD, QRIS, VIRTUAL_ACCOUNT, etc
	Card *PaymentMethodTypeCard `json:"card,omitempty"`
}

type Device struct {
	IPType      string `json:"ipType,omitempty"`
	IPAddress   string `json:"ipAddress,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
	UserAgent   string `json:"userAgent,omitempty"`
}
