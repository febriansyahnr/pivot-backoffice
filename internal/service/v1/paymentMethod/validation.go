package paymentMethodService

import (
	"errors"
	"fmt"
	"slices"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConst "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/paymentMethod"
	paymentMethodModel "github.com/paper-indonesia/pivot-backoffice/internal/model/paymentMethod"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

const (
	minBINLength = 5
	maxBINLength = 11
)

type uniqueData struct {
	travelAgents map[string]bool
}
type validateFn func() error
type cardConfig = paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj

func setupValueForSpecificPaymentType(paymentMethodType string, config *model.SetupPaymentMethodPartnerConfigForCardObj) model.SetupPaymentMethodPartnerConfigForCardObj {
	if config == nil {
		return model.SetupPaymentMethodPartnerConfigForCardObj{}
	}
	switch paymentMethodType {
	case paymentConst.PAYMENT_METHOD_VIRTUAL_TERMINAL:
		if config.SupportedUseCase == nil {
			config.SupportedUseCase = &paymentMethodModel.CardSupportedUseCase{}
		}
		config.SupportedUseCase.AllowBypass3Ds = true
		config.SupportedUseCase.AllowForeignCard = true
	default:
		// Other payment methods have no set value
	}
	return *config
}

func validateCardPartnerConfig(paymentMethodType string, config model.SetupPaymentMethodPartnerConfigForCardObj, dupChecker *uniqueData, travelAgents config.MStrStr) error {
	// Validation process for OTA (Online Travel Agent).
	if config.TravelAgentCode == "" {
		// Non-OTA config can proceed to the next process.
	} else if config.SupportedUseCase == nil || !config.SupportedUseCase.AllowBypass3Ds || !config.SupportedUseCase.AllowForeignCard || len(config.SupportedUseCase.AllowedBinNumbers) == 0 {
		return errors.New("partner configuration for OTA must allow non-3ds transactions, foreign cards and include the allowed bin numbers")

	} else if _, found := travelAgents[config.TravelAgentCode]; !found {
		return fmt.Errorf("travel agent code %s not found", config.TravelAgentCode)

	} else if _, found := dupChecker.travelAgents[config.TravelAgentCode]; found {
		return fmt.Errorf("duplicate configuration found for travel agent code %s", config.TravelAgentCode)
	}

	if paymentMethodType == paymentConst.PAYMENT_METHOD_VIRTUAL_TERMINAL && config.TravelAgentCode == "" {
		return errors.New("travel agent code is required for the virtual terminal payment method")
	}

	// The list of allowed BIN numbers is currently only used for VCC Terminal transactions and validation processes in the portal backend.
	if config.SupportedUseCase != nil && len(config.SupportedUseCase.AllowedBinNumbers) > 0 {
		for _, binNumber := range config.SupportedUseCase.AllowedBinNumbers {
			if len(binNumber) < minBINLength || len(binNumber) > maxBINLength {
				return errors.New("invalid BIN number length")
			}
		}
	}
	if config.TravelAgentCode != "" {
		dupChecker.travelAgents[config.TravelAgentCode] = true
	}

	if shouldDisableBypass3DS(config) {
		return errors.New("acquirer merchant ID (CIT) must not enable 3DS bypass")
	}

	if shouldEnableBypass3DS(config) {
		return errors.New("acquirer merchant ID (CIT/MIT) must enable 3DS bypass")
	}

	shouldEnableExternal3ds := config.SplitPaymentType == constant.CardTransactionTypeCIT &&
		(config.SupportedUseCase == nil || !config.SupportedUseCase.AllowExternalThreeDs)
	if shouldEnableExternal3ds {
		return fmt.Errorf("acquirer merchant ID %s is required to enable external 3DS for specific use cases", config.AcquirerMerchantID)
	}

	// Next, you can add other validations here.
	return nil
}

func validateCardTransactionMIDPairConfig(label string, configs []cardConfig, keyFn func(cardConfig) string) error {
	pairs := map[string]*struct {
		pairName string
		count    int
	}{
		"MPGS-CIT": {"MPGS-MIT", 1},
		"MPGS-MIT": {"MPGS-CIT", 1},
		"CYBS-CIT": {"CYBS-MIT", 1},
		"CYBS-MIT": {"CYBS-CIT", 1},
	}

	for _, config := range configs {
		key := keyFn(config)
		if check, ok := pairs[key]; ok {
			idx := slices.IndexFunc(configs, func(c cardConfig) bool {
				return keyFn(c) == check.pairName && c.IsActive
			})
			if idx < 0 {
				return pkgErr.New(response.HttpErrRequest, fmt.Errorf("invalid MID configuration for %s. %s configuration not found", label, check.pairName))
			} else if check.count == 0 {
				return pkgErr.New(response.HttpErrRequest, fmt.Errorf("invalid MID configuration for %s: duplicate configuration for %s type", label, key))
			}
			check.count--
		}
	}
	return nil
}

func validateSplitCardPaymentConfig(partnerConfig *paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest) error {
	config := partnerConfig.SplitCardPaymentConfig

	partnerConfigExist := slices.ContainsFunc(partnerConfig.Items, func(p paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj) bool {
		return p.SplitPaymentType != ""
	})

	if config == nil && !partnerConfigExist {
		return nil

	} else if config == nil && partnerConfigExist {
		return pkgErr.New(response.HttpErrRequest, errors.New("invalid config: split card payment is not set, but CIT or MIT split types are defined in partner config"))

	} else if config != nil && !partnerConfigExist {
		return pkgErr.New(response.HttpErrRequest, errors.New("invalid config: split card payment is set, but CIT or MIT split types are missing in the partner config"))
	}

	if _, ok := config.Processors[config.ActiveProcessor]; !ok {
		return pkgErr.New(response.HttpErrRequest, errors.New("missing processor limit for active processor"))
	}

	partnerConfigExist = slices.ContainsFunc(partnerConfig.Items, func(p paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj) bool {
		return p.SplitPaymentType != "" && p.PartnerProcessor == config.ActiveProcessor && p.IsActive
	})
	if !partnerConfigExist {
		return pkgErr.New(response.HttpErrRequest, errors.New("missing partner config for active processor"))
	}
	return nil
}

func validateCardFundedPayoutConfig(partnerConfig *paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest) error {
	config := partnerConfig.CardFundedPayoutConfig

	partnerConfigExist := slices.ContainsFunc(partnerConfig.Items, func(p paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj) bool {
		return p.CardFundedPayoutType != ""
	})

	if config == nil && !partnerConfigExist {
		return nil

	} else if config == nil && partnerConfigExist {
		return pkgErr.New(response.HttpErrRequest, errors.New("invalid config: card-funded payout is not set, but CIT or MIT types are defined in partner config"))

	} else if config != nil && !partnerConfigExist {
		return pkgErr.New(response.HttpErrRequest, errors.New("invalid config: card-funded payout is set, but CIT or MIT types are missing in the partner config"))
	}

	if _, ok := config.Processors[config.ActiveProcessor]; !ok {
		return pkgErr.New(response.HttpErrRequest, errors.New("missing processor limit for active processor"))
	}

	partnerConfigExist = slices.ContainsFunc(partnerConfig.Items, func(p paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj) bool {
		return p.CardFundedPayoutType != "" && p.PartnerProcessor == config.ActiveProcessor && p.IsActive
	})
	if !partnerConfigExist {
		return pkgErr.New(response.HttpErrRequest, errors.New("missing partner config for active processor"))
	}
	return nil
}

func shouldDisableBypass3DS(cardCfg paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj) bool {
	isTypeCIT := cardCfg.CardFundedPayoutType == constant.CardTransactionTypeCIT ||
		cardCfg.RecurringType == constant.CardTransactionTypeCIT

	if !isTypeCIT {
		return false
	}

	return cardCfg.SupportedUseCase != nil && cardCfg.SupportedUseCase.AllowBypass3Ds
}

// split payment using 3ds for the parent payment
// bypass the 3ds on first payment due to resuming prev session
func shouldEnableBypass3DS(cardCfg paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj) bool {
	isRequiredType := cardCfg.CardFundedPayoutType == constant.CardTransactionTypeMIT ||
		cardCfg.RecurringType == constant.CardTransactionTypeMIT ||
		cardCfg.SplitPaymentType == constant.CardTransactionTypeMIT ||
		cardCfg.SplitPaymentType == constant.CardTransactionTypeCIT

	return isRequiredType && (cardCfg.SupportedUseCase == nil || !cardCfg.SupportedUseCase.AllowBypass3Ds)
}
