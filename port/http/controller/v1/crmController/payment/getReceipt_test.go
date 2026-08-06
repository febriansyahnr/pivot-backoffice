package v1CrmPaymentController

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetReceipt(t *testing.T) {
	svc := serviceMocks.NewIPaymentService(t)
	h := New(svc)

	validReferenceID := "REF123456"
	validMerchantID := uuid.NewString()
	validPDFContent := []byte("%PDF-1.4 test pdf content")
	validFilename := "receipt.pdf"

	tests := []struct {
		name           string
		requestBody    interface{}
		mockService    func()
		wantStatusCode int
		wantHeaders    map[string]string
		wantBodyCheck  func(t *testing.T, body []byte)
	}{
		{
			name:           "ERROR: Invalid JSON payload",
			requestBody:    "invalid-json",
			mockService:    func() {},
			wantStatusCode: http.StatusBadRequest,
			wantBodyCheck: func(t *testing.T, body []byte) {
				assert.Contains(t, string(body), `"code":"40"`)
			},
		},
		{
			name: "ERROR: Validation failed - missing referenceId",
			requestBody: paymentModel.GetPaymentReceiptCRMRequest{
				MerchantID: validMerchantID,
			},
			mockService:    func() {},
			wantStatusCode: http.StatusBadRequest,
			wantBodyCheck: func(t *testing.T, body []byte) {
				assert.Contains(t, string(body), `"code":"42"`)
				assert.Contains(t, string(body), "ReferenceID")
			},
		},
		{
			name: "ERROR: Validation failed - missing merchantId",
			requestBody: paymentModel.GetPaymentReceiptCRMRequest{
				ReferenceID: validReferenceID,
			},
			mockService:    func() {},
			wantStatusCode: http.StatusBadRequest,
			wantBodyCheck: func(t *testing.T, body []byte) {
				assert.Contains(t, string(body), `"code":"42"`)
				assert.Contains(t, string(body), "MerchantID")
			},
		},
		{
			name: "ERROR: Validation failed - invalid merchantId format",
			requestBody: paymentModel.GetPaymentReceiptCRMRequest{
				ReferenceID: validReferenceID,
				MerchantID:  "invalid-uuid",
			},
			mockService:    func() {},
			wantStatusCode: http.StatusBadRequest,
			wantBodyCheck: func(t *testing.T, body []byte) {
				assert.Contains(t, string(body), `"code":"42"`)
				assert.Contains(t, string(body), "MerchantID")
			},
		},
		{
			name: "ERROR: Service error",
			requestBody: paymentModel.GetPaymentReceiptCRMRequest{
				ReferenceID: validReferenceID,
				MerchantID:  validMerchantID,
			},
			mockService: func() {
				svc.On("GetReceiptByID", mock.Anything, mock.MatchedBy(func(req *paymentModel.GetPaymentReceiptRequest) bool {
					return req.ReferenceID == validReferenceID && req.MerchantID == validMerchantID
				})).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBodyCheck: func(t *testing.T, body []byte) {
				assert.Contains(t, string(body), `"code":"99"`)
			},
		},
		{
			name: "SUCCESS: Valid request returns PDF",
			requestBody: paymentModel.GetPaymentReceiptCRMRequest{
				ReferenceID: validReferenceID,
				MerchantID:  validMerchantID,
			},
			mockService: func() {
				svc.On("GetReceiptByID", mock.Anything, mock.MatchedBy(func(req *paymentModel.GetPaymentReceiptRequest) bool {
					return req.ReferenceID == validReferenceID && req.MerchantID == validMerchantID
				})).Once().Return(&paymentModel.GetPaymentReceiptResponse{
					Filename: validFilename,
					PDF:      validPDFContent,
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantHeaders: map[string]string{
				"Content-Type":        "application/pdf",
				"Content-Disposition": `attachment; filename="receipt.pdf"`,
			},
			wantBodyCheck: func(t *testing.T, body []byte) {
				assert.Equal(t, validPDFContent, body)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.mockService()

			var body []byte
			if str, ok := test.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(test.requestBody)
			}

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/payments/receipt", bytes.NewReader(body))

			router := chi.NewRouter()
			router.Post("/payments/receipt", h.GetReceipt)
			router.ServeHTTP(rec, req)

			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)

			if test.wantHeaders != nil {
				for key, value := range test.wantHeaders {
					assert.Equal(t, value, rec.Header().Get(key))
				}
			}

			if test.wantBodyCheck != nil {
				test.wantBodyCheck(t, rec.Body.Bytes())
			}
		})
	}
}
