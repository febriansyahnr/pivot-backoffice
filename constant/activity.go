package constant

const (
	TagAccount                  = "account"
	TagCallback                 = "callback"
	TagMerchant                 = "merchant"
	TagPayment                  = "payment"
	TagDisbursement             = "disbursement"
	TagMerchantForbiddenUseCase = "merchant-forbidden-disbursement"
	TagSetAutoWithdrawalStatus  = "set-auto-withdrawal-status"
	TagTNC                      = "agreement"

	TagCredentialSettings = "credential-settings"
	TagCallbackSetting    = "callbacks-settings"
	TagManualWithdrawal   = "manual-withdrawal"
)

const (
	// User Activity
	ActivityUserLogin                   = "User login"
	ActivityUserLogout                  = "User logout"
	ActivityUserChangePassword          = "User change password"
	ActivityUserGenerateRandomPassword  = "User generate random password"
	ActivityUserRegisterCallback        = "User register merchant callback"
	ActivityUserCreateMerchant          = "User create merchant"
	ActivityUserUpdateMerchant          = "User update merchant"
	ActivityUserAssignMerchant          = "User assign merchant"
	ActivityUserCreateMerchantFee       = "User create merchant fee"
	ActivityUserUpdateMerchantFee       = "User update merchant fee"
	ActivityUserCheckBeneficiaryAccount = "User check beneficiary account"
	ActivityUserUpdatePIN               = "User update PIN"

	ActivityUserAccessCredDashboard           = "User access credential dashboard"
	ActivityUserViewClientSecretSuccess       = "User view client secret"
	ActivityUserViewClientSecretFailed        = "User failed to verify PIN to view client secret"
	ActivityUserRegenerateClientSecretSuccess = "User regenerate client secret"
	ActivityUserRegenerateClientSecretFailed  = "User failed to verify PIN to regenerate client secret"
	ActivityUserAccessCallbackDashboard       = "User accsess callbacks dashboard"
	ActivityUserViewCallbackAPIKeySuccess     = "User view callback api key"
	ActivityUserViewCallbackAPIKeyFailed      = "User failed to verify PIN to view callback api key"
	ActivityUserTestCallbackURLFailed         = "Failed to perform callback test"
	ActivityUserSaveCallbackURLFailed         = "Failed to save callback URL"
	ActivityUserTestAndSaveCallbackURLSuccess = "Successfully tested and saved the URL callback"
	ActivityUserResendInvitation              = "User resend invitation"

	ActivityUserAccessCallbackHistory     = "User access and filter callback history"
	ActivityUserViewCallbackHistoryDetail = "User view callback history detail"
	ActivityUserResendCallbackHistory     = "User resend callback"

	ActivityUserFailedCheckPin = "User failed to verify PIN"
	ActivityUserFailedLogin    = "User failed to login"

	// Merchant Activity
	ActivityMerchantCreatePayment      = "Merchant create payment"
	ActivityMerchantUpdatePayment      = "Merchant update payment"
	ActivityMerchantCreateDisbursement = "Merchant create disbursement"
	ActivityMerchantTNCSign            = "TNC Signed"

	ActivityMerchantRetryDisbursement = "Merchant retry disbursement"

	// User Block Merchant Disbursement
	ActivityBlockMerchantDisbursement   = "Block Merchant Disbursement"
	ActivityUnblockMerchantDisbursement = "Unblock Merchant Disbursement"

	// Auto Withdrawal Status
	ActivityAutoWithdrawalSetON  = "Enable auto withdrawal"
	ActivityAutoWithdrawalSetOFF = "Disable auto withdrawal"

	// CRM
	ActivityUserBlockedByOps  = "User blocked by ops"
	ActivityUserChangeKYCInfo = "User change KYC information"
)
