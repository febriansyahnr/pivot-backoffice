package paymentService

import (
	"context"
	"errors"

	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/virtualAccount"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPaymentService_UpdatePayment(t *testing.T) {
	paymentId := uuid.NewString()
	amount := decimal.NewFromInt(1000000)
	totalAmount := decimal.NewFromFloat(1000)
	now := time.Now()
	expired := time.Now().Add(time.Hour * 24)
	newUuid := uuid.NewString()
	referenceId := "reference-id"
	processorReferenceNumber := "processor-reference-number"
	fee := decimal.NewFromFloat(1000)
	discount := decimal.NewFromFloat(1000)
	//metadata := "{\"testing\":\"testing\"}"

	metadataMap := map[string]any{"snapCore": map[string]any{
		"isClosedAmount": true,
		"isSingleUse":    true,
	}, "isSnap": false}

	metadataMapSnap := map[string]any{"snapCore": map[string]any{
		"isClosedAmount": false,
		"isSingleUse":    false,
	}, "isSnap": true}

	reqPayloadWithBills := &paymentModel.PaymentUpdateRequest{
		MerchantId: "merchant-id",
		PaymentId:  paymentId,
		TotalAmount: &paymentModel.Amount{
			Value:    amount,
			Currency: "IDR",
		},
		ExpiredAt:    &expired,
		AccountEmail: "example@email.com",
		AccountPhone: "08123123123",
		BillDetails: []snapCoreModel.BillDetail{
			{
				BillName: "name",
			},
			{
				BillNo: "bill-no",
			},
		},
		AdditionalInfo: map[string]any{
			"customer": map[string]any{
				"customerId": "99999",
				"name":       "Balerina",
			},
		},
	}

	reqPayload := &paymentModel.PaymentUpdateRequest{
		MerchantId: "merchant-id",
		PaymentId:  paymentId,
		TotalAmount: &paymentModel.Amount{
			Value:    amount,
			Currency: "IDR",
		},
		ExpiredAt:    &expired,
		AccountEmail: "example@email.com",
		AccountPhone: "08123123123",
		AdditionalInfo: map[string]any{
			"customer": map[string]any{
				"customerId": "99999",
				"name":       "Balerina",
			},
		},
	}

	reqExpiredPayload := &paymentModel.PaymentUpdateRequest{
		MerchantId: "merchant-id",
		PaymentId:  paymentId,
		TotalAmount: &paymentModel.Amount{
			Value:    amount,
			Currency: "IDR",
		},
		ExpiredAt: &expired,
	}

	reqPayloadInvalidAmount := &paymentModel.PaymentUpdateRequest{
		MerchantId: "merchant-id",
		PaymentId:  paymentId,
		TotalAmount: &paymentModel.Amount{
			Value:    decimal.NewFromInt(1000),
			Currency: "IDR",
		},
		ExpiredAt: &expired,
	}

	paymentMethod := &paymentModel.PaymentMethod{
		UUID:      newUuid,
		Type:      paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
		Category:  paymentConstant.PAYMENT_METHOD_CATEGORY_DISBURSEMENT_TOPUP,
		Name:      "VA Permata",
		Acquirer:  constant.BANK_ACQUIRER_PERMATA,
		CreatedAt: now,
		UpdatedAt: now,
	}

	payment := &paymentModel.Payment{
		UUID:                     "uuid-uuid-uuid",
		ReferenceID:              &referenceId,
		MerchantID:               "merchant-id",
		CustomerID:               "customer-id",
		PaymentMethodID:          "payment-method-id",
		ProcessorReferenceNumber: &processorReferenceNumber,
		Currency:                 "IDR",
		Amount:                   amount,
		Fee:                      &fee,
		Discount:                 &discount,
		TotalAmount:              totalAmount,
		Status:                   paymentConstant.PAYMENT_STATUS_PENDING,
		Metadata:                 &metadataMap,
		CreatedAt:                now,
		UpdatedAt:                now,
		PaymentMethod:            *paymentMethod,
	}

	paymentSnapOpenStatic := &paymentModel.Payment{
		UUID:                     "uuid-uuid-uuid",
		ReferenceID:              &referenceId,
		MerchantID:               "merchant-id",
		CustomerID:               "customer-id",
		PaymentMethodID:          "payment-method-id",
		ProcessorReferenceNumber: &processorReferenceNumber,
		Currency:                 "IDR",
		Amount:                   amount,
		Fee:                      &fee,
		Discount:                 &discount,
		TotalAmount:              totalAmount,
		Status:                   paymentConstant.PAYMENT_STATUS_PENDING,
		Metadata:                 &metadataMapSnap,
		CreatedAt:                now,
		UpdatedAt:                now,
		PaymentMethod:            *paymentMethod,
	}

	expiredAt := time.Now().Add(time.Hour * -24)

	updateResponse := &snapCoreModel.UpdateVirtualAccountResponseData{
		Number:    "va-number",
		ExpiredAt: &expiredAt,
		TotalAmount: snapCoreModel.Amount{
			Value:    amount.String(),
			Currency: "IDR",
		},
	}

	customer := &customerModel.Customer{
		UUID:        "customer-id",
		FirstName:   "John Doe",
		Email:       "test@gmail.com",
		PhoneNumber: "08123123123",
	}

	paymentItem := &paymentModel.PaymentItem{
		UUID:        "uuid-uuid-uuid",
		PaymentID:   "payment-id",
		Name:        "name",
		Description: "description",
		Qty:         1,
		Currency:    "IDR",
		Amount:      amount,
		TotalAmount: totalAmount,
		Metadata:    &metadataMap,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	paymentItems := []*paymentModel.PaymentItem{paymentItem}

	testCases := []struct {
		desc         string
		setupPayload func() *paymentModel.PaymentUpdateRequest
		wantErr      bool
		setupMock    func(
			paymentRepoMocks *repositoryMocks.IPaymentRepository,

			snapCoreMocks *repositoryMocks.ISnapCoreRepository,
			customerRepoMocks *repositoryMocks.ICustomerRepository,
			merchantRepoMocks *repositoryMocks.IMerchantRepository,
			paymentMethodRepoMocks *repositoryMocks.IPaymentMethodRepository,
		)
	}{
		{
			desc:    "SUCCESS: update payment method virtual account",
			wantErr: false,
			setupPayload: func() *paymentModel.PaymentUpdateRequest {
				return reqPayloadWithBills
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,

				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				paymentMethodRepoMocks *repositoryMocks.IPaymentMethodRepository,
			) {
				paymentRepoMocks.
					On(
						"GetPaymentById",
						mock.Anything,
						mock.AnythingOfType("string")).
					Return(payment, nil)

				customerRepoMocks.
					On(
						"FindCustomerById",
						mock.Anything,
						mock.Anything).
					Return(customer, nil)

				customerRepoMocks.
					On(
						"Update",
						mock.Anything,
						mock.Anything,
					).Return(nil)

				paymentMethodRepoMocks.
					On(
						"GetPaymentMethodById",
						mock.Anything,
						mock.Anything).
					Return(paymentMethod, nil)

				paymentRepoMocks.
					On(
						"UpdatePaymentItemsFromPaymentResponseItem",
						mock.Anything,
						mock.Anything,
						mock.Anything).
					Return(nil)

				snapCoreMocks.
					On(
						"UpdateVirtualAccount",
						mock.Anything,
						mock.AnythingOfType("snapCoreModel.UpdateVirtualAccountRequest")).
					Return(updateResponse, nil)

				paymentRepoMocks.
					On(
						"BeginTransaction",
						mock.Anything).
					Return(context.Background(), nil)

				paymentRepoMocks.
					On(
						"UpdatePayment",
						mock.Anything,
						mock.Anything,
						mock.Anything,
						mock.Anything,
						mock.Anything,
						mock.Anything,
						mock.Anything).
					Return(nil)

				paymentRepoMocks.
					On(
						"CommitTransaction",
						mock.Anything).
					Return(nil)
			},
		},
		{
			desc:    "SUCCESS: update payment virtual account with no bills open static",
			wantErr: false,
			setupPayload: func() *paymentModel.PaymentUpdateRequest {
				reqPayload.MinAmount = &paymentModel.Amount{
					Value:    decimal.NewFromInt(10000),
					Currency: "IDR",
				}
				reqPayload.MaxAmount = &paymentModel.Amount{
					Value:    decimal.NewFromInt(1000000),
					Currency: "IDR",
				}

				return reqPayload
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,

				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				paymentMethodRepoMocks *repositoryMocks.IPaymentMethodRepository,
			) {
				paymentRepoMocks.
					On(
						"GetPaymentById",
						mock.Anything,
						mock.AnythingOfType("string")).
					Return(payment, nil)

				customerRepoMocks.
					On(
						"FindCustomerById",
						mock.Anything,
						mock.Anything).
					Return(customer, nil)

				customerRepoMocks.
					On(
						"Update",
						mock.Anything,
						mock.Anything,
					).Return(nil)

				paymentMethodRepoMocks.
					On(
						"GetPaymentMethodById",
						mock.Anything,
						mock.Anything).
					Return(paymentMethod, nil)

				paymentRepoMocks.
					On(
						"GetPaymentItemsByPaymentId",
						mock.Anything,
						mock.Anything).
					Return(paymentItems, nil)

				snapCoreMocks.
					On(
						"UpdateVirtualAccount",
						mock.Anything,
						mock.AnythingOfType("snapCoreModel.UpdateVirtualAccountRequest")).
					Return(updateResponse, nil)

				paymentRepoMocks.
					On(
						"BeginTransaction",
						mock.Anything).
					Return(context.Background(), nil)

				paymentRepoMocks.
					On(
						"UpdatePayment",
						mock.Anything,
						mock.Anything,
						mock.Anything,
						mock.Anything,
						mock.Anything,
						mock.Anything,
						mock.Anything).
					Return(nil)

				paymentRepoMocks.
					On(
						"CommitTransaction",
						mock.Anything).
					Return(nil)
			},
		},
		{
			desc:    "FAILED: update payment with VA expired less than now",
			wantErr: true,
			setupPayload: func() *paymentModel.PaymentUpdateRequest {
				expiredLessThanNow := time.Now().Add(-time.Hour)
				reqExpiredPayload.ExpiredAt = &expiredLessThanNow
				return reqExpiredPayload
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,

				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				paymentMethodRepoMocks *repositoryMocks.IPaymentMethodRepository,
			) {
				paymentRepoMocks.
					On(
						"GetPaymentById",
						mock.Anything,
						mock.AnythingOfType("string")).
					Return(payment, nil)

				customerRepoMocks.
					On(
						"FindCustomerById",
						mock.Anything,
						mock.Anything).
					Return(customer, nil)

				paymentMethodRepoMocks.
					On(
						"GetPaymentMethodById",
						mock.Anything,
						mock.Anything).
					Return(paymentMethod, nil)

				paymentRepoMocks.
					On(
						"GetPaymentItemsByPaymentId",
						mock.Anything,
						mock.Anything).
					Return(paymentItems, nil)
			},
		},
		{
			desc:    "FAILED: update payment amount less than 10000",
			wantErr: true,
			setupPayload: func() *paymentModel.PaymentUpdateRequest {
				return reqPayloadInvalidAmount
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,

				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				paymentMethodRepoMocks *repositoryMocks.IPaymentMethodRepository,
			) {
				paymentRepoMocks.
					On(
						"GetPaymentById",
						mock.Anything,
						mock.AnythingOfType("string")).
					Return(payment, nil)

				customerRepoMocks.
					On(
						"FindCustomerById",
						mock.Anything,
						mock.Anything).
					Return(customer, nil)

				paymentMethodRepoMocks.
					On(
						"GetPaymentMethodById",
						mock.Anything,
						mock.Anything).
					Return(paymentMethod, nil)

				paymentRepoMocks.
					On(
						"GetPaymentItemsByPaymentId",
						mock.Anything,
						mock.Anything).
					Return(paymentItems, nil)
			},
		},
		{
			desc:    "FAILED: update payment open static but min amount is above max amount",
			wantErr: true,
			setupPayload: func() *paymentModel.PaymentUpdateRequest {
				reqPayload.MinAmount = &paymentModel.Amount{
					Value:    decimal.NewFromInt(1000000),
					Currency: "IDR",
				}
				reqPayload.MaxAmount = &paymentModel.Amount{
					Value:    decimal.NewFromInt(10000),
					Currency: "IDR",
				}
				return reqPayload
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,

				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				paymentMethodRepoMocks *repositoryMocks.IPaymentMethodRepository,
			) {
				paymentRepoMocks.
					On(
						"GetPaymentById",
						mock.Anything,
						mock.AnythingOfType("string")).
					Return(paymentSnapOpenStatic, nil)

				customerRepoMocks.
					On(
						"FindCustomerById",
						mock.Anything,
						mock.Anything).
					Return(customer, nil)

				paymentMethodRepoMocks.
					On(
						"GetPaymentMethodById",
						mock.Anything,
						mock.Anything).
					Return(paymentMethod, nil)

				paymentRepoMocks.
					On(
						"GetPaymentItemsByPaymentId",
						mock.Anything,
						mock.Anything).
					Return(paymentItems, nil)
			},
		},
		{
			desc: "FAILED: payment not found",
			setupPayload: func() *paymentModel.PaymentUpdateRequest {
				return reqPayload
			},
			wantErr: true,
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,

				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				paymentMethodRepoMocks *repositoryMocks.IPaymentMethodRepository,
			) {
				paymentRepoMocks.
					On(
						"GetPaymentById",
						mock.Anything,
						mock.AnythingOfType("string")).
					Return(nil, nil)
			},
		},
		{
			desc:    "FAILED: get payment by id",
			wantErr: true,
			setupPayload: func() *paymentModel.PaymentUpdateRequest {
				return reqPayload
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,

				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				paymentMethodRepoMocks *repositoryMocks.IPaymentMethodRepository,
			) {
				paymentRepoMocks.
					On(
						"GetPaymentById",
						mock.Anything,
						mock.AnythingOfType("string")).
					Return(nil, assert.AnError)
			},
		},
		{
			desc: "FAILED: failed to get customer",
			setupPayload: func() *paymentModel.PaymentUpdateRequest {
				return reqPayload
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,

				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				paymentMethodRepoMocks *repositoryMocks.IPaymentMethodRepository,
			) {
				paymentRepoMocks.
					On(
						"GetPaymentById",
						mock.Anything,
						mock.AnythingOfType("string")).
					Return(payment, nil)

				customerRepoMocks.
					On(
						"FindCustomerById",
						mock.Anything,
						mock.Anything).
					Return(nil, fmt.Errorf("error when get customer data by id"))
			},
			wantErr: true,
		},
		{
			desc:    "FAILED: customer merchant id not match with request merchant id",
			wantErr: true,
			setupPayload: func() *paymentModel.PaymentUpdateRequest {
				return &paymentModel.PaymentUpdateRequest{
					PaymentId:  paymentId,
					MerchantId: "unmatch-merchant-id",
					TotalAmount: &paymentModel.Amount{
						Value:    amount,
						Currency: "IDR",
					},
					ExpiredAt: &expired,
				}
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,

				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				paymentMethodRepoMocks *repositoryMocks.IPaymentMethodRepository,
			) {
				paymentRepoMocks.
					On(
						"GetPaymentById",
						mock.Anything,
						mock.AnythingOfType("string")).
					Return(payment, nil)

				customerRepoMocks.
					On(
						"FindCustomerById",
						mock.Anything,
						mock.Anything).
					Return(&customerModel.Customer{
						UUID:        "customer-id",
						FirstName:   "John Doe",
						MerchantID:  "merchant-id",
						Email:       "test@gmail.com",
						PhoneNumber: "08123123123",
					}, nil)
			},
		},
		{
			desc:    "FAILED: payment method not found",
			wantErr: true,
			setupPayload: func() *paymentModel.PaymentUpdateRequest {
				return reqPayload
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,

				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				paymentMethodRepoMocks *repositoryMocks.IPaymentMethodRepository,
			) {
				paymentRepoMocks.
					On(
						"GetPaymentById",
						mock.Anything,
						mock.AnythingOfType("string")).
					Return(payment, nil)

				customerRepoMocks.
					On(
						"FindCustomerById",
						mock.Anything,
						mock.Anything).
					Return(customer, nil)

				paymentMethodRepoMocks.
					On(
						"GetPaymentMethodById",
						mock.Anything,
						mock.Anything).
					Return(nil, fmt.Errorf("error when get payment method data by id"))
			},
		},
		{
			desc:    "FAILED: payment item not found",
			wantErr: true,
			setupPayload: func() *paymentModel.PaymentUpdateRequest {
				return reqPayload
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,

				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				paymentMethodRepoMocks *repositoryMocks.IPaymentMethodRepository,
			) {
				paymentRepoMocks.
					On(
						"GetPaymentById",
						mock.Anything,
						mock.AnythingOfType("string")).
					Return(payment, nil)

				customerRepoMocks.
					On(
						"FindCustomerById",
						mock.Anything,
						mock.Anything).
					Return(customer, nil)

				paymentMethodRepoMocks.
					On(
						"GetPaymentMethodById",
						mock.Anything,
						mock.Anything).
					Return(paymentMethod, nil)

				paymentRepoMocks.
					On(
						"GetPaymentItemsByPaymentId",
						mock.Anything,
						mock.Anything).
					Return(nil, assert.AnError)
			},
		},
		{
			desc:    "FAILED: unmarshal payment metadata",
			wantErr: true,
			setupPayload: func() *paymentModel.PaymentUpdateRequest {
				return reqPayload
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,

				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				paymentMethodRepoMocks *repositoryMocks.IPaymentMethodRepository,
			) {
				paymentRepoMocks.
					On(
						"GetPaymentById",
						mock.Anything,
						mock.AnythingOfType("string")).
					Return(&paymentModel.Payment{
						UUID:                     "uuid-uuid-uuid",
						ReferenceID:              &referenceId,
						MerchantID:               "merchant-id",
						CustomerID:               "customer-id",
						PaymentMethodID:          "payment-method-id",
						ProcessorReferenceNumber: &processorReferenceNumber,
						Currency:                 "IDR",
						Amount:                   amount,
						Fee:                      &fee,
						Discount:                 &discount,
						TotalAmount:              totalAmount,
						Status:                   paymentConstant.PAYMENT_STATUS_PENDING,
						Metadata:                 &map[string]any{"snapCore": make(chan int)},
						CreatedAt:                now,
						UpdatedAt:                now,
						PaymentMethod:            *paymentMethod,
					}, nil)

				customerRepoMocks.
					On(
						"FindCustomerById",
						mock.Anything,
						mock.Anything).
					Return(customer, nil)

				paymentMethodRepoMocks.
					On(
						"GetPaymentMethodById",
						mock.Anything,
						mock.Anything).
					Return(paymentMethod, nil)

				paymentRepoMocks.
					On(
						"GetPaymentItemsByPaymentId",
						mock.Anything,
						mock.Anything).
					Return(paymentItems, nil)

			},
		},
		{
			desc:    "FAILED: error begin transaction database",
			wantErr: true,
			setupPayload: func() *paymentModel.PaymentUpdateRequest {
				return reqPayload
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,

				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				paymentMethodRepoMocks *repositoryMocks.IPaymentMethodRepository,
			) {
				paymentRepoMocks.
					On(
						"GetPaymentById",
						mock.Anything,
						mock.AnythingOfType("string")).
					Return(payment, nil)

				paymentMethodRepoMocks.
					On(
						"GetPaymentMethodById",
						mock.Anything,
						mock.Anything).
					Return(paymentMethod, nil)

				customerRepoMocks.
					On(
						"FindCustomerById",
						mock.Anything,
						mock.Anything).
					Return(customer, nil)

				paymentRepoMocks.
					On(
						"GetPaymentItemsByPaymentId",
						mock.Anything,
						mock.Anything).
					Return(paymentItems, nil)

				snapCoreMocks.
					On(
						"UpdateVirtualAccount",
						mock.Anything,
						mock.AnythingOfType("snapCoreModel.UpdateVirtualAccountRequest")).
					Return(updateResponse, nil)

				paymentRepoMocks.
					On(
						"BeginTransaction",
						mock.Anything).
					Return(nil, errors.New("error when begin transaction"))
			},
		},
		{
			desc:    "FAILED: update virtual account to snap",
			wantErr: true,
			setupPayload: func() *paymentModel.PaymentUpdateRequest {
				return reqPayload
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,

				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				paymentMethodRepoMocks *repositoryMocks.IPaymentMethodRepository,
			) {
				paymentRepoMocks.
					On(
						"GetPaymentById",
						mock.Anything,
						mock.AnythingOfType("string")).
					Return(payment, nil)

				customerRepoMocks.
					On(
						"FindCustomerById",
						mock.Anything,
						mock.Anything).
					Return(customer, nil)

				paymentMethodRepoMocks.
					On(
						"GetPaymentMethodById",
						mock.Anything,
						mock.Anything).
					Return(paymentMethod, nil)

				paymentRepoMocks.
					On(
						"GetPaymentItemsByPaymentId",
						mock.Anything,
						mock.Anything).
					Return(paymentItems, nil)

				snapCoreMocks.
					On(
						"UpdateVirtualAccount",
						mock.Anything,
						mock.AnythingOfType("snapCoreModel.UpdateVirtualAccountRequest")).
					Return(nil, assert.AnError)

			},
		},
		{
			desc:    "FAILED: failed to update payment",
			wantErr: true,
			setupPayload: func() *paymentModel.PaymentUpdateRequest {
				return reqPayload
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,

				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				paymentMethodRepoMocks *repositoryMocks.IPaymentMethodRepository,
			) {
				paymentRepoMocks.
					On(
						"GetPaymentById",
						mock.Anything,
						mock.AnythingOfType("string")).
					Return(payment, nil)

				customerRepoMocks.
					On(
						"FindCustomerById",
						mock.Anything,
						mock.Anything).
					Return(customer, nil)

				customerRepoMocks.
					On(
						"Update",
						mock.Anything,
						mock.Anything,
					).Return(nil)

				paymentMethodRepoMocks.
					On(
						"GetPaymentMethodById",
						mock.Anything,
						mock.Anything).
					Return(paymentMethod, nil)

				paymentRepoMocks.
					On(
						"GetPaymentItemsByPaymentId",
						mock.Anything,
						mock.Anything).
					Return(paymentItems, nil)

				snapCoreMocks.
					On(
						"UpdateVirtualAccount",
						mock.Anything,
						mock.AnythingOfType("snapCoreModel.UpdateVirtualAccountRequest")).
					Return(updateResponse, nil)

				paymentRepoMocks.
					On(
						"BeginTransaction",
						mock.Anything).
					Return(context.Background(), nil)

				paymentRepoMocks.
					On(
						"UpdatePayment",
						mock.Anything,
						mock.Anything,
						mock.Anything,
						mock.Anything,
						mock.Anything,
						mock.Anything,
						mock.Anything).
					Return(assert.AnError)

			},
		},
		{
			desc:    "FAILED: failed to commit transaction",
			wantErr: true,
			setupPayload: func() *paymentModel.PaymentUpdateRequest {
				return reqPayload
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,

				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				paymentMethodRepoMocks *repositoryMocks.IPaymentMethodRepository,
			) {
				paymentRepoMocks.
					On(
						"GetPaymentById",
						mock.Anything,
						mock.AnythingOfType("string")).
					Return(payment, nil)

				customerRepoMocks.
					On(
						"FindCustomerById",
						mock.Anything,
						mock.Anything).
					Return(customer, nil)

				customerRepoMocks.
					On(
						"Update",
						mock.Anything,
						mock.Anything,
					).Return(nil)

				paymentMethodRepoMocks.
					On(
						"GetPaymentMethodById",
						mock.Anything,
						mock.Anything).
					Return(paymentMethod, nil)

				paymentRepoMocks.
					On(
						"GetPaymentItemsByPaymentId",
						mock.Anything,
						mock.Anything).
					Return(paymentItems, nil)

				snapCoreMocks.
					On(
						"UpdateVirtualAccount",
						mock.Anything,
						mock.AnythingOfType("snapCoreModel.UpdateVirtualAccountRequest")).
					Return(updateResponse, nil)

				paymentRepoMocks.
					On(
						"BeginTransaction",
						mock.Anything).
					Return(context.Background(), nil)

				paymentRepoMocks.
					On(
						"UpdatePayment",
						mock.Anything,
						mock.Anything,
						mock.Anything,
						mock.Anything,
						mock.Anything,
						mock.Anything,
						mock.Anything).
					Return(nil)

				paymentRepoMocks.
					On(
						"CommitTransaction",
						mock.Anything).
					Return(assert.AnError)
			},
		},
		{
			desc: "FAILED: payment status is pending",
			setupPayload: func() *paymentModel.PaymentUpdateRequest {
				return reqPayload
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,

				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				paymentMethodRepoMocks *repositoryMocks.IPaymentMethodRepository,
			) {
				payment.Status = paymentConstant.PAYMENT_STATUS_SUCCESS
				paymentRepoMocks.
					On(
						"GetPaymentById",
						mock.Anything,
						mock.AnythingOfType("string")).
					Return(&paymentModel.Payment{
						UUID:                     "uuid-uuid-uuid",
						ReferenceID:              &referenceId,
						MerchantID:               "merchant-id",
						CustomerID:               "customer-id",
						PaymentMethodID:          "payment-method-id",
						ProcessorReferenceNumber: &processorReferenceNumber,
						Currency:                 "IDR",
						Amount:                   amount,
						Fee:                      &fee,
						Discount:                 &discount,
						TotalAmount:              totalAmount,
						Status:                   paymentConstant.PAYMENT_STATUS_SUCCESS,
						Metadata:                 &metadataMap,
						CreatedAt:                now,
						UpdatedAt:                now,
						PaymentMethod:            *paymentMethod,
					}, nil)
			},
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			paymentRepoMocks := repositoryMocks.NewIPaymentRepository(t)
			loggerMocks, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			snapCoreMocks := repositoryMocks.NewISnapCoreRepository(t)
			customerRepoMocks := repositoryMocks.NewICustomerRepository(t)
			merchantRepoMocks := repositoryMocks.NewIMerchantRepository(t)
			paymentMethodRepoMocks := repositoryMocks.NewIPaymentMethodRepository(t)
			tc.setupMock(paymentRepoMocks, snapCoreMocks, customerRepoMocks, merchantRepoMocks, paymentMethodRepoMocks)

			paymentSvc := New(paymentRepoMocks, loggerMocks, snapCoreMocks, customerRepoMocks, merchantRepoMocks, paymentMethodRepoMocks, nil)

			ctx := context.Background()
			_, err := paymentSvc.UpdatePayment(ctx, tc.setupPayload())

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			paymentRepoMocks.AssertExpectations(t)
			customerRepoMocks.AssertExpectations(t)

			snapCoreMocks.AssertExpectations(t)
			merchantRepoMocks.AssertExpectations(t)
		})
	}
}
