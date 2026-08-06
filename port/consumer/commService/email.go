package commService

import (
	"context"
	"fmt"
	"github.com/paper-indonesia/pdk/v2/logger"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/paperCommunication"
	pb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/commService"

	"google.golang.org/protobuf/proto"
)

func (h *handler) PostEmailHandler(ctx context.Context, body []byte, _ string) error {
	ctx, segment := otelTracer.Start(ctx, "port/consumer/commService/PostEmailHandler")
	defer segment.End()

	defer func() {
		if r := recover(); r != nil {
			h.log.Error(ctx, "Panic recovery from PostEmailHandler", logger.Error(fmt.Errorf("%v", r)))
		}
	}()

	payload := &pb.EmailRequest{}
	if err := proto.Unmarshal(body, payload); err != nil {
		h.log.Error(ctx, "Failed while unmarshal proto", logger.Error(err))
		return constant.ErrUnmarshalProto
	}

	request := &paperCommunication.Email{
		Header: payload.Subject,
		Event:  payload.Event,
		Body: paperCommunication.EmailBody{
			Email: payload.To,
			Data:  payload.Content.AsMap(),
		},
		Priority:          constant.EmailPriority(payload.Priority),
		OnErrCanBeRetried: payload.ToBeRetriedOnFailure,
	}

	merchantId, _ := ctx.Value(constant.CtxMerchantIDKey).(string)

	h.log.Info(ctx, "Email sending process", logger.Any("details", map[string]string{"merchantId": merchantId, "to": payload.To, "event": payload.Event}))

	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		From:        "Comm-Service-Email",
		ReferenceId: merchantId,
	})
	return h.service.PostEmailService(ctx, payload.From, request)
}
