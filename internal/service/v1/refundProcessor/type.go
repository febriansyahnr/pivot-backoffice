package refundProcessorService

import (
	"context"

	routingProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/bankTransfer"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	services "github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("RefundProcessor")

type BankTransferStrategy struct {
	snapCoreRepo repository.ISnapCoreRepository
	refundRepo   repository.IRefundRepository

	beneficiaryAccountSvc services.IBeneficiaryAccountService
	orchestratorSvc       services.IOrchestratorService

	logger logger.ILogger
	redis  redisExt.IRedisExt
}

func NewBankTransferStrategy(
	snapCoreRepo repository.ISnapCoreRepository,
	refundRepo repository.IRefundRepository,
	beneficiaryAccountSvc services.IBeneficiaryAccountService,
	orchestratorSvc services.IOrchestratorService,
	logger logger.ILogger,
	redis redisExt.IRedisExt,
) *BankTransferStrategy {
	return &BankTransferStrategy{
		snapCoreRepo:          snapCoreRepo,
		refundRepo:            refundRepo,
		beneficiaryAccountSvc: beneficiaryAccountSvc,
		orchestratorSvc:       orchestratorSvc,
		logger:                logger,
		redis:                 redis,
	}
}

type CardStrategy struct {
	logger         logger.ILogger
	creditCardRepo repository.ICreditcardCoreProcessorRepository
}

func NewCardStrategy(creditCardRepo repository.ICreditcardCoreProcessorRepository, logger logger.ILogger) *CardStrategy {
	return &CardStrategy{logger: logger, creditCardRepo: creditCardRepo}
}

type QRISStrategy struct {
	snapCoreRepo repository.ISnapCoreRepository
	logger       logger.ILogger
}

func NewQRISStrategy(snapCoreRepo repository.ISnapCoreRepository, logger logger.ILogger) *QRISStrategy {
	return &QRISStrategy{snapCoreRepo: snapCoreRepo, logger: logger}
}

type EWalletStrategy struct {
	snapCoreRepo repository.ISnapCoreRepository
	logger       logger.ILogger
}

func NewEWalletStrategy(snapCoreRepo repository.ISnapCoreRepository, logger logger.ILogger) *EWalletStrategy {
	return &EWalletStrategy{snapCoreRepo: snapCoreRepo, logger: logger}
}

type RefundProcessor struct {
	refundRepo repository.IRefundRepository

	bankTransfer    services.IRefundProcessorService
	card            services.IRefundProcessorService
	qris            services.IRefundProcessorService
	ewallet         services.IRefundProcessorService
	orchestratorSvc services.IOrchestratorService
	refundSvc       services.IRefundService
	feeSvc          services.IFeeService
	transferSvc     services.ITransferService
	settlementSvc   services.ISettlementService
	merchantSvc     services.IMerchantService

	redis  redisExt.IRedisExt
	logger logger.ILogger
}

type RefundProcessorFunc func(service *RefundProcessor)

func New(
	logger logger.ILogger,
	refundRepo repository.IRefundRepository,
	snapCoreRepo repository.ISnapCoreRepository,
	creditCardRepo repository.ICreditcardCoreProcessorRepository,
	beneficiaryAccountSvc services.IBeneficiaryAccountService,
	orchestratorSvc services.IOrchestratorService,
	redis redisExt.IRedisExt,
	depends ...RefundProcessorFunc,
) services.IRefundProcessorService {
	bankTransfer := NewBankTransferStrategy(snapCoreRepo, refundRepo, beneficiaryAccountSvc, orchestratorSvc, logger, redis)
	card := NewCardStrategy(creditCardRepo, logger)
	qris := NewQRISStrategy(snapCoreRepo, logger)
	ewallet := NewEWalletStrategy(snapCoreRepo, logger)

	s := &RefundProcessor{
		refundRepo:      refundRepo,
		bankTransfer:    bankTransfer,
		card:            card,
		qris:            qris,
		ewallet:         ewallet,
		orchestratorSvc: orchestratorSvc,
		redis:           redis,
		logger:          logger,
	}

	for _, d := range depends {
		d(s)
	}

	return s
}

func WithRefundService(svc services.IRefundService) RefundProcessorFunc {
	return func(rs *RefundProcessor) {
		rs.refundSvc = svc
	}
}

func WithFeeService(svc services.IFeeService) RefundProcessorFunc {
	return func(rs *RefundProcessor) {
		rs.feeSvc = svc
	}
}

func WithTransferService(svc services.ITransferService) RefundProcessorFunc {
	return func(rs *RefundProcessor) {
		rs.transferSvc = svc
	}
}

func WithSettlementService(svc services.ISettlementService) RefundProcessorFunc {
	return func(rs *RefundProcessor) {
		rs.settlementSvc = svc
	}
}

func WithMerchantService(svc services.IMerchantService) RefundProcessorFunc {
	return func(rs *RefundProcessor) {
		rs.merchantSvc = svc
	}
}

func (s *BankTransferStrategy) ProcessUpdateBankTransferStatus(ctx context.Context, request *routingProcessorModel.BankTransferResponseData) error {
	s.logger.Warn(ctx, "No need to implement")
	return nil
}

func (s *CardStrategy) ProcessUpdateBankTransferStatus(ctx context.Context, request *routingProcessorModel.BankTransferResponseData) error {
	s.logger.Warn(ctx, "No need to implement")
	return nil
}

func (s *QRISStrategy) ProcessUpdateBankTransferStatus(ctx context.Context, request *routingProcessorModel.BankTransferResponseData) error {
	s.logger.Warn(ctx, "No need to implement")
	return nil
}

func (s *EWalletStrategy) ProcessUpdateBankTransferStatus(ctx context.Context, request *routingProcessorModel.BankTransferResponseData) error {
	s.logger.Warn(ctx, "No need to implement")
	return nil
}
