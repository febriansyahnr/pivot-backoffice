package rabbitMqExt

import "time"

const (
	ActivityDirectExchange          = "backend-portal.activity.direct"
	CallbackExchange                = "backend-portal.callback"
	BulkDisbursementExchange        = "backend-portal.bulk-disbursement"
	MerchantExchange                = "backend-portal.merchant"
	OrchestratorTransactionExchange = "orchestrator"
	SlackExchange                   = "backend-portal.slack"
	NotificationExchange            = "backend-portal.notifications"
	UnroutedNotificationExchange    = "unrouted." + NotificationExchange
	AccountExchange                 = "backend-portal.account"
	SchedulingSettlementExchange    = "backend-portal.scheduling.settlement"
	CommServiceExchange             = "backend-portal.comm-service"
	WithdrawalExchange              = "backend-portal.withdrawals"
	RefundExchange                  = "backend-portal.refund"
	PaymentCaptureExchange          = "backend-portal.payment-capture"
	SchedulingPayoutAlertRoutingKey = "backend-portal.scheduling.payout.alert"

	QrisRegistrationCallbackExchange = "snap-core-qris-registration-callback"
	CreditcardPaymentExchange        = "creditcard.payment"
	SnapCoreExchange                 = "snap-core"
	XbCoreExchange                   = "xb-core-processor"
	SnapTransfer                     = "snap-core.transfer"
	SnapTransferCutOffReportExchange = "snap-core.transfer-cutoff-report"
	SnapTransferReconcileExchange    = "snap-core.transfer.reconcile"
	ReconProcessExchange             = "backend-portal.recon"
	PaymentExpirationExchange        = "backend-portal.payment.expiration"
	InquiryAccountCallbackExchange   = "backend-portal.inquiry-account-callback"
	SubMerchantBulkCreateExchange    = "backend-portal.sub-merchants.bulk-create"
	VccSettlementInquiryExchange     = "backend-portal.vcc.settlement.inquiry"
	VccTerminalChargeExchange        = "backend-portal.vcc-terminal"
)

const (
	ActivityInsertRoutingKey                     = "backend-portal.activity.insert"
	CallbackRoutingKey                           = "backend-portal.callback.process"
	WorkflowCallbackRoutingKey                   = "backend-portal.workflow.callback.process"
	BulkDisbursementBatchCreateRoutingKey        = "backend-portal.bulk-disbursement.batch-create"
	BulkDisbursementBatchProcessRoutingKey       = "backend-portal.bulk-disbursement.batch-process"
	BulkDisbursementBatchDelayTransferRoutingKey = "backend-portal.bulk-disbursement.batch-delay-transfer"
	MerchantActionRoutingKey                     = "backend-portal.merchant.action"
	SlackPostWebhookRoutingKey                   = "backend-portal.slack.post-webhook"
	BulkCreateAccountRoutingKey                  = "backend-portal.account.bulk-create"
	CommServiceEmailRoutingKey                   = "backend-portal.comm-service.email"
	WithdrawalProcessRoutingKey                  = "backend-portal.withdrawal.process"
	RefundProcessRoutingKey                      = "backend-portal.refund.process"
	PaymentCaptureProcessRoutingKey              = "backend-portal.payment-capture.process"
	PayoutAlertProcessingRoutingKey              = "backend-portal.payout.alert.processing"
	PayoutAlertProcessingPendingKey              = "backend-portal.payout.alert.pending"

	SettlementProcessingRoutingKey = "settlement.processing.quorum"

	QrisRegistrationCallbackRoutingKey      = "snap.qris.registration-callback"
	CreditcardPaymentNotificationRoutingKey = "creditcard.payment.notification"
	SnapVAPaymentRoutingKey                 = "snap.va.payment"
	SnapQrisPaymentRoutingKey               = "snap.qris.payment"
	SnapEwalletPaymentRoutingKey            = "snap.ewallet.payment"
	SnapTransferStatusRoutingKey            = "snap-core.transfer.status"
	SnapTransferCutOffReportRoutingKey      = "snap-core.transfer.cutoff-report"
	SnapTransferReconcileRoutingKey         = "snap-core.transfer.reconcile"
	XbPayoutStatusChangeRoutingKey          = "xb.payout.status-change"
	ReconProcessRoutingKey                  = "backend-portal.recon.process"
	PaymentExpirationRoutingKey             = "backend-portal.payment.expiration.quorum" // Taking advantage of the delayed message plugin
	InquiryCallbackRoutingKey               = "backend-portal.inquiry-account-callback"
	SubMerchantBulkCreateRoutingKey         = "backend-portal.sub-merchants.bulk-create"
	VccSettlementInquiryRoutingKey          = "backend-portal.vcc.settlement.inquiry"
	VccTerminalChargeRoutingKey             = "backend-portal.vcc-terminal.charge"
)

const (
	ActivityInsertQueueName                     = "q.backend-portal.activity.insert"
	CallbackQueueName                           = "q.backend-portal.callback.process"
	WorkflowCallbackQueueName                   = "q.backend-portal.workflow.callback.process"
	BulkDisbursementBatchCreateQueueName        = "q.backend-portal.bulk-disbursement.batch-create"
	BulkDisbursementBatchProcessQueueName       = "q.backend-portal.bulk-disbursement.batch-process"
	BulkDisbursementBatchDelayTransferQueueName = "q.backend-portal.bulk-disbursement.batch-delay-transfer"
	MerchantActionQueueName                     = "q.backend-portal.merchant.action"
	OrchestratorTransactionQueueName            = "q.orchestrator.transaction.insert"
	SlackPostWebhookQueueName                   = "q.backend-portal.slack.post-webhook"
	UnroutedNotificationQueueName               = "q.unrouted.backend-portal.notifications"
	BulkCreateAccountQueueName                  = "q.backend-portal.account.bulk-create"
	QrisRegistrationCallbackQueueName           = "q.snap.qris.registration-callback"
	CommServiceEmailQueueName                   = "q.backend-portal.comm-service.email"
	WithdrawalProcessQueueName                  = "q.backend-portal.withdrawal.process"
	RefundProcessQueueName                      = "q.backend-portal.refund.process"
	PaymentCaptureProcessQueueName              = "q.backend-portal.payment-capture.process"

	SnapVAPaymentQueueName            = "q.snap.va.payment"
	SnapQrisPaymentQueueName          = "q.snap.qris.payment"
	SnapEwalletPaymentQueueName       = "q.snap.ewallet.payment"
	SnapTransferStatusQueueName       = "q.snap-core.transfer.status"
	SnapTransferCutOffReportQueueName = "q.snap-core.transfer.cutoff-report"
	CreditcardPaymentQueueName        = "q.creditcard.payment"
	XbPayoutStatusChangeQueueName     = "q.xb.payout.status-change"
	SnapTransferReconcileQueueName    = "q.snap-core.transfer.reconcile"

	SchedulingSettlementProcessQueueName    = "q.backend-portal.scheduling.settlement.process.quorum"
	SchedulingSettlementPendingQueueNameFmt = "q.backend-portal.scheduling.settlement.pending.%s.quorum" // {settlement_time} Unique identity for settlement time
	ReconProcessQueueName                   = "q.backend-portal.recon.process"
	ProcessPaymentExpirationQueueName       = "q.backend-portal.payment.expiration.process.quorum"
	SchedulingPayoutAlertProcessQueueName   = "q.backend-portal.scheduling.payout.alert.process.quorum"
	SchedulingPayoutAlertPendingQueueName   = "q.backend-portal.scheduling.payout.alert.pending.quorum"

	ReplyToQueueName               = "amq.rabbitmq.reply-to"
	HealthCheckQueueName           = "q.backend-portal.health-check.quorum"
	InquiryCallbackQueueName       = "q.backend-portal.inquiry-account-callback"
	SubMerchantBulkCreateQueueName = "q.backend-portal.sub-merchants.bulk-create"
	VccSettlementInquiryQueueName  = "q.backend-portal.vcc.settlement.inquiry"
	VccTerminalChargeQueueName     = "q.backend-portal.vcc-terminal.charges"
)

// Dead Letter Queue
const (
	// Activity
	ActivityDirectDLQExchange  = "dle.backend-portal.activity.direct"
	ActivityInsertDLQQueueName = "dlq.backend-portal.activity.insert"

	// Payment
	SnapCoreDLQExchange            = "dle.snap-core"
	SnapVAPaymentDLQQueueName      = "dlq.snap.va.payment"
	SnapQrisPaymentDLQQueueName    = "dlq.snap.qris.payment"
	SnapEwalletPaymentDLQQueueName = "dlq.snap.ewallet.payment"

	// Callback
	CallbackDLQExchange  = "dle.backend-portal.callback"
	CallbackDLQQueueName = "dlq.backend-portal.callback.process"

	// Bulk Disbursement
	BulkDisbursementDLQExchange              = "dle.backend-portal.bulk-disbursement"
	BulkDisbursementBatchCreateDLQQueueName  = "dlq.backend-portal.bulk-disbursement.batch-create"
	BulkDisbursementBatchProcessDLQQueueName = "dlq.backend-portal.bulk-disbursement.batch-process"
	SnapTransferStatusDLQExchange            = "dle.snap-core.transfer.status"
	SnapTransferStatusDLQQueueName           = "dlq.snap-core.transfer.status"
	SnapTransferCutOffReportDLQExchange      = "dle.snap-core.transfer.cutoff-report"
	SnapTransferCutOffReportDLQQueueName     = "dlq.snap-core.transfer.cutoff-report"
	SnapTransferReconcileDLExchange          = "dle.snap-core.transfer.reconcile"
	SnapTransferReconcileDLQueueName         = "dlq.snap-core.transfer.reconcile"

	// Orchestrator
	OrchestratorTransactionDLQExchange  = "dle.orchestrator"
	OrchestratorTransactionDLQQueueName = "dlq.orchestrator.transaction.insert"

	// Merchant
	MerchantDLQExchange  = "dle.backend-portal.merchant"        // Dead letter exchange
	MerchantDLQQueueName = "dlq.backend-portal.merchant.action" // Dead letter queue

	// Slack Webhook
	SlackPostWebhookDLQExchange  = "dle.backend-portal.slack.post-webhook"
	SlackPostWebhookDLQQueueName = "dlq.backend-portal.slack.post-webhook"

	// Creditcard
	CreditcardPaymentDLQExchange  = "dle.creditcard"
	CreditcardPaymentDLQQueueName = "dlq.creditcard.payment"

	// Account
	AccountDLQExchange            = "dle.backend-portal.account"
	BulkCreateAccountDLQQueueName = "dlq.backend-portal.account.bulk-create"

	// QRIS Registration
	QrisRegistrationCallbackDLXName = "dle.snap-core.qris.registration-callback"
	QrisRegistrationCallbackDLQName = "dlq.snap.qris.registration-callback"

	// DLE XbCore
	XbCoreDLExchange = "dle.xb-core-processor"

	// XB Payout - Status Change
	XbPayoutStatusChangeDLQueueName = "dlq.xb.payout.status-change"

	// Communication Service
	CommServiceDLExchange       = "dlx.backend-portal.comm-service"
	CommServiceEmailDLQueueName = "dlq.backend-portal.comm-service.email"

	// Recon Service
	ReconProcessDLExchange = "dle.backend-portal.recon.process"
	ReconProcessDLQQueue   = "dlq.backend-portal.recon.process"

	// Withdrawal Service
	WithdrawalDLExchange         = "dlx.backend-portal.withdrawals"
	WithdrawalProcessDLQueueName = "dlq.backend-portal.withdrawal.process"

	// Unified Payment
	PendingPaymentExpirationDLQExchange  = "dle.backend-portal.payment.expiration"
	PendingPaymentExpirationDLQQueueName = "dlq.backend-portal.payment.expiration.pending"

	// Notification
	NotificationDLExchange  = "dlx.backend-portal.notifications"
	NotificationDLQueueName = "dlq.backend-portal.notifications"

	// Refund
	RefundDLExchange         = "dle.backend-portal.refund"
	RefundProcessDLQueueName = "dlq.backend-portal.refund.process"

	// Payment Capture
	PaymentCaptureDLExchange         = "dle.backend-portal.payment-capture"
	PaymentCaptureProcessDLQueueName = "dlq.backend-portal.payment-capture.process"

	// Bulk Create Submerchant
	SubMerchantBulkCreateDLQueueName = "dlq.backend-portal.sub-merchants.bulk-create"
	SubMerchantBulkCreateDLExchange  = "dle.backend-portal.sub-merchants.bulk-create"

	// VCC Settlement Inquiry
	VccSettlementInquiryDLQueueName = "dlq.backend-portal.vcc.settlement.inquiry"
	VccSettlementInquiryDLExchange  = "dle.backend-portal.vcc.settlement.inquiry"

	// VCC Terminal Charge
	VccTerminalChargeDLExchange  = "dlx.backend-portal.vcc-terminal"
	VccTerminalChargeDLQueueName = "dlq.backend-portal.vcc-terminal.charges"
)

const (
	DirectExchangeType  = "direct"
	FanoutExchangeType  = "fanout"
	DelayedExchangeType = "x-delayed-message"

	QuorumQueueType = "quorum"

	HeaderTraceId            = "trace-id"
	HeaderXParentMerchantId  = "X-Parent-Merchant-Id"
	HeaderXDerivedMerchantId = "X-Derived-Merchant-Id" // Derived merchant can be parents or another sub merchant KYC
	HeaderXMerchantId        = "X-Merchant-Id"
	HeaderXDeliveryCount     = "x-delivery-count"
	HeaderXRetryCount        = "X-Retry-Count"

	PluginHeaderXDelay = "x-delay"

	UnroutedNotificationMsgTTL    = 5 * time.Second
	UnroutedNotificationMaxLength = 1_000
)
