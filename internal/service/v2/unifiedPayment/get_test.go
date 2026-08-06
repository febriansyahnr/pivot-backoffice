package unifiedPaymentService_test

import (
	"context"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcardCoreProcessor"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v2/unifiedPayment"
	logMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	repositoryMock "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/google/uuid"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetSessionDetail(t *testing.T) {
	cfg := &config.Config{
		Environment: "test",
	}
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})
	paymentRepo := repositoryMock.NewIPaymentRepository(t)
	paymentMethodRepo := repositoryMock.NewIPaymentMethodRepository(t)
	accountTrxRepo := repositoryMock.NewIAccountTransactionRepository(t)

	merchantId := uuid.NewString()

	testCases := []struct {
		name      string
		wantErr   bool
		setupMock func()
	}{
		{
			name:    "ERROR: Got error database on GetPaymentById",
			wantErr: true,
			setupMock: func() {
				paymentRepo.On("GetPaymentById",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: Payment is not found",
			wantErr: true,
			setupMock: func() {
				paymentRepo.On("GetPaymentById",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(nil, nil)
			},
		},
		{
			name:    "ERROR: Merchant is not match",
			wantErr: true,
			setupMock: func() {
				paymentRepo.On("GetPaymentById",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&paymentModel.Payment{}, nil)
			},
		},
		{
			name:    "ERROR: Got error database on find charge by payment session ID",
			wantErr: true,
			setupMock: func() {
				paymentRepo.On("GetPaymentById",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Return(&paymentModel.Payment{MerchantID: merchantId}, nil)

				accountTrxRepo.On("FindByReference",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "SUCCESS",
			wantErr: false,
			setupMock: func() {
				accountTrxRepo.On("FindByReference",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Return(nil, nil)
			},
		},
		{
			name:    "SUCCESS: With ledgerID on inquiry request",
			wantErr: false,
			setupMock: func() {
				accountTrxRepo.On("FindByReference",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Return(&orchestrator_model.AccountTransaction{UUID: uuid.New()}, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			svc := New(cfg, log, paymentRepo, paymentMethodRepo, accountTrxRepo)
			_, err := svc.GetSessionDetail(context.Background(), &unifiedPaymentModel.GetUnifiedPaymentSessionRequest{
				PaymentSessionID: uuid.NewString(),
				MerchantID:       merchantId,
			})
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetSessionList(t *testing.T) {
	cfg := &config.Config{
		Environment: "test",
	}
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})
	paymentRepo := repositoryMock.NewIPaymentRepository(t)
	paymentMethodRepo := repositoryMock.NewIPaymentMethodRepository(t)
	accountTrxRepo := repositoryMock.NewIAccountTransactionRepository(t)

	resp := &commonModel.PaginationResponse{
		Data: []*paymentModel.Payment{{UUID: uuid.NewString()}},
		Meta: commonModel.Meta{},
	}

	testCases := []struct {
		name      string
		wantErr   bool
		setupMock func()
	}{
		{
			name:    "ERROR: Got error database on GetList",
			wantErr: true,
			setupMock: func() {
				paymentRepo.On("GetList",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*paymentModel.GetListFilterRequest"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: Got error database on get charge by payment session ID",
			wantErr: true,
			setupMock: func() {
				paymentRepo.On("GetList",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*paymentModel.GetListFilterRequest"),
				).Return(resp, nil)

				accountTrxRepo.On("FindByReference",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "SUCCESS",
			wantErr: false,
			setupMock: func() {
				accountTrxRepo.On("FindByReference",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Return(nil, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			svc := New(cfg, log, paymentRepo, paymentMethodRepo, accountTrxRepo)
			_, err := svc.GetSessionList(context.Background(), &paymentModel.GetListFilterRequest{})
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetCardBinDetail(t *testing.T) {
	logger := logMock.NewILogger(t)
	paymentMethodRepo := repositoryMock.NewIPaymentMethodRepository(t)
	cardProcessorRepo := repositoryMock.NewICreditcardCoreProcessorRepository(t)

	service := New(nil, logger, nil, paymentMethodRepo, nil, WithCreditCardCoreProcessorRepo(cardProcessorRepo))

	tests := []struct {
		name       string
		setupMock  func()
		wantError  error
		wantResult *unifiedPaymentModel.GetBinDetailResponse
	}{
		{
			name: "ERROR:Get active payment method", // NOSONAR
			setupMock: func() {
				paymentMethodRepo.On("GetActivePaymentMethodByRequest", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
				logger.On("Error", mock.Anything, "Failed while get active payment method card", mock.Anything).Once().Return()
			},
			wantError: pkgErr.New(response.HttpErrDatabase, c.ErrInternalServerForUser),
		},
		{
			name: "ERROR:Card payment method not found", // NOSONAR
			setupMock: func() {
				paymentMethodRepo.On("GetActivePaymentMethodByRequest", mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
			wantError: c.NewErrStringRequest(response.HttpErrForbidden, c.ErrCodeForbiddenAccess, "Merchant not authorized for BIN lookup"),
		},
		{
			name: "ERROR:Get BIN detail by BIN number", // NOSONAR
			setupMock: func() {
				paymentMethodRepo.On("GetActivePaymentMethodByRequest", mock.Anything, mock.Anything).Return(&paymentModel.PaymentMethodWithPivot{}, nil)
				cardProcessorRepo.On("GetBinDetailByBinNumber", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name: "ERROR:BIN detail not found", // NOSONAR
			setupMock: func() {
				cardProcessorRepo.On("GetBinDetailByBinNumber", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
			wantError: c.NewErrStringRequest(response.HttpErrNotFound, c.ErrCodeDataNotFound, "BIN not found"),
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				cardProcessorRepo.On("GetBinDetailByBinNumber", mock.Anything, mock.Anything, mock.Anything).Return(&creditcardCoreProcessorModel.GetBinDetailResponse{
					BinNumber:     "123456",  // NOSONAR
					CardType:      "CREDIT",  // NOSONAR
					CardBrand:     "VISA",    // NOSONAR
					CardLevel:     "CLASSIC", // NOSONAR
					IssuerName:    "BCA",     // NOSONAR
					IssuerCountry: "ID",      // NOSONAR
					Currency:      "IDR",     // NOSONAR
				}, nil)
			},
			wantResult: &unifiedPaymentModel.GetBinDetailResponse{
				BIN:       "123456",  // NOSONAR
				CardType:  "CREDIT",  // NOSONAR
				Principal: "VISA",    // NOSONAR
				CardLevel: "CLASSIC", // NOSONAR
				Issuer:    "BCA",     // NOSONAR
				Country:   "ID",      // NOSONAR
				Currency:  "IDR",     // NOSONAR
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := service.GetCardBinDetail(t.Context(), unifiedPaymentModel.GetBinDetailRequest{})
			assert.Equal(t, test.wantError, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
