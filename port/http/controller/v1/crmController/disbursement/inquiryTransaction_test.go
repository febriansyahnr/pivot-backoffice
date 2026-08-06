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

func TestInquiryTransaction(t *testing.T) {
	svc := serviceMocks.NewIDisbursementService(t)
	validDisbursementID := uuid.NewString()
	validMerchantID := uuid.NewString()

	validResponse := &disbursementModel.DisbursementWithTransaction{
		Disbursement: disbursementModel.Disbursement{
			UUID: validDisbursementID,
		},
	}
	validResponseInJson, err := json.Marshal(validResponse.DisbursementWithTransactionToResponse())
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return
	}

	tests := []struct {
		name           string
		disbursementID string
		setupBody      func(*testing.T) []byte
		modifierMock   func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR: Invalid disbursementID format",
			disbursementID: "invalid",
			setupBody: func(t *testing.T) []byte {
				return nil
			},
			modifierMock: func() {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"id is required"}`,
		},
		{
			name:           "ERROR: Invalid request body",
			disbursementID: validDisbursementID,
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
			name:           "ERROR: Missing required params",
			disbursementID: validDisbursementID,
			setupBody: func(t *testing.T) []byte {
				payload := disbursementModel.InquiryTransaction{
					DisbursementID: validDisbursementID,
				}
				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			modifierMock: func() {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"42","errors":{"MerchantID":"Key: 'InquiryTransaction.MerchantID' Error:Field validation for 'MerchantID' failed on the 'required' tag"}}`,
		},
		{
			name:           "ERROR: InquiryTransaction service error",
			disbursementID: validDisbursementID,
			setupBody: func(t *testing.T) []byte {
				payload := disbursementModel.InquiryTransaction{
					DisbursementID: validDisbursementID,
					MerchantID:     validMerchantID,
				}
				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			modifierMock: func() {
				svc.On(
					"InquiryTransaction",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*disbursementModel.InquiryTransaction"),
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","errors":"some error"}`,
		},
		{
			name:           "SUCCESS",
			disbursementID: validDisbursementID,
			setupBody: func(t *testing.T) []byte {
				payload := disbursementModel.InquiryTransaction{
					DisbursementID: validDisbursementID,
					MerchantID:     validMerchantID,
				}
				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			modifierMock: func() {
				svc.On(
					"InquiryTransaction",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*disbursementModel.InquiryTransaction"),
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
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/disbursements/%s/inquiry-transaction", test.disbursementID), bytes.NewBuffer(test.setupBody(t)))

			router := chi.NewRouter()
			router.Post("/disbursements/{id}/inquiry-transaction", New(svc, nil).InquiryTransaction)

			router.ServeHTTP(rec, req)

			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
