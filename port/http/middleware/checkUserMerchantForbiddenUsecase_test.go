package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	chi "github.com/go-chi/chi/v5"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/port/http/middleware"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCheckUserMerchantForbiddenUsecase(t *testing.T) {
	userClaim := user.UserTokenClaims{}

	testCases := []struct {
		Name         string
		Usecase      string
		UserClaim    *user.UserTokenClaims
		MockSetup    func(svc *mocks.IMerchantForbiddenUseCaseService)
		HttpMethod   string
		ExpectedCode int
	}{
		{
			Name:         "SUCCESS: Get Request",
			Usecase:      "DISBURSEMENT",
			UserClaim:    &userClaim,
			ExpectedCode: http.StatusOK,
			HttpMethod:   http.MethodGet,
			MockSetup: func(svc *mocks.IMerchantForbiddenUseCaseService) {
			},
		},
		{
			Name:         "SUCCESS",
			Usecase:      "DISBURSEMENT",
			UserClaim:    &userClaim,
			ExpectedCode: http.StatusOK,
			HttpMethod:   http.MethodPost,
			MockSetup: func(svc *mocks.IMerchantForbiddenUseCaseService) {
				svc.On("CheckUseCase", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name:         "ERROR: No User Claim",
			Usecase:      "DISBURSEMENT",
			UserClaim:    nil,
			ExpectedCode: http.StatusUnauthorized,
			HttpMethod:   http.MethodPost,
			MockSetup: func(svc *mocks.IMerchantForbiddenUseCaseService) {
			},
		},
		{
			Name:         "ERROR: Error check usecase",
			Usecase:      "DISBURSEMENT",
			UserClaim:    &userClaim,
			ExpectedCode: http.StatusInternalServerError,
			HttpMethod:   http.MethodPost,
			MockSetup: func(svc *mocks.IMerchantForbiddenUseCaseService) {
				svc.On("CheckUseCase", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("error"))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			forbiddenSvc := mocks.NewIMerchantForbiddenUseCaseService(t)
			tc.MockSetup(forbiddenSvc)
			middleware := middleware.CheckUserMerchantForbiddenUsecase(forbiddenSvc, tc.Usecase)

			router := chi.NewRouter()
			MountHandlers(router, middleware)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tc.HttpMethod, "/test", nil)

			if tc.UserClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tc.UserClaim))
			}

			router.ServeHTTP(rec, req)
			require.Equal(t, tc.ExpectedCode, rec.Result().StatusCode)
		})
	}
}
