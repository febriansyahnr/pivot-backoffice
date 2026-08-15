package snapCoreRepository

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	routingProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/routingProcessor/bankTransfer"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/snapCore/bankTransfer"
)

func (r *snapCoreRepository) TriggerTransfer(
	ctx context.Context,
	request *routingProcessorModel.BankTransferRequest,
) (*routingProcessorModel.BankTransferResponseData, error) {
	ctx, span := otelTracer.Start(ctx, "internal/repository/snapCore/TriggerTransfer")
	defer span.End()

	snapCoreReq := &snapCoreModel.BankTransferRequest{
		PartnerReferenceNo: request.PartnerReferenceNo,
		BTBeneficiaryRequest: snapCoreModel.BTBeneficiaryRequest{
			BeneficiaryBankCode:          request.Beneficiary.BankCode,
			BeneficiaryAccountNo:         request.Beneficiary.AccountNo,
			BeneficiaryAccountName:       request.Beneficiary.AccountName,
			BeneficiaryEmail:             request.Beneficiary.Email,
			BeneficiaryCustomerResidence: request.Beneficiary.CustomerResidence,
			BeneficiaryAddress:           request.Beneficiary.Address,
			BeneficiaryCustomerType:      request.Beneficiary.CustomerType,
			BeneficiaryCitizenStatus:     request.Beneficiary.CitizenStatus,
			BeneficiaryBICCode:           request.Beneficiary.BICCode,
		},
		Amount:               request.Amount,
		Currency:             request.Currency,
		Remark:               request.Remark,
		PurposeOfTransaction: request.PurposeOfTransaction,
		SourceAccountNo:      request.Source.AccountNo,
		SourceAccountName:    request.Source.AccountName,
		TransactionDate:      request.TransactionDate,
		AdditionalInfo:       request.AdditionalInfo,
		OriginatorInfos: []snapCoreModel.OriginatorInfosRequest{{
			OriginatorCustomerNo:   request.Source.AccountNo,
			OriginatorCustomerName: request.Source.Name,
			OriginatorBankCode:     request.Source.BankCode,
		}},
	}

	transfer, err := r.BankTransfer(ctx, snapCoreReq, &request.HeaderRequest)
	if transfer == nil && err != nil {
		return nil, err
	}

	response := &routingProcessorModel.BankTransferResponseData{
		ProcessorReference:   constant.SnapCoreProcessor,
		ResponseCode:         transfer.ResponseCode,
		ResponseMessage:      transfer.ResponseMessage,
		UUID:                 transfer.UUID,
		PartnerReferenceNo:   transfer.PartnerReferenceNo,
		BankReferenceNo:      transfer.BankReferenceNo,
		BankProcessor:        transfer.BankProcessor,
		Amount:               transfer.Amount,
		BeneficiaryAccountNo: transfer.BeneficiaryAccountNo,
		BeneficiaryBankCode:  transfer.BeneficiaryBankCode,
		SourceAccountNo:      transfer.SourceAccountNo,
		Status:               transfer.Status,
		TransferType:         transfer.TransferType,
		ExternalID:           transfer.ExternalID,
		TransactionDate:      transfer.TransactionDate,
	}

	return response, err
}
