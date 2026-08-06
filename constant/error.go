package constant

import (
	"errors"
	"fmt"
)

var (
	ErrSomeErrorForUnitTest   = errors.New("some error")
	ErrInvalidRequestPayload  = errors.New("invalid request payload")
	ErrNoRowsAffected         = errors.New("no rows affected")
	ErrInvalidValidation      = errors.New("invalid validation")
	ErrInvalidReference       = errors.New("invalid reference")
	ErrFailedValidation       = errors.New("failed validation")
	ErrNegativeValue          = errors.New("value cannot be negative")
	ErrNoRowsData             = errors.New("no rows data")
	ErrOpenFileReader         = errors.New("failed to open file reader")
	ErrInvalidPage            = errors.New("invalid page number")
	ErrInvalidPerPage         = errors.New("invalid per page number")
	ErrInvalidSize            = errors.New("invalid size")
	ErrInvalidFetchAll        = errors.New("invalid fetch all")
	ErrInvalidShowDeactivated = errors.New("invalid show deactivated")
	ErrInvalidActiveOnly      = errors.New("invalid active only")
	ErrUnmarshalProto         = errors.New("unmarshal protobuf: cannot parse invalid wire-format data")
	ErrInvalidSortOrder       = errors.New("invalid sort order")
	ErrInvalidAccess          = errors.New("invalid access")
	ErrInvalidPath            = errors.New("invalid path URL")
	ErrInvalidUnmarshalJSON   = errors.New("invalid unmarshal JSON")
	ErrInvalidMarshalData     = errors.New("invalid marshal data")
	ErrInvalidFileExtension   = errors.New("invalid file extension")
	ErrFileTooLarge           = errors.New("file size too large")

	ErrMerchantNotAllowedPerformAction = errors.New("merchant not allowed to perform this action")
	ErrMerchantNotAllowedUpdateStatus  = errors.New("merchant status cannot be updated")

	ErrPaymentAmountNotMatch                  = errors.New("payment amount not match")
	ErrPaymentMethodNotFound                  = errors.New("payment method is not found")
	ErrMerchantPaymentMethodNotFound          = errors.New("merchant payment method is not found")
	ErrIncorrectPaymentMethod                 = errors.New("incorrect payment method")
	ErrDoNotApplySplitRouteInFacilitatorModel = errors.New("split route config do not apply to facilitator payment methods")
	ErrInvalidExpiryDate                      = errors.New("invalid expiry date")
	ErrExceedMaxExpiryDate                    = "the request was invalid, or an error occurred in downstream provider. %s max expiry time is %s"
	ErrExceedMaxExpiryDateCheck               = "max expiry time"

	ErrFindMerchant                    = errors.New("error find merchant")
	ErrFindParentMerchant              = errors.New("error find parent merchant")
	ErrMerchantNotFound                = errors.New("merchant not found")
	ErrMerchantShortNameInvalid        = errors.New("invalid merchant short name")
	ErrMerchantReservedShortName       = errors.New("merchant short name already reserved")
	ErrPaymentNotFound                 = errors.New("payment not found")
	ErrPaymentInFinalState             = errors.New("payment is in final state")
	ErrPaymentCaptureNotFound          = errors.New("payment capture not found")
	ErrGetPaymentLedger                = errors.New("error get payment ledger")
	ErrMerchantBalanceNotFound         = errors.New("merchant balance not found")
	ErrMerchantTopUpReferenceNotFound  = errors.New("merchant top up reference not found")
	ErrDisbursementNotFound            = errors.New("disbursement not found")
	ErrDoubleDisbursementIndication    = errors.New("double disbursement indication")
	ErrCheckUser                       = errors.New("error check user")
	ErrUserNotFound                    = errors.New("user not found")
	ErrUserAlreadyExists               = errors.New("user already exists")
	ErrEmailAlreadyRegistered          = errors.New("email already registered")
	ErrCreateUser                      = errors.New("error create user")
	ErrUserUnauthorized                = errors.New("user unauthorized")
	ErrBulkDisbursementNotFound        = errors.New("bulk disbursement not found")
	ErrPayoutIsNotSingle               = errors.New("payout is not single")
	ErrPermissionNotFound              = errors.New("permission not found")
	ErrMenuNotFound                    = errors.New("menu not found")
	ErrMenuOrPermissionNotRegistered   = errors.New("menu or permission not registered")
	ErrGetMenuPermission               = errors.New("error get menu permission")
	ErrDeleteRoleMenuPermissions       = errors.New("error delete role menu permissions")
	ErrUpdateRoleData                  = errors.New("error update role data")
	ErrRoleNotFound                    = errors.New("role not found")
	ErrCreateRole                      = errors.New("error create role")
	ErrDataNotFound                    = errors.New("data not found")
	ErrDistrictNotFound                = errors.New("district id not found")
	ErrResourceLocked                  = errors.New("please wait for a moment")
	ErrAssignUserToRole                = errors.New("error assign user to role")
	ErrSendUserInvitation              = errors.New("error send invitation to user")
	ErrInvalidUserListSortColumn       = errors.New("invalid user list sort column")
	ErrRetryDisbursementIsNotAllowed   = errors.New("retry disbursement is not allowed")
	ErrTransactionAlreadySucceeded     = errors.New("transaction already succeeded")
	ErrInvalidProcessorReference       = errors.New("invalid processor reference")
	ErrForbiddenToChangeRole           = errors.New("forbidden to change role")
	ErrBankTransferChannelNotFound     = errors.New("channel code not found")
	ErrInvalidDisbursementAmount       = errors.New("invalid disbursement amount")
	ErrBeneficiaryNameLengthExceeded   = errors.New("beneficiary name length exceeded")
	ErrMerchantLogoAndLogoFileRequired = errors.New("either 'logo' or 'logoFile' must be provided")

	ErrCannotModifyDefaultRole = errors.New("can't modify default role")
	ErrRoleCannotBeDeleted     = errors.New("role cannot be deleted")

	ErrDisbursementStatusAlreadyApproved    = errors.New("disbursement status already approved")
	ErrDisbursementStatusHasNotBeenApproved = errors.New("disbursement status has not been approved")
	ErrDisbursementIsBeingProcessed         = errors.New("disbursement is being processed")
	ErrDisbursementIsCancelled              = errors.New("disbursement is cancelled")
	ErrDisbursementCannotBeCancelled        = errors.New("disbursement cannot be cancelled")
	ErrInvalidDisbursementType              = errors.New("disbursement type not valid")
	ErrInvalidDisbursementApprovalStatus    = errors.New("disbursement approval status not valid")
	ErrInvalidDisbursementPaymentStatus     = errors.New("disbursement payment status not valid")
	ErrInvalidDisbursementListSortColumn    = errors.New("disbursement sort column is not valid")
	ErrInvalidDisbursementCancelRequest     = errors.New("invalid cancel disbursement request")
	ErrGetDisbursementList                  = errors.New("error get disbursement list")
	ErrGetWithdrawalList                    = errors.New("error get withdrawal list")
	ErrGetChargeList                        = errors.New("error get charge list")
	ErrGetTopUpList                         = errors.New("error get top up list")
	ErrTransactionAlreadyInFinalStatus      = errors.New("transaction is already in a final status")
	ErrPaymentAlreadyInFinalStatus          = errors.New("payment is already in a final status")
	ErrInvestigationNotEnabled              = errors.New("investigation flow is not enabled for this merchant")
	ErrInvestigationAlreadyFinalized        = errors.New("investigation is already finalized and cannot be modified")
	ErrInvestigationNotFound                = errors.New("payment is not under investigation")
	ErrInvalidInvestigationStatus           = errors.New("invalid investigation status, must be INVESTIGATION_SUCCESS or INVESTIGATION_FAILED")
	ErrInvestigationNotesExceedLimit        = errors.New("investigation notes exceed 200 character limit")
	ErrBankConfirmedSuccess                 = errors.New("payment has already been confirmed successful by the bank")
	ErrBankConfirmedFailed                  = errors.New("payment has already failed based on bank confirmation")
	ErrBankConfirmedExpired                 = errors.New("payment has already expired")
	ErrBankInquiryFailed                    = errors.New("unable to verify payment status with bank")
	ErrBeneficiaryLimitRestrictions         = errors.New("payout request declined due to beneficiary limit restrictions")
	ErrBankTransferStillInPending           = errors.New("bank transfer still in pending")
	ErrProofOfPaymentRateLimitExceeded      = fmt.Errorf("Proof of payment upload rate limit exceeded")

	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrValidateBalance     = errors.New("failed validate balance")
	ErrReduceBalance       = errors.New("failed reduce balance")
	ErrTimeout             = errors.New("request timeout")

	ErrHeaderColumnDoesNotMatchWithTemplate = errors.New("header column doesn't match with template")

	ErrForbiddenAccess = errors.New("forbidden access")
	ErrInvalidToken    = errors.New("invalid token")
	ErrKeyAuthRequired = errors.New("key auth required")
	ErrInvalidKey      = errors.New("invalid key")
	ErrTokenIsExpired  = errors.New("token is expired")

	ErrIncorrectMerchantID                 = errors.New("incorrect merchant id")
	ErrMerchantIDNotValid                  = errors.New("merchant id is not valid")
	ErrMerchantIsNotMatch                  = errors.New("merchant id is not match")
	ErrMerchantFeeIDNotValid               = errors.New("merchant fee id is not valid")
	ErrPayoutsInProcess                    = errors.New("payouts are in process")
	ErrDisbursementReferenceIdAlreadyExist = errors.New("reference ID already exist")
	ErrDuplicateDisbursementReferenceId    = errors.New("duplicate reference ID")
	ErrUpdateMerchant                      = errors.New("error update merchant")
	ErrDuplicateData                       = errors.New("duplicate value found")
	ErrFeeTieringSequence                  = errors.New("invalid fee tiering sequence")
	ErrFeeTieringRange                     = errors.New("invalid fee tiering range")

	ErrMaxBulkDisbursementRequestAllowed = errors.New("max bulk disbursement request is 1000")

	// PIN
	ErrPINNotCreatedYet = errors.New("pin not created yet")
	ErrInvalidPIN       = errors.New("invalid pin")
	ErrRequiredPIN      = errors.New("request PIN is required")

	ErrInvalidPINLength       = errors.New("PIN must be 6 digits")
	ErrIdenticalPIN           = errors.New("PIN cannot consist of identical digits")
	ErrSequentialPIN          = errors.New("PIN cannot be sequential")
	ErrFailedCheckPINLimit    = errors.New("failed to check PIN, too many wrong attempts")
	ErrFailedToChangePINLimit = errors.New("failed to change PIN, too many wrong attempts")

	// Password
	ErrInvalidPassword        = errors.New("invalid password")
	ErrUserAlreadyActivated   = errors.New("user already activated")
	ErrFailedCheckPassword    = errors.New("failed to check password, too many wrong attempts")
	ErrFailedToChangePassword = errors.New("failed to change password, too many wrong attempts")

	ErrIdempotencyViolation   = errors.New("idempotency violation")
	ErrIdempotencyKeyRequired = errors.New("Idempotent-Key is a required header")
	ErrIdempotencyKeyExists   = errors.New("Idempotent-Key already exist")

	ErrDisbursementNotSuccessYet = errors.New("disbursement status has not been successful")
	ErrPaymentNotSuccessYet      = errors.New("payment status has not been successful")
	ErrFailedToGenerateReceipt   = errors.New("failed to generate receipt")
	ErrFailedToGenerateExcel     = errors.New("failed to generate excel")

	ErrRateLimiterExceedFailedAttempts = errors.New("exceed failed attempts")
	ErrRateLimiterFailedValidate       = errors.New("failed validate input attempts")
	ErrRateLimiterFailedUpdateAttempts = errors.New("failed update fail attempts")
	ErrRateLimiterFailedResetAttempts  = errors.New("failed reset attempts")

	ErrIdIsRequired                    = errors.New("id is required")
	ErrDocumentIdIsRequired            = errors.New("document id is required")
	ErrBlockedUnblockedMerchantUseCase = errors.New("failed block/unblock merchant operation")
	ErrFailedValidateUseCase           = errors.New("failed validate use case")
	ErrForbiddenUseCase                = errors.New("merchant operation is not allowed")
	ErrForbiddenUseCaseNotExist        = errors.New("use case is not exist")

	ErrInvalidEmailOrPassword = errors.New("incorrect email or password.")
	ErrInquiryIdNotFound      = errors.New("inquiry not found")
	ErrUserInvitedStatus      = errors.New("your account is not yet active. please check your email and follow the instructions to complete your onboarding")

	ErrInvalidAccount         = errors.New("invalid account")
	ErrInactiveAccount        = errors.New("inactive account")
	ErrDormantAccount         = errors.New("dormant account")
	ErrDuplicateExternailId   = errors.New(ErrMsgXExternalIdAlreadyExists)
	ErrDuplicateRequestId     = errors.New("X-Request-Id is already exists")
	ErrInvalidRequestIdFormat = errors.New("invalid X-Request-Id format")
	ErrRequestInProgress      = errors.New("request in progress")

	// Account
	ErrFailedCreateMerchantAccount = errors.New("failed create merchant accounts")
	ErrFailedCreateAccount         = errors.New("failed create account")
	ErrInvalidUserType             = errors.New("invalid user type")
	ErrInvalidUsecase              = errors.New("invalid usecase")
	ErrFindAccount                 = errors.New("error find account")
	ErrFailedBulkCreateAccounts    = errors.New("failed bulk create accounts")
	ErrBeneficiaryAlreadyExists    = errors.New("beneficiary already exists")
	ErrGetAccounts                 = errors.New("error get accounts")
	ErrGetAccount                  = errors.New("error get merchant account")

	// Account Transaction
	ErrInvalidType                        = errors.New("invalid type")
	ErrInvalidChannel                     = errors.New("invalid channel")
	ErrInvalidStatus                      = errors.New("invalid status")
	ErrInvalidTransactionTimestamp        = errors.New("invalid transaction timestamp")
	ErrInvalidAdditionalInfo              = errors.New("invalid additional info")
	ErrInvalidTransferType                = errors.New("invalid transfer type")
	ErrCreateLedgerEntry                  = errors.New("error create ledger entry")
	ErrStoreLedgerEntry                   = errors.New("error store ledger entry")
	ErrFailedValidateP2PTransfer          = errors.New("failed validate p2p transfer")
	ErrFailedValidatePayOut               = errors.New("failed validate payout")
	ErrFailedValidateCharge               = errors.New("failed validate charge")
	ErrInvalidUsecaseAndAccountType       = errors.New("invalid usecase and account type")
	ErrInvalidUsecaseAndParentAccountType = errors.New("invalid usecase and parent account type")
	ErrUpdateLedgerEntry                  = errors.New("error update ledger entry")

	// Ledger
	ErrGetBalance                  = errors.New("error get balance")
	ErrGetBulkBalance              = errors.New("error get bulk balances")
	ErrMissingRecipientAccountID   = errors.New("missing recipient account id")
	ErrMissingSenderAccountID      = errors.New("missing sender account id")
	ErrMissingParentAccountID      = errors.New("missing parent account id")
	ErrSenderSameWithRecipient     = errors.New("recipient cannot be same with sender")
	ErrRecipientAccountNotFound    = errors.New("recipient account not found")
	ErrParentAccountNotFound       = errors.New("parent account not found")
	ErrFeeRecipientAccountNotFound = errors.New("fee recipient account not found")
	ErrFeeRecipientIsNotMerchant   = errors.New("fee recipient account is not merchant")
	ErrSenderAccountNotFound       = errors.New("sender account not found")
	ErrInvalidReferenceID          = errors.New("invalid reference id")
	ErrFailedUpdateTransactions    = errors.New("failed update transactions")
	ErrUpdateTransactions          = errors.New("error update transactions")
	ErrAccountNotFound             = errors.New("account not found")
	ErrAccountAlreadyActive        = errors.New("account already active")
	ErrAccountAlreadyInactive      = errors.New("account already inactive")
	ErrAccountNotSpecified         = errors.New("must specify accounts")
	ErrGetLedgerRecords            = errors.New("error get ledger records")
	ErrInvalidDateRange            = errors.New("invalid date range")
	ErrLedgerDetailNotFound        = errors.New("ledger not found")
	ErrGetLedgerDetail             = errors.New("error get ledger detail")
	ErrMissingFeeRecipient         = errors.New("missing fee recipient")
	ErrNegativeFee                 = errors.New("negative fee is not allowed")
	ErrPayInFeeBiggerThanAmount    = errors.New("pay in fee is bigger than incoming amount")
	ErrNotAllowedCancelTransaction = errors.New("not allowed to cancel transaction")
	ErrDailyLimitReached           = errors.New("you have exceeded your daily transaction limit")
	ErrForbiddenChangeKYCStatus    = errors.New("forbidden to change KYC status")
	ErrForbiddenKYCStatusAccess    = errors.New("forbidden KYC status access ")
	ErrMerchantShouldKYC           = errors.New("merchant should KYC first")
	ErrUpdatePaymentLedger         = errors.New("error update payment ledger")
	ErrInvokeClientWebhook         = errors.New("callback delivery to the client URL failed")
	ErrCallbackURLNotConfigured    = errors.New("callback URL is not configured")
	ErrInvalidBatchPayoutItem      = errors.New("invalid batch payout items")

	// MerchantAuth
	ErrGenerateMerchantAuth        = errors.New("error generate merchant auth")
	ErrValidateMerchantAuth        = errors.New("error validate merchant auth")
	ErrMismatchMerchantAuth        = errors.New("access token is not belong to merchant")
	ErrExpiredMerchantAuth         = errors.New("access token has expired")
	ErrValidateRequestSignature    = errors.New("error validate request signature")
	ErrGenerateRequestSignature    = errors.New("error generate request signature")
	ErrInvalidRequestSignature     = errors.New("invalid request signature")
	ErrMalformedRequestBodyPayload = errors.New("malformed request body payload")

	// Customer
	ErrCustomerAlreadyExists    = errors.New("customer with phone number already exists")
	ErrPhoneNumberAlreadyExists = errors.New("phone number already exist")
	ErrCustomerNotFound         = errors.New("customer not found")
	ErrCustomerIDRequired       = errors.New("customer id is required")
	ErrMerchantIDRequired       = errors.New("merchant id is required")
	ErrBlockReasonRequired      = errors.New("blockReason is required when isBlocked is true")
	ErrCustomerIsBlocked        = errors.New("customer is blocked and cannot perform this action")
	ErrDatabaseGetCustomer      = errors.New("error when get customer data")
	ErrDatabaseCreateCustomer   = errors.New("error when create customer data")
	ErrDatabaseUpdateCustomer   = errors.New("error when update customer data")
	ErrDatabaseDeleteCustomer   = errors.New("error when delete customer data")

	// Creditcard
	ErrCreditcardPaymentMethodNotFound             = errors.New("CREDIT_CARD payment method not found")
	ErrCreditcardReferenceIdAlreadyExist           = errors.New("payment with reference ID already exist")
	ErrCreditcardReferenceIdNotFound               = errors.New("payment with reference ID not found")
	ErrCreditcardNotFound                          = errors.New("payment with UUID not found")
	ErrCreditcardReferenceIdHasBeenProcessed       = errors.New("payment with reference id has been processed")
	ErrCreditcardReferenceIdHasBeenExpired         = errors.New("payment with reference id has been expired")
	ErrCreditcardMinAmount                         = errors.New("minimum amount is 10000")
	ErrCreditcardInvalidAuthenticationMethod       = errors.New("invalid authentication method")
	ErrCreditcardInvalidUUID                       = errors.New("invalid merchant uuid")
	ErrMerchantIdIsRequired                        = errors.New("merchant id is required")
	ErrWhenUpdateCreditcardMetaData                = errors.New("error when update creditcard metadata")
	ErrWhenGetOpenAPIGetPaymentDetailByID          = errors.New("error when get open api payment detail by uuid")
	ErrWhenGetInternalGetPaymentDetailByID         = errors.New("error when get internal payment detail by uuid")
	ErrCreditcardPaymentNotSuccess                 = errors.New("payment status is not success")
	ErrCreditcardPaymentAlreadyVoid                = errors.New("payment already void")
	ErrCreditcardPaymentIsUnpaid                   = errors.New("payment is currently unpaid")
	ErrCardEncryptionIsNotEnabledForMerchant       = errors.New("card encryption is not enabled for this merchant")
	ErrClientReferenceIDMustBeInAlphanumericFormat = errors.New("client reference ID must be in alphanumeric format")

	// Sub Merchant
	ErrSubMerchantNotFound             = errors.New("sub merchant not found")
	ErrFailedValidateSubMerchantParent = errors.New("fail validate sub merchant")
	ErrIncorrectSubMerchantParent      = errors.New("incorrect submerchant parent")
	ErrIncorrectSubMerchant            = errors.New("incorrect submerchant")
	ErrRequiredSubmerchantId           = errors.New("submerchant id is required")
	ErrMissingSubMerchantId            = errors.New("missing submerchant id")
	ErrNotAllowedToCreateSubMerchant   = errors.New("not allowed to create sub merchant")
	ErrIncorrectKYCType                = errors.New("incorrect kyc type")
	ErrBulkCreateSubMerchantNoData     = errors.New("no submerchants data")

	// Error Date Format & Value
	ErrInvalidStartDateFmt  = errors.New("invalid start date format")
	ErrInvalidEndDateFmt    = errors.New("invalid end date format")
	ErrFilterDateInput      = errors.New("invalid filter date value")
	ErrDateRangeExceedLimit = errors.New("date range exceeds limits")

	// Payment method
	ErrGetPayments                     = errors.New("error get payments")
	ErrPaymentMethodAlreadyActive      = errors.New("payment method is already active")
	ErrPaymentMethodAlreadyInactive    = errors.New("payment method is already inactive")
	ErrInvalidPaymentMethod            = errors.New("invalid payment method")
	ErrInvalidPaymentStatus            = errors.New("invalid payment status")
	ErrInvalidPaymentHistorySortColumn = errors.New("invalid payment history sort column")
	ErrGetMerchantPaymentMethod        = errors.New("error get merchant payment methods")

	// Payment
	ErrPaymentIDNotValid                      = errors.New("payment id is not valid")
	ErrSplitRoutingPaymentNotValid            = errors.New("split routing payment id is not valid")
	ErrPaymentDoesNotHaveSplitRouting         = errors.New("payment does not have split routing")
	ErrPaymentSplitRoutingDestinationNotFound = errors.New("the payment split routing destination is not found")
	ErrPaymentSplitRoutingIsNotProcessed      = errors.New("the payment split routing has not been processed to the destination yet")
	ErrPaymentChargeNotFound                  = errors.New("payment charge not found")
	ErrPaymentChargeNotSettled                = errors.New("payment charge is not success yet")
	ErrUpdatePayment                          = errors.New("error update payment")
	ErrPaymentChargeIsOnHold                  = errors.New("payment charge is currently being on hold")

	ErrStatusNotAllowed = errors.New("status not allowed")

	// QRIS
	ErrInvalidQrType                   = errors.New("invalid qrType")
	ErrQrRegistrationIsNotCompleted    = errors.New("QR registration is not completed")
	ErrQrisRegistrationAlreadyExists   = errors.New("QRIS registration already exists")
	ErrQrisInvalidPartnerConfigRequest = errors.New("invalid partner config request")

	// User
	ErrNeed2FAChallengeForLogin = errors.New("need 2fa challenge for login")
	ErrDeviceIdentifierRequired = errors.New("device identifier is required")
	ErrLoginSessionTerminated   = errors.New("login session is terminated due to login on another device")
	ErrLoginRoleSessionChanged  = errors.New("login session is terminated due to user role was changed")
	ErrBlockedTooManyAttempts   = errors.New("user blocked too many login attempts")
	ErrBlockedByFDS             = errors.New("transaction blocked due to risk detected by the FDS")

	// XB
	ErrInvalidSourceCurrency      = errors.New("invalid source currency")
	ErrInvalidDestinationCurrency = errors.New("invalid destination currency")
	ErrPayoutIsNotFound           = errors.New("payout is not found")
	ErrPayoutAlreadyExpired       = errors.New("payout is expired")
	ErrInvalidPayoutId            = errors.New("invalid payout id")
	ErrPayoutStatusNotRFI         = errors.New("payout status is not RFI")
	ErrGetPayoutById              = errors.New("error processing payout by id")

	// Transfer
	ErrInvalidAmount              = errors.New("invalid amount")
	ErrInvalidMerchantId          = errors.New("invalid merchant id value")
	ErrExecuteTransfer            = errors.New("error execute transfer")
	ErrSameMerchant               = errors.New("cannot transfer to the same merchant")
	ErrGetTransferById            = errors.New("error get transfer by id")
	ErrGetTransferList            = errors.New("error get transfer list")
	ErrInvalidId                  = errors.New("invalid id")
	ErrCheckReference             = errors.New("error check reference")
	ErrReferenceIdExist           = errors.New("reference id already exists")
	ErrTransferNotFound           = errors.New("transfer detail not found")
	ErrCreateTransfer             = errors.New("error create transfer")
	ErrUpdateTransfer             = errors.New("error update transfer")
	ErrReverseTransfer            = errors.New("error reverse transfer")
	ErrReverseMerchantPlatformFee = errors.New("error reverse merchant platform fee")
	ErrIncorrectTransferType      = errors.New("incorrect transfer type")
	ErrInvalidTransferStatus      = errors.New("invalid transfer status")
	ErrInvalidTransferSortColumn  = errors.New("invalid transfer column")
	ErrInvalidStartDateValue      = errors.New("invalid start date column")
	ErrInternalServerForUser      = errors.New("an error occurred on the server. please try again later")

	// Date Range Report
	ErrDateRangeLimitDaysFmt       = "date range exceeds the maximum limit of %d days"            // $1 Days
	ErrDateRangeLimitLastMonthsFmt = "date range exceeds the maximum limit of the last %d months" // $1 Months

	// Platform
	ErrGetSubMerchantList               = errors.New("error get submerchant list")
	ErrGetSubMerchantBalance            = errors.New("error get submerchant balance")
	ErrGetSubMerchantUserList           = errors.New("error get submerchant user list")
	ErrBuildSubMerchantUserListResponse = errors.New("error build submerchant user list response")

	// Bank Account
	ErrBankAccountNotFound = errors.New("bank account not found")

	// Merchants Fee
	ErrGetMerchantFee = errors.New("error get merchant fee")

	// Unified Payment
	ErrConfirmShouldChoosePaymentMethod                             = errors.New("the confirm request should have a chosen payment method")
	ErrMerchantNotRegisteredQR                                      = errors.New("merchant not registered qr")
	ErrNotAllowedToConfirmPaymentSession                            = errors.New("not allowed to confirm payment session")
	ErrRequestIsNotCompatibleWithSnapVersion                        = errors.New("request is not compatible with SNAP version")
	ErrClientReferenceIDAlreadyExist                                = errors.New("client reference id already exist")
	ErrCannotCancelQRISPayment                                      = errors.New("could not cancel previous QRIS payment")
	ErrNotAllowedToUpdatePayment                                    = errors.New("not allowed to update payment")
	ErrPaymentMethodNotAllowed                                      = errors.New("payment method not allowed")
	ErrCustomerInformationConflict                                  = errors.New("customer information conflict between customerId and CustomerInformation")
	ErrPaymentAmountRequired                                        = errors.New("amount is required")
	ErrPaymentBelowMinAmount                                        = errors.New("amount value below the minimum")
	ErrPaymentAboveMaxAmount                                        = errors.New("amount value above the maximum")
	ErrPaymentStaticQrAmountMustBeZero                              = errors.New("QRIS with paymentType = MULTIPLE must not include an amount. Remove the amount value from the request.")
	ErrPaymentStaticQrAcquirerNotSupported                          = errors.New("QR acquirer is not supported for static payment")
	ErrPaymentStaticQrReachMaxActivePayment                         = errors.New("The number of static QRIS you've created has reached the allowed limit. Cannot create a new QRIS.")
	ErrPaymentTypeMultipleNotAllowedForNonAPI                       = errors.New("paymentType 'MULTIPLE' is only allowed when mode is set to 'API'")
	ErrPaymentTypeMultipleNotAllowedForThisMethod                   = errors.New("paymentType 'MULTIPLE' is not allowed for this payment method type")
	ErrPaymentTypeMultipleNotAllowedAutoConfirmFalse                = errors.New("paymentType 'MULTIPLE' requires autoConfirm to be set to true")
	ErrPaymentMethodIsNotAllowedToSavedFutureUse                    = errors.New("payment method type is not allowed to save for future use")
	ErrPaymentMethodIsNotAllowedToShowSavedPayment                  = errors.New("payment method type is not allowed to show saved payment")
	ErrPaymentMethodIsNotAllowedToSetThreeDsMethod                  = errors.New("payment method type is not allowed to set 3ds method")
	ErrPaymentMethodIsNotAllowedToSetExpirationMode                 = errors.New("payment method type is not allowed to set expiration mode")
	ErrPaymentDoesNotAllowedToBypass3Ds                             = errors.New("payment does not allow bypassing 3DS")
	ErrExternalThreeDsNotEnabled                                    = errors.New("external 3DS is not enabled for this merchant")
	ErrEWalletShopeePayExceedMaxExpiryTime                          = errors.New("shopeepay e-wallet max expiry time is 5 days")
	ErrEWalletDanaExceedMaxExpiryTime                               = errors.New("dana e-wallet max expiry time is 30 minutes")
	ErrPaymentMethodInvalidChannelType                              = errors.New("invalid channel type value")
	ErrProcessingConfigInvalidMID                                   = errors.New("invalid MID or MID tag not found")
	ErrCaptureMethodMustBeManual                                    = errors.New("capture method must be manual")
	ErrChargeStatusMustBeWaitingForCapture                          = errors.New("charge status must be WAITING_FOR_CAPTURE")
	ErrPaymentMethodTypeIsNotCard                                   = errors.New("payment method type is not CARD")
	ErrCurrencyIsNotMatch                                           = errors.New("currency is not match")
	ErrCaptureAmountExceedAuthorizedAmount                          = errors.New("capture amount exceeds authorized amount")
	ErrProcessorCaptureStatusNotSuccess                             = errors.New("processor capture status is not success")
	ErrAmountNotPermittedToUseDecimal                               = errors.New("amount value is not permitted to use decimal format")
	ErrCaptureIsBeingProcessed                                      = errors.New("payment capture is being processed")
	ErrAnotherCaptureInProgress                                     = errors.New("another capture is in progress for this payment")
	ErrInvalidCardNumber                                            = errors.New("invalid card number format")
	ErrFailedToDecryptCardPayment                                   = errors.New("failed to decrypt card payment")
	ErrInvalidCardInformation                                       = errors.New("invalid card information")
	ErrForeignCardBillingInformationMissing                         = errors.New("billingInformation is required for foreign card transactions")
	ErrPostalCodeTooLong                                            = errors.New("postalCode must not exceed 10 characters")
	ErrPostalCodeInvalidFormat                                      = errors.New("postalCode must contain only alphanumeric characters and hyphens")
	ErrForeignCardNotAllowed                                        = errors.New("The request was invalid because foreign cards are not allowed for this merchant.")
	ErrUnsupportedCardPrincipal                                     = errors.New("The request was invalid because the card principal is not supported for this merchant.")
	ErrUnifiedPaymentMetadataSizeLimitExceeded                      = errors.New("metadata size limit exceeded")
	ErrUnifiedPaymentModeMustBeAPIForThreeDsMethodExternal          = errors.New("mode must be API when threeDsMethod is EXTERNAL")
	ErrUnifiedPaymentAutoConfirmMustBeFalseForThreeDsMethodExternal = errors.New("autoConfirm must be false when threeDsMethod is EXTERNAL")
	ErrUnifiedPaymentRequireThreeDsInfoForThreeDsMethodExternal     = errors.New("threeDsInfo is required for external 3DS flow")
	ErrUnifiedPaymentRequireSelectMIDForThreeDsMethodExternal       = errors.New("bankMerchantId is required for external 3DS flow")
	ErrUnifiedPaymentInvalidThreeDsInfoFormat                       = errors.New("invalid threeDsInfo format")
	ErrUnifiedPayment3DsAuthenticationNotSuccessful                 = errors.New("3DS authentication was not successful")
	ErrUnifiedPaymentInvalid3DsAuthenticationResult                 = errors.New("invalid 3DS authentication result")
	ErrUnifiedPaymentThreeDsMethodExternalOnlySupportForCard        = errors.New("threeDsMethod EXTERNAL is only supported for CARD payment method")
	ErrUnifiedPaymentRedirectUrlRequiredWhenBypassStatusPage        = errors.New("redirectUrl is required when bypassStatusPage is true")
	ErrGetUnifiedPaymentSessionDetail                               = errors.New("error when retrieve unified payment session detail")
	ErrGetUnifiedPaymentSessionList                                 = errors.New("error when retrieve unified payment session list")
	ErrInvalidUnifiedPaymentSessionListDataType                     = errors.New("invalid unified payment session list data type")
	ErrUpdateUnifiedPaymentSessionLedger                            = errors.New("error update ledger detail")
	ErrAcquireLockUnifiedPaymentNotification                        = errors.New("error acquire unified payment notification lock")
	ErrInquiryEwalletPaymentStatus                                  = errors.New("error inquiry ewallet payment")

	// Payment Link
	ErrPaymentLinkMinAmount = fmt.Errorf("min amount is %d", DashboardPaymentLinkMinAmount)
	ErrPaymentLinkMaxAmount = fmt.Errorf("max amount is %d", DashboardPaymentLinkMaxAmount)

	// ShortLink
	ErrCreateShortLink              = errors.New("error when create short link")
	ErrGetShortLink                 = errors.New("error when retrieve link")
	ErrShortLinkNotFound            = errors.New("link not found")
	ErrShortLinkExpired             = errors.New("link expired")
	ErrShortLinkDestinationRequired = errors.New("destination url is required")

	// FDS
	ErrRuleEvaluationsNotFound = errors.New("rule evaluation not found")
	ErrFraudRulesNotFound      = errors.New("fraud rule not found")
	ErrFraudRuleWeight         = errors.New("weight must be between 0 and 1")
	ErrGetFraudRuleList        = errors.New("error when get fraud rule list")
	ErrGetFraudRuleDetail      = errors.New("error when get fraud rule detail")
	ErrDeleteFraudRule         = errors.New("error when delete fraud rule")
	ErrCreateFraudRule         = errors.New("error when create fraud rule")
	ErrUpdateFraudRule         = errors.New("error when update fraud rule")
	ErrFdsTimeout              = errors.New("error fds timeout")
	ErrFdsUpdateTransaction    = errors.New("error when fds update transaction")

	// Vendor
	ErrVendorNotFound             = errors.New("vendor not found")
	ErrVendorNameAlreadyExists    = errors.New("vendor name already exists")
	ErrVendorNotActive            = errors.New("vendor status must be ACTIVE")
	ErrInvalidMerchantID          = errors.New("invalid merchant id")
	ErrInvalidAvgMonthlyTpvAmount = errors.New("invalid avg monthly tpv amount")

	// Payout Manual Processing Account
	ErrPayoutManualProcessingAccountAlreadyExists = errors.New("payout manual processing account already exists")
	ErrPayoutManualProcessingAccountNotFound      = errors.New("payout manual processing account not found")
	ErrGetPayoutManualProcessingAccountList       = errors.New("error when getting payout manual processing account list")
	ErrInvalidPayoutManualProcessingAccountStatus = errors.New("invalid status, must be ACTIVE or INACTIVE")

	// TNC (Terms and Conditions)
	ErrMerchantAlreadySignedTNC = errors.New("merchant already signed the active tnc version")
	ErrGetTNCSigningHistory     = errors.New("error when getting tnc signing history")
	ErrNoActiveTNCVersion       = errors.New("no active tnc version")
	ErrTNCVersionAlreadyExists  = errors.New("tnc version already exists")
	ErrTNCVersionIsLocked       = errors.New("tnc version already exists")
	ErrTNCVersionNotFound       = errors.New("tnc version not found")
	ErrGetTNCVersionList        = errors.New("error get tnc version list")
	ErrSubmerchantCannotSignTNC = errors.New("submerchant cannot sign tnc")

	// AML
	ErrProviderNotFound         = errors.New("provider not found")
	ErrAmlCheck                 = errors.New("error aml check")
	ErrAmlInquiry               = errors.New("error aml inquiry")
	ErrJourneyIDRequired        = errors.New("journeyId is required")
	ErrProviderRequired         = errors.New("provider is required")
	ErrAmlTransactionIDNotExist = errors.New("transaction id not exist")

	// Refund
	ErrRefundNotFound                                    = errors.New("refund not found")
	ErrClientReferenceIdIsExist                          = errors.New("client reference id already exist")
	ErrPaymentAlreadyRefunded                            = errors.New("payment already refunded")
	ErrFailedToRefundPayment                             = errors.New("failed to refund payment")
	ErrRefundTransferDestinationNotAvailable             = errors.New("refund transfer destination not available")
	ErrRefundNotInPendingStatus                          = errors.New("refund not in pending status")
	ErrRefundIsBeingProcessed                            = errors.New("refund is being processed")
	ErrRefundAlreadyProcessed                            = errors.New("refund already processed")
	ErrRefundAmountExceedPaymentCharge                   = errors.New("refund amount must not exceed the original payment charge amount")
	ErrRefundNotAllowedForPaymentMethodFacilitatorConfig = errors.New("refund not allowed for non facilitator payment method configuration")
	ErrRefundIncorrectRequestMethodForFacilitator        = errors.New("refund incorrect request method for facilitator")
	ErrRefundPartialIsNotYetAvailable                    = errors.New("partial refund isn't yet available, please try again after 24 hours from the original payment")
	ErrRefundIsNotYetAvailable                           = errors.New("refund isn't yet available, please try again after 24 hours from the original payment")
	ErrRefundFindExisting                                = errors.New("error find existing refunds for the payment")
	ErrRefundPaymentProcess                              = errors.New("error process payment refund")

	ErrProcessorNotRegistered = errors.New("processor not registered")

	ErrMessageForApiErrorSimulation = errors.New("message for api error simulation")

	// Countries
	ErrGetAllCountries  = errors.New("error when get all countries")
	ErrGetCountryByCode = errors.New("error when get country by code")

	// Industries
	ErrInvalidIndustryID      = errors.New("invalid industry id")
	ErrGetIndustryByID        = errors.New("error when get industry by id")
	ErrIndustryNotFound       = errors.New("industry not found")
	ErrDuplicateIndustry      = errors.New("industry with this parent and child already exists")
	ErrCreateIndustry         = errors.New("failed to create industry")
	ErrUpdateIndustry         = errors.New("failed to update industry")
	ErrDeleteIndustry         = errors.New("failed to delete industry")
	ErrIndustryInUse          = errors.New("cannot delete industry that is in use by merchants")
	ErrInvalidIndustryRisk    = errors.New("invalid risk level, must be one of: Low, Medium, High")
	ErrIndustryIDRequired     = errors.New("industry id is required")
	ErrParentIndustryRequired = errors.New("parent industry is required")
	ErrChildIndustryRequired  = errors.New("child industry is required")
	ErrMCCRequired            = errors.New("mcc is required")
	ErrCommonMCCRequired      = errors.New("common mcc is required")

	// Recon
	ErrReconTransactionTypeInvalid = errors.New("recon transaction type invalid")

	// Dukcapil
	ErrSetupDukcapilConfig       = errors.New("error setup dukcapil config")
	ErrDukcapilInvalidIdentity   = errors.New("invalid dukcapil identity")
	ErrEmptyDukcapilResponse     = errors.New("empty dukcapil response")
	ErrMalformedDukcapilMetadata = errors.New("malformed dukcapil metadata")

	// Withdrawal
	ErrInvalidBalanceValue = errors.New("invalid balance value")

	// Adjustment
	ErrAdjustmentInvalidBalanceType = errors.New("invalid balance type")

	// Merchant BOD
	ErrInvalidMerchantBODPosition   = errors.New("invalid position value")
	ErrMerchantBODMandatoryIdentity = errors.New("position require identity file & number")
	ErrMerchantBODMandatoryShares   = errors.New("mandatory shares value")
	ErrMerchantBODInvalidShares     = errors.New("invalid shares value")

	ErrEmptyProcessorResponse = errors.New("empty processor response")

	// VA
	ErrBlockVA = errors.New("error when block VA")

	// Partner Configs
	ErrPartnerConfigNotFound = errors.New("partner configuration was not found")

	// Network Token
	ErrMissingCardOnFile                                    = errors.New("card on file is required")
	ErrMissingNetworkTokenDetail                            = errors.New("network token detail is required")
	ErrMissingMerchantPrevioustNetworkTransactionID         = errors.New("merchant previous network transaction id is required")
	ErrMerchantPreviousNetworkTransactionIDNotAllowedForCIT = errors.New("merchant previous network transaction id is not allowed for customer initiated transaction")
	ErrInvalidMITThreeDSMethod                              = errors.New("invalid merchant initiated threeds method")
	ErrInvalidCITThreeDSMethod                              = errors.New("invalid customer initiated threeds method")

	// Others
	ErrFeatureIsNotYetEnable = errors.New("this feature is not yet enabled. please contact our operations team to activate it")
)

// tools error
var (
	ErrDatabaseGetData                         = errors.New("error when get data")
	ErrPartnerInGeneral                        = errors.New("an error occurred while processing your request with the partner system")
	ErrInitDatabaseTransaction                 = errors.New("error initiate database transaction")
	ErrCommitDatabaseTransaction               = errors.New("error commit database transaction")
	ErrRollbackDatabaseTransaction             = errors.New("error rollback database transaction")
	ErrTransactionAlreadyCommittedOrRolledBack = errors.New("transaction already committed or rolled back")
	ErrInvalidTimeZone                         = errors.New("invalid time zone")
	ErrStoreInCache                            = errors.New("error store in cache")
	ErrPaymentPartnerInGeneral                 = errors.New("there was an error with the payment partner. please contact the operations team")
	ErrVANotFoundInProcessor                   = errors.New("virtual account not found in processor")
	ErrVAAlreadyPaidInProcessor                = errors.New("virtual account already paid in processor")
	ErrRecipientIdNotFound                     = errors.New("recipient id not found")
)

const (
	InternalErrorFmt = "there is an internal error with the id %s" // %s = Trace ID

	ErrWhenMarshalCreditcardMetadata   = "error when marshal credit card metadata"
	ErrWhenUnMarshalCreditcardMetadata = "error when unmarshal credit card metadata"
	ErrWhenPublishCreditcardData       = "error when publish credit card data"
	ErrWhenBuildCallbackDataRequest    = "error when build creditcard callback data request"
)

const (
	// V1 Error Code
	ErrCodeInvalidCredential             string = "credentials_invalid"
	ErrCodeFieldFormatInvalid            string = "field_format_invalid"
	ErrCodeFieldValueInvalid             string = "field_value_invalid"
	ErrCodeFieldRequired                 string = "field_required"
	ErrCodeInvalidAmount                 string = "invalid_amount"
	ErrCodeAmountBelowLimit              string = "amount_below_limit"
	ErrCodeAmountAboveLimit              string = "amount_above_limit"
	ErrCodeResourceAlreadyExists         string = "resource_already_exists"
	ErrCodePayoutInProcess               string = "payouts_in_process"
	ErrCodeServiceUnavailable            string = "service_unavailable"
	ErrCodeUnsupportedChannelCode        string = "unsupported_channel_code"
	ErrCodeResourceMissing               string = "resource_missing"
	ErrCodeFormatInvalid                 string = "format_invalid"
	ErrCodeDataNotFound                  string = "data_not_found"
	ErrCodeGeneral                       string = "general_error"
	ErrCodeTimeout                       string = "gateway_timeout"
	ErrCodeBalanceInsufficient           string = "balance_insufficient"
	ErrCodeUnprocessableEntity           string = "unprocessable_entity"
	ErrCodeDailyLimitReached             string = "daily_limit_reached"
	ErrCodeDailyPayoutLimitReached       string = "daily_payout_limit_exceeded"
	ErrCodeForbiddenAccess               string = "forbidden_access"
	ErrCodeInvalidStatusInquiry          string = "invalid_status_inquiry"
	ErrCodeInvalidAccountNumberFormat    string = "invalid_account_number_format"
	ErrCodeInvalidACHCodeFormat          string = "invalid_ach_code_format"
	ErrCodeInvalidSwiftCodeFormat        string = "invalid_swift_code_format"
	ErrCodeInvalidRemarkPattern          string = "invalid_remarks_format"
	ErrCodeInvalidBankNameFormat         string = "invalid_bank_name_format"
	ErrCodeUnallowedPurpose              string = "unallowed_purpose"
	ErrCodeDuplicateError                string = "duplicate_error"
	ErrCodeBadGateway                    string = "bad_gateway"
	ErrCodeFrequencyAboveLimit           string = "frequency_above_limit"
	ErrCodeSwiftCodeNotFound             string = "swift_code_not_found"
	ErrCodeCurrencyNotEnabled            string = "currency_not_enabled"
	ErrCodeInvalidAddressFormat          string = "invalid_address_format"
	ErrCodeFileSizeExceedsLimit          string = "file_size_exceeds_limit"
	ErrCodePayoutAlreadyExpired          string = "payout_expired"
	ErrCodeMerchantNotFound              string = "merchant_not_found"
	ErrCodeInvalidMerchantStatus         string = "invalid_merchant_status"
	ErrCodeResendFailed                  string = "resend_failed"
	ErrCodeAssignUserFailed              string = "assign_user_failed"
	ErrCodeTopupVaFailed                 string = "topup_va_failed"
	ErrCodeInvalidRecipient              string = "invalid_recepient"
	ErrCodeInsufficientFund              string = "insufficient_fund"
	ErrCodePayoutSessionAlreadyConfirmed string = "payout_session_already_confirmed"
	ErrCodeInvalidCardNumber             string = "invalid_card_number"
	ErrCodeCardDecryption                string = "card_decryption_failed"
	ErrCodeInvalidCardInfo               string = "invalid_card_information"
	ErrCodePayoutNotEligible             string = "payout_not_eligible"
)

const (
	// V2 Error Code
	ErrCodeV2InvalidCredential             string = "credentials_invalid"
	ErrCodeV2ResourceNotComplete           string = "resource_not_complete"
	ErrCodeV2APIValidationError            string = "api_validation_error"
	ErrCodeV2RequestForbidden              string = "forbidden_access"
	ErrCodeV2NotFound                      string = "not_found"
	ErrCodeV2ResourceNotFound              string = "resource_missing"
	ErrCodeV2DuplicateError                string = "duplicate_error"
	ErrCodeV2IdempotencyError              string = "idempotency_error"
	ErrCodeV2FrequencyAboveLimit           string = "frequency_above_limit"
	ErrCodeV2DatabaseError                 string = "database_error"
	ErrCodeV2InternalError                 string = "internal_error"
	ErrCodeV2BadGateway                    string = "bad_gateway"
	ErrCodeV2ServiceUnavailable            string = "service_unavailable"
	ErrCodeV2GatewayTimeout                string = "gateway_timeout"
	ErrCodeV2InvalidCardNumber             string = "invalid_card_number"
	ErrCodeV2CardDecryption                string = "card_decryption_failed"
	ErrCodeV2InvalidCardInfo               string = "invalid_card_information"
	ErrCodeV2ForeignCardBillingInfoMissing string = "foreign_card_billing_info_missing"
)

var (
	// V2 Error Messages
	ErrMessageV2InvalidCredential             string = "Access token is invalid, please verify that the authentication is provided and valid"
	ErrMessageV2ResourceNotComplete           string = "Please verify that the setup is complete"
	ErrMessageV2APIValidationError            string = "The request was invalid, or an error occurred in downstream provider"
	ErrMessageV2RequestForbidden              string = "Provided API Key does not have the correct permissions to perform the operation"
	ErrMessageV2NotFound                      string = "The requested URL does not exist"
	ErrMessageV2ResourceNotFound              string = "The $resource with ID $id cannot be found"
	ErrMessageV2DuplicateError                string = "There's already existing record with the provided details"
	ErrMessageV2IdempotencyError              string = "The same Idempotency-key was provided with a different payload"
	ErrMessageV2FrequencyAboveLimit           string = "The frequency limit of resource is reached for operation operation"
	ErrMessageV2DatabaseError                 string = "An internal error was encountered. Please Try again later"
	ErrMessageV2InternalError                 string = "An internal error was encountered. Please Try again later"
	ErrMessageV2BadGateway                    string = "An internal error was encountered. Please Try again later"
	ErrMessageV2ServiceUnavailable            string = "An internal error was encountered. Please Try again later"
	ErrMessageV2GatewayTimeout                string = "An internal error was encountered. Please Try again later"
	ErrMessageV2InvalidCardPaymentNumber      string = "The card number entered is invalid. Please check the number and try again"
	ErrMessageV2InvalidCardPaymentDecryption  string = "Failed to decrypt card information. Please try again later or contact support if the issue persists"
	ErrMessageV2InvalidCardPaymentInformation string = "The entered card information is invalid. Please check the number and try again"
	ErrMessageV2ForeignCardBillingInfoMissing string = "billingInformation is required for foreign card transactions."
)

const (
	ErrMsgAccessTokenInvalid                              = "Access token is invalid"
	ErrMsgFieldFormatInvalid                              = "Format Field is invalid"
	ErrMsgMandatoryField                                  = "Mandatory field is missing"
	ErrMsgAmountBelowMinimum                              = "Amount value below the minimum"
	ErrMsgAmountAboveMaximum                              = "Amount value above the maximum"
	ErrMsgInvalidAmount                                   = "Invalid amount"
	ErrMsgXExternalIdAlreadyExists                        = "X-EXTERNAL-ID is already exists"
	ErrMsgXRequestAlreadyExists                           = "X-Request-Id is already exists"
	ErrMsgXRequestFormatInvalid                           = "X-Request-Id format is invalid"
	ErrMsgXRequestMaximum                                 = "X-Request-Id is too long"
	ErrMsgXRequestMinimum                                 = "X-Request-Id is too short"
	ErrMsgIdAlreadyExists                                 = "ID is already exists"
	ErrMsgPayoutAreBeingInProcess                         = "Payouts are being process"
	ErrMsgAccountNumberAlreadyExists                      = "Account Number is already exists"
	ErrMsgServiceUnavailable                              = "Gateway / Partner service is unavailable"
	ErrMsgChannelCodeNotSupported                         = "Channel Code is currently not supported"
	ErrMsgPayoutIdNotExist                                = "Payout ID is not Exist"
	ErrMsgReferenceIdNotExist                             = "Reference ID is not Exist"
	ErrMsgInquiryIdNotExist                               = "Inquiry ID is not Exist"
	ErrMsgPayoutFormatInvalid                             = "Payouts format is invalid"
	ErrMsgGeneral                                         = "General error"
	ErrMsgTimeout                                         = "Gateway / Partner service is unavailable"
	ErrMsgBalanceInsufficient                             = "Merchant Balance is Insufficient"
	ErrMsgUnauthorized                                    = "Unauthorized"
	ErrMsgInvalidMerchantStatus                           = "Merchant status is invalid, please make sure the merchant is active."
	ErrMsgClientReferenceIdAlreadyExist                   = "Client Reference ID already exists"
	ErrMsgExpiryLessThanCurrentTime                       = "ExpiryAt is not allowed to be less than current time"
	ErrMsgRfiIdNotExist                                   = "Payout status is not RFI"
	ErrMsgInvalidStatusInquiry                            = "Invalid status inquiry"
	ErrMsgUnprocessableEntity                             = "Unprocessable entity"
	ErrMsgPayoutDailyLimitExceeded                        = "Daily Payout limit exceeded"
	ErrMsgPayoutDailyLimitRemainingToday                  = "Remaining amount today: Rp %s. Please try again tomorrow or contact Helpdesk"
	ErrMsgPayoutDeclinedDueToBeneficiaryLimitRestrictions = "Payout request declined due to beneficiary limit restrictions"
	ErrMsgInvalidAccountFormat                            = `Invalid account number format`
	ErrMsgInvalidACHCodeFormat                            = `Invalid ACH code format`
	ErrMsgInvalidSwiftCodeFormat                          = `Invalid SWIFT code format`
	ErrMsgUnallowedPurpose                                = "Unallowed purpose"
	ErrMsgInvalidRemarkPattern                            = "Invalid remark pattern"
	ErrMsgInvalidBankNameFormat                           = "Invalid bank name format"
	ErrMsgSwiftCodeNotFound                               = "Swift code not found"
	ErrMsgCurrencyNotEnabled                              = "Currency not enabled"
	ErrMsgInvalidAddressFormat                            = "Invalid address format"
	ErrMsgFileSizeExceedsLimit                            = "File size exceeds maximum limit"
	ErrMsgPayoutAlreadyExpired                            = "Payout is expired"
	ErrMsgMerchantNotFound                                = "Merchant not found"
	ErrMsgResendInvitationFailed                          = "Resend Invitation Failed: %s"
	ErrMsgAssignAdminUserFailed                           = "Assign Admin User: %s"
	ErrMsgPayoutSessionAlreadyConfirmed                   = "Payout session has been confirmed"
)

// Error Details
var (
	ErrDetailMsgSignature                     = "Make sure the signature value and method is correct"
	ErrDetailMsgAccessToken                   = "Request new access token"
	ErrDetailMsgFormatField                   = "Make sure %s format is correct"
	ErrDetailMsgRequestFormatField            = "Make sure %s request format is correct"
	ErrDetailMsgMandatoryField                = "Make sure %s value is fulfilled"
	ErrDetailMsgAmount                        = "Make sure %s amount is above minimum"
	ErrDetailAmountAboveMaxFmt                = "Make sure %s amount is below maximum" // $1 is field name
	ErrDetailMsgXExternalId                   = "Use unique X-EXTERNAL-ID"
	ErrDetailMsgXRequestId                    = "Use unique X-Request-Id"
	ErrDetailMsgXRequestIdFormat              = "Make sure X-Request-Id format is correct"
	ErrDetailMsgId                            = "Use unique %s"
	ErrDetailMsgPayoutInProcess               = "Wait a moment to retrieve the Payout"
	ErrDetailMsgServiceUnavailable            = "Please hit periodically"
	ErrDetailMsgChannelCode                   = "Make sure channel code following on the API Document"
	ErrDetailMsgPayoutIdNotExist              = "Make sure payout id is correct"
	ErrDetailMsgReferenceIdNotExist           = "Make sure reference id is correct"
	ErrDetailMsgInquiryIdNotExist             = "Make sure Inquiry ID is correct"
	ErrDetailMsgInvalidStatusInquiry          = "Invalid Account inquiry status"
	ErrDetailMsgPayoutFormatInvalid           = "Make sure payout object no more than 1000 items"
	ErrDetailMsgNotExists                     = "%s is not exists"
	ErrDetailMsgGeneralError                  = "Please contact our representative team"
	ErrDetailMsgBalanceInsufficient           = "Re Top-up your Balance"
	ErrDetailMsgRfiIdNotExist                 = "Make sure payout status is RFI"
	ErrDetailMsgInvalidAmount                 = "Invalid amount. Please refer to requirement of each currency."
	ErrDetailMsgInvalidAccountFormat          = `Beneficiary account number should match the pattern "^[A-Z0-9a-z]{1,30}$"`
	ErrDetailMsgInvalidAchCodeFormat          = `ACH code should match the pattern "^[0-9]{9}$"`
	ErrDetailMsgInvalidSwiftCodeFormat        = `SWIFT code should match the pattern "^[A-Za-z0-9]{8}$|^[A-Za-z0-9]{11}$""`
	ErrDetailMsgInvalidSalaryPayment          = "Salary payments should be B2P only"
	ErrDetailMsgInvalidRemarkPattern          = `Remarks should match the pattern "^[a-zA-Z0-9_\s]+$"`
	ErrDetailMsgInvalidBankNameFormat         = `Bank name should match the pattern "^[A-z0-9., -]{1,100}$"`
	ErrDetailMsgSwiftCodeNotFound             = "Please check the SWIFT code"
	ErrDetailMsgCurrencyNotEnabled            = "Currency not enabled, please contact the representative team"
	ErrDetailMsgInvalidAddressFormat          = "Invalid address format"
	ErrDetailMsgFileSizeExceedsLimit          = "File size exceeds the allowed limit"
	ErrDetailMsgPayoutAlreadyExpired          = "Payout is expired, please try to create new payout"
	ErrDetailMsgMerchantNotFound              = "Invalid Merchant request"
	ErrDetailMsgFieldFormatInvalid            = "Please check the format of the field"
	ErrDetailMsgResendInvitationFailed        = "Cannot resend invitation: %s"
	ErrDetailMsgAssignAdminUserFailed         = "Cannot assign admin user: %s"
	ErrDetailMsgUnallowedPurposeCode          = "Unallowed purpose code"
	ErrDetailMsgPayoutSessionAlreadyConfirmed = "Payout session has been confirmed before. Please check the payout status."
	ErrDetailMsgInvalidPayoutAmount           = "Payout amount should be integer"

	ErrDetailMsgCharExceedTheMax        = "Make sure the number of characters in the %s column is not greater than %s"
	ErrDetailMsgMakeSureValueIsCorrect  = "Make sure the %s is correct"
	ErrDetailMsgMakeSureValueIsAboveMin = "Make sure %s is above minimum"
	ErrDetailMsgMakeSureValueIsBelowMax = "Make sure %s is below maximum"

	ErrDetailMsgVaNumberIsOutsideValidRangeFmt = "The requested Virtual Account number is not within your assigned %s VA range. Please check your configuration." // {static|dynamic}
	ErrDetailMsgVaNumberStillInUse             = "The Virtual Account number you requested has already been assigned. Please choose a different number."
	ErrDetailMsgNoAvailableVaNumberToAssignFmt = "No available VA number to assign. All %s VA numbers in the configured range are in use. Please try again later or adjust your range." // {static|dynamic}
	ErrDetailMsgPayoutDstNotEligible           = "Destination account is not eligible for payout"
	ErrDetailMsgPayoutDstNotEligibleFmt        = ErrDetailMsgPayoutDstNotEligible + ". bankCode=%s accountNumber=%s" // $1 bank code, $2 account number
)

// XB
var (
	ErrWhenCreateSenderData      = errors.New("error when create sender data")
	ErrWhenCreateBeneficiaryData = errors.New("error when create beneficiary data")
)

// Product
var (
	ErrGetExistingMerchantSelectedProduct   = errors.New("error when get existing merchant selected product")
	ErrMerchantSelectedProductAlreadyExists = errors.New("merchant selected product already exists")
	ErrMerchantNotSelectedProduct           = errors.New("merchant is not selected product")
	ErrAddMerchantSelectedProduct           = errors.New("error when register merchant selected product")
	ErrGetMerchantSelectedProducts          = errors.New("error when get merchant selected products")
	ErrProductNotFound                      = errors.New("product not found")
	ErrGetProduct                           = errors.New("error when get product")
	ErrGetProductList                       = errors.New("error when get product list")
	ErrUpdateMerchantProductAvailability    = errors.New("error when update merchant product availability")
	ErrUpdateProductAvailability            = errors.New("error when update product availability")
	ErrValidateMerchantProductAvailability  = errors.New("error when validate merchant product availability")

	MerchantIsNotAllowedToUseProductMsgFormat = fmt.Sprintf("merchant is not allowed to use %s product", "%v")
)

// Creditcard
const (
	ErrWhenGetCreditcardMetaData      = "error when get creditcard metadata: %w"
	ErrWhenCreditcardGetPaymentStatus = "error when get credit card payment status: %w"
)

// Creditcard Validation Error Message
const (
	ErrMsgForeignCardBillingInformationValidation = "foreign cards require billing information:"
)

// Orchestrator
var (
	ErrWhenFindAccountTransaction   = errors.New("error when find account transaction")
	ErrWhenUpdateAccountTransaction = errors.New("error when update account transaction")
)

// CRM
var (
	ErrMissingXIdentifier = errors.New("x-identifier header is required")
)

// IP Whitelist
var (
	ErrInvalidIPAddress                  = errors.New("invalid ip address")
	ErrGetIPWhitelistConfigurationList   = errors.New("error when get ip whitelist configuration list")
	ErrGetIPWhitelistConfigurationDetail = errors.New("error when get ip whitelist configuration detail")
	ErrDeleteIPWhitelistConfiguration    = errors.New("error when delete ip whitelist configuration")
	ErrIPWhitelistConfigurationNotFound  = errors.New("ip whitelist configuration detail not found")
	ErrCreateIPWhitelistConfiguration    = errors.New("error create ip whitelist configuration")
	ErrUpdateIPWhitelistConfiguration    = errors.New("error update ip whitelist configuration")
	ErrIPConfigurationAlreadyExists      = errors.New("ip whitelist configuration already exists")
	ErrInvalidIPConfigurationID          = errors.New("invalid ip whitelist configuration id")
	ErrCachedIPWhitelistConfiguration    = errors.New("error cached ip whitelist configuration")
	ErrForbiddenIPAddress                = errors.New("IP Address not allowed to access resources")
)

// rate limiter
var (
	ErrGetRateLimiterConfigurationList   = errors.New("error when get rate limiter configuration list")
	ErrGetRateLimiterConfigurationDetail = errors.New("error when get rate limiter configuration detail")
	ErrRateLimiterConfigurationNotFound  = errors.New("rate limiter configuration detail not found")
	ErrInvalidRateLimiterConfigurationID = errors.New("invalid rate limiter configuration id")
	ErrCreateRateLimiterConfiguration    = errors.New("error create rate limiter configuration")
	ErrUpdateRateLimiterConfiguration    = errors.New("error update rate limiter configuration")
)

// Error get mid detail
var (
	ErrGetMIDDetail             = errors.New("error get mid detail")
	ErrInvalidMIDSettlementType = errors.New("invalid mid settlement type")
)

// Installment Plan
var (
	ErrCreateInstallmentPlan                = errors.New("error create installment plan")
	ErrUpdateInstallmentPlan                = errors.New("error update installment plan")
	ErrInvalidMIDTransactionType            = errors.New("invalid mid transaction type")
	ErrMismatchInstallmentTenor             = errors.New("installment tenor not matched with mid tenor")
	ErrGetInstallmentPlan                   = errors.New("error get installment plan")
	ErrInstallmentPlanNotFound              = errors.New("installment plan not found")
	ErrActiveInstallmentEmptyBins           = errors.New("active installment plan must have allowed bins")
	ErrDependentCardPaymentMethodNotActive  = errors.New("card payment method is not active")
	ErrCardInstallmentNotConfigured         = errors.New("card installment not configured yet")
	ErrInstallmentGetDependentPaymentMethod = errors.New("error get installment dependent payment method")
)

// Merchant RCNs
var (
	ErrMerchantRcnNotFound = errors.New("merchant rcn not found")
	ErrDecodeMerchantRcn   = errors.New("error decode merchant rcn")
	ErrDecryptMerchantRcn  = errors.New("error decrypt merchant rcn")
)

// VCC Settlement
var (
	ErrRequestCimbTransactionInquiry                = errors.New("error when request transaction inquiry cimb")
	ErrFailCimbTransactionInquiry                   = errors.New("fail request transaction inquiry to cimb")
	ErrGeneralServiceProviderCimbTransactionInquiry = errors.New("error transaction inquiry to cimb")
	ErrAcquireTransactionInquiryLock                = errors.New("error acquire transaction inquiry lock")
	ErrProcessSettlementTransactionInquiry          = errors.New("error process settlement transaction inquiry")
	ErrPublishSettlementTransactionProcess          = errors.New("error when publish transaction inquiry process")
)

// SettlementHold
var (
	ErrValidateSettlementHold       = errors.New("error validate settlement hold")
	ErrStoredSettlementHold         = errors.New("error stored settlement hold")
	ErrProcessSettlementHoldRelease = errors.New("error process settlement hold/release")
	ErrGetSettlementHold            = errors.New("error get settlement hold config")
	ErrProcessManualSettlement      = errors.New("error process manual settlement")
)

// Card funded payout
var (
	ErrIncorrectCardFundedPayoutPaymentSessionDetailType = errors.New("payment session detail type is not card funded payout")
)

// MapErrorCodeToV2 maps V1 error codes to V2 error codes.
func MapErrorCodeToV2(v1Code string) string {
	switch v1Code {
	case ErrCodeInvalidCredential:
		return ErrCodeV2InvalidCredential
	case ErrCodeFieldFormatInvalid, ErrCodeFormatInvalid, ErrCodeFieldRequired, ErrCodeFieldValueInvalid, ErrCodeUnsupportedChannelCode, ErrCodeUnprocessableEntity, ErrCodeInvalidStatusInquiry, ErrCodeV2ForeignCardBillingInfoMissing:
		return ErrCodeV2APIValidationError
	case ErrCodeAmountBelowLimit, ErrCodeAmountAboveLimit, ErrCodeDailyLimitReached, ErrCodeDailyPayoutLimitReached:
		return ErrCodeV2FrequencyAboveLimit
	case ErrCodeResourceAlreadyExists, ErrCodePayoutInProcess, ErrCodeDuplicateError:
		return ErrCodeV2DuplicateError
	case ErrCodeServiceUnavailable:
		return ErrCodeV2ServiceUnavailable
	case ErrCodeTimeout:
		return ErrCodeV2GatewayTimeout
	case ErrCodeResourceMissing:
		return ErrCodeV2ResourceNotFound
	case ErrCodeDataNotFound:
		return ErrCodeV2NotFound
	case ErrCodeGeneral:
		return ErrCodeV2InternalError
	case ErrCodeBalanceInsufficient, ErrCodeForbiddenAccess:
		return ErrCodeV2RequestForbidden
	case ErrCodeBadGateway:
		return ErrCodeV2BadGateway
	case ErrCodeFrequencyAboveLimit:
		return ErrCodeV2FrequencyAboveLimit
	case ErrCodeInvalidCardNumber:
		return ErrCodeV2InvalidCardNumber
	case ErrCodeInvalidCardInfo:
		return ErrCodeV2InvalidCardInfo
	case ErrCodeCardDecryption:
		return ErrCodeV2CardDecryption

	default:
		return ErrCodeV2InternalError // Default mapping for unknown errors
	}
}

// MapV2ErrorCodeToMessage maps V2 error codes to error messages.
func MapV2ErrorCodeToMessage(v2Code string) string {
	switch v2Code {
	case ErrCodeV2InvalidCredential:
		return ErrMessageV2InvalidCredential
	case ErrCodeV2ResourceNotComplete:
		return ErrMessageV2ResourceNotComplete
	case ErrCodeV2APIValidationError, ErrCodeV2ForeignCardBillingInfoMissing:
		return ErrMessageV2APIValidationError
	case ErrCodeV2RequestForbidden:
		return ErrMessageV2RequestForbidden
	case ErrCodeV2NotFound:
		return ErrMessageV2NotFound
	case ErrCodeV2ResourceNotFound:
		return ErrMessageV2ResourceNotFound
	case ErrCodeV2DuplicateError:
		return ErrMessageV2DuplicateError
	case ErrCodeV2IdempotencyError:
		return ErrMessageV2IdempotencyError
	case ErrCodeV2FrequencyAboveLimit:
		return ErrMessageV2FrequencyAboveLimit
	case ErrCodeV2DatabaseError:
		return ErrMessageV2DatabaseError
	case ErrCodeV2InternalError:
		return ErrMessageV2InternalError
	case ErrCodeV2BadGateway:
		return ErrMessageV2BadGateway
	case ErrCodeV2ServiceUnavailable:
		return ErrMessageV2ServiceUnavailable
	case ErrCodeV2GatewayTimeout:
		return ErrMessageV2GatewayTimeout
	case ErrCodeV2InvalidCardNumber:
		return ErrMessageV2InvalidCardPaymentNumber
	case ErrCodeV2CardDecryption:
		return ErrMessageV2InvalidCardPaymentDecryption
	case ErrCodeV2InvalidCardInfo:
		return ErrMessageV2InvalidCardPaymentInformation
	default:
		return "Unknown error"
	}
}

type GeneralError interface {
	error
	OriginalError() error
	GetResponseCode() string
	GetResponseMessage() (message, detail string)
}

func NewErrInternalPartner(processName string, err error) error {
	return &ErrInternalPartner{processName, err}
}

func NewErrRequiredField(fieldName string) error               { return &ErrRequiredField{fieldName} }
func NewErrInvalidFieldFmt(fieldName string) error             { return &ErrInvalidFieldFmt{fieldName} }
func NewErrResourceNotFound(resource, id string) error         { return &ErrResourceNotFound{resource, id} }
func NewErrInvalidPayload(err error) error                     { return &ErrInvalidPayload{err} }
func NewErrFieldValidation(err error) error                    { return &ErrFieldValidation{err} }
func NewErrInvalidData(process string, err error) GeneralError { return &ErrInvalidData{process, err} }
func NewErrInvalidRules(process string, err error) GeneralError {
	return &ErrInvalidRules{process, err}
}
func NewErrStringRequest(errType, code, message string) error {
	return &ErrStringRequest{errType, code, message}
}

type ErrInternalPartner struct {
	processName string
	err         error
}

func (e ErrInternalPartner) OriginalError() error { return e.err }
func (e ErrInternalPartner) Error() string {
	switch e.processName {
	default:
		return e.err.Error()

	case ProcessNameGenerateVaTopup:
		return "ERROR_REQUEST | Internal partner"
	}
}

func (e ErrInternalPartner) GetResponseCode() string {
	switch e.processName {
	default:
		return ErrCodeGeneral

	case ProcessNameGenerateVaTopup:
		return ErrCodeTopupVaFailed
	}
}

func (e ErrInternalPartner) GetResponseMessage() (message, detail string) {
	switch e.processName {
	default:
		return ErrMessageV2InternalError, ErrDetailMsgGeneralError

	case ProcessNameGenerateVaTopup:
		return "Failed to Generate Virtual Account for Topup", "Failed to generate VA for topup"
	}
}

type ErrInvalidData struct {
	processName string
	err         error
}

func (e ErrInvalidData) Error() string        { return fmt.Sprintf("ERROR_REQUEST | %v", e.err) }
func (e ErrInvalidData) OriginalError() error { return e.err }
func (e ErrInvalidData) GetResponseCode() string {
	switch e.processName {
	default:
		return ""

	case ProcessNamePlatformTransfer:
		if errors.Is(e.err, ErrRecipientIdNotFound) || errors.Is(e.err, ErrSameMerchant) {
			return ErrCodeInvalidRecipient
		} else if errors.Is(e.err, ErrInsufficientBalance) {
			return ErrCodeInsufficientFund
		}
	}
	return ""
}

func (e ErrInvalidData) GetResponseMessage() (message, detail string) {
	switch e.processName {
	default:
		return "", ""

	case ProcessNamePlatformTransfer:
		if errors.Is(e.err, ErrRecipientIdNotFound) {
			return "Recipient ID is invalid", fmt.Sprintf(ErrDetailMsgMakeSureValueIsCorrect, "Recipient ID")
		} else if errors.Is(e.err, ErrSameMerchant) {
			return "Recipient ID is invalid", "Cannot transfer to the same merchant"
		} else if errors.Is(e.err, ErrInsufficientBalance) {
			return "Insufficient balance for Transfer", "Make sure balance is sufficient for Transfer"
		}
	}
	return "", ""
}

type ErrInvalidRules struct {
	processName string
	err         error
}

func (e ErrInvalidRules) Error() string        { return "ERROR_REQUEST | Invalid rules" }
func (e ErrInvalidRules) OriginalError() error { return e.err }
func (e ErrInvalidRules) ProcessName() string  { return e.processName }
func (e ErrInvalidRules) GetResponseCode() string {
	switch e.processName {
	default:
		return ""

	case ProcessNameResendInvitation:
		return ErrCodeResendFailed

	case ProcessNameAssignAdminUser:
		return ErrCodeAssignUserFailed
	}
}

func (e ErrInvalidRules) GetResponseMessage() (message, detail string) {
	switch e.processName {
	case ProcessNameResendInvitation:
		if errors.Is(e.err, ErrUserNotFound) {
			return fmt.Sprintf(ErrMsgResendInvitationFailed, "User Is Invalid"), fmt.Sprintf(ErrDetailMsgResendInvitationFailed, "the user is invalid or no longer eligible for an invitation")
		} else if errors.Is(e.err, ErrUserAlreadyActivated) {
			return fmt.Sprintf(ErrMsgResendInvitationFailed, "User Already Active"), fmt.Sprintf(ErrDetailMsgResendInvitationFailed, "the user is already active")
		}
	case ProcessNameAssignAdminUser:
		if errors.Is(e.err, ErrUserAlreadyExists) {
			return fmt.Sprintf(ErrMsgAssignAdminUserFailed, "Email Already Exists"), fmt.Sprintf(ErrDetailMsgAssignAdminUserFailed, "the email is already exists")
		}
	}
	return "", ""
}

type ErrStringRequest struct {
	errType      string
	responseCode string
	message      string
}

func (e ErrStringRequest) Error() string           { return e.errType + " | " + e.message }
func (e ErrStringRequest) Message() string         { return e.message }
func (e ErrStringRequest) GetResponseCode() string { return e.responseCode }

type ErrRequiredField struct{ field string }

func (e ErrRequiredField) Error() string        { return "ERROR_REQUEST | Field is required" }
func (e ErrRequiredField) Message() string      { return fmt.Sprintf(ErrDetailMsgMandatoryField, e.field) }
func (e ErrRequiredField) GetFieldName() string { return e.field }

type ErrInvalidFieldFmt struct{ field string }

func (e ErrInvalidFieldFmt) Error() string { return "ERROR_REQUEST | Invalid field format" }
func (e ErrInvalidFieldFmt) Message() string {
	return fmt.Sprintf("Make sure %s format is correct", e.field)
}
func (e ErrInvalidFieldFmt) GetFieldName() string { return e.field }

type ErrResourceNotFound struct {
	resource string
	id       string
}

func (e ErrResourceNotFound) Error() string { return "ERROR_NOT_FOUND | Resource not found" }
func (e ErrResourceNotFound) Message() string {
	return fmt.Sprintf("The %s with ID %s cannot be found", e.resource, e.id)
}

type ErrInvalidPayload struct{ err error }

func (e ErrInvalidPayload) Error() string   { return "ERROR_REQUEST | Invalid payload" }
func (e ErrInvalidPayload) Message() string { return ErrMessageV2APIValidationError }

type ErrFieldValidation struct{ err error }

func (e ErrFieldValidation) Error() string        { return "ERROR_REQUEST | Invalid field validation" }
func (e ErrFieldValidation) OriginalError() error { return e.err }
