package cardFundedPayoutController_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/cardFundedPayout"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetTransactionConfig(t *testing.T) {
	cardFundedPayoutSvc := serviceMocks.NewICardFundedPayoutService(t)
	feeSvc := serviceMocks.NewIFeeService(t)
	merchantSvc := serviceMocks.NewIMerchantService(t)
	cfg := &config.Config{}
	controller := New(cfg, validatorExt.New(), cardFundedPayoutSvc,
		WithFeeService(feeSvc),
		WithMerchantService(merchantSvc),
	)

	merchantId := "merchant-123"
	parentMerchantId := "parent-merchant-123"
	userClaim := &userModel.UserTokenClaims{
		UUID:       "user-uuid-123",
		MerchantId: merchantId,
	}
	merchantConfigs := &merchantModel.MerchantIdForConfigs{
		MerchantType:              constant.MerchantTypeMerchant,
		MerchantTransactionConfig: merchantId,
	}
	ctxWithParent := context.WithValue(context.Background(), constant.CtxParentMerchantId, parentMerchantId)

	tests := []struct {
		name         string
		userClaim    *userModel.UserTokenClaims
		setupMock    func()
		queryParams  string
		wantStatus   int
		wantResponse func() string
	}{
		{
			name:        "ERROR: User not found",
			queryParams: "",
			wantStatus:  http.StatusUnauthorized,
			wantResponse: func() string {
				return `{"code":"41","message":"user not found","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`
			},
		},
		{
			name:        "ERROR: Get merchant id for configs returns error",
			userClaim:   userClaim,
			queryParams: "",
			setupMock: func() {
				merchantSvc.On("GetMerchantIdForConfigs", mock.Anything, merchantId, false).
					Return(context.Background(), nil, pkgErrors.New(response.HttpErrInternal, errors.New("service error"))).Once()
			},
			wantStatus: http.StatusInternalServerError,
			wantResponse: func() string {
				return `{"code":"99","message":"service error","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`
			},
		},
		{
			name:        "ERROR: Get fee calculation and detail returns error for merchant",
			userClaim:   userClaim,
			queryParams: "",
			setupMock: func() {
				merchantSvc.On("GetMerchantIdForConfigs", mock.Anything, merchantId, false).
					Return(context.Background(), merchantConfigs, nil).Once()
				feeSvc.On("GetFeeCalculationAndDetail", mock.Anything, mock.Anything).
					Return(0.0, nil, pkgErrors.New(response.HttpErrInternal, errors.New("fee error"))).Once()
			},
			wantStatus: http.StatusInternalServerError,
			wantResponse: func() string {
				return `{"code":"99","message":"fee error","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`
			},
		},
		{
			name:        "ERROR: Get transaction fee on behalf returns error for sub-merchant",
			userClaim:   userClaim,
			queryParams: "",
			setupMock: func() {
				subMerchantConfigs := &merchantModel.MerchantIdForConfigs{
					MerchantType:              constant.MerchantTypeSubMerchant,
					MerchantTransactionConfig: merchantId,
				}
				merchantSvc.On("GetMerchantIdForConfigs", mock.Anything, merchantId, false).
					Return(ctxWithParent, subMerchantConfigs, nil).Once()
				feeSvc.On("GetTransactionFeeOnBehalf", mock.Anything, mock.Anything).
					Return(nil, pkgErrors.New(response.HttpErrInternal, errors.New("fee on behalf error"))).Once()
			},
			wantStatus: http.StatusInternalServerError,
			wantResponse: func() string {
				return `{"code":"99","message":"fee on behalf error","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`
			},
		},
		{
			name:        "SUCCESS: Get transaction config for merchant",
			userClaim:   userClaim,
			queryParams: "",
			setupMock: func() {
				merchantSvc.On("GetMerchantIdForConfigs", mock.Anything, merchantId, false).
					Return(context.Background(), merchantConfigs, nil).Once()
				feeSvc.On("GetFeeCalculationAndDetail", mock.Anything, mock.Anything).
					Return(4000.0, &feeModel.FeeMetadataObject{
						Type:          "CARD_FUNDED_PAYOUT",
						DeductionType: "DIRECT",
						AmountType:    "AMOUNT",
						Amount:        4000,
						TaxType:       "NON_PKP",
						FinalAmount:   4000,
					}, nil).Once()
			},
			wantStatus: http.StatusOK,
			wantResponse: func() string {
				return `{"code":"00","message":"OK","data":{"feeDetail":{"type":"CARD_FUNDED_PAYOUT","referenceType":"","method":"","deductionType":"DIRECT","amountType":"AMOUNT","amount":4000,"percentage":0,"taxType":"NON_PKP","taxPercentage":0,"taxAmount":0,"finalAmount":4000}}}`
			},
		},
		{
			name:        "SUCCESS: Get transaction config for merchant with amount and settlement method",
			userClaim:   userClaim,
			queryParams: "?amount=100000&settlementMethod=INSTANT",
			setupMock: func() {
				merchantSvc.On("GetMerchantIdForConfigs", mock.Anything, merchantId, false).
					Return(context.Background(), merchantConfigs, nil).Once()
				feeSvc.On("GetFeeCalculationAndDetail", mock.Anything, mock.Anything).
					Return(5000.0, &feeModel.FeeMetadataObject{
						Type:          "CARD_FUNDED_PAYOUT",
						DeductionType: "DIRECT",
						AmountType:    "PERCENTAGE",
						Amount:        5000,
						TaxType:       "PKP",
						FinalAmount:   5500,
					}, nil).Once()
			},
			wantStatus: http.StatusOK,
			wantResponse: func() string {
				return `{"code":"00","message":"OK","data":{"feeDetail":{"type":"CARD_FUNDED_PAYOUT","referenceType":"","method":"","deductionType":"DIRECT","amountType":"PERCENTAGE","amount":5000,"percentage":0,"taxType":"PKP","taxPercentage":0,"taxAmount":0,"finalAmount":5500}}}`
			},
		},
		{
			name:        "SUCCESS: Get transaction config for sub-merchant",
			userClaim:   userClaim,
			queryParams: "",
			setupMock: func() {
				subMerchantConfigs := &merchantModel.MerchantIdForConfigs{
					MerchantType:              constant.MerchantTypeSubMerchant,
					MerchantTransactionConfig: merchantId,
				}
				merchantSvc.On("GetMerchantIdForConfigs", mock.Anything, merchantId, false).
					Return(ctxWithParent, subMerchantConfigs, nil).Once()
				feeSvc.On("GetTransactionFeeOnBehalf", mock.Anything, mock.Anything).
					Return(&feeModel.TrxFeeOnBehalfMetadata{
						Reference:   "CARD_FUNDED_PAYOUT",
						AmountType:  "AMOUNT",
						Amount:      2500,
						FinalAmount: 2500,
					}, nil).Once()
			},
			wantStatus: http.StatusOK,
			wantResponse: func() string {
				return `{"code":"00","message":"OK","data":{"feeDetail":{"type":"","amountType":"AMOUNT","amount":2500,"percentage":0,"finalAmount":2500}}}`
			},
		},
		{
			name:        "SUCCESS: Get transaction config with invalid settlement method (uses default)",
			userClaim:   userClaim,
			queryParams: "?settlementMethod=INVALID",
			setupMock: func() {
				merchantSvc.On("GetMerchantIdForConfigs", mock.Anything, merchantId, false).
					Return(context.Background(), merchantConfigs, nil).Once()
				feeSvc.On("GetFeeCalculationAndDetail", mock.Anything, mock.Anything).
					Return(4000.0, &feeModel.FeeMetadataObject{
						Type:          "CARD_FUNDED_PAYOUT",
						DeductionType: "DIRECT",
						AmountType:    "AMOUNT",
						Amount:        4000,
						TaxType:       "NON_PKP",
						FinalAmount:   4000,
					}, nil).Once()
			},
			wantStatus: http.StatusOK,
			wantResponse: func() string {
				return `{"code":"00","message":"OK","data":{"feeDetail":{"type":"CARD_FUNDED_PAYOUT","referenceType":"","method":"","deductionType":"DIRECT","amountType":"AMOUNT","amount":4000,"percentage":0,"taxType":"NON_PKP","taxPercentage":0,"taxAmount":0,"finalAmount":4000}}}`
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setupMock != nil {
				test.setupMock()
			}

			req := httptest.NewRequest(http.MethodGet, "/card-funded-payout/configs"+test.queryParams, nil)
			rec := httptest.NewRecorder()

			ctx := req.Context()
			if test.userClaim != nil {
				ctx = context.WithValue(ctx, constant.CtxUserInfoKey, test.userClaim)
			}
			req = req.WithContext(ctx)

			controller.GetTransactionConfig(rec, req)

			assert.Equal(t, test.wantStatus, rec.Result().StatusCode)
			if test.wantResponse != nil {
				if !assert.JSONEq(t, test.wantResponse(), rec.Body.String()) {
					t.Log("Result:", rec.Body.String())
				}
			}
		})
	}
}