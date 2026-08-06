package crmDisbursementController

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	"github.com/stretchr/testify/mock"

	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
)

func TestGetReceipt(t *testing.T) {
	svc := serviceMocks.NewIDisbursementService(t)
	validDisbursementID := uuid.NewString()
	validMerchantID := uuid.NewString()

	validResponse := &disbursementModel.GetDisbursementReceiptResponse{
		ReceiptURL: "url://receipt-downloadable-url",
	}
	validResponseInJson, err := json.Marshal(validResponse)
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return
	}

	tests := []struct {
		name           string
		setupBody      func(*testing.T) []byte
		modifierMock   func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name: "ERROR: Invalid request body",
			setupBody: func(t *testing.T) []byte {
				return []byte("{invalid json}")
			},
			modifierMock: func() {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"invalid character 'i' looking for beginning of object key string"}`,
		},
		{
			name: "ERROR: Failed Validation",
			setupBody: func(t *testing.T) []byte {
				payload := disbursementModel.GetDisbursementReceiptCRMRequest{
					ReferenceID: validDisbursementID,
				}
				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			modifierMock: func() {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"42","errors":{"MerchantID":"Key: 'GetDisbursementReceiptCRMRequest.MerchantID' Error:Field validation for 'MerchantID' failed on the 'required' tag"}}`,
		},
		{
			name: "ERROR: InquiryTransaction service error",
			setupBody: func(t *testing.T) []byte {
				payload := disbursementModel.GetDisbursementReceiptCRMRequest{
					ReferenceID: validDisbursementID,
					MerchantID:  validMerchantID,
				}
				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			modifierMock: func() {
				svc.On(
					"GetReceiptByID",
					constant.ValueCtxMockType(),
					mock.Anything,
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","errors":"some error"}`,
		},
		{
			name: "SUCCESS",
			setupBody: func(t *testing.T) []byte {
				payload := disbursementModel.GetDisbursementReceiptCRMRequest{
					ReferenceID: validDisbursementID,
					MerchantID:  validMerchantID,
				}
				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			modifierMock: func() {
				svc.On(
					"GetReceiptByID",
					constant.ValueCtxMockType(),
					mock.Anything,
				).Return(validResponse, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":` + string(validResponseInJson) + `}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.modifierMock()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/disbursements/receipt", bytes.NewBuffer(test.setupBody(t)))

			router := chi.NewRouter()
			router.Post("/disbursements/receipt", New(svc, nil).GetReceipt)

			router.ServeHTTP(rec, req)

			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
