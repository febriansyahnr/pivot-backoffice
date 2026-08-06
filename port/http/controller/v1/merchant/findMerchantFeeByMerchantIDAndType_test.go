package merchant

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockRabbitMq "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
)

func TestMerchantControllerFindMerchantFeeByMerchantIDAndType(t *testing.T) {
	merchantID := uuid.NewString()

	validUserClaims := &user.UserTokenClaims{
		UUID:       uuid.NewString(),
		MerchantId: merchantID,
	}

	testCase := []struct {
		name           string
		merchantId     string
		feeType        string
		mockSetup      func(feeSvc *serviceMocks.IFeeService)
		userClaim      *user.UserTokenClaims
		expectedStatus int
	}{
		{
			name:       "SUCCESS",
			merchantId: merchantID,
			feeType:    constant.TypeDisbursement,
			mockSetup: func(feeSvc *serviceMocks.IFeeService) {
				feeSvc.On("GetFeeCalculationAndDetail", constant.ValueCtxMockType(), constant.PtrGetFeeRequestMockType()).
					Return(1_000.0, &feeModel.FeeMetadataObject{}, nil)
			},
			userClaim:      validUserClaims,
			expectedStatus: 200,
		},
		{
			name:       "ERROR: Service Error",
			merchantId: merchantID,
			feeType:    constant.TypeDisbursement,
			mockSetup: func(feeSvc *serviceMocks.IFeeService) {
				feeSvc.On("GetFeeCalculationAndDetail", constant.ValueCtxMockType(), constant.PtrGetFeeRequestMockType()).
					Return(1_000.0, &feeModel.FeeMetadataObject{}, constant.ErrSomeErrorForUnitTest)
			},
			userClaim:      validUserClaims,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:       "ERROR: User not found",
			merchantId: merchantID,
			feeType:    constant.TypeDisbursement,
			mockSetup: func(feeSvc *serviceMocks.IFeeService) {
				// Empty mock setup
			},
			userClaim:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			feeSvc := serviceMocks.NewIFeeService(t)
			mockValidator := validator.New()
			mockRmq := mockRabbitMq.NewRabbitMQExt(t)
			tt.mockSetup(feeSvc)

			mc := New(nil, mockValidator, mockRmq, WithFeeService(feeSvc))

			// Create the HTTP request for the test case
			chiRouterCtx := chi.NewRouteContext()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/merchants/fee", nil)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))
			if tt.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tt.userClaim))
			}

			chiRouterCtx.URLParams.Add("type", tt.feeType)

			// Create the handler and serve the request
			httpRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(mc.FindMerchantFeeByMerchantIDAndType)
			handler.ServeHTTP(httpRecorder, req)

			if httpRecorder.Body.String() != "" {
				t.Logf("Handler response body: %s", httpRecorder.Body.String())
			}

			// Assertions
			assert.Equal(t, tt.expectedStatus, httpRecorder.Code)
			feeSvc.AssertExpectations(t)
		})
	}
}
