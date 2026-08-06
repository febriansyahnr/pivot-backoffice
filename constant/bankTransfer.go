package constant

const (
	SnapCoreBankTransferStatusInitiated = "INITIATED"
	SnapCoreBankTransferStatusPending   = "PENDING"
	SnapCoreBankTransferStatusFailed    = "FAILED"
	SnapCoreBankTransferStatusSuccess   = "SUCCESS"
)

const (
	SnapCoreResponseCodeInsufficientFund          = "4031714" // use pattern instead
	SnapCoreResponseCodeInterbankInsufficientFund = "4031814" // use pattern instead
	SnapCoreResponseCodeInsufficientFundPattern   = "^403..14$"
	SnapCoreResponseCodeInvalidFieldPattern       = "^403..11$"

	SnapCoreResponseCodeDormantAccountPattern     = "^403..09$"
	SnapCoreResponseCodeInactiveAccountPattern    = "^403..18$"
	SnapCoreSuspectedFraudCodePattern             = "^403..03$"
	SnapCoreActivityCountLimitExceededCodePattern = "^403..04$"
	SnapCoreDoNotHonorCodePattern                 = "^403..05$"
	SnapCoreResponseCodeInvalidAccountPattern     = "^404..11$"
	SnapCoreResponseCodeConflictPattern           = "^409..00$"
	SnapCoreResponseCodeToManyRequestPattern      = "^429..00$"
	SnapCoreResponseCodeRequestInProgress         = "^202..00$"
	SnapCoreResponseCodeSuccessPattern            = "^200..00$"

	SnapCoreResponseCodeGeneralErrorPattern  = "^500..00$"
	SnapCoreResponseCodeInternalErrorPattern = "^500..01$"
	SnapCoreResponseCodeTimeoutPattern       = "^504..00$"

	// Used for inquiry status
	SnapCoreResponseCodeVANotFoundPattern = "^404"
	SnapCoreResponseCodeVAConflictPattern = "^409"
)

const (
	SnapCoreResponseInactiveAccountMessage = "Inactive account"
	SnapCoreResponseDormantAccountMessage  = "Dormant account"
	SnapCoreResponseInvalidAccountMessage  = "Invalid account"
)

// FLIP Response Code
const (
	FlipBankTransferStatusPending   = "PENDING"
	FlipBankTransferStatusCancelled = "CANCELLED"
	FlipBankTransferStatusDone      = "DONE"

	FlipAccountInquiryStatusPending          = "PENDING"
	FlipAccountInquiryStatusSuccess          = "SUCCESS"
	FlipAccountInquiryStatusInvalid          = "INVALID_ACCOUNT_NUMBER"
	FlipAccountInquiryStatusSuspectedAccount = "SUSPECTED_ACCOUNT"
	FlipAccountInquiryStatusBlacklisted      = "BLACK_LISTED"

	FlipReasonInactiveAccount                         = "INACTIVE_ACCOUNT"
	FlipReasonNotRegisteredAccount                    = "NOT_REGISTERED_ACCOUNT"
	FlipReasonCantReceiveTransfer                     = "CANT_RECEIVE_TRANSFER"
	FlipReasonIntermittenDisturbanceOnBeneficiaryBank = "INTERMITTENT_DISTURBANCE_ON_BENEFICIARY_BANK"
	FlipReasonBeneficiaryAccountNotVerified           = "BENEFICIARY_ACCOUNT_NOT_VERIFIED"
	FlipReasonInsufficientBalance                     = "INSUFFICIENT_BALANCE"
	FlipReasonTransactionAmountLimit                  = "EXCEEDED_TRANSACTION_AMOUNT_LIMIT"
	FlipReasonInternalError                           = "INTERNAL_ERROR"
	FlipReasonExceedAmountLimit                       = "EXCEED_AMOUNT_LIMIT"
	FlipReasonDormantAccount                          = "DORMANT_ACCOUNT"
	FlipReasonInvalidAccount                          = "INVALID_ACCOUNT"
	FlipReasonInvalidBill                             = "INVALID_BILL"
	FlipReasonInvalidAmount                           = "INVALID_AMOUNT"
	FlipReasonPaidBill                                = "PAID_BILL"
	FlipReasonExpiredBill                             = "EXPIRED_BILL"

	FlipGatewayTimeoutFieldCode              = "998"
	FlipUndefinedError                       = "999"
	FlipInvalidMandatoryFieldCode            = "1001"
	FlipInvalidFieldCode                     = "1002"
	FlipAcceptedNumericCode                  = "1020"
	FlipInvalidMinimumAmountCode             = "1021"
	FlipInvalidMaximumAmountCode             = "1022"
	FlipInvalidMaxLengthFieldCode            = "1024"
	FlipInvalidBankAccountVACode             = "1025"
	FlipSuspectedFraudAccountCode            = "1026"
	FlipAccountClosedCode                    = "1027"
	FlipInvalidPaginationLengthCode          = "1032"
	FlipInvalidBankCode                      = "1033"
	FlipInvalidCountryCode                   = "1034"
	FlipInsufficientBalanceCode              = "1035"
	FlipInvalidBankAccountCode               = "1036"
	FlipInvalidCityCode                      = "1038"
	FlipInvalidDateFormatCode                = "1039"
	FlipInvalidAttributeCode                 = "1041"
	FlipInvalidIdempotencyKeyCode            = "1042"
	FlipInvalidBillTitleCode                 = "1043"
	FlipInvalidMaxLengthBeneficiaryEmailCode = "1070"
	FlipInvalidBeneficiaryEmailCode          = "1071"
	FlipDisbursementIDNotFoundCode           = "1072"
	FlipDisbursementIdempotencyNotFoundCode  = "1073"
	FlipDailyTransactionLimitCode            = "1074"
	FlipPageNotFoundCode                     = "1075"
	FlipMaxLimitActiveTrxCode                = "1080"
	FlipBeneficiaryBankDisturbanceCode       = "1088"
	FlipInvalidAccountToFlipCode             = "1089"
	FlipAgentKYCNotApprovedCode              = "1090"
	FlipAgentStatusNotActiveCode             = "1091"
	FlipAgentNotAllowedToUpdateCode          = "1092"
	FlipUnprocessedTransactionTimeOutCode    = "1093"
	FlipStaleRequestCode                     = "1094"
	FlipInvalidXTimestampCode                = "1095"
	FlipInvalidRegexFieldCode                = "2001"
	FlipDuplicateFieldCode                   = "2002"
	FlipInvalidRegexAlphanumericCode         = "2003"
	FlipInvalidRegexNonAlphanumericCode      = "2004"
	FlipInvalidAmountByBankCode              = "2005"
)

const (
	DanaEscrowBalanceSuccessCode          = "00000000"
	DanaEscrowBalanceSystemErrorCode      = "00000900"
	DanaEscrowBalanceParamIllegalCode     = "00000004"
	DanaEscrowBalanceParamMissingCode     = "00000002"
	DanaEscrowBalanceInvalidSignatureCode = "00000007"
	DanaEscrowBalanceKeyNotFoundCode      = "00000008"
	DanaEscrowBalanceNoInterfaceDefCode   = "00000013"
	DanaEscrowBalanceApiInvalidCode       = "00000014"
	DanaEscrowBalanceMsgParseErrorCode    = "00000015"
	DanaEscrowBalanceOAuthFailedCode      = "00000016"
	DanaEscrowBalanceFunctionMismatchCode = "00000017"
	DanaEscrowBalanceVerifySecretFailCode = "12014151"
	DanaEscrowBalanceForbiddenAccessCode  = "12014152"
	DanaEscrowBalanceUnknownClientCode    = "12014155"
	DanaEscrowBalanceInvalidClientStatus  = "12014156"
	DanaEscrowBalanceAccessTokenMissing   = "12005015"
	DanaEscrowBalanceAccessTokenExpired   = "12005016"
	DanaEscrowBalanceAuthorizationExpired = "12014200"
	DanaEscrowBalanceTooManyRequestsCode  = "00000026"
	DanaEscrowBalanceMerchantNotExistCode = "12158006901"
)
