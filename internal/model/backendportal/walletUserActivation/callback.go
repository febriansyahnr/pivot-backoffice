package walletUserActivation

import (
	"time"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/proto/common"
)

type UserActivationCallbackRequest struct {
	PartnerReferenceNo  string               `json:"partnerReferenceNo"`
	BindingID           string               `json:"bindingId"`
	UserData            *UserData            `json:"user"`
	UserApplicationData *UserApplicationData `json:"userApplication,omitempty"`
}

type UserData struct {
	UUID        string `json:"uuid,omitempty"`
	PhoneNumber string `json:"phoneNumber"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	CreatedAt   string `json:"createdAt"`
}

type UserApplicationData struct {
	Status    string           `json:"status"`
	CreatedAt string           `json:"createdAt"`
	UpdatedAt string           `json:"updatedAt"`
	Data      *ApplicationData `json:"data,omitempty"`
}

type ApplicationData struct {
	IdentityNumber   string `json:"identityNumber"`
	IdentityType     string `json:"identityType"`
	FullName         string `json:"fullName"`
	POB              string `json:"pob"`
	DOB              string `json:"dob"`
	Nationality      string `json:"nationality"`
	DomicileAddress  string `json:"domicileAddress"`
	DomicileProvince string `json:"domicileProvince"`
	DomicileCity     string `json:"domicileCity"`
	DomicilePostcode string `json:"domicilePostcode"`
	Province         string `json:"province"`
	City             string `json:"city"`
	PostalCode       string `json:"postalCode"`
	Address          string `json:"address"`
	Occupation       string `json:"occupation"`
	Gender           string `json:"gender"`
}

type AccountLinkageRequest struct {
	UserID             string `json:"userId,omitempty"`
	PartnerReferenceNo string `json:"partnerReferenceNo"`
	PhoneNumber        string `json:"phoneNumber"`
	MerchantID         string `json:"merchantId"`
	BindingID          string `json:"bindingId"`
	Status             string `json:"status"`
	CreatedAt          string `json:"createdAt"`
}

type WalletTransaction struct {
	ID                  string      `json:"id"`
	OrderID             string      `json:"orderID"`
	ReferenceID         string      `json:"referenceId"`
	CustomerRefNo       string      `json:"customerRefNo"`
	Currency            string      `json:"currency"`
	Amount              float64     `json:"amount"`
	Fee                 float64     `json:"fee"`
	Category            string      `json:"category"`
	CategoryType        string      `json:"categoryType"`
	CategoryTypeProduct string      `json:"categoryTypeProduct"`
	ProductCode         string      `json:"productCode"`
	Status              string      `json:"status"`
	Remarks             string      `json:"remarks"`
	CreatedAt           time.Time   `json:"createdAt"`
	UpdatedAt           time.Time   `json:"updatedAt"`
	AdditionalData      interface{} `json:"additionalData"`
}

type SnapQrisMpmCallbackRequest struct {
	OriginalReferenceNo        string         `json:"originalReferenceNo"`
	OriginalPartnerReferenceNo string         `json:"originalPartnerReferenceNo"`
	LatestTransactionStatus    string         `json:"latestTransactionStatus"`
	TransactionStatusDesc      string         `json:"transactionStatusDesc"`
	Amount                     *common.Amount `json:"amount"`
	AdditionalData             map[string]any `json:"additionalData"`
}

type SnapDirectDebitCallbackRequest struct {
	OriginalReferenceNo        string         `json:"originalReferenceNo"`
	OriginalPartnerReferenceNo string         `json:"originalPartnerReferenceNo"`
	LatestTransactionStatus    string         `json:"latestTransactionStatus"`
	TransactionStatusDesc      string         `json:"transactionStatusDesc"`
	Amount                     *common.Amount `json:"amount"`
	AdditionalData             map[string]any `json:"additionalData"`
}
