package recurringContractModel

import (
	"math"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConst "github.com/paper-indonesia/pivot-backoffice/constant/payment"

	"github.com/jmoiron/sqlx/types"
)

type RecurringContract struct {
	UUID              string             `db:"uuid" json:"uuid"`
	MerchantID        string             `db:"merchant_id" json:"merchantId"`
	ClientReferenceID string             `db:"client_reference_id" json:"clientReferenceId"`
	CustomerID        string             `db:"customer_id" json:"customerId"`
	PaymentMethodID   *string            `db:"payment_method_id" json:"paymentMethodId,omitempty"`
	PaymentTokenID    *string            `db:"payment_token_id" json:"paymentTokenId,omitempty"`
	AuthMethod        string             `db:"auth_method" json:"authMethod" examples:"FIRST_PAYMENT or ONE_DOLLAR"`
	AuthTransactionID string             `db:"auth_transaction_id" json:"authTransactionId"`
	StartDate         *time.Time         `db:"start_date" json:"startDate"`
	EndDate           time.Time          `db:"end_date" json:"endDate"`
	Plan              Plan               `db:"-" json:"plan"`
	RawPlan           types.JSONText     `db:"plan" json:"-"`
	Trials            []Trial            `db:"-" json:"trials"`
	RawTrials         types.NullJSONText `db:"trials" json:"-"`
	Billing           Billing            `db:"-" json:"billing"`
	RawBilling        types.JSONText     `db:"billing" json:"-"`
	SchedulerMode     string             `db:"scheduler_mode" json:"schedulerMode" examples:"SELF_MANAGED or PLATFORM_MANAGED"`
	Currency          string             `db:"currency" json:"currency"`
	Amount            float64            `db:"amount" json:"amount"`
	Status            string             `db:"status" json:"status" examples:"CREATED or PENDING_INITIAL_AUTH or ACTIVE or INACTIVE"`
	ActivatedAt       *time.Time         `db:"activated_at" json:"activatedAt"`
	DeactivatedAt     *time.Time         `db:"deactivated_at" json:"deactivatedAt"`
	CreatedBy         string             `db:"created_by" json:"createdBy"`
	CreatedAt         time.Time          `db:"created_at" json:"createdAt"`
	UpdatedBy         string             `db:"updated_by" json:"updatedBy"`
	UpdatedAt         time.Time          `db:"updated_at" json:"updatedAt"`
	DeletedAt         *time.Time         `db:"deleted_at" json:"deletedAt"`
}

type Plan struct {
	PlanId   string `json:"planId" validate:"required,max=255"`
	PlanName string `json:"planName" validate:"required,max=255"`
}

type Trial struct {
	TrialStart uint16  `json:"trialStart" validate:"required,min=1"`
	TrialEnd   uint16  `json:"trialEnd" validate:"required,min=1,gtefield=TrialStart"`
	Type       string  `json:"type" validate:"required,oneof=FREE FIXED PERCENTAGE" examples:"FREE or FIXED or PERCENTAGE"`
	Amount     float64 `json:"amount,omitempty" validate:"required_if=Type FIXED,excluded_if=Type FREE,excluded_if=Type PERCENTAGE,omitempty,min=1"`
	Percentage float64 `json:"percentage,omitempty" validate:"required_if=Type PERCENTAGE,excluded_if=Type FREE,excluded_if=Type FIXED,omitempty,min=1,max=100"`
}

func (t *Trial) CalculateDiscount(amount float64) float64 {
	switch t.Type {
	case constant.RecurringContractTrialTypeFree:
		return amount

	case constant.RecurringContractTrialTypeFixed:
		return t.Amount

	case constant.RecurringContractTrialTypePercentage:
		return math.Round((t.Percentage / 100) * amount)
	}
	return 0.00
}

type Billing struct {
	Interval     uint8  `json:"interval"`
	IntervalUnit string `json:"intervalUnit" examples:"DAY or MONTH or YEAR"`
	Count        uint16 `json:"count"`
}

type RecurringContractDetail struct {
	UUID                   string             `db:"uuid" json:"uuid"`
	MerchantID             string             `db:"merchant_id" json:"merchantId"`
	CustomerID             string             `db:"customer_id" json:"customerId"`
	ClientReferenceID      string             `db:"client_reference_id" json:"clientReferenceId"`
	PaymentMethodID        *string            `db:"payment_method_id" json:"paymentMethodId"`
	PaymentTokenID         *string            `db:"payment_token_id" json:"paymentTokenId"`
	AuthMethod             string             `db:"auth_method" json:"authMethod"`
	AuthTransactionID      *string            `db:"auth_transaction_id" json:"authTransactionId"`
	StartDate              *time.Time         `db:"start_date" json:"startDate"`
	EndDate                time.Time          `db:"end_date" json:"endDate"`
	Plan                   Plan               `db:"-" json:"plan"`
	RawPlan                types.JSONText     `db:"plan" json:"-"`
	Trials                 []Trial            `db:"-" json:"trials"`
	RawTrials              types.NullJSONText `db:"trials" json:"-"`
	Billing                Billing            `db:"-" json:"billing"`
	RawBilling             types.JSONText     `db:"billing" json:"-"`
	Currency               string             `db:"currency" json:"currency"`
	Amount                 float64            `db:"amount" json:"amount"`
	Status                 string             `db:"status" json:"status"`
	CreatedAt              time.Time          `db:"created_at" json:"createdAt"`
	UpdatedAt              time.Time          `db:"updated_at" json:"updatedAt"`
	PaymentMethodCategory  *string            `db:"category" json:"paymentMethodCategory"`
	PaymentMethodType      *string            `db:"type" json:"paymentMethodType"`
	PaymentMethodSubType   *string            `db:"sub_type" json:"paymentMethodSubType"`
	ProcessorReference     *string            `db:"processor_reference" json:"processorReference"`
	ProcessorReferenceID   *string            `db:"processor_reference_id" json:"processorReferenceId"`
	ProcessorTransactionID *string            `db:"processor_transaction_id" json:"processorTransactionId"`
	ProcessorOrderID       *string            `db:"processor_order_id" json:"processorOrderId"`
}

func (r *RecurringContractDetail) IsFirstAuthorization() bool {
	switch r.Status {
	default:
		return false

	case constant.RecurringContractStatusCreated, constant.RecurringContractStatusPendInitialAuth:
		return true
	}
}

func (r *RecurringContractDetail) GetRecurringAmountForBillingCycle(initiateFirstAuthorization bool) float64 {

	billingCycle := r.Billing.Count + 1

	if (r.IsFirstAuthorization() && r.AuthMethod == constant.RecurringContractAuthMethodOneDollar) ||
		(r.Status == constant.RecurringContractStatusActive && initiateFirstAuthorization) {

		return constant.RecurringContractOneDollarAuthAmountIDR

	} else if len(r.Trials) == 0 {
		return r.Amount
	}

	var trial *Trial

	for _, t := range r.Trials {
		if billingCycle >= t.TrialStart && billingCycle <= t.TrialEnd {
			trial = &t

			break
		}
	}
	if trial == nil {
		return r.Amount
	}

	discount := trial.CalculateDiscount(r.Amount)

	return r.Amount - discount
}

func (r *RecurringContractDetail) GetUnifiedPaymentMethodType() string {
	if r.PaymentMethodType == nil {
		return ""
	}

	switch *r.PaymentMethodType {
	default:
		return *r.PaymentMethodType

	case paymentConst.PAYMENT_METHOD_CREDIT_CARD:
		return constant.UnifiedPaymentMethodCard
	}
}

func (r *RecurringContractDetail) GetMinMaxAmountPerPayment() (minAmount, maxAmount float64) {

	minAmount, maxAmount = r.Amount, r.Amount

	for _, trial := range r.Trials {
		minAmount = min(minAmount, (r.Amount - trial.CalculateDiscount(r.Amount)))
	}

	return
}
