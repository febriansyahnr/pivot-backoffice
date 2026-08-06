package withdrawalService

import (
	"context"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/proto/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/callback"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"

	"google.golang.org/protobuf/types/known/anypb"
)

// Send a callback for the withdrawal status with the destination to bank transfer.
func (s *withdrawalService) SendWithdrawalStatusCallback(ctx context.Context, request withdrawal.WithdrawalStatusCallbackRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/withdrawal/SendWithdrawalStatusCallback")
	defer segment.End()

	// Initiate recipient ids
	recipientIds := []string{request.MerchantId}

	if parentId, _ := ctx.Value(constant.CtxParentMerchantId).(string); parentId != "" {
		// Main-Merchant initiate the withdrawal (on behalf of sub-merchant) -> Send the callback to Main-Merchant only.
		recipientIds = []string{parentId}

	} else {

		if merchant, err := s.merchantRepo.FindMerchantByID(ctx, request.MerchantId); err != nil {
			return err

		} else if merchant != nil && merchant.ParentID.Valid {
			// Sub-Merchant initiate the withdrawal -> Send the callback to Sub-Merchant AND Main-Merchant.
			recipientIds = append(recipientIds, merchant.ParentID.String)
		}
	}

	// Callback payload
	withdrawalStatus := &callback.WithdrawalStatus{
		Id:         request.ID,
		MerchantId: request.MerchantId,
		Withdrawal: &callback.WithdrawalDetail{
			ReferenceId:  request.Withdrawal.ReferenceID,
			WithdrawType: request.Withdrawal.WithdrawType,
			BalanceType:  request.Withdrawal.BalanceType,
			IsFullAmount: request.Withdrawal.IsFullAmount,
			Amount: &common.Amount{
				Currency: request.Withdrawal.Amount.Currency,
				Value:    request.Withdrawal.Amount.Value,
			},
			Description: request.Withdrawal.Description,
		},
		Status:    request.Status,
		CreatedAt: request.CreatedAt,
		UpdatedAt: request.UpdatedAt,
	}
	withdrawalStatusWrapped, err := anypb.New(withdrawalStatus)
	if err != nil {
		return err
	}

	for _, recipientId := range recipientIds {

		callbackRequest := &callback.ProcessCallbackRequest{
			Name:       constant.CallbackNameWithdrawal,
			Event:      fmt.Sprintf(constant.CallbackEventWithdrawPattern, request.Status),
			MerchantId: recipientId,
			Request:    withdrawalStatusWrapped,
		}

		_ = s.rmq.PublishMerchantCallback(ctx, callbackRequest)
	}

	return nil
}
