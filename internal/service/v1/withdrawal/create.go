package withdrawalService

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/bankAccount"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/bankTransfer"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"

	"github.com/go-redsync/redsync/v4"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
)

var balanceNames = map[string]string{
	// Balance Enum: Balance Name
	constant.AccountNameDisbursement: "Payout Balance",
	// Currently only supports withdrawals to payout balance
}

var (
	BankLimitExceededErrMessageFmt              = "The maximum withdrawal limit for this bank is IDR %s. Please adjust the amount or use a supported bank to withdraw a higher amount."
	MerchantLimitExceededErrMessageFmt          = "The maximum withdrawal limit for your account is IDR %s. Please adjust the amount."
	TransactionAmountBelowMinLimitErrMessageFmt = "The minimum withdrawal is IDR %s. Please adjust the amount."
)

func (s *withdrawalService) withdrawalRequestValidation(ctx context.Context, request *withdrawal.WithdrawalRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/withdrawal/withdrawalRequestValidation")
	defer segment.End()

	var (
		traceId, _ = ctx.Value(pdkConst.CtxTraceIdKey).(string)

		// Set default withdrawal config
		trxConfig = merchant.WithdrawalConfig{
			MinAmount: s.config.MinAmount,
			MaxAmount: s.config.MaxAmount,
		}
	)

	if withdrawalConfig, err := s.repo.GetTransactionConfig(ctx, request.MerchantId); err != nil {
		s.logger.Error(ctx, "Failed while get withdrawal transaction config", logger.Error(err))
		return err

	} else if withdrawalConfig != nil {
		trxConfig.MinAmount = withdrawalConfig.MinAmount
		trxConfig.MaxAmount = withdrawalConfig.MaxAmount
	}

	if request.IsFullAmount {
		availableBalance, err := s.orchestratorSvc.GetAvailableMerchantBalance(ctx, request.MerchantId, request.AccountName)
		if err != nil {
			s.logger.Error(ctx, "Get available merchant balance", logger.Error(err))
			return pkgErrs.New(response.HttpErrInternal, fmt.Errorf(constant.InternalErrorFmt, traceId))
		}
		request.Amount = math.Floor(availableBalance)
	}

	if util.HasDecimal(request.Amount) {
		return pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("The request was invalid because withdrawal amount must be an integer value."))
	}

	if request.Amount < trxConfig.MinAmount {
		return pkgErrs.New(response.HttpErrUnprocessableContent, fmt.Errorf(TransactionAmountBelowMinLimitErrMessageFmt, util.ConvertFloatToCurrency(trxConfig.MinAmount)))
	}

	// Maximum withdrawal amount validation is skipped for transactions originating from auto-withdrawal or balance transfer processes.
	if request.Type == constant.WithdrawalAutomated || request.Destination == constant.WithdrawalDestBalanceTransfer {
		return nil
	}

	if request.Amount > trxConfig.MaxAmount {
		return pkgErrs.New(response.HttpErrUnprocessableContent, fmt.Errorf(MerchantLimitExceededErrMessageFmt, util.ConvertFloatToCurrency(trxConfig.MaxAmount)))
	}

	isOverbooking := s.bankTransferConfig.IsBankcodeOverbookingChannelAllowed(ctx, request.BeneficiaryBankCode, request.MerchantId)
	if isOverbooking && request.Amount > s.config.LimitOverbooking {
		return pkgErrs.New(response.HttpErrUnprocessableContent, fmt.Errorf(BankLimitExceededErrMessageFmt, util.ConvertFloatToCurrency(s.config.LimitOverbooking)))

	} else if !isOverbooking && request.Amount > s.config.LimitNonOverbooking {
		return pkgErrs.New(response.HttpErrUnprocessableContent, fmt.Errorf(BankLimitExceededErrMessageFmt, util.ConvertFloatToCurrency(s.config.LimitNonOverbooking)))
	}
	return nil
}

func (s *withdrawalService) Create(ctx context.Context, request *withdrawal.WithdrawalRequest) (result *withdrawal.WithdrawalProcessResponse, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/withdrawal/Create")
	defer segment.End()

	traceId, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)

	if request.Source == constant.SourceOpenApi && request.Destination == constant.WithdrawalDestBankTransfer {
		bankAccounts, err := s.bankAccountRepo.GetListBankAccount(ctx, request.MerchantId)
		if err != nil {
			s.logger.Error(ctx, "Get list bank account", logger.Error(err))
			return nil, pkgErrs.New(response.HttpErrDatabase, err)

		} else if bankAccounts == nil {
			return nil, pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("bank account not found"))
		}

		request.BeneficiaryBankCode = bankAccounts[0].BeneficiaryBankCode
		request.BeneficiaryAccountNo = bankAccounts[0].BeneficiaryAccountNo
		request.DestinationAccountName = bankAccounts[0].BeneficiaryAccountName
	}

	if err = s.withdrawalRequestValidation(ctx, request); err != nil {
		return nil, err
	}

	if request.ReferenceID != "" {
		withdrawal, err := s.repo.GetByReferenceId(ctx, request.MerchantId, request.ReferenceID)
		if err != nil {
			s.logger.Error(ctx, "error when get withdrawal record by reference id", logger.String("merchantId", request.MerchantId), logger.String("referenceId", request.ReferenceID), logger.Error(err))
			return nil, pkgErrs.New(response.HttpErrInternal, fmt.Errorf(constant.InternalErrorFmt, traceId))
		}
		if withdrawal != nil {
			// Duplicate referenceID is a conflict on an existing resource → 409 (consistent with payout).
			return nil, pkgErrs.New(response.HttpErrDupCheck, fmt.Errorf("withdrawal with reference id %s already exists", request.ReferenceID))
		}
	}

	bankAccount := &bankAccount.BankAccountResponse{
		BeneficiaryAccountName: balanceNames[request.DestinationAccountName],
	}
	if request.Destination == constant.WithdrawalDestBankTransfer {
		// Bank Account Validation
		bankAccount, err = s.bankAccountRepo.GetBankAccountValidation(ctx, request.MerchantId, request.BeneficiaryBankCode, request.BeneficiaryAccountNo)
		if err != nil {
			s.logger.Error(ctx, "Get bank account validation", logger.Error(err))
			return nil, pkgErrs.New(response.HttpErrInternal, fmt.Errorf(constant.InternalErrorFmt, traceId))

		} else if bankAccount == nil {
			return nil, pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("bank account not found"))
		}
	}

	mutex := s.redis.NewMutex(
		"backend-portal:merchant-balances:"+request.MerchantId+":deduct:"+request.AccountName,
		redsync.WithExpiry(10*time.Second),
		redsync.WithRetryDelay(50*time.Millisecond),
		redsync.WithFailFast(true),
		redsync.WithTries(256),
	)
	if err := mutex.LockContext(ctx); err != nil {
		s.logger.Error(ctx, "Distributed lock for balance deduction", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, err)
	}
	unlockProcess := func() {
		if _, err := mutex.UnlockContext(ctx); err != nil {
			s.logger.Warn(ctx, "Failed unlock distributed lock", logger.Error(err))
		}
	}

	withdrawalData, err := s.CreateHistoryAndLedger(ctx, request, bankAccount)
	if err != nil {
		unlockProcess()
		return nil, err
	}

	unlockProcess()
	if request.Destination == constant.WithdrawalDestBalanceTransfer {
		return &withdrawal.WithdrawalProcessResponse{
			Id:          withdrawalData.id,
			Type:        request.Type,
			AccountName: request.AccountName,
			Amount: commonModel.Amount{
				Currency: constant.CurrencyIDR,
				Value:    fmt.Sprintf("%.0f", request.Amount),
			},
			Status:    constant.StatusSuccess,
			CreatedAt: withdrawalData.createdAt,
			UpdatedAt: withdrawalData.updatedAt,
		}, nil
	}

	// Withdrawal With Destination Bank Transfer

	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		RequestId:   traceId,
		From:        "Withdrawal",
		OriginId:    withdrawalData.id,
		ReferenceId: request.MerchantId,
	})

	snapCoreResp, err := s.snapCoreRepo.BankTransfer(ctx, &snapCoreModel.BankTransferRequest{
		BTBeneficiaryRequest: snapCoreModel.BTBeneficiaryRequest{
			BeneficiaryBankCode:    bankAccount.BeneficiaryBankCode,
			BeneficiaryAccountNo:   bankAccount.BeneficiaryAccountNo,
			BeneficiaryAccountName: bankAccount.BeneficiaryAccountName,
		},
		Currency:             constant.CurrencyIDR,
		Amount:               commonModel.Amount{Currency: constant.CurrencyIDR, Value: fmt.Sprintf("%.f", request.Amount)},
		Remark:               withdrawalData.id[24:],
		PurposeOfTransaction: snapCoreModel.DefaultPurchaseOfTransaction,
		TransactionDate:      withdrawalData.createdAt,
	}, &snapCoreModel.BankTransferHeaderRequest{
		ExternalId: withdrawalData.transactionId,
		MerchantId: request.MerchantId,
	})
	if snapCoreResp != nil && snapCoreResp.UUID != "" {

		withdrawalData.metadata.BankTransfer = &withdrawal.BankTransfer{
			UUID:               snapCoreResp.UUID,
			ExternalId:         snapCoreResp.ExternalID,
			BankReferenceNo:    snapCoreResp.BankReferenceNo,
			PartnerReferenceNo: snapCoreResp.PartnerReferenceNo,
		}
		if errUpdate := s.repo.UpdateMetadataById(ctx, withdrawalData.id, &withdrawalData.metadata); errUpdate != nil {
			s.logger.Error(ctx, "Update withdrawal metadata (bank transfer)", logger.Error(errUpdate))
		}

		if e := s.orchestratorSvc.UpdateProcessorAndReconReferenceByID(ctx, withdrawalData.transactionId, constant.SnapCoreProcessor, snapCoreResp.UUID, snapCoreResp.GetReconReferenceNo()); e != nil {
			s.logger.Error(ctx, "Update account transactions additional info", logger.Error(e))
		}
	}

	accTrxStatus, reasonType, reasonDesc, errMessage := constant.StatusPending, "", "", ""

	if request.Type == constant.WithdrawalManual {
		defer func() {
			message := fmt.Sprintf("Withdrawal request of Rp. %s ", util.ConvertFloatToCurrency(request.Amount))
			switch accTrxStatus {
			case constant.StatusSuccess:
				message += "successfully processed"

			case constant.StatusFailed:
				message += "failed to process"

			default:
				message += "is being processed"
			}
			_ = s.rmq.PublishActivity(ctx, &request.MerchantId, &request.UserId, constant.TagManualWithdrawal, message, result)
		}()
	}

	var updatedAtTrx time.Time

	if err != nil {

		accTrxStatus, reasonType, reasonDesc, errMessage = constant.StatusFailed, constant.ReasonTypeOtherReason, "", err.Error()
		if snapCoreResp != nil {
			accTrxStatus, reasonType, reasonDesc = snapCoreResp.MappingAccountTransactionErrStatus()
		}

		updatedAtTrx = time.Now().UTC()
		if errUpdate := s.orchestratorSvc.UpdateStatusAccountTransaction(ctx, withdrawalData.transactionId, accTrxStatus, &reasonType, &reasonDesc); errUpdate != nil {
			updatedAtTrx = time.Time{}
			s.logger.Error(ctx, "Update status account transaction (failed)", logger.Error(errUpdate))
		}

	} else if snapCoreResp != nil && snapCoreResp.Status == constant.SnapCoreBankTransferStatusSuccess {

		updatedAtTrx = time.Now().UTC()
		if errUpdate := s.orchestratorSvc.UpdateStatusAccountTransaction(ctx, withdrawalData.transactionId, constant.StatusSuccess, nil, nil); errUpdate == nil {
			accTrxStatus = constant.StatusSuccess

		} else {
			updatedAtTrx = time.Time{}
			s.logger.Error(ctx, "Update status account transaction (success)", logger.Error(errUpdate))
		}
	}

	s.logger.Info(
		ctx, "Withdrawal transaction status",
		logger.Any("details", map[string]string{
			"id":                     withdrawalData.id,
			"type":                   request.Type,
			"accountName":            request.AccountName,
			"merchantId":             request.MerchantId,
			"destination":            request.Destination,
			"beneficiaryBankName":    bankAccount.BeneficiaryBankName,
			"beneficiaryAccountNo":   bankAccount.BeneficiaryAccountNo,
			"beneficiaryAccountName": bankAccount.BeneficiaryAccountName,
		}),
		logger.String("status", accTrxStatus), logger.String("reasonType", reasonType), logger.String("reasonDescription", reasonDesc), logger.String("error", errMessage),
	)

	// TODO: Withdrawal failed, send alert to fin ops slack channel
	if request.Type == constant.WithdrawalAutomated && reasonType == constant.ReasonTypeBeneficiaryAccountReason {
		err = s.notificationSvc.SendFailedWithdrawalAlert(ctx, &withdrawal.FailedWithdrawalAlertRequest{
			AlertTitle:                 constant.WithdrawalFailedAlertTitle,
			WithdrawalID:               withdrawalData.id,
			MerchantID:                 request.MerchantId,
			BalanceName:                request.AccountName,
			BeneficiaryAccountName:     bankAccount.BeneficiaryAccountName,
			BeneficiaryAccountNo:       request.BeneficiaryAccountNo,
			BeneficiaryAccountBankName: bankAccount.BeneficiaryBankName,
			WithdrawType:               request.Type,
			Status:                     accTrxStatus,
			Amount:                     request.Amount,
			Reason:                     reasonType,
		})
		if err != nil {
			s.logger.Error(ctx, "error sending failed withdrawal alert", logger.Error(err))
		}
	}

	withdrawalResponse := &withdrawal.WithdrawalProcessResponse{
		Id:          withdrawalData.id,
		Type:        request.Type,
		AccountName: request.AccountName,
		Amount: commonModel.Amount{
			Currency: constant.CurrencyIDR,
			Value:    fmt.Sprintf("%.0f", request.Amount),
		},
		Status:    accTrxStatus,
		Reason:    reasonType,
		CreatedAt: withdrawalData.createdAt,
		UpdatedAt: withdrawalData.updatedAt,
	}
	if !updatedAtTrx.IsZero() {
		withdrawalResponse.UpdatedAt = updatedAtTrx
	}

	if request.Source == constant.SourceOpenApi && withdrawalResponse.Status != constant.StatusPending {
		callbackRequest := withdrawal.WithdrawalStatusCallbackRequest{
			ID:         withdrawalResponse.Id,
			MerchantId: request.MerchantId,
			Withdrawal: withdrawal.OpenAPIWithdrawalDetailResponse{
				ReferenceID:  request.ReferenceID,
				WithdrawType: request.Destination,
				IsFullAmount: request.IsFullAmount,
				Amount:       &withdrawalResponse.Amount,
				Description:  request.Description,
			},
			Status:    withdrawalResponse.Status,
			CreatedAt: withdrawalResponse.CreatedAt.Format(time.RFC3339),
			UpdatedAt: withdrawalResponse.UpdatedAt.Format(time.RFC3339),
		}
		if err := s.SendWithdrawalStatusCallback(ctx, callbackRequest); err != nil {
			s.logger.Error(ctx, "Failed to send withdrawal final status", logger.Error(err), logger.Any("callbackRequest", callbackRequest))
		}
	}
	return withdrawalResponse, nil
}

func (s *withdrawalService) CreateHistoryAndLedger(ctx context.Context, request *withdrawal.WithdrawalRequest, bankAccount *bankAccount.BankAccountResponse) (*withdrawalCreateResult, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/withdrawal/CreateHistoryAndLedger")
	defer segment.End()

	traceId, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)

	ctxTx, err := s.repo.BeginTransaction(ctx)
	if err != nil {
		s.logger.Error(ctx, "Begin transaction", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrInternal, fmt.Errorf(constant.InternalErrorFmt, traceId))
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

	availableBalance, err := s.orchestratorSvc.GetAvailableMerchantBalance(ctxTx, request.MerchantId, request.AccountName)
	if err != nil {
		s.logger.Error(ctx, "Get available merchant balance", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrInternal, fmt.Errorf(constant.InternalErrorFmt, traceId))

	} else if availableBalance < request.Amount {
		// Insufficient balance → 403 for consistency with /payouts/{id}/retry and /inquiry-account handling.
		return nil, pkgErrs.New(response.HttpErrForbidden, constant.ErrInsufficientBalance)
	}

	withdrawalId, _ := uuid.NewV7()
	withdrawalData := &withdrawal.Withdrawal{
		Id:                     withdrawalId.String(),
		MerchantId:             request.MerchantId,
		ReferenceId:            request.ReferenceID,
		BeneficiaryBankCode:    bankAccount.BeneficiaryBankCode,
		BeneficiaryBankName:    bankAccount.BeneficiaryBankName,
		BeneficiaryAccountNo:   bankAccount.BeneficiaryAccountNo,
		BeneficiaryAccountName: bankAccount.BeneficiaryAccountName,
		Type:                   request.Type,
		Currency:               constant.CurrencyIDR,
		Description:            request.Description,
		Amount:                 request.Amount,
		CreatedBy:              request.UserId,
		CreatedAt:              time.Now().UTC(), UpdatedAt: time.Now().UTC(),
		Metadata: withdrawal.Metadata{
			Source:       request.Source,
			WithdrawType: request.Destination,
			IsFullAmount: request.IsFullAmount,
		},
	}
	if request.DestinationAccountName == constant.AccountNameDisbursement {
		withdrawalData.Metadata.BalanceType = constant.WithdrawalPayoutBalanceDestination
	}
	if err = s.repo.Create(ctxTx, withdrawalData); err != nil {
		s.logger.Error(ctx, "Create withdrawal history", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrInternal, fmt.Errorf(constant.InternalErrorFmt, traceId))
	}

	transactionStatus := constant.StatusPending
	transactionChannel := constant.ChannelBankTransfer
	if request.Destination == constant.WithdrawalDestBalanceTransfer {
		transactionStatus = constant.StatusSuccess
		transactionChannel = constant.ChannelBalanceTransfer
	}
	// Debit Transaction
	merchantUUID, _ := uuid.Parse(request.MerchantId)
	ledgerRequest := &orchestrator_model.CreateAccountTransactionRequest{
		UUID:                 uuid.New(),
		ReferenceID:          withdrawalId.String(),
		MerchantID:           merchantUUID,
		Currency:             constant.CurrencyIDR,
		Debit:                request.Amount,
		Type:                 constant.TypeWithdrawal,
		Channel:              transactionChannel,
		Status:               transactionStatus,
		TransactionTimestamp: withdrawalData.CreatedAt,
		Usecase:              request.AccountName,
		AdditionalInfo: types.NullJSONText{
			Valid:    true,
			JSONText: fmt.Appendf(nil, `{"type": "%s"}`, request.Type),
		},
	}

	if err = s.orchestratorSvc.PostAccountTransaction(ctxTx, ledgerRequest); err != nil {
		s.logger.Error(ctx, "Create ledger of withdrawal transaction", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrInternal, fmt.Errorf(constant.InternalErrorFmt, traceId))
	}
	if request.Destination == constant.WithdrawalDestBalanceTransfer {
		// Credit Transactions
		ledgerRequest := &orchestrator_model.CreateAccountTransactionRequest{
			UUID:                 uuid.New(),
			ReferenceID:          withdrawalId.String(),
			MerchantID:           merchantUUID,
			Currency:             constant.CurrencyIDR,
			Credit:               request.Amount,
			Type:                 constant.TypeGeneralTopUp,
			Channel:              constant.ChannelBalanceTransfer,
			Status:               transactionStatus,
			TransactionTimestamp: withdrawalData.CreatedAt,
			Usecase:              request.DestinationAccountName,
			AdditionalInfo: types.NullJSONText{
				Valid:    true,
				JSONText: fmt.Appendf(nil, `{"type": "%s", "source": "%s"}`, constant.TypeWithdrawal, request.AccountName),
			},
		}
		if err = s.orchestratorSvc.PostAccountTransaction(ctxTx, ledgerRequest); err != nil {
			s.logger.Error(ctx, "Create ledger of top up balance transaction", logger.Error(err))
			return nil, pkgErrs.New(response.HttpErrInternal, fmt.Errorf(constant.InternalErrorFmt, traceId))
		}
	}

	if err = s.repo.CommitTransaction(ctxTx); err != nil {
		s.logger.Error(ctx, "Commit session transaction", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrInternal, fmt.Errorf(constant.InternalErrorFmt, traceId))
	}

	isCompleted = true

	return &withdrawalCreateResult{
		id:            withdrawalData.Id,
		transactionId: ledgerRequest.UUID.String(),
		createdAt:     withdrawalData.CreatedAt,
		updatedAt:     withdrawalData.UpdatedAt,
		metadata:      withdrawalData.Metadata,
	}, nil
}
