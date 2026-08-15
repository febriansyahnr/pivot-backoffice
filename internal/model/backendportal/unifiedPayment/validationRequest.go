package unifiedPaymentModel

import (
	paymentMethodModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/paymentMethod"
)

type ValidateCardRequest struct {
	IsConfirmStep            bool
	Mode                     string
	CardPaymentMethod        *CardPaymentMethodDetail
	CardPaymentMethodOptions *PaymentMethodOptionCard
	CardSupportedUseCases    []*paymentMethodModel.CardSupportedUseCase
	IsRecurringPayment       bool
	IsVirtualTerminal        bool
	IsCardFundedPayout       bool
	IsAutoSplitPayment       bool
	HasCardOnFile            bool
}
