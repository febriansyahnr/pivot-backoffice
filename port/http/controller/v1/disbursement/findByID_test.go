package disbursementController_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	disbursementController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/disbursement"
)

func TestFindByID(t *testing.T) {
	validMerchantID := uuid.NewString()
	validUserClaims := &user.UserTokenClaims{
		MerchantId: validMerchantID,
	}

	testCases := []struct {
		name           string
		mockSetup      func(mockService *serviceMocks.IDisbursementService)
		expectedStatus int
		disbursementID string
		userClaims     *user.UserTokenClaims
	}{
		{
			name: "SUCCESS: Find by ID",
			mockSetup: func(mockService *serviceMocks.IDisbursementService) {
				mockService.On(
					"FindByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(&disbursementModel.DisbursementWithTransaction{Disbursement: disbursementModel.Disbursement{
					MerchantID: validMerchantID,
				}}, nil)
			},
			expectedStatus: http.StatusOK,
			disbursementID: uuid.NewString(),
			userClaims:     validUserClaims,
		},
		{
			name: "ERROR: Invalid merchant ID",
			mockSetup: func(mockService *serviceMocks.IDisbursementService) {
				mockService.On(
					"FindByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(&disbursementModel.DisbursementWithTransaction{Disbursement: disbursementModel.Disbursement{
					MerchantID: uuid.NewString(),
				}}, nil)
			},
			expectedStatus: http.StatusBadRequest,
			disbursementID: uuid.NewString(),
			userClaims:     validUserClaims,
		},
		{
			name: "ERROR: Service FindByID error",
			mockSetup: func(mockService *serviceMocks.IDisbursementService) {
				mockService.On(
					"FindByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			expectedStatus: http.StatusInternalServerError,
			disbursementID: uuid.NewString(),
			userClaims:     validUserClaims,
		},
		{
			name: "ERROR: Empty ID",
			mockSetup: func(mockService *serviceMocks.IDisbursementService) {
				// invalid disbursement ID
			},
			expectedStatus: http.StatusBadRequest,
			disbursementID: "invalid",
			userClaims:     validUserClaims,
		},
		{
			name: "ERROR: User not in Context",
			mockSetup: func(mockService *serviceMocks.IDisbursementService) {
				// Empty mock setup
			},
			disbursementID: uuid.NewString(),
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
			req := httptest.NewRequest(http.MethodGet, "/api/v1/disbursements/"+tc.disbursementID, nil)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))
			if tc.userClaims != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tc.userClaims))
			}

			httpRecorder := httptest.NewRecorder()

			router := chi.NewRouter()
			router.Get(
				"/api/v1/disbursements/{id}", disbursementController.New(cfg, nil, nil, disbursementController.Services{DisbursementSvc: mockService}, nil, nil).FindByID,
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

func TestFindByIDWithSubMerchantId(t *testing.T) {
	validMerchantID := uuid.NewString()
	validSubMerchantID := uuid.NewString()
	validUserClaims := &user.UserTokenClaims{
		MerchantId: validMerchantID,
	}

	testCases := []struct {
		name               string
		mockDisbursement   func(mockService *serviceMocks.IDisbursementService)
		mockMerchant       func(mockService *serviceMocks.IMerchantService)
		expectedStatus     int
		disbursementID     string
		subMerchantIDQuery string
	}{
		{
			name: "SUCCESS: Find by ID with valid subMerchantId query",
			mockDisbursement: func(mockService *serviceMocks.IDisbursementService) {
				mockService.On(
					"FindByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(&disbursementModel.DisbursementWithTransaction{Disbursement: disbursementModel.Disbursement{
					MerchantID: validSubMerchantID,
				}}, nil)
			},
			mockMerchant: func(mockService *serviceMocks.IMerchantService) {
				mockService.On(
					"ValidateSubMerchantParent",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					validMerchantID,
					validSubMerchantID,
				).Return(nil)
			},
			expectedStatus:     http.StatusOK,
			disbursementID:     uuid.NewString(),
			subMerchantIDQuery: validSubMerchantID,
		},
		{
			name: "ERROR: ValidateSubMerchantParent returns error",
			mockDisbursement: func(mockService *serviceMocks.IDisbursementService) {
				// Not called when validation fails
			},
			mockMerchant: func(mockService *serviceMocks.IMerchantService) {
				mockService.On(
					"ValidateSubMerchantParent",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					validMerchantID,
					validSubMerchantID,
				).Return(pkgErrors.New(response.HttpErrForbidden, constant.ErrIncorrectSubMerchantParent))
			},
			expectedStatus:     http.StatusForbidden,
			disbursementID:     uuid.NewString(),
			subMerchantIDQuery: validSubMerchantID,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDisbursementSvc := serviceMocks.NewIDisbursementService(t)
			mockMerchantSvc := serviceMocks.NewIMerchantService(t)
			tc.mockDisbursement(mockDisbursementSvc)
			tc.mockMerchant(mockMerchantSvc)

			cfg := &config.Config{}
			chiRouterCtx := chi.NewRouteContext()

			url := "/api/v1/disbursements/" + tc.disbursementID + "?subMerchantId=" + tc.subMerchantIDQuery
			req := httptest.NewRequest(http.MethodGet, url, nil)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))
			req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, validUserClaims))

			httpRecorder := httptest.NewRecorder()

			router := chi.NewRouter()
			router.Get(
				"/api/v1/disbursements/{id}", disbursementController.New(cfg, nil, nil, disbursementController.Services{
					DisbursementSvc: mockDisbursementSvc,
					MerchantSvc:     mockMerchantSvc,
				}, nil, nil).FindByID,
			)
			router.ServeHTTP(httpRecorder, req)

			if httpRecorder.Body.String() != "" {
				t.Logf("response: %s", httpRecorder.Body.String())
			}

			assert.Equal(t, tc.expectedStatus, httpRecorder.Code)
			mockDisbursementSvc.AssertExpectations(t)
			mockMerchantSvc.AssertExpectations(t)
		})
	}
}
