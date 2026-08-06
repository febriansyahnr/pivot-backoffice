package withdrawalService

import (
	"context"
	"errors"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

// Tech-Debt: The inquiry transaction status feature needs to be updated to exclude balance transfer transactions and to include the balance source derived from withdrawal transactions.
func (s *withdrawalService) InquiryTransaction(ctx context.Context, request *withdrawal.InquiryTransactionRequest) (*withdrawal.InquiryTransactionResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/withdrawal/InquiryTransaction")
	defer segment.End()

	traceId, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)

	details, err := s.GetById(ctx, &withdrawal.WithdrawalDetailRequest{Id: request.Id, MerchantId: request.MerchantId})
	if err != nil {
		return nil, err

	} else if details.Status != constant.StatusPending {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrTransactionAlreadyInFinalStatus)

	} else if details.ExternalID == "" {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("external id for bank transfer not found"))
	}

	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		RequestId:   traceId,
		From:        "Withdrawal-Inquiry-Transaction",
		OriginId:    request.Id,
		ReferenceId: request.MerchantId,
	})
	bankTransfer, err := s.snapCoreRepo.FindBankTransferByExternalID(ctx, details.ExternalID, false)
	if err != nil {
		s.logger.Error(ctx, "Find bank transfer by external id", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrInternal, err)
	}

	status, reasonType, reasonDesc := bankTransfer.MappingInquiryTransactionStatus()

	if status == constant.StatusSuccess {
		metadata := &withdrawal.Metadata{
			BankTransfer: &withdrawal.BankTransfer{
				UUID:               details.BankTransferUUID,
				ExternalId:         details.ExternalID,
				BankReferenceNo:    details.BankReferenceNo,
				PartnerReferenceNo: bankTransfer.PartnerReferenceNo,
			},
		}
		if bankTransfer.BankReferenceNo != "" {
			metadata.BankTransfer.BankReferenceNo = bankTransfer.BankReferenceNo
		}
		if err = s.repo.UpdateMetadataById(ctx, request.Id, metadata); err != nil {
			s.logger.Error(ctx, "Update withdrawal metadata by id", logger.Error(err))
			return nil, pkgErrs.New(response.HttpErrDatabase, err)
		}
	}

	if err = s.orchestratorSvc.UpdateStatusAccountTransaction(ctx, details.TransactionID, status, &reasonType, &reasonDesc); err != nil {
		s.logger.Error(ctx, "Update status account transaction", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}

	return &withdrawal.InquiryTransactionResponse{
		WithdrawalDetailResponse: details,
		UpdatedAt:                time.Now().UTC(),
		Status:                   status,
		ReasonType:               reasonType,
		ReasonDescription:        reasonDesc,
	}, nil
}
