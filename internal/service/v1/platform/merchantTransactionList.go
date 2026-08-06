package platformService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	merchantTopUp "github.com/paper-indonesia/pivot-backoffice/internal/model/merchantTopUp"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/platform"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/transfer"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	errPkg "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *PlatformService) GetMerchantTransactionList(ctx context.Context, request *platform.TransactionRequest) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "/internal/service/v1/platform/GetMerchantTransactionList")
	defer segment.End()

	err := s.merchantService.ValidateSubMerchantParent(ctx, request.ParentMerchantId, request.MerchantId)
	if err != nil {
		return nil, err
	}

	err = request.Validate()
	if err != nil {
		return nil, errPkg.New(response.HttpErrRequest, err)
	}

	if request.Reference == constant.ReferenceDisbursement {
		filterDisbursementRequest := &disbursementModel.GetDisbursementFilterRequest{
			MerchantID:        request.MerchantId,
			StartCreatedAt:    &request.StartDate,
			EndCreatedAt:      &request.EndDate,
			Status:            request.ApprovalStatus,
			TransactionStatus: request.Status,
			Type:              request.ReferenceType,
			Keyword:           request.Keyword,
			Sort:              request.SortOrder,
			SortBy:            request.SortBy,
		}
		data, err := s.disbursementService.GetList(ctx, filterDisbursementRequest, request.Page, request.PerPage)
		if err != nil {
			return nil, errPkg.New(response.HttpErrInternal, constant.ErrGetDisbursementList)
		}
		return data, nil
	}

	if request.Reference == constant.ReferencePayment {
		filterPaymentRequest := paymentModel.FilterPaymentHistoryOption{
			MerchantID:       request.MerchantId,
			ReferenceID:      request.ReferenceID,
			Status:           request.Status,
			PaymentMethod:    request.PaymentMethod,
			StartDate:        request.StartDate,
			EndDate:          request.EndDate,
			PaymentStartDate: request.PaymentStartDate,
			PaymentEndDate:   request.PaymentEndDate,
			Sort:             request.SortOrder,
			SortBy:           request.SortBy,
			Page:             int(request.Page),
			PerPage:          int(request.PerPage),
		}
		data, err := s.paymentService.FilterPaymentHistory(ctx, filterPaymentRequest)
		if err != nil {
			return nil, errPkg.New(response.HttpErrInternal, constant.ErrGetPayments)
		}
		return data, nil
	}

	if request.Reference == constant.ReferencePlatformTransfer {
		filterTransferRequest := transfer.GetTransferListRequest{
			UUID:        request.UUID,
			MerchantID:  request.MerchantId,
			ReferenceID: request.ReferenceID,
			Type:        request.ReferenceType,
			Status:      request.Status,
			StartDate:   request.StartDate,
			EndDate:     request.EndDate,
			SortBy:      request.SortBy,
			SortOrder:   request.SortOrder,
		}
		data, err := s.transferService.GetList(ctx, &filterTransferRequest, request.Page, request.PerPage)
		if err != nil {
			return nil, errPkg.New(response.HttpErrInternal, constant.ErrGetTransferList)
		}
		return data, nil
	}

	if request.Reference == constant.ReferenceWithdrawal {
		filterWithdrawalRequest := &withdrawal.WithdrawalHistoryRequest{
			WithdrawalListRequest: &withdrawal.WithdrawalListRequest{
				MerchantId:   request.MerchantId,
				AccountName:  constant.AccountNamePayment, // Fixed: always PAYMENT for withdrawal list
				StrStartDate: request.StartDate.Format("2006-01-02"),
				StrEndDate:   request.EndDate.Format("2006-01-02"),
				StartDate:    request.StartDate,
				EndDate:      request.EndDate,
				Status:       request.Status,
				Id:           request.UUID,
				Sort:         request.SortOrder,
			},
			Page:    int(request.Page),
			PerPage: int(request.PerPage),
		}
		data, err := s.withdrawalService.GetList(ctx, filterWithdrawalRequest)
		if err != nil {
			return nil, errPkg.New(response.HttpErrInternal, constant.ErrGetWithdrawalList)
		}
		return data, nil
	}

	if request.Reference == constant.ReferenceTopUp {
		filterTopUpRequest := &merchantTopUp.TopUpTransactionListRequest{
			MerchantId:    request.MerchantId,
			StartDate:     request.StartDate,
			EndDate:       request.EndDate,
			Status:        request.Status,
			TransactionID: request.UUID,
			ReferenceID:   request.ReferenceID,
			SortOrder:     request.SortOrder,
			Page:          request.Page,
			PerPage:       request.PerPage,
		}
		data, err := s.merchantTopUpService.GetList(ctx, filterTopUpRequest)
		if err != nil {
			return nil, errPkg.New(response.HttpErrInternal, constant.ErrGetTopUpList)
		}
		return data, nil
	}

	if request.Reference == constant.ReferenceCharge {
		changeFilter := unifiedPaymentModel.FilterChargeRequest{
			MerchantID:        request.MerchantId,
			Status:            request.Status,
			ClientReferenceID: request.ReferenceID,
			StartCreatedAt:    request.StartDate,
			EndCreatedAt:      request.EndDate,
			StartPaymentDate:  request.PaymentStartDate,
			EndPaymentDate:    request.PaymentEndDate,
			Sort:              request.SortOrder,
			SortBy:            request.SortBy,
			Page:              int(request.Page),
			PerPage:           int(request.PerPage),
		}
		data, err := s.unifiedPaymentService.GetChargeList(ctx, &changeFilter)
		if err != nil {
			return nil, errPkg.New(response.HttpErrInternal, constant.ErrGetChargeList)
		}
		return data, nil
	}

	return nil, nil
}
