package routingprocessorService

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	requestAccountInquiries "github.com/paper-indonesia/pivot-backoffice/internal/model/requestAccountInquiry"
	routingProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/accountInquiry"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
	amqp "github.com/rabbitmq/amqp091-go"
)

func (s *routingProcessorService) ProcessAccountInquiryCallback(ctx context.Context, payload *routingProcessorModel.InquiryAccountResponseData) error {
	ctx, span := tracer.Start(ctx, "internal/service/v1/routingProcessor/ProcessAccountInquiryCallback")
	defer span.End()

	var (
		err error
	)

	// get outbound first

	ob, err := s.outboundRepository.FindByID(ctx, payload.PartnerReferenceNo)
	if err != nil {
		s.logger.Error(ctx, "RequestAccountInquiry - error when find outbound", logger.Error(err))
		return err
	}

	if ob == nil {
		s.logger.Error(ctx, "RequestAccountInquiry - outbound not found")
		return errors.New(response.HttpErrNotFound)
	}

	// publish to reply queue
	var client outbound.Client
	err = ob.Client.Unmarshal(&client)
	if err != nil {
		s.logger.Error(ctx, "RequestAccountInquiry - error when unmarshal client", logger.Error(err))
		return err
	}

	err = s.updateAccountInquiryData(ctx, client.OriginId, payload)
	if err != nil {
		s.logger.Warn(ctx, "inquiry account not found", logger.Error(err))
	}

	if client.ReplyToAddress == "" {
		s.logger.Error(ctx, "RequestAccountInquiry - reply to address not found")
		return errors.New(response.HttpErrNotFound)
	}

	payloadB, _ := json.Marshal(payload)
	err = s.rabbitMq.PublishToReplyQueue(ctx, client.ReplyToAddress, amqp.Publishing{
		ContentType: constant.MIMEApplicationJSON,
		Body:        payloadB,
	})
	if err != nil {
		s.logger.Error(ctx, "RequestAccountInquiry - error when publish to reply queue", logger.Error(err))
	}

	return err
}

func (s *routingProcessorService) AddressingReplyToAccountInquiry(ctx context.Context, payload *routingProcessorModel.InquiryAccountResponseData) error {
	ctx, span := tracer.Start(ctx, "internal/service/v1/routingProcessor/AddressingReplyToAccountInquiry")
	defer span.End()

	var (
		err     error
		client  outbound.Client
		replyTo string
	)

	// select outbound
	ob, err := s.outboundRepository.FindByID(ctx, payload.PartnerReferenceNo)
	if err != nil {
		s.logger.Error(ctx, "AddressingReplyToAccountInquiry - error when find outbound", logger.Error(err))
		return err
	}

	if ob == nil {
		s.logger.Error(ctx, "AddressingReplyToAccountInquiry - outbound not found")
		return errors.New(response.HttpErrNotFound)
	}

	err = ob.Client.Unmarshal(&client)
	if err != nil {
		s.logger.Error(ctx, "AddressingReplyToAccountInquiry - error when unmarshal client", logger.Error(err))
		return err
	}

	replyTo, ok := ctx.Value(constant.CtxRabbitMQReplyTo).(string)
	if !ok {
		err = errors.New(response.HttpErrInternal)
		s.logger.Error(ctx, "AddressingReplyToAccountInquiry - error when get reply to", logger.Error(err))
		return err
	}

	client.ReplyToAddress = replyTo
	err = s.outboundRepository.UpdateClient(ctx, ob.Id, &client)
	if err != nil {
		s.logger.Error(ctx, "AddressingReplyToAccountInquiry - error when update client", logger.Error(err))
		return err
	}

	return nil
}

// this function will called and can used when waiting time for callback was timeout
// and data already set in database
// check to beneficiary account
func (s *routingProcessorService) updateAccountInquiryData(ctx context.Context, clientOriginID string, payload *routingProcessorModel.InquiryAccountResponseData) error {
	getAccount, err := s.requestAccountInquiryRepo.FindByID(ctx, clientOriginID)
	if err != nil || getAccount == nil {
		s.logger.Warn(ctx, "inquiry account not found", logger.Error(err))
		return err
	}

	if util.IsPatternMatch(constant.SnapCoreResponseCodeInvalidFieldPattern, payload.ResponseCode) && getAccount.Status.String == constant.RequestAccountInquiryStatusPending {
		getAccount.Status.String = constant.RequestAccountInquiryStatusInvalid

		_, getAccount.MetadataObj.DetailStatus = requestAccountInquiries.NewDetailStatusRequestInquiry(constant.RequestAccountInquiryStatusInvalid, "", "", "", "")
		getAccount.MetadataObj.SnapCoreResponse = payload.ToSnapCoreResponseData()
		getAccount.SetMetadataNullJSONText()
		err = s.requestAccountInquiryRepo.Update(ctx, getAccount)
		if err != nil {
			return err
		}

		return nil
	}

	return nil
}
