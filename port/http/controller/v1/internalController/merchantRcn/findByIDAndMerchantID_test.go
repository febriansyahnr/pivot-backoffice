package merchantRcn

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchantRcn"
	mockMerchant "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMerchantRcnController_FindByIDAndMerchantID(t *testing.T) {

	expectedMerchant := merchantRcn.MerchantRcnResponse{
		PartnerReferenceNo: "xxx",
	}

	testCase := []struct {
		name           string
		rcnID          string
		merchantID     string
		mockSetup      func(merchantSvcMocks *mockMerchant.IMerchantRcnService)
		expectedStatus int
	}{
		{
			name:       "SUCCESS",
			rcnID:      "550e8400-e29b-41d4-a716-446655440000",
			merchantID: "550e8400-e29b-41d4-a716-446655440001",
			mockSetup: func(merchantSvcMocks *mockMerchant.IMerchantRcnService) {
				merchantSvcMocks.On("FindByIDAndMerchantID",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(&expectedMerchant, nil)
			},
			expectedStatus: 200,
		},
		{
			name:       "ERROR: Service Error",
			rcnID:      "550e8400-e29b-41d4-a716-446655440000",
			merchantID: "550e8400-e29b-41d4-a716-446655440001",
			mockSetup: func(merchantSvcMocks *mockMerchant.IMerchantRcnService) {
				merchantSvcMocks.On("FindByIDAndMerchantID",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:  "ERROR: Merchant not found",
			rcnID: "",
			mockSetup: func(merchantSvcMocks *mockMerchant.IMerchantRcnService) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:       "ERROR: Merchant not found",
			rcnID:      "550e8400-e29b-41d4-a716-446655440000",
			merchantID: "550e8400-e29b-41d4-a716-446655440001",
			mockSetup: func(merchantSvcMocks *mockMerchant.IMerchantRcnService) {
				merchantSvcMocks.On("FindByIDAndMerchantID",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil, nil)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			mockMerchantSvc := mockMerchant.NewIMerchantRcnService(t)
			mockValidator := validator.New()

			tt.mockSetup(mockMerchantSvc)

			mc := New(mockMerchantSvc, mockValidator)

			// Create the HTTP request for the test case
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/merchants/rcns/%s", tt.rcnID), nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("rcn_id", tt.rcnID)

			rr := httptest.NewRecorder()

			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			req.Header.Set(constant.HeaderXMerchantId, tt.merchantID)
			// Create the handler and serve the request
			handler := http.HandlerFunc(mc.FindByIDAndMerchantID)
			handler.ServeHTTP(rr, req)

			if rr.Body.String() != "" {
				t.Logf("Handler response body: %s", rr.Body.String())
			}

			// Assertions
			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockMerchantSvc.AssertExpectations(t)
		})
	}
}
