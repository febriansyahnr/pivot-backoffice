package merchant_test

import (
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
)

func TestTransactionConfig(t *testing.T) {
	validator := validator.New()
	merchantSvc := mocks.NewIMerchantService(t)
	userSvc := mocks.NewIUserService(t)

	router := chi.NewRouter()
	router.Patch("/merchants/{id}/transaction-configs", New(merchantSvc, userSvc, validator, nil).TransactionConfig)

	merchantId := uuid.NewString()
	bodyRequest := `{"disbursement":{"minAmount":10000,"maxAmount":250000000},"withdrawal":{"minAmount":10000,"maxAmount":250000000}}`

	tests := []struct {
		name           string
		merchantId     string
		bodyRequest    string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:Invalid merchant id",
			merchantId:     "-",
			bodyRequest:    ``,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrRespForTest(40, "merchant id is not valid"),
		},
		{
			name:           "ERROR:Malformed request body",
			bodyRequest:    `B`,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrRespForTest(40, "parser: invalid character 'B' looking for beginning of value"),
		},
		{
			name:        "ERROR:Transaction configs",
			bodyRequest: bodyRequest,
			setupMock: func() {
				merchantSvc.On(
					"TransactionConfig", c.ValueCtxMockType(), merchantId, c.PtrMerchantTransactionConfigMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrRespForTest(99, "some error"),
		},
		{
			name:        "SUCCESS",
			bodyRequest: bodyRequest,
			setupMock: func() {
				merchantSvc.On("TransactionConfig", c.ValueCtxMockType(), merchantId, c.PtrMerchantTransactionConfigMockType()).Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   fmt.Sprintf(`{"code":"00","data":{"merchantId":"%s","transactionConfigs":{"disbursement":{"minAmount":10000,"maxAmount":250000000},"withdrawal":{"minAmount":10000,"maxAmount":250000000}}}}`, merchantId),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			if test.merchantId == "" {
				test.merchantId = merchantId
			}
			req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/merchants/%s/transaction-configs", test.merchantId), strings.NewReader(test.bodyRequest))

			if test.setupMock != nil {
				test.setupMock()
			}

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}

func TestFDSConfig(t *testing.T) {
	validator := validator.New()
	merchantSvc := mocks.NewIMerchantService(t)

	handler := New(merchantSvc, nil, validator, nil)

	router := chi.NewRouter()
	router.Put("/merchants/{id}/fds-configs", handler.FDSConfig)

	merchantID := "17cd5cfc-e0e0-47f4-b951-7d701dc28da9"
	bodyRequest := `{"proofOfPayment": {"velocity": {"enabled": true,"window": {"interval": 1,"unit": "HOUR"},"threshold": {"count": 5},"action": "BLOCK"}}}`

	tests := []struct {
		name           string
		merchantId     string
		bodyRequest    string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:Invalid merchant id",
			merchantId:     "ABC",
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrRespForTest(40, "merchant id is not valid"),
		},
		{
			name:           "ERROR:Malformed request body",
			bodyRequest:    `B`,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrRespForTest(40, "invalid character 'B' looking for beginning of value"),
		},
		{
			name:           "ERROR:Invalid request body",
			bodyRequest:    `{"proofOfPayment": {"velocity": {"enabled": true,"window": {"interval": 1,"unit": "HOUR"},"threshold": {"count": 5}}}}`,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":{"Action":"Key: 'FDSConfigRequest.FDSConfig.ProofOfPayment.Velocity.Action' Error:Field validation for 'Action' failed on the 'required' tag"}}`,
		},
		{
			name:        "ERROR:FDS configs",
			bodyRequest: bodyRequest,
			setupMock: func() {
				merchantSvc.On(
					"FDSConfig", c.ValueCtxMockType(), merchantID, mock.Anything,
				).Once().Return(nil, assert.AnError)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrRespForTest(99, "assert.AnError general error for testing"),
		},
		{
			name:        "SUCCESS",
			bodyRequest: bodyRequest,
			setupMock: func() {
				merchantSvc.On("FDSConfig", c.ValueCtxMockType(), merchantID, mock.Anything).Return(&merchant.FDSConfigResponse{}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":{"proofOfPayment":null}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setupMock != nil {
				test.setupMock()
			}
			if test.merchantId == "" {
				test.merchantId = merchantID
			}
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPut, "/merchants/"+test.merchantId+"/fds-configs", strings.NewReader(test.bodyRequest))

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Actual:", rec.Body.String())
			}

			merchantSvc.AssertExpectations(t)
		})
	}
}

func TestGetFDSConfig(t *testing.T) {

	merchantSvc := mocks.NewIMerchantService(t)

	handler := New(merchantSvc, nil, nil, nil)

	router := chi.NewRouter()
	router.Get("/merchants/{id}/fds-configs", handler.GetFDSConfig)

	merchantID := "072a41b4-e41e-4e87-8ae5-4f2df1665c9c"
	responseBody := `{"code":"00","data":{"merchantId":"072a41b4-e41e-4e87-8ae5-4f2df1665c9c","merchantName":"TEST","merchantType":"MERCHANT","fdsConfig":{"proofOfPayment":{"velocity":{"enabled":true,"window":{"interval":1,"unit":"MINUTE"},"threshold":{"count":10},"action":"BLOCK"}}}}}`

	tests := []struct {
		name           string
		merchantId     string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:Invalid merchant id", // NOSONAR
			merchantId:     "ABC",
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrRespForTest(40, "merchant id is not valid"),
		},
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				merchantSvc.On("GetFDSConfig", c.ValueCtxMockType(), merchantID).Once().Return(nil, assert.AnError)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrRespForTest(99, "assert.AnError general error for testing"),
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				merchantSvc.On(
					"GetFDSConfig", c.ValueCtxMockType(), merchantID,
				).Once().Return(&merchant.GetFDSConfigResponse{
					MerchantID:   merchantID, // NOSONAR
					MerchantName: "TEST",     // NOSONAR
					MerchantType: "MERCHANT", // NOSONAR
					FDSConfig: merchant.FDSConfig{
						ProofOfPayment: &merchant.FDSFeatureProofOfPayment{
							Velocity: merchant.FDSRuleVelocityConfig{
								Enabled: true, // NOSONAR
								Window: merchant.FDSWindowConfig{
									Interval: 1,        // NOSONAR
									Unit:     "MINUTE", // NOSONAR
								},
								Threshold: merchant.FDSThresholdConfig{
									Count: 10, // NOSONAR
								},
								Action: "BLOCK", // NOSONAR
							},
						},
					},
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   responseBody,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setupMock != nil {
				test.setupMock()
			}
			if test.merchantId == "" {
				test.merchantId = merchantID
			}
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/merchants/"+test.merchantId+"/fds-configs", nil)

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Actual:", rec.Body.String())
			}

			merchantSvc.AssertExpectations(t)
		})
	}
}

func TestUpdateSettlementConfig(t *testing.T) {
	validator := validator.New()
	merchantSvc := mocks.NewIMerchantService(t)

	router := chi.NewRouter()
	router.Patch("/merchants/fee/{id}/settlement-configs", New(merchantSvc, nil, validator, nil).UpdateSettlementConfig)

	merchantFeeId := uuid.NewString()
	bodyRequestWithoutCutOff := `{"type":"INSTANT", "cutOff": null}`
	bodyRequestWithCutOff := `{"type":"T+1","cutOff":{"window":{"startTime":"22:00:00","endTime":"06:00:00"},"deferral":{"offsetDays":1,"executionTime":"07:00:00"}}}`

	tests := []struct {
		name           string
		merchantFeeId  string
		bodyRequest    string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:Invalid merchant fee id",
			merchantFeeId:  "-",
			bodyRequest:    ``,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrRespForTest(40, "merchant fee id is not valid"),
		},
		{
			name:           "ERROR:Malformed request body",
			bodyRequest:    `B`,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrRespForTest(40, "parser: invalid character 'B' looking for beginning of value"),
		},
		{
			name:           "ERROR:Validate payload",
			bodyRequest:    `{"type":"N"}`,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"invalid settlement type format"}`,
		},
		{
			name:           "ERROR:Empty object cut-off time",
			bodyRequest:    `{"type":"T+1","cutOff": {}}`,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":{"EndTime":"Key: 'SettlementConfig.CutOff.Window.EndTime' Error:Field validation for 'EndTime' failed on the 'required' tag","ExecutionTime":"Key: 'SettlementConfig.CutOff.Deferral.ExecutionTime' Error:Field validation for 'ExecutionTime' failed on the 'required' tag","OffsetDays":"Key: 'SettlementConfig.CutOff.Deferral.OffsetDays' Error:Field validation for 'OffsetDays' failed on the 'required' tag","StartTime":"Key: 'SettlementConfig.CutOff.Window.StartTime' Error:Field validation for 'StartTime' failed on the 'required' tag"}}`,
		},
		{
			name:        "ERROR:Transaction configs",
			bodyRequest: bodyRequestWithoutCutOff,
			setupMock: func() {
				merchantSvc.On(
					"UpdateSettlementConfig", c.ValueCtxMockType(), merchantFeeId, c.PtrMerchantSettlementConfigMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrRespForTest(99, "some error"),
		},
		{
			name:        "SUCCESS:Without cut-off time",
			bodyRequest: bodyRequestWithoutCutOff,
			setupMock: func() {
				merchantSvc.On("UpdateSettlementConfig", c.ValueCtxMockType(), merchantFeeId, c.PtrMerchantSettlementConfigMockType()).Once().Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":{"type":"INSTANT","cutOff":null,"isOnHold":false}}`,
		},
		{
			name:        "SUCCESS:With cut-off time",
			bodyRequest: bodyRequestWithCutOff,
			setupMock: func() {
				merchantSvc.On("UpdateSettlementConfig", c.ValueCtxMockType(), merchantFeeId, c.PtrMerchantSettlementConfigMockType()).Once().Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":{"type":"T+1","cutOff":{"window":{"startTime":"22:00:00","endTime":"06:00:00"},"deferral":{"offsetDays":1,"executionTime":"07:00:00"}},"isOnHold":false}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			if test.merchantFeeId == "" {
				test.merchantFeeId = merchantFeeId
			}
			req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/merchants/fee/%s/settlement-configs", test.merchantFeeId), strings.NewReader(test.bodyRequest))

			if test.setupMock != nil {
				test.setupMock()
			}

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Actual Response:", rec.Body.String())
			}
		})
	}
}

func TestGetTransactionConfig(t *testing.T) {
	service := mocks.NewIMerchantService(t)

	handler := New(service, nil, nil, nil)

	router := chi.NewRouter()
	router.Get("/{id}/transaction-configs", handler.GetTransactionConfig)

	merchantId := "18c23032-0a31-4cb2-ad0f-5ad06caebe89"

	tests := []struct {
		name           string
		merchantId     string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:Invalid merchant id",
			setupMock:      func() { /*No Body*/ }, // NOSONAR
			merchantId:     "ABC",
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40", "errors":"merchant id is not valid"}`,
		},
		{
			name:       "ERROR:Some error",
			merchantId: merchantId,
			setupMock: func() {
				service.On(
					"GetTransactionConfig", c.ValueCtxMockType(), merchantId,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99", "errors":"some error"}`,
		},
		{
			name:       "SUCCESS",
			merchantId: merchantId,
			setupMock: func() {
				service.On(
					"GetTransactionConfig", c.ValueCtxMockType(), merchantId,
				).Return(&merchant.TransactionConfigResp{
					MerchantId: merchantId,
					TransactionConfigs: merchant.TransactionConfigs{
						Disbursement: merchant.DisbursementConfig{
							MinAmount: 10_001,
							MaxAmount: 1_000_001,
						},
						Withdrawal: merchant.WithdrawalConfig{
							MinAmount: 10_000,
							MaxAmount: 20_000,
						},
					},
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   fmt.Sprintf(`{"code":"00","data":{"merchantId":"%s","transactionConfigs":{"disbursement":{"minAmount":10001,"maxAmount":1000001},"withdrawal":{"minAmount":10000,"maxAmount":20000}}}}`, merchantId),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/"+test.merchantId+"/transaction-configs", nil)

			test.setupMock()
			router.ServeHTTP(rec, req)

			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Output:", rec.Body.String())
			}
		})
	}
}

func TestPaymentInvestigationConfig(t *testing.T) {
	validator := validator.New()
	merchantSvc := mocks.NewIMerchantService(t)

	handler := New(merchantSvc, nil, validator, nil)

	router := chi.NewRouter()
	router.Patch("/merchants/{id}/payment-investigation-configs", handler.PaymentInvestigationConfig)

	merchantID := "e4086ea1-40f0-4843-a013-8ff65d64d011"
	bodyRequest := `{"enabled": true,"pivotPercentageLoss": 50,"pivotMaxLoss": 500000}`

	tests := []struct {
		name           string
		merchantId     string
		bodyRequest    string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:Invalid merchant id",
			merchantId:     "ABC",
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrRespForTest(40, "merchant id is not valid"),
		},
		{
			name:           "ERROR:Malformed request body",
			bodyRequest:    `B`,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrRespForTest(40, "invalid character 'B' looking for beginning of value"),
		},
		{
			name:           "ERROR:Invalid request body",
			bodyRequest:    `{"enabled": true,"pivotPercentageLoss": 120,"pivotMaxLoss": -1}`,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":{"PivotMaxLoss":"Key: 'PaymentInvestigationConfigRequest.PivotMaxLoss' Error:Field validation for 'PivotMaxLoss' failed on the 'min' tag","PivotPercentageLoss":"Key: 'PaymentInvestigationConfigRequest.PivotPercentageLoss' Error:Field validation for 'PivotPercentageLoss' failed on the 'max' tag"}}`,
		},
		{
			name:        "ERROR:Update payment investigation configs",
			bodyRequest: bodyRequest,
			setupMock: func() {
				merchantSvc.On(
					"PaymentInvestigationConfig", mock.Anything, merchantID, mock.Anything,
				).Once().Return(nil, assert.AnError)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrRespForTest(99, "assert.AnError general error for testing"),
		},
		{
			name:        "SUCCESS",
			bodyRequest: bodyRequest,
			setupMock: func() {
				merchantSvc.On(
					"PaymentInvestigationConfig", mock.Anything, merchantID, mock.Anything,
				).Once().Return(&merchant.PaymentInvestigationConfigResponse{
					MerchantID:          merchantID,
					MerchantName:        "TEST",
					Enabled:             true,
					PivotPercentageLoss: 50,
					PivotMaxLoss:        500_000,
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":{"enabled":true,"merchantId":"e4086ea1-40f0-4843-a013-8ff65d64d011","merchantName":"TEST","pivotPercentageLoss":50,"pivotMaxLoss":500000}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setupMock != nil {
				test.setupMock()
			}
			if test.merchantId == "" {
				test.merchantId = merchantID
			}
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPatch, "/merchants/"+test.merchantId+"/payment-investigation-configs", strings.NewReader(test.bodyRequest))

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Actual:", rec.Body.String())
			}

			merchantSvc.AssertExpectations(t)
		})
	}
}
