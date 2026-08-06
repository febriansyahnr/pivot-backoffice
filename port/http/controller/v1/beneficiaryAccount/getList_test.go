package beneficiaryAccountController

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	beneficiaryAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/beneficiaryAccount"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockRabbitMq "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetList(t *testing.T) {
	data := make([]beneficiaryAccountModel.BeneficiaryAccount, 0)
	data = append(data, beneficiaryAccountModel.BeneficiaryAccount{
		UUID: uuid.NewString(),
	})
	expectedResponse := &commonModel.PaginationResponse{
		Data: data,
		Meta: commonModel.Meta{
			Page:       1,
			PerPage:    20,
			TotalItems: 1,
			TotalPages: 1,
		},
	}
	validUserClaims := &user.UserTokenClaims{
		MerchantId: uuid.NewString(),
	}

	testCases := []struct {
		name            string
		mockSetup       func(mockService *serviceMocks.IBeneficiaryAccountService)
		expectedStatus  int
		funcQueryParams func() *url.Values
		userClaims      *user.UserTokenClaims
	}{
		{
			name: "SUCCESS: Get List without any filter",
			mockSetup: func(mockService *serviceMocks.IBeneficiaryAccountService) {
				mockService.On(
					"GetList",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*beneficiaryAccountModel.GetBeneficiaryAccountFilterRequest"),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(expectedResponse, nil)
			},
			expectedStatus: http.StatusOK,
			funcQueryParams: func() *url.Values {
				return nil
			},
			userClaims: validUserClaims,
		},
		{
			name: "SUCCESS: Get List with created_at filter",
			mockSetup: func(mockService *serviceMocks.IBeneficiaryAccountService) {
				mockService.On(
					"GetList",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*beneficiaryAccountModel.GetBeneficiaryAccountFilterRequest"),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(expectedResponse, nil)
			},
			expectedStatus: http.StatusOK,
			funcQueryParams: func() *url.Values {
				queryParams := url.Values{}
				queryParams.Add("startCreatedAt", util.TimeNow.Add(-24*time.Hour).Format(util.UTCLayout))
				queryParams.Add("endCreatedAt", util.TimeNow.Format(util.UTCLayout))
				return &queryParams
			},
			userClaims: validUserClaims,
		},
		{
			name: "SUCCESS: Get List that has page values",
			mockSetup: func(mockService *serviceMocks.IBeneficiaryAccountService) {
				mockService.On(
					"GetList",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*beneficiaryAccountModel.GetBeneficiaryAccountFilterRequest"),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(expectedResponse, nil)
			},
			expectedStatus: http.StatusOK,
			funcQueryParams: func() *url.Values {
				queryParams := url.Values{}
				queryParams.Add("page", "2")
				return &queryParams
			},
			userClaims: validUserClaims,
		},
		{
			name: "FAILED: Got error 500 on get list caused by service error",
			mockSetup: func(mockService *serviceMocks.IBeneficiaryAccountService) {
				mockService.On(
					"GetList",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*beneficiaryAccountModel.GetBeneficiaryAccountFilterRequest"),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(nil, errors.New("some-error"))
			},
			expectedStatus: http.StatusInternalServerError,
			funcQueryParams: func() *url.Values {
				return nil
			},
			userClaims: validUserClaims,
		},
		{
			name:           "FAILED: Got error 400 on get list caused by invalid startCreatedAt",
			mockSetup:      func(mockService *serviceMocks.IBeneficiaryAccountService) {},
			expectedStatus: http.StatusBadRequest,
			funcQueryParams: func() *url.Values {
				queryParams := url.Values{}
				queryParams.Add("startCreatedAt", "invalid format")
				return &queryParams
			},
			userClaims: validUserClaims,
		},
		{
			name:           "FAILED: Got error 400 on get list caused by invalid endCreatedAt",
			mockSetup:      func(mockService *serviceMocks.IBeneficiaryAccountService) {},
			expectedStatus: http.StatusBadRequest,
			funcQueryParams: func() *url.Values {
				queryParams := url.Values{}
				queryParams.Add("endCreatedAt", "invalid format")
				return &queryParams
			},
			userClaims: validUserClaims,
		},
		{
			name:           "FAILED: Got error 400 on get list caused by invalid page",
			mockSetup:      func(mockService *serviceMocks.IBeneficiaryAccountService) {},
			expectedStatus: http.StatusBadRequest,
			funcQueryParams: func() *url.Values {
				queryParams := url.Values{}
				queryParams.Add("page", "invalid format")
				return &queryParams
			},
			userClaims: validUserClaims,
		},
		{
			name: "ERROR: User not in Context",
			mockSetup: func(mockService *serviceMocks.IBeneficiaryAccountService) {
			},
			userClaims: nil,
			funcQueryParams: func() *url.Values {
				return nil
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := serviceMocks.NewIBeneficiaryAccountService(t)
			mockDis := serviceMocks.NewIDisbursementService(t)
			validate := validator.New()
			mockRmq := mockRabbitMq.NewRabbitMQExt(t)

			tc.mockSetup(mockService)

			cfg := &config.Config{}
			cfg.AppConfig.PaginationPerPage = 20
			beneficiaryAccountController := New(cfg, validate, mockRmq, mockService, mockDis, nil, nil)
			baseUrl := "/api/v1/beneficiary-accounts"

			// Append query parameters to the URL
			if tc.funcQueryParams() != nil {
				baseUrl += "?" + tc.funcQueryParams().Encode()
			}

			chiRouterCtx := chi.NewRouteContext()
			req := httptest.NewRequest(http.MethodGet, baseUrl, nil)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))
			if tc.userClaims != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tc.userClaims))
			}

			httpRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(beneficiaryAccountController.GetList)
			handler.ServeHTTP(httpRecorder, req)

			if httpRecorder.Body.String() != "" {
				t.Logf("response: %s", httpRecorder.Body.String())
			}

			assert.Equal(t, tc.expectedStatus, httpRecorder.Code)
			mockService.AssertExpectations(t)
		})
	}
}
