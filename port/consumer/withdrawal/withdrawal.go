package withdrawalConsumer

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	pb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/withdrawal"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"

	"github.com/paper-indonesia/pdk/v2/logger"
	"google.golang.org/protobuf/proto"
)

func (h *handler) WithdrawalProcess(ctx context.Context, body []byte, _ string) error {
	ctx, segment := otelTracer.Start(ctx, "port/consumer/withdrawal/WithdrawalProcess")
	defer segment.End()

	defer func() {
		if r := recover(); r != nil {
			h.logger.Error(ctx, "Panic recovery from WithdrawalProcess", logger.Error(fmt.Errorf("%v", r)))
		}
	}()

	payload := &pb.WithdrawalRequest{}
	if err := proto.Unmarshal(body, payload); err != nil {
		h.logger.Error(ctx, "Failed while unmarshal proto", logger.Error(err))
		return constant.ErrUnmarshalProto
	}

	request := &withdrawal.WithdrawalRequest{
		AccountName:          payload.AccountName,
		Amount:               payload.Amount,
		IsFullAmount:         payload.IsFullAmount,
		Type:                 payload.Type,
		Destination:          constant.WithdrawalDestBankTransfer,
		BeneficiaryBankCode:  payload.BeneficiaryBankCode,
		BeneficiaryAccountNo: payload.BeneficiaryAccountNo,
		UserId:               "", // User ID is not set when the process is auto withdrawal
		MerchantId:           payload.MerchantId,
		Reason:               payload.Reason,
		Source:               constant.SourceSystem,
	}

	start, message := time.Now().UTC(), ""

	defer func() {
		duration := time.Now().UTC().Sub(start)

		h.logger.Info(
			ctx, "Proccess automatic withdrawal",
			logger.Any("details", map[string]interface{}{
				"merchantId":           request.MerchantId,
				"accountName":          request.AccountName,
				"beneficiaryBankCode":  request.BeneficiaryBankCode,
				"beneficiaryAccountNo": request.BeneficiaryAccountNo,
				"amount":               request.Amount,
			}), logger.String("message", message), logger.Int64("duration", duration.Milliseconds()), logger.String("durationHuman", duration.String()))
	}()
	if _, err := h.service.Create(ctx, request); err != nil {
		if strings.Contains(err.Error(), constant.ErrInsufficientBalance.Error()) {
			message = "Merchant balance is insufficient"
			return nil
		}

		message = err.Error()
		return err
	}
	return nil
}
