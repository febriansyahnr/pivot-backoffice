package customerModel

import (
	"time"

	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/unifiedPayment"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
)

type GeneralCustomerResponse struct {
	UUID         string    `json:"uuid"`
	MerchantID   string    `json:"merchantId"`
	Email        string    `json:"email"`
	PhoneNumber  string    `json:"phoneNumber"`
	FirstName    string    `json:"firstName"`
	LastName     string    `json:"lastName"`
	BusinessName string    `json:"businessName"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`

	City         string `json:"city"`
	Country      string `json:"country"`
	AddressLine1 string `json:"addressLine1"`
	AddressLine2 string `json:"addressLine2"`
	PostalCode   string `json:"postalCode"`
	State        string `json:"state"`

	IsBlocked   bool   `json:"isBlocked"`
	BlockReason string `json:"blockReason"`

	Metadata map[string]interface{} `json:"metadata"`
}

func (m *Customer) ToGeneralCustomerResponse() *GeneralCustomerResponse {
	return &GeneralCustomerResponse{
		UUID:         m.UUID,
		MerchantID:   m.MerchantID,
		Email:        m.Email,
		PhoneNumber:  m.PhoneNumber,
		FirstName:    m.FirstName,
		LastName:     m.LastName,
		BusinessName: m.BusinessName,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
		City:         m.City,
		Country:      m.Country,
		AddressLine1: m.AddressLine1,
		AddressLine2: m.AddressLine2,
		PostalCode:   m.PostalCode,
		State:        m.State,
		IsBlocked:    m.IsBlocked,
		BlockReason:  m.BlockReason,
		Metadata:     m.Metadata,
	}
}

func (m *Customer) ToUnifiedPaymentCustomerResponse() *unifiedPaymentModel.CustomerInformationResponse {
	refundReference, _ := util.ConvertToStruct[*unifiedPaymentModel.UnifiedPaymentRefundPreference](m.Metadata["refundPreference"])
	paymentMethods, _ := util.ConvertToStruct[[]*unifiedPaymentModel.CustomerPaymentMethodResponse](m.Metadata["paymentMethods"])
	return &unifiedPaymentModel.CustomerInformationResponse{
		GivenName: m.FirstName,
		Surname:   &m.LastName,
		SureName:  m.LastName,
		Email:     m.Email,
		PhoneNumber: &unifiedPaymentModel.UnifiedPaymentPhoneNumber{
			Number:      util.CleanUpIDNPhoneNumber(m.PhoneNumber),
			CountryCode: m.PhoneCountryCode,
		},
		RefundPreference:     refundReference,
		StoredPaymentMethods: paymentMethods,
	}
}
