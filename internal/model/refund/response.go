package refundModel

import (
	"encoding/json"
	"fmt"
	"time"

	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
)

type RefundResponse struct {
	ID                  string                                            `json:"id" db:"id"`
	MerchantID          string                                            `json:"-" db:"merchant_id"`
	ClientReferenceID   string                                            `json:"clientReferenceId" db:"client_reference_id"`
	PaymentSessionID    string                                            `json:"paymentSessionId" db:"payment_session_id"`
	ChargeID            string                                            `json:"chargeId" db:"charge_id"`
	CapturedAmount      commonModel.Amount                                `json:"capturedAmount" db:"captured_amount"`
	IsFullAmount        bool                                              `json:"isFullAmount" db:"is_full_amount"`
	Amount              commonModel.Amount                                `json:"amount" db:"amount"`
	Status              string                                            `json:"status" db:"status"`
	Reason              string                                            `json:"reason" db:"reason"`
	Description         string                                            `json:"description" db:"description"`
	DestinationType     string                                            `json:"destinationType,omitempty" db:"destination_type"`
	Method              string                                            `json:"method" db:"method"`
	TransferDestination *TransferDestination                              `json:"transferDestination,omitempty" db:"transfer_destination"`
	ChannelDestination  *ChannelDestination                               `json:"channelDestination,omitempty" db:"-"`
	Metadata            interface{}                                       `json:"metadata,omitempty" db:"metadata"` // client-sent metadata
	FailureCode         string                                            `json:"failureCode,omitempty" db:"failure_code"`
	CreatedAt           time.Time                                         `json:"createdAt" db:"created_at"`
	UpdatedAt           time.Time                                         `json:"updatedAt" db:"updated_at"`
	StatusHistory       []unifiedPaymentModel.RefundStatusHistoryResponse `json:"statusHistory,omitempty"`

	PaymentChannel          string      `json:"-" db:"payment_channel"`
	RawChargeAdditionalInfo interface{} `json:"-" db:"charge_additional_info"`
}

// ChannelDestination contains payment method details when destinationType is CHANNEL
type ChannelDestination struct {
	PaymentMethod  string                 `json:"paymentMethod"`
	PaymentChannel string                 `json:"paymentChannel,omitempty"`
	PaymentDetail  map[string]interface{} `json:"paymentDetail,omitempty"`
}

// RefundDetailResponse represents the detailed refund response including original payment info
type RefundDetailResponse struct {
	Refund          *RefundResponse      `json:"refund"`
	OriginalPayment *OriginalPaymentInfo `json:"originalPayment,omitempty"`
}

// OriginalPaymentInfo represents simplified payment information for refund detail response
type OriginalPaymentInfo struct {
	ID            string             `json:"id"`
	Amount        commonModel.Amount `json:"amount"`
	PaymentMethod string             `json:"paymentMethod"`
	Status        string             `json:"status"`
	CreatedAt     time.Time          `json:"createdAt"`
}

// BuildChannelDestination populates the ChannelDestination field based on the charge's payment method details.
// This should be called when DestinationType is "CHANNEL" and RawChargeAdditionalInfo contains payment data.
func (r *RefundResponse) BuildChannelDestination() {
	if r.DestinationType != "CHANNEL" || r.RawChargeAdditionalInfo == nil {
		return
	}

	// Convert raw additional info to map
	var additionalInfo map[string]interface{}
	switch v := r.RawChargeAdditionalInfo.(type) {
	case []byte:
		if err := json.Unmarshal(v, &additionalInfo); err != nil {
			return
		}
	case string:
		if err := json.Unmarshal([]byte(v), &additionalInfo); err != nil {
			return
		}
	case map[string]interface{}:
		additionalInfo = v
	default:
		return
	}

	methodDetail, ok := additionalInfo["methodDetail"].(map[string]interface{})
	if !ok {
		return
	}

	destination := &ChannelDestination{
		PaymentChannel: r.PaymentChannel,
	}

	// Extract details based on payment method type
	if qrDetail, ok := methodDetail["qr"].(map[string]interface{}); ok {
		destination.PaymentMethod = "QRIS"
		destination.PaymentDetail = map[string]interface{}{}
		if acquirer, ok := qrDetail["acquirer"].(string); ok {
			destination.PaymentDetail["acquirer"] = acquirer
		}
		if merchantName, ok := qrDetail["merchantName"].(string); ok {
			destination.PaymentDetail["merchantName"] = merchantName
		}
		if rrn, ok := qrDetail["retrievalReferenceNumber"].(string); ok && rrn != "" {
			destination.PaymentDetail["rrn"] = rrn
		}
	} else if cardDetail, ok := methodDetail["card"].(map[string]interface{}); ok {
		destination.PaymentMethod = "CREDIT_CARD"
		destination.PaymentDetail = map[string]interface{}{}
		if last4, ok := cardDetail["last4"].(string); ok {
			destination.PaymentDetail["last4Digit"] = last4
		}
		if binInfo, ok := cardDetail["binInformations"].(map[string]interface{}); ok {
			if issuingBank, ok := binInfo["issuingBank"].(string); ok {
				destination.PaymentDetail["cardIssuing"] = issuingBank
			}
			if brand, ok := binInfo["brand"].(string); ok {
				destination.PaymentDetail["cardBrand"] = brand
			}
		}
		// Extract RRN from authorizationResult if available
		if authResult, ok := cardDetail["authorizationResult"].(map[string]interface{}); ok {
			if rrn, ok := authResult["retrievalReferenceNumber"].(string); ok && rrn != "" {
				destination.PaymentDetail["rrn"] = rrn
			}
		}
	} else if ewalletDetail, ok := methodDetail["ewallet"].(map[string]interface{}); ok {
		destination.PaymentMethod = "EWALLET"
		destination.PaymentDetail = map[string]interface{}{}
		if channel, ok := ewalletDetail["channel"].(string); ok && channel != "" {
			destination.PaymentDetail["channel"] = channel
		}
	}

	r.ChannelDestination = destination
}

func (r *RefundResponse) UnmarshalJSON(data []byte) error {
	type Alias RefundResponse // Prevent recursion
	raw := &struct {
		Metadata json.RawMessage `json:"metadata"` // intercept just this one
		*Alias
	}{
		Alias: (*Alias)(r),
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if len(raw.Metadata) == 0 {
		return nil
	}

	var m map[string]any
	if err := json.Unmarshal(raw.Metadata, &m); err != nil {
		return fmt.Errorf("failed to unmarshal additionalInfo: %w", err)
	}

	// Use the inner "value" if present and it's a map
	if v, ok := m["value"].(map[string]any); ok {
		r.Metadata = &v
	} else {
		r.Metadata = &m
	}

	return nil
}

// GetRefundReceiptRequest is the request for getting refund receipt
type GetRefundReceiptRequest struct {
	RefundID   string `json:"-"`
	MerchantID string `json:"-"`
}

// GetRefundReceiptResponse is the response for refund receipt
type GetRefundReceiptResponse struct {
	ReceiptURL string `json:"receiptUrl"`
}

// RefundReceiptData is the data structure for refund receipt PDF template
type RefundReceiptData struct {
	Amount             string
	CompletedAt        string
	RefundReferenceID  string
	PaymentReferenceID string
	RefundReason       string
	RefundDestination  string
	RRN                string

	// Images
	ImageHeader     string
	ImageBackground string
}
