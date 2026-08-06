package internal_merchant

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUnblockMerchantController(t *testing.T) {
	merchantID := uuid.NewString()
	token := &merchantModel.MerchantAuthTokenClaims{
		MerchantId: merchantID,
	}
	testcases := []struct {
		Name           string
		Request        []byte
		MockSetup      func(svc *mockSvc.IMerchantForbiddenUseCaseService)
		token          *merchantModel.MerchantAuthTokenClaims
		WantErr        bool
		ExpectedStatus int
	}{
		{
			Name: "SUCCESS",
			Request: []byte(`{
				"merchantId": "` + merchantID + `",
				"usecase":"DISBURSEMENT"
			}`),
			MockSetup: func(svc *mockSvc.IMerchantForbiddenUseCaseService) {
				svc.On("BlockUseCase", mock.Anything, mock.Anything).Return(nil)
			},
			token:          token,
			WantErr:        false,
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:    "ERROR: Bad Request",
			Request: []byte("{invalid JSON"),
			MockSetup: func(svc *mockSvc.IMerchantForbiddenUseCaseService) {
			},
			token:          token,
			WantErr:        true,
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name: "ERROR: Failed Validation",
			Request: []byte(`{
				"merchantId": "123",
				"usecase":"123"
			}`),
			MockSetup: func(svc *mockSvc.IMerchantForbiddenUseCaseService) {
			},
			token:          token,
			WantErr:        true,
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name: "ERROR: Error block usecase",
			Request: []byte(`{
				"merchantId": "` + merchantID + `",
				"usecase":"DISBURSEMENT"
			}`),
			MockSetup: func(svc *mockSvc.IMerchantForbiddenUseCaseService) {
				svc.On("BlockUseCase", mock.Anything, mock.Anything).Return(errors.New("error"))
			},
			token:          token,
			WantErr:        true,
			ExpectedStatus: http.StatusInternalServerError,
		},
		{
			Name: "ERROR: Unauthorized",
			Request: []byte(`{
				"merchantId": "` + merchantID + `",
				"usecase":"DISBURSEMENT"
			}`),
			MockSetup: func(_ *mockSvc.IMerchantForbiddenUseCaseService) {
			},
			token:          nil,
			WantErr:        true,
			ExpectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.Name, func(t *testing.T) {
			mockForbiddenSvc := mockSvc.NewIMerchantForbiddenUseCaseService(t)
			mockValidator := validator.New()
			tt.MockSetup(mockForbiddenSvc)

			controller := New(mockForbiddenSvc, nil, mockValidator)
			req := httptest.NewRequest(http.MethodPost, "/merchants/forbidden-usecase/block", bytes.NewBuffer(tt.Request))
			if tt.token != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxMerchantInfo, tt.token))
			}
			rr := httptest.NewRecorder()

			// Create the handler and serve the request
			handler := http.HandlerFunc(controller.Block)
			handler.ServeHTTP(rr, req)

			if rr.Body.String() != "" {
				t.Logf("Handler response body: %s", rr.Body.String())
			}

			// Assertions
			assert.Equal(t, tt.ExpectedStatus, rr.Code)
			mockForbiddenSvc.AssertExpectations(t)

		})
	}
}
