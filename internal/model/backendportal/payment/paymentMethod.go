package paymentModel

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	constant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	paymentMethodModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/paymentMethod"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/unifiedPayment"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/jmoiron/sqlx/types"
)

type PaymentMethod struct {
	UUID                    string                         `json:"uuid" db:"uuid"`
	Type                    string                         `json:"type" db:"type"`
	Subtype                 string                         `json:"subtype" db:"sub_type"`
	Category                string                         `json:"category" db:"category"`
	Name                    string                         `json:"name" db:"name"`
	Description             *string                        `json:"description" db:"description"`
	Logo                    *string                        `json:"logo" db:"logo"`
	Acquirer                string                         `json:"acquirer" db:"acquirer"`
	BankName                *string                        `json:"bankName" db:"bank_name"`
	Instructions            *string                        `json:"instructions" db:"instructions"`
	Processor               string                         `json:"processor" db:"processor"`
	ActivationMethod        string                         `json:"activationMethod" db:"activation_method"`
	CountryOfOperation      string                         `json:"countryOfOperation" db:"country_of_operation"`
	SupportedCurrency       string                         `json:"supportedCurrency" db:"supported_currency"`
	Config                  types.NullJSONText             `json:"-" db:"config"`
	ConfigObj               *PaymentMethodConfig           `json:"-" db:"-"`
	RequiredDocuments       types.NullJSONText             `json:"-" db:"required_document"`
	RequiredDocumentObjects *[]PaymentMethodDocumentObject `json:"requiredDocuments" db:"-"`

	CreatedAt time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time  `json:"updatedAt" db:"updated_at"`
	DeletedAt *time.Time `json:"deletedAt" db:"deleted_at"`
}

// UnmarshalConfigObj unmarshals the config field into a PaymentMethodConfig object,
// if the config field is not valid, the ConfigObj field will be set to nil
func (pm *PaymentMethod) UnmarshalConfigObj() {
	if pm.Config.Valid {
		err := json.Unmarshal([]byte(pm.Config.String()), &pm.ConfigObj)
		if err != nil {
			pm.ConfigObj = nil
		}
	}
}

type PaymentMethodWithPivot struct {
	PaymentMethod

	MerchantID        string                             `json:"merchantId" db:"merchant_id"`
	IsActive          bool                               `json:"isActive" db:"is_active"`
	ActivationStatus  string                             `json:"activationStatus" db:"activation_status"`
	ChannelType       string                             `json:"channelType" db:"channel_type"`
	MerchantConfig    types.NullJSONText                 `json:"-" db:"merchant_config"`
	MerchantConfigObj *PaymentMethodMerchantConfigObject `json:"merchantConfig" db:"-"`
	IsDerivedMerchant *bool                              `json:"isDerivedMerchant,omitempty" db:"-"`
	QRPayments        []StaticQRPaymentItem              `json:"QRPayments,omitempty" db:"-"`
}

type PaymentMethodMerchantConfigObject struct {
	ChannelConfig          *paymentMethodModel.SetupPaymentMethodChannelConfigRequest `json:"channelConfig,omitempty"`
	SplitCardPaymentConfig *paymentMethodModel.SplitCardPaymentConfig                 `json:"splitCardPaymentConfig,omitempty"`
	CardFundedPayoutConfig *paymentMethodModel.CardFundedPayoutConfig                 `json:"cardFundedPayoutConfig,omitempty"`
	VirtualTerminalConfig  *paymentMethodModel.VirtualTerminalConfig                  `json:"virtualTerminalConfig,omitempty"`
	PartnerConfig          *paymentMethodModel.SetupPaymentMethodPartnerConfigRequest `json:"partnerConfig,omitempty"`
}

type PaymentMethodMerchantPartnerVAConfig struct {
	BINPrefix       string `json:"binPrefix,omitempty"`
	Type            string `json:"type,omitempty" validate:"oneof=OPEN_STATIC CLOSED_DYNAMIC CLOSED_STATIC"`
	IntegrationType string `json:"integrationType,omitempty" validate:"oneof=SERVER CLIENT"`
}

type PaymentMethodDocumentObject struct {
	Name   string `json:"name"`
	Format string `json:"format"` // data / document
}

type PlatformPaymentMethodResponse struct {
	UUID        string  `json:"uuid" db:"uuid"`
	Name        string  `json:"name" db:"name"`
	Description *string `json:"description" db:"description"`
	Logo        *string `json:"logo" db:"logo"`
	Acquirer    string  `json:"acquirer" db:"acquirer"`
	BankName    *string `json:"bankName" db:"bank_name"`
}

type StaticQRPaymentItem struct {
	MerchantID               string    `json:"merchantId"`
	PaymentSessionID         string    `json:"paymentSessionId"`
	PaymentClientReferenceID string    `json:"paymentClientReferenceId"`
	StoreID                  string    `json:"storeId"`
	IsDerived                bool      `json:"isDerived"`
	CreatedAt                time.Time `json:"createdAt"`
	ExpiredAt                time.Time `json:"expiredAt"`
}

func (p *PaymentMethodWithPivot) ToPlatformResponseModel() *PlatformPaymentMethodResponse {
	return &PlatformPaymentMethodResponse{
		UUID:        p.UUID,
		Name:        p.Name,
		Description: p.Description,
		Logo:        p.Logo,
		Acquirer:    p.Acquirer,
		BankName:    p.BankName,
	}
}

func (p PaymentMethodWithPivot) IsCardPartnerConfigFound() bool {
	return p.MerchantConfigObj != nil &&
		p.MerchantConfigObj.PartnerConfig != nil &&
		p.MerchantConfigObj.PartnerConfig.Card != nil &&
		len(p.MerchantConfigObj.PartnerConfig.Card.Items) > 0
}

func (p PaymentMethodWithPivot) GetCardPartnerConfigForOnlineTravelAgent(def config.VCCTerminalDefaultConfig) map[string]*paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj {
	travelAgents := map[string]*paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
		c.DefaultConfig: {
			CardTypes:          def.CardTypes,
			AcquirerMerchantID: def.AcquirerMerchantID,
			PrincipalAvailable: def.PrincipalAvailable,
			SupportedUseCase: &paymentMethodModel.CardSupportedUseCase{
				AllowedBinNumbers: util.ValueOfPtr(def.AllowedBinNumbers),
			},
		},
	}
	if def.AllowedBinNumbers == nil {
		travelAgents[c.DefaultConfig].SupportedUseCase.AllowedBinNumbers = []string{"ALL"}
	}

	if p.MerchantConfigObj == nil {
		return travelAgents
	}

	if cnf := p.MerchantConfigObj.VirtualTerminalConfig; cnf != nil {
		travelAgents[c.DefaultConfig].AcquirerMerchantID = cnf.AcquirerMerchantID
		travelAgents[c.DefaultConfig].SupportedUseCase.AllowedBinNumbers = cnf.AllowedBinNumbers
		travelAgents[c.DefaultConfig].CardTypes = cnf.CardTypes
		travelAgents[c.DefaultConfig].PrincipalAvailable = cnf.PrincipalAvailable
	}

	if !p.IsCardPartnerConfigFound() {
		return travelAgents
	}

	for _, config := range p.MerchantConfigObj.PartnerConfig.Card.Items {
		if config.TravelAgentCode == "" || !config.IsActive {
			continue
		}
		travelAgents[config.TravelAgentCode] = &config
	}
	return travelAgents
}

func (p PaymentMethodWithPivot) EnableSplitCardPayment() bool {
	return p.MerchantConfigObj != nil &&
		p.MerchantConfigObj.SplitCardPaymentConfig != nil &&
		p.MerchantConfigObj.SplitCardPaymentConfig.Enabled
}

func (p PaymentMethodWithPivot) EnableCardFundedPayout() bool {
	return p.MerchantConfigObj != nil &&
		p.MerchantConfigObj.CardFundedPayoutConfig != nil &&
		p.MerchantConfigObj.CardFundedPayoutConfig.Enabled
}

func ValidatePaymentMethod(paymentMethod string) error {
	switch strings.ToUpper(paymentMethod) {
	case constant.PAYMENT_METHOD_VIRTUAL_ACCOUNT, constant.PAYMENT_METHOD_CREDIT_CARD, constant.PAYMENT_METHOD_QRIS, constant.PAYMENT_METHOD_EWALLET, constant.PAYMENT_METHOD_INSTALLMENT:
		return nil
	default:
		return c.ErrInvalidPaymentMethod
	}
}

type PaymentMethodConfig struct {
	ExpiryConfig PaymentMethodExpiryConfig `json:"expiryConfig"`
}

type PaymentMethodExpiryConfig struct {
	Duration int    `json:"duration"`
	Unit     string `json:"unit"`
}

func (p *PaymentMethodExpiryConfig) ToDateTime() time.Time {
	switch p.Unit {
	case constant.UnifiedPaymentExpiryUnitMinutes:
		return time.Now().Add(time.Duration(p.Duration) * time.Minute)
	case constant.UnifiedPaymentExpiryUnitHours:
		return time.Now().Add(time.Duration(p.Duration) * time.Hour)
	case constant.UnifiedPaymentExpiryUnitDays:
		return time.Now().Add(time.Duration(p.Duration) * 24 * time.Hour)
	default:
		return time.Now().Add(time.Duration(p.Duration) * time.Second)
	}
}

func (p *PaymentMethodMerchantConfigObject) GetCardAcquirer(mid string) string {
	if p.PartnerConfig == nil || p.PartnerConfig.Card == nil {
		return ""
	}

	for _, cfg := range p.PartnerConfig.Card.Items {
		if cfg.AcquirerMerchantID == mid {
			return cfg.Acquirer
		}
	}

	return ""
}

type PaymentRequestExpiryValidation struct {
	Method                string // payment method: VIRTUAL_ACCOUNT or QRIS
	MerchantID            string
	Request               *PaymentRequest
	PaymentMethod         *PaymentMethodWithPivot
	UnifiedPaymentRequest *unifiedPaymentModel.CreateUnifiedPaymentSessionRequest
}
