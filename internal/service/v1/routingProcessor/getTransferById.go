package routingprocessorService

import (
	"context"
	"errors"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	routingProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/bankTransfer"
)

func (s *routingProcessorService) GetTransferByID(ctx context.Context, transaction *orchestratorModel.AccountTransactionWithUseCase, forceFailed bool) (*routingProcessorModel.BankTransferResponseData, error) {
	ctx, span := tracer.Start(ctx, "internal/service/v1/routingProcessor/GetTransferByID")
	defer span.End()

	var (
		transfer *routingProcessorModel.BankTransferResponseData
		err      = errors.New("processor not found")
	)

	processorName := transaction.ProcessorReference

	if processorName == "" {
		processorName = constant.SnapCoreProcessor
	}

	// if processorName == constant.DanaPGProcessor && bankTransfer.IsWalletCode(transaction.BeneficiaryBankName) {
	// 	disbursmentType = constant.DisbursementTypeWallet
	// }
	ctx = context.WithValue(ctx, constant.CtxDisburesementType, constant.DisbursementTypeTransfer)

	if s.routingProcessor == nil {
		return nil, fmt.Errorf("routingProcessor map is nil")
	}

	processor, exists := s.routingProcessor[processorName]
	if !exists {
		return nil, fmt.Errorf("processor %s not found", processorName)
	}

	transfer, err = processor.GetTransferById(ctx, transaction.UUID.String(), forceFailed)

	return transfer, err
}
