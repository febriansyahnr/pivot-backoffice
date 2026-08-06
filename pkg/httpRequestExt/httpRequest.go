package httpRequestExt

import (
	"context"
	"encoding/json"
	"maps"
	"mime/multipart"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/logger/encoder"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	pdkLogger "github.com/paper-indonesia/pdk/v2/logger"
)

type IHTTPRequest interface {
	GET(ctx context.Context, uri string, header map[string]string) ([]byte, int, error)
	POST(ctx context.Context, uri string, data interface{}, header map[string]string) ([]byte, int, error)
	PUT(ctx context.Context, uri string, data interface{}, header map[string]string) ([]byte, int, error)
	PATCH(ctx context.Context, uri string, data interface{}, header map[string]string) ([]byte, int, error)
	DELETE(ctx context.Context, uri string, data interface{}, header map[string]string) ([]byte, int, error)
	POSTWithFormData(ctx context.Context, uri string, formData map[string]string, files map[string]*multipart.FileHeader, header map[string]string) ([]byte, int, error)
}

type HTTPRequest struct {
	outbound  repository.IOutboundRepository
	skipper   func(url string) bool
	logger    pdkLogger.ILogger
	inspector encoder.Inspector
}

type HTTPRequestFunc func(*HTTPRequest)

func WithOutbound(outbound repository.IOutboundRepository) HTTPRequestFunc {
	return func(h *HTTPRequest) {
		h.outbound = outbound
	}
}

func WithSkipper(s func(url string) bool) HTTPRequestFunc {
	return func(h *HTTPRequest) {
		h.skipper = s
	}
}

func WithLogger(log pdkLogger.ILogger) HTTPRequestFunc {
	return func(h *HTTPRequest) {
		h.logger = log
	}
}

func WithMaskingSensitiveData(fields []string) HTTPRequestFunc {
	return func(h *HTTPRequest) {
		h.inspector = encoder.NewInspector(fields)
	}
}

func New(opts ...HTTPRequestFunc) IHTTPRequest {
	r := &HTTPRequest{
		inspector: encoder.NewInspector(nil),
	}

	for _, opt := range opts {
		opt(r)
	}
	return r
}

// POST implements IHTTPRequest.
func (r *HTTPRequest) POST(ctx context.Context, uri string, data interface{}, header map[string]string) ([]byte, int, error) {
	return r.Do(ctx, "POST", uri, data, header)
}

// GET implements IHTTPRequest.
func (r *HTTPRequest) GET(ctx context.Context, uri string, header map[string]string) ([]byte, int, error) {
	return r.Do(ctx, "GET", uri, nil, header)
}

// PUT implements IHTTPRequest.
func (r *HTTPRequest) PUT(ctx context.Context, uri string, data interface{}, header map[string]string) ([]byte, int, error) {
	return r.Do(ctx, "PUT", uri, data, header)
}

// PATCH implements IHTTPRequest.
func (r *HTTPRequest) PATCH(ctx context.Context, uri string, data interface{}, header map[string]string) ([]byte, int, error) {
	return r.Do(ctx, "PATCH", uri, data, header)
}

// DELETE implements IHTTPRequest.
func (r *HTTPRequest) DELETE(ctx context.Context, uri string, data interface{}, header map[string]string) ([]byte, int, error) {
	return r.Do(ctx, "DELETE", uri, data, header)
}

// POSTWithFormData implements IHTTPRequest.
func (r *HTTPRequest) POSTWithFormData(ctx context.Context, uri string, formData map[string]string, files map[string]*multipart.FileHeader, header map[string]string) (respBody []byte, statusCode int, err error) {
	method := http.MethodPost
	if r.skipper != nil && r.skipper(uri) {
		return httputil.RequestHitAPIWithFormData(ctx, method, uri, formData, files, header)
	}

	client, _ := ctx.Value(constant.CtxClientReqKey).(*outbound.Client)
	if client == nil {
		client = &outbound.Client{}
	}
	if client.RequestId, _ = ctx.Value(pdkConst.CtxRequestIdKey).(string); client.RequestId == "" {
		id, _ := uuid.NewV7()
		client.RequestId = id.String()
	}

	now := time.Now().UTC()
	defer func() {
		id, _ := uuid.NewV7()

		outboundReq := &outbound.OutboundRequest{
			Id:           id.String(),
			Client:       client,
			Date:         now,
			Method:       method,
			URL:          uri,
			Headers:      r.inspector.Inspects(header).(map[string]string),
			Body:         r.inspector.Inspects(formData),
			StatusCode:   statusCode,
			ResponseTime: time.Now().UTC().Sub(now).String(),
			ResponseBody: respBody,
			Error:        err,
		}
		go func() {
			newCtx := context.WithValue(
				context.Background(), pdkConst.CtxRequestIdKey, outboundReq.Client.RequestId,
			)

			if e := r.outbound.Insert(newCtx, outboundReq); e != nil && r.logger != nil {
				r.logger.Error(newCtx, "insert outbound request", pdkLogger.Error(err))
			}
		}()
	}()
	return httputil.RequestHitAPIWithFormData(ctx, method, uri, formData, files, header)
}

func (r *HTTPRequest) Do(ctx context.Context, method, uri string, data interface{}, headers map[string]string) ([]byte, int, error) {
	var xRequestID string
	if xRequestID, _ = ctx.Value(pdkConst.CtxRequestIdKey).(string); xRequestID == "" {
		id, _ := uuid.NewV7()
		xRequestID = id.String()

		ctx = context.WithValue(ctx, pdkConst.CtxRequestIdKey, xRequestID)
	}

	headers["X-Request-Id"] = xRequestID

	if r.outbound != nil {
		return r.RequestWithOutbound(ctx, method, uri, data, headers)
	}
	return httputil.RequestHitAPI(ctx, method, uri, data, headers)
}

func (r *HTTPRequest) RequestWithOutbound(ctx context.Context, method, url string, data interface{}, headers map[string]string) (respBody []byte, statusCode int, err error) {
	defer func() {
		if re := recover(); re != nil {
			r.logger.Error(ctx, "Panic recovery from invoked RequestWithOutbound", pdkLogger.Any("detail", re))
		}
	}()

	if r.skipper != nil && r.skipper(url) {
		return httputil.RequestHitAPI(ctx, method, url, data, headers)
	}

	client, _ := ctx.Value(constant.CtxClientReqKey).(*outbound.Client)
	if client == nil {
		client = &outbound.Client{}
	}
	if client.RequestId, _ = ctx.Value(pdkConst.CtxRequestIdKey).(string); client.RequestId == "" {
		id, _ := uuid.NewV7()
		client.RequestId = id.String()
	}

	now := time.Now().UTC()
	wait := make(chan struct{}, 1)
	isWait, _ := ctx.Value(constant.CtxSyncKey).(bool)
	defer func() {
		id, _ := ctx.Value(constant.CtxOutboundID).(uuid.UUID)
		if id == uuid.Nil {
			id, _ = uuid.NewV7()
		}

		outboundReq := &outbound.OutboundRequest{
			Id:           id.String(),
			Client:       client,
			Date:         now,
			Method:       method,
			URL:          url,
			Headers:      r.inspector.Inspects(headers).(map[string]string),
			Body:         r.inspector.Inspects(data),
			StatusCode:   statusCode,
			ResponseTime: time.Now().UTC().Sub(now).String(),
			ResponseBody: respBody,
			Error:        err,
		}
		go func() {
			newCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			newCtx = context.WithValue(
				newCtx, pdkConst.CtxRequestIdKey, outboundReq.Client.RequestId,
			)

			if !json.Valid(outboundReq.ResponseBody) {
				outboundReq.ResponseBody, _ = json.Marshal(map[string]string{
					"wrappingReason": "Invalid Response Format", "content": string(outboundReq.ResponseBody),
				})
			}
			if e := r.outbound.Insert(newCtx, outboundReq); e != nil && r.logger != nil {
				r.logger.Error(newCtx, "insert outbound request", pdkLogger.Error(e))
			}

			wait <- struct{}{}
		}()

		if isWait {
			// waiting for response outbound created
			<-wait
		}
	}()
	return httputil.RequestHitAPI(ctx, method, url, data, headers)
}

func (r *HTTPRequest) MaskingData(fields []string, src map[string]string) map[string]string {
	if fields == nil {
		fields = defMaskingField
	} else {
		fields = append(fields, defMaskingField...)
	}

	dst := map[string]string{}

	maps.Copy(dst, src)

	for key, val := range dst {
		ok := slices.ContainsFunc(fields, func(field string) bool {
			return strings.EqualFold(key, field)
		})
		if ok {
			dst[key] = strings.Repeat("*", len(val))
		}
	}
	return dst
}
