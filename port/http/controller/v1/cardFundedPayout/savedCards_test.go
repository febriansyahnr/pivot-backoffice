package cardFundedPayoutController_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	cardFundedPayoutModel "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/cardFundedPayout"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateSavedCard(t *testing.T) {
	cardFundedPayoutSvc := serviceMocks.NewICardFundedPayoutService(t)
	cfg := &config.Config{}
	controller := New(cfg, validatorExt.New(), cardFundedPayoutSvc)

	validRequest := &cardFundedPayoutModel.CreateSavedCardRequest{
		ReferenceID: "ref-123456",
	}

	tests := []struct {
		name         string
		userClaim    *userModel.UserTokenClaims
		setupMock    func()
		requestBody  func() []byte
		wantStatus   int
		wantResponse func() string
	}{
		{
			name:       "ERROR: User not found",
			wantStatus: http.StatusUnauthorized,
			wantResponse: func() string {
				return `{"code":"41","message":"user not found","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`
			},
		},
		{
			name: "ERROR: Invalid Payload - malformed JSON",
			userClaim: &userModel.UserTokenClaims{
				UUID:       "user-uuid-123",
				MerchantId: "merchant-123",
			},
			requestBody: func() []byte {
				return []byte(`{"missing-payload"}`)
			},
			wantStatus: http.StatusBadRequest,
			wantResponse: func() string {
				return `{"code":"40","message":"invalid character '}' after object key","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`
			},
		},
		{
			name: "ERROR: Validation error - missing referenceId",
			userClaim: &userModel.UserTokenClaims{
				UUID:       "user-uuid-123",
				MerchantId: "merchant-123",
			},
			requestBody: func() []byte {
				return []byte(`{}`)
			},
			wantStatus: http.StatusBadRequest,
			wantResponse: func() string {
				return `{"code":"40","message":"invalid validation","error":{"type":"API_ERROR","details":[{"field":"ReferenceID","message":"Key: 'CreateSavedCardRequest.ReferenceID' Error:Field validation for 'ReferenceID' failed on the 'required' tag"}],"traceId":""},"data":null}`
			},
		},
		{
			name: "ERROR: Service returns error",
			userClaim: &userModel.UserTokenClaims{
				UUID:       "user-uuid-123",
				MerchantId: "merchant-123",
			},
			requestBody: func() []byte {
				reqBody, _ := json.Marshal(validRequest)
				return reqBody
			},
			setupMock: func() {
				cardFundedPayoutSvc.On("CreateSavedCard", mock.Anything, mock.Anything).
					Return(nil, pkgErrors.New(response.HttpErrInternal, errors.New("service error"))).Once()
			},
			wantStatus: http.StatusInternalServerError,
			wantResponse: func() string {
				return `{"code":"99","message":"service error","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`
			},
		},
		{
			name: "SUCCESS: Create saved card",
			userClaim: &userModel.UserTokenClaims{
				UUID:       "user-uuid-123",
				MerchantId: "merchant-123",
			},
			requestBody: func() []byte {
				reqBody, _ := json.Marshal(validRequest)
				return reqBody
			},
			setupMock: func() {
				cardFundedPayoutSvc.On("CreateSavedCard", mock.Anything, mock.Anything).
					Return(&cardFundedPayoutModel.CreateSavedCardResponse{
						ReferenceID: "ref-123456",
						PaymentUrl:  "https://payment.url",
					}, nil).Once()
			},
			wantStatus: http.StatusOK,
			wantResponse: func() string {
				return `{"code":"00","data":{"referenceId":"ref-123456","paymentUrl":"https://payment.url"},"message":"OK"}`
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setupMock != nil {
				test.setupMock()
			}

			reqBody := []byte{}
			if test.requestBody != nil {
				reqBody = test.requestBody()
			}
			req := httptest.NewRequest(http.MethodPost, "/card-funded-payout/saved-cards", bytes.NewBuffer(reqBody))
			rec := httptest.NewRecorder()

			ctx := req.Context()
			if test.userClaim != nil {
				ctx = context.WithValue(ctx, constant.CtxUserInfoKey, test.userClaim)
			}
			req = req.WithContext(ctx)

			controller.CreateSavedCard(rec, req)

			assert.Equal(t, test.wantStatus, rec.Result().StatusCode)
			if test.wantResponse != nil {
				if !assert.JSONEq(t, test.wantResponse(), rec.Body.String()) {
					t.Log("Result:", rec.Body.String())
				}
			}
		})
	}
}

func TestGetSavedCardList(t *testing.T) {
	cardFundedPayoutSvc := serviceMocks.NewICardFundedPayoutService(t)
	cfg := &config.Config{}
	controller := New(cfg, validatorExt.New(), cardFundedPayoutSvc)

	now := time.Now().UTC()
	startDate := now.Add(-24 * time.Hour).Format(util.UTCLayout)
	endDate := now.Format(util.UTCLayout)

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
			name: "ERROR: Invalid page format",
			userClaim: &userModel.UserTokenClaims{
				UUID:       "user-uuid-123",
				MerchantId: "merchant-123",
			},
			queryParams: "?page=invalid",
			wantStatus:  http.StatusBadRequest,
			wantResponse: func() string {
				return `{"code":"40","message":"invalid page format. Use number format instead","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`
			},
		},
		{
			name: "ERROR: Invalid perPage format",
			userClaim: &userModel.UserTokenClaims{
				UUID:       "user-uuid-123",
				MerchantId: "merchant-123",
			},
			queryParams: "?perPage=invalid",
			wantStatus:  http.StatusBadRequest,
			wantResponse: func() string {
				return `{"code":"40","message":"invalid perPage format. Use number format instead","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`
			},
		},
		{
			name: "ERROR: Invalid startDate format",
			userClaim: &userModel.UserTokenClaims{
				UUID:       "user-uuid-123",
				MerchantId: "merchant-123",
			},
			queryParams: "?startDate=invalid-date",
			wantStatus:  http.StatusBadRequest,
			wantResponse: func() string {
				return `{"code":"40","message":"invalid startDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`
			},
		},
		{
			name: "ERROR: Invalid endDate format",
			userClaim: &userModel.UserTokenClaims{
				UUID:       "user-uuid-123",
				MerchantId: "merchant-123",
			},
			queryParams: "?endDate=invalid-date",
			wantStatus:  http.StatusBadRequest,
			wantResponse: func() string {
				return `{"code":"40","message":"invalid endDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`
			},
		},
		{
			name: "ERROR: Service returns error",
			userClaim: &userModel.UserTokenClaims{
				UUID:       "user-uuid-123",
				MerchantId: "merchant-123",
			},
			queryParams: "",
			setupMock: func() {
				cardFundedPayoutSvc.On("GetSavedCardList", mock.Anything, mock.Anything).
					Return(nil, pkgErrors.New(response.HttpErrInternal, errors.New("service error"))).Once()
			},
			wantStatus: http.StatusInternalServerError,
			wantResponse: func() string {
				return `{"code":"99","message":"service error","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`
			},
		},
		{
			name: "SUCCESS: Get saved card list with default pagination",
			userClaim: &userModel.UserTokenClaims{
				UUID:       "user-uuid-123",
				MerchantId: "merchant-123",
			},
			queryParams: "",
			setupMock: func() {
				cardFundedPayoutSvc.On("GetSavedCardList", mock.Anything, mock.Anything).
					Return(&commonModel.PaginationResponse{
						Data: []cardFundedPayoutModel.GetSavedCardResponse{},
						Meta: commonModel.Meta{
							Page:       1,
							PerPage:    1000,
							TotalItems: 0,
							TotalPages: 0,
						},
					}, nil).Once()
			},
			wantStatus: http.StatusOK,
			wantResponse: func() string {
				return `{"code":"00","message":"OK","data":[],"pagination":{"page":1,"perPage":1000,"totalItems":0,"totalPages":0}}`
			},
		},
		{
			name: "SUCCESS: Get saved card list with custom pagination",
			userClaim: &userModel.UserTokenClaims{
				UUID:       "user-uuid-123",
				MerchantId: "merchant-123",
			},
			queryParams: "?page=2&perPage=10",
			setupMock: func() {
				cardFundedPayoutSvc.On("GetSavedCardList", mock.Anything, mock.Anything).
					Return(&commonModel.PaginationResponse{
						Data: []cardFundedPayoutModel.GetSavedCardResponse{
							{
								ID:             "customer-123",
								CardName:       "VISA",
								PaymentChannel: "CHANNEL",
								IssuingBank:    "BANK",
								Last4:          "1234",
								ExpiryMonth:    "12",
								ExpiryYear:     "2025",
								CardOrigin:     "LOCAL",
							},
						},
						Meta: commonModel.Meta{
							Page:       2,
							PerPage:    10,
							TotalItems: 11,
							TotalPages: 2,
						},
					}, nil).Once()
			},
			wantStatus: http.StatusOK,
			wantResponse: func() string {
				return `{"code":"00","message":"OK","data":[{"id":"customer-123","cardName":"VISA","paymentChannel":"CHANNEL","issuingBank":"BANK","last4":"1234","cardOrigin":"LOCAL", "expiryMonth":"12","expiryYear":"2025"}],"pagination":{"page":2,"perPage":10,"totalItems":11,"totalPages":2}}`
			},
		},
		{
			name: "SUCCESS: Get saved card list with date filter",
			userClaim: &userModel.UserTokenClaims{
				UUID:       "user-uuid-123",
				MerchantId: "merchant-123",
			},
			queryParams: "?startDate=" + url.QueryEscape(startDate) + "&endDate=" + url.QueryEscape(endDate) + "&sort=ASC&sortBy=updatedAt",
			setupMock: func() {
				cardFundedPayoutSvc.On("GetSavedCardList", mock.Anything, mock.Anything).
					Return(&commonModel.PaginationResponse{
						Data: []cardFundedPayoutModel.GetSavedCardResponse{},
						Meta: commonModel.Meta{
							Page:       1,
							PerPage:    1000,
							TotalItems: 0,
							TotalPages: 0,
						},
					}, nil).Once()
			},
			wantStatus: http.StatusOK,
			wantResponse: func() string {
				return `{"code":"00","message":"OK","data":[],"pagination":{"page":1,"perPage":1000,"totalItems":0,"totalPages":0}}`
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setupMock != nil {
				test.setupMock()
			}

			req := httptest.NewRequest(http.MethodGet, "/card-funded-payout/saved-cards"+test.queryParams, nil)
			rec := httptest.NewRecorder()

			ctx := req.Context()
			if test.userClaim != nil {
				ctx = context.WithValue(ctx, constant.CtxUserInfoKey, test.userClaim)
			}
			req = req.WithContext(ctx)

			controller.GetSavedCardList(rec, req)

			assert.Equal(t, test.wantStatus, rec.Result().StatusCode)
			if test.wantResponse != nil {
				if !assert.JSONEq(t, test.wantResponse(), rec.Body.String()) {
					t.Log("Result:", rec.Body.String())
				}
			}
		})
	}
}
