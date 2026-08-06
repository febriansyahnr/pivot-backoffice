package constant

import "time"

const (
	DisbursementTypeLocalPayout         = "LOCAL_PAYOUT"
	DisbursementTypeInternationalPayout = "INTERNATIONAL_PAYOUT"
	DisbursementTypeCardFundedPayout    = "CARD_FUNDED_PAYOUT"

	DisbursementStatusWaiting  = "WAITING"
	DisbursementStatusApproved = "APPROVED"
	DisbursementStatusRejected = "REJECTED"
	DisbursementStatusPending  = "PENDING"

	BulkDisbursementStatusUploading  = "UPLOADING"
	BulkDisbursementStatusWaiting    = "WAITING"
	BulkDisbursementStatusInProgress = "IN_PROGRESS"
	BulkDisbursementStatusDone       = "DONE"
	BulkDisbursementStatusPending    = "PENDING"

	DisbursementCreatedFromMerchantPortal = "MERCHANT_PORTAL"
	DisbursementCreatedFromOpenApi        = "OPEN_API"

	BulkPreviewResultValid    = "VALID"
	BulkPreviewResultInvalid  = "INVALID"
	BulkPreviewResultWarning  = "WARNING"
	BulkPreviewResultRejected = "REJECTED"

	DisbursementTypeSingle = "SINGLE"
	DisbursementTypeBulk   = "BULK"

	DisbursementCutOffTimeStatusOffSchedule = "OFF_SCHEDULE"
	DisbursementCutOffTimeStatusOngoing     = "ONGOING"
	DisbursementCutOffTimeStatusWhitelisted = "WHITELISTED"
	DisbursementCutOffTimeStatusDeactive    = "DEACTIVE"

	DisbursementTypeSingleTitle = "Single Transaction"
	DisbursementTypeBulkTitle   = "Bulk Transaction"

	DisbursementDailyLimitMerchant         = "merchant"
	DisbursementDailyLimitMerchantPlatform = "merchant-platform"

	BulkDisbursementQueueLockFmt                = "backend-portal:locks:disbursements:bulk-create:queueus:%s:%s"          // $1 is merchant_id and $2 is reference id
	BulkDisbursementInProgressQueueLockFmt      = "backend-portal:locks:disbursements:bulk-create:in-progress:%s:%s"      // $1 is merchant_id and $2 is bulk disbursement id
	DisbursementProcessQueueLockFmt             = "backend-portal:locks:disbursements:process:queues:%s"                  // $1 is disbursement ID
	DisbursementTransactionConfigFmt            = "backend-portal:transaction-configs:merchants:%s:disbursement"          // $1 is Merchant Id or Sub-Merchant Id
	DailyDisbursementTransactionConfigFmt       = "backend-portal:transaction-configs:merchants:%s:daily-disbursement:%s" // $1 is Merchant Id and $2 is Merchant Type
	DelayTransferProcessRedisKey                = "backend-portal:delay-transfers"
	CutOffReportMemberDedupKeyFmt               = "backend-portal:cutoff-report:member:%s:%s:%s"
	CutOffReportMemberIndexKeyFmt               = "backend-portal:cutoff-report:members:%s:%s"
	CutOffReportMemberLastSeenKeyFmt            = "backend-portal:cutoff-report:last-seen:%s:%s"
	CutOffReportExecutionDedupKeyFmt            = "backend-portal:cutoff-report:dedup:%s:%s"
	BeneficiaryPayoutDefaultRuleLimitFmt        = "backend-portal:beneficiary-payout-limit:bank-code:%s:account-no:%s"                     // $1 is Bank Code $2 is Account No
	BeneficiaryPayoutCustomRuleLimitFmt         = "backend-portal:beneficiary-payout-limit:merchants:%s:bank-code:%s:account-no:%s"        // $1 is Merchant Id and $2 is Bank Code $3 is Account No
	BeneficiaryPayoutMerchantPolicyRuleLimitFmt = "backend-portal:beneficiary-payout-limit:merchants:%s:policy:bank-code:%s:account-no:%s" // $1 is Merchant Id and $2 is Bank Code $3 is Account No
	ListOverbookingBankCacheKey                 = "backend-portal:payout:list-overbooking-bank"
	DisbursementApprovalBeneficiaryLockFmt      = "backend-portal:locks:disbursements:approval:beneficiary:%s:%s" // $1 is bank code and $2 is account no
	DisbursementCallbackEventLockFmt            = "backend-portal:locks:disbursements:callback:event:%s:%s:%s"    // $1 merchant id, $2 bulk id, $3 event
)

const (
	CutOffReportRedisTTL                 = 72 * 60 * 60 // seconds
	CutOffReportSettleQuietWindowSeconds = 5
	CutOffReportSettleMaxWaitSeconds     = 30
	CallbackEventLockTTL                 = time.Minute * 15
)

// Reason Type
const (
	// General
	DisbursementReasonTypeInsufficientBalance = "INSUFFICIENT_BALANCE"
	DisbursementReasonTypeCancelled           = "CANCELLED"
)

const (
	BulkDisbursementMaxDataRequest         = 1000
	BulkDisbursementMaxDataRequestPerBatch = 100
)

const (
	DisbursementMinAmount                = 10_000      // Merchant Config: merchant.DisbursementConfig.MinAmount
	DisbursementMaxAmount                = 250_000_000 // Merchant Config: merchant.DisbursementConfig.MaxAmount
	DisbursementMaxLengthRemark          = 40
	DisbursementMaxLengthBeneficiaryName = 100
)

const (
	DisbursementTypeTransfer = "TRANSFER"
	DisbursementTypeWallet   = "WALLET"
)

const (
	DisbursementBankTransferTypeOverbooking = "INTRABANK"
)

const (
	DisbursementBeneficiaryLimitMerchantPolicy = "merchant-policy"
	DisbursementBeneficiaryLimitCustom         = "custom"
)
