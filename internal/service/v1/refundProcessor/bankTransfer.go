package refundProcessorService

import (
	"context"
	"fmt"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	beneficiaryAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/beneficiaryAccount"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	refundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/refund"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/bankTransfer"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/snap/bankTransfer"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *BankTransferStrategy) Process(ctx context.Context, request *refundModel.RefundProcessRequest) error {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/refundProcessor/BankTransferStrategyProcess")
	defer span.End()

	// Update processor reference
	processorReferenceID, reconReferenceNo := "", ""
	defer func() {
		if e := s.orchestratorSvc.UpdateProcessorAndReconReferenceByID(ctx, request.RefundLedgerID, constant.SnapCoreProcessor, processorReferenceID, reconReferenceNo); e != nil {
			s.logger.Error(ctx, "[BankTransferStrategyProcess] Update account transactions additional info", logger.Error(e))
		}
	}()

	// Validate channel information
	if request.MetadataObj.TransferDestination == nil {
		return pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrRefundTransferDestinationNotAvailable)
	}

	bankDb := bankTransfer.NewBankDB()
	bank := bankDb.FindByChannelCode(request.MetadataObj.TransferDestination.ChannelCode)
	if bank == nil {
		return pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrBankTransferChannelNotFound)
	}

	beneficiaryBankCode := bank.Code
	beneficiaryAccountNo := request.MetadataObj.TransferDestination.ChannelInformation.AccountNumber
	beneficiaryAccountName := request.MetadataObj.TransferDestination.ChannelInformation.AccountName
	remark := request.MetadataObj.TransferDestination.Description

	// Do Inquiry
	if _, err := s.beneficiaryAccountSvc.FindByBankCodeAndAccountNo(ctx, &beneficiaryAccountModel.CheckAccountRequest{
		BeneficiaryBankCode:  beneficiaryBankCode,
		BeneficiaryAccountNo: beneficiaryAccountNo,
		MerchantID:           request.MerchantID,
		AdditionalInfo:       map[string]any{},
	}); err != nil {
		return err
	}

	traceId, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)
	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		RequestId:   traceId,
		From:        "Refund",
		OriginId:    request.RefundID,
		ReferenceId: request.MerchantID,
	})
	snapCoreResp, err := s.snapCoreRepo.BankTransfer(ctx, &snapCoreModel.BankTransferRequest{
		BTBeneficiaryRequest: snapCoreModel.BTBeneficiaryRequest{
			BeneficiaryBankCode:    beneficiaryBankCode,
			BeneficiaryAccountNo:   beneficiaryAccountNo,
			BeneficiaryAccountName: beneficiaryAccountName,
		},
		Currency:             constant.CurrencyIDR,
		Amount:               commonModel.Amount{Currency: constant.CurrencyIDR, Value: fmt.Sprintf("%.f", request.Amount)},
		PurposeOfTransaction: snapCoreModel.DefaultPurchaseOfTransaction,
		TransactionDate:      request.CreatedAt,
		Remark:               remark,
	}, &snapCoreModel.BankTransferHeaderRequest{
		ExternalId: request.RefundLedgerID,
		MerchantId: request.MerchantID,
	})
	if snapCoreResp != nil && snapCoreResp.UUID != "" {
		processorReferenceID = snapCoreResp.UUID
		reconReferenceNo = snapCoreResp.GetReconReferenceNo()

		if snapCoreResp.Status == constant.SnapCoreBankTransferStatusPending {
			return constant.ErrBankTransferStillInPending
		}
	}

	if err != nil {
		if snapCoreResp != nil {
			_, reasonType, reasonDesc := snapCoreResp.MappingAccountTransactionErrStatus()

			// Passing reference for reason type and desc
			request.RefundLedgerReasonType = &reasonType
			request.RefundLedgerReasonDesc = &reasonDesc
		}

		return err
	}

	return nil
}
