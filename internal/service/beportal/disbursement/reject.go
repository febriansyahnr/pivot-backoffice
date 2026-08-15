package disbursementService

import (
	"context"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/snap/bankTransfer"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *DisbursementService) Reject(ctx context.Context, request *disbursementModel.RejectRequest) (result string, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/Approve")
	defer segment.End()

	var (
		rejectedRow []*disbursementModel.BulkPreviewResponse
		bankDB      = bankTransfer.NewBankDB()
	)

	// For reject action we can't do bulk update, because we need to set the reason 1 by 1
	// But we still need to validate merchant first

	var disbursementIDs []string
	for _, rejectAction := range request.RejectAction {
		disbursementIDs = append(disbursementIDs, rejectAction.DisbursementID)
	}

	if len(disbursementIDs) == 0 {
		return "", nil
	}

	// Validate merchant
	countByMerchant := s.disbursementRepo.CountByIDsAndMerchantID(ctx, disbursementIDs, request.MerchantID)
	if countByMerchant != len(disbursementIDs) {
		err := constant.ErrMerchantNotAllowedPerformAction
		s.logger.Error(ctx, err.Error(), logger.Error(err))
		return "", pkgErrors.New(httpResponse.HttpErrRequest, err)
	}

	// Get disbursement list by UUID
	rejectedDisbursements, err := s.disbursementRepo.GetByIDs(ctx, disbursementIDs)
	if err != nil {
		return "", err
	}

	// Begin Tx
	ctxTx, err := s.disbursementRepo.BeginTransaction(ctx)
	if err != nil {
		return "", err
	}
	isCompleted := false
	defer func() {
		if !isCompleted {
			if e := s.disbursementRepo.RollbackTransaction(ctxTx); e != nil {
				result, err = "", fmt.Errorf("failed to rollback transaction: %v", e)
			}
		}
	}()

	// update reject 1 by 1
	for _, rejectAction := range request.RejectAction {
		err = s.disbursementRepo.Reject(
			ctxTx,
			rejectAction.DisbursementID,
			rejectAction.ReasonType,
			rejectAction.ReasonDescription,
			request.RejectedBy,
		)
		if err != nil {
			return "", err
		}

		// Record status history for rejected disbursement
		s.recordDisbursementRejected(ctx, rejectAction.DisbursementID, request.RejectedBy)

		// rejected file
		rejectedDisbursement := findDisbursementByID(rejectedDisbursements, rejectAction.DisbursementID)
		beneficiaryBankName := ""
		if rejectedDisbursement.BeneficiaryBankName != nil {
			beneficiaryBankName = *rejectedDisbursement.BeneficiaryBankName
		}

		remark := ""
		if rejectedDisbursement.Remark != nil {
			remark = *rejectedDisbursement.Remark
		}

		channelCode := ""
		bank := bankDB.FindByCode(rejectedDisbursement.BeneficiaryBankCode)
		if bank != nil {
			channelCode = bank.ChannelCode
		}

		rejectedRow = append(rejectedRow, &disbursementModel.BulkPreviewResponse{
			ReferenceID:            rejectedDisbursement.ReferenceID,
			BeneficiaryBankCode:    rejectedDisbursement.BeneficiaryBankCode,
			BeneficiaryBankName:    beneficiaryBankName,
			BeneficiaryAccountNo:   rejectedDisbursement.BeneficiaryAccountNo,
			BeneficiaryAccountName: rejectedDisbursement.BeneficiaryAccountName,
			Amount:                 rejectedDisbursement.Amount.String(),
			Remark:                 remark,
			Error:                  fmt.Sprintf("Reject reason : %s", rejectAction.ReasonDescription),
			Result:                 constant.BulkPreviewResultRejected,
			ChannelCode:            channelCode,
		})
	}

	if err = s.disbursementRepo.CommitTransaction(ctxTx); err != nil {
		return "", err
	}
	isCompleted = true

	// Generate and update fileRejected for bulk transaction
	if request.BulkID != "" {
		return s.GenerateExcelAndUpdateRejectedBulkDisbursement(ctx, request.BulkID, rejectedRow)
	}
	return "", nil
}

// FindPersonByName finds a person by name in a slice of persons
func findDisbursementByID(disbursements []*disbursementModel.Disbursement, uuid string) *disbursementModel.Disbursement {
	for _, disbursement := range disbursements {
		if disbursement.UUID == uuid {
			return disbursement
		}
	}
	return nil
}
