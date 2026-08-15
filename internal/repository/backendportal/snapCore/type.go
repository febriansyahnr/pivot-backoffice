package snapCoreRepository

import (
	"context"
	"strconv"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/httpRequestExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("SnapCoreRepository")

type snapCoreRepository struct {
	config      *config.Config
	secret      *config.Secret
	logger      logger.ILogger
	httpRequest httpRequestExt.IHTTPRequest

	// Test hook for post-processing virtual account config data
	testVAConfigPostProcessor func(interface{})

	// Test hooks for QrUploadDocument error injection
	testMultipartCreateFileHook  func() error
	testMultipartCopyHook        func() error
	testMultipartCreateFieldHook func(string) error
	testMultipartWriteToHook     func(string) error
	testMultipartCloseHook       func() error
	testHTTPNewRequestHook       func() error
	testIOReadAllHook            func() error
}

func New(
	config *config.Config,
	secret *config.Secret,
	logger logger.ILogger,
	httpRequest httpRequestExt.IHTTPRequest,
) *snapCoreRepository {
	return &snapCoreRepository{
		config:      config,
		secret:      secret,
		logger:      logger,
		httpRequest: httpRequest,
	}
}

// Test hook setter methods for testing purposes
func (r *snapCoreRepository) SetTestVAConfigPostProcessor(fn func(interface{})) {
	r.testVAConfigPostProcessor = fn
}

func (r *snapCoreRepository) SetTestMultipartCreateFileHook(fn func() error) {
	r.testMultipartCreateFileHook = fn
}

func (r *snapCoreRepository) SetTestMultipartCopyHook(fn func() error) {
	r.testMultipartCopyHook = fn
}

func (r *snapCoreRepository) SetTestMultipartCreateFieldHook(fn func(string) error) {
	r.testMultipartCreateFieldHook = fn
}

func (r *snapCoreRepository) SetTestMultipartWriteToHook(fn func(string) error) {
	r.testMultipartWriteToHook = fn
}

func (r *snapCoreRepository) SetTestMultipartCloseHook(fn func() error) {
	r.testMultipartCloseHook = fn
}

func (r *snapCoreRepository) SetTestHTTPNewRequestHook(fn func() error) {
	r.testHTTPNewRequestHook = fn
}

func (r *snapCoreRepository) SetTestIOReadAllHook(fn func() error) {
	r.testIOReadAllHook = fn
}

var _ repository.IRoutingProcessorRepository = (*snapCoreRepository)(nil)
var _ repository.ISnapCoreRepository = (*snapCoreRepository)(nil)

func overridePaymentSimulationForHeader(ctx context.Context, headers map[string]string) map[string]string {
	if simulationModeValue, ok := ctx.Value(constant.CtxPaymentSimulationMode).(string); ok && simulationModeValue == strconv.FormatBool(true) {
		headers[constant.HeaderXSimulationMode] = simulationModeValue
	}
	if simulationTokenValue, ok := ctx.Value(constant.CtxPaymentSimulationToken).(string); ok && simulationTokenValue != "" {
		headers[constant.HeaderXSimulationToken] = simulationTokenValue
	}

	return headers
}
