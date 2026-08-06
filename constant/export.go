package constant

const (
	ExportTempDir    = "tmp"
	DefaultExtXlsx   = ".xlsx"
	DefaultSheetName = "Sheet1"
)

const (
	DefaultFilenameTransactionHistory       = "transaction_history_%s"
	DefaultFilenameUploadBulkDisbursement   = "uploaded_bulk_disbursement_%s"
	DefaultFilenameInvalidBulkDisbursement  = "invalid_bulk_disbursement_%s"
	DefaultFilenameRejectedBulkDisbursement = "rejected_bulk_disbursement_%s"
	DefaultFilenameDisbursementHistory      = "disbursement_history_%s"
	DefaultFilenameUploadReconciliation     = "uploaded_reconciliation_%s"
	DefaultFilenameResultReconciliation     = "result_reconciliation_%s"
	DefaultFilenameCardFundedPayoutHistory  = "card_funded_payout_history_%s"
)

const (
	ExportBulkDisbursementSuccessDir  = "success"
	ExportBulkDisbursementFailedDir   = "failed"
	ExportBulkDisbursementTemplateDir = "template"
)

const (
	DisbursementHistoryBucketDir       = "disbursements/history"
	CardFundedPayoutHistoryBucketDir   = "card-funded-payouts/history"
)

const (
	ReconciliationUploadDir = "reconciliations/uploads"
	ReconciliationResultDir = "reconciliations/results"
)

const (
	RedisKeyDownloadWithdrawalHistoryFmt                = "backend-portal:downloads:withdrawal-histories:%s"
	RedisKeyDownloadPaymentHistoryFmt                   = "backend-portal:downloads:payment-histories:%s"
	RedisKeyDownloadWalletMerchantTransactionHistoryFmt = "backend-portal:downloads:wallet-merchant-transaction-histories:%s"
	RedisKeyDownloadChargeHistoryFmt                    = "backend-portal:downloads:charge-histories:%s"
	RedisKeyInquiryCooldownFmt                          = "backend-portal:inquiry:cooldown:%s"
)
