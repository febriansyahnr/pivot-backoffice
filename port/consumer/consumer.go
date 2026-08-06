package consumer

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
)

type ICallbackConsumer interface {
	ProcessCallback(ctx context.Context, body []byte, channel string) error
}

type IActivityConsumer interface {
	Insert(ctx context.Context, body []byte, channel string) error
}

type PaymentConsumer interface {
	ProcessPaymentNotification(ctx context.Context, body []byte, channel string) error
	ProcessPaymentExpiration(ctx context.Context, body []byte, channel string) error
	VCCTerminalSubmitCharge(ctx context.Context, body []byte, _ string) error
}

type IDisbursementConsumer interface {
	BatchCreateDisbursement(ctx context.Context, body []byte, channel string) error
	BatchProcessDisbursement(ctx context.Context, body []byte, channel string) error
	PayoutTransactionAlertProcess(ctx context.Context, body []byte, channel string) (err error)
}

type ISlackConsumer interface {
	ProcessSlackPostWebhook(ctx context.Context, body []byte, channel string) error
}

type ICreditCardService interface {
	PaymentNotification(ctx context.Context, body []byte, channel string) error
}

type IAccountConsumer interface {
	BulkCreateAccount(ctx context.Context, body []byte, channel string) error
}

type IProcessor interface {
	Process(ctx context.Context, body []byte, channel string) error
}

type IXbPayoutConsumer interface {
	UpdateStatus(ctx context.Context, body []byte, channel string) error
}

type ISettlementConsumer interface {
	ProcessPaymentSettlement(ctx context.Context, body []byte, channel string) error
}

type CommServiceHandler interface {
	PostEmailHandler(ctx context.Context, body []byte, _ string) error
}

type IWithdrawalConsumer interface {
	WithdrawalProcess(ctx context.Context, body []byte, _ string) error
}

type IReconciliationConsumer interface {
	ReconciliationProcess(ctx context.Context, body []byte, _ string) error
	SnapCoreTransferReconcile(ctx context.Context, body []byte, _ string) error
}

type INotificationConsumer interface {
	RetryNotification(ctx context.Context, msg *rabbitMqExt.Delivery) error
}
type IDirectReplyConsumer interface {
	AddressingWaitReplyAccountInquiry(ctx context.Context, body []byte, channel string) error
}

type IRefundConsumer interface {
	RefundProcess(ctx context.Context, body []byte, _ string) error
}

type IBankTransferConsumer interface {
	UpdateTransferStatus(ctx context.Context, body []byte, _ string) error
	CutOffReportTrigger(ctx context.Context, body []byte, _ string) error
}

type IMerchantConsumer interface {
	ProcessBulkCreateSubMerchant(ctx context.Context, body []byte, channel string) error
}

type IPaymentCaptureConsumer interface {
	PaymentCaptureProcess(ctx context.Context, body []byte, _ string) error
}

type IVccSettlementConsumer interface {
	ProcessSettlementTransactionInquiry(ctx context.Context, body []byte, _ string) error
}

type IReportingConsumer interface {
	BalanceHistory(context.Context, []byte) error
}
