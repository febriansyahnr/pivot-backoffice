package recurringContractModel

import (
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
)

type CreateRecurringContractRequest struct {
	ClientReferenceID   string                                   `json:"clientReferenceId" validate:"required,max=100"`
	Mode                string                                   `json:"mode" validate:"required,oneof=SELF_MANAGED"`
	Plan                Plan                                     `json:"plan" validate:"required"`
	Amount              Amount                                   `json:"amount" validate:"required"`
	Trials              []Trial                                  `json:"trials" validate:"omitempty,dive,required"`
	EndDate             string                                   `json:"endDate" validate:"required,datetime=2006-01-02T15:04:05Z"`
	BillingInterval     uint8                                    `json:"billingInterval" validate:"required,min=1"`
	BillingIntervalUnit string                                   `json:"billingIntervalUnit" validate:"required,oneof=DAY MONTH YEAR"`
	FirstAuthorization  string                                   `json:"firstAuthorization" validate:"required,oneof=ONE_DOLLAR FIRST_PAYMENT"`
	Customer            *unifiedPaymentModel.CustomerInformation `json:"customer" validate:"omitnil"`
	CustomerID          *string                                  `json:"customerId" validate:"omitnil,uuid"`
	MerchantID          string                                   `json:"-" validate:"uuid"`
	CreatedBy           string                                   `json:"-" validate:"uuid"`
}

type CancelRecurringContractRequest struct {
	MerchantID  string `json:"-"`
	RecurringID string `json:"-"`
	UpdatedBy   string `json:"-"`
}

type Amount struct {
	Currency string  `json:"currency" validate:"iso4217,oneof=IDR"`
	Value    float64 `json:"value" validate:"min=1"`
}

func (r *CreateRecurringContractRequest) Validate() error {

	firstPaymentAmount := r.Amount.Value

	if len(r.Trials) > 0 {
		if r.Trials[0].TrialStart != 1 {
			return fmt.Errorf("%s", "Ensure the initial trial value starts from 1")
		}

		firstPaymentAmount = min(
			firstPaymentAmount, (r.Amount.Value - r.Trials[0].CalculateDiscount(r.Amount.Value)),
		)

		for i, trial := range r.Trials {
			if i > 0 && (r.Trials[i-1].TrialEnd+1) != trial.TrialStart {
				return fmt.Errorf("%s", "Ensure the next trial value is sequential")
			}

			if trial.Type == constant.RecurringContractTrialTypeFree || trial.Type == constant.RecurringContractTrialTypePercentage {
				continue

			} else if trial.Amount > r.Amount.Value {
				return fmt.Errorf("%s", "Ensure the trial discount does not exceed the transaction amount")
			}
		}
	}
	if r.Customer != nil && r.CustomerID != nil {
		return fmt.Errorf("%s", "Only one of Customer Object or Customer ID can be used")

	} else if firstPaymentAmount == 0 && r.FirstAuthorization == constant.RecurringContractAuthMethodFirstPayment {
		return fmt.Errorf("%s", "First authorization must use the ONE_DOLLAR method because the first payment amount is 0")
	}
	return nil
}

type UpdateRecurringPaymentRequest struct {
	MerchantID       string
	RecurringID      string
	TransactionID    string
	PaymentTokenID   string
	PaymentMethodID  string
	RecurringPayment *unifiedPaymentModel.MetadataRecurringPayment
	UpdatedBy        string
}

type UpdateRecurringContractRequest struct {
	RecurringID       string
	TransactionID     string
	PaymentTokenID    string
	PaymentMethodID   string
	BillingCycleCount uint16
	Status            string
	UpdatedAt         time.Time
	UpdatedBy         string
	ActivatedAt       time.Time
}
