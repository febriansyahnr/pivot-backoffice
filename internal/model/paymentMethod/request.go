package paymentMethodModel

import (
	"encoding/json"
	"slices"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
)

type SetupPaymentMethodConfigRequest struct {
	ChannelType            string                                  `json:"channelType,omitempty" validate:"oneof=DIRECT AGGREGATOR"`
	ChannelConfig          *SetupPaymentMethodChannelConfigRequest `json:"channelConfig,omitempty"`
	SplitCardPaymentConfig *SplitCardPaymentConfig                 `json:"splitCardPaymentConfig,omitempty" validate:"omitempty,required"`
	CardFundedPayoutConfig *CardFundedPayoutConfig                 `json:"cardFundedPayoutConfig,omitempty" validate:"omitempty,required"`
	VirtualTerminalConfig  *VirtualTerminalConfig                  `json:"virtualTerminalConfig,omitempty" validate:"omitempty,required"`
	PartnerConfig          *SetupPaymentMethodPartnerConfigRequest `json:"partnerConfig,omitempty" validate:"omitempty"`

	PaymentMethodID string `json:"-"`
	MerchantID      string `json:"-"`
}

type SetupPaymentMethodChannelConfigRequest struct{}

type SetupPaymentMethodPartnerConfigRequest struct {
	VirtualAccount *SetupPaymentMethodPartnerConfigForVARequest          `json:"virtualAccount,omitempty" validate:"omitempty"`
	Card           *SetupPaymentMethodPartnerConfigForCardRequest        `json:"card,omitempty" validate:"omitempty"`
	Qris           *SetupPaymentMethodPartnerConfigForQrisRequest        `json:"qris,omitempty" validate:"omitempty"`
	EWallet        *SetupPaymentMethodPartnerConfigForEWalletRequest     `json:"eWallet,omitempty" validate:"omitempty"`
	Installment    *SetupPaymentMethodPartnerConfigForInstallmentRequest `json:"installment,omitempty" validate:"omitempty"`
}

type SetupPaymentMethodPartnerConfigForVARequest struct {
	Items []SetupPaymentMethodPartnerConfigForVAObj

	PartnerConfigReferenceID string `json:"-"`
	MerchantID               string `json:"-"`
	MerchantMID              string `json:"-"`
	ChannelType              string `json:"-"`
	Acquirer                 string `json:"-"`
}

// Custom JSON marshal: only serialize the slice
func (r SetupPaymentMethodPartnerConfigForVARequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.Items)
}

// Custom JSON unmarshal: deserialize the slice into Items
func (r *SetupPaymentMethodPartnerConfigForVARequest) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &r.Items)
}

type SetupPaymentMethodPartnerConfigForVAObj struct {
	BINPrefix  string `json:"binPrefix,omitempty"`
	Type       string `json:"type,omitempty" validate:"oneof=OPEN_STATIC CLOSED_DYNAMIC CLOSED_STATIC"`
	StartRange string `json:"startRange,omitempty"`
	EndRange   string `json:"endRange,omitempty"`
}

type SetupPaymentMethodPartnerConfigForCardRequest struct {
	Items []SetupPaymentMethodPartnerConfigForCardObj `validate:"omitempty,dive"`

	MerchantID             string                  `json:"-"`
	MerchantShortName      string                  `json:"-"`
	ChannelType            string                  `json:"-"`
	SplitCardPaymentConfig *SplitCardPaymentConfig `json:"-"`
	CardFundedPayoutConfig *CardFundedPayoutConfig `json:"-"`
	PaymentMethodType      string                  `json:"-"`
}

// Custom JSON marshal: only serialize the slice
func (r SetupPaymentMethodPartnerConfigForCardRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.Items)
}

// Custom JSON unmarshal: deserialize the slice into Items
func (r *SetupPaymentMethodPartnerConfigForCardRequest) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &r.Items)
}

func (s SetupPaymentMethodPartnerConfigForCardRequest) GetPartnerConfigByPaymentType(paymentType string) *SetupPaymentMethodPartnerConfigForCardRequest {
	if len(s.Items) == 0 {
		return nil
	}

	s.Items = slices.DeleteFunc(s.Items, func(p SetupPaymentMethodPartnerConfigForCardObj) bool {
		switch paymentType {
		case constant.GroupPaymentTypePayment, constant.GroupPaymentTypeOneDollarAuth:
			return p.RecurringType != "" || p.TravelAgentCode != "" || p.CardFundedPayoutType != "" || p.SplitPaymentType != ""

		case constant.GroupPaymentTypeRecurringPayment:
			return p.RecurringType == ""

		case constant.GroupPaymentTypeVirtualTerminal:
			return p.TravelAgentCode == ""

		case constant.GroupPaymentTypeCardFundedPayout:
			return p.CardFundedPayoutType == ""

		case constant.GroupPaymentTypeSplitPayment:
			return p.SplitPaymentType == ""

		default:
			return false
		}
	})
	return &s
}

func (r SetupPaymentMethodPartnerConfigForCardRequest) GetNetworkTokenPartnerConfig(cofInitiatorType string) *SetupPaymentMethodPartnerConfigForCardObj {
	for _, item := range r.Items {
		if item.NetworkToken == nil {
			continue
		}

		if cofInitiatorType == "" && item.NetworkToken.Type == constant.NetworkTokenDefaultType {
			return &item
		}

		if item.NetworkToken.Type == constant.NetworkTokenCOFType && item.NetworkToken.COFInitiator == cofInitiatorType {
			return &item
		}
	}
	return nil
}

type SetupPaymentMethodPartnerConfigForCardObj struct {
	NetworkToken         *CardNetworkTokenPartnerConfigObj `json:"networkToken,omitempty"`
	TravelAgentCode      string                            `json:"travelAgentCode,omitempty" validate:"omitempty,uppercase"`
	CardFundedPayoutType string                            `json:"cardFundedPayoutType,omitempty" validate:"omitempty,oneof=CIT MIT"`
	RecurringType        string                            `json:"recurringType,omitempty" validate:"omitempty,oneof=CIT MIT"`
	SplitPaymentType     string                            `json:"splitPaymentType,omitempty" validate:"omitempty,oneof=CIT MIT"`
	PartnerProcessor     string                            `json:"partnerProcessor" validate:"required,oneof=MPGS CYBS"`
	PartnerBaseURL       string                            `json:"partnerBaseURL" validate:"-"`
	AcquirerMerchantID   string                            `json:"acquirerMerchantId" validate:"required"`
	Description          string                            `json:"description"`
	Priority             int                               `json:"priority"`
	IsActive             bool                              `json:"isActive"`
	PFID                 string                            `json:"pfID,omitempty"`
	PFName               string                            `json:"pfName,omitempty"`
	CardTypes            []string                          `json:"cardTypes,omitempty"`
	PrincipalAvailable   []string                          `json:"principalAvailable,omitempty"`
	PrioritizedBIN       []string                          `json:"prioritizedBIN,omitempty"`
	Acquirer             string                            `json:"acquirer,omitempty"`
	MerchantIDTag        string                            `json:"merchantIdTag,omitempty"`
	SupportedUseCase     *CardSupportedUseCase             `json:"supportedUseCase"`
	AggregatorConfig     map[string]CardAggregatorConfig   `json:"aggregatorConfig,omitempty"` // the key is the principal uppercase name

	// due to merchant able to use multiple MID
	// which each MID can have different channelType
	// so this field is to override the channelType
	// if empty, it will use the channelType from parent struct
	// which is SetupPaymentMethodPartnerConfigForCardRequest
	ChannelType string `json:"channelType,omitempty" validate:"oneof=DIRECT AGGREGATOR"`
}

type CardNetworkTokenPartnerConfigObj struct {
	Type         string `json:"type" validate:"required,oneof=DEFAULT COF"`
	COFInitiator string `json:"cofInitiator" validate:"omitempty,required_if=Type COF,oneof=MERCHANT CUSTOMER"`
}

type CardSupportedUseCase struct {
	AllowBypass3Ds                bool     `json:"allowBypass3ds"`
	AllowForeignCard              bool     `json:"allowForeignCard"`
	AllowExternalThreeDs          bool     `json:"allowExternalThreeDs"`
	AllowedCountryRiskLevelNon3ds []string `json:"allowedCountryRiskLevelNon3ds" validate:"omitempty,dive,oneof=LOW MEDIUM HIGH"` // LOW, MEDIUM, HIGH
	AllowedCountryRiskLevel3ds    []string `json:"allowedCountryRiskLevel3ds" validate:"omitempty,dive,oneof=LOW MEDIUM HIGH"`    // LOW, MEDIUM, HIGH
	AllowedECICodes               []string `json:"allowedECICodes,omitempty" validate:"omitempty,slicestrcontains=02 05,unique,dive,number"`
	AllowedBinNumbers             []string `json:"allowedBinNumbers" validate:"omitempty,unique,dive,required,number"`
}

type SetupPaymentMethodPartnerConfigForQrisRequest struct {
	AcquirerMerchantID string   `json:"acquirerMerchantId"`
	AcquirerTerminalID string   `json:"acquirerTerminalId"`
	AcquirerStoreIDs   []string `json:"acquirerStoreIdList"` // for BNC Static QRIS purpose
	QRType             string   `json:"qrType"`

	Acquirer     string `json:"-"`
	MerchantType string `json:"merchantType"`
	CreatedBy    string `json:"createdBy"`
	ChannelType  string `json:"-"`

	MerchantID         string `json:"-"`
	MerchantExternalID string `json:"-"`
}

type SetupPaymentMethodPartnerConfigForEWalletRequest struct {
	ExternalMerchantID string `json:"externalMerchantId"`
	ExternalStoreID    string `json:"externalStoreId"`

	SubMerchantID string `json:"subMerchantId"`

	ChannelType string `json:"-"`
}

type SetupPaymentMethodPartnerConfigForInstallmentRequest struct {
	InstallmentPlanIDs []string `json:"installmentPlanIds"`
}

type ChangeActivationStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=NOT_REQUESTED REQUESTED SUBMITTED APPROVED REJECTED"`

	PaymentMethodID string `json:"-"`
	MerchantID      string `json:"-"`
}

type GetRequiredMerchantDocumentsRequest struct {
	PaymentMethodID string `json:"-"`
	MerchantID      string `json:"-"`
}

type UpdateVAStaticRangeRequest struct {
	OpenRange  *VAStaticRange `json:"openRange"`
	CloseRange *VAStaticRange `json:"closeRange"`
}

type VAStaticRange struct {
	BinPrefix string `json:"binPrefix"`
	Start     string `json:"start"`
	End       string `json:"end"`
}

type CreatePaymentMethodRequest struct {
	UUID               string  `json:"-" db:"uuid"`
	Type               string  `json:"type" db:"type" validate:"required,oneof=VIRTUAL_ACCOUNT EWALLET INSTALLMENT QRIS CREDIT_CARD"`
	Subtype            string  `json:"subtype" db:"sub_type"`
	Category           string  `json:"category" db:"category" validate:"required,oneof=PAYMENT MERCHANT_TOP_UP"`
	Name               string  `json:"name" db:"name" validate:"required,max=60"`
	Description        *string `json:"description" db:"description" validate:"max=255"`
	Logo               *string `json:"logo" db:"logo" validate:"max=255"`
	Acquirer           string  `json:"acquirer" db:"acquirer" validate:"required,max=20"`
	BankName           string  `json:"bankName" db:"bank_name" validate:"max=60"`
	Instructions       *string `json:"instructions" db:"instructions"`
	Processor          string  `json:"processor" db:"processor" validate:"max=40"`
	ActivationMethod   string  `json:"activationMethod" db:"activation_method" validate:"required,oneof=INSTANT MANUAL API"`
	CountryOfOperation string  `json:"countryOfOperation" db:"country_of_operation" validate:"max=2"`
	SupportedCurrency  string  `json:"supportedCurrency" db:"supported_currency" validate:"max=3"`
}

// GetMaxBNCQRStaticLimit returns the number of AcquirerStoreIDs in the Qris configuration.
// it add 1 after calculation because BNC allowed merchant MID to create 1 static QRIS without storeID
// When merchant have 2 storeID, then allowed static QRIS is 3 (1 static QRIS from MID)
// If the Qris field is nil, it returns 0.
func (s *SetupPaymentMethodPartnerConfigRequest) GetMaxBNCQRStaticLimit() int {
	if s == nil || s.Qris == nil {
		return 0
	}

	if !strings.EqualFold(s.Qris.Acquirer, constant.BANK_ACQUIRER_BNC) {
		return 0
	}

	return len(s.Qris.AcquirerStoreIDs) + 1
}

type CardFundedPayoutPartnerConfig struct {
	CIT *SetupPaymentMethodPartnerConfigForCardObj `json:"-"`
	MIT *SetupPaymentMethodPartnerConfigForCardObj `json:"-"`
}

type CardAggregatorConfig struct {
	PFID   string `json:"pfID"`
	PFName string `json:"pfName"`
}

type SplitCardPaymentConfig struct {
	Enabled         bool                                  `json:"enabled" validate:"-"`
	ActiveProcessor string                                `json:"activeProcessor" validate:"required,oneof=MPGS CYBS"`
	Processors      map[string]CardPartnerProcessorConfig `json:"processors" validate:"required,min=1,dive,keys,oneof=MPGS CYBS,endkeys,required"`
}

type CardFundedPayoutConfig struct {
	Enabled         bool                                  `json:"enabled" validate:"-"`
	ActiveProcessor string                                `json:"activeProcessor" validate:"required,oneof=MPGS CYBS"`
	Processors      map[string]CardPartnerProcessorConfig `json:"processors" validate:"required,min=1,dive,keys,oneof=MPGS CYBS,endkeys,required"`
}

type VirtualTerminalConfig struct {
	AcquirerMerchantID string   `json:"acquirerMerchantId"`
	AllowedBinNumbers  []string `json:"allowedBinNumbers" validate:"omitempty,min=1,unique,dive,number|eq=ALL,required"`
	CardTypes          []string `json:"cardTypes" validate:"omitempty,min=1,unique,dive,required"`
	PrincipalAvailable []string `json:"principalAvailable" validate:"omitempty,min=1,unique,dive,required"`
}

type CardPartnerProcessorConfig struct {
	Limit float64 `json:"limit" validate:"required,min=1"`
}
