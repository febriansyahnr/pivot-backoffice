package disbursementService

import (
	"context"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/bankTransfer"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *DisbursementService) GetPayoutStatusAndRouting(ctx context.Context, request *disbursementModel.CRMSinglePayoutStatusRequest) (*disbursementModel.CRMPayoutStatusResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/GetPayoutStatusAndRouting")
	defer segment.End()

	return s.processSinglePayoutStatus(ctx, request.ReferenceID)
}

func (s *DisbursementService) findDisbursementByReferenceID(ctx context.Context, referenceID string) (*disbursementModel.DisbursementWithTransaction, error) {
	// Try to find by disbursement UUID first
	disbursement, err := s.disbursementRepo.FindByID(ctx, referenceID)
	if err != nil {
		s.logger.Debug(ctx, "disbursement not found by ID", logger.String("referenceID", referenceID), logger.Error(err))
	} else if disbursement != nil {
		return disbursement, nil
	}

	// If not found by ID, try to find by reference ID
	disbursement, err = s.disbursementRepo.FindByReference(ctx, referenceID)
	if err != nil {
		s.logger.Debug(ctx, "disbursement not found by reference", logger.String("referenceID", referenceID), logger.Error(err))
		return nil, fmt.Errorf("transaction not found")
	}

	if disbursement == nil {
		s.logger.Debug(ctx, "disbursement found but data is null", logger.String("referenceID", referenceID))
		return nil, fmt.Errorf("transaction not found")
	}

	return disbursement, nil
}

func (s *DisbursementService) buildTransferLogsFromStatusData(statusData *snapCoreModel.BankTransferCheckStatusResponseData) []disbursementModel.RoutingHistoryItem {
	transferLogs := make([]disbursementModel.RoutingHistoryItem, 0)

	for i, log := range statusData.TransferLogs {
		responseMsg := ""

		if log.Status != "SUCCESS" {
			if log.ResponsePayload.ResponseMessage != "" {
				responseMsg = log.ResponsePayload.ResponseMessage
			} else if log.ResponsePayload.ResponseCode != "" {
				responseMsg = fmt.Sprintf("Response Code: %s", log.ResponsePayload.ResponseCode)
			} else {
				responseMsg = s.extractFailedReasonFromAdditionalInfo(log.AdditionalInfo)
			}
		}

		transferLogs = append(transferLogs, disbursementModel.RoutingHistoryItem{
			Order:       i + 1,
			BankName:    log.Bank,
			Status:      log.Status,
			ResponseMsg: responseMsg,
			Timestamp:   log.CreatedAt,
		})
	}

	return transferLogs
}

func (s *DisbursementService) extractFailedReasonFromAdditionalInfo(additionalInfo interface{}) string {
	if additionalInfo == nil {
		return ""
	}

	if additionalMap, ok := additionalInfo.(map[string]interface{}); ok {
		if failedReason, exists := additionalMap["failedReason"]; exists {
			if failedReasonMap, ok := failedReason.(map[string]interface{}); ok {
				if data, exists := failedReasonMap["data"]; exists {
					if dataMap, ok := data.(map[string]interface{}); ok {
						if responseMessage, exists := dataMap["responseMessage"]; exists {
							if msg, ok := responseMessage.(string); ok {
								return msg
							}
						}
					}
				}
			}
		}
	}

	return ""
}

func (s *DisbursementService) GetBatchPayoutStatusAndRouting(ctx context.Context, request *disbursementModel.CRMBatchPayoutStatusRequest) (*disbursementModel.CRMBatchPayoutStatusResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/GetBatchPayoutStatusAndRouting")
	defer segment.End()

	batchResponse := &disbursementModel.CRMBatchPayoutStatusResponse{
		Code: "00",
		Data: make([]disbursementModel.CRMPayoutStatusResult, 0, len(request.ReferenceIDs)),
	}

	for _, refID := range request.ReferenceIDs {
		result := disbursementModel.CRMPayoutStatusResult{
			ReferenceID: refID,
			Success:     false,
		}

		response, err := s.processSinglePayoutStatus(ctx, refID)
		if err != nil {
			result.Error = &disbursementModel.CRMPayoutStatusError{
				Code:    "ERROR",
				Message: err.Error(),
			}
		} else {
			result.Success = true
			result.Data = response.Data
		}

		batchResponse.Data = append(batchResponse.Data, result)
	}

	return batchResponse, nil
}

func (s *DisbursementService) processSinglePayoutStatus(ctx context.Context, referenceID string) (*disbursementModel.CRMPayoutStatusResponse, error) {
	disbursement, err := s.findDisbursementByReferenceID(ctx, referenceID)
	if err != nil {
		s.logger.Error(ctx, "error finding disbursement", logger.Error(err))
		return nil, pkgErrors.New(httpResponse.HttpErrDatabase, err)
	}

	if disbursement == nil {
		return nil, pkgErrors.New(httpResponse.HttpErrNotFound, fmt.Errorf("transaction not found"))
	}

	beneficiaryBankName := ""
	if disbursement.BeneficiaryBankName != nil {
		beneficiaryBankName = *disbursement.BeneficiaryBankName
	}

	response := &disbursementModel.CRMPayoutStatusResponse{
		Code: "00",
		Data: &disbursementModel.CRMPayoutStatusResponseData{
			DisbursementUUID:   disbursement.UUID,
			ReferenceID:        disbursement.ReferenceID,
			ApprovalStatus:     disbursement.Status,
			Amount:             disbursement.Amount.String(),
			BeneficiaryAccount: disbursement.BeneficiaryAccountNo,
			BeneficiaryName:    disbursement.BeneficiaryAccountName,
			BeneficiaryBank:    beneficiaryBankName,
			TransactionDate:    disbursement.CreatedAt.Format(time.RFC3339),
			CreatedAt:          disbursement.CreatedAt.Format(time.RFC3339),
			UpdatedAt:          disbursement.UpdatedAt.Format(time.RFC3339),
			TransferLogs:       make([]disbursementModel.RoutingHistoryItem, 0),
		},
	}

	if disbursement.ProcessorReferenceID != nil && *disbursement.ProcessorReferenceID != "" {
		accountTxn, err := s.accountTransactionRepo.FindByReference(ctx, disbursement.UUID, constant.TypeDisbursement)
		response.Data.Status = accountTxn.Status
		if err == nil && accountTxn != nil {
			statusData, err := s.snapCoreRepo.CheckStatusByExternalId(ctx, accountTxn.UUID.String(), false)
			if err != nil {
				s.logger.Warn(ctx, "error getting transfer status", logger.Error(err))
				response.Data.TransferLogs = []disbursementModel.RoutingHistoryItem{
					{
						Order:       1,
						BankName:    beneficiaryBankName,
						Status:      disbursement.Status,
						ResponseMsg: "Transaction processed through " + beneficiaryBankName,
						Timestamp:   disbursement.CreatedAt.Format(time.RFC3339),
					},
				}
			} else {
				response.Data.TransferLogs = s.buildTransferLogsFromStatusData(statusData)
			}
		}
	}

	return response, nil
}
