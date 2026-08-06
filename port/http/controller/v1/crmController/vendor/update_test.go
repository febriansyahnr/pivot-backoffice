package vendor

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	vendorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/vendor"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdate(t *testing.T) {
	updatedVendor := &vendorModel.VendorResponse{
		UUID:                "123e4567-e89b-12d3-a456-426614174000",
		MerchantID:          "merchant-123",
		Name:                "Updated Vendor",
		BeneficialOwner:     "John Doe",
		BusinessCategory:    "E-Commerce",
		AvgMonthlyTpvAmount: decimal.NewFromInt(1000000),
		BankName:            "Bank ABC",
		BankCode:            "ABC",
		AccountNumber:       "1234567890",
		AccountName:         "Test Account",
		Status:              "ACTIVE",
	}

	testCases := []struct {
		name         string
		paramID      string
		fields       map[string]string
		files        map[string]string
		setup        func(svc *mockSvc.IVendorService)
		expectedCode int
	}{
		{
			name:    "SUCCESS: Update Vendor name",
			paramID: "123e4567-e89b-12d3-a456-426614174000",
			fields: map[string]string{
				"name": "Updated Vendor",
			},
			setup: func(svc *mockSvc.IVendorService) {
				svc.On(
					"Update",
					mock.Anything,
					mock.AnythingOfType("*vendor.UpdateVendorRequest"),
				).Return(updatedVendor, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:    "SUCCESS: Update Vendor with multiple fields",
			paramID: "123e4567-e89b-12d3-a456-426614174000",
			fields: map[string]string{
				"name":                "Updated Vendor",
				"beneficialOwner":     "Jane Doe",
				"avgMonthlyTpvAmount": "2000000",
			},
			setup: func(svc *mockSvc.IVendorService) {
				svc.On(
					"Update",
					mock.Anything,
					mock.AnythingOfType("*vendor.UpdateVendorRequest"),
				).Return(updatedVendor, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:    "SUCCESS: Update Vendor with document files",
			paramID: "123e4567-e89b-12d3-a456-426614174000",
			fields: map[string]string{
				"name": "Updated Vendor",
			},
			files: map[string]string{
				"documents": "updated_document.pdf",
			},
			setup: func(svc *mockSvc.IVendorService) {
				svc.On(
					"Update",
					mock.Anything,
					mock.AnythingOfType("*vendor.UpdateVendorRequest"),
				).Return(updatedVendor, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:    "ERROR: Invalid UUID",
			paramID: "not-a-valid-uuid",
			fields: map[string]string{
				"name": "Updated Vendor",
			},
			setup:        func(svc *mockSvc.IVendorService) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:    "ERROR: Invalid avgMonthlyTpvAmount",
			paramID: "123e4567-e89b-12d3-a456-426614174000",
			fields: map[string]string{
				"avgMonthlyTpvAmount": "invalid-amount",
			},
			setup:        func(svc *mockSvc.IVendorService) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:    "ERROR: Service error - vendor not found",
			paramID: "123e4567-e89b-12d3-a456-426614174000",
			fields: map[string]string{
				"name": "Updated Vendor",
			},
			setup: func(svc *mockSvc.IVendorService) {
				svc.On(
					"Update",
					mock.Anything,
					mock.AnythingOfType("*vendor.UpdateVendorRequest"),
				).Return(nil, errors.New("vendor not found"))
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			svc := mockSvc.NewIVendorService(t)
			validate := validator.New()

			tc.setup(svc)

			ctrl := New(svc, validate)

			body, contentType, err := createMultipartFormData(tc.fields, tc.files)
			assert.NoError(t, err)

			req := httptest.NewRequest(http.MethodPut, "/card-funded-payout/vendors/"+tc.paramID, body)
			req.Header.Set("Content-Type", contentType)
			req = req.WithContext(ctx)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.paramID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(ctrl.Update)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedCode, rr.Code)
		})
	}
}
