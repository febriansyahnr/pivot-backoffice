package xbModel

import (
	"time"

	"github.com/shopspring/decimal"
)

type XbPayoutMetadata struct {
	Uuid                string                  `json:"uuid"`
	SenderId            string                  `json:"senderId"`
	BeneficiaryId       string                  `json:"beneficiaryId"`
	SourceCurrency      string                  `json:"sourceCurrency"`
	DestinationCurrency string                  `json:"destinationCurrency"`
	PurposeCode         string                  `json:"purposeCode"`
	FxRate              decimal.Decimal         `json:"fxRate"`
	FeeFxRate           decimal.Decimal         `json:"feeFxRate"`
	DestinationFxRate   decimal.Decimal         `json:"destinationFxRate"`
	DestinationAmount   decimal.Decimal         `json:"destinationAmount"`
	SpreadValue         decimal.Decimal         `json:"spreadValue"`
	SpreadType          string                  `json:"spreadType"`
	SourceAmount        decimal.Decimal         `json:"sourceAmount"`
	TotalAmount         decimal.Decimal         `json:"totalAmount"`
	ExpiredAt           time.Time               `json:"expiredAt"`
	SenderData          SenderDataResponse      `json:"senderData"`
	BeneficiaryData     BeneficiaryDataResponse `json:"beneficiaryData"`
	RoutingCode         string                  `json:"routingCode"`
	RoutingValue        string                  `json:"routingValue"`
}
