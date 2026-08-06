package withdrawalService

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *withdrawalService) TransferBalance(ctx context.Context, request *withdrawal.WithdrawalTransferBalanceRequest) (*withdrawal.WithdrawalTransferBalanceResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/withdrawal/TransferBalance")
	defer segment.End()

	traceId, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)

	if util.HasDecimal(request.Amount) {
		return nil, pkgErr.New(response.HttpErrUnprocessableContent, errors.New("The request was invalid because withdrawal amount must be an integer value."))
	}

	// Get available
	availableBalance, err := s.orchestratorSvc.GetAvailableMerchantBalance(ctx, request.MerchantID, request.SourceAccountName)
	if err != nil {
		s.logger.Error(ctx, "Get available merchant balance", logger.Error(err))
		return nil, pkgErr.New(response.HttpErrInternal, fmt.Errorf(constant.InternalErrorFmt, traceId))

	} else if request.Amount > availableBalance {
		return nil, pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrInsufficientBalance)
	}

	ctxTx, err := s.repo.BeginTransaction(ctx)
	if err != nil {
		s.logger.Error(ctx, "Start transaction", logger.Error(err))
		return nil, pkgErr.New(response.HttpErrInternal, err)
	}

	isCompleted := false
	defer func() {
		if isCompleted {
			return
		}
		if e := s.repo.RollbackTransaction(ctxTx); e != nil {
			s.logger.Error(ctx, "Rollback session transaction", logger.Error(e))
		}
	}()

	beneficiaryAccountName := "Payout Balance"
	if request.DestinationAccountName == constant.TypePayment {
		beneficiaryAccountName = "Payment Balance"
	}

	withdrawalId, _ := uuid.NewV7()
	withdrawalData := &withdrawal.Withdrawal{
		Id:                     withdrawalId.String(),
		MerchantId:             request.MerchantID,
		BeneficiaryBankCode:    "",
		BeneficiaryBankName:    "",
		BeneficiaryAccountNo:   "",
		BeneficiaryAccountName: beneficiaryAccountName,
		Type:                   constant.WithdrawalManual,
		Currency:               constant.CurrencyIDR,
		Amount:                 request.Amount,
		CreatedBy:              request.UserID,
		CreatedAt:              time.Now().UTC(), UpdatedAt: time.Now().UTC(),
		RawMetadata: types.NullJSONText{
			Valid:    true,
			JSONText: []byte(fmt.Sprintf(`{"source":"%s","destination":"%s"}`, request.SourceAccountName, request.DestinationAccountName)),
		},
	}
	if err = s.repo.Create(ctxTx, withdrawalData); err != nil {
		s.logger.Error(ctx, "Create withdrawal history", logger.Error(err))
		return nil, pkgErr.New(response.HttpErrInternal, fmt.Errorf(constant.InternalErrorFmt, traceId))
	}

	merchantUUID, _ := uuid.Parse(request.MerchantID)
	debitLedgerRequest := &orchestratorModel.CreateAccountTransactionRequest{
		UUID:                 uuid.New(),
		ReferenceID:          withdrawalId.String(),
		MerchantID:           merchantUUID,
		Currency:             constant.CurrencyIDR,
		Debit:                request.Amount,
		Type:                 constant.TypeWithdrawal,
		Channel:              constant.ChannelBalanceTransfer,
		Status:               constant.StatusSuccess,
		TransactionTimestamp: withdrawalData.CreatedAt,
		Usecase:              request.SourceAccountName,
		AdditionalInfo: types.NullJSONText{
			Valid:    true,
			JSONText: []byte(fmt.Sprintf(`{"type":"%s"}`, constant.WithdrawalManual)),
		},
	}
	if err = s.orchestratorSvc.PostAccountTransaction(ctxTx, debitLedgerRequest); err != nil {
		s.logger.Error(ctx, "Create account transaction", logger.Error(err))
		return nil, pkgErr.New(response.HttpErrInternal, fmt.Errorf(constant.InternalErrorFmt, traceId))
	}

	creditLedgerRequest := &orchestratorModel.CreateAccountTransactionRequest{
		UUID:                 uuid.New(),
		ReferenceID:          withdrawalId.String(),
		MerchantID:           merchantUUID,
		Currency:             constant.CurrencyIDR,
		Credit:               request.Amount,
		Type:                 constant.TypeGeneralTopUp,
		Channel:              constant.ChannelBalanceTransfer,
		Status:               constant.StatusSuccess,
		TransactionTimestamp: withdrawalData.CreatedAt,
		Usecase:              request.DestinationAccountName,
		AdditionalInfo: types.NullJSONText{
			Valid:    true,
			JSONText: []byte(fmt.Sprintf(`{"type":"%s"}`, constant.TypeWithdrawal)),
		},
	}
	if err = s.orchestratorSvc.PostAccountTransaction(ctxTx, creditLedgerRequest); err != nil {
		s.logger.Error(ctx, "Create account transaction", logger.Error(err))
		return nil, pkgErr.New(response.HttpErrInternal, fmt.Errorf(constant.InternalErrorFmt, traceId))
	}

	if err = s.repo.CommitTransaction(ctxTx); err != nil {
		s.logger.Error(ctx, "Commit session transaction", logger.Error(err))
		return nil, pkgErr.New(response.HttpErrInternal, fmt.Errorf(constant.InternalErrorFmt, traceId))
	}
	isCompleted = true

	return &withdrawal.WithdrawalTransferBalanceResponse{
		Id: withdrawalId.String(),
	}, nil
}
