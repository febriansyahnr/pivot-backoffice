package xbPayoutController_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/xbPayout"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUploadUnderlyingDocument(t *testing.T) {
	const (
		testDocumentPDF = "document.pdf"
		testContent     = "test content"
	)

	cfg := &config.Config{}
	xbPayoutSvc := serviceMock.NewIXbPayoutService(t)

	ctrl := New(cfg, WithXbPayoutService(xbPayoutSvc))

	validRequestID := func(req *http.Request) {
		*req = *req.WithContext(context.WithValue(req.Context(), c.CtxUserInfoKey, &userModel.UserTokenClaims{
			MerchantId: "12345",
		}))
	}

	createMultipartRequest := func(payoutID string, filename string, content string) *http.Request {
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)

		part, err := writer.CreateFormFile("document", filename)
		if err != nil {
			t.Fatal(err)
		}

		if _, err := io.WriteString(part, content); err != nil {
			t.Fatal(err)
		}

		writer.Close()

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/xb/payout/%s/upload", payoutID), &body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		return req
	}

	tests := []struct {
		name             string
		payoutID         string
		filename         string
		fileContent      string
		mockSetup        func()
		reqSetting       func(r *http.Request)
		expectedStatus   int
		expectedRespBody string
	}{
		{
			name: "ERROR: Invalid user info",
			mockSetup: func() {
				// empty modifier
			},
			expectedStatus:   http.StatusUnauthorized,
			expectedRespBody: `{"code":"41", "data":null, "error":{"details":[], "traceId":"", "type":"API_ERROR"}, "message":"user not found"}`,
		},
		{
			name:        "ERROR: Invalid payout ID",
			payoutID:    "invalid-uuid",
			filename:    testDocumentPDF,
			fileContent: testContent,
			mockSetup: func() {
				// empty modifier
			},
			reqSetting:       validRequestID,
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"42", "data":null, "error":{"details":[], "traceId":"", "type":"API_VALIDATION_ERROR"}, "message":"id is required"}`,
		},
		{
			name:        "ERROR: Upload service error",
			payoutID:    uuid.NewString(),
			filename:    testDocumentPDF,
			fileContent: testContent,
			mockSetup: func() {
				xbPayoutSvc.On("UploadUnderlyingDocument",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbModel.UploadUnderlyingDocumentRequest"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			reqSetting:       validRequestID,
			expectedStatus:   http.StatusInternalServerError,
			expectedRespBody: `{"code":"99", "data":null, "error":{"details":[], "traceId":"", "type":"UNKNOWN"}, "message":"some error"}`,
		},
		{
			name:        "SUCCESS",
			payoutID:    uuid.NewString(),
			filename:    testDocumentPDF,
			fileContent: testContent,
			mockSetup: func() {
				xbPayoutSvc.On("UploadUnderlyingDocument",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbModel.UploadUnderlyingDocumentRequest"),
				).Once().Return(&xbModel.UploadUnderlyingDocumentResponse{}, nil)
			},
			reqSetting:       validRequestID,
			expectedStatus:   http.StatusOK,
			expectedRespBody: `{"code":"00", "data":{"documentReference":""}, "message":"OK"}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			var req *http.Request
			if tc.name == "ERROR: Invalid user info" {
				req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/xb/payout/%s/upload", uuid.NewString()), nil)
			} else {
				req = createMultipartRequest(tc.payoutID, tc.filename, tc.fileContent)
			}

			if tc.reqSetting != nil {
				tc.reqSetting(req)
			}

			rec := httptest.NewRecorder()

			router := chi.NewRouter()
			router.Post("/api/v1/xb/payout/{id}/upload", ctrl.UploadUnderlyingDocument)

			router.ServeHTTP(rec, req)

			require.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
			assert.JSONEq(t, tc.expectedRespBody, rec.Body.String())
		})
	}
}
