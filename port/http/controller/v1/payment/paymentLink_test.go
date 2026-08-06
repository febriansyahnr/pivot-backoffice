package payment_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	s "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/payment"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreatePaymentLink(t *testing.T) {
	unifiedPaymentService := serviceMocks.NewIUnifiedPaymentService(t)
	handler := New(nil, validator.New(), nil, WithUnifiedPaymentService(unifiedPaymentService))

	router := chi.NewRouter()
	router.Post("/payment-links", handler.CreatePaymentLink)

	userTokenClaims := &user.UserTokenClaims{
		UUID:       "user-123",
		MerchantId: "merchant-123",
	}

	validPayload := unifiedPaymentModel.DashboardPaymentLinkCreateRequest{
		ClientReferenceID: "ref-123",
		ExpiredAt:         time.Now().Add(time.Hour),
		Amount:            unifiedPaymentModel.Amount{Value: 50000},
		Customer:          unifiedPaymentModel.PaymentLinkCustomerRequest{Email: "test@example.com"},
	}

	successResponse := &unifiedPaymentModel.UnifiedPaymentSessionResponse{
		ID:                "session-123",
		ClientReferenceID: "ref-123",
		Amount:            unifiedPaymentModel.Amount{Value: 50000},
		PaymentUrl:        "https://payment.link/123",
		ShortPaymentUrl:   "https://payment.link/123",
		Status:            "ACTIVE",
		CreatedAt:         time.Now(),
	}

	tests := []struct {
		name            string
		userTokenClaims *user.UserTokenClaims
		requestBody     interface{}
		setupMock       func()
		wantStatusCode  int
		wantRespCheck   func(t *testing.T, body string)
	}{
		{
			name:           "error - user not found in context",
			wantStatusCode: http.StatusUnauthorized,
			requestBody:    validPayload,
			wantRespCheck: func(t *testing.T, body string) {
				expectedErr := c.WrapErrApiRespForTest(41, s.ErrTypeAPI, "user not found")
				assert.JSONEq(t, expectedErr, body)
			},
		},
		{
			name:            "error - invalid json body",
			userTokenClaims: userTokenClaims,
			requestBody:     "invalid json",
			wantStatusCode:  http.StatusBadRequest,
			wantRespCheck: func(t *testing.T, body string) {
				expectedErr := c.WrapErrApiRespForTest(40, s.ErrTypeAPI, "invalid request payload")
				assert.JSONEq(t, expectedErr, body)
			},
		},
		{
			name:            "error - validation fails - missing required fields",
			userTokenClaims: userTokenClaims,
			requestBody:     unifiedPaymentModel.DashboardPaymentLinkCreateRequest{
				// Missing required fields
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespCheck: func(t *testing.T, body string) {
				// Should contain validation error
				assert.Contains(t, body, "\"code\":\"40\"")
				assert.Contains(t, body, "ClientReferenceID")
			},
		},
		{
			name:            "error - custom validation fails - amount too low",
			userTokenClaims: userTokenClaims,
			requestBody: unifiedPaymentModel.DashboardPaymentLinkCreateRequest{
				ClientReferenceID: "ref-123",
				ExpiredAt:         time.Now().Add(time.Hour),
				Amount:            unifiedPaymentModel.Amount{Value: 5000}, // Below minimum
				Customer:          unifiedPaymentModel.PaymentLinkCustomerRequest{Email: "test@example.com"},
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespCheck: func(t *testing.T, body string) {
				expectedErr := c.WrapErrApiRespForTest(40, s.ErrTypeAPI, "min amount is 10000")
				assert.JSONEq(t, expectedErr, body)
			},
		},
		{
			name:            "error - service returns error",
			userTokenClaims: userTokenClaims,
			requestBody:     validPayload,
			setupMock: func() {
				unifiedPaymentService.On("CreateDashboardPaymentLink", mock.Anything, mock.MatchedBy(func(req *unifiedPaymentModel.DashboardPaymentLinkCreateRequest) bool {
					return req.MerchantID == "merchant-123" && req.UserID == "user-123"
				})).Return(nil, errors.New(s.HttpErrDatabase, c.ErrSomeErrorForUnitTest)).Once()
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespCheck: func(t *testing.T, body string) {
				// Error gets wrapped differently by the service layer
				assert.Contains(t, body, "\"code\":\"98\"")
				assert.Contains(t, body, "some error")
			},
		},
		{
			name:            "success - payment link created",
			userTokenClaims: userTokenClaims,
			requestBody:     validPayload,
			setupMock: func() {
				unifiedPaymentService.On("CreateDashboardPaymentLink", mock.Anything, mock.MatchedBy(func(req *unifiedPaymentModel.DashboardPaymentLinkCreateRequest) bool {
					return req.MerchantID == "merchant-123" &&
						req.UserID == "user-123" &&
						req.ClientReferenceID == "ref-123" &&
						req.Amount.Value == 50000
				})).Return(successResponse, nil).Once()
			},
			wantStatusCode: http.StatusOK,
			wantRespCheck: func(t *testing.T, body string) {
				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				assert.NoError(t, err)

				assert.Equal(t, "00", response["code"])
				assert.Equal(t, "OK", response["message"])

				data := response["data"].(map[string]interface{})
				assert.Equal(t, "session-123", data["id"])
				assert.Equal(t, "ref-123", data["clientReferenceId"])
				assert.Equal(t, "https://payment.link/123", data["paymentLink"])
				assert.Equal(t, "ACTIVE", data["status"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			var reqBody string
			if str, ok := tt.requestBody.(string); ok {
				reqBody = str
			} else {
				bodyBytes, _ := json.Marshal(tt.requestBody)
				reqBody = string(bodyBytes)
			}

			req := httptest.NewRequest(http.MethodPost, "/payment-links", bytes.NewBuffer([]byte(reqBody)))
			req.Header.Set("Content-Type", "application/json")

			if tt.userTokenClaims != nil {
				ctx := context.WithValue(req.Context(), c.CtxUserInfoKey, tt.userTokenClaims)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatusCode, rr.Code)

			if tt.wantRespCheck != nil {
				tt.wantRespCheck(t, strings.TrimSpace(rr.Body.String()))
			}

			unifiedPaymentService.AssertExpectations(t)
		})
	}
}
