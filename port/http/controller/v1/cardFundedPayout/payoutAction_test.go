package cardFundedPayoutController

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreatePayout(t *testing.T) {
	validator := validator.New()
	service := serviceMocks.NewICardFundedPayoutService(t)

	validUserClaims := &userModel.UserTokenClaims{
		UUID:       uuid.NewString(),
		Name:       "John Doe", // NOSONAR
		MerchantId: uuid.NewString(),
	}

	validRequestBody := model.CreatePayoutRequest{
		VendorID:    uuid.NewString(),
		ReferenceID: "REF-123", // NOSONAR
		Amount: commonModel.AmountRequest{
			Currency: constant.CurrencyIDR,
			Value:    100000,
		},
		Remarks:          "Test payout", // NOSONAR
		SettlementMethod: constant.PaymentSettlementMethodStandard,
		CardID:           uuid.NewString(),
	}

	cfg := &config.Config{ServiceName: "testing"}
	handler := New(cfg, validator, service)

	router := chi.NewRouter()
	router.Post("/", handler.CreatePayout)

	testCases := []struct {
		name             string
		requestBody      any
		userClaim        *userModel.UserTokenClaims
		setupMock        func()
		wantStatusCode   int
		wantResponseBody string
	}{

		{
			name:             "ERROR: User not in Context",
			userClaim:        nil,
			requestBody:      validRequestBody,
			wantStatusCode:   http.StatusUnauthorized,
			wantResponseBody: `{"code":"41","message":"user not found","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:             "ERROR: Invalid request body",
			userClaim:        validUserClaims,
			requestBody:      "invalid json",
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: `{"code":"40","message":"json: cannot unmarshal string into Go value of type cardFundedPayoutModel.CreatePayoutRequest","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:      "ERROR: Validation error - missing required fields",
			userClaim: validUserClaims,
			requestBody: func() model.CreatePayoutRequest {
				req := validRequestBody
				req.ReferenceID = ""
				return req
			}(),
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: `{"code":"40","message":"invalid validation","error":{"type":"API_ERROR","details":[{"field":"ReferenceID","message":"Key: 'CreatePayoutRequest.ReferenceID' Error:Field validation for 'ReferenceID' failed on the 'required' tag"}],"traceId":""},"data":null}`,
		},
		{
			name:        "ERROR: Service error",
			userClaim:   validUserClaims,
			requestBody: validRequestBody,
			setupMock: func() {
				service.On("CreatePayout", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
			},
			wantStatusCode:   http.StatusInternalServerError,
			wantResponseBody: `{"code":"99","message":"assert.AnError general error for testing","error":{"type":"UNKNOWN","details":[],"traceId":""},"data":null}`,
		},
		{
			name:        "SUCCESS",
			userClaim:   validUserClaims,
			requestBody: validRequestBody,
			setupMock: func() {
				service.On(
					"CreatePayout", mock.Anything, mock.Anything,
				).Once().Return(&model.PayoutActionResponse{
					ID:          "6e34cfca-de00-4aa1-bc6b-069d878012c8",
					ReferenceID: "REF-123",
					CreatedAt:   util.ValueToPtr(time.Date(2026, 3, 16, 12, 0, 0, 0, time.UTC)),
				}, nil)
			},
			wantStatusCode:   http.StatusOK,
			wantResponseBody: `{"code":"00","message":"OK","data":{"id":"6e34cfca-de00-4aa1-bc6b-069d878012c8","vendorId":"","vendorName":"","referenceId":"REF-123","feeAmount":0,"amount":{"currency":"","value":0},"remarks":"","settlementMethod":"","cardId":"","cardName":"","createdAt":"2026-03-16T12:00:00Z"}}`,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {

			if tt.setupMock != nil {
				tt.setupMock()
			}
			bodyBytes, _ := json.Marshal(tt.requestBody)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))

			if tt.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tt.userClaim))
			}

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, tt.wantResponseBody, rec.Body.String()) {
				t.Log("Actual:", rec.Body.String())
			}

			service.AssertExpectations(t)
		})
	}
}

func TestApprovePayout(t *testing.T) {
	validator := validator.New()
	service := serviceMocks.NewICardFundedPayoutService(t)

	payoutID := "6e34cfca-de00-4aa1-bc6b-069d878012c8"
	validUserClaims := &userModel.UserTokenClaims{
		UUID:       uuid.NewString(),
		Name:       "John Doe", // NOSONAR
		MerchantId: uuid.NewString(),
	}
	validRequestBody := model.ApprovePayoutRequest{
		ID:  payoutID,
		CVC: "123",
	}

	cfg := &config.Config{ServiceName: "testing"}
	handler := New(cfg, validator, service)

	router := chi.NewRouter()
	router.Post("/{payoutId}/approve", handler.ApprovePayout)

	testCases := []struct {
		name             string
		requestBody      any
		userClaim        *userModel.UserTokenClaims
		payoutID         string
		setupMock        func()
		wantStatusCode   int
		wantResponseBody string
	}{
		{
			name:        "SUCCESS",
			requestBody: validRequestBody,
			userClaim:   validUserClaims,
			payoutID:    payoutID,
			setupMock: func() {
				service.On(
					"ApprovePayout", mock.Anything, mock.Anything,
				).Once().Return(&model.PayoutActionResponse{
					ID:                payoutID,
					ReferenceID:       "REF-123",
					AuthenticationUrl: util.ValueToPtr("http://example.com/auth"),
					ApprovedAt:        util.ValueToPtr(time.Date(2026, 3, 16, 12, 0, 0, 0, time.UTC)),
				}, nil)
			},
			wantStatusCode:   http.StatusOK,
			wantResponseBody: `{"code":"00","message":"OK","data":{"id":"6e34cfca-de00-4aa1-bc6b-069d878012c8","vendorId":"","vendorName":"","referenceId":"REF-123","feeAmount":0,"amount":{"currency":"","value":0},"remarks":"","settlementMethod":"","cardId":"","cardName":"","authenticationUrl":"http://example.com/auth","approvedAt":"2026-03-16T12:00:00Z"}}`,
		},
		{
			name:             "ERROR: User not in Context",
			requestBody:      validRequestBody,
			userClaim:        nil,
			payoutID:         payoutID,
			wantStatusCode:   http.StatusUnauthorized,
			wantResponseBody: `{"code":"41","message":"user not found","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:             "ERROR: Invalid request body",
			requestBody:      "invalid json",
			userClaim:        validUserClaims,
			payoutID:         payoutID,
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: `{"code":"40","message":"json: cannot unmarshal string into Go value of type cardFundedPayoutModel.ApprovePayoutRequest","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name: "ERROR: Validation error - missing CVC",
			requestBody: func() model.ApprovePayoutRequest {
				req := validRequestBody
				req.CVC = ""
				return req
			}(),
			userClaim:        validUserClaims,
			payoutID:         payoutID,
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: `{"code":"40","message":"invalid validation","error":{"type":"API_ERROR","details":[{"field":"CVC","message":"Key: 'ApprovePayoutRequest.CVC' Error:Field validation for 'CVC' failed on the 'required' tag"}],"traceId":""},"data":null}`,
		},
		{
			name: "ERROR: Validation error - CVC too short",
			requestBody: func() model.ApprovePayoutRequest {
				req := validRequestBody
				req.CVC = "12"
				return req
			}(),
			userClaim:        validUserClaims,
			payoutID:         payoutID,
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: `{"code":"40","message":"invalid validation","error":{"type":"API_ERROR","details":[{"field":"CVC","message":"Key: 'ApprovePayoutRequest.CVC' Error:Field validation for 'CVC' failed on the 'min' tag"}],"traceId":""},"data":null}`,
		},
		{
			name:        "ERROR: Service error",
			requestBody: validRequestBody,
			userClaim:   validUserClaims,
			payoutID:    payoutID,
			setupMock: func() {
				service.On("ApprovePayout", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
			},
			wantStatusCode:   http.StatusInternalServerError,
			wantResponseBody: `{"code":"99","message":"assert.AnError general error for testing","error":{"type":"UNKNOWN","details":[],"traceId":""},"data":null}`,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}
			bodyBytes, _ := json.Marshal(tt.requestBody)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/"+tt.payoutID+"/approve", bytes.NewReader(bodyBytes))

			if tt.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tt.userClaim))
			}

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, tt.wantResponseBody, rec.Body.String()) {
				t.Log("Actual:", rec.Body.String())
			}

			service.AssertExpectations(t)
		})
	}
}

func TestRejectPayout(t *testing.T) {
	validator := validator.New()
	service := serviceMocks.NewICardFundedPayoutService(t)

	payoutID := "6e34cfca-de00-4aa1-bc6b-069d878012c8"
	validUserClaims := &userModel.UserTokenClaims{
		UUID:       uuid.NewString(),
		Name:       "John Doe", // NOSONAR
		MerchantId: uuid.NewString(),
	}
	validRequestBody := model.RejectPayoutRequest{
		ID:     payoutID,
		Reason: "Invalid card details",
	}

	cfg := &config.Config{ServiceName: "testing"}
	handler := New(cfg, validator, service)

	router := chi.NewRouter()
	router.Post("/{payoutId}/reject", handler.RejectPayout)

	testCases := []struct {
		name             string
		requestBody      any
		userClaim        *userModel.UserTokenClaims
		payoutID         string
		setupMock        func()
		wantStatusCode   int
		wantResponseBody string
	}{
		{
			name:        "SUCCESS",
			requestBody: validRequestBody,
			userClaim:   validUserClaims,
			payoutID:    payoutID,
			setupMock: func() {
				service.On(
					"RejectPayout", mock.Anything, mock.Anything,
				).Once().Return(&model.PayoutActionResponse{
					ID:           payoutID,
					ReferenceID:  "REF-123",
					RejectReason: util.ValueToPtr("some reason"),
					RejectedAt:   util.ValueToPtr(time.Date(2026, 3, 16, 12, 0, 0, 0, time.UTC)),
				}, nil)
			},
			wantStatusCode:   http.StatusOK,
			wantResponseBody: `{"code":"00","message":"OK","data":{"id":"6e34cfca-de00-4aa1-bc6b-069d878012c8","vendorId":"","vendorName":"","referenceId":"REF-123","feeAmount":0,"amount":{"currency":"","value":0},"remarks":"","settlementMethod":"","cardId":"","cardName":"","rejectReason":"some reason","rejectedAt":"2026-03-16T12:00:00Z"}}`,
		},
		{
			name:             "ERROR: User not in Context",
			requestBody:      validRequestBody,
			userClaim:        nil,
			payoutID:         payoutID,
			wantStatusCode:   http.StatusUnauthorized,
			wantResponseBody: `{"code":"41","message":"user not found","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:             "ERROR: Invalid request body",
			requestBody:      "invalid json",
			userClaim:        validUserClaims,
			payoutID:         payoutID,
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: `{"code":"40","message":"json: cannot unmarshal string into Go value of type cardFundedPayoutModel.RejectPayoutRequest","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name: "ERROR: Validation error - missing reason",
			requestBody: func() model.RejectPayoutRequest {
				req := validRequestBody
				req.Reason = ""
				return req
			}(),
			userClaim:        validUserClaims,
			payoutID:         payoutID,
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: `{"code":"40","message":"invalid validation","error":{"type":"API_ERROR","details":[{"field":"Reason","message":"Key: 'RejectPayoutRequest.Reason' Error:Field validation for 'Reason' failed on the 'required' tag"}],"traceId":""},"data":null}`,
		},
		{
			name: "ERROR: Validation error - reason exceeds max length",
			requestBody: func() model.RejectPayoutRequest {
				req := validRequestBody
				req.Reason = "This is a very long reason that exceeds the maximum allowed length of 255 characters. This is a very long reason that exceeds the maximum allowed length of 255 characters. This is a very long reason that exceeds the maximum allowed length of 255 characters. This is a very long reason that exceeds the maximum allowed length of 255 characters."
				return req
			}(),
			userClaim:        validUserClaims,
			payoutID:         payoutID,
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: `{"code":"40","message":"invalid validation","error":{"type":"API_ERROR","details":[{"field":"Reason","message":"Key: 'RejectPayoutRequest.Reason' Error:Field validation for 'Reason' failed on the 'max' tag"}],"traceId":""},"data":null}`,
		},
		{
			name:        "ERROR: Service error",
			requestBody: validRequestBody,
			userClaim:   validUserClaims,
			payoutID:    payoutID,
			setupMock: func() {
				service.On("RejectPayout", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
			},
			wantStatusCode:   http.StatusInternalServerError,
			wantResponseBody: `{"code":"99","message":"assert.AnError general error for testing","error":{"type":"UNKNOWN","details":[],"traceId":""},"data":null}`,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}
			bodyBytes, _ := json.Marshal(tt.requestBody)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/"+tt.payoutID+"/reject", bytes.NewReader(bodyBytes))

			if tt.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tt.userClaim))
			}

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, tt.wantResponseBody, rec.Body.String()) {
				t.Log("Actual:", rec.Body.String())
			}

			service.AssertExpectations(t)
		})
	}
}
