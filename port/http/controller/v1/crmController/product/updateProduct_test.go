package crmProductController

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/product"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateProductAvailability(t *testing.T) {
	testcases := []struct {
		Name             string
		SetupRequest     func(req *http.Request) *http.Request
		SetupBody        func() []byte
		MockSetup        func(svc *mockSvc.IProductService)
		WantErr          bool
		ExpectedStatus   int
		ExpectedResponse string
	}{
		{
			Name: "SUCCESS",
			SetupRequest: func(req *http.Request) *http.Request {
				ctx := req.Context()
				chiCtx := chi.NewRouteContext()
				chiCtx.URLParams.Add("merchantId", uuid.NewString())
				return req.WithContext(context.WithValue(ctx, chi.RouteCtxKey, chiCtx))
			},
			SetupBody: func() []byte {
				payload := product.UpdateProductRequest{
					ID:     uuid.Max.String(),
					Active: true,
				}
				b, _ := json.Marshal(payload)
				return b
			},
			MockSetup: func(svc *mockSvc.IProductService) {
				svc.On(
					"UpdateProductAvailability",
					mock.Anything,
					mock.Anything,
				).Return(
					nil,
				)
			},
			WantErr:          false,
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: `{"code":"00"}`,
		},
		{
			Name: "ERROR: Failed decode payload",
			SetupRequest: func(req *http.Request) *http.Request {
				ctx := req.Context()
				chiCtx := chi.NewRouteContext()
				chiCtx.URLParams.Add("merchantId", uuid.NewString())
				return req.WithContext(context.WithValue(ctx, chi.RouteCtxKey, chiCtx))
			},
			SetupBody: func() []byte {
				b, _ := json.Marshal("string")
				return b
			},
			MockSetup: func(svc *mockSvc.IProductService) {
			},
			WantErr:          true,
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedResponse: `{"code":"40","errors":"json: cannot unmarshal string into Go value of type product.UpdateProductRequest"}`,
		},
		{
			Name: "ERROR: Failed payload validation",
			SetupRequest: func(req *http.Request) *http.Request {
				ctx := req.Context()
				chiCtx := chi.NewRouteContext()
				chiCtx.URLParams.Add("merchantId", uuid.NewString())
				return req.WithContext(context.WithValue(ctx, chi.RouteCtxKey, chiCtx))
			},
			SetupBody: func() []byte {
				payload := product.UpdateMerchantSelectedProductAvailabilityRequest{
					ProductID: "",
					Active:    true,
				}
				b, _ := json.Marshal(payload)
				return b
			},
			MockSetup: func(svc *mockSvc.IProductService) {
			},
			WantErr:          true,
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedResponse: `{"code":"40","errors":{"ID":"Key: 'UpdateProductRequest.ID' Error:Field validation for 'ID' failed on the 'required' tag"}}`,
		},
		{
			Name: "ERROR: Error update product",
			SetupRequest: func(req *http.Request) *http.Request {
				ctx := req.Context()
				chiCtx := chi.NewRouteContext()
				chiCtx.URLParams.Add("merchantId", uuid.NewString())
				return req.WithContext(context.WithValue(ctx, chi.RouteCtxKey, chiCtx))
			},
			SetupBody: func() []byte {
				payload := product.UpdateProductRequest{
					ID:     uuid.Max.String(),
					Active: true,
				}
				b, _ := json.Marshal(payload)
				return b
			},
			MockSetup: func(svc *mockSvc.IProductService) {
				svc.On(
					"UpdateProductAvailability",
					mock.Anything,
					mock.Anything,
				).Return(
					errors.New("error"),
				)
			},
			WantErr:          true,
			ExpectedStatus:   http.StatusInternalServerError,
			ExpectedResponse: `{"code":"99","errors":"error"}`,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.Name, func(t *testing.T) {
			productSvc := mockSvc.NewIProductService(t)
			mockValidator := validator.New()
			tt.MockSetup(productSvc)

			controller := New(productSvc, mockValidator)
			req := httptest.NewRequest(http.MethodPut, "/products", bytes.NewBuffer(tt.SetupBody()))
			req = tt.SetupRequest(req)
			rr := httptest.NewRecorder()

			// Create the handler and serve the request
			handler := http.HandlerFunc(controller.UpdateProductAvailability)
			handler.ServeHTTP(rr, req)

			if rr.Body.String() != "" {
				t.Logf("Handler response body: %s", rr.Body.String())
			}

			// Assertions
			assert.Equal(t, tt.ExpectedStatus, rr.Code)
			assert.JSONEqf(t, string(tt.ExpectedResponse), rr.Body.String(), "Handler response body is not equal to expected. got %s, want %s", rr.Body.String(), tt.ExpectedResponse)

		})
	}
}
