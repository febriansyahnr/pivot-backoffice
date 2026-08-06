package refundService

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/proto/common"
	pb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/callback"
	refundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/refund"
	"github.com/paper-indonesia/pdk/v2/logger"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// SendCallback sends a callback notification for a refund to the merchant.
//
// It retrieves the refund details, prepares the callback data with refund information,
// and publishes it to a RabbitMQ queue for processing.

// The callback includes detailed refund information such as:
//   - Refund ID, status, and reason
//   - Payment details (charge ID, session ID)
//   - Amount information (captured and refunded amounts)
//   - Transfer destination details
//   - Metadata and timestamps
func (s *RefundService) SendCallback(ctx context.Context, refundID string, merchantID string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/refund/sendCallback")
	defer segment.End()

	var (
		callbackRequestWrapper *anypb.Any
		callbackEvent          string
		err                    error
	)

	refund, err := s.GetRefundDetail(ctx, refundModel.FilterRefundRequest{
		UUID:       refundID,
		MerchantID: merchantID,
	})
	if err != nil {
		s.logger.Error(ctx, "GetRefundDetail", logger.Error(err))
		return err
	}

	pbRequest := &pb.RefundCallbackRequest{
		Id:                refund.ID,
		ClientReferenceId: refund.ClientReferenceID,
		PaymentSessionId:  refund.PaymentSessionID,
		ChargeId:          refund.ChargeID,
		CapturedAmount: &common.Amount{
			Value:    refund.CapturedAmount.Value,
			Currency: refund.CapturedAmount.Currency,
		},
		Amount: &common.Amount{
			Value:    refund.Amount.Value,
			Currency: refund.Amount.Currency,
		},
		Status:          refund.Status,
		Reason:          refund.Reason,
		Description:     refund.Description,
		DestinationType: refund.DestinationType,
		Method:          refund.Method,
		IsFullAmount:    refund.IsFullAmount,
		Metadata:        &anypb.Any{},
		FailureCode:     &refund.FailureCode,
		CreatedAt:       timestamppb.New(refund.CreatedAt),
		UpdatedAt:       timestamppb.New(refund.UpdatedAt),
	}

	if refund.Metadata != nil {
		metadata, err := s.GetMetadata(ctx, refund.Metadata)
		if err != nil {
			s.logger.Error(ctx, "invalid metadata", logger.Error(err), logger.Any("metadata", refund.Metadata))
		}

		structPB, _ := structpb.NewStruct(metadata)
		pbRequest.Metadata, _ = anypb.New(structPB)
	}

	if refund.TransferDestination != nil {
		pbRequest.TransferDestination = &pb.TransferDestination{
			ChannelCode: refund.TransferDestination.ChannelCode,
			ChannelInformation: &pb.ChannelInformation{
				AccountNumber: refund.TransferDestination.ChannelInformation.AccountNumber,
				AccountName:   refund.TransferDestination.ChannelInformation.AccountName,
			},
			Description: refund.TransferDestination.Description,
		}
	}

	callbackName := constant.CallbackNameRefund
	callbackRequestWrapper, _ = anypb.New(pbRequest)
	callbackEvent = fmt.Sprintf(constant.CallbackEventRefundPattern, refund.Status)

	// publish callback
	callbackRequest := &pb.ProcessCallbackRequest{
		Name:       callbackName,
		Event:      callbackEvent,
		MerchantId: refund.MerchantID,
		Request:    callbackRequestWrapper,
		IsSnap:     false,
	}

	err = s.rabbitMqExt.PublishMerchantCallback(ctx, callbackRequest)
	if err != nil {
		s.logger.Error(ctx, "Publish callback request", logger.Error(err))
		return err
	}

	s.logger.Info(ctx, "Send callback request", logger.String("callbackName", callbackName), logger.String("callbackEvent", callbackEvent), logger.String("merchantId", refund.MerchantID))
	return nil
}

// GetMetadata converts any metadata interface{} into a map[string]interface{}.
// It handles nil metadata by returning an empty map, marshals the metadata
// to JSON and then unmarshals it back into a map structure.
func (s *RefundService) GetMetadata(ctx context.Context, metadata interface{}) (map[string]interface{}, error) {
	_, span := otelTracer.Start(ctx, "internal/service/v1/refund/GetMetadata")
	defer span.End()

	metadataMap := make(map[string]interface{})
	if metadata == nil {
		return metadataMap, nil
	}

	bytes, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(bytes, &metadataMap)
	return metadataMap, nil
}
