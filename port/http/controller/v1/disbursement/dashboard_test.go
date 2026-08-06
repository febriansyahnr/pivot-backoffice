package disbursementController

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementDashboardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursementDashboard"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestController_GetDisbursementDashboard(t *testing.T) {
	validUserClaims := &user.UserTokenClaims{
		UUID:       uuid.NewString(),
		MerchantId: uuid.NewString(),
	}

	startDate := util.GetCurrentDateOfLocation(util.GetJakartaTimeLocation()).UTC()
	endDate := startDate.AddDate(0, 0, 1)

	testCases := []struct {
		name           string
		params         string
		mockSetup      func(disbursementDashboardSvc *mocks.IDisbursementDashboardService)
		expectedStatus int
		userClaim      *user.UserTokenClaims
	}{
		{
			name: "SUCCESS",
			mockSetup: func(disbursementDashboardSvc *mocks.IDisbursementDashboardService) {
				disbursementDashboardSvc.
					On(
						"Get",
						mock.Anything,
						disbursementDashboardModel.GetDisbursementDashboardFilter{
							MerchantID:       validUserClaims.MerchantId,
							InsightStartDate: startDate,
							InsightEndDate:   endDate,
						}).
					Return(&disbursementDashboardModel.DisbursementDashboardResponse{}, nil)
			},
			expectedStatus: http.StatusOK,
			userClaim:      validUserClaims,
		},
		{
			name:           "ERROR: User not in Context",
			mockSetup:      func(disbursementDashboardSvc *mocks.IDisbursementDashboardService) {},
			expectedStatus: http.StatusUnauthorized,
			userClaim:      nil,
		},
		{
			name: "ERROR: Service error",
			mockSetup: func(disbursementDashboardSvc *mocks.IDisbursementDashboardService) {
				disbursementDashboardSvc.
					On(
						"Get",
						mock.Anything,
						disbursementDashboardModel.GetDisbursementDashboardFilter{
							MerchantID:       validUserClaims.MerchantId,
							InsightStartDate: startDate,
							InsightEndDate:   endDate,
						}).
					Return(nil, errors.New("some error"))
			},
			expectedStatus: http.StatusInternalServerError,
			userClaim:      validUserClaims,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				ServiceName: "testing",
			}

			mockValidator := validator.New()
			mockDisbursementDashboardSvc := mocks.NewIDisbursementDashboardService(t)

			tt.mockSetup(mockDisbursementDashboardSvc)

			mc := New(cfg, mockValidator, nil, Services{DisbursementDashboardSvc: mockDisbursementDashboardSvc}, nil, nil)

			// Create the HTTP request for the test case
			path := fmt.Sprintf("/api/v1/disbursements/dashboard%s", tt.params)
			req := httptest.NewRequest(http.MethodGet, path, nil)
			if tt.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tt.userClaim))
			}
			rr := httptest.NewRecorder()

			// Create the handler and serve the request
			handler := http.HandlerFunc(mc.GetDisbursementDashboard)
			handler.ServeHTTP(rr, req)

			if rr.Body.String() != "" {
				t.Logf("Handler response body: %s", rr.Body.String())
			}

			// Assertions
			assert.Equal(t, tt.expectedStatus, rr.Code)

			mockDisbursementDashboardSvc.AssertExpectations(t)
		})
	}
}
