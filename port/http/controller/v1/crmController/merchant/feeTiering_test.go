package merchant_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/merchant"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpdateFeeTieringConfig(t *testing.T) {
	validator := validator.New()
	merchantSvc := mocks.NewIMerchantService(t)

	handler := New(merchantSvc, nil, validator, nil)

	router := chi.NewRouter()
	router.Patch("/merchants/fee/{id}/tiers", handler.UpdateFeeTieringConfig)

	feeId := uuid.NewString()
	merchantId := uuid.NewString()
	requestBody := fmt.Sprintf(`{"merchantId": "%s","type": "TPV","configs": [{"tier": 1,"min": 1,"max": 10,"amountType": "AMOUNT","amount": 4000,"percentage": 0,"taxType": "NON_PKP","taxPercentage": 0}]}`, merchantId)

	response := merchant.FeeTieringResponse{
		MerchantId: merchantId,
		Reference:  c.ReferenceDisbursement,
		Type:       "TPV",
		Configs: []merchant.FeeTieringConfig{
			{
				Tier:       1,
				Min:        1,
				Max:        10,
				AmountType: "AMOUNT",
				Amount:     4_000,
				TaxType:    "NON_PKP",
			},
		},
	}
	rawResponse, err := json.Marshal(response)
	require.NoError(t, err)

	tests := []struct {
		name           string
		feeId          string
		bodyRequest    string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:Malformed request body",
			bodyRequest:    "A",
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrRespForTest(40, c.ErrInvalidRequestPayload.Error()),
		},
		{
			name:           "ERROR:Merchant fee ID invalid format",
			feeId:          "123456",
			bodyRequest:    requestBody,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":{"FeeId":"Key: 'FeeTieringRequest.FeeId' Error:Field validation for 'FeeId' failed on the 'uuid' tag"}}`,
		},
		{
			name:        "ERROR:Some error", // NOSONAR
			bodyRequest: requestBody,
			setupMock: func() {
				merchantSvc.On(
					"UpdateFeeTieringConfig", c.ValueCtxMockType(), mock.AnythingOfType("*merchant.FeeTieringRequest"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrRespForTest(99, "some error"),
		},
		{
			name:        "SUCCESS", // NOSONAR
			bodyRequest: requestBody,
			setupMock: func() {
				merchantSvc.On(
					"UpdateFeeTieringConfig", c.ValueCtxMockType(), mock.AnythingOfType("*merchant.FeeTieringRequest"),
				).Return(&response, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   fmt.Sprintf(`{"code":"00","data":%s}`, string(rawResponse)),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.feeId == "" {
				test.feeId = feeId
			}
			if test.setupMock != nil {
				test.setupMock()
			}
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/merchants/fee/%s/tiers", test.feeId), strings.NewReader(test.bodyRequest))

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Output:", rec.Body.String())
			}
		})
	}
}
