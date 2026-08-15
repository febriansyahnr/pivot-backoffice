package disbursementService

import (
	"context"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/disbursement"
	repository "github.com/paper-indonesia/pivot-backoffice/internal/repository/backendportal"
	service "github.com/paper-indonesia/pivot-backoffice/internal/service/backendportal"
	"github.com/paper-indonesia/pivot-backoffice/pkg/gcs"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/xlsx"
	"go.opentelemetry.io/otel"
)

var (
	otelTracer = otel.Tracer("DisbursementService")
	local, _   = time.LoadLocation(constant.TimeLoc)
)

type DisbursementService struct {
	config *config.Config
	logger logger.ILogger

	// repository
	merchantRepo                      repository.IMerchantRepository
	disbursementRepo                  repository.IDisbursementRepository
	snapCoreRepo                      repository.ISnapCoreRepository
	accountRepo                       repository.IAccountRepository
	accountTransactionRepo            repository.IAccountTransactionRepository
	bankAccountRepo                   repository.IBankAccountRepository
	statusHistoriesRepo               repository.IStatusHistoriesRepository
	workflowFDSRepo                   repository.IWorkflowFDSRepository
	payoutManualProcessingAccountRepo repository.IPayoutManualProcessingAccountRepository

	// service
	orchestratorSvc             service.IOrchestratorService
	beneficiaryAccountSvc       service.IBeneficiaryAccountService
	merchantForbiddenUsecaseSvc service.IMerchantForbiddenUseCaseService
	feeSvc                      service.IFeeService
	routingProcessorSvc         service.IRoutingProcessorService
	transferSvc                 service.ITransferService
	ledgerSvc                   service.ILedgerService
	merchantSvc                 service.IMerchantService

	// ext
	rabbitMqExt rabbitMqExt.IRabbitMQExt
	gcs         gcs.IGCSService
	redisExt    redisExt.IRedisExt
	excel       xlsx.Exceler

	// Workerpool
	totalWorkerPool int
	batchProcessWP  *ants.PoolWithFunc
	batchCreateWP   *ants.PoolWithFunc
	batchApprovalWP *ants.PoolWithFunc
	statusHistoryWP *ants.PoolWithFunc
}

type DisbursementServiceFunc func(*DisbursementService)

func New(
	config *config.Config,
	logger logger.ILogger,
	merchantRepo repository.IMerchantRepository,
	disbursementRepo repository.IDisbursementRepository,
	snapCoreRepo repository.ISnapCoreRepository,
	bankAccountRepo repository.IBankAccountRepository,
	depends ...DisbursementServiceFunc,
) service.IDisbursementService {
	d := &DisbursementService{
		config:           config,
		logger:           logger,
		merchantRepo:     merchantRepo,
		disbursementRepo: disbursementRepo,
		snapCoreRepo:     snapCoreRepo,
		bankAccountRepo:  bankAccountRepo,
		totalWorkerPool:  config.WorkerPoolConfig.Disbursement, // Default
	}
	for _, fn := range depends {
		fn(d)
	}

	return d
}

func WithRedisClient(rdb redisExt.IRedisExt) DisbursementServiceFunc {
	return func(ds *DisbursementService) {
		ds.redisExt = rdb
	}
}

func WithGoogleCloudStorage(gcs gcs.IGCSService) DisbursementServiceFunc {
	return func(ds *DisbursementService) {
		ds.gcs = gcs
	}
}

func WithRabbitMQClient(rmq rabbitMqExt.IRabbitMQExt) DisbursementServiceFunc {
	return func(ds *DisbursementService) {
		ds.rabbitMqExt = rmq
	}
}

func WithOrchestratorService(svc service.IOrchestratorService) DisbursementServiceFunc {
	return func(ds *DisbursementService) {
		ds.orchestratorSvc = svc
	}
}

func WithBeneficiaryAccService(svc service.IBeneficiaryAccountService) DisbursementServiceFunc {
	return func(ds *DisbursementService) {
		ds.beneficiaryAccountSvc = svc
	}
}

func WithMerchantForbiddenUseCaseService(svc service.IMerchantForbiddenUseCaseService) DisbursementServiceFunc {
	return func(ds *DisbursementService) {
		ds.merchantForbiddenUsecaseSvc = svc
	}
}

func WithFeeService(svc service.IFeeService) DisbursementServiceFunc {
	return func(ds *DisbursementService) {
		ds.feeSvc = svc
	}
}

func WithRoutingProcessorService(svc service.IRoutingProcessorService) DisbursementServiceFunc {
	return func(ds *DisbursementService) {
		ds.routingProcessorSvc = svc
	}
}

func WithAccountTransactionRepository(repo repository.IAccountTransactionRepository) DisbursementServiceFunc {
	return func(ds *DisbursementService) {
		ds.accountTransactionRepo = repo
	}
}

func WithStatusHistoriesRepository(repo repository.IStatusHistoriesRepository) DisbursementServiceFunc {
	return func(ds *DisbursementService) {
		ds.statusHistoriesRepo = repo
	}
}

func WithExcelLibrary(excel xlsx.Exceler) DisbursementServiceFunc {
	return func(ds *DisbursementService) {
		ds.excel = excel
	}
}

func WithDisbursementWorkerPool(total int) DisbursementServiceFunc {
	return func(ds *DisbursementService) {
		ds.totalWorkerPool = total
	}
}

func WithTransferService(service service.ITransferService) DisbursementServiceFunc {
	return func(ds *DisbursementService) {
		ds.transferSvc = service
	}
}

func WithWorkflowFDSRepository(repo repository.IWorkflowFDSRepository) DisbursementServiceFunc {
	return func(ds *DisbursementService) {
		ds.workflowFDSRepo = repo
	}
}

func WithPayoutManualProcessingAccountRepository(repo repository.IPayoutManualProcessingAccountRepository) DisbursementServiceFunc {
	return func(ds *DisbursementService) {
		ds.payoutManualProcessingAccountRepo = repo
	}
}

func (s *DisbursementService) WPRelease() {
	if s.batchProcessWP != nil {
		s.batchProcessWP.Release()
	}

	if s.batchCreateWP != nil {
		s.batchCreateWP.Release()
	}

	if s.batchApprovalWP != nil {
		s.batchApprovalWP.Release()
	}

	if s.statusHistoryWP != nil {
		s.statusHistoryWP.Release()
	}
}

type batchApprovalWPData struct {
	ctx                     context.Context
	wg                      *sync.WaitGroup
	disbursementId          string
	cutOffTimeStatusOngoing bool
	cutOffTimeProcessedAt   time.Time
	mx                      *sync.Mutex
	approvalValidationResp  *[]disbursementModel.ApprovalValidation
}

type statusHistoryWPData struct {
	ctx            context.Context
	wg             *sync.WaitGroup
	disbursementId string
	actor          string
	statusType     string // "approved" or "waiting_for_topup"
}

func WithLedgerService(service service.ILedgerService) DisbursementServiceFunc {
	return func(ds *DisbursementService) {
		ds.ledgerSvc = service
	}
}

func WithMerchantService(service service.IMerchantService) DisbursementServiceFunc {
	return func(ds *DisbursementService) {
		ds.merchantSvc = service
	}
}

func WithAccountRepository(repo repository.IAccountRepository) DisbursementServiceFunc {
	return func(ds *DisbursementService) {
		ds.accountRepo = repo
	}
}
