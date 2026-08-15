package unifiedPaymentModel

import (
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
)

type DashboardPaymentLinkCreateRequest struct {
	MerchantID        string
	UserID            string
	ClientReferenceID string                     `json:"clientReferenceId" validate:"required"`
	ExpiredAt         time.Time                  `json:"expiredAt" validate:"required"`
	Amount            Amount                     `json:"amount" validate:"required"`
	Customer          PaymentLinkCustomerRequest `json:"customer" validate:"required"`
}

type PaymentLinkCustomerRequest struct {
	Email string `json:"email" validate:"required,email"`
}

func (r *DashboardPaymentLinkCreateRequest) Validate() error {
	if r.Amount.Value < constant.DashboardPaymentLinkMinAmount {
		return constant.ErrPaymentLinkMinAmount
	}
	if r.Amount.Value > constant.DashboardPaymentLinkMaxAmount {
		return constant.ErrPaymentLinkMaxAmount
	}
	if r.ExpiredAt.UTC().Before(time.Now().UTC()) {
		return fmt.Errorf(constant.ErrMsgExpiryLessThanCurrentTime)
	}
	return nil
}

type DashboardPaymentLinkResponse struct {
	ID                string    `json:"id"`
	ClientReferenceID string    `json:"clientReferenceId"`
	Amount            Amount    `json:"amount"`
	PaymentLink       string    `json:"paymentLink"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"createdAt"`
	ExpiryAt          time.Time `json:"expiryAt"`
}

func (r *UnifiedPaymentSessionResponse) ToDashboardPaymentLinkResponse() *DashboardPaymentLinkResponse {
	response := &DashboardPaymentLinkResponse{
		ID:                r.ID,
		ClientReferenceID: r.ClientReferenceID,
		Amount:            r.Amount,
		PaymentLink:       r.ShortPaymentUrl,
		Status:            r.Status,
		CreatedAt:         r.CreatedAt,
	}
	if r.ExpiryAt != nil {
		response.ExpiryAt = *r.ExpiryAt
	}
	return response
}
