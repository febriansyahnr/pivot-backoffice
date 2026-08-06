package v2InternalUnifiedPaymentController

import (
	"context"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/stretchr/testify/assert"
)

func TestValidateCustomerPayload(t *testing.T) {
	unifiedPaymentSvc := serviceMocks.NewIUnifiedPaymentService(t)
	customerSvc := serviceMocks.NewICustomerService(t)
	controller := paymentController{
		unifiedPaymentSvc: unifiedPaymentSvc,
		customerSvc:       customerSvc,
	}

	validCustomerInfo := &unifiedPaymentModel.CustomerInformation{
		GivenName: "John",
		Surname:   util.ValueToPtr("Doe"),
		Email:     "john@example.com",
		RefundPreference: &unifiedPaymentModel.UnifiedPaymentRefundPreference{
			Method: "bank_transfer",
			TransferDestination: &unifiedPaymentModel.RefundTransferDestination{
				ChannelCode: "PERMATA",
				ChannelInformation: &unifiedPaymentModel.RefundChannelInformation{
					AccountNumber: "1234567890",
					AccountName:   "John Doe",
				},
			},
		},
		StoredPaymentMethods: []*unifiedPaymentModel.CustomerPaymentMethod{},
	}

	validCustomerInfoWithPhone := *validCustomerInfo
	validCustomerInfoWithPhone.RefundPreference = nil
	validCustomerInfoWithPhone.PhoneNumber = &unifiedPaymentModel.UnifiedPaymentPhoneNumber{
		CountryCode: "+62",
		Number:      "8123456789",
	}

	tests := []struct {
		name      string
		payload   *unifiedPaymentModel.CreateUnifiedPaymentSessionRequest
		setupMock func()
		wantErr   bool
		errMsg    string
	}{
		{
			name: "SUCCESS: Neither CustomerID nor CustomerInformation provided",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				MerchantID: "merchant-123",
			},
			wantErr: false,
		},
		{
			name: "ERROR: Both CustomerID and CustomerInformation provided",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				MerchantID:          "merchant-123",
				CustomerID:          "customer-123",
				CustomerInformation: validCustomerInfo,
			},
			wantErr: true,
			errMsg:  "customer information conflict",
		},
		{
			name: "SUCCESS: Only CustomerInformation provided",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				MerchantID:          "merchant-123",
				CustomerInformation: validCustomerInfo,
			},
			setupMock: func() {
				customerSvc.On("CreateUnfiedPaymentCustomer", constant.ValueCtxMockType(), customerModel.CreateUnifiedPaymentCustomerRequest{
					MerchantID: "merchant-123",
					FirstName:  validCustomerInfo.GivenName,
					LastName:   *validCustomerInfo.Surname,
					Email:      validCustomerInfo.Email,
					Metadata: map[string]interface{}{
						"refundPreference": validCustomerInfo.RefundPreference,
					},
				}).Return(&customerModel.GeneralCustomerResponse{UUID: "new-customer-123"}, nil).Once()
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: With Phone Number",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				MerchantID:          "merchant-1234",
				CustomerInformation: &validCustomerInfoWithPhone,
			},
			setupMock: func() {
				customerSvc.On("CreateUnfiedPaymentCustomer", constant.ValueCtxMockType(), customerModel.CreateUnifiedPaymentCustomerRequest{
					MerchantID:       "merchant-1234",
					FirstName:        validCustomerInfoWithPhone.GivenName,
					LastName:         *validCustomerInfoWithPhone.Surname,
					Email:            validCustomerInfoWithPhone.Email,
					PhoneNumber:      validCustomerInfoWithPhone.PhoneNumber.Number,
					PhoneCountryCode: validCustomerInfoWithPhone.PhoneNumber.CountryCode,
					Metadata: map[string]interface{}{
						"refundPreference": validCustomerInfoWithPhone.RefundPreference,
					},
				}).Return(&customerModel.GeneralCustomerResponse{UUID: "new-customer-123"}, nil).Once()
			},
			wantErr: false,
		},
		{
			name: "ERROR: Creating customer fails",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				MerchantID:          "merchant-123",
				CustomerInformation: validCustomerInfo,
			},
			setupMock: func() {
				customerSvc.On("CreateUnfiedPaymentCustomer", constant.ValueCtxMockType(), customerModel.CreateUnifiedPaymentCustomerRequest{
					MerchantID: "merchant-123",
					FirstName:  validCustomerInfo.GivenName,
					LastName:   *validCustomerInfo.Surname,
					Email:      validCustomerInfo.Email,
					Metadata: map[string]interface{}{
						"refundPreference": validCustomerInfo.RefundPreference,
					},
				}).Return(nil, errors.New("failed to create customer")).Once()
			},
			wantErr: true,
			errMsg:  "failed to create customer",
		},
		{
			name: "SUCCESS: CustomerID provided and customer exists",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				MerchantID: "merchant-123",
				CustomerID: "existing-customer",
			},
			setupMock: func() {
				customerSvc.On("GetCustomerById", constant.ValueCtxMockType(), "existing-customer", "merchant-123").
					Return(&customerModel.GeneralCustomerResponse{UUID: "existing-customer"}, nil).Once()
			},
			wantErr: false,
		},
		{
			name: "ERROR: CustomerID provided but customer not found",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				MerchantID: "merchant-123",
				CustomerID: "non-existent-customer",
			},
			setupMock: func() {
				customerSvc.On("GetCustomerById", constant.ValueCtxMockType(), "non-existent-customer", "merchant-123").
					Return(nil, nil).Once()
			},
			wantErr: true,
			errMsg:  "customer not found",
		},
		{
			name: "ERROR: GetCustomerById returns error",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				MerchantID: "merchant-123",
				CustomerID: "customer-with-error",
			},
			setupMock: func() {
				customerSvc.On("GetCustomerById", constant.ValueCtxMockType(), "customer-with-error", "merchant-123").
					Return(nil, errors.New("database error")).Once()
			},
			wantErr: true,
			errMsg:  "database error",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setupMock != nil {
				test.setupMock()
			}

			ctx := context.Background()
			err := controller.ValidateCustomerPayload(ctx, test.payload)

			if test.wantErr {
				assert.Error(t, err)
				if test.errMsg != "" {
					assert.Contains(t, err.Error(), test.errMsg)
				}
			} else {
				assert.NoError(t, err)
				if test.payload.CustomerInformation != nil {
					assert.Equal(t, "new-customer-123", test.payload.CustomerID)
				}
			}
		})
	}
}
