package constant

import (
	"encoding/json"
	"slices"
	"time"

	"github.com/google/uuid"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
)

const (
	// Feature Flag Key
	FeatureFlagKeyForceAutoWithdrawalProcess                                   = "backend-portal-force-auto-withdrawal-process"
	FeatureFlagKeyBankTransferCreatePendingTrxIfNotExists                      = "backend-portal-bank-transfer-create-pending-trx-if-not-exists"
	FeatureFlagKeyDoNotApplyPayoutCutOffTime                                   = "backend-portal-do-not-apply-payout-cut-off-time"
	FeatureFlagKeyFlipProcessorUrl                                             = "backend-portal-flip-processor-url"
	FeatureFlagKeyDanaProcessorUrl                                             = "backend-portal-dana-processor-url"
	FeatureFlagKeyEnableMerchantRateLimitMiddleware                            = "backend-portal-merchant-rate-limit-middleware"
	FeatureFlagKeyEnableFrontEndAppVersion                                     = "backend-portal-front-end-app-version"
	FeatureFlagKeyEnableMerchantDefaultRateLimitMiddleware                     = "backend-portal-merchant-default-rate-limit-middleware"
	FeatureFlagKeyEnableAllowedBeneficiaryPayoutDefaultRule                    = "backend-portal-merchant-allowed-beneficiary-payout-default-rule"
	FeatureFlagKeyMerchantAllowedBeneficiaryPayoutCustomRule                   = "backend-portal-merchant-allowed-beneficiary-payout-custom-rule"
	FeatureFlagKeyMerchantAllowedExcludeBeneficiaryPayoutRules                 = "backend-portal-merchant-allowed-exclude-beneficiary-payout-rules"
	FeatureFlagKeyCustomContextTimeout                                         = "backend-portal-custom-context-timeout"
	FeatureFlagKeyHealthCheckCustomContextTimeout                              = "backend-portal-health-check-custom-context-timeout"
	FeatureFlagKeyDisbursementOverbookingBankCodeListViaFlip                   = "backend-portal-disbursement-overbooking-bankcode-list-via-flip"
	FeatureFlagKeyHttpRequestCacheDuration                                     = "backend-portal-http-request-cache-duration"
	FeatureFlagKeyFdsConfig                                                    = "backend-portal-fds-config"
	FeatureFlagFraudNetConfig                                                  = "backend-portal-fraud-net-config"
	FeatureFlagKeyUnifiedPaymentCustomerAndOrderObjectEligibleMerchant         = "backend-portal-unfied-payment-customer-order-object-eligible-merchant"
	FeatureFlagKeyUnifiedPaymentCardEncryptionWhitelistedMerchant              = "backend-portal-card-encryption-whitelisted-merchant"
	FeatureFlagKeyUnifiedPaymentClientReferenceSpecialCharsWhitelistedMerchant = "backend-portal-unified-payment-client-reference-special-chars-whitelisted-merchant"
	FeatureFlagAdvanceAiConfig                                                 = "backend-portal-advance-ai-config"
	FeatureFlagKeyDukcapilConfig                                               = "backend-portal-dukcapil-config"
	FeatureFlagPaymentMigrationV1toV2Enabled                                   = "backend-portal-payment-migration-v1-to-v2-enabled"
	FeatureFlagPlatformWhitelistOldResponseFormat                              = "backend-portal-platform-whitelist-old-response-format"
	FeatureFlagAccountOTPBypass                                                = "backend-portal-account-otp-bypass"
	FeatureFlagQrGenerateSeparateDb                                            = "backend-portal-qr-generate-separate-db"
	FeatureFlagQrMultiAcquirerRouting                                          = "backend-portal-qr-multi-acquirer-routing"
	FeatureFlagSnapMultiplePaymentDelegation                                   = "backend-portal-snap-multiple-payment-delegation"
	FeatureFlagExpireProcessedPaymentInMinute                                  = "backend-portal-expire-processed-payment-in-minute"
	FeatureFlagAccountInquiryMerchantNameUseRuneCheck                          = "backend-portal-account-inquiry-merchant-name-use-rune-check"
	FeatureFlagAccountInquiryIgnoreNamePrefix                                  = "backend-portal-account-inquiry-ignore-name-prefix"
	FeatureFlagKeyPaymentUIAuthCaptureBannerMessage                            = "backend-portal-payment-ui-auth-capture-banner-message"
	FeatureFlagKeyEnableMerchantCallbackViaWorkflow                            = "backend-portal-enable-merchant-callback-via-workflow"
	FeatureFlagKeyMerchantShouldEnforceNewMandatoryFields                      = "backend-portal-merchant-should-enforce-new-mandatory-fields"
	FeatureFlagMerchantExcludedSendCaptureHistory                              = "backend-portal-merchant-excluded-send-capture-history"
	FeatureFlagMerchantExcludedSendSurname                                     = "backend-portal-merchant-excluded-send-surname"
	FeatureFlagEwalletPaymentSimulationFlow                                    = "backend-portal-ewallet-payment-simulation-flow"
	FeatureFlagAccountInquiryDisplayVirtualAccountFlagForWhitelistedMerchant   = "backend-portal-account-inquiry-display-virtual-account-flag-for-whitelisted-merchant"
	FeatureFlagSinglePayoutCallbackWhitelistedMerchant                         = "backend-portal-single-payout-callback-whitelisted-merchant"
	FeatureFlagPaymentAutoInquiry                                              = "backend-portal-payment-auto-inquiry"
	FeatureFlagPayoutFDSResultAllowed                                          = "backend-portal-payout-fds-result-allowed"
	FeatureFlagPayoutExternalFDSEnabled                                        = "backend-portal-payout-external-fds-enabled"
	FeatureFlagCybersourceTestAmountWhitelist                                  = "backend-portal-cybersource-test-amount-whitelist"
	FeatureFlagEnableBalanceHistoryViaDataReporting                            = "backend-portal-enable-balance-history-via-data-reporting"
	FeatureFlagCardFundedPayoutManualProcessingAccounts                        = "backend-portal-card-funded-payout-manual-processing-accounts"
	FeatureFlagKeyBulkValidateWorkers                                          = "backend-portal-bulk-validate-workers"
	FeatureFlagRestrictPayoutToInternalVirtualAccounts                         = "backend-portal-restrict-payout-to-internal-virtual-accounts"

	// Feature Flag Targeting Queries
	FeatureFlagTargetQueryNameMerchantId       = "merchant_id"
	FeatureFlagTargetQueryNameEnv              = "environment"
	FeatureFlagTargetQueryNameEvent            = "event"
	FeatureFlagTargetQueryNameURLPath          = "url_path"
	FeatureFlagTargetQueryNameResult           = "result"
	FeatureFlagTargetQueryNameDependentService = "dependent_service"
	FeatureFlagTargetQueryNameBankCode         = "bank_code"
	FeatureFlagTargetQueryNameAccountNumber    = "account_number"
)

type FdsFeatureFlag struct {
	ScoreThreshold int64 `json:"scoreThreshold"`
	Timeout        int64 `json:"timeout"`
	BinLength      int64 `json:"binLength"`
	SendMidInfo    bool  `json:"sendMidInfo"`
}

type FraudNetFeatureFlag struct {
	BaseURL string `json:"baseUrl"`
}

type AdvanceAiFeatureFlag struct {
	BaseURL   string `json:"baseUrl"`
	JourneyID string `json:"journeyId"`
}

type MandatoryAddressEnforcementConfig struct {
	CutoffDate  string   `json:"cutoffDate"`
	MerchantIds []string `json:"merchantIds"`
}

type PaymentAutoInquiryConfig struct {
	CooldownSeconds int      `json:"cooldownSeconds"`
	EnabledMethods  []string `json:"enabledMethods"`
}

func GetFdsFeatureFlag(key string) *FdsFeatureFlag {
	flagEval := ffcontext.NewEvaluationContext(key)
	flagEval.AddCustomAttribute(FeatureFlagTargetQueryNameEnv, key)
	dataJson, err := ffclient.JSONVariation(FeatureFlagKeyFdsConfig, flagEval, nil)
	if err != nil || dataJson == nil {
		return nil
	}

	dataBytes, err := json.Marshal(dataJson)
	if err != nil {
		return nil
	}

	var result FdsFeatureFlag
	err = json.Unmarshal(dataBytes, &result)
	if err != nil {
		return nil
	}

	return &result
}

func GetFraudNetFeatureFlag(key string) *FraudNetFeatureFlag {
	flagEval := ffcontext.NewEvaluationContext(key)
	flagEval.AddCustomAttribute(FeatureFlagTargetQueryNameEnv, key)
	dataJson, err := ffclient.JSONVariation(FeatureFlagFraudNetConfig, flagEval, nil)
	if err != nil || dataJson == nil {
		return nil
	}

	dataBytes, err := json.Marshal(dataJson)
	if err != nil {
		return nil
	}

	var result FraudNetFeatureFlag
	err = json.Unmarshal(dataBytes, &result)
	if err != nil {
		return nil
	}

	return &result
}

func GetAdvanceAiFeatureFlag(key string) *AdvanceAiFeatureFlag {
	flagEval := ffcontext.NewEvaluationContext(key)
	flagEval.AddCustomAttribute(FeatureFlagTargetQueryNameEnv, key)
	dataJson, err := ffclient.JSONVariation(FeatureFlagAdvanceAiConfig, flagEval, nil)
	if err != nil || dataJson == nil {
		return nil
	}

	dataBytes, err := json.Marshal(dataJson)
	if err != nil {
		return nil
	}

	var result AdvanceAiFeatureFlag
	err = json.Unmarshal(dataBytes, &result)
	if err != nil {
		return nil
	}

	return &result
}

func IsQrGenerateSeparateDbFeatureEnabled(merchantID string) bool {
	flagEval := ffcontext.NewEvaluationContext(merchantID)
	flagEval.AddCustomAttribute(FeatureFlagTargetQueryNameMerchantId, merchantID)
	dataJson, err := ffclient.BoolVariation(FeatureFlagQrGenerateSeparateDb, flagEval, false)
	if err != nil {
		return false
	}

	return dataJson
}

func IsQrMultiAcquirerRoutingEnabled(merchantID string) bool {
	flagEval := ffcontext.NewEvaluationContext(merchantID)
	flagEval.AddCustomAttribute(FeatureFlagTargetQueryNameMerchantId, merchantID)
	enabled, err := ffclient.BoolVariation(FeatureFlagQrMultiAcquirerRouting, flagEval, false)
	if err != nil {
		return false
	}

	return enabled
}

func IsAccountInquiryMerchantNameUseRuneCheck(merchantID string) bool {
	flagEval := ffcontext.NewEvaluationContext(merchantID)
	flagEval.AddCustomAttribute(FeatureFlagTargetQueryNameMerchantId, merchantID)
	enabled, err := ffclient.BoolVariation(FeatureFlagAccountInquiryMerchantNameUseRuneCheck, flagEval, false)
	if err != nil {
		return false
	}

	return enabled
}

func IsAccountInquiryIgnoreNamePrefix(merchantID string) bool {
	flagEval := ffcontext.NewEvaluationContext(merchantID)
	flagEval.AddCustomAttribute(FeatureFlagTargetQueryNameMerchantId, merchantID)
	enabled, err := ffclient.BoolVariation(FeatureFlagAccountInquiryIgnoreNamePrefix, flagEval, false)
	if err != nil {
		return false
	}

	return enabled
}

func IsMerchantCallbackWorkflowEnabled(environment string) bool {
	fcontext := ffcontext.NewEvaluationContext(environment)
	fcontext.AddCustomAttribute(FeatureFlagTargetQueryNameEnv, environment)

	result, _ := ffclient.BoolVariation(FeatureFlagKeyEnableMerchantCallbackViaWorkflow, fcontext, false)
	return result
}

func ShouldEnforceStandardizedAddress(merchantID string, merchantCreatedAt time.Time) bool {
	flagEval := ffcontext.NewEvaluationContext(merchantID)
	flagEval.AddCustomAttribute(FeatureFlagTargetQueryNameMerchantId, merchantID)

	dataJson, err := ffclient.JSONVariation(FeatureFlagKeyMerchantShouldEnforceNewMandatoryFields, flagEval, nil)
	if err != nil || dataJson == nil {
		return false
	}

	dataBytes, err := json.Marshal(dataJson)
	if err != nil {
		return false
	}

	var config MandatoryAddressEnforcementConfig
	err = json.Unmarshal(dataBytes, &config)
	if err != nil {
		return false
	}

	if len(config.MerchantIds) > 0 && slices.Contains(config.MerchantIds, merchantID) {
		return true
	}

	cutoffDate, err := time.Parse(time.RFC3339, config.CutoffDate)
	if err != nil {
		return false
	}

	return merchantCreatedAt.After(cutoffDate) || merchantCreatedAt.Equal(cutoffDate)
}

func IsAccountInquiryVirtualAccountFlagDisplayedForMerchant(merchantID string) bool {
	flagEval := ffcontext.NewEvaluationContext(merchantID)
	flagEval.AddCustomAttribute(FeatureFlagTargetQueryNameMerchantId, merchantID)
	enabled, err := ffclient.BoolVariation(FeatureFlagAccountInquiryDisplayVirtualAccountFlagForWhitelistedMerchant, flagEval, false)
	if err != nil {
		return false
	}

	return enabled
}

func IsSinglePayoutCallbackWhitelistedForMerchant(merchantID string) bool {
	flagEval := ffcontext.NewEvaluationContext(merchantID)
	flagEval.AddCustomAttribute(FeatureFlagTargetQueryNameMerchantId, merchantID)
	enabled, err := ffclient.BoolVariation(FeatureFlagSinglePayoutCallbackWhitelistedMerchant, flagEval, false)
	if err != nil {
		return false
	}

	return enabled
}

func GetPaymentAutoInquiryCooldownSeconds(environment string) *int {
	flagEval := ffcontext.NewEvaluationContext(environment)
	flagEval.AddCustomAttribute(FeatureFlagTargetQueryNameEnv, environment)

	configData, err := ffclient.JSONVariation(FeatureFlagPaymentAutoInquiry, flagEval, nil)
	if err != nil || configData == nil {
		return nil
	}

	if cooldown, ok := configData["cooldownSeconds"].(float64); ok {
		cooldownInt := int(cooldown)
		return &cooldownInt
	}

	return nil
}

func IsPayoutFDSResultAllowed(result string) bool {
	eval := ffcontext.NewEvaluationContext(result)
	eval.AddCustomAttribute(FeatureFlagTargetQueryNameResult, result)

	evalResult, _ := ffclient.BoolVariation(FeatureFlagPayoutFDSResultAllowed, eval, false)
	return evalResult
}

func IsPayoutExternalFDSEnabled(env string) bool {
	eval := ffcontext.NewEvaluationContext(env)
	eval.AddCustomAttribute(FeatureFlagTargetQueryNameEnv, env)

	evalResult, _ := ffclient.BoolVariation(FeatureFlagPayoutExternalFDSEnabled, eval, false)
	return evalResult
}

func IsEnableBalanceHistoryViaDataReporting(merchantID string) bool {
	eval := ffcontext.NewEvaluationContext(merchantID)
	eval.AddCustomAttribute(FeatureFlagTargetQueryNameMerchantId, merchantID)

	result, _ := ffclient.BoolVariation(FeatureFlagEnableBalanceHistoryViaDataReporting, eval, false)
	return result
}

func IsCardFundedPayoutManualProcessingAccount(bankCode, accountNumber string) bool {
	eval := ffcontext.NewEvaluationContext(uuid.NewString())
	eval.AddCustomAttribute(FeatureFlagTargetQueryNameBankCode, bankCode)
	eval.AddCustomAttribute(FeatureFlagTargetQueryNameAccountNumber, accountNumber)

	result, _ := ffclient.BoolVariation(FeatureFlagCardFundedPayoutManualProcessingAccounts, eval, false)
	return result
}

func IsPayoutToVirtualAccountAllowed(bankCode, accountNumber string) (allowed bool) {
	attr := ffcontext.NewEvaluationContext(bankCode + "_" + accountNumber)
	attr.AddCustomAttribute(FeatureFlagTargetQueryNameBankCode, bankCode)
	attr.AddCustomAttribute(FeatureFlagTargetQueryNameAccountNumber, accountNumber)

	allowed, err := ffclient.BoolVariation(FeatureFlagRestrictPayoutToInternalVirtualAccounts, attr, true)
	if err != nil {
		return true
	}
	return allowed
}
