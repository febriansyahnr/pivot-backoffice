package callback

import (
	"context"
	"fmt"
	"net/http"

	callbackModel "github.com/paper-indonesia/pivot-backoffice/pkg/callback/model"
	"github.com/paper-indonesia/pivot-backoffice/pkg/logger"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"

	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("CallbackService")

type ICallbackPartner interface {
	Callback(ctx context.Context, request callbackModel.CallbackRequest, headers map[string]string) (string, error)
}

type callbackPartner struct {
	logger     logger.ILogger
	httpClient httputil.HTTPClient
}

func New(logger logger.ILogger, httpClient httputil.HTTPClient) ICallbackPartner {
	return &callbackPartner{
		logger:     logger,
		httpClient: httpClient,
	}
}

type ErrHttpClient struct {
	statusCode   int
	responseBody []byte
	err          error
}

func (e *ErrHttpClient) Error() string {
	return fmt.Sprintf("%v", e.err)
}

func (e *ErrHttpClient) StatusCode() int {
	return e.statusCode
}

func (e *ErrHttpClient) ResponseBody() []byte {
	return e.responseBody
}

var ErrTestBadRequest = &ErrHttpClient{statusCode: http.StatusBadRequest, responseBody: []byte(`{"message":"bad request"}`)}
