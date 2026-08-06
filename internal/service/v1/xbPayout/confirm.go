package xbPayoutService

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	statusHistoryModel "github.com/paper-indonesia/pivot-backoffice/internal/model/statusHistory"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	xbCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xbCoreProcessor"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *xbPayoutService) Confirm(ctx context.Context, request *xbModel.ConfirmPayoutRequest) (*xbModel.ConfirmPayoutResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/xbPayout/Confirm")
	defer segment.End()

	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		OriginId:    request.PayoutId,
		ReferenceId: request.MerchantId,
		From:        serviceName,
	})

	// Find payout by ID
	disbursement, err := s.disbursementRepo.FindByID(ctx, request.PayoutId)
	if err != nil {
		s.logger.Error(ctx, "Confirm - Find disbursement by uuid", logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrDatabase, err)

	} else if disbursement == nil || disbursement.MerchantID != request.MerchantId || disbursement.MetadataObj.XbDetail == nil {
		s.logger.Info(ctx, "Confirm - Payout not found or merchant ID does not match or missing XB metadata", logger.Any("request", map[string]string{
			"merchantId": request.MerchantId,
			"payoutId":   request.PayoutId,
		}))
		return nil, pkgErrors.New(response.HttpErrNotFound, constant.ErrPayoutIsNotFound)
	}

	// Return if expired
	if time.Now().UTC().After(disbursement.MetadataObj.XbDetail.ExpiredAt) {
		s.logger.Error(
			ctx,
			"Confirm - Payout has already expired",
			logger.Any("expired_at", disbursement.MetadataObj.XbDetail.ExpiredAt),
			logger.Any("current_time", time.Now().UTC()),
			logger.String("payout_id", request.PayoutId),
		)
		return nil, pkgErrors.New(response.HttpErrRequest, constant.ErrPayoutAlreadyExpired)
	}

	// update status = approved
	err = s.disbursementRepo.ApproveInBulk(ctx, []string{request.PayoutId}, request.ApprovedBy)
	if err != nil {
		if errors.Is(err, constant.ErrNoRowsAffected) {
			err = constant.ErrDisbursementStatusAlreadyApproved
			s.logger.Error(
				ctx,
				"Confirm - Disbursement already approved, no rows affected",
				logger.String("payoutId", request.PayoutId),
				logger.Error(err),
			)
			return nil, pkgErrors.New(response.HttpErrRequest, err)
		}

		s.logger.Error(
			ctx,
			"Confirm - Failed to approve disbursement in bulk",
			logger.String("payoutId", request.PayoutId),
			logger.Error(err),
		)
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	}

	// Convert amount
	amountToBePaidInSourceCurrency, _ := disbursement.MetadataObj.XbDetail.TotalAmount.Float64()
	amountToBePaidInSourceCurrency += disbursement.MetadataObj.FeeDetail.FinalAmount

	// Check available balance in IDR
	availableBalance, err := s.orchestratorSvc.GetAvailableMerchantBalance(ctx, request.MerchantId, constant.TypeDisbursement)
	if err != nil {
		s.logger.Error(
			ctx,
			"Confirm - Failed to get available merchant balance",
			logger.String("merchantId", request.MerchantId),
			logger.Error(err),
		)
		return nil, err
	}

	// simulation case when account number is XBSimulationInsufficientBalanceNumber and first confirm,
	// it should get insufficient balance, but after re confirm, it should get success
	isSimulationCase := s.config.Environment != constant.EnvironmentProduction &&
		disbursement.BeneficiaryAccountNo == constant.XBSimulationInsufficientBalanceNumber &&
		util.ValueOfPtr(disbursement.ReasonType) != constant.XbDisbursementReasonTypeInsufficientBalance

	// Return error if insufficient
	if amountToBePaidInSourceCurrency > availableBalance || isSimulationCase {
		if errUpdate := s.disbursementRepo.UpdateReasonByIDs(
			ctx,
			[]string{request.PayoutId},
			constant.XbDisbursementReasonTypeInsufficientBalance,
			constant.XbDisbursementReasonDescInsufficientBalance,
		); errUpdate != nil {
			s.logger.Error(
				ctx,
				"Confirm - Failed to update disbursement reason due to insufficient balance",
				logger.String("payoutId", request.PayoutId),
				logger.Error(errUpdate),
			)
			return nil, pkgErrors.New(response.HttpErrDatabase, errUpdate)
		}

		s.logger.Error(
			ctx,
			"Confirm - Insufficient balance for payout",
			logger.String("payoutId", request.PayoutId),
			logger.Float64("amountToBePaid", amountToBePaidInSourceCurrency),
			logger.Float64("availableBalance", availableBalance),
		)
		return nil, pkgErrors.New(response.HttpErrForbidden, constant.ErrInsufficientBalance)
	}

	// If sufficient, Call to XB Core processor
	xbAcquirerTransactionId := ""
	if disbursement.ProcessorReferenceID != nil {
		xbAcquirerTransactionId = *disbursement.ProcessorReferenceID
	}

	xbResp, err := s.xbCoreProcessorRepo.ConfirmPayout(ctx, &xbCoreProcessorModel.ConfirmPayoutRequest{
		XbPayoutId:            disbursement.MetadataObj.XbDetail.Uuid,
		AcquirerTransactionId: xbAcquirerTransactionId,
		MerchantId:            request.MerchantId,
	})

	// Tracking status
	actor := request.ApprovedBy
	if actor == "" {
		actor = constant.UserSystemType
	}
	_ = s.RecordStatusHistory(ctx, &statusHistoryModel.RecordDisbursementStatusHistoryRequest{
		DisbursementID: disbursement.UUID,
		Status:         constant.XbStatusConfirmed,
		Actor:          actor,
	})

	if err != nil {
		s.logger.Error(
			ctx,
			"Confirm - Failed to confirm payout to xb-core-processor",
			logger.String("payoutId", request.PayoutId),
			logger.String("xbPayoutId", disbursement.MetadataObj.XbDetail.Uuid),
			logger.String("merchantId", request.MerchantId),
			logger.Error(err),
		)
		errHandleConfirmation := s.handleConfirmationError(ctx, request, disbursement)
		if errHandleConfirmation != nil {
			s.logger.Error(ctx, "Confirm - Error handle confirmation error",
				logger.String("payoutId", request.PayoutId),
				logger.String("xbPayoutId", disbursement.MetadataObj.XbDetail.Uuid),
				logger.Error(errHandleConfirmation),
			)

		}
		return nil, err
	}

	processorReferenceID := ""
	if disbursement.ProcessorReferenceID != nil {
		processorReferenceID = *disbursement.ProcessorReferenceID
	}
	if errUpdate := s.disbursementRepo.UpdateProcessorReferenceIdAndBankReferenceNo(
		ctx, disbursement.UUID, processorReferenceID, xbResp.PartnerTransactionId); errUpdate != nil {
		s.logger.Error(
			ctx,
			"Confirm - Failed to update processorReferenceId and bankReferenceNo",
			logger.String("payoutId", request.PayoutId),
			logger.String("disbursementId", disbursement.UUID),
			logger.String("processorReferenceId", processorReferenceID),
			logger.String("bankReferenceNo", xbResp.PartnerTransactionId),
			logger.Error(errUpdate),
		)
		return nil, pkgErrors.New(response.HttpErrDatabase, errUpdate)
	}

	if errUpdate := s.disbursementRepo.UpdateReasonByIDs(
		ctx,
		[]string{request.PayoutId},
		constant.XbDisbursementReasonTypePending,
		constant.XbDisbursementReasonDescPending,
	); errUpdate != nil {
		s.logger.Error(
			ctx,
			"Confirm - Failed to update disbursement reason to pending",
			logger.String("payoutId", request.PayoutId),
			logger.Error(errUpdate),
		)
		return nil, pkgErrors.New(response.HttpErrDatabase, errUpdate)
	}

	// Create ledger or update ledger
	merchantUUID, _ := uuid.Parse(disbursement.MerchantID)
	totalAmountInSourceCurrency, _ := disbursement.MetadataObj.XbDetail.TotalAmount.Round(2).Float64()
	remark := ""
	if disbursement.Remark != nil {
		remark = *disbursement.Remark
	}

	// check if ledger already exists
	ledger, err := s.orchestratorSvc.FindByReference(ctx, disbursement.UUID, orchestratorModel.TypeDisbursement)
	if err != nil {
		s.logger.Error(
			ctx,
			"Confirm - Failed to find ledger",
			logger.String("payoutId", request.PayoutId),
			logger.String("disbursementId", disbursement.UUID),
			logger.Error(err),
		)
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	}

	feeAmount := disbursement.MetadataObj.FeeDetail.FinalAmount

	if ledger != nil {
		_, reasonType, status := constant.MapXbProcessorStatusToCoreStatus(xbResp.Status)
		if status == "" {
			status = constant.StatusPending
		}
		if errLedger := s.orchestratorSvc.
			UpdateStatusAccountTransactionByReferenceID(ctx, disbursement.UUID, status, &reasonType, nil); errLedger != nil {
			s.logger.Error(
				ctx,
				"Confirm - Failed to update ledger status",
				logger.String("payoutId", request.PayoutId),
				logger.String("disbursementId", disbursement.UUID),
				logger.String("status", xbResp.Status),
				logger.String("reasonType", reasonType),
				logger.Error(errLedger),
			)
			return nil, pkgErrors.New(response.HttpErrDatabase, errLedger)
		}
	} else {
		if errLedger := s.orchestratorSvc.
			PostAccountTransaction(ctx, &orchestratorModel.CreateAccountTransactionRequest{
				UUID:                 uuid.New(),
				ReferenceID:          disbursement.UUID,
				Type:                 orchestratorModel.TypeDisbursement,
				MerchantID:           merchantUUID,
				Currency:             disbursement.MetadataObj.XbDetail.SourceCurrency,
				Credit:               0.00,
				Debit:                totalAmountInSourceCurrency,
				Channel:              constant.ChannelXB,
				Status:               constant.StatusPending,
				Remarks:              remark,
				TransactionTimestamp: xbResp.UpdatedAt,
				Usecase:              constant.TypeDisbursement,
			}); errLedger != nil {
			s.logger.Error(
				ctx,
				"Confirm - Failed to post account transaction to ledger",
				logger.String("payoutId", request.PayoutId),
				logger.String("referenceId", disbursement.UUID),
				logger.Any("merchantId", merchantUUID),
				logger.Float64("amount", totalAmountInSourceCurrency),
				logger.String("currency", disbursement.MetadataObj.XbDetail.SourceCurrency),
				logger.Error(errLedger),
			)
			return nil, pkgErrors.New(response.HttpErrDatabase, errLedger)
		}
		// create transaction fee - use stored final fee amount (already calculated with FX conversion during creation)
		feeTrxRequest := &orchestratorModel.CreateAccountTransactionRequest{
			UUID:                 uuid.New(),
			ReferenceID:          disbursement.UUID,
			Type:                 orchestratorModel.TypeFee,
			MerchantID:           merchantUUID,
			Currency:             constant.CurrencyIDR,
			Credit:               0.00,
			Debit:                feeAmount,
			Channel:              "",
			Status:               constant.StatusPending,
			Remarks:              remark,
			TransactionTimestamp: disbursement.CreatedAt,
			Usecase:              constant.TypeDisbursement,
		}
		feeTrxRequest.AdditionalInfo.Valid = true
		feeTrxRequest.AdditionalInfo.JSONText, _ = json.Marshal(disbursement.MetadataObj.FeeDetail)

		if errFeeLedger := s.orchestratorSvc.PostAccountTransaction(ctx, feeTrxRequest); errFeeLedger != nil {
			s.logger.Error(
				ctx,
				"Confirm - Failed to post account transaction fee to ledger",
				logger.String("payoutId", request.PayoutId),
				logger.String("referenceId", disbursement.UUID),
				logger.Any("merchantId", merchantUUID),
				logger.Float64("amount", feeAmount),
				logger.String("currency", constant.CurrencyIDR),
				logger.Error(errFeeLedger),
			)
			return nil, pkgErrors.New(response.HttpErrDatabase, errFeeLedger)
		}
	}

	_ = s.RecordStatusHistory(ctx, &statusHistoryModel.RecordDisbursementStatusHistoryRequest{
		DisbursementID: disbursement.UUID,
		Status:         xbResp.Status,
		Actor:          actor,
	})

	return &xbModel.ConfirmPayoutResponse{
		Uuid:                disbursement.UUID,
		MerchantId:          disbursement.MerchantID,
		ReferenceId:         disbursement.ReferenceID,
		SourceCurrency:      xbResp.SourceCurrency,
		DestinationCurrency: xbResp.DestinationCurrency,
		DestinationAmount:   xbResp.DestinationAmount,
		FxRate:              xbResp.FxRate,
		DestinationFxRate:   xbResp.DestinationFxRate,
		Fee:                 decimal.NewFromFloat(feeAmount),
		TotalAmount:         xbResp.TotalAmount,
		Remark:              xbResp.StatementNarrative,
		CreatedAt:           xbResp.CreatedAt,
		Status:              s.mapStatus(xbResp.Status),
		SenderId:            xbResp.SenderId,
		BeneficiaryId:       xbResp.BeneficiaryId,
		BeneficiaryData: xbModel.BeneficiaryDataResponse{
			Name:               xbResp.BeneficiaryData.Name,
			Address:            xbResp.BeneficiaryData.Address,
			City:               xbResp.BeneficiaryData.City,
			Postcode:           xbResp.BeneficiaryData.Postcode,
			State:              xbResp.BeneficiaryData.State,
			CountryCode:        xbResp.BeneficiaryData.CountryCode,
			CountryName:        xbResp.BeneficiaryData.CountryName,
			AccountType:        xbResp.BeneficiaryData.AccountType,
			AccountNumber:      xbResp.BeneficiaryData.AccountNumber,
			BankName:           xbResp.BeneficiaryData.BankName,
			BankCode:           xbResp.BeneficiaryData.BankCode,
			ContactCountryCode: xbResp.BeneficiaryData.ContactCountryCode,
			ContactNumber:      xbResp.BeneficiaryData.ContactNumber,
			Email:              xbResp.BeneficiaryData.Email,
		},
		SenderData: xbModel.SenderDataResponse{
			Name:                 xbResp.SenderData.Name,
			Address:              xbResp.SenderData.Address,
			City:                 xbResp.SenderData.City,
			Postcode:             xbResp.SenderData.Postcode,
			State:                xbResp.SenderData.State,
			CountryCode:          xbResp.SenderData.CountryCode,
			CountryName:          xbResp.SenderData.CountryName,
			AccountType:          xbResp.SenderData.AccountType,
			IdentificationType:   xbResp.SenderData.IdentificationType,
			IdentificationNumber: xbResp.SenderData.IdentificationNumber,
			BankAccountNumber:    xbResp.SenderData.BankAccountNumber,
			ContactCountryCode:   xbResp.SenderData.ContactCountryCode,
			ContactNumber:        xbResp.SenderData.ContactNumber,
			Dob:                  xbResp.SenderData.Dob,
			SourceOfIncome:       xbResp.SenderData.SourceOfIncome,
		},
	}, nil
}

// handleConfirmationError handles the error scenario when payout confirmation fails.
// It performs the following operations in sequence:
// 1. Updates the disbursement status reason to failed
// 2. Records a status history entry with the provided actor (defaults to system user if not provided)
// 3. Attempts to find the associated ledger entry by disbursement UUID
// 4. Updates the ledger status if it exists
// 5. Sends a callback notification to the client
//
// Returns:
//   - error: returns a wrapped database error if any of the critical operations (update reason, find ledger, or update ledger status) fail
//   - Note: status history update failures are logged but do not cause the function to return an error
func (s *xbPayoutService) handleConfirmationError(ctx context.Context, request *xbModel.ConfirmPayoutRequest, disbursement *disbursementModel.DisbursementWithTransaction) error {
	err := s.disbursementRepo.UpdateReasonByIDs(ctx, []string{request.PayoutId}, constant.XbDisbursementReasonTypeError, constant.XbDisbursementReasonDescError)
	if err != nil {
		s.logger.Error(
			ctx,
			"Confirm - Failed to update disbursement reason to failed",
			logger.String("payoutId", request.PayoutId),
			logger.Error(err),
		)
		return pkgErrors.New(response.HttpErrDatabase, err)
	}

	actor := request.ApprovedBy
	if actor == "" {
		actor = constant.UserSystemType
	}

	errStatusHistory := s.RecordStatusHistory(ctx, &statusHistoryModel.RecordDisbursementStatusHistoryRequest{
		DisbursementID: request.PayoutId,
		Status:         constant.XbStatusHttpError,
		Actor:          actor,
	})
	if errStatusHistory != nil {
		s.logger.Error(ctx, "Confirm - Failed to update status history", logger.String("payoutId", request.PayoutId), logger.Error(errStatusHistory))
	}

	ledger, err := s.orchestratorSvc.FindByReference(ctx, disbursement.UUID, orchestratorModel.TypeDisbursement)
	if err != nil {
		s.logger.Error(
			ctx,
			"Confirm - Failed to find ledger",
			logger.String("payoutId", request.PayoutId),
			logger.String("disbursementId", disbursement.UUID),
			logger.Error(err),
		)
		return pkgErrors.New(response.HttpErrDatabase, err)
	}

	// update ledger when its exist
	if ledger != nil {
		_, reasonType, status := constant.MapXbProcessorStatusToCoreStatus(constant.XbStatusHttpError)
		err = s.orchestratorSvc.UpdateStatusAccountTransactionByReferenceID(ctx, disbursement.UUID, status, &reasonType, nil)
		if err != nil {
			s.logger.Error(
				ctx,
				"Confirm - Failed to update ledger status",
				logger.String("payoutId", request.PayoutId),
				logger.String("disbursementId", disbursement.UUID),
				logger.String("reasonType", reasonType),
				logger.Error(err),
			)
			return pkgErrors.New(response.HttpErrDatabase, err)
		}
	}

	// s.sendCallbackToClient(ctx, disbursement.UUID) // RnD will decided later whether to send callback in case of confirmation error, since currently client only receive callback when payout is successful
	return nil
}
