package activity

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	activityModel "github.com/paper-indonesia/pivot-backoffice/internal/model/activity"
	pb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/activity"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/consumer"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type activity struct {
	logger      logger.ILogger
	activitySvc service.IActivityService
}

var otelTracer = otel.Tracer("ActivityConsumer")

func New(
	log logger.ILogger, activitySvc service.IActivityService,
) consumer.IActivityConsumer {
	return &activity{
		logger:      log,
		activitySvc: activitySvc,
	}
}

func (c *activity) Insert(ctx context.Context, body []byte, _ string) error {
	ctx, segment := otelTracer.Start(ctx, "port/consumer/activity/Insert")
	defer segment.End()

	defer func() {
		if r := recover(); r != nil {
			c.logger.Error(ctx, "Panic recovery from ActivityInsert", logger.Error(fmt.Errorf("%v", r)))
		}
	}()

	if _, ok := ctx.Value(pdkConst.CtxTraceIdKey).(string); !ok {
		ctx = context.WithValue(ctx, pdkConst.CtxTraceIdKey, uuid.NewString())
	}

	payload := &pb.Activity{}
	if err := proto.Unmarshal(body, payload); err != nil {
		c.logger.Error(ctx, "Unmarshal protobuf data", logger.Error(err))
		return errors.New("unmarshal protobuf: cannot parse invalid wire-format data")
	}

	request := &activityModel.Activity{
		ID:          payload.Id,
		MerchantID:  payload.MerchantId,
		UserID:      payload.UserId,
		Tag:         payload.Tag,
		Activity:    payload.Activity,
		ServiceName: payload.ServiceName,
		CreatedAt:   payload.CreatedAt.AsTime(),
		UpdatedAt:   payload.UpdatedAt.AsTime(),
	}

	if payload.Parameter != nil {
		rawParams := &wrapperspb.BytesValue{}
		if err := anypb.UnmarshalTo(payload.Parameter, rawParams, proto.UnmarshalOptions{}); err != nil {
			c.logger.Error(ctx, "Unmarshal wrappers bytes value type", logger.Error(err))
			return errors.New(`unmarshal wrappers bytes type: mismatched message type`)
		}

		var params map[string]any

		_ = json.Unmarshal(rawParams.Value, &params)

		if len(params) > 0 {
			request.Parameter = &params
		}
	}
	err := c.activitySvc.Create(ctx, request)
	if err != nil {
		if mySqlErr, ok := err.(*mysql.MySQLError); ok && mySqlErr.Number == 1062 {
			c.logger.Error(ctx, "duplicate entry when inserting activity_logs", logger.Error(err))
			return nil
		}

		if err == constant.ErrNoRowsAffected {
			return nil
		}

		c.logger.Error(ctx, "failed when inserting activity_logs", logger.Error(err))
		return err
	}

	return nil
}
