package crmXbController

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"net/http"
	"net/http/httptest"

	"github.com/google/uuid"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
)

func TestGetRfiDetails(t *testing.T) {
	svc := serviceMocks.NewIXbPayoutService(t)
	validPayoutID := uuid.NewString()
	validMerchantID := uuid.NewString()

	validResponse := &xbModel.GetRfiDetailsResponse{
		Uuid:       validPayoutID,
		MerchantId: validMerchantID,
	}
	validResponseInJson, err := json.Marshal(validResponse)
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return
	}

	tests := []struct {
		name           string
		PayoutID       string
		setupBody      func(*testing.T) []byte
		modifierMock   func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:     "ERROR: Invalid PayoutID format",
			PayoutID: "invalid",
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
			name:     "ERROR: Invalid request body",
			PayoutID: validPayoutID,
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
			name:     "ERROR: Missing required params",
			PayoutID: validPayoutID,
			setupBody: func(t *testing.T) []byte {
				payload := xbModel.GetRfiDetailsRequest{
					PayoutId: validPayoutID,
				}
				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			modifierMock: func() {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"42","errors":{"MerchantId":"Key: 'GetRfiDetailsRequest.MerchantId' Error:Field validation for 'MerchantId' failed on the 'required' tag"}}`,
		},
		{
			name:     "ERROR: GetRfiDetails service error",
			PayoutID: validPayoutID,
			setupBody: func(t *testing.T) []byte {
				payload := xbModel.GetRfiDetailsRequest{
					PayoutId:   validPayoutID,
					MerchantId: validMerchantID,
				}
				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			modifierMock: func() {
				svc.On(
					"GetRfiDetails",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*xbModel.GetRfiDetailsRequest"),
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","errors":"some error"}`,
		},
		{
			name:     "SUCCESS",
			PayoutID: validPayoutID,
			setupBody: func(t *testing.T) []byte {
				payload := xbModel.GetRfiDetailsRequest{
					PayoutId:   validPayoutID,
					MerchantId: validMerchantID,
				}
				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			modifierMock: func() {
				svc.On(
					"GetRfiDetails",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*xbModel.GetRfiDetailsRequest"),
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
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/xb/payout/%s/get-rfi", test.PayoutID), bytes.NewBuffer(test.setupBody(t)))

			router := chi.NewRouter()
			router.Post("/xb/payout/{id}/get-rfi", New(svc).GetRfiDetails)

			router.ServeHTTP(rec, req)

			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
