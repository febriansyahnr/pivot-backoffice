package fdscommon

type CRMUpdateRequest struct {
	AgentCode string            `json:"agentCode"`
	IsFraud   bool              `json:"isFraud"`
	Status    string            `json:"status" validate:"omitempty,oneof=new hold queued approved cancelled"`
	FraudType string            `json:"fraudType,omitempty" validate:"required_if=IsFraud true,omitempty,oneof='payment risk' 'policy abuse' 'friendly fraud' 'identity theft' 'synthetic identity'"`
	Note      string            `json:"note"`
	Payment   *CRMPaymentUpdate `json:"payment"`
}

type CRMPaymentUpdate struct {
	CardStatus       string `json:"cardStatus,omitempty" validate:"omitempty,oneof=active cancelled decline deleted inactive lost new pick-up request restricted stolen suspended delinquency damaged expired"`
	PaymentStatus    string `json:"paymentStatus,omitempty" validate:"omitempty,oneof=auth paid 'partially paid' invoiced refunded 'partially refunded' default 'partially default' declined chargeback void"`
	ChargebackStatus string `json:"chargebackStatus,omitempty" validate:"omitempty,oneof=opened won lost 'settled by merchant'"`
}

func (c *CRMUpdateRequest) ToUpdateRequest() *UpdateRequest {
	updateReq := &UpdateRequest{
		AgentCode: &c.AgentCode,
		IsFraud:   &c.IsFraud,
		Status:    c.Status,
		FraudType: &c.FraudType,
		Note:      &c.Note,
	}

	if c.Payment != nil {
		var chargebackStatus *string
		if c.Payment.ChargebackStatus != "" {
			chargebackStatus = &c.Payment.ChargebackStatus
		}

		var cardStatus *string
		if c.Payment.CardStatus != "" {
			cardStatus = &c.Payment.CardStatus
		}

		updateReq.Payment = &PaymentUpdate{
			CardStatus:       cardStatus,
			PaymentStatus:    c.Payment.PaymentStatus,
			ChargebackStatus: chargebackStatus,
		}
	}

	return updateReq
}
