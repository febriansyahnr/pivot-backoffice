package cardFundedPayoutService

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	port "github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/encryption"
	"github.com/paper-indonesia/pivot-backoffice/pkg/gcs"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
	"golang.org/x/sync/singleflight"
)

var (
	otelTracer = otel.Tracer("CardFundedPayoutService")
)

type service struct {
	config *config.Config
	logger logger.ILogger
	// service
	customerSvc       port.ICustomerService
	unifiedPaymentSvc port.IUnifiedPaymentService
	vendorSvc         port.IVendorService
	feeSvc            port.IFeeService
	creditCardSvc     port.ICreditCardService
	orchestratorSvc   port.IOrchestratorService
	// Repository
	disbursementRepo       repository.IDisbursementRepository
	statusHistoryRepo      repository.IStatusHistoriesRepository
	paymentMethodRepo      repository.IPaymentMethodRepository
	paymentRepo            repository.IPaymentRepository
	accountTransactionRepo repository.IAccountTransactionRepository
	snapCoreRepo           repository.ISnapCoreRepository
	// Others
	cacheClient    redisExt.IRedisExt
	cryptoProvider encryption.CryptoProvider
	gcs            gcs.IGCSService
	receiptSf      singleflight.Group
}

type Option func(*service)

func New(
	config *config.Config,
	logger logger.ILogger,
	depends ...Option,
) port.ICardFundedPayoutService {
	d := &service{
		config: config,
		logger: logger,
	}

	for _, fn := range depends {
		fn(d)
	}

	return d
}

func WithCustomerService(svc port.ICustomerService) Option {
	return func(s *service) {
		s.customerSvc = svc
	}
}

func WithUnifiedPaymentService(svc port.IUnifiedPaymentService) Option {
	return func(s *service) {
		s.unifiedPaymentSvc = svc
	}
}

func WithFeeService(svc port.IFeeService) Option {
	return func(s *service) {
		s.feeSvc = svc
	}
}

func WithVendorService(svc port.IVendorService) Option {
	return func(s *service) {
		s.vendorSvc = svc
	}
}

func WithCreditCardService(svc port.ICreditCardService) Option {
	return func(s *service) {
		s.creditCardSvc = svc
	}
}

func WithOrchestratorService(svc port.IOrchestratorService) Option {
	return func(s *service) {
		s.orchestratorSvc = svc
	}
}

func WithDisbursementRepository(repo repository.IDisbursementRepository) Option {
	return func(s *service) {
		s.disbursementRepo = repo
	}
}

func WithStatusHistoriesRepository(repo repository.IStatusHistoriesRepository) Option {
	return func(s *service) {
		s.statusHistoryRepo = repo
	}
}

func WithPaymentMethodRepository(repo repository.IPaymentMethodRepository) Option {
	return func(s *service) {
		s.paymentMethodRepo = repo
	}
}

func WithPaymentRepository(repo repository.IPaymentRepository) Option {
	return func(s *service) {
		s.paymentRepo = repo
	}
}

func WithAccountTransactionRepository(repo repository.IAccountTransactionRepository) Option {
	return func(s *service) {
		s.accountTransactionRepo = repo
	}
}

func WithSnapCoreRepository(repo repository.ISnapCoreRepository) Option {
	return func(s *service) {
		s.snapCoreRepo = repo
	}
}

func WithCacheClient(client redisExt.IRedisExt) Option {
	return func(s *service) {
		s.cacheClient = client
	}
}

func WithCryptoProvider(provider encryption.CryptoProvider) Option {
	return func(s *service) {
		s.cryptoProvider = provider
	}
}

func WithGCS(gcsService gcs.IGCSService) Option {
	return func(s *service) {
		s.gcs = gcsService
	}
}
