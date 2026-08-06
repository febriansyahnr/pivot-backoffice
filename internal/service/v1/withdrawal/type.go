package withdrawalService

import (
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/gcs"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pdk/v2/logger"

	"go.opentelemetry.io/otel"
)

type withdrawalService struct {
	logger             logger.ILogger
	repo               repository.IWithdrawalRepository
	orchestratorSvc    service.IOrchestratorService
	userSvc            service.IUserService
	bankAccountRepo    repository.IBankAccountRepository
	snapCoreRepo       repository.ISnapCoreRepository
	redis              redisExt.IRedisExt
	accountRepo        repository.IAccountRepository
	config             *config.WithdrawalConfig
	merchantRepo       repository.IMerchantRepository
	rmq                rabbitMqExt.IRabbitMQExt
	gcs                gcs.IGCSService
	accountTrxRepo     repository.IAccountTransactionRepository
	internal           service.IWithdrawalExporterService
	bankTransferConfig service.IBankTransferConfig
	notificationSvc    service.INotificationService
}

type Option func(*withdrawalService)

var (
	loc, _     = time.LoadLocation(constant.TimeLoc)
	otelTracer = otel.Tracer("WithdrawalService")
)

var downloadWithdrawalHeaders = []string{
	"Created Date", "Last Update", "Created By", "Amount", "Status", "Transaction ID", "Bank Reference", "Bank Name", "Account Number", "Beneficiary Name",
}

func New(
	log logger.ILogger,
	repo repository.IWithdrawalRepository,
	orchestratorSvc service.IOrchestratorService,
	bankAccountRepo repository.IBankAccountRepository,
	userSvc service.IUserService,
	options ...Option,
) service.IWithdrawalService {
	service := &withdrawalService{
		logger:          log,
		repo:            repo,
		orchestratorSvc: orchestratorSvc,
		bankAccountRepo: bankAccountRepo,
		userSvc:         userSvc,
	}
	service.internal = service

	for _, opt := range options {
		opt(service)
	}
	return service
}

func WithRedisClient(rdb redisExt.IRedisExt) Option {
	return func(ws *withdrawalService) {
		ws.redis = rdb
	}
}

func WithSnapCoreRepository(snap repository.ISnapCoreRepository) Option {
	return func(ws *withdrawalService) {
		ws.snapCoreRepo = snap
	}
}

func WithAccountRepository(accountRepo repository.IAccountRepository) Option {
	return func(ws *withdrawalService) {
		ws.accountRepo = accountRepo
	}
}

func WithWithdrawalConfig(cfg *config.WithdrawalConfig) Option {
	return func(ws *withdrawalService) {
		ws.config = cfg

		if ws.config.AutoWithdrawalWorker <= 0 {
			ws.config.AutoWithdrawalWorker = 10
		}
	}
}

func WithMerchantRepository(merchantRepo repository.IMerchantRepository) Option {
	return func(ws *withdrawalService) {
		ws.merchantRepo = merchantRepo
	}
}

func WithRabbitMQClient(rmq rabbitMqExt.IRabbitMQExt) Option {
	return func(ws *withdrawalService) {
		ws.rmq = rmq
	}
}

func WithGCSService(gcs gcs.IGCSService) Option {
	return func(ws *withdrawalService) {
		ws.gcs = gcs
	}
}

func WithWithdrawalExporter(exporter service.IWithdrawalExporterService) Option {
	return func(ws *withdrawalService) {
		ws.internal = exporter
	}
}

func WithAccountTransactionRepository(repo repository.IAccountTransactionRepository) Option {
	return func(ws *withdrawalService) {
		ws.accountTrxRepo = repo
	}
}

func WithBankTransferConfig(config service.IBankTransferConfig) Option {
	return func(ws *withdrawalService) {
		ws.bankTransferConfig = config
	}
}

func WithNotificationService(notificationSvc service.INotificationService) Option {
	return func(ws *withdrawalService) {
		ws.notificationSvc = notificationSvc
	}
}

// A private data type used only in this package to simplify function return values.
type withdrawalCreateResult struct {
	id            string
	transactionId string
	createdAt     time.Time
	updatedAt     time.Time
	metadata      withdrawal.Metadata
}
