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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateFeeConfigOnBehalf(t *testing.T) {
	validator := validator.New()
	merchantSvc := mocks.NewIMerchantService(t)

	handler := New(merchantSvc, nil, validator, nil)

	router := chi.NewRouter()
	router.Post("/fee-on-behalf", handler.CreateFeeConfigOnBehalf)

	feeId := "0192f6c3-5d09-7ced-ad1e-f79a222fa8ab"
	merchantId := "58e57e0d-0bb8-482f-9b9f-8608ed4a9b0e"

	requestFmt := `{"merchantId": "%s", "type": "ALL", "subMerchantId": null, "reference": "DISBURSEMENT", "paymentMethod": null, "amountType": "AMOUNT", "amount": 2000, "percentage": 0}`
	requestBody := fmt.Sprintf(requestFmt, merchantId)
	invalidRequestBody := fmt.Sprintf(requestFmt, "ABC")

	response := &merchant.CreateFeeConfigOnBehalfResponse{
		Id: feeId,
		CreateFeeConfigOnBehalfRequest: &merchant.CreateFeeConfigOnBehalfRequest{
			MerchantId: merchantId,
			Type:       c.FeeOnBehalfTypeAll,
			Reference:  c.ReferenceDisbursement,
			AmountType: c.MerchantFeeAmountType,
			Amount:     2_000,
		},
	}
	rawResponse, _ := json.Marshal(response)

	tests := []struct {
		name           string
		bodyRequest    string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:Invalid request body",
			bodyRequest:    "A",
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrRespForTest(40, c.ErrInvalidRequestPayload.Error()),
		},
		{
			name:           "ERROR:Invalid merchant ID format",
			bodyRequest:    invalidRequestBody,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":{"MerchantId":"Key: 'CreateFeeConfigOnBehalfRequest.MerchantId' Error:Field validation for 'MerchantId' failed on the 'uuid' tag"}}`,
		},
		{
			name:        "ERROR:Some error", // NOSONAR
			bodyRequest: requestBody,
			setupMock: func() {
				merchantSvc.On("CreateFeeConfigOnBehalf", c.ValueCtxMockType(), mock.Anything).Once().Return("", c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrRespForTest(99, "some error"),
		},
		{
			name:        "SUCCESS", // NOSONAR
			bodyRequest: requestBody,
			setupMock: func() {
				merchantSvc.On("CreateFeeConfigOnBehalf", c.ValueCtxMockType(), mock.Anything).Return(feeId, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   fmt.Sprintf(`{"code":"00","data":%s}`, string(rawResponse)),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setupMock != nil {
				test.setupMock()
			}
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/fee-on-behalf", strings.NewReader(test.bodyRequest))

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Output:", rec.Body.String())
			}
		})
	}
}

func TestGetFeeConfigOnBehalf(t *testing.T) {
	validator := validator.New()
	merchantSvc := mocks.NewIMerchantService(t)

	handler := New(merchantSvc, nil, validator, nil)

	router := chi.NewRouter()
	router.Get("/fee-on-behalf/details", handler.GetFeeConfigOnBehalf)

	tests := []struct {
		name           string
		query          string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:Invalid merchant ID",
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrRespForTest(40, "invalid merchant id value"),
		},
		{
			name:           "ERROR:Invalid reference",
			query:          "merchantId=3fc96de8-f65e-4b16-90a1-e2a00d1bae29",
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrRespForTest(40, "invalid reference"),
		},
		{
			name:  "ERROR:Some error", // NOSONAR
			query: "merchantId=3fc96de8-f65e-4b16-90a1-e2a00d1bae29&reference=PAYMENT",
			setupMock: func() {
				merchantSvc.On(
					"GetFeeConfigOnBehalf", c.ValueCtxMockType(), mock.Anything,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrRespForTest(99, "some error"),
		},
		{
			name:  "SUCCESS", // NOSONAR
			query: "merchantId=3fc96de8-f65e-4b16-90a1-e2a00d1bae29&reference=PAYMENT&paymentMethod=VIRTUAL_ACCOUNT",
			setupMock: func() {
				merchantSvc.On(
					"GetFeeConfigOnBehalf", c.ValueCtxMockType(), mock.Anything,
				).Return([]merchant.FeeConfigOnBehalfResponse{
					{
						Id:         "0192f6a9-ed2c-7ec3-8a5d-06c502cc08f6",
						Type:       "DEFAULT",
						AmountType: "AMOUNT",
						Amount:     2_000,
					},
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":{"reference":"PAYMENT","referenceType":"","paymentMethod":"VIRTUAL_ACCOUNT","merchantId":"3fc96de8-f65e-4b16-90a1-e2a00d1bae29","configs":[{"id":"0192f6a9-ed2c-7ec3-8a5d-06c502cc08f6","type":"DEFAULT","subMerchantId":null,"amountType":"AMOUNT","amount":2000,"percentage":0,"createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}]}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setupMock != nil {
				test.setupMock()
			}
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/fee-on-behalf/details", nil)

			req.URL.RawQuery = test.query

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Output:", rec.Body.String())
			}
		})
	}
}

func TestUpdateFeeConfigOnBehalf(t *testing.T) {
	validator := validator.New()
	merchantSvc := mocks.NewIMerchantService(t)

	handler := New(merchantSvc, nil, validator, nil)

	router := chi.NewRouter()
	router.Patch("/fee-on-behalf/{id}", handler.UpdateFeeConfigOnBehalf)

	feeId := "a5108d16-2b86-4b8d-9f11-fac0111465be"

	tests := []struct {
		name           string
		id             string
		requestBody    string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:Invalid on-behalf fee id",
			id:             "-",
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrRespForTest(40, "invalid id"),
		},
		{
			name:           "ERROR:Invalid request body",
			requestBody:    "A",
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrRespForTest(40, "invalid request payload"),
		},
		{
			name:           "ERROR:Invalid amountType",
			requestBody:    `{"amountType": "XXX","amount": 2500,"percentage": 0}`,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":{"AmountType":"Key: 'UpdateFeeConfigOnBehalfRequest.AmountType' Error:Field validation for 'AmountType' failed on the 'oneof' tag"}}`,
		},
		{
			name:        "ERROR:Some error", // NOSONAR
			requestBody: `{"amountType": "AMOUNT","amount": 2500,"percentage": 0}`,
			setupMock: func() {
				merchantSvc.On(
					"UpdateFeeConfigOnBehalf", c.ValueCtxMockType(), c.StringMockType(), mock.Anything,
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrRespForTest(99, "some error"),
		},
		{
			name:        "SUCCESS", // NOSONAR
			requestBody: `{"amountType": "AMOUNT","amount": 2500,"percentage": 0}`,
			setupMock: func() {
				merchantSvc.On(
					"UpdateFeeConfigOnBehalf", c.ValueCtxMockType(), c.StringMockType(), mock.Anything,
				).Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":{"message":"data successfully updated"}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setupMock != nil {
				test.setupMock()
			}
			if test.id == "" {
				test.id = feeId
			}
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPatch, "/fee-on-behalf/"+test.id, strings.NewReader(test.requestBody))

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Output:", rec.Body.String())
			}
		})
	}
}
