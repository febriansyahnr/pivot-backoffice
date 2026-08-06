package callback

import (
	"context"
	"errors"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	callbackModel "github.com/paper-indonesia/pivot-backoffice/pkg/callback/model"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"go.uber.org/zap"
)

func (s *callbackPartner) Callback(ctx context.Context, request callbackModel.CallbackRequest, headers map[string]string) (string, error) {
	ctx, segment := otelTracer.Start(ctx, "pkg/callback/Callback")
	defer segment.End()

	response, statusCode, err := s.httpClient.Request(
		ctx, http.MethodPost, request.URL, request.Request, headers,
	)
	if err != nil {
		s.logger.Error(
			ctx, "Failed to send callback to merchant",
			zap.Any(
				"request", map[string]any{
					"url":  request.URL,
					"body": request.Request,
				}),
			zap.Error(err),
		)
		if errors.Is(err, http.ErrHandlerTimeout) || errors.Is(err, context.DeadlineExceeded) {
			return "", pkgErrs.New(httpResponse.HttpErrRequest, errors.New("request timeout"))
		}
		return "", pkgErrs.New(httpResponse.HttpErrInternal, constant.ErrInternalServerForUser)
	}

	if statusCode >= 400 {
		s.logger.Info(ctx, "Failed to send callback to merchant: Received non-2xx response",
			zap.Any(
				"request", map[string]any{
					"url":  request.URL,
					"body": request.Request,
				},
			),
			zap.Any(
				"response", map[string]any{
					"statusCode":   statusCode,
					"responseBody": string(response),
				},
			),
		)
		return string(response), &ErrHttpClient{
			statusCode:   statusCode,
			responseBody: response,
			err:          constant.ErrInvokeClientWebhook,
		}
	}

	s.logger.Info(ctx, "Response from callback", zap.String("body", string(response)))
	return string(response), nil
}
