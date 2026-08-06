package constant

const (
	PermissionByRoleKeyPattern = "backend-portal:permissions:%s:cache"
)

const (
	// Navbar region
	// Team Member
	PermissionSlugTeamMemberView   = "team-members.view"
	PermissionSlugTeamMemberCreate = "team-members.create"
	PermissionSlugTeamMemberEdit   = "team-members.edit"

	//Roles and Permissions
	PermissionSlugRolesAndPermissionsView   = "roles-and-permissions.view"
	PermissionSlugRolesAndPermissionsCreate = "roles-and-permissions.create"
	PermissionSlugRolesAndPermissionsEdit   = "roles-and-permissions.edit"
	PermissionSlugRolesAndPermissionsDelete = "roles-and-permissions.delete"

	// Activity Log
	PermissionSlugActivityLogView = "activity-logs.view"

	// Developer Setting
	PermissionSlugDeveloperSettingView   = "developer-setting.view"
	PermissionSlugDeveloperSettingCreate = "developer-setting.create"
	PermissionSlugDeveloperSettingEdit   = "developer-setting.edit"
	PermissionSlugDeveloperSettingDelete = "developer-setting.delete"

	// Deposit Setting
	PermissionSlugDepositSettingView = "deposit-setting.view"
	PermissionSlugDepositSettingEdit = "deposit-setting.edit"

	//end region

	// Sidebar region
	// Transaction History
	PermissionSlugTransactionHistoryView = "transaction-history.view"

	// Disbursement
	PermissionSlugDisbursementView = "disbursement.view"

	// Disbursement History
	PermissionSlugDisbursementHistoryView = "disbursement.disbursement-history.view"

	// Disbursement Approval
	PermissionSlugDisbursementApprovalView   = "disbursement.disbursement-approval.view"
	PermissionSlugDisbursementApprovalCreate = "disbursement.disbursement-approval.create"

	// Merchant Topup
	PermissionSlugDisbursementTopUpView   = "disbursement.top-up.view"
	PermissionSlugDisbursementTopUpCreate = "disbursement.top-up.create"

	// Wallet
	PermissionSlugWalletMerchantBalanceView              = "wallet.merchant.balance.view"
	PermissionSlugWalletMerchantWithdrawCreate           = "wallet.merchant.balance.create"
	PermissionSlugWalletMerchantTransactionHistoriesView = "wallet.merchant.transaction-histories.view"

	PermissionSlugWalletCustomersInsightsView     = "wallet.customers.insights.view"
	PermissionSlugWalletCustomersView             = "wallet.customers.view"
	PermissionSlugWalletCustomersVerificationView = "wallet.customers.verification.view"
	PermissionSlugWalletCustomersTransactionsView = "wallet.customers.transactions.view"

	// Create Transaction
	PermissionSlugCreateTransactionView   = "disbursement.disbursement-create.view"
	PermissionSlugCreateTransactionCreate = "disbursement.disbursement-create.create"

	// Beneficiary List
	PermissionSlugDisbursementBeneficiaryListView = "disbursement.disbursement-beneficiary-list.view"

	// Payment Dashboard
	PermissionSlugPaymentInsightView   = "payment.insight.view"
	PermissionSlugPaymentHistoriesView = "payment.histories.view"
	// Withdrawal from payment balance
	PermissionSlugPaymentWithdrawalView   = "payment.withdrawal.view"
	PermissionSlugPaymentWithdrawalCreate = "payment.withdrawal.create"
	// Payment Transfer
	PermissionSlugPaymentTransferView = "payment.transfer.view"
	PermissionSlugPaymentLinkCreate   = "payment.link.create"
	// Cases Management (Payment Investigation)
	PermissionSlugCasesManagementView = "payment.cases.view"
	// Payment Refund
	PermissionSlugPaymentRefundCreate = "payment.refund.create"
	//end region

	// Platform
	PermissionSlugPlatformView                 = "platform.view"
	PermissionSlugPlatformEdit                 = "platform.edit"
	PermissionSlugPlatformMerchantCreate       = "platform.merchant.create"
	PermissionSlugPlatformMerchantEdit         = "platform.merchant.edit"
	PermissionSlugPlatformMerchantDeactivate   = "platform.merchant.deactivate"
	PermissionSlugPlatformPayoutDailyLimitView = "platform.payout.dailylimit.view"
	// End Platform

	// XB
	PermissionSlugInternationalPayoutView   = "international-payout.view"
	PermissionSlugInternationalPayoutCreate = "international-payout.create"

	// Home
	PermissionSlugHomeView = "home.view"

	// VCC Terminal
	PermissionSlugVccTerminalBalanceView           = "vcc-terminal.balance.view"
	PermissionSlugVccTerminalBalanceWithdraw       = "vcc-terminal.balance.withdraw"
	PermissionSlugVccTerminalChargeHistoryView     = "vcc-terminal.charge-history.view"
	PermissionSlugVccTerminalChargeHistoryDownload = "vcc-terminal.charge-history.download"
	PermissionSlugVccTerminalCreate                = "vcc-terminal.create"
	PermissionSlugVccTerminalChargeDetailView      = "vcc-terminal.charge-detail.view"
	PermissionSlugVccTerminalChargeDetailDownload  = "vcc-terminal.charge-detail.download"

	// Notification Settings
	PermissionSlugNotificationSettingsView = "notification-settings.view"
	PermissionSlugNotificationSettingsEdit = "notification-settings.edit"

	// Card Funded Payouts
	PermissionSlugCardFundedPayoutView             = "card-funded-payout.view"
	PermissionSlugCardFundedPayoutCreate           = "card-funded-payout.create"
	PermissionSlugCardFundedPayoutSavedCardsView   = "card-funded-payout.saved-cards.view"
	PermissionSlugCardFundedPayoutSavedCardsCreate = "card-funded-payout.saved-cards.create"
	PermissionSlugCardFundedPayoutVendorView       = "card-funded-payout.vendor.view"
	PermissionSlugCardFundedPayoutNeedActionView   = "card-funded-payout.need-action.view"
	PermissionSlugCardFundedPayoutNeedActionCreate = "card-funded-payout.need-action.create"
)

// HomeMenuRequiredPermissionPrefixes defines permission prefixes that grant access to Home menu
// Users with any permission starting with these prefixes will automatically see the Home menu
var HomeMenuRequiredPermissionPrefixes = []string{
	"disbursement.",
	"payment.",
	"international-payout.",
}
