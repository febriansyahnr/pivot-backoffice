package rabbitMqExt

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	messagingQueueModel "github.com/paper-indonesia/pivot-backoffice/internal/model/messagingQueue"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/notification"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/callback"
	pdkLogger "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/paper-indonesia/pdk/v2/amqp"
	pdkNewRelic "github.com/paper-indonesia/pdk/v2/newRelicExt"
)

type IRabbitMQExt interface {
	HealthCheck(ctx context.Context) error
	PushNotification(ctx context.Context, data *notification.PushNotification) (err error)
	PublishActivity(ctx context.Context, merchantID, userID *string, tag, activity string, parameter interface{}) error
	Consume(ctx context.Context, document string, exchangeType *string, process func(context.Context, []byte, string) error) error
	ConsumeWithOpts(ctx context.Context, config *ConsumeOptsConfig, handler HandlerFunc, opts ...ChannelOpt) error
	Publish(ctx context.Context, document string, exchangeType *string, message interface{}, failedMsg ...amqp.Delivery) error
	PublishWithTTL(ctx context.Context, document string, message interface{}, ttl time.Duration) error
	PublishWithDelay(ctx context.Context, document string, message interface{}, delay time.Duration) error

	PublishAndWaitReply(ctx context.Context, document string, message interface{}) (event *amqp.Event, close io.Closer, err error)
	PublishToReplyQueue(ctx context.Context, replyToAddress string, msg amqp.Publishing) error
	PublishForSettlementProcess(ctx context.Context, payload messagingQueueModel.PublishSettlementProcessPayload) error
	PublishMerchantCallback(ctx context.Context, payload *callback.ProcessCallbackRequest) error

	Close() error
}

type (
	Delivery   = amqp.Delivery
	Table      = amqp.Table
	Publishing = amqp.Publishing
)

type rabbitMQExt struct {
	url        string
	connection *amqp.Connection
	logger     pdkLogger.ILogger
	once       *once
	config     config.RabbitMQConfig[string]
	secret     config.RabbitMQSecret

	nrApp pdkNewRelic.INewRelicExt
}

type once struct {
	directReplyTo *sync.Map
}

var pool = &sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

type ConsumeOptsConfig struct {
	QueueName      string
	Args           Table
	RetryMechanism bool
	RetryLimit     uint64
}

type ChannelOpt func(*amqp.Channel) error
type HandlerFunc func(context.Context, *Delivery) error

func New(
	config config.RabbitMQConfig[string],
	secret config.RabbitMQSecret,
	logger pdkLogger.ILogger,
	nrApp pdkNewRelic.INewRelicExt,
) (IRabbitMQExt, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s", secret.Username, secret.Password, config.Host, config.Port)
	if config.VHost != "" {
		url = fmt.Sprintf("amqp://%s:%s@%s:%s/%s", secret.Username, secret.Password, config.Host, config.Port, config.VHost)
	}

	if config.HeartbeatInSeconds <= 0 {
		config.HeartbeatInSeconds = 10
	}

	dialConfig := amqp.Config{
		ClientName: config.ServiceName,
		Locale:     amqp.DefaultLocale,
		Heartbeat:  time.Duration(config.HeartbeatInSeconds) * time.Second,
	}
	conn, err := amqp.DialConfig(url, dialConfig, amqp.WithConnectionHealthCheck(HealthCheckQueueName))
	if err != nil {
		return nil, err
	}

	return &rabbitMQExt{
		connection: conn,
		url:        url,
		logger:     logger,
		once: &once{
			directReplyTo: new(sync.Map),
		},
		config: config,
		secret: secret,
		nrApp:  nrApp,
	}, nil
}

func (r *rabbitMQExt) getChannel() (*amqp.Channel, error) {
	if r.connection == nil || r.connection.IsClosed() {
		url := fmt.Sprintf("amqp://%s:%s@%s:%s", r.secret.Username, r.secret.Password, r.config.Host, r.config.Port)
		if r.config.VHost != "" {
			url = fmt.Sprintf("amqp://%s:%s@%s:%s/%s", r.secret.Username, r.secret.Password, r.config.Host, r.config.Port, r.config.VHost)
		}

		dialConfig := amqp.Config{
			ClientName: r.config.ServiceName,
			Locale:     amqp.DefaultLocale,
			Heartbeat:  time.Duration(r.config.HeartbeatInSeconds) * time.Second,
		}

		conn, err := amqp.DialConfig(url, dialConfig, amqp.WithConnectionHealthCheck(HealthCheckQueueName))
		if err != nil {
			return nil, err
		}
		r.connection = conn
	}

	ch, err := r.connection.Channel()
	if err != nil {
		return nil, err
	}

	return ch, nil
}

func (r *rabbitMQExt) HealthCheck(ctx context.Context) error {
	return r.connection.HealthCheck(ctx)
}

func (r *rabbitMQExt) Close() error {
	defer log.Println("RabbitMQ health check stopped")

	return r.connection.Close()
}

func routingRabbitMqConfig(channel string) RabbitMqExchangeConfig {
	var result RabbitMqExchangeConfig

	switch channel {
	case SnapVAPaymentRoutingKey:
		result = RabbitMqExchangeConfig{
			Exchange:     SnapCoreExchange,
			QueueName:    SnapVAPaymentQueueName,
			RoutingKey:   SnapVAPaymentRoutingKey,
			Channel:      constant.ChannelVirtualAccount,
			DLQExchange:  SnapCoreDLQExchange,
			DLQQueueName: SnapVAPaymentDLQQueueName,
		}

	case SnapQrisPaymentRoutingKey:
		result = RabbitMqExchangeConfig{
			Exchange:     SnapCoreExchange,
			QueueName:    SnapQrisPaymentQueueName,
			RoutingKey:   SnapQrisPaymentRoutingKey,
			Channel:      constant.ChannelQris,
			DLQExchange:  SnapCoreDLQExchange,
			DLQQueueName: SnapQrisPaymentDLQQueueName,
		}

	case SnapEwalletPaymentRoutingKey:
		result = RabbitMqExchangeConfig{
			Exchange:     SnapCoreExchange,
			QueueName:    SnapEwalletPaymentQueueName,
			RoutingKey:   SnapEwalletPaymentRoutingKey,
			Channel:      constant.ChannelEwallet,
			DLQExchange:  SnapCoreDLQExchange,
			DLQQueueName: SnapEwalletPaymentDLQQueueName,
		}

	case CallbackRoutingKey:
		result = RabbitMqExchangeConfig{
			Exchange:     CallbackExchange,
			QueueName:    CallbackQueueName,
			RoutingKey:   CallbackRoutingKey,
			Channel:      "",
			DLQExchange:  CallbackDLQExchange,
			DLQQueueName: CallbackDLQQueueName,
		}

	case WorkflowCallbackRoutingKey:
		result = RabbitMqExchangeConfig{
			Exchange:   CallbackExchange,
			QueueName:  WorkflowCallbackQueueName,
			RoutingKey: WorkflowCallbackRoutingKey,
		}

	case MerchantActionRoutingKey:
		result = RabbitMqExchangeConfig{
			Exchange:     MerchantExchange,
			QueueName:    MerchantActionQueueName,
			RoutingKey:   MerchantActionRoutingKey,
			Channel:      "",
			DLQExchange:  MerchantDLQExchange,
			DLQQueueName: MerchantDLQQueueName,
		}

	case ActivityInsertRoutingKey:
		result = RabbitMqExchangeConfig{
			Exchange:     ActivityDirectExchange,
			QueueName:    ActivityInsertQueueName,
			RoutingKey:   ActivityInsertRoutingKey,
			Channel:      "",
			DLQExchange:  ActivityDirectDLQExchange,
			DLQQueueName: ActivityInsertDLQQueueName,
		}

	case BulkDisbursementBatchCreateRoutingKey:
		result = RabbitMqExchangeConfig{
			Exchange:     BulkDisbursementExchange,
			QueueName:    BulkDisbursementBatchCreateQueueName,
			RoutingKey:   BulkDisbursementBatchCreateRoutingKey,
			Channel:      "",
			DLQExchange:  BulkDisbursementDLQExchange,
			DLQQueueName: BulkDisbursementBatchCreateDLQQueueName,
		}

	case BulkDisbursementBatchProcessRoutingKey:
		result = RabbitMqExchangeConfig{
			Exchange:     BulkDisbursementExchange,
			QueueName:    BulkDisbursementBatchProcessQueueName,
			RoutingKey:   BulkDisbursementBatchProcessRoutingKey,
			Channel:      "",
			DLQExchange:  BulkDisbursementDLQExchange,
			DLQQueueName: BulkDisbursementBatchProcessDLQQueueName,
		}

	case SlackPostWebhookRoutingKey:
		result = RabbitMqExchangeConfig{
			Exchange:     SlackExchange,
			QueueName:    SlackPostWebhookQueueName,
			RoutingKey:   SlackPostWebhookRoutingKey,
			Channel:      constant.ChannelSlackNotifier,
			DLQExchange:  SlackPostWebhookDLQExchange,
			DLQQueueName: SlackPostWebhookDLQQueueName,
		}

	case CreditcardPaymentNotificationRoutingKey:
		result = RabbitMqExchangeConfig{
			Exchange:     CreditcardPaymentExchange,
			QueueName:    CreditcardPaymentQueueName,
			RoutingKey:   CreditcardPaymentNotificationRoutingKey,
			Channel:      "",
			DLQExchange:  CreditcardPaymentDLQExchange,
			DLQQueueName: CreditcardPaymentDLQQueueName,
		}
	case BulkCreateAccountRoutingKey:
		result = RabbitMqExchangeConfig{
			Exchange:     AccountExchange,
			QueueName:    BulkCreateAccountQueueName,
			RoutingKey:   BulkCreateAccountRoutingKey,
			Channel:      "",
			DLQExchange:  AccountDLQExchange,
			DLQQueueName: BulkCreateAccountDLQQueueName,
		}

	case QrisRegistrationCallbackRoutingKey:
		result = RabbitMqExchangeConfig{
			Exchange:     QrisRegistrationCallbackExchange,
			QueueName:    QrisRegistrationCallbackQueueName,
			RoutingKey:   QrisRegistrationCallbackRoutingKey,
			Channel:      "",
			DLQExchange:  QrisRegistrationCallbackDLXName,
			DLQQueueName: QrisRegistrationCallbackDLQName,
		}

	case XbPayoutStatusChangeRoutingKey:
		result = RabbitMqExchangeConfig{
			Exchange:     XbCoreExchange,
			QueueName:    XbPayoutStatusChangeQueueName,
			RoutingKey:   XbPayoutStatusChangeRoutingKey,
			Channel:      constant.ChannelXB,
			DLQExchange:  XbCoreDLExchange,
			DLQQueueName: XbPayoutStatusChangeDLQueueName,
		}

	case SettlementProcessingRoutingKey:
		result = RabbitMqExchangeConfig{
			Exchange:    SchedulingSettlementExchange,
			QueueName:   SchedulingSettlementProcessQueueName,
			RoutingKey:  SettlementProcessingRoutingKey,
			Channel:     "",
			DLQExchange: "",
		}

	case CommServiceEmailRoutingKey:
		result = RabbitMqExchangeConfig{
			Exchange:     CommServiceExchange,
			QueueName:    CommServiceEmailQueueName,
			RoutingKey:   CommServiceEmailRoutingKey,
			Channel:      "",
			DLQExchange:  CommServiceDLExchange,
			DLQQueueName: CommServiceEmailDLQueueName,
		}

	case SnapTransferStatusRoutingKey:
		result = RabbitMqExchangeConfig{
			Exchange:     SnapTransfer,
			QueueName:    SnapTransferStatusQueueName,
			RoutingKey:   SnapTransferStatusRoutingKey,
			Channel:      "",
			DLQExchange:  SnapTransferStatusDLQExchange,
			DLQQueueName: SnapTransferStatusDLQQueueName,
		}

	case SnapTransferCutOffReportRoutingKey:
		result = RabbitMqExchangeConfig{
			Exchange:     SnapTransferCutOffReportExchange,
			QueueName:    SnapTransferCutOffReportQueueName,
			RoutingKey:   SnapTransferCutOffReportRoutingKey,
			Channel:      "",
			DLQExchange:  SnapTransferCutOffReportDLQExchange,
			DLQQueueName: SnapTransferCutOffReportDLQQueueName,
		}

	case ReconProcessRoutingKey:
		result = RabbitMqExchangeConfig{
			Exchange:     ReconProcessExchange,
			QueueName:    ReconProcessQueueName,
			RoutingKey:   ReconProcessRoutingKey,
			Channel:      "",
			DLQExchange:  ReconProcessDLExchange,
			DLQQueueName: ReconProcessDLQQueue,
		}

	case WithdrawalProcessRoutingKey:
		result = RabbitMqExchangeConfig{
			Exchange:     WithdrawalExchange,
			QueueName:    WithdrawalProcessQueueName,
			RoutingKey:   WithdrawalProcessRoutingKey,
			Channel:      "",
			DLQExchange:  WithdrawalDLExchange,
			DLQQueueName: WithdrawalProcessDLQueueName,
		}
	case PaymentExpirationRoutingKey:
		result = RabbitMqExchangeConfig{
			Exchange:     PaymentExpirationExchange,
			QueueName:    ProcessPaymentExpirationQueueName,
			RoutingKey:   PaymentExpirationRoutingKey,
			Channel:      "",
			DLQExchange:  "",
			DLQQueueName: "",
		}
	case InquiryCallbackRoutingKey:
		result = RabbitMqExchangeConfig{
			Exchange:     InquiryAccountCallbackExchange,
			QueueName:    InquiryCallbackQueueName,
			RoutingKey:   InquiryCallbackRoutingKey,
			Channel:      "",
			DLQExchange:  "",
			DLQQueueName: "",
		}

	case BulkDisbursementBatchDelayTransferRoutingKey:
		result = RabbitMqExchangeConfig{
			Exchange:      BulkDisbursementExchange,
			QueueName:     BulkDisbursementBatchDelayTransferQueueName,
			RoutingKey:    BulkDisbursementBatchDelayTransferRoutingKey,
			DLQExchange:   BulkDisbursementExchange,
			DLQRoutingKey: BulkDisbursementBatchProcessRoutingKey,
		}

	case RefundProcessRoutingKey:
		result = RabbitMqExchangeConfig{
			Exchange:     RefundExchange,
			QueueName:    RefundProcessQueueName,
			RoutingKey:   RefundProcessRoutingKey,
			DLQExchange:  RefundDLExchange,
			DLQQueueName: RefundProcessDLQueueName,
		}

	case PaymentCaptureProcessRoutingKey:
		result = RabbitMqExchangeConfig{
			Exchange:     PaymentCaptureExchange,
			QueueName:    PaymentCaptureProcessQueueName,
			RoutingKey:   PaymentCaptureProcessRoutingKey,
			DLQExchange:  PaymentCaptureDLExchange,
			DLQQueueName: PaymentCaptureProcessDLQueueName,
		}

	case SnapTransferReconcileRoutingKey:
		result = RabbitMqExchangeConfig{
			Exchange:     SnapTransferReconcileExchange,
			QueueName:    SnapTransferReconcileQueueName,
			RoutingKey:   SnapTransferReconcileRoutingKey,
			DLQExchange:  SnapTransferReconcileDLExchange,
			DLQQueueName: SnapTransferReconcileDLQueueName,
		}

	case PayoutAlertProcessingRoutingKey:
		result = RabbitMqExchangeConfig{
			Exchange:      SchedulingPayoutAlertRoutingKey,
			QueueName:     SchedulingPayoutAlertProcessQueueName,
			RoutingKey:    PayoutAlertProcessingRoutingKey,
			Channel:       "",
			DLQExchange:   "",
			DLQQueueName:  SchedulingPayoutAlertPendingQueueName,
			DLQRoutingKey: PayoutAlertProcessingPendingKey, // Previously, this configuration was set as an empty string.
		}

	case SubMerchantBulkCreateRoutingKey:
		result = RabbitMqExchangeConfig{
			Exchange:     SubMerchantBulkCreateExchange,
			QueueName:    SubMerchantBulkCreateQueueName,
			RoutingKey:   SubMerchantBulkCreateRoutingKey,
			DLQExchange:  SubMerchantBulkCreateDLExchange,
			DLQQueueName: SubMerchantBulkCreateDLQueueName,
		}

	case VccSettlementInquiryRoutingKey:
		result = RabbitMqExchangeConfig{
			Exchange:     VccSettlementInquiryExchange,
			QueueName:    VccSettlementInquiryQueueName,
			RoutingKey:   VccSettlementInquiryRoutingKey,
			DLQExchange:  VccSettlementInquiryDLExchange,
			DLQQueueName: VccSettlementInquiryDLQueueName,
		}

	case VccTerminalChargeRoutingKey:
		result = RabbitMqExchangeConfig{
			Exchange:     VccTerminalChargeExchange,
			QueueName:    VccTerminalChargeQueueName,
			RoutingKey:   VccTerminalChargeRoutingKey,
			DLQExchange:  VccTerminalChargeDLExchange,
			DLQQueueName: VccTerminalChargeDLQueueName,
		}

	default:
		return result
	}

	return result
}

func (r *rabbitMQExt) incrementRetryCountAndPrepareMessage(msg amqp.Delivery) amqp.Publishing {
	retryCount, ok := msg.Headers["x-delivery-count"].(int)
	if !ok {
		retryCount = 0
	}
	retryCount++

	headers := amqp.Table{
		"x-delivery-count": retryCount,
	}

	publishing := amqp.Publishing{
		ContentType: "text/plain",
		Body:        msg.Body,
		Headers:     headers,
	}

	return publishing
}

func ConsumerSetupNotification(ch *amqp.Channel) error {
	err := ch.ExchangeDeclare(NotificationDLExchange, amqp.ExchangeTopic, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("DLX Declare: %w", err)
	}

	_, err = ch.QueueDeclare(
		NotificationDLQueueName, true, false, false, false, amqp.Table{"x-queue-type": amqp.QueueTypeQuorum},
	)
	if err != nil {
		return fmt.Errorf("DLQ Declare: %w", err)
	}

	if err = ch.QueueBind(NotificationDLQueueName, "*", NotificationDLExchange, false, nil); err != nil {
		return fmt.Errorf("Binding Queue: %w", err)
	}
	return ch.Qos(1, 0, false)
}
