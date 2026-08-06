package refundService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	statusHistoryModel "github.com/paper-indonesia/pivot-backoffice/internal/model/statusHistory"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	services "github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/gcs"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var (
	otelTracer = otel.Tracer("RefundService")
)

type RefundService struct {
	config                 *config.Config
	logger                 logger.ILogger
	refundRepo             repository.IRefundRepository
	paymentRepo            repository.IPaymentRepository
	accountTransactionRepo repository.IAccountTransactionRepository
	merchantRepo           repository.IMerchantRepository
	snapCoreRepo           repository.ISnapCoreRepository
	callbackRepo           repository.ICallbackRepository
	paymentMethodRepo      repository.IPaymentMethodRepository
	statusHistoriesRepo    repository.IStatusHistoriesRepository

	feeSvc          services.IFeeService
	orchestratorSvc services.IOrchestratorService

	rabbitMqExt rabbitMqExt.IRabbitMQExt
	redis       redisExt.IRedisExt
	gcs         gcs.IGCSService
}

type RefundServiceFunc func(service *RefundService)

func New(
	config *config.Config,
	logger logger.ILogger,
	refundRepo repository.IRefundRepository,
	paymentRepo repository.IPaymentRepository,
	accountTransactionRepo repository.IAccountTransactionRepository,
	merchantRepo repository.IMerchantRepository,
	snapCoreRepo repository.ISnapCoreRepository,
	callbackRepo repository.ICallbackRepository,
	depends ...RefundServiceFunc,
) services.IRefundService {
	s := &RefundService{
		config:                 config,
		logger:                 logger,
		refundRepo:             refundRepo,
		paymentRepo:            paymentRepo,
		accountTransactionRepo: accountTransactionRepo,
		merchantRepo:           merchantRepo,
		snapCoreRepo:           snapCoreRepo,
		callbackRepo:           callbackRepo,
	}

	for _, fn := range depends {
		fn(s)
	}

	return s
}

func WithOrchestratorService(svc services.IOrchestratorService) RefundServiceFunc {
	return func(rs *RefundService) {
		rs.orchestratorSvc = svc
	}
}

func WithFeeService(svc services.IFeeService) RefundServiceFunc {
	return func(rs *RefundService) {
		rs.feeSvc = svc
	}
}

func WithRabbitMQClient(rmq rabbitMqExt.IRabbitMQExt) RefundServiceFunc {
	return func(ds *RefundService) {
		ds.rabbitMqExt = rmq
	}
}

func WithRedisClient(rdb redisExt.IRedisExt) RefundServiceFunc {
	return func(ps *RefundService) {
		ps.redis = rdb
	}
}

func WithPaymentMethodRepository(repo repository.IPaymentMethodRepository) RefundServiceFunc {
	return func(rs *RefundService) {
		rs.paymentMethodRepo = repo
	}
}

func WithStatusHistoriesRepository(repo repository.IStatusHistoriesRepository) RefundServiceFunc {
	return func(rs *RefundService) {
		rs.statusHistoriesRepo = repo
	}
}

func WithGCS(gcsService gcs.IGCSService) RefundServiceFunc {
	return func(rs *RefundService) {
		rs.gcs = gcsService
	}
}

// RecordRefundStatusHistory records refund status history synchronously
func (s *RefundService) RecordRefundStatusHistory(ctx context.Context, refundID, actor, statusType string) {
	if s.statusHistoriesRepo == nil {
		return
	}

	switch statusType {
	case constant.RefundStatusHistoryPending:
		s.recordRefundPending(ctx, refundID, actor)
	case constant.RefundStatusHistoryWaitingBankTransfer:
		s.recordRefundWaitingBankTransfer(ctx, refundID, actor)
	case constant.RefundStatusHistorySuccess:
		s.recordRefundSuccess(ctx, refundID, actor)
	case constant.RefundStatusHistoryFailed:
		s.recordRefundFailed(ctx, refundID, actor)
	case constant.RefundStatusHistoryCancelled:
		s.recordRefundCancelled(ctx, refundID, actor)
	default:
		s.recordStatusHistory(ctx, &statusHistoryModel.RecordRefundStatusHistoryRequest{
			RefundID: refundID,
			Status:   statusType,
			Actor:    actor,
		})
	}
}
