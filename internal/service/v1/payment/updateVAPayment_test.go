package paymentService

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/virtualAccount"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetAndUpdateVirtualAccountPayment(t *testing.T) {
	validRequest := &paymentModel.VirtualAccountPaymentNotificationRequest{
		Acquirer: constant.BANK_ACQUIRER_PERMATA,
		Number:   VaNumber,
		Status:   paymentConstant.VirtualAccountStatusPaid,
		PaidAmount: commonModel.Amount{
			Currency: "IDR",
			Value:    "1000000.00",
		},
	}
	invalidPaidAmountRequest := &paymentModel.VirtualAccountPaymentNotificationRequest{
		Acquirer: constant.BANK_ACQUIRER_PERMATA,
		Number:   VaNumber,
		Status:   paymentConstant.VirtualAccountStatusPaid,
		PaidAmount: commonModel.Amount{
			Currency: "IDR",
			Value:    "s1000000.00",
		},
	}
	invalidStatusRequest := &paymentModel.VirtualAccountPaymentNotificationRequest{
		Acquirer: constant.BANK_ACQUIRER_PERMATA,
		Number:   VaNumber,
		Status:   "invalid-status",
		PaidAmount: commonModel.Amount{
			Currency: "IDR",
			Value:    "1000000.00",
		},
	}

	invalidDifferentAcquirerRequest := &paymentModel.VirtualAccountPaymentNotificationRequest{
		Acquirer: "not-registered-acquirer",
		Number:   VaNumber,
		Status:   paymentConstant.VirtualAccountStatusPaid,
		PaidAmount: commonModel.Amount{
			Currency: "IDR",
			Value:    "s1000000.00",
		},
	}

	snapCoreResp := snapCoreModel.CreateVirtualAccountResponseData{
		IsSingleUse:    true,
		IsClosedAmount: true,
	}

	paymentAmount, _ := decimal.NewFromString("1000000.00")
	validPayment := &paymentModel.Payment{
		UUID:        uuid.NewString(),
		Currency:    "IDR",
		Amount:      paymentAmount,
		TotalAmount: paymentAmount,
		Metadata:    &map[string]any{"snapCore": snapCoreResp},

		PaymentMethod: paymentModel.PaymentMethod{
			Acquirer: constant.BANK_ACQUIRER_PERMATA,
		},
	}

	testCases := []struct {
		name       string
		input      *paymentModel.VirtualAccountPaymentNotificationRequest
		mocksSetup func(
			paymentRepoMocks *repositoryMocks.IPaymentRepository,

			snapCoreMocks *repositoryMocks.ISnapCoreRepository,
			customerRepoMocks *repositoryMocks.ICustomerRepository,
			merchantRepoMocks *repositoryMocks.IMerchantRepository,
		)
		wantErr bool
	}{
		{
			name:  "SUCCESS: Get and update VA payment",
			input: validRequest,
			mocksSetup: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,

				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
			) {
				paymentRepoMocks.On(
					"GetActivePaymentByProcessorReferenceNumber",
					constant.ValueCtxMockType(),
					mock.Anything,
				).Return(validPayment, nil)

				paymentRepoMocks.On(
					"UpdatePaymentStatus",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.TimeMockType(),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name:  "ERROR: Invalid status on request",
			input: invalidStatusRequest,
			mocksSetup: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,

				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
			) {
			},
			wantErr: true,
		},
		{
			name:  "ERROR: Active payment not found",
			input: validRequest,
			mocksSetup: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,

				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
			) {
				paymentRepoMocks.On(
					"GetActivePaymentByProcessorReferenceNumber",
					constant.ValueCtxMockType(),
					mock.Anything,
				).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name:  "ERROR: GetActivePaymentByProcessorReferenceNumber",
			input: validRequest,
			mocksSetup: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,

				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
			) {
				paymentRepoMocks.On(
					"GetActivePaymentByProcessorReferenceNumber",
					constant.ValueCtxMockType(),
					mock.Anything,
				).Return(nil, errors.New("some-error"))
			},
			wantErr: true,
		},
		{
			name:  "ERROR: Different acquirer",
			input: invalidDifferentAcquirerRequest,
			mocksSetup: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,

				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
			) {
				paymentRepoMocks.On(
					"GetActivePaymentByProcessorReferenceNumber",
					constant.ValueCtxMockType(),
					mock.Anything,
				).Return(validPayment, nil)
			},
			wantErr: true,
		},
		{
			name:  "ERROR: Invalid paid amount value",
			input: invalidPaidAmountRequest,
			mocksSetup: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,

				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
			) {
				paymentRepoMocks.On(
					"GetActivePaymentByProcessorReferenceNumber",
					constant.ValueCtxMockType(),
					mock.Anything,
				).Return(validPayment, nil)

			},
			wantErr: true,
		},
		{
			name:  "ERROR: UpdatePaymentStatus service",
			input: validRequest,
			mocksSetup: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,

				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
			) {
				paymentRepoMocks.On(
					"GetActivePaymentByProcessorReferenceNumber",
					constant.ValueCtxMockType(),
					mock.Anything,
				).Return(validPayment, nil)

				paymentRepoMocks.On(
					"UpdatePaymentStatus",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.TimeMockType(),
				).Return(errors.New("some-error"))
			},
			wantErr: true,
		},
		{
			name:  "ERROR: Paid amount not match",
			input: validRequest,
			mocksSetup: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,

				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
			) {
				changeAmount, _ := decimal.NewFromString("1500000.00")
				validPayment.Amount = changeAmount
				paymentRepoMocks.On(
					"GetActivePaymentByProcessorReferenceNumber",
					constant.ValueCtxMockType(),
					mock.Anything,
				).Return(validPayment, nil)

			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockPaymentRepo := repositoryMocks.NewIPaymentRepository(t)
			mockSnapCoreSvc := repositoryMocks.NewISnapCoreRepository(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockCustomerRepo := repositoryMocks.NewICustomerRepository(t)
			mockMerchantRepo := repositoryMocks.NewIMerchantRepository(t)
			tc.mocksSetup(mockPaymentRepo, mockSnapCoreSvc, mockCustomerRepo, mockMerchantRepo)

			paymentSvc := New(mockPaymentRepo, mockLogger, mockSnapCoreSvc, mockCustomerRepo, mockMerchantRepo, nil, nil)
			ctx := context.Background()
			_, err := paymentSvc.GetAndUpdateVirtualAccountPayment(ctx, tc.input)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockPaymentRepo.AssertExpectations(t)
			mockSnapCoreSvc.AssertExpectations(t)

		})
	}
}

func TestExpireVirtualAccountPayment(pt *testing.T) {
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	mockPayment := repositoryMocks.NewIPaymentRepository(pt)
	mockCustomer := repositoryMocks.NewICustomerRepository(pt)

	mockCustomer.On("FindCustomerById", mock.Anything, mock.Anything).Return(&customerModel.Customer{}, nil)

	service := New(
		mockPayment, mockLogger, repositoryMocks.NewISnapCoreRepository(pt),
		mockCustomer, repositoryMocks.NewIMerchantRepository(pt), repositoryMocks.NewIPaymentMethodRepository(pt), nil,
	)

	tests := []struct {
		name       string
		request    paymentModel.VirtualAccountPaymentNotificationRequest
		mockSetup  func(p *repositoryMocks.IPaymentRepository)
		wantErr    string
		wantResult *paymentModel.Payment
	}{
		{
			name: "ERROR:Get active payment/Invalid session id",
			mockSetup: func(p *repositoryMocks.IPaymentRepository) {
				p.On(
					"GetActivePaymentByProcessorReferenceNumber", constant.ValueCtxMockType(), mock.Anything,
				).Once().Return(nil, errors.New("invalid session id"))
			},
			wantErr: "invalid session id",
		},
		{
			name: "ERROR:Get active payment/Payment not found",
			mockSetup: func(p *repositoryMocks.IPaymentRepository) {
				p.On(
					"GetActivePaymentByProcessorReferenceNumber", constant.ValueCtxMockType(), mock.Anything,
				).Once().Return(nil, nil)
			},
			wantErr: "payment not found",
		},
		{
			name:    "ERROR:Identity of the acquirer is different",
			request: paymentModel.VirtualAccountPaymentNotificationRequest{Acquirer: "ACQ-0001"},
			mockSetup: func(p *repositoryMocks.IPaymentRepository) {
				p.On(
					"GetActivePaymentByProcessorReferenceNumber", constant.ValueCtxMockType(), mock.Anything,
				).Once().Return(&paymentModel.Payment{PaymentMethod: paymentModel.PaymentMethod{Acquirer: ""}}, nil)
			},
			wantErr: "payment not found",
		},
		{
			name: "ERROR:Invalid payment metadata format",
			mockSetup: func(p *repositoryMocks.IPaymentRepository) {
				p.On(
					"GetActivePaymentByProcessorReferenceNumber", constant.ValueCtxMockType(), mock.Anything,
				).Once().Return(&paymentModel.Payment{Metadata: &map[string]any{"chan": make(chan struct{}, 1)}}, nil)
			},
			wantErr: "json: unsupported type: chan struct {}",
		},
		{
			name:    "ERROR:Update payment",
			request: paymentModel.VirtualAccountPaymentNotificationRequest{ExpiredAt: &time.Time{}},
			mockSetup: func(p *repositoryMocks.IPaymentRepository) {
				p.On(
					"GetActivePaymentByProcessorReferenceNumber", constant.ValueCtxMockType(), mock.Anything,
				).Return(&paymentModel.Payment{}, nil)
				p.On(
					"UpdatePayment", constant.ValueCtxMockType(), constant.StringMockType(), constant.DecimalMockType(), constant.DecimalMockType(), constant.StringMockType(), constant.StringMockType(), constant.TimeMockType(),
				).Once().Return(errors.New("invalid session id"))
			},
			wantErr: "invalid session id",
		},
		{
			name:    "ERROR:Update payment status",
			request: paymentModel.VirtualAccountPaymentNotificationRequest{ExpiredAt: &time.Time{}},
			mockSetup: func(p *repositoryMocks.IPaymentRepository) {
				p.On(
					"UpdatePayment", constant.ValueCtxMockType(), constant.StringMockType(), constant.DecimalMockType(), constant.DecimalMockType(), constant.StringMockType(), constant.StringMockType(), constant.TimeMockType(),
				).Return(nil)
				p.On(
					"UpdatePaymentStatus", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(), constant.StringMockType(), constant.TimeMockType(),
				).Once().Return(errors.New("duplicate unique ID"))
			},
			wantErr: "duplicate unique ID",
		},
		{
			name:    "SUCCESS",
			request: paymentModel.VirtualAccountPaymentNotificationRequest{ExpiredAt: &time.Time{}},
			mockSetup: func(p *repositoryMocks.IPaymentRepository) {
				p.On(
					"UpdatePaymentStatus", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(), constant.StringMockType(), constant.TimeMockType(),
				).Return(nil)
			},
			wantResult: &paymentModel.Payment{Status: paymentConstant.PAYMENT_STATUS_VOID},
		},
	}
	for _, test := range tests {
		pt.Run(test.name, func(t *testing.T) {

			test.mockSetup(mockPayment)
			test.request.Status = paymentConstant.VirtualAccountStatusExpired

			result, err := service.GetAndUpdateVirtualAccountPayment(context.Background(), &test.request)
			if test.wantErr != "" {
				require.NotNil(t, err)
				assert.ErrorContains(t, err, test.wantErr)
				return
			}
			require.Nil(t, err)
			assert.Equal(t, test.wantResult.Status, result.Status)
		})
	}
}
