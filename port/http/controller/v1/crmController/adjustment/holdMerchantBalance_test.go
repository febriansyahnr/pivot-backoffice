package adjustment_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	adjustmentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/adjustment"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/adjustment"
)

func TestHoldMerchantBalance(t *testing.T) {
	adjustSvcMock := serviceMocks.NewIAdjustmentService(t)

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
			modifierMock:   func() {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"invalid character 'i' looking for beginning of object key string"}`,
		},
		{
			name: "ERROR: Missing required fields",
			setupBody: func(t *testing.T) []byte {
				payload := adjustmentModel.HoldMerchantBalanceRequest{}
				b, err := json.Marshal(payload)
				assert.NoError(t, err)
				return b
			},
			modifierMock:   func() {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":{"MerchantId":"Key: 'HoldMerchantBalanceRequest.MerchantId' Error:Field validation for 'MerchantId' failed on the 'required' tag","AccountType":"Key: 'HoldMerchantBalanceRequest.AccountType' Error:Field validation for 'AccountType' failed on the 'required' tag","Type":"Key: 'HoldMerchantBalanceRequest.Type' Error:Field validation for 'Type' failed on the 'required' tag","Amount":"Key: 'HoldMerchantBalanceRequest.Amount' Error:Field validation for 'Amount' failed on the 'required' tag"}}`,
		},
		{
			name: "ERROR: HoldMerchantBalance service returns error",
			setupBody: func(t *testing.T) []byte {
				payload := adjustmentModel.HoldMerchantBalanceRequest{
					MerchantId:  uuid.NewString(),
					AccountType: "PAYMENT",
					Type:        "HOLD",
					Amount:      50000,
				}
				b, err := json.Marshal(payload)
				assert.NoError(t, err)
				return b
			},
			modifierMock: func() {
				adjustSvcMock.On(
					"HoldMerchantBalance", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("*adjustment.HoldMerchantBalanceRequest"),
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","errors":"some error"}`,
		},
		{
			name: "SUCCESS: Hold merchant balance",
			setupBody: func(t *testing.T) []byte {
				payload := adjustmentModel.HoldMerchantBalanceRequest{
					MerchantId:  uuid.NewString(),
					AccountType: "PAYMENT",
					Type:        "HOLD",
					Amount:      50000,
				}
				b, err := json.Marshal(payload)
				assert.NoError(t, err)
				return b
			},
			modifierMock: func() {
				adjustSvcMock.On(
					"HoldMerchantBalance", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("*adjustment.HoldMerchantBalanceRequest"),
				).Once().Return(&adjustmentModel.HoldMerchantBalanceResponse{
					Amount:      50000,
					MerchantID:  "merchant-123",
					AccountType: "PAYMENT",
					Type:        "HOLD",
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody: `{
				"code": "00",
				"data": {
					"amount": 50000,
					"merchantId": "merchant-123",
					"accountType": "PAYMENT",
					"type": "HOLD"
				}
			}`,
		},
		{
			name: "ERROR: ReleaseHoldedMerchantBalance service returns error",
			setupBody: func(t *testing.T) []byte {
				payload := adjustmentModel.HoldMerchantBalanceRequest{
					MerchantId:  uuid.NewString(),
					AccountType: "PAYMENT",
					Type:        "RELEASE",
					Amount:      50000,
				}
				b, err := json.Marshal(payload)
				assert.NoError(t, err)
				return b
			},
			modifierMock: func() {
				adjustSvcMock.On(
					"ReleaseHoldedMerchantBalance", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("*adjustment.HoldMerchantBalanceRequest"),
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","errors":"some error"}`,
		},
		{
			name: "SUCCESS: Release merchant balance",
			setupBody: func(t *testing.T) []byte {
				payload := adjustmentModel.HoldMerchantBalanceRequest{
					MerchantId:  uuid.NewString(),
					AccountType: "PAYMENT",
					Type:        "RELEASE",
					Amount:      50000,
				}
				b, err := json.Marshal(payload)
				assert.NoError(t, err)
				return b
			},
			modifierMock: func() {
				adjustSvcMock.On(
					"ReleaseHoldedMerchantBalance", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("*adjustment.HoldMerchantBalanceRequest"),
				).Once().Return(&adjustmentModel.HoldMerchantBalanceResponse{
					Amount:      50000,
					MerchantID:  "merchant-123",
					AccountType: "PAYMENT",
					Type:        "RELEASE",
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody: `{
				"code": "00",
				"data": {
					"amount": 50000,
					"merchantId": "merchant-123",
					"accountType": "PAYMENT",
					"type": "RELEASE"
				}
			}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.modifierMock()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/crm/v1/balances/hold", bytes.NewBuffer(test.setupBody(t)))

			router := chi.NewRouter()
			router.Post("/crm/v1/balances/hold", New(adjustSvcMock).HoldMerchantBalance)

			router.ServeHTTP(rec, req)

			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
