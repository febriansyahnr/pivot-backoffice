package v2InternalUnifiedPaymentController_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConst "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	loggerMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v2/internalController/unifiedPayment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUploadProofOfPayment(t *testing.T) {
	unifiedPaymentSvc := serviceMocks.NewIUnifiedPaymentService(t)
	loggerMock := loggerMocks.NewILogger(t)
	loggerMock.On("Warn", mock.Anything, mock.Anything, mock.Anything).Maybe()
	loggerMock.On("Info", mock.Anything, mock.Anything, mock.Anything).Maybe()
	loggerMock.On("Error", mock.Anything, mock.Anything, mock.Anything).Maybe()

	cfg := &config.Config{
		Investigation: config.InvestigationConfig{
			MaxFileSizeMB:   5,
			AllowedFileExts: []string{"png", "jpg", "jpeg", "pdf"},
		},
	}

	controller := New(cfg, nil, WithUnifiedPaymentService(unifiedPaymentSvc), WithLogger(loggerMock))

	tests := []struct {
		name          string
		paymentID     string
		merchantClaim *merchant.MerchantAuthTokenClaims
		setupMock     func()
		requestHeader map[string]string
		requestBody   func() (*bytes.Buffer, string)
		wantStatus    int
		wantContains  string
	}{
		{
			name:       "ERROR: Merchant not found",
			paymentID:  uuid.NewString(),
			wantStatus: http.StatusUnauthorized,
			requestBody: func() (*bytes.Buffer, string) {
				return createMultipartForm(t, "test.png", "Test reason", "dummy content")
			},
			wantContains: "merchant_not_found",
		},
		{
			name:      "ERROR: Invalid UUID",
			paymentID: "invalid-uuid-format",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "merchant-123",
			},
			requestBody: func() (*bytes.Buffer, string) {
				return createMultipartForm(t, "test.png", "Test reason", "dummy content")
			},
			wantStatus:   http.StatusBadRequest,
			wantContains: "field_format_invalid",
		},
		{
			name:      "ERROR: Missing file",
			paymentID: uuid.NewString(),
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "merchant-123",
			},
			requestBody: func() (*bytes.Buffer, string) {
				return createMultipartFormWithoutFile(t, "Test reason")
			},
			wantStatus:   http.StatusBadRequest,
			wantContains: "invalid request payload",
		},
		{
			name:      "ERROR: Missing reason",
			paymentID: uuid.NewString(),
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "merchant-123",
			},
			requestBody: func() (*bytes.Buffer, string) {
				return createMultipartFormWithoutReason(t, "test.png", "dummy content")
			},
			wantStatus:   http.StatusBadRequest,
			wantContains: "invalid request payload",
		},
		{
			name:      "ERROR: Invalid file type",
			paymentID: uuid.NewString(),
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "merchant-123",
			},
			requestBody: func() (*bytes.Buffer, string) {
				return createMultipartForm(t, "test.exe", "Test reason", "dummy content")
			},
			wantStatus:   http.StatusBadRequest,
			wantContains: "invalid file extension",
		},
		{
			name:      "ERROR: Service returns payment not found",
			paymentID: uuid.NewString(),
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "merchant-123",
			},
			setupMock: func() {
				unifiedPaymentSvc.On("UploadProofOfPayment", mock.Anything, mock.AnythingOfType("*unifiedPaymentModel.UploadProofOfPaymentRequest")).
					Return(nil, pkgErr.New(response.HttpErrNotFound, constant.ErrPaymentNotFound)).Once()
			},
			requestBody: func() (*bytes.Buffer, string) {
				return createMultipartForm(t, "test.png", "Test reason", "dummy content")
			},
			wantStatus:   http.StatusNotFound,
			wantContains: "payment not found",
		},
		{
			name:      "ERROR: Service returns investigation not enabled",
			paymentID: uuid.NewString(),
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "merchant-123",
			},
			setupMock: func() {
				unifiedPaymentSvc.On("UploadProofOfPayment", mock.Anything, mock.AnythingOfType("*unifiedPaymentModel.UploadProofOfPaymentRequest")).
					Return(nil, pkgErr.New(response.HttpErrForbidden, constant.ErrInvestigationNotEnabled)).Once()
			},
			requestBody: func() (*bytes.Buffer, string) {
				return createMultipartForm(t, "test.png", "Test reason", "dummy content")
			},
			wantStatus:   http.StatusForbidden,
			wantContains: "investigation flow is not enabled",
		},
		{
			name:      "ERROR: Service returns payment already final",
			paymentID: uuid.NewString(),
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "merchant-123",
			},
			setupMock: func() {
				unifiedPaymentSvc.On("UploadProofOfPayment", mock.Anything, mock.AnythingOfType("*unifiedPaymentModel.UploadProofOfPaymentRequest")).
					Return(nil, pkgErr.New(response.HttpErrRequest, constant.ErrPaymentAlreadyInFinalStatus)).Once()
			},
			requestBody: func() (*bytes.Buffer, string) {
				return createMultipartForm(t, "test.png", "Test reason", "dummy content")
			},
			wantStatus:   http.StatusBadRequest,
			wantContains: "final status",
		},
		{
			name:      "ERROR: Service returns bank confirmed success",
			paymentID: uuid.NewString(),
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "merchant-123",
			},
			setupMock: func() {
				unifiedPaymentSvc.On("UploadProofOfPayment", mock.Anything, mock.AnythingOfType("*unifiedPaymentModel.UploadProofOfPaymentRequest")).
					Return(nil, pkgErr.New(response.HttpErrRequest, constant.ErrBankConfirmedSuccess)).Once()
			},
			requestBody: func() (*bytes.Buffer, string) {
				return createMultipartForm(t, "test.png", "Test reason", "dummy content")
			},
			wantStatus:   http.StatusBadRequest,
			wantContains: "bank",
		},
		{
			name:      "ERROR: Service returns generic error",
			paymentID: uuid.NewString(),
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "merchant-123",
			},
			setupMock: func() {
				unifiedPaymentSvc.On("UploadProofOfPayment", mock.Anything, mock.AnythingOfType("*unifiedPaymentModel.UploadProofOfPaymentRequest")).
					Return(nil, errors.New("internal error")).Once()
			},
			requestBody: func() (*bytes.Buffer, string) {
				return createMultipartForm(t, "test.png", "Test reason", "dummy content")
			},
			wantStatus:   http.StatusInternalServerError,
			wantContains: "general_error",
		},
		{
			name:      "SUCCESS",
			paymentID: uuid.NewString(),
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "merchant-123",
			},
			setupMock: func() {
				unifiedPaymentSvc.On("UploadProofOfPayment", mock.Anything, mock.AnythingOfType("*unifiedPaymentModel.UploadProofOfPaymentRequest")).
					Return(&unifiedPaymentModel.UploadProofOfPaymentResponse{
						PaymentID:           uuid.NewString(),
						Status:              constant.ChargeStatusSuccess,
						InvestigationStatus: paymentConst.InvestigationStatusInProcess,
						CreatedAt:           time.Now().UTC(),
						UpdatedAt:           time.Now().UTC(),
					}, nil).Once()
			},
			requestBody: func() (*bytes.Buffer, string) {
				return createMultipartForm(t, "test.png", "Customer showed screenshot", "dummy content")
			},
			wantStatus:   http.StatusOK,
			wantContains: `"code":"00"`,
		},
		{
			name:      "SUCCESS: With sub-merchant ID",
			paymentID: uuid.NewString(),
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "parent-merchant-123",
			},
			requestHeader: map[string]string{
				constant.HeaderXSubMerchantID: "sub-merchant-456",
			},
			setupMock: func() {
				unifiedPaymentSvc.On("UploadProofOfPayment", mock.Anything, mock.MatchedBy(func(req *unifiedPaymentModel.UploadProofOfPaymentRequest) bool {
					return req.MerchantID == "sub-merchant-456"
				})).
					Return(&unifiedPaymentModel.UploadProofOfPaymentResponse{
						PaymentID:           uuid.NewString(),
						Status:              constant.ChargeStatusSuccess,
						InvestigationStatus: paymentConst.InvestigationStatusInProcess,
						CreatedAt:           time.Now().UTC(),
						UpdatedAt:           time.Now().UTC(),
					}, nil).Once()
			},
			requestBody: func() (*bytes.Buffer, string) {
				return createMultipartForm(t, "test.jpg", "Customer showed screenshot", "dummy content")
			},
			wantStatus:   http.StatusOK,
			wantContains: `"code":"00"`,
		},
		{
			name:      "SUCCESS: With PDF file",
			paymentID: uuid.NewString(),
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "merchant-123",
			},
			setupMock: func() {
				unifiedPaymentSvc.On("UploadProofOfPayment", mock.Anything, mock.AnythingOfType("*unifiedPaymentModel.UploadProofOfPaymentRequest")).
					Return(&unifiedPaymentModel.UploadProofOfPaymentResponse{
						PaymentID:           uuid.NewString(),
						Status:              constant.ChargeStatusSuccess,
						InvestigationStatus: paymentConst.InvestigationStatusInProcess,
						CreatedAt:           time.Now().UTC(),
						UpdatedAt:           time.Now().UTC(),
					}, nil).Once()
			},
			requestBody: func() (*bytes.Buffer, string) {
				return createMultipartForm(t, "receipt.pdf", "Bank transfer receipt", "%PDF-1.4 dummy pdf content")
			},
			wantStatus:   http.StatusOK,
			wantContains: `"code":"00"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			body, contentType := tt.requestBody()
			req := httptest.NewRequest(http.MethodPost, "/v2/payments/"+tt.paymentID+"/investigation/proof-of-payment", body)
			req.Header.Set("Content-Type", contentType)

			for key, value := range tt.requestHeader {
				req.Header.Set(key, value)
			}

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("uuid", tt.paymentID)

			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
			if tt.merchantClaim != nil {
				ctx = context.WithValue(ctx, constant.CtxMerchantInfo, tt.merchantClaim)
			}
			req = req.WithContext(ctx)

			rec := httptest.NewRecorder()
			controller.UploadProofOfPayment(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.wantContains)
		})
	}
}

func createMultipartForm(t *testing.T, filename, reason, content string) (*bytes.Buffer, string) {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("proofOfTransaction", filename)
	assert.NoError(t, err)
	_, err = io.WriteString(part, content)
	assert.NoError(t, err)

	err = writer.WriteField("reason", reason)
	assert.NoError(t, err)

	err = writer.Close()
	assert.NoError(t, err)

	return body, writer.FormDataContentType()
}

func createMultipartFormWithoutFile(t *testing.T, reason string) (*bytes.Buffer, string) {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	err := writer.WriteField("reason", reason)
	assert.NoError(t, err)

	err = writer.Close()
	assert.NoError(t, err)

	return body, writer.FormDataContentType()
}

func createMultipartFormWithoutReason(t *testing.T, filename, content string) (*bytes.Buffer, string) {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("proofOfTransaction", filename)
	assert.NoError(t, err)
	_, err = io.WriteString(part, content)
	assert.NoError(t, err)

	err = writer.Close()
	assert.NoError(t, err)

	return body, writer.FormDataContentType()
}

func TestUploadProofOfPaymentFileTooLarge(t *testing.T) {
	unifiedPaymentSvc := serviceMocks.NewIUnifiedPaymentService(t)
	loggerMock := loggerMocks.NewILogger(t)
	loggerMock.On("Warn", mock.Anything, mock.Anything, mock.Anything).Maybe()

	cfg := &config.Config{
		Investigation: config.InvestigationConfig{
			MaxFileSizeMB:   1,
			AllowedFileExts: []string{"png", "jpg", "jpeg", "pdf"},
		},
	}

	controller := New(cfg, nil, WithUnifiedPaymentService(unifiedPaymentSvc), WithLogger(loggerMock))

	paymentID := uuid.NewString()
	merchantClaim := &merchant.MerchantAuthTokenClaims{
		MerchantId: "merchant-123",
	}

	largeContent := make([]byte, 2*1024*1024)
	for i := range largeContent {
		largeContent[i] = 'a'
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("proofOfTransaction", "large.png")
	assert.NoError(t, err)
	_, err = part.Write(largeContent)
	assert.NoError(t, err)

	err = writer.WriteField("reason", "Test reason")
	assert.NoError(t, err)

	err = writer.Close()
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v2/payments/"+paymentID+"/investigation/proof-of-payment", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("uuid", paymentID)

	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, constant.CtxMerchantInfo, merchantClaim)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	controller.UploadProofOfPayment(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "file size too large")
}
