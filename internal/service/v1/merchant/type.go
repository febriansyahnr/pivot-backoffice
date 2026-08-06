package merchant

import (
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/encryption"
	"github.com/paper-indonesia/pivot-backoffice/pkg/gcs"
	"github.com/paper-indonesia/pivot-backoffice/pkg/jwt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/vault"
	"github.com/paper-indonesia/pivot-backoffice/pkg/xlsx"

	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("MerchantService")

const merchantDocumentExp = 5 * time.Minute

type MerchantService struct {
	merchantTopUpRepo repository.IMerchantTopUpRepository
	repo              repository.IMerchantRepository
	User              repository.IUserRepository
	accountRepo       repository.IAccountRepository
	accountService    service.IAccountService
	locationRepo      repository.IAddrLocationRepository
	UserSvc           service.IUserService
	feeSvc            service.IFeeCalculator
	orchestratorSvc   service.IOrchestratorService
	qrisSvc           service.IQrisService
	industrySvc       service.IIndustryService
	countrySvc        service.ICountryService
	logger            logger.ILogger
	JWT               jwt.IJwt
	rabbitMqExt       rabbitMqExt.IRabbitMQExt
	cryptoExt         encryption.ICrypto
	gcs               gcs.IGCSService
	config            *config.Config
	redis             redisExt.IRedisExt
	validator         *validatorExt.Validate

	// optional
	activityRepo           repository.IActivityRepository
	accountTransactionRepo repository.IAccountTransactionRepository
	bankAccountRepo        repository.IBankAccountRepository
	productRepo            repository.IProductRepository
	beneficiaryRepo        repository.IBeneficiaryAccountRepository
	beneficiaryAccountSvc  service.IBeneficiaryAccountService
	paymentMethod          repository.IPaymentMethodRepository
	excel                  xlsx.Exceler
	encryption             vault.IVaultTransit
	snapCoreRepo           repository.ISnapCoreRepository
}

type OptionFunc func(*MerchantService)

func New(
	repo repository.IMerchantRepository,
	logger logger.ILogger,
	User repository.IUserRepository,
	JWT jwt.IJwt,
	rabbitMqExt rabbitMqExt.IRabbitMQExt,
	encrypt encryption.ICrypto,
	opts ...OptionFunc,
) service.IMerchantService {
	svc := &MerchantService{
		repo:        repo,
		logger:      logger,
		User:        User,
		JWT:         JWT,
		rabbitMqExt: rabbitMqExt,
		cryptoExt:   encrypt,
	}

	for _, opt := range opts {
		opt(svc)
	}

	return svc
}

func WithAccountRepository(repo repository.IAccountRepository) OptionFunc {
	return func(svc *MerchantService) {
		svc.accountRepo = repo
	}
}

func WithLocationRepository(repo repository.IAddrLocationRepository) OptionFunc {
	return func(svc *MerchantService) {
		svc.locationRepo = repo
	}
}

func WithAccountService(accountSvc service.IAccountService) OptionFunc {
	return func(svc *MerchantService) {
		svc.accountService = accountSvc
	}
}

func WithGCSService(gcs gcs.IGCSService) OptionFunc {
	return func(svc *MerchantService) {
		svc.gcs = gcs
	}
}

func WithServiceConfig(cfg *config.Config) OptionFunc {
	return func(svc *MerchantService) {
		svc.config = cfg
	}
}

func WithRedisClient(rdb redisExt.IRedisExt) OptionFunc {
	return func(svc *MerchantService) {
		svc.redis = rdb
	}
}

func WithUserService(userSvc service.IUserService) OptionFunc {
	return func(svc *MerchantService) {
		svc.UserSvc = userSvc
	}
}

func WithFeeCalculation(fee service.IFeeCalculator) OptionFunc {
	return func(ms *MerchantService) {
		ms.feeSvc = fee
	}
}

func WithActivityRepo(repo repository.IActivityRepository) OptionFunc {
	return func(ms *MerchantService) {
		ms.activityRepo = repo
	}
}

func WithAccountTransactionRepo(repo repository.IAccountTransactionRepository) OptionFunc {
	return func(ms *MerchantService) {
		ms.accountTransactionRepo = repo
	}
}

func WithOrchestratorService(orchestrator service.IOrchestratorService) OptionFunc {
	return func(ms *MerchantService) {
		ms.orchestratorSvc = orchestrator
	}
}

func WithBankAccountRepository(repo repository.IBankAccountRepository) OptionFunc {
	return func(ms *MerchantService) {
		ms.bankAccountRepo = repo
	}
}

func WithProductRepository(repo repository.IProductRepository) OptionFunc {
	return func(ms *MerchantService) {
		ms.productRepo = repo
	}
}

func WithBeneficiaryAccountRepo(repo repository.IBeneficiaryAccountRepository) OptionFunc {
	return func(ms *MerchantService) {
		ms.beneficiaryRepo = repo
	}
}

func WithBeneficiaryAccountService(service service.IBeneficiaryAccountService) OptionFunc {
	return func(ms *MerchantService) {
		ms.beneficiaryAccountSvc = service
	}
}

func WithPaymentMethodRepository(repo repository.IPaymentMethodRepository) OptionFunc {
	return func(ms *MerchantService) {
		ms.paymentMethod = repo
	}
}

func WithQrisService(service service.IQrisService) OptionFunc {
	return func(ms *MerchantService) {
		ms.qrisSvc = service
	}
}

func WithIndustryService(service service.IIndustryService) OptionFunc {
	return func(ms *MerchantService) {
		ms.industrySvc = service
	}
}

func WithCountryService(service service.ICountryService) OptionFunc {
	return func(ms *MerchantService) {
		ms.countrySvc = service
	}
}

func WithValidator(vld *validatorExt.Validate) OptionFunc {
	return func(ms *MerchantService) {
		ms.validator = vld
	}
}

func WithExcelLibrary(lib xlsx.Exceler) OptionFunc {
	return func(ms *MerchantService) {
		ms.excel = lib
	}
}

func WithVaultTransit(transit vault.IVaultTransit) OptionFunc {
	return func(ms *MerchantService) {
		ms.encryption = transit
	}
}

func WithSnapCoreRepo(repo repository.ISnapCoreRepository) OptionFunc {
	return func(ms *MerchantService) {
		ms.snapCoreRepo = repo
	}
}

func WithMerchantTopUpRepo(repo repository.IMerchantTopUpRepository) OptionFunc {
	return func(ms *MerchantService) {
		ms.merchantTopUpRepo = repo
	}
}
