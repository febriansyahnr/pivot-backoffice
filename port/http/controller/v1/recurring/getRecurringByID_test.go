package recurring

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	recurringContractModel "github.com/paper-indonesia/pivot-backoffice/internal/model/recurringContract"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockServices "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetRecurringByID(pt *testing.T) {
	const validRecurringID = "e39d1d21-0a7b-49e5-8cd9-404ac75d54be"
	const merchantID = "merchant-uuid-123"

	validUserClaims := &userModel.UserTokenClaims{
		UUID:       "user-uuid-123",
		Email:      "test@example.com",
		Name:       "Test User",
		MerchantId: merchantID,
	}

	tests := []struct {
		name           string
		recurringID    string
		userClaims     *userModel.UserTokenClaims
		mockSetup      func(m *mockServices.IRecurringContractService)
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:        "ERROR:User not found in context",
			recurringID: validRecurringID,
			userClaims:  nil,
			mockSetup: func(m *mockServices.IRecurringContractService) {
				// No mock setup needed - should fail before service call
			},
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   `{"code":"41","data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message":"user not found"}`,
		},
		{
			name:        "ERROR:Empty recurring ID",
			recurringID: "",
			userClaims:  validUserClaims,
			mockSetup: func(m *mockServices.IRecurringContractService) {
				// No mock setup needed - should fail validation
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message":"invalid request payload"}`,
		},
		{
			name:        "ERROR:Invalid UUID format",
			recurringID: "invalid-uuid",
			userClaims:  validUserClaims,
			mockSetup: func(m *mockServices.IRecurringContractService) {
				// No mock setup needed - should fail validation
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message":"invalid request payload"}`,
		},
		{
			name:        "ERROR:Service returns error",
			recurringID: validRecurringID,
			userClaims:  validUserClaims,
			mockSetup: func(m *mockServices.IRecurringContractService) {
				m.On(
					"GetRecurringByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.MatchedBy(func(req recurringContractModel.GetRecurringByIDRequest) bool {
						return req.RecurringID == validRecurringID && req.MerchantID == merchantID
					}),
				).Once().Return(nil, errors.New("database error"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","data":null,"error":{"details":[],"traceId":"","type":"UNKNOWN"},"message":"database error"}`,
		},
		{
			name:        "SUCCESS:Get recurring by ID",
			recurringID: validRecurringID,
			userClaims:  validUserClaims,
			mockSetup: func(m *mockServices.IRecurringContractService) {
				mockResponse := &recurringContractModel.GetRecurringByIDDashboardResponse{
					UUID:              validRecurringID,
					MerchantID:        merchantID,
					CustomerID:        "customer-uuid-123",
					ClientReferenceID: "client-ref-123",
					Plan: recurringContractModel.Plan{
						PlanId:   "plan-1",
						PlanName: "Premium Plan",
					},
					Trials: []recurringContractModel.Trial{
						{
							TrialStart: 1,
							TrialEnd:   3,
							Type:       "PERCENTAGE",
							Percentage: 50,
						},
					},
					Billing: recurringContractModel.Billing{
						Interval:     1,
						IntervalUnit: "MONTH",
						Count:        12,
					},
					Amount: commonModel.Amount{
						Value:    "100000",
						Currency: "IDR",
					},
					Status:    constant.RecurringContractStatusActive,
					CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
				}

				m.On(
					"GetRecurringByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.MatchedBy(func(req recurringContractModel.GetRecurringByIDRequest) bool {
						return req.RecurringID == validRecurringID && req.MerchantID == merchantID
					}),
				).Once().Return(mockResponse, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"uuid":"e39d1d21-0a7b-49e5-8cd9-404ac75d54be","merchantId":"merchant-uuid-123","customerId":"customer-uuid-123","clientReferenceId":"client-ref-123","plan":{"planId":"plan-1","planName":"Premium Plan"},"trials":[{"trialStart":1,"trialEnd":3,"type":"PERCENTAGE","percentage":50}],"billing":{"interval":1,"intervalUnit":"MONTH","count":12},"amount":{"currency":"IDR","value":"100000"},"status":"ACTIVE","updatedAt":"2024-01-02T00:00:00Z","createdAt":"2024-01-01T00:00:00Z"}}`,
		},
	}

	for _, test := range tests {
		pt.Run(test.name, func(t *testing.T) {
			mockRecurringService := mockServices.NewIRecurringContractService(pt)
			test.mockSetup(mockRecurringService)

			controller := &RecurringContractController{
				recurringContractService: mockRecurringService,
			}

			req := httptest.NewRequest(http.MethodGet, "/api/v1/recurrings/"+test.recurringID, nil)

			// Add chi route context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("uuid", test.recurringID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Add user claims to context if provided
			if test.userClaims != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, test.userClaims))
			}

			rec := httptest.NewRecorder()

			handler := http.HandlerFunc(controller.GetRecurringByID)
			handler.ServeHTTP(rec, req)

			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
