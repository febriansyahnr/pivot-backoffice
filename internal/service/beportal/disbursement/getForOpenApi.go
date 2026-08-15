package disbursementService

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/snap/bankTransfer"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *DisbursementService) GetBulkDisbursementForOpenApiByID(ctx context.Context, filter *disbursementModel.GetDisbursementFilterRequest, page, perPage int64) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/GetBulkDisbursementForOpenApiByID")
	defer segment.End()

	var (
		isSimulation  = false
		bankDB        = bankTransfer.NewBankDB()
		payoutObjects []disbursementModel.PayoutObject
		errSimulation error
	)

	// get bulk disbursement first
	bulkDisbursement, err := s.disbursementRepo.FindBulkDisbursementByID(ctx, filter.BulkID)
	if err != nil {
		return nil, err
	} else if bulkDisbursement == nil {
		return nil, pkgErrors.New(response.HttpErrNotFound, constant.ErrBulkDisbursementNotFound)
	}

	// validate merchant
	if bulkDisbursement.MerchantID != filter.MerchantID {
		s.logger.Error(ctx, "merchant id not match", logger.Error(fmt.Errorf("bulk disbursement not found, merchant id not match")))
		return nil, pkgErrors.New(response.HttpErrNotFound, constant.ErrBulkDisbursementNotFound)
	}

	// get all disbursement by bulkID
	disbursements, err := s.disbursementRepo.GetList(ctx, filter, page, perPage)
	if err != nil {
		return nil, err
	}

	disbursementData, ok := disbursements.Data.([]*disbursementModel.DisbursementWithTransactionResponse)
	if !ok {
		s.logger.Error(ctx, "mismatch disbursement model", logger.Error(fmt.Errorf("mismatch disbursement model")))
		return nil, pkgErrors.New(response.HttpErrRequest, constant.ErrDisbursementNotFound)
	}

	// build payout response
	for _, disbursement := range disbursementData {
		bank := bankDB.FindByCode(disbursement.BeneficiaryBankCode)

		remark := ""
		if disbursement.Remark != nil {
			remark = *disbursement.Remark
		}

		if s.config.Environment != constant.EnvironmentProduction {
			isSimulation, errSimulation = util.ValidateMagicNumber(http.MethodGet, disbursement.BeneficiaryAccountNo)
		}

		accountInquiryId := ""
		if disbursement.AccountInquiryID != nil {
			accountInquiryId = *disbursement.AccountInquiryID
		}

		payoutObject := disbursementModel.PayoutObject{
			ReferenceID: disbursement.ReferenceID,
			InquiryID:   accountInquiryId,
			ChannelCode: bank.ChannelCode,
			ChannelInformation: disbursementModel.PayoutChannelInformation{
				AccountNumber: disbursement.BeneficiaryAccountNo,
				AccountName:   disbursement.BeneficiaryAccountName,
			},
			Amount: commonModel.Amount{
				Currency: disbursement.Currency,
				Value:    disbursement.Amount.String(),
			},
			Description: remark,
			Status:      BuildChildStatus(disbursement.DisbursementWithTransaction),
			Reason:      BuildReason(disbursement.DisbursementWithTransaction),
			CreatedAt:   disbursement.CreatedAt,
			UpdatedAt:   disbursement.UpdatedAt,
		}
		if isSimulation {
			payoutObject.Status = constant.DisbursementStatusApproved
			if errSimulation != nil && errSimulation.Error() == constant.ErrInsufficientBalance.Error() {
				payoutObject.Status = constant.StatusPending
			}
		}

		payoutObjects = append(payoutObjects, payoutObject)
	}

	bulkStatus := bulkDisbursement.Status
	if isSimulation {
		bulkStatus = constant.BulkDisbursementStatusPending
	}

	// build response
	disbursementResponse := &disbursementModel.GetBulkDisbursementForOpenApiByIDResponse{
		UUID:          bulkDisbursement.UUID,
		MerchantID:    bulkDisbursement.MerchantID,
		PayoutResults: s.getDisbursementSummaryByBulkID(ctx, bulkDisbursement.UUID),
		Payouts:       payoutObjects,
		Status:        bulkStatus,
		CreatedAt:     bulkDisbursement.CreatedAt,
		UpdatedAt:     bulkDisbursement.UpdatedAt,
	}

	return &commonModel.PaginationResponse{
		Data: disbursementResponse,
		Meta: disbursements.Meta,
	}, nil
}

func (s *DisbursementService) GetBulkDisbursementForOpenApiByReferenceID(ctx context.Context, bulkID, referenceID, merchantID string) (*disbursementModel.GetBulkDisbursementForOpenApiByReferenceIDResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/GetBulkDisbursementForOpenApiByReferenceID")
	defer segment.End()

	var bankDB = bankTransfer.NewBankDB()

	// get bulk disbursement first
	bulkDisbursement, err := s.disbursementRepo.FindBulkDisbursementByID(ctx, bulkID)
	if err != nil {
		return nil, err
	}

	if bulkDisbursement == nil {
		return nil, pkgErrors.New(response.HttpErrNotFound, constant.ErrBulkDisbursementNotFound)
	}

	// validate merchant
	if bulkDisbursement.MerchantID != merchantID {
		s.logger.Error(ctx, "merchant id not match", logger.Error(fmt.Errorf("bulk disbursement not found, merchant id not match")))
		return nil, pkgErrors.New(response.HttpErrNotFound, constant.ErrBulkDisbursementNotFound)
	}

	// if disbursement by reference not found, then return err
	disbursement, err := s.disbursementRepo.FindByMerchantAndReference(ctx, merchantID, referenceID)
	if err != nil {
		return nil, err
	}

	if disbursement == nil {
		return nil, pkgErrors.New(response.HttpErrNotFound, constant.ErrBulkDisbursementNotFound)
	}

	// validate disbursement and bulkID
	disbursementBulkID := ""
	if disbursement.BulkID != nil {
		disbursementBulkID = *disbursement.BulkID
	}

	if disbursementBulkID != bulkID {
		s.logger.Error(ctx, "mismatch disbursement bulkID", logger.Error(fmt.Errorf("mismatch disbursement bulkID")))
		return nil, pkgErrors.New(response.HttpErrNotFound, constant.ErrBulkDisbursementNotFound)
	}

	bank := bankDB.FindByCode(disbursement.BeneficiaryBankCode)
	remark := ""
	if disbursement.Remark != nil {
		remark = *disbursement.Remark
	}

	accountInquiryId := ""
	if disbursement.AccountInquiryID != nil {
		accountInquiryId = *disbursement.AccountInquiryID
	}

	// build payoutObject
	payout := disbursementModel.PayoutObject{
		ReferenceID: disbursement.ReferenceID,
		InquiryID:   accountInquiryId,
		ChannelCode: bank.ChannelCode,
		ChannelInformation: disbursementModel.PayoutChannelInformation{
			AccountNumber: disbursement.BeneficiaryAccountNo,
			AccountName:   disbursement.BeneficiaryAccountName,
		},
		Amount: commonModel.Amount{
			Currency: disbursement.Currency,
			Value:    disbursement.Amount.String(),
		},
		Description: remark,
		Status:      BuildChildStatus(*disbursement),
		Reason:      BuildReason(*disbursement),
		CreatedAt:   disbursement.CreatedAt,
		UpdatedAt:   disbursement.UpdatedAt,
	}

	// build payoutResults
	payoutResults := disbursementModel.PayoutResultObject{}
	payoutAmount, _ := strconv.ParseFloat(payout.Amount.Value, 64)
	if payout.Status == constant.StatusSuccess {
		payoutResults.TotalSuccessCount = 1
		payoutResults.TotalSuccessAmount = payoutAmount
	} else if payout.Status == constant.StatusFailed {
		payoutResults.TotalFailedCount = 1
		payoutResults.TotalFailedAmount = payoutAmount
	} else {
		payoutResults.TotalPendingCount = 1
		payoutResults.TotalPendingAmount = payoutAmount
	}

	disbursementResponse := &disbursementModel.GetBulkDisbursementForOpenApiByReferenceIDResponse{
		UUID:          bulkDisbursement.UUID,
		MerchantID:    bulkDisbursement.MerchantID,
		PayoutResults: payoutResults,
		Payouts:       payout,
	}

	return disbursementResponse, nil
}

func (s *DisbursementService) getDisbursementSummaryByBulkID(ctx context.Context, bulkID string) disbursementModel.PayoutResultObject {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/getDisbursementSummaryByBulkID")
	defer segment.End()

	// build payout result
	var (
		payoutResultObject = disbursementModel.PayoutResultObject{}
		errG               = new(errgroup.Group)
	)

	// get summary success
	errG.Go(func() error {
		summary := s.disbursementRepo.SummarySuccessByBulkID(ctx, bulkID)
		payoutResultObject.TotalSuccessCount = summary.Count
		payoutResultObject.TotalSuccessAmount = summary.Sum

		return nil
	})

	// get summary failed
	errG.Go(func() error {
		summary := s.disbursementRepo.SummaryFailedByBulkID(ctx, bulkID)
		payoutResultObject.TotalFailedCount = summary.Count
		payoutResultObject.TotalFailedAmount = summary.Sum

		return nil
	})

	// get summary pending
	errG.Go(func() error {
		summary := s.disbursementRepo.SummaryPendingByBulkID(ctx, bulkID)
		payoutResultObject.TotalPendingCount = summary.Count
		payoutResultObject.TotalPendingAmount = summary.Sum

		return nil
	})

	// get summary cancelled
	errG.Go(func() error {
		summary := s.disbursementRepo.SummaryCancelledByBulkID(ctx, bulkID)
		payoutResultObject.TotalCancelledCount = summary.Count
		payoutResultObject.TotalCancelledAmount = summary.Sum

		return nil
	})

	// Wait for all goroutines to complete and handle any errors
	if err := errG.Wait(); err != nil {
		// Log error but continue as this is for data enrichment
		s.logger.Error(ctx, "error in data enrichment goroutines", logger.Error(err))
	}

	return payoutResultObject
}

func BuildReason(disbursement disbursementModel.DisbursementWithTransaction) string {
	var (
		reasonType            = ""
		reason                = ""
		transactionReasonType = ""
	)
	if util.ValueOfPtr(disbursement.TransactionStatus) != constant.StatusSuccess && disbursement.ReasonType != nil {
		reasonType = *disbursement.ReasonType
	}

	if util.ValueOfPtr(disbursement.TransactionStatus) != constant.StatusSuccess && disbursement.TransactionReasonType != nil {
		transactionReasonType = *disbursement.TransactionReasonType
	}

	if util.ValueOfPtr(disbursement.TransactionStatus) == constant.StatusFailed { // Define default Value
		reason = "Failed to process by Bank Network"
	}
	if slices.Contains([]string{constant.DisbursementReasonTypeInsufficientBalance, constant.DisbursementReasonTypeCancelled}, reasonType) &&
		disbursement.ReasonDescription != nil {
		reason = *disbursement.ReasonDescription
	} else if slices.Contains(
		[]string{constant.ReasonTypeBeneficiaryAccountReason}, transactionReasonType) &&
		(strings.Contains(*disbursement.TransactionReasonDescription, "invalid account") ||
			strings.Contains(*disbursement.TransactionReasonDescription, "inactive account")) {
		reason = "Invalid Account"
	} else if slices.Contains(
		[]string{constant.ReasonTypeBeneficiaryAccountReason}, transactionReasonType) &&
		strings.Contains(*disbursement.TransactionReasonDescription, "dormant account") {
		reason = "Dormant Account"
	} else if slices.Contains(
		[]string{constant.ReasonTypeBlockedByHarsya}, transactionReasonType) && disbursement.TransactionReasonDescription != nil {
		reason = "Feature not allowed at this time"
	} else if slices.Contains(
		[]string{constant.ReasonTypeOtherReason}, transactionReasonType) && disbursement.TransactionReasonDescription == nil {
		reason = "Unknown Bank Network error"
	} else if slices.Contains(
		[]string{constant.ReasonTypeOtherReason}, transactionReasonType) &&
		(strings.Contains(strings.ToLower(*disbursement.TransactionReasonDescription), "record not found") ||
			strings.Contains(strings.ToLower(*disbursement.TransactionReasonDescription), "transaction not found") ||
			strings.Contains(strings.ToLower(*disbursement.TransactionReasonDescription), "transaction failed") ||
			strings.Contains(strings.ToLower(*disbursement.TransactionReasonDescription), "bounceback")) {
		reason = "Failed to process by Bank Network"
	} else if slices.Contains(
		[]string{constant.ReasonTypeOtherReason}, transactionReasonType) &&
		strings.Contains(strings.ToLower(*disbursement.TransactionReasonDescription), strings.ToLower("Manual update from PENDING to FAILED")) {
		reason = "Failed to process by Bank Network"
	} else if slices.Contains([]string{
		constant.ReasonTypeBeneficiaryAccountReason,
		constant.ReasonTypeDeclinedBeneficiaryRestriction,
	}, transactionReasonType) && disbursement.TransactionReasonDescription != nil {
		reason = *disbursement.TransactionReasonDescription
	} else if slices.Contains(
		[]string{constant.ReasonTypeBankNetworkError}, transactionReasonType) && disbursement.TransactionReasonDescription != nil {
		reason = "Failed to process by Bank Network"
	} else if transactionReasonType == constant.ReasonTypePayoutDelayed && disbursement.TransactionReasonDescription != nil {
		reason = *disbursement.TransactionReasonDescription
	}

	return reason
}

func BuildChildStatus(disbursement disbursementModel.DisbursementWithTransaction) string {
	status := disbursement.Status
	if util.ValueOfPtr(disbursement.ReasonType) == constant.DisbursementReasonTypeCancelled {
		status = constant.DisbursementReasonTypeCancelled
	}
	if disbursement.TransactionStatus != nil {
		status = *disbursement.TransactionStatus
	}
	if status != constant.StatusSuccess && util.ValueOfPtr(disbursement.TransactionReasonType) == constant.DisbursementReasonTypeDelayed {
		status = constant.DisbursementReasonTypeDelayed
	}

	return status
}

func (s *DisbursementService) GetDisbursementByReferenceID(ctx context.Context, referenceID, merchantID string) (*disbursementModel.PayoutObject, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/GetDisbursementByReferenceID")
	defer segment.End()

	var bankDB = bankTransfer.NewBankDB()

	// Find the disbursement by merchant and reference
	disbursement, err := s.disbursementRepo.FindByMerchantAndReference(ctx, merchantID, referenceID)
	if err != nil {
		return nil, err
	}

	if disbursement == nil {
		return nil, pkgErrors.New(response.HttpErrNotFound, constant.ErrDisbursementNotFound)
	}

	bank := bankDB.FindByCode(disbursement.BeneficiaryBankCode)
	remark := ""
	if disbursement.Remark != nil {
		remark = *disbursement.Remark
	}

	status := disbursement.Status
	if disbursement.TransactionStatus != nil {
		status = *disbursement.TransactionStatus
	}

	accountInquiryId := ""
	if disbursement.AccountInquiryID != nil {
		accountInquiryId = *disbursement.AccountInquiryID
	}

	// Check if it's a simulation in non-production environments
	var isSimulation bool
	if s.config.Environment != constant.EnvironmentProduction {
		isSimulation, _ = util.ValidateMagicNumber(http.MethodGet, disbursement.BeneficiaryAccountNo)
	}

	// Build the payout object response
	payout := &disbursementModel.PayoutObject{
		ReferenceID: disbursement.ReferenceID,
		InquiryID:   accountInquiryId,
		ChannelCode: bank.ChannelCode,
		ChannelInformation: disbursementModel.PayoutChannelInformation{
			AccountNumber: disbursement.BeneficiaryAccountNo,
			AccountName:   disbursement.BeneficiaryAccountName,
		},
		Amount: commonModel.Amount{
			Currency: disbursement.Currency,
			Value:    disbursement.Amount.String(),
		},
		Description: remark,
		Status:      status,
		Reason:      BuildReason(*disbursement),
		CreatedAt:   disbursement.CreatedAt,
		UpdatedAt:   disbursement.UpdatedAt,
	}

	if isSimulation {
		payout.Status = constant.DisbursementStatusApproved
	}

	return payout, nil
}
