package crmProductController

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/product"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetProductList(t *testing.T) {
	testcases := []struct {
		Name             string
		MockSetup        func(svc *mockSvc.IProductService)
		WantErr          bool
		ExpectedStatus   int
		ExpectedResponse string
	}{
		{
			Name: "SUCCESS",
			MockSetup: func(svc *mockSvc.IProductService) {
				svc.On(
					"GetProductList",
					mock.Anything,
				).Return(
					[]*product.Product{
						{
							UUID:   uuid.Max.String(),
							Name:   "PRODUCT 1",
							Active: true,
						},
					},
					nil,
				)
			},
			WantErr:          false,
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: `{"code":"00","data":[{"uuid":"ffffffff-ffff-ffff-ffff-ffffffffffff","name":"PRODUCT 1","active":true,"createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}]}`,
		},
		{
			Name: "ERROR: Error get product list",
			MockSetup: func(svc *mockSvc.IProductService) {
				svc.On("GetProductList", mock.Anything).Return(nil, errors.New("error"))
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
			req := httptest.NewRequest(http.MethodGet, "/products", nil)
			rr := httptest.NewRecorder()

			// Create the handler and serve the request
			handler := http.HandlerFunc(controller.GetProductList)
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
