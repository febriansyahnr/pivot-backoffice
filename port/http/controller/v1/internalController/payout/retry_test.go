package internalPayoutController_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/func"
	mockMQ "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	internalPayoutController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/internalController/payout"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestInternalPayoutControllerRetryBulk(t *testing.T) {
	mockDisbursement := serviceMocks.NewIDisbursementService(t)
	mockRabbitMQ := mockMQ.NewRabbitMQExt(t)
	mockAccountInquirySvc := serviceMocks.NewIAccountInquiryService(t)
	conf := config.Config{
		Environment: "development",
	}

	router := chi.NewRouter()
	router.Post(
		"/payouts/{id}/retry", internalPayoutController.New(
			&conf, validator.New(), mockDisbursement, serviceMocks.NewIMerchantService(t), mockAccountInquirySvc, mockRabbitMQ,
		).RetryBulk,
	)

	validRequestID := func(req *http.Request) {
		*req = *req.WithContext(context.WithValue(req.Context(), constant.CtxMerchantInfo, &merchantModel.MerchantAuthTokenClaims{
			MerchantId: "193b2a46-3ca0-4d1a-9dae-7536b1b4a5f5",
		}))
		req.Header.Set("Content-Type", "application/json")
	}

	getBulkDisbursementResp := &disbursementModel.GetBulkDisbursementForOpenApiByIDResponse{
		UUID:       "186ea62c-d02c-4b72-a2bf-c3214593f880",
		MerchantID: "193b2a46-3ca0-4d1a-9dae-7536b1b4a5f5",
		PayoutResults: disbursementModel.PayoutResultObject{
			TotalPendingCount:  0,
			TotalPendingAmount: 0,
			TotalSuccessCount:  2,
			TotalSuccessAmount: 25000,
			TotalFailedCount:   0,
			TotalFailedAmount:  0,
		},
		Payouts: []disbursementModel.PayoutObject{
			{
				ReferenceID: "testing-prod2",
				InquiryID:   "",
				ChannelCode: "BRI",
				ChannelInformation: disbursementModel.PayoutChannelInformation{
					AccountNumber: "111501016013508",
					AccountName:   "Rizaldy Septa Amanda",
				},
				Amount: commonModel.Amount{
					Currency: "IDR",
					Value:    "10000",
				},
				Description: "This is remark",
				Status:      "SUCCESS",
				Reason:      "Insufficient Balance",
				CreatedAt:   time.Date(2024, 7, 26, 11, 57, 55, 0, time.UTC),
				UpdatedAt:   time.Date(2024, 7, 28, 8, 2, 54, 0, time.UTC),
			},
			{
				ReferenceID: "testing-prod1",
				InquiryID:   "",
				ChannelCode: "BRI",
				ChannelInformation: disbursementModel.PayoutChannelInformation{
					AccountNumber: "111501016013508",
					AccountName:   "Rizaldy Septa Amanda",
				},
				Amount: commonModel.Amount{
					Currency: "IDR",
					Value:    "10000",
				},
				Description: "This is remark",
				Status:      "SUCCESS",
				Reason:      "Insufficient Balance",
				CreatedAt:   time.Date(2024, 7, 26, 11, 57, 55, 0, time.UTC),
				UpdatedAt:   time.Date(2024, 7, 28, 8, 2, 54, 0, time.UTC),
			},
		},
		Status:    "PENDING",
		CreatedAt: time.Date(2024, 7, 26, 11, 57, 55, 0, time.UTC),
		UpdatedAt: time.Date(2024, 7, 28, 8, 2, 54, 0, time.UTC),
	}
	commonResp := &commonModel.PaginationResponse{
		Data: getBulkDisbursementResp,
	}

	tests := []struct {
		name           string
		path           string
		reqBody        string
		reqSetting     func(r *http.Request)
		mockSetup      func(d *serviceMocks.IDisbursementService, r *mockMQ.RabbitMQExt)
		setHeaders     func(req *http.Request)
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR: invalid bulk disbursement id",
			path:           "/invalid",
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   mocks.ResponseFormattingTest(40, "disbursement id is required"),
		},
		{
			name:           "ERROR: invalid merchant info",
			path:           "/3a2675d6-5a0b-4bf6-9650-d8e58394ebe3",
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   mocks.ResponseFormattingTest(41, "merchant not found"),
		},
		{
			name:       "ERROR: failed to get bulk disbursement for open-api",
			path:       "/3a2675d6-5a0b-4bf6-9650-d8e58394ebe3",
			reqBody:    `{"id": "uuid-uuid-uuid-uuid"}`,
			reqSetting: validRequestID,
			mockSetup: func(d *serviceMocks.IDisbursementService, r *mockMQ.RabbitMQExt) {
				d.On(
					"GetBulkDisbursementForOpenApiByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeDisbursementFilterRequestReference),
					constant.Int64MockType(),
					constant.Int64MockType(),
				).Once().
					Return(nil, pkgErrors.New(response.HttpErrRequest, errors.New("failed on get bulk disbursement for open-api")))
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   mocks.ResponseFormattingTest(40, "failed on get bulk disbursement for open-api"),
		},
		{
			name:       "ERROR: bulk disbursement not found",
			path:       "/3a2675d6-5a0b-4bf6-9650-d8e58394ebe3",
			reqBody:    `{"id": "uuid-uuid-uuid-uuid"}`,
			reqSetting: validRequestID,
			mockSetup: func(d *serviceMocks.IDisbursementService, r *mockMQ.RabbitMQExt) {
				d.On(
					"GetBulkDisbursementForOpenApiByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeDisbursementFilterRequestReference),
					constant.Int64MockType(),
					constant.Int64MockType(),
				).Once().
					Return(nil, nil)
			},
			wantStatusCode: http.StatusNotFound,
			wantRespBody:   mocks.ResponseFormattingTest(44, "bulk disbursement not found"),
		},
		{
			name:       "ERROR: failed to retry bulk",
			path:       "/3a2675d6-5a0b-4bf6-9650-d8e58394ebe3",
			reqBody:    `{"id": "uuid-uuid-uuid-uuid"}`,
			reqSetting: validRequestID,
			mockSetup: func(d *serviceMocks.IDisbursementService, r *mockMQ.RabbitMQExt) {
				d.On(
					"GetBulkDisbursementForOpenApiByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeDisbursementFilterRequestReference),
					constant.Int64MockType(),
					constant.Int64MockType(),
				).Once().
					Return(&commonModel.PaginationResponse{}, nil)

				d.On(
					"RetryBulk",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeDisbursementRetryBulkRequestReference),
				).Once().
					Return(pkgErrors.New(response.HttpErrRequest, errors.New("failed to retry disbursement bulk")))

				r.On(
					"PublishActivity",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.PtrStringMockType(),
					constant.PtrStringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeRetryDisbursementFromOpenAPIResponse)).
					Return(nil)
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   mocks.ResponseFormattingTest(40, "failed to retry disbursement bulk"),
		},
		{
			name:       "SUCCESS",
			path:       "/3a2675d6-5a0b-4bf6-9650-d8e58394ebe3",
			reqBody:    `{"id": "uuid-uuid-uuid-uuid"}`,
			reqSetting: validRequestID,
			mockSetup: func(d *serviceMocks.IDisbursementService, r *mockMQ.RabbitMQExt) {
				d.On(
					"GetBulkDisbursementForOpenApiByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeDisbursementFilterRequestReference),
					constant.Int64MockType(),
					constant.Int64MockType(),
				).Once().
					Return(commonResp, nil)

				d.On(
					"RetryBulk",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeDisbursementRetryBulkRequestReference),
				).Once().
					Return(nil)

				r.On(
					"PublishActivity",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.PtrStringMockType(),
					constant.PtrStringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeRetryDisbursementFromOpenAPIResponse)).
					Return(nil)
			},
			wantStatusCode: http.StatusOK,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			var body io.Reader

			if test.reqBody != "" {
				body = strings.NewReader(test.reqBody)
			}
			req := httptest.NewRequest(http.MethodPost, "/payouts"+test.path+"/retry", body)
			req = req.WithContext(context.WithValue(req.Context(), constant.CtxMerchantIDKey, merchantPlatformWhitelistedOldResponseFormat))

			if test.reqSetting != nil {
				test.reqSetting(req)
			}
			if test.mockSetup != nil {
				test.mockSetup(mockDisbursement, mockRabbitMQ)
			}
			if test.setHeaders != nil {
				test.setHeaders(req)
			}
			router.ServeHTTP(rec, req)
			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
		})
	}
}
