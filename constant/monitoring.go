package constant

const (
	MetricInstrumentTypeCounter   = "COUNTER"
	MetricInstrumentTypeGauge     = "GAUGE"
	MetricInstrumentTypeHistogram = "HISTOGRAM"
)

const ( // In otel, this will set the Meter Name (lowercase)
	ComponentNameUnifiedPayment = "unified-payment"
	ComponentNameInquiryAccount = "inquiry-account"
	ComponentNamePayout         = "payout"
	ComponentNameAccount        = "account"
	ComponentNameXB             = "xb"
	ComponentMerchantCallback   = "merchant-callback"
)

const ( // In otel, this will set the Meter Name (lowercase)
	MetricNameUnifiedPaymentConfirmSession      = "unified-payment.payment.confirm-session"
	MetricNameUnifiedPaymentPaymentProcessed    = "unified-payment.payment.processed"
	MetricNameUnifiedPaymentRefund              = "unified-payment.payment.refund"
	MetricNameUnifiedPaymentRefundProcessed     = "unified-payment.payment.refund.processed"
	MetricNameUnifiedPaymentExpired             = "unified-payment.payment.expired"
	MetricNameUnifiedPaymentSettlementProcessed = "unified-payment.payment.settlement.processed"
	MetricNameUnifiedPaymentCaptureProcessed    = "unified-payment.payment.capture.processed"
	MetricNameInquiryAccount                    = "inquiry-account"
	MetricNamePayout                            = "payout"
	MetricNamePayoutProcessed                   = "payout.processed"
	MetricNameAccountBalance                    = "account.balance"
	MetricNameXBCreatePayout                    = "xb.payout.create"
	MetricNameXBUploadPayout                    = "xb.payout.upload"
	MetricNameXBConfirmPayout                   = "xb.payout.confirm"
	MetricNameXBSubmitRFI                       = "xb.payout.submit-rfi"
	MetricNameXBUpdateStatus                    = "xb.payout.update-status"
	MetricNameMerchantCallbackCount             = "merchant-callback.process.count"
	MetricNameMerchantCallbackDuration          = "merchant-callback.process.duration"
	MetricNameMerchantCallbackRetryCount        = "merchant-callback.process.retry-count"
)
