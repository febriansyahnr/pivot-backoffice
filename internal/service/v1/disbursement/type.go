package disbursementService

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/panjf2000/ants/v2"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	orchestraModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository/datamart"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/gcs"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/pkg/xlsx"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
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
	disbursementMetricsRepo           datamart.IDatamartDisbursementMetrics
	workflowFDSRepo                   repository.IWorkflowFDSRepository
	payoutManualProcessingAccountRepo repository.IPayoutManualProcessingAccountRepository

	// service
	orchestratorSvc             service.IOrchestratorService
	beneficiaryAccountSvc       service.IBeneficiaryAccountService
	merchantForbiddenUsecaseSvc service.IMerchantForbiddenUseCaseService
	feeSvc                      service.IFeeService
	routingProcessorSvc         service.IRoutingProcessorService
	transferSvc                 service.ITransferService
	self                        service.IDisbursementInternalService
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
	d.self = d
	d.newBatchApprovalWP()
	d.newStatusHistoryWP()

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

type batchProcessWPData struct {
	ctx            context.Context
	wg             *sync.WaitGroup
	disbursementID string
}

func (s *DisbursementService) batchProcessWPFunc(data interface{}) {
	req := data.(batchProcessWPData)
	defer req.wg.Done()

	var (
		queueTTLLock = time.Duration(s.config.AppConfig.DisbursementProcessExpireLockSecond) * time.Second
		traceId, _   = req.ctx.Value(pdkConst.CtxTraceIdKey).(string)
	)

	queueKey := fmt.Sprintf(constant.DisbursementProcessQueueLockFmt, req.disbursementID)
	if ok, errLock := s.redisExt.SetNX(req.ctx, queueKey, true, queueTTLLock).Result(); errLock != nil {
		s.logger.Error(req.ctx, "[create-bank-transfer] set exclusive queue with key "+queueKey, logger.Error(pkgErrors.New(httpResponse.HttpErrDatabase, fmt.Errorf("QUEUE: "+constant.InternalErrorFmt, traceId))))
		return

	} else if !ok {
		s.logger.Error(req.ctx, "[create-bank-transfer] duplicate check", logger.Error(pkgErrors.New(httpResponse.HttpErrDupCheck, constant.ErrDisbursementIsBeingProcessed)))
		return
	}

	defer func() {
		if e := s.redisExt.Del(req.ctx, queueKey).Err(); e != nil {
			s.logger.Error(req.ctx, "[create-bank-transfer] clears the disbursement process queue lock", logger.Error(e))
		}
	}()

	disbursement, err := s.disbursementRepo.FindByID(req.ctx, req.disbursementID)
	if err != nil {
		s.logger.Error(req.ctx, "[create-bank-transfer] error when find disbursement by id: "+req.disbursementID, logger.Error(err))
		return
	}

	if err = s.self.CreateBankTransfer(req.ctx, disbursement); err != nil {
		s.logger.Error(req.ctx, fmt.Sprintf("[create-bank-transfer] error processing batch process disbursement %s", req.disbursementID), logger.Error(err), logger.Any("disbursement", disbursement))
	}
}

func (s *DisbursementService) newBatchProcessWP() {
	var opts []ants.Option
	s.batchProcessWP, _ = ants.NewPoolWithFunc(s.totalWorkerPool, s.batchProcessWPFunc, opts...)
}

type batchCreateWPData struct {
	ctx           context.Context
	wg            *sync.WaitGroup
	bulkID        string
	createRequest disbursementModel.CreateSingleRequest
}

func (s *DisbursementService) batchCreateWPFunc(data interface{}) {
	wpData := data.(batchCreateWPData)
	defer wpData.wg.Done()

	ctx := wpData.ctx

	_, err := s.CreateSingle(ctx, &wpData.createRequest)
	if err != nil {
		s.logger.Error(ctx, fmt.Sprintf("error creating single disbursement from %s", wpData.bulkID), logger.Error(err))
	}

	queueKey := fmt.Sprintf(
		constant.BulkDisbursementQueueLockFmt, wpData.createRequest.MerchantID, wpData.createRequest.ReferenceID,
	)
	if err := s.redisExt.Del(ctx, queueKey).Err(); err != nil {
		s.logger.Error(ctx, "delete exclusive queue with key "+queueKey, logger.Error(err))
	}
}

func (s *DisbursementService) newBatchCreateWP() {
	s.batchCreateWP, _ = ants.NewPoolWithFunc(s.totalWorkerPool, s.batchCreateWPFunc)
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

func (s *DisbursementService) newBatchApprovalWP() {
	s.batchApprovalWP, _ = ants.NewPoolWithFunc(s.totalWorkerPool, func(data interface{}) {
		approval := data.(*batchApprovalWPData)
		defer approval.wg.Done()

		request := disbursementModel.DisbursementWithTransaction{} // Must be empty

		if disbursement, _ := s.disbursementRepo.FindByID(approval.ctx, approval.disbursementId); disbursement != nil {
			request = *disbursement // Handle race condition when run unit test with flag -race
		}
		request.CutOffTimeStatusOngoing = approval.cutOffTimeStatusOngoing

		if request.CutOffTimeStatusOngoing {
			batchTotal, _ := s.redisExt.HIncrByFloat(approval.ctx, constant.DelayTransferProcessRedisKey, "total", 1)
			if batchTotal == 1 {
				_ = s.redisExt.Expire(
					approval.ctx, constant.DelayTransferProcessRedisKey, approval.cutOffTimeProcessedAt.Add((3 * time.Minute)).Sub(time.Now().UTC()),
				)
			}

			amount := request.Amount.InexactFloat64()
			_, _ = s.redisExt.HIncrByFloat(approval.ctx, constant.DelayTransferProcessRedisKey, "amount", amount)
			_, _ = s.redisExt.HIncrByFloat(approval.ctx, constant.DelayTransferProcessRedisKey, fmt.Sprintf("bank_%s_total", request.BeneficiaryBankCode), 1)
			_, _ = s.redisExt.HIncrByFloat(approval.ctx, constant.DelayTransferProcessRedisKey, fmt.Sprintf("bank_%s_amount", request.BeneficiaryBankCode), amount)
		}

		mutex := s.redisExt.NewMutex(
			fmt.Sprintf(constant.DisbursementApprovalBeneficiaryLockFmt, request.BeneficiaryBankCode, request.BeneficiaryAccountNo),
			redsync.WithExpiry(5*time.Second),
			redsync.WithRetryDelay(80*time.Millisecond),
			redsync.WithFailFast(true),
			redsync.WithTries(256),
		)
		if err := mutex.LockContext(approval.ctx); err != nil {
			s.logger.Warn(approval.ctx, "Failed lock distributed lock", logger.Error(err))
		}
		defer func() {
			if _, err := mutex.UnlockContext(approval.ctx); err != nil {
				s.logger.Warn(approval.ctx, "Failed unlock distributed lock", logger.Error(err))
			}
		}()

		var (
			err                             error
			transactionId, transactionFeeId = "", ""
		)
		if transactionId, transactionFeeId, err = s.self.CreatePendingOrchestratorTransaction(approval.ctx, &request); err != nil {
			s.logger.Error(approval.ctx, "error when create pending orchestrator transaction", logger.Any("details", request))
		}

		// Validate beneficiary payout limit
		if errBeneficiaryLimit := s.self.ValidateBankAccountAndUpdateTransaction(approval.ctx, &request, &orchestraModel.TransactionAndFeeObject{
			TransactionID: transactionId,
			FeeID:         transactionFeeId,
			MerchantID:    request.MerchantID,
		}); errBeneficiaryLimit != nil {
			s.logger.Error(approval.ctx, "validate bank account", logger.Any("details", map[string]any{
				"transactionId": transactionId,
				"feeId":         transactionFeeId,
				"merchantId":    request.MerchantID,
				"bankCode":      request.BeneficiaryBankCode,
				"accountNo":     request.BeneficiaryAccountNo,
			}))

			// catch errBeneficiaryLimit
			approval.mx.Lock()
			*approval.approvalValidationResp = append(*approval.approvalValidationResp, disbursementModel.ApprovalValidation{
				Amount:    request.Amount.InexactFloat64(),
				AccountNo: request.BeneficiaryAccountNo,
				Error:     errBeneficiaryLimit,
			})
			approval.mx.Unlock()
		}
	})
}

func (s *DisbursementService) newStatusHistoryWP() {
	s.statusHistoryWP, _ = ants.NewPoolWithFunc(s.totalWorkerPool, func(data interface{}) {
		statusHistory := data.(*statusHistoryWPData)
		defer statusHistory.wg.Done()

		switch statusHistory.statusType {
		case constant.DisbursementStatusHistoryApproved:
			s.recordDisbursementApproved(statusHistory.ctx, statusHistory.disbursementId, statusHistory.actor)
		case constant.DisbursementStatusHistoryWaitingForTopUp:
			s.recordDisbursementWaitingForTopUp(statusHistory.ctx, statusHistory.disbursementId, statusHistory.actor)
		}
	})
}

func WithDisbursementInternalService(fn service.IDisbursementInternalService) DisbursementServiceFunc {
	return func(ds *DisbursementService) {
		ds.self = fn
	}
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

func WithDisbursementMetricsRepository(repo datamart.IDatamartDisbursementMetrics) DisbursementServiceFunc {
	return func(ds *DisbursementService) {
		ds.disbursementMetricsRepo = repo
	}
}
