package disbursementController_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	s "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/disbursement"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetTransactionConfig(t *testing.T) {
	feeSvc := serviceMocks.NewIFeeService(t)
	merchantSvc := serviceMocks.NewIMerchantService(t)
	disbursementSvc := serviceMocks.NewIDisbursementService(t)

	handler := New(&config.Config{}, validator.New(), nil, Services{
		DisbursementSvc: disbursementSvc, FeeSvc: feeSvc, MerchantSvc: merchantSvc,
	}, nil, nil)

	router := chi.NewRouter()
	router.Get("/disbursements/configs", handler.GetTransactionConfig)

	merchantId := "9a63b7fe-8434-4260-abab-f8ff14acc204"
	parentMerchantId := "7ce63495-8803-483c-a64c-5eaf1861488c"
	userClaim := &user.UserTokenClaims{
		UUID: uuid.NewString(), MerchantId: merchantId,
	}
	trxConfig := &disbursementModel.TransactionConfig{
		MinAmount: 10_000, MaxAmount: 250_000,
	}
	merchantConfigs := &merchant.MerchantIdForConfigs{
		MerchantType:              c.MerchantTypeMerchant,
		MerchantTransactionConfig: merchantId,
	}
	ctx := context.WithValue(context.Background(), c.CtxParentMerchantId, parentMerchantId)

	tests := []struct {
		name           string
		userClaim      *user.UserTokenClaims
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:User not found",
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   c.WrapErrApiRespForTest(41, s.ErrTypeAPI, "user not found"),
		},
		{
			name:      "ERROR:Get merchant id for configs",
			userClaim: userClaim,
			setupMock: func() {
				merchantSvc.On("GetMerchantIdForConfigs", c.ValueCtxMockType(), merchantId, false).Once().Return(ctx, nil, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrApiRespForTest(99, s.ErrTypeUnknown, "some error"),
		},
		{
			name:      "ERROR:Get transaction config",
			userClaim: userClaim,
			setupMock: func() {
				merchantSvc.On("GetMerchantIdForConfigs", c.ValueCtxMockType(), merchantId, false).Return(ctx, merchantConfigs, nil)
				disbursementSvc.On("GetTransactionConfig", c.ValueCtxMockType(), merchantId).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrApiRespForTest(99, s.ErrTypeUnknown, "some error"),
		},
		{
			name:      "ERROR:Get fee calculation and detail",
			userClaim: userClaim,
			setupMock: func() {
				disbursementSvc.On("GetTransactionConfig", c.ValueCtxMockType(), merchantId).Return(trxConfig, nil)
				feeSvc.On(
					"GetFeeCalculationAndDetail", c.ValueCtxMockType(), mock.Anything,
				).Once().Return(0.0, nil, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrApiRespForTest(99, s.ErrTypeUnknown, "some error"),
		},
		{
			name:      "SUCCESS:Merchant",
			userClaim: userClaim,
			setupMock: func() {
				feeSvc.On(
					"GetFeeCalculationAndDetail", c.ValueCtxMockType(), mock.Anything,
				).Return(4_000.0, &feeModel.FeeMetadataObject{
					Type:          "DISBURSEMENT",
					DeductionType: "DIRECT",
					AmountType:    "AMOUNT",
					Amount:        4_000,
					TaxType:       "NON_PKP",
					FinalAmount:   4_000,
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"minAmount":10000,"maxAmount":250000,"feeDetail":{"type":"DISBURSEMENT","method":"","deductionType":"DIRECT","amountType":"AMOUNT","amount":4000,"percentage":0,"taxType":"NON_PKP","taxPercentage":0,"taxAmount":0,"totalAmount":4000}}}`,
		},
		{
			name:      "SUCCESS:Sub-Merchant",
			userClaim: userClaim,
			setupMock: func() {
				merchantConfigs.MerchantType = c.MerchantTypeSubMerchant

				feeSvc.On(
					"GetTransactionFeeOnBehalf", c.ValueCtxMockType(), mock.Anything,
				).Return(&feeModel.TrxFeeOnBehalfMetadata{
					Reference: "DISBURSEMENT", AmountType: "AMOUNT", Amount: 2_500, FinalAmount: 2_500,
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"minAmount":10000,"maxAmount":250000,"feeDetail":{"type":"DISBURSEMENT","method":"","deductionType":"DIRECT","amountType":"AMOUNT","amount":2500,"percentage":0,"taxType":"NON_PKP","taxPercentage":0,"taxAmount":0,"totalAmount":2500}}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/disbursements/configs", nil)

			if test.setupMock != nil {
				test.setupMock()
			}
			if test.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), c.CtxUserInfoKey, test.userClaim))
			}

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Result:", rec.Body.String())
			}
		})
	}
}

func TestGetDailyTransactionLimit(t *testing.T) {
	service := serviceMocks.NewIDisbursementService(t)

	handler := New(nil, validator.New(), nil, Services{DisbursementSvc: service}, nil, nil)

	router := chi.NewRouter()
	router.Get("/daily-limits/{type}", handler.GetDailyTransactionLimit)

	userClaims := &user.UserTokenClaims{
		MerchantId: "dae00c25-a223-4e7c-81bc-89f631b77490",
	}

	tests := []struct {
		name           string
		merchantType   string
		userClaims     *user.UserTokenClaims
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:User not found",
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   c.WrapErrApiRespForTest(41, s.ErrTypeAPI, "user not found"),
		},
		{
			name:           "ERROR:Invalid merchant type",
			merchantType:   "test",
			userClaims:     userClaims,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrApiRespForTest(40, s.ErrTypeAPI, "merchant type not registered"),
		},
		{
			name:       "ERROR:Some error",
			userClaims: userClaims,
			setupMock: func() {
				service.On(
					"GetDailyTransactionLimit", c.ValueCtxMockType(), userClaims.MerchantId, c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrApiRespForTest(99, s.ErrTypeUnknown, "some error"),
		},
		{
			name:       "SUCCESS:Sub-Merchant request",
			userClaims: userClaims,
			setupMock: func() {
				service.On(
					"GetDailyTransactionLimit", c.ValueCtxMockType(), userClaims.MerchantId, c.StringMockType(),
				).Once().Return(nil, c.ErrForbiddenAccess)
			},
			wantStatusCode: http.StatusNoContent,
		},
		{
			name:       "SUCCESS:Merchant",
			userClaims: userClaims,
			setupMock: func() {
				service.On(
					"GetDailyTransactionLimit", c.ValueCtxMockType(), userClaims.MerchantId, c.StringMockType(),
				).Return(&disbursementModel.DailyTransactionLimitResponse{
					Limit:     util.ValueToPtr(10_000.00),
					Processed: 1_000,
					Remaining: 9_000,
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"limit":10000,"processed":1000,"remaining":9000}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.merchantType == "" {
				test.merchantType = "merchant"
			}
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/daily-limits/"+test.merchantType, nil)

			if test.setupMock != nil {
				test.setupMock()
			}
			if test.userClaims != nil {
				req = req.WithContext(context.WithValue(req.Context(), c.CtxUserInfoKey, test.userClaims))
			}

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if rec.Result().StatusCode == http.StatusNoContent {
				assert.Empty(t, rec.Body.String())

			} else {
				assert.JSONEq(t, test.wantRespBody, rec.Body.String())
			}
		})
	}
}

func TestGetTransactionLimit(t *testing.T) {
	merchantSvc := serviceMocks.NewIMerchantService(t)

	handler := New(nil, nil, nil, Services{MerchantSvc: merchantSvc}, nil, nil)

	router := chi.NewRouter()
	router.Get("/limits", handler.GetTransactionLimit)

	userClaims := &user.UserTokenClaims{
		MerchantId: "ceba442d-54a6-41c3-8825-501d665bc8f6",
	}

	tests := []struct {
		name           string
		userClaims     *user.UserTokenClaims
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:User not found",
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   c.WrapErrApiRespForTest(41, s.ErrTypeAPI, "user not found"),
		},
		{
			name:       "ERROR:Some error",
			userClaims: userClaims,
			setupMock: func() {
				merchantSvc.On(
					"GetTransactionConfig", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrApiRespForTest(99, s.ErrTypeUnknown, "some error"),
		},
		{
			name:       "SUCCESS",
			userClaims: userClaims,
			setupMock: func() {
				merchantSvc.On(
					"GetTransactionConfig", c.ValueCtxMockType(), c.StringMockType(),
				).Return(&merchant.TransactionConfigResp{
					MerchantId:   userClaims.MerchantId,
					MerchantName: "Dummy",
					MerchantType: "MERCHANT",
					TransactionConfigs: merchant.TransactionConfigs{
						Disbursement: merchant.DisbursementConfig{
							MinAmount: 10_000,
							MaxAmount: 1_000_000,
						},
						Withdrawal: merchant.WithdrawalConfig{
							MinAmount: 10_500,
							MaxAmount: 20_500,
						},
					},
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   fmt.Sprintf(`{"code":"00","message":"OK","data":{"merchantId":"%s","merchantName":"Dummy","merchantType":"MERCHANT","transactionConfigs":{"disbursement":{"minAmount":10000,"maxAmount":1000000},"withdrawal":{"minAmount":10500,"maxAmount":20500}}}}`, userClaims.MerchantId),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/limits", nil)

			if test.setupMock != nil {
				test.setupMock()
			}
			if test.userClaims != nil {
				req = req.WithContext(context.WithValue(req.Context(), c.CtxUserInfoKey, test.userClaims))
			}

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Output:", rec.Body.String())
			}
		})
	}
}

func TestGetTransactionLimitSubMerchant(t *testing.T) {
	merchantSvc := serviceMocks.NewIMerchantService(t)

	handler := New(nil, nil, nil, Services{MerchantSvc: merchantSvc}, nil, nil)

	router := chi.NewRouter()
	router.Get("/limits/sub-merchants/{id}", handler.GetTransactionLimitSubMerchant)

	userClaims := &user.UserTokenClaims{
		MerchantId: "ceba442d-54a6-41c3-8825-501d665bc8f6",
	}
	subMerchantId := "f2faac05-530a-413c-bbaa-6edde714530e"

	tests := []struct {
		name           string
		id             string
		userClaims     *user.UserTokenClaims
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:User not found",
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   c.WrapErrApiRespForTest(41, s.ErrTypeAPI, "user not found"),
		},
		{
			name:           "ERROR:Invalid sub-merchant id",
			id:             "XXXX",
			userClaims:     userClaims,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrApiRespForTest(40, s.ErrTypeAPI, "invalid merchant id value"),
		},
		{
			name:       "ERROR:Find merchant by id",
			userClaims: userClaims,
			setupMock: func() {
				merchantSvc.On(
					"FindMerchantByID", c.ValueCtxMockType(), subMerchantId,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrApiRespForTest(99, s.ErrTypeUnknown, "some error"),
		},
		{
			name:       "ERROR:Sub-merchant not found",
			userClaims: userClaims,
			setupMock: func() {
				merchantSvc.On(
					"FindMerchantByID", c.ValueCtxMockType(), subMerchantId,
				).Once().Return(nil, nil)
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrApiRespForTest(40, s.ErrTypeAPI, "merchant not found"),
		},
		{
			name:       "ERROR:Forbidden access",
			userClaims: userClaims,
			setupMock: func() {
				merchantSvc.On(
					"FindMerchantByID", c.ValueCtxMockType(), subMerchantId,
				).Once().Return(&merchant.Merchant{}, nil)
			},
			wantStatusCode: http.StatusForbidden,
			wantRespBody:   c.WrapErrApiRespForTest(43, s.ErrTypeAPI, "forbidden access"),
		},
		{
			name:       "ERROR:Some error",
			userClaims: userClaims,
			setupMock: func() {
				merchantSvc.On(
					"FindMerchantByID", c.ValueCtxMockType(), subMerchantId,
				).Return(&merchant.Merchant{
					UUID: subMerchantId, ParentID: sql.NullString{String: userClaims.MerchantId},
				}, nil)

				merchantSvc.On(
					"GetTransactionConfig", c.ValueCtxMockType(), subMerchantId,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrApiRespForTest(99, s.ErrTypeUnknown, "some error"),
		},
		{
			name:       "SUCCESS",
			userClaims: userClaims,
			setupMock: func() {
				merchantSvc.On(
					"GetTransactionConfig", c.ValueCtxMockType(), subMerchantId,
				).Return(&merchant.TransactionConfigResp{
					MerchantId:   userClaims.MerchantId,
					MerchantName: "Dummy",
					MerchantType: "MERCHANT",
					TransactionConfigs: merchant.TransactionConfigs{
						Disbursement: merchant.DisbursementConfig{
							MinAmount: 10_000,    // NOSONAR
							MaxAmount: 1_000_000, // NOSONAR
						},
						Withdrawal: merchant.WithdrawalConfig{
							MinAmount: 10_000,    // NOSONAR
							MaxAmount: 1_000_000, // NOSONAR
						},
					},
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   fmt.Sprintf(`{"code":"00","message":"OK","data":{"merchantId":"%s","merchantName":"Dummy","merchantType":"MERCHANT","transactionConfigs":{"disbursement":{"minAmount":10000,"maxAmount":1000000},"withdrawal":{"minAmount":10000,"maxAmount":1000000}}}}`, userClaims.MerchantId),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			if test.id == "" {
				test.id = subMerchantId
			}

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/limits/sub-merchants/"+test.id, nil)

			if test.setupMock != nil {
				test.setupMock()
			}
			if test.userClaims != nil {
				req = req.WithContext(context.WithValue(req.Context(), c.CtxUserInfoKey, test.userClaims))
			}

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Output:", rec.Body.String())
			}
		})
	}
}
