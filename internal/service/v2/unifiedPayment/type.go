package unifiedPaymentService

import (
	"context"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	services "github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/encryption"
	"github.com/paper-indonesia/pivot-backoffice/pkg/fds"
	"github.com/paper-indonesia/pivot-backoffice/pkg/gcs"
	jwtExt "github.com/paper-indonesia/pivot-backoffice/pkg/jwt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"go.opentelemetry.io/otel"
)

var (
	otelTracer        = otel.Tracer("UnifiedPaymentService")
	asiaJakartaLoc, _ = time.LoadLocation(constant.TimeLoc)
)

const (
	chargeExportCacheExpiration = 15 * time.Minute
	maxPaymentCreatedDays       = 14
	workerCount                 = 5
)

type UnifiedPaymentService struct {
	config                  *config.Config
	secret                  *config.Secret
	logger                  logger.ILogger
	paymentRepo             repository.IPaymentRepository
	paymentMethodRepo       repository.IPaymentMethodRepository
	accountTransactionRepo  repository.IAccountTransactionRepository
	customerRepo            repository.ICustomerRepository
	creditCardProcessorRepo repository.ICreditcardCoreProcessorRepository
	statusHistoriesRepo     repository.IStatusHistoriesRepository
	paymentCaptureRepo      repository.IPaymentCaptureRepository

	merchantRepo repository.IMerchantRepository
	snapCoreRepo repository.ISnapCoreRepository

	feeSvc                    services.IFeeService
	orchestratorSvc           services.IOrchestratorService
	qrisSvc                   services.IQrisService
	paymentSvc                services.IPaymentService
	paymentMethodSvc          services.IPaymentMethodService
	creditcardSvc             services.ICreditCardService
	fdsSvc                    services.IFdsService
	installmentPlanSvc        services.IInstallmentPlanService
	merchantSvc               services.IMerchantService
	internalUnifiedPaymentSvc services.IInternalUnifiedPaymentService
	shortLinkSvc              services.IShortLinkService
	customerSvc               services.ICustomerService
	recurringContractRepo     repository.IRecurringContractRepository
	recurringContractSvc      services.IRecurringContractService
	cardFundedPayoutSvc       services.ICardFundedPayoutService

	rabbitMqExt      rabbitMqExt.IRabbitMQExt
	redis            redisExt.IRedisExt
	jwt              jwtExt.IJwt
	cache            redisExt.IRedisExt
	storage          gcs.IGCSService
	fdsVelocityCheck fds.VelocityChecker
	cryptoProvider   encryption.CryptoProvider
}

type UnifiedPaymentServiceFunc func(*UnifiedPaymentService)

const serviceName = "Unified-Payment-Service"

func New(
	config *config.Config,
	logger logger.ILogger,
	paymentRepo repository.IPaymentRepository,
	paymentMethodRepo repository.IPaymentMethodRepository,
	accountTransactionRepo repository.IAccountTransactionRepository,
	depends ...UnifiedPaymentServiceFunc,
) services.IUnifiedPaymentService {
	s := &UnifiedPaymentService{
		config:                 config,
		logger:                 logger,
		paymentRepo:            paymentRepo,
		paymentMethodRepo:      paymentMethodRepo,
		accountTransactionRepo: accountTransactionRepo,
	}

	for _, fn := range depends {
		fn(s)
	}

	s.internalUnifiedPaymentSvc = s

	return s
}

func WithMerchantRepo(repo repository.IMerchantRepository) UnifiedPaymentServiceFunc {
	return func(ds *UnifiedPaymentService) {
		ds.merchantRepo = repo
	}
}

func WithSnapCoreRepo(repo repository.ISnapCoreRepository) UnifiedPaymentServiceFunc {
	return func(ds *UnifiedPaymentService) {
		ds.snapCoreRepo = repo
	}
}

func WithJWTExt(jwt jwtExt.IJwt) UnifiedPaymentServiceFunc {
	return func(ds *UnifiedPaymentService) {
		ds.jwt = jwt
	}
}

func WithRabbitMQClient(rmq rabbitMqExt.IRabbitMQExt) UnifiedPaymentServiceFunc {
	return func(ds *UnifiedPaymentService) {
		ds.rabbitMqExt = rmq
	}
}

func WithRedisClient(rdb redisExt.IRedisExt) UnifiedPaymentServiceFunc {
	return func(ps *UnifiedPaymentService) {
		ps.redis = rdb
	}
}

func WithFeeService(svc services.IFeeService) UnifiedPaymentServiceFunc {
	return func(ds *UnifiedPaymentService) {
		ds.feeSvc = svc
	}
}

func WithOrchestratorService(svc services.IOrchestratorService) UnifiedPaymentServiceFunc {
	return func(ds *UnifiedPaymentService) {
		ds.orchestratorSvc = svc
	}
}

func WithQRISService(svc services.IQrisService) UnifiedPaymentServiceFunc {
	return func(ds *UnifiedPaymentService) {
		ds.qrisSvc = svc
	}
}

func WithPaymentService(svc services.IPaymentService) UnifiedPaymentServiceFunc {
	return func(ds *UnifiedPaymentService) {
		ds.paymentSvc = svc
	}
}

func WithCustomerRepo(repo repository.ICustomerRepository) UnifiedPaymentServiceFunc {
	return func(ds *UnifiedPaymentService) {
		ds.customerRepo = repo
	}
}

func WithPaymentMethodService(svc services.IPaymentMethodService) UnifiedPaymentServiceFunc {
	return func(ds *UnifiedPaymentService) {
		ds.paymentMethodSvc = svc
	}
}

func WithCreditCardService(svc services.ICreditCardService) UnifiedPaymentServiceFunc {
	return func(ds *UnifiedPaymentService) {
		ds.creditcardSvc = svc
	}
}

func WithFdsService(svc services.IFdsService) UnifiedPaymentServiceFunc {
	return func(ds *UnifiedPaymentService) {
		ds.fdsSvc = svc
	}
}

func WithCreditCardCoreProcessorRepo(repo repository.ICreditcardCoreProcessorRepository) UnifiedPaymentServiceFunc {
	return func(ups *UnifiedPaymentService) {
		ups.creditCardProcessorRepo = repo
	}
}

func WithSecret(secret *config.Secret) UnifiedPaymentServiceFunc {
	return func(ds *UnifiedPaymentService) {
		ds.secret = secret
	}
}

func WithCache(cache redisExt.IRedisExt) UnifiedPaymentServiceFunc {
	return func(ds *UnifiedPaymentService) {
		ds.cache = cache
	}
}

func WithStorage(storage gcs.IGCSService) UnifiedPaymentServiceFunc {
	return func(ds *UnifiedPaymentService) {
		ds.storage = storage
	}
}

func WithStatusHistoriesRepository(repo repository.IStatusHistoriesRepository) UnifiedPaymentServiceFunc {
	return func(ds *UnifiedPaymentService) {
		ds.statusHistoriesRepo = repo
	}
}

func WithPaymentCaptureRepository(repo repository.IPaymentCaptureRepository) UnifiedPaymentServiceFunc {
	return func(ds *UnifiedPaymentService) {
		ds.paymentCaptureRepo = repo
	}
}

func WithRecurringContractRepository(repo repository.IRecurringContractRepository) UnifiedPaymentServiceFunc {
	return func(ds *UnifiedPaymentService) {
		ds.recurringContractRepo = repo
	}
}

func WithRecurringContractService(svc services.IRecurringContractService) UnifiedPaymentServiceFunc {
	return func(ds *UnifiedPaymentService) {
		ds.recurringContractSvc = svc
	}
}

func WithFDSVelocityCheck(velocity fds.VelocityChecker) UnifiedPaymentServiceFunc {
	return func(ds *UnifiedPaymentService) {
		ds.fdsVelocityCheck = velocity
	}
}

func WithCryptoProvider(cryptoProvider encryption.CryptoProvider) UnifiedPaymentServiceFunc {
	return func(ds *UnifiedPaymentService) {
		ds.cryptoProvider = cryptoProvider
	}
}

func WithInstallmentPlanService(svc services.IUnifiedPaymentService, installmentSvc services.IInstallmentPlanService) {
	svc.(*UnifiedPaymentService).installmentPlanSvc = installmentSvc
}

func WithMerchantService(svc services.IUnifiedPaymentService, merchantSvc services.IMerchantService) {
	svc.(*UnifiedPaymentService).merchantSvc = merchantSvc
}

func WithInternalUnifiedPaymentService(svc services.IUnifiedPaymentService, internalUnifiedPaymentSvc services.IInternalUnifiedPaymentService) {
	svc.(*UnifiedPaymentService).internalUnifiedPaymentSvc = internalUnifiedPaymentSvc
}

func WithShortLinkService(svc services.IUnifiedPaymentService, shortLinkSvc services.IShortLinkService) {
	svc.(*UnifiedPaymentService).shortLinkSvc = shortLinkSvc
}

func WithCustomerService(svc services.IUnifiedPaymentService, customerSvc services.ICustomerService) {
	svc.(*UnifiedPaymentService).customerSvc = customerSvc
}

func WithCardFundedPayoutService(svc services.IUnifiedPaymentService, service services.ICardFundedPayoutService) {
	svc.(*UnifiedPaymentService).cardFundedPayoutSvc = service
}

func (s *UnifiedPaymentService) isPaymentMigrationV1ToV2Enabled(ctx context.Context, merchantId string) bool {
	_, segment := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/isPaymentMigrationV1ToV2Enabled")
	defer segment.End()

	attr := ffcontext.NewEvaluationContext(s.config.Environment)
	attr.AddCustomAttribute(constant.FeatureFlagTargetQueryNameEnv, s.config.Environment)
	attr.AddCustomAttribute(constant.FeatureFlagTargetQueryNameMerchantId, merchantId)

	enabled, _ := ffclient.BoolVariation(constant.FeatureFlagPaymentMigrationV1toV2Enabled, attr, false)
	return enabled
}
func (s *UnifiedPaymentService) isMerchantExcludedToSendCaptureHistoryOnCallback(ctx context.Context, merchantId string) bool {
	_, segment := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/isMerchantExcludedToSendCaptureHistoryOnCallback")
	defer segment.End()

	attr := ffcontext.NewEvaluationContext(s.config.Environment)
	attr.AddCustomAttribute(constant.FeatureFlagTargetQueryNameEnv, s.config.Environment)
	attr.AddCustomAttribute(constant.FeatureFlagTargetQueryNameMerchantId, merchantId)

	enabled, _ := ffclient.BoolVariation(constant.FeatureFlagMerchantExcludedSendCaptureHistory, attr, false)
	return enabled
}
func (s *UnifiedPaymentService) isEWalletPaymentSimulationFlowEnabled(ctx context.Context, merchantId string) bool {
	_, segment := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/isEWalletPaymentSimulationFlowEnabled")
	defer segment.End()

	attr := ffcontext.NewEvaluationContext(merchantId)
	attr.AddCustomAttribute(constant.FeatureFlagTargetQueryNameMerchantId, merchantId)

	enabled, _ := ffclient.BoolVariation(constant.FeatureFlagEwalletPaymentSimulationFlow, attr, true)
	return enabled
}

// RecordChargeStatusHistory records charge status history synchronously
func (s *UnifiedPaymentService) RecordChargeStatusHistory(ctx context.Context, chargeID, actor, statusType string) {
	if s.statusHistoriesRepo == nil {
		return
	}

	switch statusType {
	case constant.ChargeStatusHistoryWaitingForUserAction:
		s.recordChargeWaitingForUserAction(ctx, chargeID, actor)
	case constant.ChargeStatusHistoryWaitingForAuthentication:
		s.recordChargeWaitingForAuthentication(ctx, chargeID, actor)
	case constant.ChargeStatusHistoryWaitingForExternalFDS:
		s.recordChargeWaitingForExternalFDS(ctx, chargeID, actor)
	case constant.ChargeStatusHistoryProcessing:
		s.recordChargeProcessing(ctx, chargeID, actor)
	case constant.ChargeStatusHistoryWaitingForCapture:
		s.recordChargeWaitingForCapture(ctx, chargeID, actor)
	case constant.ChargeStatusHistorySuccess:
		s.recordChargeSuccess(ctx, chargeID, actor)
	case constant.ChargeStatusHistoryFailed:
		s.recordChargeFailed(ctx, chargeID, actor)
	case constant.ChargeStatusHistoryExpired:
		s.recordChargeExpired(ctx, chargeID, actor)
	}
}

func (s *UnifiedPaymentService) getPaymentAutoInquiryConfig(ctx context.Context) *constant.PaymentAutoInquiryConfig {
	_, segment := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/getPaymentAutoInquiryConfig")
	defer segment.End()

	cooldownSeconds := 30
	if s.config.UnifiedPaymentConfig.AutoInquiryConfig != nil &&
		s.config.UnifiedPaymentConfig.AutoInquiryConfig.CooldownSeconds > 0 {
		cooldownSeconds = s.config.UnifiedPaymentConfig.AutoInquiryConfig.CooldownSeconds
	}

	if ffCooldown := constant.GetPaymentAutoInquiryCooldownSeconds(s.config.Environment); ffCooldown != nil {
		cooldownSeconds = *ffCooldown
	}

	return &constant.PaymentAutoInquiryConfig{
		CooldownSeconds: cooldownSeconds,
		EnabledMethods:  []string{paymentConstant.PAYMENT_METHOD_QRIS, paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT},
	}
}
