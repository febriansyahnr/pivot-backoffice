package paymentModel

import (
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/unifiedPayment"
	"github.com/shopspring/decimal"
)

type BookingPayload struct {
	BookingID       string `json:"bookingId" validate:"required"`
	ReferenceID     string `json:"referenceId" validate:"required"`
	Amount          Amount `json:"amount"`
	TravelAgentCode string `json:"travelAgentCode" validate:"required"`
	Card            Card   `json:"card" validate:"required"`
	Remarks         string `json:"remarks" validate:"omitempty,maxChar=20"`
	// For internal used
	MerchantID        string   `json:"-" validate:"-"`
	BankMerchantID    string   `json:"-" validate:"-"`
	AllowedBinNumbers []string `json:"-" validate:"-"`
	AllowedCardTypes  []string `json:"-" validate:"-"`
	AllowedPrincipal  []string `json:"-" validate:"-"`
	BatchID           string   `json:"-" validate:"-"`
	PaymentID         string   `json:"-" validate:"-"`
	CreatedBy         string   `json:"-" validate:"-"`
	ChargeID          string   `json:"-" validate:"-"`
}

func (b BookingPayload) ToResponse() BookingPayload {
	b.Card.Number = b.Card.Number[:6] + "xxxxxx" + b.Card.Number[12:]
	b.Card.SecurityCode = strings.Repeat("x", len(b.Card.SecurityCode))

	return b
}

func (b BookingPayload) ToLog() BookingPayload {
	b = b.ToResponse()
	b.Card.Expiry.Month = "xx"
	b.Card.Expiry.Year = "xx"
	return b
}

func (b *BookingPayload) ToCardChargePayload() CardChargePayload {
	return CardChargePayload{
		PaymentID:           b.PaymentID,
		MerchantID:          b.MerchantID,
		ClientTransactionID: b.ReferenceID,
		Amount:              b.Amount.Value.InexactFloat64(),
		Currency:            b.Amount.Currency,
		Card: CardChargePayloadCHD{
			Number:       b.Card.Number,
			SecurityCode: b.Card.SecurityCode,
			Expiry:       b.Card.Expiry,
		},
		VCCTerminal: true,
	}
}

func (b *BookingPayload) ToVCCTerminalChargeMessage(encryptedPayload string) VCCTerminalChargeMessage {
	return VCCTerminalChargeMessage{
		MerchantID:       b.MerchantID,
		BatchID:          b.BatchID,
		PaymentID:        b.PaymentID,
		ChargeID:         b.ChargeID,
		EncryptedPayload: encryptedPayload,
	}
}

func (b *BookingPayload) ToCreateUnifiedPaymentSessionRequest(conf config.VccTerminalConfig) *unifiedPaymentModel.CreateUnifiedPaymentSessionRequest {
	return &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
		ClientReferenceID: b.ReferenceID,
		PaymentMethod: &unifiedPaymentModel.PaymentMethod{
			Type: constant.UnifiedPaymentMethodCard,
		},
		PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
			Card: &unifiedPaymentModel.PaymentMethodOptionCard{
				ThreeDsMethod: constant.CardThreeDsMethodNever,
				ProcessingConfig: &unifiedPaymentModel.PaymentMethodOptionCardProcessingConfig{
					BankMerchantId: b.BankMerchantID,
				},
			},
		},
		Mode:        constant.UnifiedPaymentModeRedirect,
		AutoConfirm: true,
		ExpiryAt:    time.Now().Add(time.Duration(conf.ExpiryAfterMinutes) * time.Minute).UTC(),
		Amount: unifiedPaymentModel.Amount{
			Value:    b.Amount.Value.InexactFloat64(),
			Currency: b.Amount.Currency,
		},
		MerchantID:  b.MerchantID,
		PaymentID:   b.PaymentID,
		PaymentType: constant.TypeVirtualTerminal,
		CreatedBy:   b.CreatedBy,
		CreatedFrom: constant.SourceDashboard,
		VirtualTerminal: &unifiedPaymentModel.VirtualTerminal{
			BatchID:            b.BatchID,
			BookingID:          b.BookingID,
			TravelAgentCode:    b.TravelAgentCode,
			TravelAgentName:    conf.TravelAgents[b.TravelAgentCode],
			Remarks:            b.Remarks,
			AcquirerMerchantID: b.BankMerchantID,
			AllowedBinNumbers:  b.AllowedBinNumbers,
			AllowedCardTypes:   b.AllowedCardTypes,
			AllowedPrincipal:   b.AllowedPrincipal,
		},
	}
}

type Card struct {
	Number       string `json:"number" validate:"required,luhn"`
	SecurityCode string `json:"securityCode" validate:"omitempty,min=3,max=4"`
	Expiry       Expiry `json:"expiry" validate:"required"`
}

type Expiry struct {
	Month string `json:"month" validate:"required,number,len=2"`
	Year  string `json:"year" validate:"required,number,len=2"`
}

// VCCTerminalChargeMessage is the RabbitMQ message payload published
// per booking to be processed for payment creation and charge.
type VCCTerminalChargeMessage struct {
	MerchantID       string `json:"merchantId"`
	BatchID          string `json:"batchId"`
	PaymentID        string `json:"paymentId"`
	ChargeID         string `json:"chargeId"`
	EncryptedPayload string `json:"encryptedPayload"`
}

// CardChargePayload is the plaintext payload to be encrypted
// before sending to the card core processor.
type CardChargePayload struct {
	PaymentID           string               `json:"payment_id"`
	MerchantID          string               `json:"merchant_id"`
	ClientTransactionID string               `json:"client_transaction_id"`
	Amount              float64              `json:"amount"`
	Currency            string               `json:"currency"`
	Card                CardChargePayloadCHD `json:"card"`
	VCCTerminal         bool                 `json:"vcc_terminal"`
}

type CardChargePayloadCHD struct {
	Number       string `json:"number"`
	SecurityCode string `json:"security_code"`
	Expiry       Expiry `json:"expiry"`
}

type VccTerminalItem struct {
	BulkID      string    `json:"bulkId" db:"bulk_id"`
	ChargeID    string    `json:"chargeId" db:"charge_id"`
	ChargeDate  time.Time `json:"chargeDate" db:"created_at"`
	Amount      Amount    `json:"amount" db:"-"`
	Status      string    `json:"status" db:"status"`
	TravelAgent string    `json:"travelAgent" db:"travel_agent"`
	ReferenceID string    `json:"referenceId" db:"reference_id"`
	BookingID   string    `json:"bookingId" db:"booking_id"`

	ChargeAmount   decimal.Decimal `json:"-" db:"amount"`
	ChargeCurrency string          `json:"-" db:"currency"`
}
