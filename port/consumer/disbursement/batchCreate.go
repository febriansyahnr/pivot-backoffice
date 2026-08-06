package disbursementConsumerController

import (
	"context"
	"encoding/base64"
	"errors"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	pb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/disbursement"

	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
	"google.golang.org/protobuf/proto"
)

func (c *DisbursementConsumer) BatchCreateDisbursement(ctx context.Context, body []byte, _ string) (err error) {
	ctx, segment := otelTracer.Start(ctx, "port/consumer/disbursement/BatchCreateDisbursement")
	defer segment.End()

	payload := &pb.BatchCreateDisbursementRequest{}
	if err = proto.Unmarshal(body, payload); err != nil {
		return constant.ErrUnmarshalProto
	}
	c.logger.Info(
		ctx, "Incoming message on BatchCreateDisbursement",
		logger.String("base64Payload", base64.StdEncoding.EncodeToString(body)), logger.Any("payload", payload),
	)

	request := &disbursementModel.BatchCreateDisbursementRequest{
		BulkID:       payload.BulkId,
		MerchantID:   payload.MerchantId,
		MerchantName: payload.MerchantName,
		CreatedBy:    payload.CreatedBy,
		CreatedFrom:  payload.CreatedFrom,
		TotalTrx:     int(payload.TotalTrx),
		AutoApprove:  payload.AutoApprove,
		Data:         make([]disbursementModel.CreateSingleRequest, len(payload.Data)),
	}
	for i, data := range payload.Data {
		request.Data[i] = disbursementModel.CreateSingleRequest{
			ReferenceID:            data.ReferenceId,
			BeneficiaryBankCode:    data.BeneficiaryBankCode,
			BeneficiaryBankName:    data.BeneficiaryBankName,
			BeneficiaryAccountNo:   data.BeneficiaryAccountNo,
			BeneficiaryAccountName: data.BeneficiaryAccountName,
			Remark:                 data.Remark,
			PurposeID:              data.PurposeId,
			InquiryID:              data.InquiryId,
		}
		request.Data[i].Amount, _ = decimal.NewFromString(data.Amount)
	}

	// Call BatchCreateDisbursement
	err = c.disbursementSvc.BatchCreateDisbursement(ctx, request)
	if err != nil {
		// when its beneficiary limit err then return nil to avoid increase error rate
		if _, ok := err.(*disbursementModel.ApprovalResultErr); ok || errors.Is(err, constant.ErrBeneficiaryLimitRestrictions) {
			return nil
		}

		return err
	}
	return nil
}
