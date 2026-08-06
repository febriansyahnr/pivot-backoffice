package internalXbController_test

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	funcMock "github.com/paper-indonesia/pivot-backoffice/mocks/func"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/internalController/xb"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUploadUnderlyingDocument(t *testing.T) {
	cfg := &config.Config{}
	xbPayoutSvc := serviceMock.NewIXbPayoutService(t)
	logger := logger.NewSlogger(logger.Config{})
	svc := New(cfg, WithXbPayoutService(xbPayoutSvc), WithLogger(logger))

	validPayoutId := uuid.NewString()
	postForm := func(req *http.Request) {
		*req = *req.WithContext(context.WithValue(req.Context(), c.CtxMerchantInfo, &merchantModel.MerchantAuthTokenClaims{
			MerchantId: "12345",
		}))

		req.PostForm = url.Values{}
	}

	tests := []struct {
		name             string
		mockSetup        func()
		reqSetting       func(r *http.Request)
		payoutId         string
		fileName         string
		fileContent      string
		expectedStatus   int
		expectedRespBody string
	}{
		{
			name: "ERROR: Invalid merchant info",
			mockSetup: func() {
				// empty modifier
			},
			expectedStatus:   http.StatusUnauthorized,
			expectedRespBody: `{"code":"merchant_not_found","error":{"details":[{"field":"","message":"Invalid Merchant request"}],"traceId":"","type":"API_ERROR"},"message":"Merchant not found"}`,
		},
		{
			name: "ERROR: Invalid id",
			mockSetup: func() {
				// empty modifier
			},
			reqSetting: func(r *http.Request) {
				postForm(r)
			},
			payoutId:         "invalid",
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"field_required","error":{"details":[{"field":"id","message":"Make sure id value is fulfilled"}],"traceId":"","type":"API_ERROR"},"message":"Mandatory field is missing"}`,
		},
		{
			name: "ERROR: Invalid document file",
			mockSetup: func() {
				// empty mock setup
			},
			reqSetting: func(r *http.Request) {
				postForm(r)
				r.MultipartForm = &multipart.Form{}
			},
			payoutId:         validPayoutId,
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
		{
			name: "ERROR: Document file not supported",
			mockSetup: func() {
				// empty mock setup
			},
			reqSetting: func(r *http.Request) {
				postForm(r)
			},
			fileName:         "file-test.png",
			payoutId:         validPayoutId,
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
		{
			name: "ERROR: UploadUnderlyingDocument service error",
			mockSetup: func() {
				xbPayoutSvc.On("UploadUnderlyingDocument",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbModel.UploadUnderlyingDocumentRequest"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			reqSetting: func(r *http.Request) {
				postForm(r)
			},
			payoutId:         validPayoutId,
			expectedStatus:   http.StatusInternalServerError,
			expectedRespBody: `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
		{
			name: "SUCCESS",
			mockSetup: func() {
				xbPayoutSvc.On("UploadUnderlyingDocument",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbModel.UploadUnderlyingDocumentRequest"),
				).Return(&xbModel.UploadUnderlyingDocumentResponse{}, nil)
			},
			reqSetting: func(r *http.Request) {
				postForm(r)
			},
			payoutId:         validPayoutId,
			expectedStatus:   http.StatusOK,
			expectedRespBody: `{"code":"00", "data":{"documentReference":""}, "message":"Success"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			if tt.fileName == "" {
				tt.fileName = "test-file.zip"
			}
			body, contentType, err := funcMock.CreateMultipartFormFile("document", tt.fileName, tt.fileContent)

			require.NoError(t, err)
			require.NotEqual(t, "", contentType)
			defer body.Reset()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/open-api/v1/xb/create-payout-session/%s/upload", tt.payoutId), body)
			req.Header.Set("Content-Type", contentType)

			if tt.reqSetting != nil {
				tt.reqSetting(req)
			}

			router := chi.NewRouter()
			router.Post("/open-api/v1/xb/create-payout-session/{id}/upload", svc.UploadUnderlyingDocument)

			router.ServeHTTP(rec, req)

			require.Equal(t, tt.expectedStatus, rec.Result().StatusCode)
			assert.JSONEq(t, tt.expectedRespBody, rec.Body.String())
		})
	}
}
