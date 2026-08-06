package crmDisbursementController

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
)

func TestChangeStatus(t *testing.T) {

	validTransactionID := uuid.NewString()
	validStatus := "SUCCESS"
	validReason := "Transaction completed successfully"

	validResponse := []disbursementModel.ChangeDisbursementTransactionStatusResponse{
		{
			DisbursementID: validTransactionID,
			Updated:        true,
			Reason:         "Status updated successfully",
		},
	}
	validResponseInJson, err := json.Marshal(validResponse)
	if err != nil {
		t.Fatalf("Error marshalling to JSON: %v", err)
	}

	tests := []struct {
		name           string
		setupBody      func(*testing.T) []byte
		modifierMock   func(svc *serviceMocks.IDisbursementService)
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name: "ERROR: Invalid request body",
			setupBody: func(t *testing.T) []byte {
				return []byte("{invalid json}")
			},
			modifierMock: func(svc *serviceMocks.IDisbursementService) {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"invalid character 'i' looking for beginning of object key string"}`,
		},
		{
			name: "ERROR: Validation failed - missing required fields",
			setupBody: func(t *testing.T) []byte {
				payload := disbursementModel.ChangeDisbursementTransactionStatusRequest{
					DisbursementIDS: []string{},
					Status:          "",
				}
				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			modifierMock: func(svc *serviceMocks.IDisbursementService) {
				// empty modifier - validation will fail naturally
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"42","errors":{"DisbursementIDS":"Key: 'ChangeDisbursementTransactionStatusRequest.DisbursementIDS' Error:Field validation for 'DisbursementIDS' failed on the 'min' tag","Status":"Key: 'ChangeDisbursementTransactionStatusRequest.Status' Error:Field validation for 'Status' failed on the 'required' tag"}}`,
		},
		{
			name: "ERROR: Validation failed - referenceNumber provided with multiple disbursementIds",
			setupBody: func(t *testing.T) []byte {
				payload := disbursementModel.ChangeDisbursementTransactionStatusRequest{
					DisbursementIDS: []string{validTransactionID, uuid.NewString()},
					Status:          validStatus,
					ReferenceNumber: "REF-12345",
				}
				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			modifierMock: func(svc *serviceMocks.IDisbursementService) {
				// validation fails before service is called
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"42","errors":"when referenceNumber is provided, exactly one disbursementId is allowed"}`,
		},
		{
			name: "SUCCESS: Change status with referenceNumber and single disbursementId",
			setupBody: func(t *testing.T) []byte {
				payload := disbursementModel.ChangeDisbursementTransactionStatusRequest{
					DisbursementIDS: []string{validTransactionID},
					Status:          validStatus,
					ReferenceNumber: "REF-12345",
				}
				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			modifierMock: func(svc *serviceMocks.IDisbursementService) {
				svc.On(
					"ChangeDisbursementTransactionStatus",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("disbursementModel.ChangeDisbursementTransactionStatusRequest"),
				).Return(validResponse)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":` + string(validResponseInJson) + `}`,
		},
		{
			name: "SUCCESS: Change status with all required fields",
			setupBody: func(t *testing.T) []byte {
				payload := disbursementModel.ChangeDisbursementTransactionStatusRequest{
					DisbursementIDS: []string{validTransactionID},
					Status:          validStatus,
					ReasonDescription: &validReason,
				}
				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			modifierMock: func(svc *serviceMocks.IDisbursementService) {
				svc.On(
					"ChangeDisbursementTransactionStatus",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("disbursementModel.ChangeDisbursementTransactionStatusRequest"),
				).Return(validResponse)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":` + string(validResponseInJson) + `}`,
		},
		{
			name: "SUCCESS: Change status with minimum required fields",
			setupBody: func(t *testing.T) []byte {
				reasonType := "OTHER"
				payload := disbursementModel.ChangeDisbursementTransactionStatusRequest{
					DisbursementIDS: []string{validTransactionID},
					Status:          "SUCCESS",
					ReasonType:      &reasonType,
				}
				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			modifierMock: func(svc *serviceMocks.IDisbursementService) {
				minimalResponse := []disbursementModel.ChangeDisbursementTransactionStatusResponse{
					{
						DisbursementID: validTransactionID,
						Updated:       true,
						Reason:        "Status updated successfully",
					},
				}
				svc.On(
					"ChangeDisbursementTransactionStatus",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("disbursementModel.ChangeDisbursementTransactionStatusRequest"),
				).Return(minimalResponse)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":[{"disbursementId":"` + validTransactionID + `","updated":true,"reason":"Status updated successfully"}]}`,
		},
		{
			name: "SUCCESS: Change multiple transaction statuses",
			setupBody: func(t *testing.T) []byte {
				payload := disbursementModel.ChangeDisbursementTransactionStatusRequest{
					DisbursementIDS: []string{validTransactionID},
					Status:          validStatus,
					ReasonDescription: &validReason,
				}
				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			modifierMock: func(svc *serviceMocks.IDisbursementService) {
				secondTransactionID := uuid.NewString()
				multipleResponse := []disbursementModel.ChangeDisbursementTransactionStatusResponse{
					{
						DisbursementID: validTransactionID,
						Updated:       true,
						Reason:        "Status updated successfully",
					},
					{
						DisbursementID: secondTransactionID,
						Updated:       true,
						Reason:        "Status updated successfully",
					},
				}
				svc.On(
					"ChangeDisbursementTransactionStatus",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("disbursementModel.ChangeDisbursementTransactionStatusRequest"),
				).Return(multipleResponse)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   "",
		},
		{
			name: "SUCCESS: Empty response from service",
			setupBody: func(t *testing.T) []byte {
				payload := disbursementModel.ChangeDisbursementTransactionStatusRequest{
					DisbursementIDS: []string{validTransactionID},
					Status:          validStatus,
				}
				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			modifierMock: func(svc *serviceMocks.IDisbursementService) {
				emptyResponse := []disbursementModel.ChangeDisbursementTransactionStatusResponse{}
				svc.On(
					"ChangeDisbursementTransactionStatus",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("disbursementModel.ChangeDisbursementTransactionStatusRequest"),
				).Return(emptyResponse)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":[]}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			svc := serviceMocks.NewIDisbursementService(t)
			test.modifierMock(svc)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/crm/v1/disbursements/change-transaction-status", bytes.NewBuffer(test.setupBody(t)))

			handler := &handler{
				disbursementSvc: svc,
				validator:       validatorExt.New(),
			}

			router := chi.NewRouter()
			router.Post("/crm/v1/disbursements/change-transaction-status", handler.ChangeStatus)

			router.ServeHTTP(rec, req)

			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)

			// For cases with dynamic UUIDs or empty wantRespBody, use Contains instead of JSONEq
			if test.name == "SUCCESS: Change multiple transaction statuses" || test.wantRespBody == "" {
				assert.Contains(t, rec.Body.String(), `"code":"00"`)
				assert.Contains(t, rec.Body.String(), `"data":[`)
				if test.name == "SUCCESS: Change multiple transaction statuses" {
					assert.Contains(t, rec.Body.String(), validTransactionID)
					assert.Contains(t, rec.Body.String(), "Status updated successfully")
				}
			} else {
				assert.JSONEq(t, test.wantRespBody, rec.Body.String())
			}
		})
	}
}
