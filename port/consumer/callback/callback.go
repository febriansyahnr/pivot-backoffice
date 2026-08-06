package callback

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/monitoring"
	pb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/callback"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	callbackPkg "github.com/paper-indonesia/pivot-backoffice/pkg/callback"
	"github.com/paper-indonesia/pivot-backoffice/pkg/customMetric"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/port/consumer"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
	"google.golang.org/protobuf/proto"
)

var otelTracer = otel.Tracer("CallbackConsumer")

type callback struct {
	logger      logger.ILogger
	callbackSvc service.ICallbackService
	queue       rabbitMqExt.IRabbitMQExt
}

func New(
	log logger.ILogger, callbackSvc service.ICallbackService, rmq rabbitMqExt.IRabbitMQExt,
) consumer.ICallbackConsumer {
	return &callback{
		logger:      log,
		callbackSvc: callbackSvc,
		queue:       rmq,
	}
}

func (c *callback) ProcessCallback(ctx context.Context, body []byte, _ string) (err error) {
	ctx, segment := otelTracer.Start(ctx, "port/consumer/callback/ProcessCallback")
	defer segment.End()

	var (
		start   = time.Now()
		payload = &pb.ProcessCallbackRequest{}
	)
	defer func() {
		if r := recover(); r != nil {
			if err == nil {
				err = fmt.Errorf("panic recovery: %v", r)
			}
			c.logger.Error(ctx, "Panic recovery from ProcessCallback", logger.Error(fmt.Errorf("%v", r)))
		}

		metricCounter := monitoring.CustomMetric{
			MetricInstrumentType: constant.MetricInstrumentTypeCounter,
			MetricValue:          1,
			ComponentName:        constant.ComponentMerchantCallback,
			MetricName:           constant.MetricNameMerchantCallbackCount,
			Attributes: map[string]any{
				"merchantId": payload.MerchantId,
				"eventName":  payload.Event,
				"statusCode": int64(0),
			},
		}

		metricDuration := monitoring.CustomMetric{
			MetricInstrumentType: constant.MetricInstrumentTypeHistogram,
			MetricValue:          time.Since(start).Milliseconds(),
			ComponentName:        constant.ComponentMerchantCallback,
			MetricName:           constant.MetricNameMerchantCallbackDuration,
			Attributes:           map[string]any{},
		}
		maps.Copy(metricDuration.Attributes, metricCounter.Attributes)

		if err == nil {
			metricCounter.Attributes["statusCode"], metricDuration.Attributes["statusCode"] = int64(200), int64(200)

		} else {
			if deliveryFail, ok := err.(*callbackPkg.ErrHttpClient); ok {
				statusCode := int64(deliveryFail.StatusCode())
				metricCounter.Attributes["statusCode"], metricDuration.Attributes["statusCode"] = statusCode, statusCode
			}
			metricCounter.Attributes["errorDetail"] = err.Error()
		}

		for _, metricData := range []monitoring.CustomMetric{metricCounter, metricDuration} {
			if errMetric := customMetric.RecordCustomMetric(ctx, &metricData); errMetric != nil {
				c.logger.Warn(ctx, "Failed while sending metrics on merchant callback process", logger.Error(errMetric))
			}
		}

		if errors.Is(err, constant.ErrCallbackURLNotConfigured) {
			err = nil // Return the error value as nil since ErrCallbackURLNotConfigured is only used for metrics and the process will not be retried.
		}

		// If callback delivery is redirected to a non-workflow process due to specific conditions (e.g., maintenance) and the delivery fails,
		// the received message will be formatted into data to be rescheduled via the workflow once the system becomes active again.
		if err != nil && !errors.Is(err, constant.ErrUnmarshalProto) && !constant.IsMerchantCallbackWorkflowEnabled(config.Environment()) {
			workflowPayload, _ := json.Marshal(map[string]string{
				"payload": base64.StdEncoding.EncodeToString(body),
			})

			messageId := uuid.NewString()
			ctx = context.WithValue(ctx, constant.CtxMessageId, messageId)

			c.logger.Info(ctx, "Publish for resubmission via workflow", logger.String("messageId", messageId), logger.ByteString("payload", workflowPayload))

			_ = c.queue.Publish(ctx, rabbitMqExt.WorkflowCallbackRoutingKey, nil, workflowPayload)
		}
	}()

	if err := proto.Unmarshal(body, payload); err != nil {
		return constant.ErrUnmarshalProto
	}

	request := &callbackModel.ProcessCallbackRequest{
		Name:   payload.Name,
		Event:  payload.Event,
		IsSnap: payload.IsSnap,
	}
	request.MerchantID, _ = uuid.Parse(payload.MerchantId)

	if err = request.Bind(payload.Request); err != nil {
		return err
	}

	c.logger.Info(ctx, "Callback request body", logger.Any("request", request.Request))

	return c.callbackSvc.ProcessCallback(ctx, request)
}
