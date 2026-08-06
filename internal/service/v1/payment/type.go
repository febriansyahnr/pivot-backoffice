package paymentService

import (
	"context"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository/datamart"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/encryption"
	"github.com/paper-indonesia/pivot-backoffice/pkg/gcs"
	jwtExt "github.com/paper-indonesia/pivot-backoffice/pkg/jwt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/vault"

	"github.com/paper-indonesia/pdk/v2/encrypt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
	"golang.org/x/sync/singleflight"
)

var (
	loc, _     = time.LoadLocation(constant.TimeLoc)
	otelTracer = otel.Tracer("PaymentService")

	downloadHeaders = []string{
		"Created Date", "Method", "Type", "Bill Amount", "Payment Status", "Payment Date", "Payment Channel", "Expiry Time", "Paid Amount", "Customer No", "Reference ID", "Transaction ID", "Recurring ID", "Bank Reference", "VA Name", "VA Number", "QR Merchant Name", "QR URL", "Acquiring Bank", "Bank Merchant ID (MID)", "Card Issuer Bank", "Card Number", "Card Expiry Date", "Refund Date", "Refund Amount", "Refund Status",
	}
	allowedStatusForPaymentUpdate = []string{
		paymentConstant.PAYMENT_STATUS_PENDING,
		paymentConstant.UnifiedPaymentStatusWaitingForPayment,
	}
	allowedPaymentMethods = []string{
		paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
		paymentConstant.PAYMENT_METHOD_QRIS,
		paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
	}
)

const proofOfTransactionSignedURLExpiry = 10 * time.Minute

type PaymentService struct {
	paymentRepo            repository.IPaymentRepository
	logger                 logger.ILogger
	snapCoreRepo           repository.ISnapCoreRepository
	customerRepo           repository.ICustomerRepository
	merchantRepo           repository.IMerchantRepository
	paymentMethodRepo      repository.IPaymentMethodRepository
	accountRepo            repository.IAccountRepository
	accountTransactionRepo repository.IAccountTransactionRepository
	statusHistoriesRepo    repository.IStatusHistoriesRepository
	disbursementRepo       repository.IDisbursementRepository

	orchestratorSvc   service.IOrchestratorService
	qrisSvc           service.IQrisService
	feeSvc            service.IFeeService
	merchantSvc       service.IMerchantService
	transferSvc       service.ITransferService
	internal          service.IPaymentInternalDirectFunc
	creditCardSvc     service.ICreditCardService
	paymentMethodSvc  service.IPaymentMethodService
	ledgerSvc         service.ILedgerService
	unifiedPaymentSvc service.IUnifiedPaymentService
	refundSvc         service.IRefundService
	settlementHoldSvc service.ISettlementHoldService

	rabbitMqExt        rabbitMqExt.IRabbitMQExt
	config             *config.Config
	gcs                gcs.IGCSService
	redis              redisExt.IRedisExt
	jwt                jwtExt.IJwt
	paymentMetricsRepo datamart.IDatamartPaymentMetrics
	secretManager      vault.IVaultKeyValue
	cryptoAesGcm       encrypt.Encrypter
	cryptoProvider     encryption.CryptoProvider
	validator          *validatorExt.Validate
	receiptSf          singleflight.Group
}

type PaymentServiceFunc func(*PaymentService)

const serviceName = "Payment-Service"

func New(
	paymentRepo repository.IPaymentRepository,
	logger logger.ILogger,
	snapCoreRepo repository.ISnapCoreRepository,
	customerRepo repository.ICustomerRepository,
	merchantRepo repository.IMerchantRepository,
	paymentMethodRepo repository.IPaymentMethodRepository,
	accountRepo repository.IAccountRepository,
	depends ...PaymentServiceFunc,
) *PaymentService {
	s := &PaymentService{
		paymentRepo:       paymentRepo,
		logger:            logger,
		snapCoreRepo:      snapCoreRepo,
		customerRepo:      customerRepo,
		merchantRepo:      merchantRepo,
		paymentMethodRepo: paymentMethodRepo,
		accountRepo:       accountRepo,
	}
	s.internal = s

	for _, fn := range depends {
		fn(s)
	}

	return s
}

func WithOrchestratorService(svc service.IOrchestratorService) PaymentServiceFunc {
	return func(ds *PaymentService) {
		ds.orchestratorSvc = svc
	}
}

func WithQrisService(svc service.IQrisService) PaymentServiceFunc {
	return func(ds *PaymentService) {
		ds.qrisSvc = svc
	}
}

func WithRabbitMQClient(rmq rabbitMqExt.IRabbitMQExt) PaymentServiceFunc {
	return func(ds *PaymentService) {
		ds.rabbitMqExt = rmq
	}
}

func WithConfig(c *config.Config) PaymentServiceFunc {
	return func(ds *PaymentService) {
		ds.config = c
	}
}

func WithFeeService(svc service.IFeeService) PaymentServiceFunc {
	return func(ds *PaymentService) {
		ds.feeSvc = svc
	}
}

func WithMerchantService(svc service.IMerchantService) PaymentServiceFunc {
	return func(ds *PaymentService) {
		ds.merchantSvc = svc
	}
}

func WithAccountTransactionRepository(repo repository.IAccountTransactionRepository) PaymentServiceFunc {
	return func(ds *PaymentService) {
		ds.accountTransactionRepo = repo
	}
}

func WithTransferService(service service.ITransferService) PaymentServiceFunc {
	return func(ps *PaymentService) {
		ps.transferSvc = service
	}
}

func WithGCSService(gcs gcs.IGCSService) PaymentServiceFunc {
	return func(ps *PaymentService) {
		ps.gcs = gcs
	}
}

func WithRedisClient(rdb redisExt.IRedisExt) PaymentServiceFunc {
	return func(ps *PaymentService) {
		ps.redis = rdb
	}
}

func WithInternalDirectFunc(funcs service.IPaymentInternalDirectFunc) PaymentServiceFunc {
	return func(ps *PaymentService) {
		ps.internal = funcs
	}
}

func WithJWTExt(jwt jwtExt.IJwt) PaymentServiceFunc {
	return func(ds *PaymentService) {
		ds.jwt = jwt
	}
}

func WithCreditCardService(svc service.ICreditCardService) PaymentServiceFunc {
	return func(ds *PaymentService) {
		ds.creditCardSvc = svc
	}
}

func WithPaymentMethodService(svc service.IPaymentMethodService) PaymentServiceFunc {
	return func(ps *PaymentService) {
		ps.paymentMethodSvc = svc
	}
}

func WithLedgerService(svc service.ILedgerService) PaymentServiceFunc {
	return func(ps *PaymentService) {
		ps.ledgerSvc = svc
	}
}

func WithUnifiedPaymentService(svc service.IPaymentService, unifiedSvc service.IUnifiedPaymentService) {
	svc.(*PaymentService).unifiedPaymentSvc = unifiedSvc
}

func WithRefundService(svc service.IRefundService) PaymentServiceFunc {
	return func(ps *PaymentService) {
		ps.refundSvc = svc
	}
}

func WithStatusHistoriesRepository(repo repository.IStatusHistoriesRepository) PaymentServiceFunc {
	return func(ps *PaymentService) {
		ps.statusHistoriesRepo = repo
	}
}

func WithPaymentMetricsRepository(repo datamart.IDatamartPaymentMetrics) PaymentServiceFunc {
	return func(ps *PaymentService) {
		ps.paymentMetricsRepo = repo
	}
}

func WithSecretManager(kv vault.IVaultKeyValue) PaymentServiceFunc {
	return func(ps *PaymentService) {
		ps.secretManager = kv
	}
}

func WithCryptoAesGcm(aesGcm encrypt.Encrypter) PaymentServiceFunc {
	return func(ps *PaymentService) {
		ps.cryptoAesGcm = aesGcm
	}
}

func WithCryptoProvider(provider encryption.CryptoProvider) PaymentServiceFunc {
	return func(ps *PaymentService) {
		ps.cryptoProvider = provider
	}
}

func WithValidator(vld *validatorExt.Validate) PaymentServiceFunc {
	return func(ps *PaymentService) {
		ps.validator = vld
	}
}

func WithSettlementHoldService(svc service.IPaymentService, settlementHoldSvc service.ISettlementHoldService) {
	svc.(*PaymentService).settlementHoldSvc = settlementHoldSvc
}

func WithDisbursementRepository(repo repository.IDisbursementRepository) PaymentServiceFunc {
	return func(ps *PaymentService) {
		ps.disbursementRepo = repo
	}
}

func NewPaymentLedgerService(
	c *config.Config,
	logger logger.ILogger,
	paymentRepo repository.IPaymentRepository,
	merchantRepo repository.IMerchantRepository,
	accountTransactionRepo repository.IAccountTransactionRepository,
	orchestratorSvc service.IOrchestratorService,
	feeSvc service.IFeeService,
	transferSvc service.ITransferService,
	ledgerSvc service.ILedgerService,
	rabbitMqExt rabbitMqExt.IRabbitMQExt,
) service.IPaymentLedgerService {
	return &PaymentService{
		config:                 c,
		logger:                 logger,
		paymentRepo:            paymentRepo,
		merchantRepo:           merchantRepo,
		accountTransactionRepo: accountTransactionRepo,
		orchestratorSvc:        orchestratorSvc,
		feeSvc:                 feeSvc,
		transferSvc:            transferSvc,
		ledgerSvc:              ledgerSvc,
		rabbitMqExt:            rabbitMqExt,
	}
}

// RecordPaymentStatusHistory records payment status history synchronously
func (s *PaymentService) RecordPaymentStatusHistory(ctx context.Context, paymentID, actor, statusType string) {
	if s.statusHistoriesRepo == nil {
		return
	}

	switch statusType {
	case constant.PaymentStatusHistoryPending:
		s.recordPaymentPending(ctx, paymentID, actor)
	case constant.PaymentStatusHistoryRequirePaymentMethod:
		s.recordPaymentRequirePaymentMethod(ctx, paymentID, actor)
	case constant.PaymentStatusHistoryRequireConfirmation:
		s.recordPaymentRequireConfirmation(ctx, paymentID, actor)
	case constant.PaymentStatusHistoryRequireAction:
		s.recordPaymentRequireAction(ctx, paymentID, actor)
	case constant.PaymentStatusHistoryProcessing:
		s.recordPaymentProcessing(ctx, paymentID, actor)
	case constant.PaymentStatusHistorySuccess:
		s.recordPaymentSuccess(ctx, paymentID, actor)
	case constant.PaymentStatusHistoryPaid:
		s.recordPaymentPaid(ctx, paymentID, actor)
	case constant.PaymentStatusHistoryVoid:
		s.recordPaymentVoid(ctx, paymentID, actor)
	case constant.PaymentStatusHistoryExpired:
		s.recordPaymentExpired(ctx, paymentID, actor)
	case constant.PaymentStatusHistoryCancelled:
		s.recordPaymentCancelled(ctx, paymentID, actor)
	case constant.PaymentStatusHistoryInvestigationInProcess:
		s.recordInvestigationInProcess(ctx, paymentID, actor)
	case constant.PaymentStatusHistoryInvestigationSuccess:
		s.recordInvestigationSuccess(ctx, paymentID, actor)
	case constant.PaymentStatusHistoryInvestigationFailed:
		s.recordInvestigationFailed(ctx, paymentID, actor)
	default:
		s.logger.Warn(ctx, "unknown payment status type, skipping status history recording",
			logger.String("statusType", statusType),
			logger.String("paymentID", paymentID),
			logger.String("actor", actor))
	}
}
