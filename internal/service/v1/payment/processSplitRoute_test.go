package paymentService_test

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	splitRoutingPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/splitRoutingPayment"
	transferModel "github.com/paper-indonesia/pivot-backoffice/internal/model/transfer"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/payment"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
)

func TestProcessSplitRoute(t *testing.T) {
	paymentRepo := repositoryMocks.NewIPaymentRepository(t)
	merchantRepo := repositoryMocks.NewIMerchantRepository(t)
	transferSvc := serviceMocks.NewITransferService(t)
	ledgerSvc := serviceMocks.NewILedgerService(t)
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	paymentID := uuid.NewString()
	merchantID := uuid.NewString()
	parentMerchantID := uuid.NewString()
	referenceID := "ref-id"

	splitRouteConfigs := []*splitRoutingPaymentModel.PaymentSplitRoutingConfiguration{
		{
			TransferID:       "",
			MerchantId:       uuid.NewString(),
			Type:             constant.SplitRoutingPaymentTypePercentage,
			PercentageAmount: 10,
		},
	}
	validPayment := &paymentModel.Payment{
		UUID:       paymentID,
		MerchantID: merchantID,
		Metadata: &map[string]interface{}{
			constant.SplitRoutingPaymentConfigKey: splitRouteConfigs,
		},
		ReferenceID: &referenceID,
		TotalAmount: decimal.NewFromFloat(10000),
	}
	validMerchant := &merchantModel.Merchant{
		ParentID: sql.NullString{
			Valid:  true,
			String: parentMerchantID,
		},
	}
	validTransfer := &transferModel.Transfer{
		UUID: uuid.New(),
	}

	testCases := []struct {
		name        string
		wantErr     bool
		setupMock   func()
		expectError error
	}{
		{
			name:    "ERROR: Get payment by ID",
			wantErr: true,
			setupMock: func() {
				paymentRepo.On("GetPaymentById", constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			expectError: pkgErr.New(httpResponse.HttpErrDatabase, constant.ErrSomeErrorForUnitTest),
		},
		{
			name:    "ERROR: Get payment by ID not found data",
			wantErr: true,
			setupMock: func() {
				paymentRepo.On("GetPaymentById", constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(nil, nil)
			},
			expectError: pkgErr.New(httpResponse.HttpErrUnprocessableContent, constant.ErrPaymentNotFound),
		},
		{
			name:    "ERROR: Get merchant by ID",
			wantErr: true,
			setupMock: func() {
				paymentRepo.On("GetPaymentById", constant.ValueCtxMockType(), constant.StringMockType()).
					Return(validPayment, nil)
				merchantRepo.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			expectError: pkgErr.New(httpResponse.HttpErrDatabase, constant.ErrSomeErrorForUnitTest),
		},
		{
			name:    "ERROR: Get merchant by ID not found data",
			wantErr: true,
			setupMock: func() {
				merchantRepo.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(nil, nil)
			},
			expectError: pkgErr.New(httpResponse.HttpErrUnprocessableContent, constant.ErrMerchantNotFound),
		},
		{
			name:    "ERROR: Transfer service",
			wantErr: true,
			setupMock: func() {
				merchantRepo.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).
					Return(validMerchant, nil)
				transferSvc.On("Transfer", constant.ValueCtxMockType(), mock.AnythingOfType("*transfer.TransferRequest")).
					Once().Return(nil, pkgErr.New(httpResponse.HttpErrInternal, constant.ErrSomeErrorForUnitTest))
			},
			expectError: pkgErr.New(httpResponse.HttpErrInternal, constant.ErrSomeErrorForUnitTest),
		},
		{
			name:    "ERROR: Update payment data",
			wantErr: true,
			setupMock: func() {
				transferSvc.On("Transfer", constant.ValueCtxMockType(), mock.AnythingOfType("*transfer.TransferRequest")).
					Return(validTransfer, nil)

				paymentRepo.On("UpdatePaymentData", constant.ValueCtxMockType(), mock.AnythingOfType("*paymentModel.PaymentDTO")).
					Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			expectError: pkgErr.New(httpResponse.HttpErrDatabase, constant.ErrSomeErrorForUnitTest),
		},
		{
			name:    "SUCCESS",
			wantErr: false,
			setupMock: func() {
				transferSvc.On("UpdateTransferStatus", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(), constant.StringMockType(), mock.Anything).
					Return(nil)

				ledgerSvc.On("UpdateTransaction", constant.ValueCtxMockType(), mock.Anything).
					Return(nil)

				paymentRepo.On("UpdatePaymentData", constant.ValueCtxMockType(), mock.AnythingOfType("*paymentModel.PaymentDTO")).
					Return(nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()
			paymentSvc := New(paymentRepo, logger, nil, nil, merchantRepo, nil, nil,
				WithTransferService(transferSvc),
				WithLedgerService(ledgerSvc),
			)

			ctx := context.Background()
			err := paymentSvc.ProcessSplitRoute(ctx, paymentID)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tc.expectError, err)
			} else {
				assert.NoError(t, err)
			}

			paymentRepo.AssertExpectations(t)
			merchantRepo.AssertExpectations(t)
			transferSvc.AssertExpectations(t)
		})
	}
}
