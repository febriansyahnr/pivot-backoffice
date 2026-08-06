package disbursementController_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	disbursementController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/disbursement"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetBulkDisbursementDetail(t *testing.T) {
	validMerchantID := uuid.NewString()
	validUserClaims := &user.UserTokenClaims{
		MerchantId: validMerchantID,
	}

	testCases := []struct {
		name           string
		mockSetup      func(mockService *serviceMocks.IDisbursementService)
		expectedStatus int
		bulkID         string
		userClaims     *user.UserTokenClaims
	}{
		{
			name: "SUCCESS: Get bulk disbursement detail",
			mockSetup: func(mockService *serviceMocks.IDisbursementService) {
				mockService.On(
					"GetBulkDisbursementDetail",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(&disbursementModel.BulkDisbursementDetail{
					MerchantID: validMerchantID,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			bulkID:         uuid.NewString(),
			userClaims:     validUserClaims,
		},
		{
			name: "ERROR: Merchant mismatch",
			mockSetup: func(mockService *serviceMocks.IDisbursementService) {
				mockService.On(
					"GetBulkDisbursementDetail",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(&disbursementModel.BulkDisbursementDetail{
					MerchantID: uuid.NewString(),
				}, nil)
			},
			expectedStatus: http.StatusBadRequest,
			bulkID:         uuid.NewString(),
			userClaims:     validUserClaims,
		},
		{
			name: "ERROR: Service GetBulkDisbursementDetail error",
			mockSetup: func(mockService *serviceMocks.IDisbursementService) {
				mockService.On(
					"GetBulkDisbursementDetail",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			expectedStatus: http.StatusInternalServerError,
			bulkID:         uuid.NewString(),
			userClaims:     validUserClaims,
		},
		{
			name: "ERROR: Invalid ID",
			mockSetup: func(mockService *serviceMocks.IDisbursementService) {
				// invalid bulk ID, service not called
			},
			expectedStatus: http.StatusBadRequest,
			bulkID:         "invalid",
			userClaims:     validUserClaims,
		},
		{
			name: "ERROR: User not in Context",
			mockSetup: func(mockService *serviceMocks.IDisbursementService) {
				// Empty mock setup
			},
			bulkID:         uuid.NewString(),
			userClaims:     nil,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := serviceMocks.NewIDisbursementService(t)
			tc.mockSetup(mockService)

			cfg := &config.Config{}
			chiRouterCtx := chi.NewRouteContext()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/disbursements/bulk/"+tc.bulkID, nil)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))
			if tc.userClaims != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tc.userClaims))
			}

			httpRecorder := httptest.NewRecorder()

			router := chi.NewRouter()
			router.Get(
				"/api/v1/disbursements/bulk/{id}", disbursementController.New(cfg, nil, nil, disbursementController.Services{DisbursementSvc: mockService}, nil, nil).GetBulkDisbursementDetail,
			)
			router.ServeHTTP(httpRecorder, req)

			if httpRecorder.Body.String() != "" {
				t.Logf("response: %s", httpRecorder.Body.String())
			}

			assert.Equal(t, tc.expectedStatus, httpRecorder.Code)
			mockService.AssertExpectations(t)
		})
	}
}
