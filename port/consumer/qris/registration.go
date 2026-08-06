package qris

import (
	"context"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/proto/qr_mpm"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/qris"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
	"google.golang.org/protobuf/proto"

	"github.com/google/uuid"
)

func (c *consumerHandler) Process(ctx context.Context, body []byte, _ string) (err error) {
	ctx, segment := otelTracer.Start(ctx, "port/consumer/qris/QrisRegistrationCallback")
	defer segment.End()

	defer func() {
		if r := recover(); r != nil {
			c.logger.Error(ctx, "Panic recovery from QrisProcess", logger.Error(fmt.Errorf("%v", r)))
		}
	}()

	if _, ok := ctx.Value(pdkConst.CtxTraceIdKey).(string); !ok {
		ctx = context.WithValue(ctx, pdkConst.CtxTraceIdKey, uuid.NewString())
	}

	start := time.Now().UTC()

	protoReq := &qr_mpm.RegistrationCallbackRequest{}
	if err = proto.Unmarshal(body, protoReq); err != nil {
		return fmt.Errorf("unmarshal %w", err)
	}
	defer func() {
		c.logger.Info(
			ctx, fmt.Sprintf("Message is received from the %s queue", rabbitMqExt.QrisRegistrationCallbackQueueName),
			logger.Any("payload", protoReq), logger.Bool("status", err == nil), logger.String("duration", time.Now().UTC().Sub(start).String()),
		)
	}()

	request := &qris.RegistrationCallback{
		Id:            protoReq.ApplicationCode,
		ApplymentCode: protoReq.ApplymentCode,
		MerchantId:    protoReq.MID,
		Status:        protoReq.AuditStatus,
	}
	if protoReq.DateTime != nil {
		request.Datetime = protoReq.DateTime.AsTime()
	}

	return c.service.RegistrationCallback(ctx, request)
}
