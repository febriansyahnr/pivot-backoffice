package adjustment

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	adjustModel "github.com/paper-indonesia/pivot-backoffice/internal/model/adjustment"
	orchestraModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *adjustmentService) CreateMerchantBalanceAdjustment(ctx context.Context, req *adjustModel.MerchantBalanceAdjustmentRequest) (*adjustModel.ManualAdjustmentHistory, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/adjustment/CreateMerchantBalanceAdjustment")
	defer segment.End()

	merchant, err := s.merchantRepo.FindMerchantByID(ctx, req.MerchantId)
	if err != nil {
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	} else if merchant == nil {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrMerchantNotFound)
	}

	if ctx, err = s.repo.BeginTransaction(ctx); err != nil {
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}
	isCompleted := false
	defer func() {
		if !isCompleted {
			if e := s.repo.RollbackTransaction(ctx); e != nil {
				s.logger.Error(ctx, "failed to rollback transaction", logger.Error(e))
			}
		}
	}()

	now := time.Now().UTC()
	amount := req.Credit
	adjustmentType := constant.AccountNameDisbursement
	if amount == 0 && req.Debit > 0 {
		amount = req.Debit * -1
	}
	if amount == 0 {
		return nil, pkgErrs.New(response.HttpErrRequest, constant.ErrInvalidAmount)
	}
	switch req.BalanceType {
	case constant.AdjustmentPayoutBalanceDestination:
		adjustmentType = constant.AccountNameDisbursement
	case constant.AdjustmentPaymentBalanceDestination:
		adjustmentType = constant.AccountNamePayment
	default:
		return nil, pkgErrs.New(response.HttpErrRequest, constant.ErrAdjustmentInvalidBalanceType)
	}

	jsonBankAccount, _ := json.Marshal("{}")
	data := &adjustModel.ManualAdjustmentHistory{
		UUID:            uuid.NewString(),
		MerchantID:      req.MerchantId,
		TransactionDate: time.Now().UTC(),
		BankAccount:     string(jsonBankAccount),
		ProofOfTransfer: "",
		Type:            adjustmentType,
		Currency:        req.Currency,
		Amount:          amount,
		Notes:           req.Remarks,
		CreatedBy:       req.CreatedBy,
		ReferenceID:     req.ReferenceId,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err = s.repo.CreateAdjustment(ctx, data); err != nil {
		s.logger.Error(ctx, "failed to insert adjustment", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}

	ledger := &orchestraModel.CreateAccountTransactionRequest{
		UUID:                 uuid.New(),
		ReferenceID:          data.UUID,
		Currency:             req.Currency,
		Credit:               req.Credit,
		Debit:                req.Debit,
		Type:                 constant.TypeManualAdjust,
		Channel:              constant.ChannelBalanceAdjustment,
		Status:               constant.StatusSuccess,
		Remarks:              req.Remarks,
		TransactionTimestamp: data.TransactionDate,
		Usecase:              adjustmentType,
	}

	ledger.MerchantID, _ = uuid.Parse(req.MerchantId)

	if err = s.orchestrator.PostAccountTransaction(ctx, ledger); err != nil {
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}

	if err = s.repo.CommitTransaction(ctx); err != nil {
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}
	isCompleted = true

	return data, nil
}
