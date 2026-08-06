package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	chi "github.com/go-chi/chi/v5"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/port/http/middleware"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCheckMerchantForbiddenUsecase(t *testing.T) {
	merchantClaim := merchant.MerchantAuthTokenClaims{}
	subMerchantId := "submerchantId"
	useCase := "DISBURSEMENT"

	testCases := []struct {
		Name                     string
		Usecase                  string
		MerchantClaim            *merchant.MerchantAuthTokenClaims
		MockSetup                func(svc *mocks.IMerchantForbiddenUseCaseService)
		HttpMethod               string
		ExpectedCode             int
		IsActOnBehalfSubmerchant bool
	}{
		{
			Name:          "SUCCESS: Get Request",
			Usecase:       useCase,
			MerchantClaim: &merchantClaim,
			ExpectedCode:  http.StatusOK,
			HttpMethod:    http.MethodGet,
			MockSetup: func(svc *mocks.IMerchantForbiddenUseCaseService) {
			},
		},
		{
			Name:          "SUCCESS",
			Usecase:       useCase,
			MerchantClaim: &merchantClaim,
			ExpectedCode:  http.StatusOK,
			HttpMethod:    http.MethodPost,
			MockSetup: func(svc *mocks.IMerchantForbiddenUseCaseService) {
				svc.On("CheckUseCase", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name:          "ERROR: No Merchant Claim",
			Usecase:       useCase,
			MerchantClaim: nil,
			ExpectedCode:  http.StatusUnauthorized,
			HttpMethod:    http.MethodPost,
			MockSetup: func(svc *mocks.IMerchantForbiddenUseCaseService) {
			},
		},
		{
			Name:          "ERROR: Error check usecase",
			Usecase:       useCase,
			MerchantClaim: &merchantClaim,
			ExpectedCode:  http.StatusInternalServerError,
			HttpMethod:    http.MethodPost,
			MockSetup: func(svc *mocks.IMerchantForbiddenUseCaseService) {
				svc.On("CheckUseCase", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("error"))
			},
		},
		{
			Name:          "ERROR: Error check usecase",
			Usecase:       useCase,
			MerchantClaim: &merchantClaim,
			ExpectedCode:  http.StatusInternalServerError,
			HttpMethod:    http.MethodPost,
			MockSetup: func(svc *mocks.IMerchantForbiddenUseCaseService) {
				svc.On("CheckUseCase", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("error"))
			},
		},
		{
			Name:          "SUCCESS: Merchant act on behalf sub-merchant where merchant is blocked and submerchant is not",
			Usecase:       useCase,
			MerchantClaim: &merchantClaim,
			ExpectedCode:  http.StatusOK,
			HttpMethod:    http.MethodPost,
			MockSetup: func(svc *mocks.IMerchantForbiddenUseCaseService) {
				svc.On("CheckUseCase", mock.Anything, merchantClaim.MerchantId, useCase).Return(errors.New("error")).Maybe()
				svc.On("CheckUseCase", mock.Anything, subMerchantId, useCase).Return(nil).Maybe()
			},
			IsActOnBehalfSubmerchant: true,
		},
		{
			Name:          "ERROR: Merchant act on behalf sub-merchant where merchant is not blocked and submerchant is",
			Usecase:       useCase,
			MerchantClaim: &merchantClaim,
			ExpectedCode:  http.StatusInternalServerError,
			HttpMethod:    http.MethodPost,
			MockSetup: func(svc *mocks.IMerchantForbiddenUseCaseService) {
				svc.On("CheckUseCase", mock.Anything, merchantClaim.MerchantId, useCase).Return(nil).Maybe()
				svc.On("CheckUseCase", mock.Anything, subMerchantId, useCase).Return(errors.New("error")).Maybe()
			},
			IsActOnBehalfSubmerchant: true,
		},
		{
			Name:          "ERROR: Merchant act on behalf sub-merchant where both merchant and submerchant are blocked",
			Usecase:       useCase,
			MerchantClaim: &merchantClaim,
			ExpectedCode:  http.StatusInternalServerError,
			HttpMethod:    http.MethodPost,
			MockSetup: func(svc *mocks.IMerchantForbiddenUseCaseService) {
				svc.On("CheckUseCase", mock.Anything, merchantClaim.MerchantId, useCase).Return(errors.New("error")).Maybe()
				svc.On("CheckUseCase", mock.Anything, subMerchantId, useCase).Return(errors.New("error")).Maybe()
			},
			IsActOnBehalfSubmerchant: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			forbiddenSvc := mocks.NewIMerchantForbiddenUseCaseService(t)
			tc.MockSetup(forbiddenSvc)
			middleware := middleware.CheckMerchantForbiddenUsecase(forbiddenSvc, tc.Usecase)

			router := chi.NewRouter()
			MountHandlers(router, middleware)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tc.HttpMethod, "/test", nil)

			if tc.IsActOnBehalfSubmerchant {
				req.Header.Add(constant.HeaderXSubMerchantID, subMerchantId)
			}
			if tc.MerchantClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxMerchantInfo, tc.MerchantClaim))
			}

			router.ServeHTTP(rec, req)
			require.Equal(t, tc.ExpectedCode, rec.Result().StatusCode)
		})
	}
}
