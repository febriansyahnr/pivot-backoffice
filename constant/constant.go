package constant

import "time"

type ContextKey string

const (
	// CtxRequestIdKey is the context key for request id
	CtxRequestIdKey ContextKey = "request_id"
	// CtxTraceIdKey is the context key for trace id
	CtxTraceIdKey ContextKey = "trace_id"
	// ctxUserInfoKey is the context key for user info from JWT
	CtxUserInfoKey ContextKey = "userInfo"
	// ctxMerchantIDKey is the context key for get merchant id
	CtxMerchantIDKey               ContextKey = "merchant_id"
	CtxSubMerchantIDKey            ContextKey = "sub_merchant_id"
	CtxMerchantInfo                ContextKey = "merchantInfo"
	CtxMerchantData                ContextKey = "merchant"
	CtxUserPINKey                  ContextKey = "internal_user_pin"
	CtxClientReqKey                ContextKey = "internal_client_request"
	CtxSnapApiName                 ContextKey = "internal_snap_api_name"
	CtxProcessorName               ContextKey = "payout_processor_name"
	CtxPaymentID                   ContextKey = "payment_id"
	CtxExposeUnmappingRequestError ContextKey = "expose_unmapping_request_error"
	CtxChangePaymentMethod         ContextKey = "internal_change_payment_method"
	CtxCurrentPaymentMethod        ContextKey = "internal_current_payment_method"
	CtxOutboundID                  ContextKey = "outbound_id"
	CtxSyncKey                     ContextKey = "internal_sync_key"
	CtxTimeLocation                ContextKey = "internal_time_location"
	CtxUseV2ErrorCode              ContextKey = "internal_use_v2_error_code"
	CtxUseErrorSource              ContextKey = "internal_use_error_source"
	CtxTest                        ContextKey = "internal_test"
	CtxTestCardNumber              ContextKey = "internal_test_card_number"
	CtxFeatureName                 ContextKey = "internal_feature_name"
	CtxErrorInfo                   ContextKey = "internal_error_info"
	CtxCustomErrorResponse         ContextKey = "internal_custom_error_response"

	CtxEntryPoint ContextKey = "entry_point"

	CtxParentMerchantId                 ContextKey = "internal_parent_merchant_id"
	CtxSetPendingTransaction            ContextKey = "internal_set_pending_transaction"
	CtxSetBypassBalanceCheckTransaction ContextKey = "internal_set_bypass_balance_check_transaction"
	CtxSetPendingSettlementTransaction  ContextKey = "internal_set_pending_settlement_transaction"

	// CtxDerivedMerchantID context for merchant that has parent as derived merchant.
	CtxDerivedMerchantID ContextKey = "internal_derived_merchant_id"

	// CtxMockInsufficientBalanceMerchant is the context key for mock insufficient balance merchant
	CtxMockInsufficientBalanceMerchant ContextKey = "mock_insufficient_balance_merchant"

	// ctxOTPKey is the context key for verify token otp process
	CtxTokenOTPKey ContextKey = "internal_auth_token_from_otp"

	// CtxAcceptLanguage is the context key for dictionary language
	CtxAcceptLanguage string = "accept_language"

	CtxPaymentSimulationMode  ContextKey = "internal_payment_simulation_mode"
	CtxPaymentSimulationToken ContextKey = "internal_payment_simulation_token"

	CtxUserAgentKey            ContextKey = "internal_user_agent"
	CtxUserDeviceIdentifierKey ContextKey = "internal_user_device_identifier"
	CtxUserIPKey               ContextKey = "internal_user_ip"
	CtxIsRemember              ContextKey = "internal_is_remember"
	CtxDisburesementType       ContextKey = "disbursement_type"
	CtxForceFailed             ContextKey = "force_failed"
	CtxAccountName             ContextKey = "internal_account_name"
	CtxTimeZone                ContextKey = "time_zone"
	CtxMessageId               ContextKey = "internal_message_id"
	CtxFromRetry               ContextKey = "from_retry"

	EnvironmentDevelopment = "development"
	EnvironmentLocal       = "local"
	EnvironmentStaging     = "staging"
	EnvironmentProduction  = "production"
	EnvironmentTest        = "test"

	LOGIN_EXPIRATION_DURATION   = time.Duration(24) * time.Hour
	REFRESH_EXPIRATION_DURATION = time.Duration(30) * 24 * time.Hour
	BLOCKED_DURATION            = time.Duration(3) * time.Hour

	MerchantAuthExpirationDuration = time.Duration(15) * time.Minute
)

const (
	DefaultPage               = 1
	DefaultPaginationPageSize = 20

	DefaultPageSize                           = 50
	DefaultPlatformSubMerchantBalancePageSize = 30

	DefaultMerchantFee = float64(2500)

	DefaultPhoneCountryCode = "+62"
)

const (
	ClientIdKey                  = "X-CLIENT-KEY"
	ClientSecretKey              = "X-CLIENT-SECRET"
	ContentTypeApplicationJson   = "application/json"
	ContentTypeMultipartFormData = "multipart/form-data"
	HeaderDeviceIdentifier       = "X-Device-Identifier"
	XIdentifierKey               = "X-Identifier"
	XIPInfoKey                   = "X-IP-Info"
	XIsRemember                  = "X-Is-Remember"
	XRequestIdKey                = "X-Request-Id"
	HeaderTimeZoneKey            = "Time-Zone"
	HeaderXMerchantID            = "X-MERCHANT-ID"
	HeaderToken                  = "Token"
	HeaderDataOrigin             = "Data-Origin"
)

const (
	DataOriginRaw       = "raw"
	DataOriginReporting = "reporting"
)

const (
	NonPaymentFeeConfigsFmt         = "backend-portal:non-payment-fee-configs:merchants:%s:%s" // $1 is merchant id and $2 referece
	CacheKeyFmtPayoutTransactionFee = "backend-portal:fees:payouts:%s:%s"                      // $1 merchant id $2 bank channel code
	CacheKeyFmtMerchantFeeCounter   = "backend-portal:merchant-fee-counter:%s:%s"              // $1 fee UUID $2 YYYY-MM
)

const (
	WithdrawalManual    = "MANUAL"
	WithdrawalAutomated = "AUTOMATED"

	AutoWithdrawalStateON  = "ON"
	AutoWithdrawalStateOFF = "OFF"

	WithdrawalDestBankTransfer    = "BANK_TRANSFER"
	WithdrawalDestBalanceTransfer = "BALANCE_TRANSFER"

	WithdrawalReasonScheduled       = "SCHEDULED"
	WithdrawalReasonDormantMerchant = "DORMANT_MERCHANT"
)

const (
	FeatureBalanceHistoryDashboard = "BALANCE_HISTORY_DASHBOARD"
	FeatureBalanceHistoryOpenApi   = "BALANCE_HISTORY_OPEN_API"

	FeatureConfirmPaymentCheckoutUI = "CONFIRM_PAYMENT_CHECKOUT_UI"
	FeatureConfirmOpenAPI           = "CONFIRM_PAYMENT_OPEN_API"
)

const (
	FileExtensionCsv = ".csv"
	FileSize5MB      = 5 * 1024 * 1024 // 5MB
)

const (
	DefaultConfig                         = "DEFAULT_CONFIG"
	DefaultPendingTransactionBackdateDays = 90
)
