package adjustment_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/adjustment"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"

	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/adjustment"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateAdjustmentFromManualTopup(t *testing.T) {
	adjustSvcMock := serviceMocks.NewIAdjustmentService(t)
	uniqueID := "66eb9f9e-2b7b-4267-9666-589e051cfa08"
	notes := "This is notes"

	tests := []struct {
		name           string
		id             string
		setupBody      func(*testing.T) []byte
		modifierMock   func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name: "ERROR: Invalid ID",
			setupBody: func(t *testing.T) []byte {
				return nil
			},
			modifierMock: func() {
				// empty modifier
			},
			id:             "invalid",
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"id is required"}`,
		},
		{
			name: "ERROR: Invalid request body",
			setupBody: func(t *testing.T) []byte {
				return []byte("{invalid json}")
			},
			modifierMock: func() {
				// empty modifier
			},
			id:             uuid.NewString(),
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"invalid character 'i' looking for beginning of object key string"}`,
		},
		{
			name: "ERROR: Missing required request",
			setupBody: func(t *testing.T) []byte {
				payload := adjustment.BalanceAdjustmentRequest{
					Currency:  "IDR",
					Amount:    1000000,
					CreatedBy: "John",
					Notes:     notes,
				}
				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			modifierMock: func() {
				// empty modifier
			},
			id:             uuid.NewString(),
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":{"MerchantID":"Key: 'BalanceAdjustmentRequest.MerchantID' Error:Field validation for 'MerchantID' failed on the 'required' tag"}}`,
		},
		{
			name: "ERROR: CreateBalanceAdjustmentFromManualTopUp service",
			setupBody: func(t *testing.T) []byte {
				payload := adjustment.BalanceAdjustmentRequest{
					MerchantID: uuid.NewString(),
					Currency:   "IDR",
					Amount:     1000000,
					CreatedBy:  "John",
					Notes:      notes,
				}
				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			modifierMock: func() {
				adjustSvcMock.On(
					"CreateBalanceAdjustmentFromManualTopUp", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("*adjustment.BalanceAdjustmentRequest"),
				).Once().Return("", constant.ErrSomeErrorForUnitTest)
			},
			id:             uuid.NewString(),
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","errors":"some error"}`,
		},
		{
			name: "SUCCESS",
			setupBody: func(t *testing.T) []byte {
				payload := adjustment.BalanceAdjustmentRequest{
					MerchantID: uuid.NewString(),
					Currency:   "IDR",
					Amount:     1000000,
					CreatedBy:  "John",
					Notes:      notes,
				}
				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			modifierMock: func() {
				adjustSvcMock.On(
					"CreateBalanceAdjustmentFromManualTopUp", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("*adjustment.BalanceAdjustmentRequest"),
				).Return(uniqueID, nil)
			},
			id:             uuid.NewString(),
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":{"id": "` + uniqueID + `"}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.modifierMock()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/balances/topup/manual/%s/adjustment", test.id), bytes.NewBuffer(test.setupBody(t)))

			router := chi.NewRouter()
			router.Post("/balances/topup/manual/{id}/adjustment", New(adjustSvcMock).CreateAdjustmentFromManualTopup)

			router.ServeHTTP(rec, req)

			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
