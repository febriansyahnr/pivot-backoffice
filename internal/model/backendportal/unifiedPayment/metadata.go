package unifiedPaymentModel

import (
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	splitRoutingPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/splitRoutingPayment"

	"github.com/paper-indonesia/pdk/v2/encrypt"
)

type MetadataUnifiedPayment struct {
	FeeDetail                  *feeModel.FeeMetadataObject                                  `json:"feeDetail,omitempty"`
	FeeOnBehalf                *feeModel.TrxFeeOnBehalfMetadata                             `json:"feeOnBehalf,omitempty"`
	OnBehalf                   *merchantModel.OnBehalfObject                                `json:"onBehalf,omitempty"`
	ClientRedirectUrl          *RedirectUrl                                                 `json:"clientRedirectUrl,omitempty"`
	IsUnifiedPaymentV2         bool                                                         `json:"isUnifiedPaymentV2"`
	SplitRoutingConfigurations *[]splitRoutingPaymentModel.PaymentSplitRoutingConfiguration `json:"splitRoutingConfigurations,omitempty"`
	CanceledAt                 *time.Time                                                   `json:"canceledAt,omitempty"`
	CancellationReason         string                                                       `json:"cancellationReason,omitempty"`
	IsMigratingFromV1          bool                                                         `json:"isMigratingFromV1"`
	SaveForFutureUse           *bool                                                        `json:"saveForFutureUse,omitempty"`
	ShowSavedPayment           *bool                                                        `json:"showSavedPayment,omitempty"`
	IsSnap                     *bool                                                        `json:"isSnap,omitempty"`
	VirtualTerminal            *VirtualTerminal                                             `json:"virtualTerminal,omitempty"`
	AutoSplitPayment           *AutoSplitPayment                                            `json:"autoSplitPayment,omitempty"`
	AutoSplitDetails           *AutoSplitDetails                                            `json:"autoSplitDetails,omitempty"`
	// For CC only
	RedirectPaymentUIUrl   *RedirectPaymentUIUrl    `json:"redirectUrl,omitempty"`
	EncryptedEncryptionKey *encrypt.RSAEncryptedKey `json:"encryptedEncryptionKey,omitempty"`
	RetryableConfirmation  *bool                    `json:"retryableConfirmation,omitempty"`

	// Add From Request
	AutoConfirm          bool                      `json:"autoConfirm"`
	Mode                 string                    `json:"mode"`
	StatementDescriptor  string                    `json:"statementDescriptor"`
	PaymentMethod        *PaymentMethod            `json:"paymentMethod"`
	PaymentMethodOptions PaymentMethodOptions      `json:"paymentMethodOptions,omitempty"`
	ClientMetadata       interface{}               `json:"clientMetadata,omitempty"`
	PaymentOrder         *PaymentOrder             `json:"orderInformation,omitempty"`
	ExpirationMode       string                    `json:"expirationMode,omitempty"`
	ShortPaymentURL      string                    `json:"shortPaymentUrl,omitempty"`
	RecurringPayment     *MetadataRecurringPayment `json:"recurringPayment,omitempty"`
	CardFundedPayout     *MetadataCardFundedPayout `json:"cardFundedPayout,omitempty"`

	// Static MethodDetail, the value should be similar with charge method detail
	MethodDetail *ChargePaymentMethodDetails `json:"methodDetail,omitempty"`
	// SummaryTransaction, only for static payment
	SummaryTransaction *SummaryTransaction `json:"summaryTransaction,omitempty"`
	// Bypass status page for mode redirect
	BypassStatusPage bool `json:"bypassStatusPage"`
	// OneDollarAuthorization object
	OneDollarAuthorization *OneDollarAuthorization `json:"oneDollarAuthorization,omitempty"`
}

type MetadataRecurringPayment struct {
	InitiateFirstAuthorization bool                  `json:"initiateFirstAuthorization"`
	FirstAuthorizationMethod   string                `json:"firstAuthorizationMethod,omitempty"`
	FirstAuthorizationOrderID  *string               `json:"firstAuthorizationOrderID,omitempty"`
	BillingCycle               RecurringBillingCycle `json:"billingCycle"`
}

type MetadataCardFundedPayout struct {
	Sequence         int    `json:"sequence"`
	Count            int    `json:"count"`
	FirstPaymentID   string `json:"firstPaymentId,omitempty"`
	SettlementMethod string `json:"settlementMethod"`
	CardID           string `json:"cardId"`
}

// GetBypassStatusPageState returns whether the status page should be bypassed for this unified payment metadata.
// It returns true if the payment mode is API, otherwise it returns the BypassStatusPage field value.
func (m MetadataUnifiedPayment) GetBypassStatusPageState() bool {
	if m.Mode == constant.UnifiedPaymentModeAPI {
		return true
	}

	return m.BypassStatusPage
}

func (m MetadataUnifiedPayment) IsAutoSplitPaymentAuth() bool {
	return m.AutoSplitPayment != nil &&
		m.AutoSplitPayment.TransactionType == constant.AutoSplitPaymentTypeAuthentication
}
