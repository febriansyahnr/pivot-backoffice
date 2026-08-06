package v1InternalUnifiedPaymentController

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	merchant "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	mockService "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgError "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFindPaymentByReferenceId(t *testing.T) {
	claim := merchant.MerchantAuthTokenClaims{
		MerchantId: uuid.NewString(),
	}
	paymentReferenceID := uuid.NewString()
	expiredAt, _ := time.Parse(time.RFC3339, "9999-12-31T23:59:59+07:00")

	testCases := []struct {
		Name             string
		MockSetup        func(svc *mockService.IPaymentService)
		SetUrlParam      bool
		Claim            *merchant.MerchantAuthTokenClaims
		WantErr          bool
		ExpectedCode     int
		ExpectedResponse string
	}{
		{
			Name: "SUCCESS",
			MockSetup: func(svc *mockService.IPaymentService) {
				svc.On("GetPaymentByReferenceId", mock.Anything, paymentReferenceID, claim.MerchantId).
					Return(&paymentModel.UnifiedPaymentResponse{
						UUID:       paymentReferenceID,
						MerchantID: claim.MerchantId,
						Amount:     commonModel.Amount{Currency: "IDR", Value: "100000.00"},
						Status:     "success",
						ExpiryAt:   &expiredAt,
					}, nil).Once()
			},
			SetUrlParam:      true,
			Claim:            &claim,
			WantErr:          false,
			ExpectedCode:     http.StatusOK,
			ExpectedResponse: `{"code":"00","message":"Success","data":{"uuid":"` + paymentReferenceID + `","merchantId":"` + claim.MerchantId + `","amount":{"currency":"IDR","value":"100000.00"},"status":"success","expiryAt":"9999-12-31T23:59:59+07:00","paidAmount":{"currency":"","value":""},"paymentTypeDetail":{},"referenceId":""}}`,
		},
		{
			Name:             "ERROR: Missing Merchant Claim",
			MockSetup:        func(svc *mockService.IPaymentService) {},
			Claim:            nil,
			WantErr:          true,
			ExpectedCode:     http.StatusUnauthorized,
			ExpectedResponse: `{"code":"401","message":"Unauthorized","error":{"type":"API_ERROR","message":"Unauthorized access","recommendation":""},"data":null}`,
		},
		{
			Name:             "ERROR: Invalid ReferenceID",
			MockSetup:        func(svc *mockService.IPaymentService) {},
			SetUrlParam:      false,
			Claim:            &claim,
			WantErr:          true,
			ExpectedCode:     http.StatusBadRequest,
			ExpectedResponse: `{"code":"400","message":"invalid request payload","error":{"type":"API_ERROR","message":"invalid request payload","recommendation":""},"data":null}`,
		},
		{
			Name: "ERROR: Internal Server Error",
			MockSetup: func(svc *mockService.IPaymentService) {
				svc.On("GetPaymentByReferenceId", mock.Anything, paymentReferenceID, claim.MerchantId).
					Return(nil, pkgError.New("internal server error", errors.New("internal service error"))).Once()
			},
			SetUrlParam:      true,
			Claim:            &claim,
			WantErr:          true,
			ExpectedCode:     http.StatusInternalServerError,
			ExpectedResponse: `{"code":"500","message":"internal server error","error":{"type":"SERVER_ERROR","message":"internal service error","recommendation":""},"data":null}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mockPaymentService := new(mockService.IPaymentService)
			tc.MockSetup(mockPaymentService)

			controller := paymentController{
				paymentSvc: mockPaymentService,
			}

			req := httptest.NewRequest(http.MethodGet, "/api/v1/unified-payments/"+paymentReferenceID, nil)
			if tc.SetUrlParam {
				routeCtx := chi.NewRouteContext()
				routeCtx.URLParams.Add("referenceId", paymentReferenceID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
			}

			if tc.Claim != nil {
				ctx := context.WithValue(req.Context(), constant.CtxMerchantInfo, tc.Claim)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(controller.FindPaymentByReferenceId)
			handler.ServeHTTP(rr, req)

			if rr.Body.String() != "" {
				t.Logf("Handler response body: %s", rr.Body.String())
			}

			assert.Equal(t, tc.ExpectedCode, rr.Code)
			if !tc.WantErr {
				assert.JSONEq(t, tc.ExpectedResponse, rr.Body.String())
			}
		})
	}
}
