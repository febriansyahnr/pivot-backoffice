package vendor

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	vendorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/vendor"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func createMultipartFormData(fields map[string]string, files map[string]string) (*bytes.Buffer, string, error) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			return nil, "", err
		}
	}

	for key, filename := range files {
		part, err := writer.CreateFormFile(key, filename)
		if err != nil {
			return nil, "", err
		}
		part.Write([]byte("test file content"))
	}

	if err := writer.Close(); err != nil {
		return nil, "", err
	}

	return body, writer.FormDataContentType(), nil
}

func TestCreate(t *testing.T) {
	createdVendor := &vendorModel.VendorResponse{
		UUID:                "test-uuid",
		MerchantID:          "merchant-123",
		Name:                "Test Vendor",
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
		fields       map[string]string
		files        map[string]string
		setup        func(svc *mockSvc.IVendorService)
		expectedCode int
	}{
		{
			name: "SUCCESS: Create Vendor",
			fields: map[string]string{
				"merchantId":          "merchant-123",
				"name":                "Test Vendor",
				"beneficialOwner":     "John Doe",
				"businessCategory":    "E-Commerce",
				"avgMonthlyTpvAmount": "1000000",
				"bankName":            "Bank ABC",
				"bankCode":            "ABC",
				"accountNumber":       "1234567890",
				"accountName":         "Test Account",
			},
			setup: func(svc *mockSvc.IVendorService) {
				svc.On(
					"Create",
					mock.Anything,
					mock.AnythingOfType("*vendor.CreateVendorRequest"),
				).Return(createdVendor, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "SUCCESS: Create Vendor with document files",
			fields: map[string]string{
				"merchantId":          "merchant-123",
				"name":                "Test Vendor",
				"beneficialOwner":     "John Doe",
				"businessCategory":    "E-Commerce",
				"avgMonthlyTpvAmount": "1000000",
				"bankName":            "Bank ABC",
				"bankCode":            "ABC",
				"accountNumber":       "1234567890",
				"accountName":         "Test Account",
			},
			files: map[string]string{
				"documents": "test_document.pdf",
			},
			setup: func(svc *mockSvc.IVendorService) {
				svc.On(
					"Create",
					mock.Anything,
					mock.AnythingOfType("*vendor.CreateVendorRequest"),
				).Return(createdVendor, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "ERROR: Validation failed - missing merchantId",
			fields: map[string]string{
				"name":                "Test Vendor",
				"beneficialOwner":     "John Doe",
				"businessCategory":    "E-Commerce",
				"avgMonthlyTpvAmount": "1000000",
				"bankName":            "Bank ABC",
				"bankCode":            "ABC",
				"accountNumber":       "1234567890",
				"accountName":         "Test Account",
			},
			setup:        func(svc *mockSvc.IVendorService) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "ERROR: Validation failed - missing name",
			fields: map[string]string{
				"merchantId":          "merchant-123",
				"beneficialOwner":     "John Doe",
				"businessCategory":    "E-Commerce",
				"avgMonthlyTpvAmount": "1000000",
				"bankName":            "Bank ABC",
				"bankCode":            "ABC",
				"accountNumber":       "1234567890",
				"accountName":         "Test Account",
			},
			setup:        func(svc *mockSvc.IVendorService) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "ERROR: Invalid avgMonthlyTpvAmount",
			fields: map[string]string{
				"merchantId":          "merchant-123",
				"name":                "Test Vendor",
				"beneficialOwner":     "John Doe",
				"businessCategory":    "E-Commerce",
				"avgMonthlyTpvAmount": "invalid-amount",
				"bankName":            "Bank ABC",
				"bankCode":            "ABC",
				"accountNumber":       "1234567890",
				"accountName":         "Test Account",
			},
			setup:        func(svc *mockSvc.IVendorService) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "ERROR: Vendor name already exists",
			fields: map[string]string{
				"merchantId":          "merchant-123",
				"name":                "Test Vendor",
				"beneficialOwner":     "John Doe",
				"businessCategory":    "E-Commerce",
				"avgMonthlyTpvAmount": "1000000",
				"bankName":            "Bank ABC",
				"bankCode":            "ABC",
				"accountNumber":       "1234567890",
				"accountName":         "Test Account",
			},
			setup: func(svc *mockSvc.IVendorService) {
				svc.On(
					"Create",
					mock.Anything,
					mock.AnythingOfType("*vendor.CreateVendorRequest"),
				).Return(nil, errors.New("vendor name already exists"))
			},
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "ERROR: Service error",
			fields: map[string]string{
				"merchantId":          "merchant-123",
				"name":                "Test Vendor",
				"beneficialOwner":     "John Doe",
				"businessCategory":    "E-Commerce",
				"avgMonthlyTpvAmount": "1000000",
				"bankName":            "Bank ABC",
				"bankCode":            "ABC",
				"accountNumber":       "1234567890",
				"accountName":         "Test Account",
			},
			setup: func(svc *mockSvc.IVendorService) {
				svc.On(
					"Create",
					mock.Anything,
					mock.AnythingOfType("*vendor.CreateVendorRequest"),
				).Return(nil, errors.New("internal server error"))
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

			req := httptest.NewRequest(http.MethodPost, "/card-funded-payout/vendors", body)
			req.Header.Set("Content-Type", contentType)
			req = req.WithContext(ctx)
			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(ctrl.Create)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedCode, rr.Code)
		})
	}
}
