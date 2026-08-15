package snapCoreRepository

import (
	"context"

	routingProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/bankTransfer"
)

func (r *snapCoreRepository) GetTransferById(ctx context.Context, id string, forceFailed bool) (*routingProcessorModel.BankTransferResponseData, error) {
	ctx, span := otelTracer.Start(ctx, "internal/repository/snapCore/GetTransferById")
	defer span.End()

	resp, err := r.FindBankTransferByExternalID(ctx, id, forceFailed)
	var response *routingProcessorModel.BankTransferResponseData

	if resp != nil {
		response = &routingProcessorModel.BankTransferResponseData{
			ResponseCode:         resp.ResponseCode,
			ResponseMessage:      resp.ResponseMessage,
			UUID:                 resp.UUID,
			PartnerReferenceNo:   resp.PartnerReferenceNo,
			BankReferenceNo:      resp.BankReferenceNo,
			Amount:               resp.Amount,
			BeneficiaryAccountNo: resp.BeneficiaryAccountNo,
			BeneficiaryBankCode:  resp.BeneficiaryBankCode,
			SourceAccountNo:      resp.SourceAccountNo,
			Status:               resp.Status,
			TransferType:         resp.TransferType,
			ExternalID:           resp.ExternalID,
		}
	}

	return response, err
}
