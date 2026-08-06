package merchantTopUp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/merchantTopUp"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTopUp(t *testing.T) {

	expectedReference := &model.MerchantTopUp{
		ID:              "uuid-uuid-uuid",
		MerchantID:      "merchant-id",
		PaymentMethodID: "payment-method-id",
		ReferenceNumber: "reference-number",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	payloadRequest := &model.MerchantTopUpRequest{
		AccountName:     constant.TypeWallet,
		PaymentMethodID: "payment-method-id",
	}
	payloadRequestByte, err := json.Marshal(payloadRequest)
	require.NoError(t, err)

	userClaims := &user.UserTokenClaims{}

	testCases := []struct {
		name           string
		requestBody    []byte
		userClaims     *user.UserTokenClaims
		mockSetup      func(*mocks.IMerchantTopUpService)
		expectedStatus int
	}{
		{
			name:           "ERROR:User not found",
			mockSetup:      func(*mocks.IMerchantTopUpService) { /* Empty */ },
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "ERROR:Invalid JSON",
			userClaims:     userClaims,
			requestBody:    []byte("{invalid JSON"),
			mockSetup:      func(*mocks.IMerchantTopUpService) { /* Empty */ },
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "ERROR:Failed Validation",
			userClaims:     userClaims,
			requestBody:    []byte(`{}`),
			mockSetup:      func(*mocks.IMerchantTopUpService) { /* Empty */ },
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "ERROR:Failed to find or create reference",
			userClaims:  userClaims,
			requestBody: payloadRequestByte,
			mockSetup: func(topUpSvc *mocks.IMerchantTopUpService) {
				topUpSvc.On(
					"FindOrCreate", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:        "SUCCESS",
			userClaims:  userClaims,
			requestBody: payloadRequestByte,
			mockSetup: func(topUpSvc *mocks.IMerchantTopUpService) {
				topUpSvc.On(
					"FindOrCreate", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(expectedReference, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {

			topUpSvc := mocks.NewIMerchantTopUpService(t)

			tt.mockSetup(topUpSvc)

			mc := New(validatorExt.New(), topUpSvc)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/disbursements/top-up", bytes.NewBuffer(tt.requestBody))

			ctx := context.WithValue(req.Context(), constant.CtxAccountName, constant.TypeWallet)
			if tt.userClaims != nil {
				ctx = context.WithValue(ctx, constant.CtxUserInfoKey, tt.userClaims)
			}
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			// Create the handler and serve the request
			handler := http.HandlerFunc(mc.Topup)
			handler.ServeHTTP(rr, req)
			if !assert.Equal(t, tt.expectedStatus, rr.Code) {
				t.Log("Request Body :", string(tt.requestBody))
				t.Log("Response Body:", rr.Body.String())
			}

			topUpSvc.AssertExpectations(t)
		})
	}
}
