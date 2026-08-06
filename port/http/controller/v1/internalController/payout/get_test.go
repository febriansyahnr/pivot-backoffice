package internalPayoutController_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	mockMQ "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	mockRedis "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	internalPayoutController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/internalController/payout"
	"github.com/redis/go-redis/v9"

	chi "github.com/go-chi/chi/v5"
	validator "github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestFindByBulkId(pt *testing.T) {
	mockDisbursement := serviceMocks.NewIDisbursementService(pt)
	mockAccountInquirySvc := serviceMocks.NewIAccountInquiryService(pt)
	mockRedisClient := mockRedis.NewIRedisExt(pt)
	conf := config.Config{
		Environment: "development",
	}

	router := chi.NewRouter()
	router.Get(
		"/payouts/{id}", internalPayoutController.New(
			&conf, validator.New(), mockDisbursement, serviceMocks.NewIMerchantService(pt), mockAccountInquirySvc, mockMQ.NewRabbitMQExt(pt),
			internalPayoutController.WithRedisClient(mockRedisClient),
		).FindByBulkId,
	)

	validRequestID := func(req *http.Request) {
		*req = *req.WithContext(context.WithValue(req.Context(), constant.CtxMerchantInfo, &merchantModel.MerchantAuthTokenClaims{
			MerchantId: "12345",
		}))
		req.Header.Set("Content-Type", "application/json")
	}

	tests := []struct {
		name           string
		path           string
		reqSetting     func(r *http.Request)
		mockSetup      func(d *serviceMocks.IDisbursementService)
		setHeaders     func(req *http.Request)
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:Invalid request ID",
			path:           "/invalid",
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   wrapErrOpenApiNonSnap(40, "uuid is required"),
		},
		{
			name:           "ERROR:Invalid merchant info",
			path:           "/3a2675d6-5a0b-4bf6-9650-d8e58394ebe3",
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   wrapErrOpenApiNonSnap(41, "merchant not found"),
		},
		{
			name:       "ERROR:Request with reference ID[data not found]",
			path:       "/3a2675d6-5a0b-4bf6-9650-d8e58394ebe3?referenceId=1",
			reqSetting: validRequestID,
			mockSetup: func(d *serviceMocks.IDisbursementService) {
				// Mock redis Get operation to return empty string
				mockRedisClient.On("Get", mock.Anything, mock.AnythingOfType("string")).Return(redis.NewStringCmd(context.Background()))

				d.On(
					"GetBulkDisbursementForOpenApiByReferenceID", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"),
				).Once().Return(nil, pkgErrors.New(response.HttpErrRequest, errors.New("data not found")))
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   wrapErrOpenApiNonSnap(40, "data not found"),
		},
		{
			name:       "SUCCESS:Request with reference ID[data found]",
			path:       "/3a2675d6-5a0b-4bf6-9650-d8e58394ebe3?referenceId=2",
			reqSetting: validRequestID,
			mockSetup: func(d *serviceMocks.IDisbursementService) {
				// Mock redis Get operation to return empty string
				mockRedisClient.On("Get", mock.Anything, mock.AnythingOfType("string")).Return(redis.NewStringCmd(context.Background()))

				d.On(
					"GetBulkDisbursementForOpenApiByReferenceID", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"),
				).Return(&disbursementModel.GetBulkDisbursementForOpenApiByReferenceIDResponse{
					UUID:       "3a2675d6-5a0b-4bf6-9650-d8e58394ebe3",
					MerchantID: "123456",
					PayoutResults: disbursementModel.PayoutResultObject{
						TotalSuccessCount:    1,
						TotalSuccessAmount:   10_000,
						TotalPendingCount:    0,
						TotalPendingAmount:   0,
						TotalFailedCount:     0,
						TotalFailedAmount:    0,
						TotalCancelledCount:  0,
						TotalCancelledAmount: 0,
					},
					Payouts: disbursementModel.PayoutObject{
						ReferenceID: "REF-001",
						ChannelCode: "BRI",
						Description: "DES",
					},
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"Success","data":{"uuid":"3a2675d6-5a0b-4bf6-9650-d8e58394ebe3","merchantId":"123456","payoutResults":{"totalPendingCount":0,"totalPendingAmount":0,"totalSuccessCount":1,"totalSuccessAmount":10000,"totalFailedCount":0,"totalFailedAmount":0},"payouts":{"referenceId":"REF-001","channelCode":"BRI","channelInformation":{"accountNumber":"","accountName":""},"amount":{"currency":"","value":""},"description":"DES","inquiryId":"", "reason":"","status":"","created":"0001-01-01T00:00:00Z","updated":"0001-01-01T00:00:00Z"}}}`,
		},
		{
			name:       "ERROR:Invalid page format",
			path:       "/3a2675d6-5a0b-4bf6-9650-d8e58394ebe3?page=invalid",
			reqSetting: validRequestID,
			mockSetup: func(d *serviceMocks.IDisbursementService) {
				// Mock redis Get operation to return empty string
				mockRedisClient.On("Get", mock.Anything, mock.AnythingOfType("string")).Return(redis.NewStringCmd(context.Background()))
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   wrapErrOpenApiNonSnap(40, "invalid page format. Use number format instead"),
		},
		{
			name:       "ERROR:Invalid per page format",
			path:       "/3a2675d6-5a0b-4bf6-9650-d8e58394ebe3?page=1&perPage=invalid",
			reqSetting: validRequestID,
			mockSetup: func(d *serviceMocks.IDisbursementService) {
				// Mock redis Get operation to return empty string
				mockRedisClient.On("Get", mock.Anything, mock.AnythingOfType("string")).Return(redis.NewStringCmd(context.Background()))
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   wrapErrOpenApiNonSnap(40, "invalid perPage format. Use number format instead"),
		},
		{
			name:       "ERROR:Invalid per page value",
			path:       "/3a2675d6-5a0b-4bf6-9650-d8e58394ebe3?page=1&perPage=0",
			reqSetting: validRequestID,
			mockSetup: func(d *serviceMocks.IDisbursementService) {
				// Mock redis Get operation to return empty string
				mockRedisClient.On("Get", mock.Anything, mock.AnythingOfType("string")).Return(redis.NewStringCmd(context.Background()))
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   wrapErrOpenApiNonSnap(40, "perPage value must be 1 to 100"),
		},
		{
			name:       "ERROR:Get bulk disbursement for open API by ID",
			path:       "/3a2675d6-5a0b-4bf6-9650-d8e58394ebe3?page=1&perPage=10",
			reqSetting: validRequestID,
			mockSetup: func(d *serviceMocks.IDisbursementService) {
				// Mock redis Get operation to return empty string
				mockRedisClient.On("Get", mock.Anything, mock.AnythingOfType("string")).Return(redis.NewStringCmd(context.Background()))

				d.On(
					"GetBulkDisbursementForOpenApiByID", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("*disbursementModel.GetDisbursementFilterRequest"), mock.AnythingOfType("int64"), mock.AnythingOfType("int64"),
				).Once().Return(nil, pkgErrors.New(response.HttpErrRequest, constant.ErrBulkDisbursementNotFound))
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   wrapErrOpenApiNonSnap(40, "bulk disbursement not found"),
		},
		{
			name:       "SUCCESS",
			path:       "/3a2675d6-5a0b-4bf6-9650-d8e58394ebe3?page=1&perPage=10",
			reqSetting: validRequestID,
			mockSetup: func(d *serviceMocks.IDisbursementService) {
				// Mock redis Get operation to return empty string
				mockRedisClient.On("Get", mock.Anything, mock.AnythingOfType("string")).Return(redis.NewStringCmd(context.Background()))

				d.On(
					"GetBulkDisbursementForOpenApiByID", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("*disbursementModel.GetDisbursementFilterRequest"), mock.AnythingOfType("int64"), mock.AnythingOfType("int64"),
				).Return(&commonModel.PaginationResponse{
					Data: map[string]string{"message": "OK"},
					Meta: commonModel.Meta{
						Page: 1, PerPage: 10, TotalItems: 12, TotalPages: 2,
					},
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"Success","data":{"message":"OK"},"pagination":{"page":1,"perPage":10,"totalItems":12,"totalPages":2}}`,
		},
		{
			name: "SUCCESS in behalf of submerchant",
			path: "/3a2675d6-5a0b-4bf6-9650-d8e58394ebe3?page=1&perPage=10",
			setHeaders: func(req *http.Request) {
				req.Header.Set(constant.HeaderXSubMerchantID, uuid.Max.String())
			},
			reqSetting: validRequestID,
			mockSetup: func(d *serviceMocks.IDisbursementService) {
				// Mock redis Get operation to return empty string
				mockRedisClient.On("Get", mock.Anything, mock.AnythingOfType("string")).Return(redis.NewStringCmd(context.Background()))

				d.On(
					"GetBulkDisbursementForOpenApiByID", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("*disbursementModel.GetDisbursementFilterRequest"), mock.AnythingOfType("int64"), mock.AnythingOfType("int64"),
				).Return(&commonModel.PaginationResponse{
					Data: map[string]string{"message": "OK"},
					Meta: commonModel.Meta{
						Page: 1, PerPage: 10, TotalItems: 12, TotalPages: 2,
					},
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"Success","data":{"message":"OK"},"pagination":{"page":1,"perPage":10,"totalItems":12,"totalPages":2}}`,
		},
	}
	for _, test := range tests {
		pt.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/payouts"+test.path, nil)
			req = req.WithContext(context.WithValue(req.Context(), constant.CtxMerchantIDKey, merchantPlatformWhitelistedOldResponseFormat))

			if test.reqSetting != nil {
				test.reqSetting(req)
			}
			if test.mockSetup != nil {
				test.mockSetup(mockDisbursement)
			}
			if test.setHeaders != nil {
				test.setHeaders(req)
			}
			router.ServeHTTP(rec, req)
			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
