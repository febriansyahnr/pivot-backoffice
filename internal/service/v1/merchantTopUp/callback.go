package merchantTopUp

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchantTopUp"
	commonPb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/common"
	pb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/callback"
	"github.com/paper-indonesia/pdk/v2/logger"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *merchantTopUpService) SendCallback(ctx context.Context, event string, request *merchantTopUp.MerchantTopUpCallbackRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchantTopUp/SendCallback")
	defer segment.End()

	s.logger.Info(ctx, "Process Send Callback Merchant Top Up")

	pbRequest := &pb.MerchantTopUpCallbackRequest{
		Uuid:         request.UUID,
		MerchantId:   request.MerchantID,
		MerchantName: request.MerchantName,
		AccountName:  request.AccountName,
		Amount: &commonPb.Amount{
			Currency: request.Amount.Currency,
			Value:    request.Amount.Value,
		},
		BalanceBefore: &commonPb.Amount{
			Currency: request.BalanceBefore.Currency,
			Value:    request.BalanceBefore.Value,
		},
		BalanceAfter: &commonPb.Amount{
			Currency: request.BalanceAfter.Currency,
			Value:    request.BalanceAfter.Value,
		},
		PaymentMethod: &pb.MerchantTopUpCallbackPaymentMethodObject{
			Type: request.PaymentMethod.Type,
		},
		PaymentMethodOptions: &pb.MerchantTopUpCallbackPaymentMethodOptionsObject{},
		TransactionTime:      timestamppb.New(request.TransactionTime),
	}
	if request.PaymentMethodOptions.VirtualAccount != nil {
		pbRequest.PaymentMethodOptions.VirtualAccount = &pb.MerchantTopUpCallbackPaymentMethodOptionVAObject{
			Channel:              request.PaymentMethodOptions.VirtualAccount.Channel,
			VirtualAccountNumber: request.PaymentMethodOptions.VirtualAccount.VirtualAccountNumber,
			VirtualAccountName:   request.PaymentMethodOptions.VirtualAccount.VirtualAccountName,
		}
	}

	callbackRequestWrapper, _ := anypb.New(pbRequest)

	// Send Callback to Merchant (Main or Sub)
	callbackRequest := &pb.ProcessCallbackRequest{
		Name:       constant.CallbackNameMerchantTopUp,
		Event:      event,
		MerchantId: request.MerchantID,
		Request:    callbackRequestWrapper,
		IsSnap:     false,
	}

	err := s.rabbitMqExt.PublishMerchantCallback(ctx, callbackRequest)
	if err != nil {
		s.logger.Error(ctx, "Publish callback request", logger.Error(err))
		return err
	}

	// If parentMerchant exist then send to Main
	if request.ParentMerchantID != "" {
		s.logger.Info(ctx, "Process Send Callback Merchant Top Up For Main Account")

		// Send Callback to Main Merchant
		callbackRequest = &pb.ProcessCallbackRequest{
			Name:       constant.CallbackNameMerchantTopUp,
			Event:      event,
			MerchantId: request.ParentMerchantID,
			Request:    callbackRequestWrapper,
			IsSnap:     false,
		}

		err = s.rabbitMqExt.PublishMerchantCallback(ctx, callbackRequest)
		if err != nil {
			s.logger.Error(ctx, "Publish callback request", logger.Error(err))
			return err
		}
	}

	return nil
}
