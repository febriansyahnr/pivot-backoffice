package subMerchant

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockServices "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/stretchr/testify/assert"
)

func TestGetSubMerchantDailyLimit(t *testing.T) {
	subMerchantID := uuid.NewString()
	mockDisbursementSvc := mockServices.NewIDisbursementService(t)

	controller := &SubMerchantController{
		validate:        validator.New(),
		disbursementSvc: mockDisbursementSvc,
	}

	router := chi.NewRouter()
	router.Get("/sub-merchants/{id}/daily-limits/{type}", controller.GetSubMerchantDailyLimit)

	tests := []struct {
		name           string
		subMerchantID  string
		merchantType   string
		userClaims     *user.UserTokenClaims
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:User not found",
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   constant.WrapErrApiRespForTest(41, response.ErrTypeAPI, "user not found"),
		},
		{
			name:          "ERROR:Invalid sub merchant ID",
			subMerchantID: "invalid-sub-merchant-id",
			merchantType:  "merchant",
			userClaims: &user.UserTokenClaims{
				MerchantId: "valid-merchant-id",
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   constant.WrapErrApiRespForTest(40, response.ErrTypeAPI, "merchant id is not valid"),
		},
		{
			name:          "ERROR:Invalid merchant type",
			subMerchantID: subMerchantID,
			merchantType:  "invalid-merchant-type",
			userClaims: &user.UserTokenClaims{
				MerchantId: "valid-merchant-id",
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   constant.WrapErrApiRespForTest(40, response.ErrTypeAPI, "merchant type not registered"),
		},
		{
			name:          "ERROR:Some error",
			subMerchantID: subMerchantID,
			userClaims: &user.UserTokenClaims{
				MerchantId: "valid-merchant-id",
			},
			setupMock: func() {
				mockDisbursementSvc.On(
					"GetDailyTransactionLimit", constant.ValueCtxMockType(), subMerchantID, constant.StringMockType(),
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   constant.WrapErrApiRespForTest(99, response.ErrTypeUnknown, "some error"),
		},
		{
			name:          "SUCCESS:Sub-Merchant request",
			subMerchantID: subMerchantID,
			userClaims: &user.UserTokenClaims{
				MerchantId: "valid-merchant-id",
			},
			setupMock: func() {
				mockDisbursementSvc.On(
					"GetDailyTransactionLimit", constant.ValueCtxMockType(), subMerchantID, constant.StringMockType(),
				).Once().Return(nil, constant.ErrForbiddenAccess)
			},
			wantStatusCode: http.StatusNoContent,
		},
		{
			name:          "SUCCESS:when everything is fine",
			subMerchantID: subMerchantID,
			userClaims: &user.UserTokenClaims{
				MerchantId: "valid-merchant-id",
			},
			setupMock: func() {
				mockDisbursementSvc.On(
					"GetDailyTransactionLimit", constant.ValueCtxMockType(), subMerchantID, constant.StringMockType(),
				).Return(&disbursementModel.DailyTransactionLimitResponse{
					Limit:     util.ValueToPtr(10_000.00),
					Processed: 1_000,
					Remaining: 9_000,
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"limit":10000,"processed":1000,"remaining":9000}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.merchantType == "" {
				test.merchantType = "merchant"
			}
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/sub-merchants/"+test.subMerchantID+"/daily-limits/"+test.merchantType, nil)

			if test.setupMock != nil {
				test.setupMock()
			}
			if test.userClaims != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, test.userClaims))
			}

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if rec.Result().StatusCode == http.StatusNoContent {
				assert.Empty(t, rec.Body.String())

			} else {
				assert.JSONEq(t, test.wantRespBody, rec.Body.String())
			}
		})
	}

}
