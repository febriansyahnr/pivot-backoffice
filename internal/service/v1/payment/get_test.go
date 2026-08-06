package paymentService

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	constant "github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	accountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	cardFundedPayoutModel "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	refundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/refund"
	splitRoutingPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/splitRoutingPayment"
	statusHistoryModel "github.com/paper-indonesia/pivot-backoffice/internal/model/statusHistory"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/transfer"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
)

func TestPaymentServiceGetPaymentById(t *testing.T) {
	amount := decimal.NewFromFloat(1000)
	now := time.Now()
	fee := decimal.NewFromFloat(1000)
	discount := decimal.NewFromFloat(1000)
	totalAmount := decimal.NewFromFloat(1000)
	//metadata := "{\"testing\":\"testing\"}"
	description := "description"
	metadataMap := map[string]any{"testing": "testing", "maxAmount": map[string]any{"value": "100000", "currency": "IDR"}, "minAmount": map[string]any{"value": "1000", "currency": "IDR"}}
	email := "test@gmail.com"
	phone := "1234123112"

	paymentUuid := uuid.NewString()
	paymentReferenceId := "123"
	payment := &paymentModel.Payment{
		UUID:            paymentUuid,
		ReferenceID:     &paymentReferenceId,
		MerchantID:      "merchant-id",
		CustomerID:      "customer-id",
		PaymentMethodID: "payment-method-id",
		Currency:        "IDR",
		Amount:          amount,
		Fee:             &fee,
		Discount:        &discount,
		TotalAmount:     amount,
		Status:          "pending",
		Metadata:        &metadataMap,
		CreatedAt:       now,
		UpdatedAt:       now,
		PaymentMethod: paymentModel.PaymentMethod{
			Type:     "VA",
			Name:     "VA Permata",
			Acquirer: "Permata",
		},
	}
	paymentInvalidMetadata := &paymentModel.Payment{
		UUID:            paymentUuid,
		MerchantID:      "merchant-id",
		CustomerID:      "customer-id",
		PaymentMethodID: "payment-method-id",
		Currency:        "IDR",
		Amount:          amount,
		Fee:             &fee,
		Discount:        &discount,
		TotalAmount:     amount,
		Status:          "pending",
		Metadata:        &map[string]any{"invalidKey": make(chan int)},
		CreatedAt:       now,
		UpdatedAt:       now,
		PaymentMethod: paymentModel.PaymentMethod{
			Type:     "VA",
			Name:     "VA Permata",
			Acquirer: "Permata",
		},
	}

	paymentItem := &paymentModel.PaymentItem{
		UUID:        "uuid-uuid-uuid",
		PaymentID:   "payment-id",
		Name:        "name",
		Description: description,
		Qty:         1,
		Currency:    "IDR",
		Amount:      amount,
		TotalAmount: totalAmount,
		Metadata:    &metadataMap,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	paymentItems := []*paymentModel.PaymentItem{paymentItem}

	paymentMethod := &paymentModel.PaymentMethod{
		UUID:      uuid.NewString(),
		Type:      paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
		Category:  paymentConstant.PAYMENT_METHOD_CATEGORY_DISBURSEMENT_TOPUP,
		Name:      "VA Permata",
		Acquirer:  constant.BANK_ACQUIRER_PERMATA,
		CreatedAt: util.TimeNow,
		UpdatedAt: util.TimeNow,
	}

	customer := &customerModel.Customer{
		UUID:        "123",
		FirstName:   "John Doe",
		Email:       email,
		PhoneNumber: phone,
	}

	testCases := []struct {
		name       string
		id         string
		setupMocks func(paymentRepo *repositoryMocks.IPaymentRepository, customerRepo *repositoryMocks.ICustomerRepository, paymentMethodRepo *repositoryMocks.IPaymentMethodRepository, accountTransactionRepo *repositoryMocks.IAccountTransactionRepository)
		wantErr    bool
	}{
		{
			name: "SUCCESS: Get Payment By ID",
			id:   paymentUuid,
			setupMocks: func(paymentRepo *repositoryMocks.IPaymentRepository, customerRepo *repositoryMocks.ICustomerRepository, paymentMethodRepo *repositoryMocks.IPaymentMethodRepository, accountTransactionRepo *repositoryMocks.IAccountTransactionRepository) {
				paymentRepo.
					On(
						"GetPaymentById",
						mock.Anything,
						mock.Anything).
					Return(payment, nil)
				paymentMethodRepo.
					On(
						"GetPaymentMethodById",
						mock.Anything,
						mock.Anything).
					Return(paymentMethod, nil)
				customerRepo.
					On(
						"FindCustomerById",
						mock.Anything,
						mock.Anything).
					Return(customer, nil)
				paymentRepo.
					On(
						"GetPaymentItemsByPaymentId",
						mock.Anything,
						mock.Anything).
					Return(paymentItems, nil)
				// Mock accountTransactionRepo to return nil (FDS data not found, which is OK)
				accountTransactionRepo.
					On(
						"FindByReference",
						mock.Anything,
						mock.Anything,
						mock.Anything).
					Return(nil, nil)
			},
			wantErr: false,
		},
		{
			name: "FAILED: Get Payment By ID - Payment Not Found",
			id:   paymentUuid,
			setupMocks: func(paymentRepo *repositoryMocks.IPaymentRepository, customerRepo *repositoryMocks.ICustomerRepository, paymentMethodRepo *repositoryMocks.IPaymentMethodRepository, accountTransactionRepo *repositoryMocks.IAccountTransactionRepository) {
				paymentRepo.
					On(
						"GetPaymentById",
						mock.Anything,
						mock.Anything).
					Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "FAILED: Get Payment By ID - Failed Payment",
			id:   paymentUuid,
			setupMocks: func(paymentRepo *repositoryMocks.IPaymentRepository, customerRepo *repositoryMocks.ICustomerRepository, paymentMethodRepo *repositoryMocks.IPaymentMethodRepository, accountTransactionRepo *repositoryMocks.IAccountTransactionRepository) {
				paymentRepo.
					On(
						"GetPaymentById",
						mock.Anything,
						mock.Anything).
					Return(nil, assert.AnError)

			},
			wantErr: true,
		},
		{
			name: "FAILED: Get Payment By ID - Payment Method Not Found",
			id:   paymentUuid,
			setupMocks: func(paymentRepo *repositoryMocks.IPaymentRepository, customerRepo *repositoryMocks.ICustomerRepository, paymentMethodRepo *repositoryMocks.IPaymentMethodRepository, accountTransactionRepo *repositoryMocks.IAccountTransactionRepository) {
				paymentRepo.
					On(
						"GetPaymentById",
						mock.Anything,
						mock.Anything).
					Return(payment, nil)
				paymentMethodRepo.
					On(
						"GetPaymentMethodById",
						mock.Anything,
						mock.Anything).
					Return(nil, assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "FAILED: Get Payment By ID - Customer Not Found",
			id:   paymentUuid,
			setupMocks: func(paymentRepo *repositoryMocks.IPaymentRepository, customerRepo *repositoryMocks.ICustomerRepository, paymentMethodRepo *repositoryMocks.IPaymentMethodRepository, accountTransactionRepo *repositoryMocks.IAccountTransactionRepository) {
				paymentRepo.
					On(
						"GetPaymentById",
						mock.Anything,
						mock.Anything).
					Return(payment, nil)
				paymentMethodRepo.
					On(
						"GetPaymentMethodById",
						mock.Anything,
						mock.Anything).
					Return(paymentMethod, nil)
				customerRepo.
					On(
						"FindCustomerById",
						mock.Anything,
						mock.Anything).
					Return(nil, assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "FAILED: Get Payment By ID - Payment Items Not Found",
			id:   paymentUuid,
			setupMocks: func(paymentRepo *repositoryMocks.IPaymentRepository, customerRepo *repositoryMocks.ICustomerRepository, paymentMethodRepo *repositoryMocks.IPaymentMethodRepository, accountTransactionRepo *repositoryMocks.IAccountTransactionRepository) {
				paymentRepo.
					On(
						"GetPaymentById",
						mock.Anything,
						mock.Anything).
					Return(payment, nil)
				paymentMethodRepo.
					On(
						"GetPaymentMethodById",
						mock.Anything,
						mock.Anything).
					Return(paymentMethod, nil)
				customerRepo.
					On(
						"FindCustomerById",
						mock.Anything,
						mock.Anything).
					Return(customer, nil)
				// Mock accountTransactionRepo to return nil (FDS data not found, which is OK)
				accountTransactionRepo.
					On(
						"FindByReference",
						mock.Anything,
						mock.Anything,
						mock.Anything).
					Return(nil, nil)
				paymentRepo.
					On(
						"GetPaymentItemsByPaymentId",
						mock.Anything,
						mock.Anything).
					Return(nil, assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "FAILED: Get Payment By ID - Marshal Metadata Error",
			id:   paymentUuid,
			setupMocks: func(paymentRepo *repositoryMocks.IPaymentRepository, customerRepo *repositoryMocks.ICustomerRepository, paymentMethodRepo *repositoryMocks.IPaymentMethodRepository, accountTransactionRepo *repositoryMocks.IAccountTransactionRepository) {
				payment.Metadata = nil
				paymentRepo.
					On(
						"GetPaymentById",
						mock.Anything,
						mock.Anything).
					Return(paymentInvalidMetadata, nil)
				paymentMethodRepo.
					On(
						"GetPaymentMethodById",
						mock.Anything,
						mock.Anything).
					Return(paymentMethod, nil)
				customerRepo.
					On(
						"FindCustomerById",
						mock.Anything,
						mock.Anything).
					Return(customer, nil)
				// Mock accountTransactionRepo to return nil (FDS data not found, which is OK)
				accountTransactionRepo.
					On(
						"FindByReference",
						mock.Anything,
						mock.Anything,
						mock.Anything).
					Return(nil, nil)
				paymentRepo.
					On(
						"GetPaymentItemsByPaymentId",
						mock.Anything,
						mock.Anything).
					Return(paymentItems, nil)

			},
			wantErr: true,
		},
		{
			name: "FAILED: Get Payment By ID - Payment ReferenceID is nil",
			id:   paymentUuid,
			setupMocks: func(paymentRepo *repositoryMocks.IPaymentRepository, customerRepo *repositoryMocks.ICustomerRepository, paymentMethodRepo *repositoryMocks.IPaymentMethodRepository, accountTransactionRepo *repositoryMocks.IAccountTransactionRepository) {
				paymentRepo.
					On(
						"GetPaymentById",
						mock.Anything,
						mock.Anything).
					Return(payment, nil)

				// set payment referenceID to nil
				payment.ReferenceID = nil

				paymentMethodRepo.
					On(
						"GetPaymentMethodById",
						mock.Anything,
						mock.Anything).
					Return(paymentMethod, nil)
				customerRepo.
					On(
						"FindCustomerById",
						mock.Anything,
						mock.Anything).
					Return(customer, nil)
				// Mock accountTransactionRepo to return nil (FDS data not found, which is OK)
				accountTransactionRepo.
					On(
						"FindByReference",
						mock.Anything,
						mock.Anything,
						mock.Anything).
					Return(nil, nil)
				paymentRepo.
					On(
						"GetPaymentItemsByPaymentId",
						mock.Anything,
						mock.Anything).
					Return(paymentItems, nil)
			},
			wantErr: false,
		},
		{
			name: "FAILED: Get Payment By ID - Merchant Not Found",
			id:   paymentUuid,
			setupMocks: func(paymentRepo *repositoryMocks.IPaymentRepository, customerRepo *repositoryMocks.ICustomerRepository, paymentMethodRepo *repositoryMocks.IPaymentMethodRepository, accountTransactionRepo *repositoryMocks.IAccountTransactionRepository) {
				paymentResponse := payment
				paymentResponse.MerchantID = "invalid merchantID"

				paymentRepo.
					On(
						"GetPaymentById",
						mock.Anything,
						mock.Anything).
					Return(paymentResponse, nil)
				paymentMethodRepo.
					On(
						"GetPaymentMethodById",
						mock.Anything,
						mock.Anything).
					Return(paymentMethod, nil)
				customerRepo.
					On(
						"FindCustomerById",
						mock.Anything,
						mock.Anything).
					Return(customer, nil)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockPaymentRepo := repositoryMocks.NewIPaymentRepository(t)
			mockCustomerRepo := repositoryMocks.NewICustomerRepository(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockSnapCore := repositoryMocks.NewISnapCoreRepository(t)
			mockMerchantRepo := repositoryMocks.NewIMerchantRepository(t)
			mockPaymentMethodRepo := repositoryMocks.NewIPaymentMethodRepository(t)
			mockAccountTransactionRepo := repositoryMocks.NewIAccountTransactionRepository(t)

			tc.setupMocks(mockPaymentRepo, mockCustomerRepo, mockPaymentMethodRepo, mockAccountTransactionRepo)

			merchantInfo := &merchantModel.MerchantAuthTokenClaims{
				MerchantId: "merchant-id",
			}

			ctx := context.Background()
			ctx = context.WithValue(ctx, constant.CtxMerchantInfo, merchantInfo)

			svc := New(mockPaymentRepo, mockLogger, mockSnapCore, mockCustomerRepo, mockMerchantRepo, mockPaymentMethodRepo, nil, WithAccountTransactionRepository(mockAccountTransactionRepo))
			_, err := svc.FindPaymentById(ctx, tc.id, merchantInfo.MerchantId)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockPaymentRepo.AssertExpectations(t)
			mockCustomerRepo.AssertExpectations(t)
			mockSnapCore.AssertExpectations(t)
			mockMerchantRepo.AssertExpectations(t)
			mockAccountTransactionRepo.AssertExpectations(t)
		})
	}
}

func TestGetTotalPaymentBalance(t *testing.T) {
	var (
		ctx               = context.Background()
		invalidMerchantID = uuid.New()
		validMerchantID   = uuid.New()
		validUSMerchantID = uuid.New()

		mockAccountRepo     repositoryMocks.IAccountRepository
		mockOrchestratorSvc serviceMocks.IOrchestratorService
		paymentService      = PaymentService{
			accountRepo:     &mockAccountRepo,
			orchestratorSvc: &mockOrchestratorSvc,
		}
	)
	testCases := []struct {
		name      string
		payload   uuid.UUID
		callMock  func()
		want      *commonModel.Amount
		wantErr   error
		shouldErr bool
	}{
		{
			name:    "when failed to get merchant payment account, then should return error",
			payload: invalidMerchantID,
			callMock: func() {
				mockAccountRepo.On("FindMerchantAccountByName", mock.Anything, invalidMerchantID, constant.TypePayment).
					Return(&accountModel.Account{}, errors.New("merchant account not found")).Once()
			},
			wantErr:   errors.New("merchant account not found"),
			shouldErr: true,
		},
		{
			name:    "when failed to get merchant payment balance, then should return error",
			payload: validMerchantID,
			callMock: func() {
				mockAccountRepo.On("FindMerchantAccountByName", mock.Anything, validMerchantID, constant.TypePayment).
					Return(&accountModel.Account{
						UUID:     uuid.New(),
						Name:     constant.TypePayment,
						Currency: constant.CurrencyIDR,
					}, nil).
					Once()
				mockOrchestratorSvc.On("GetAvailableMerchantBalance", mock.Anything, validMerchantID.String(), constant.TypePayment).
					Return(0.0, errors.New("merchant balance not found")).
					Once()

			},
			wantErr:   errors.New("merchant balance not found"),
			shouldErr: true,
		},
		{
			name:    "when merchant payment balance was found, then should return the available amount",
			payload: validMerchantID,
			callMock: func() {
				mockAccountRepo.On("FindMerchantAccountByName", mock.Anything, validMerchantID, constant.TypePayment).
					Return(&accountModel.Account{
						UUID:     uuid.New(),
						Name:     constant.TypePayment,
						Currency: constant.CurrencyIDR,
					}, nil).
					Once()
				mockOrchestratorSvc.On("GetAvailableMerchantBalance", mock.Anything, validMerchantID.String(), constant.TypePayment).
					Return(1945000.0, nil).
					Once()

			},
			want: &commonModel.Amount{
				Currency: constant.CurrencyIDR,
				Value:    "1945000.00",
			},
		},
		{
			name:    "when merchant payment balance was found, then should return the available amount with USD currency",
			payload: validUSMerchantID,
			callMock: func() {
				mockAccountRepo.On("FindMerchantAccountByName", mock.Anything, validUSMerchantID, constant.TypePayment).
					Return(&accountModel.Account{
						UUID:     uuid.New(),
						Name:     constant.TypePayment,
						Currency: "USD",
					}, nil).
					Once()
				mockOrchestratorSvc.On("GetAvailableMerchantBalance", mock.Anything, validUSMerchantID.String(), constant.TypePayment).
					Return(1011000.0, nil).
					Once()

			},
			want: &commonModel.Amount{
				Currency: "USD",
				Value:    "1011000.00",
			},
		},
		{
			name:    "when merchant payment balance was nil, should not return error",
			payload: validUSMerchantID,
			callMock: func() {
				mockAccountRepo.On("FindMerchantAccountByName", mock.Anything, validUSMerchantID, constant.TypePayment).
					Return(nil, nil).
					Once()
			},
			want: &commonModel.Amount{
				Currency: "",
				Value:    "0.00",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.callMock()

			amount, err := paymentService.GetTotalPaymentBalance(ctx, tc.payload)

			if tc.shouldErr {
				assert.NotNil(t, err)
				assert.Equal(t, tc.wantErr, err)
				return
			}

			assert.Nil(t, err)
			assert.Equal(t, tc.want.Currency, amount.Currency)
			assert.Equal(t, tc.want.Value, amount.Value)
		})
	}
}

func TestGetPaymentHistoryDetail(t *testing.T) {
	var (
		mockLogger, _              = loggerMocks.NewZapLogger(loggerMocks.Config{})
		mockPaymentRepo            = repositoryMocks.NewIPaymentRepository(t)
		mockAccountTransactionRepo = repositoryMocks.NewIAccountTransactionRepository(t)
		mockCustomerRepo           = repositoryMocks.NewICustomerRepository(t)
		mockStatusHistoriesRepo    = repositoryMocks.NewIStatusHistoriesRepository(t)
		transferSvc                = serviceMocks.NewITransferService(t)
		refundSvc                  = serviceMocks.NewIRefundService(t)
	)

	now := time.Now()
	fiveMinLater := now.Add(5 * time.Minute)

	// Create valid UUIDs for transfers
	transfer1UUID := uuid.New()
	transfer2UUID := uuid.New()

	testCases := []struct {
		name      string
		option    paymentModel.PaymentHistoryDetailOption
		setupMock func()
		want      *paymentModel.PaymentHistoryDetailResponse
		wantErr   error
	}{
		{
			name: "when payment not found, should return error",
			option: paymentModel.PaymentHistoryDetailOption{
				PaymentID:  "invalid-payment-id",
				MerchantID: "valid-merchant-id",
			},
			setupMock: func() {
				mockPaymentRepo.On("GetPaymentByIdAndMerchantId", mock.Anything, "invalid-payment-id", "valid-merchant-id").
					Return(nil, errors.New("payment not found")).
					Once()
			},
			wantErr: errors.New("payment not found"),
		},
		{
			name: "when payment is nil, should return error not found",
			option: paymentModel.PaymentHistoryDetailOption{
				PaymentID:  "valid-payment-id",
				MerchantID: "valid-merchant-id",
			},
			setupMock: func() {
				mockPaymentRepo.On("GetPaymentByIdAndMerchantId", mock.Anything, "valid-payment-id", "valid-merchant-id").
					Return(nil, nil).
					Once()
			},
			wantErr: pkgErrors.New(response.HttpErrNotFound, errors.New("payment not found")),
		},
		{
			name: "when account transaction not found, should return error",
			option: paymentModel.PaymentHistoryDetailOption{
				PaymentID:  "valid-payment-id",
				MerchantID: "valid-merchant-id",
			},
			setupMock: func() {
				// Mock payment retrieval
				mockPaymentRepo.On("GetPaymentByIdAndMerchantId", mock.Anything, "valid-payment-id", "valid-merchant-id").
					Return(&paymentModel.Payment{
						UUID:       "valid-payment-uuid",
						MerchantID: "valid-merchant-id",
						CustomerID: "customer-id",
						PaymentMethod: paymentModel.PaymentMethod{
							Type:     paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
							BankName: util.ValueToPtr("valid-bank"),
						},
						Currency:                 "IDR",
						Amount:                   decimal.New(100000, 10),
						Status:                   "success",
						CreatedAt:                time.Now(),
						UpdatedAt:                time.Now(),
						ExpiredAt:                nil,
						ProcessorReferenceNumber: util.ValueToPtr("proc-ref"),
					}, nil).Once()

				// Mock failure in GetChargeList
				mockPaymentRepo.On("GetChargeList", mock.Anything, &unifiedPaymentModel.FilterChargeRequest{
					MerchantID:       "valid-merchant-id",
					PaymentSessionID: "valid-payment-uuid",
					Page:             1,
					PerPage:          1000,
				}).Return(nil, errors.New("charge list not found")).Once()
			},
			wantErr: errors.New("charge list not found"),
		},
		{
			name: "when payment and transaction found, should return payment history detail",
			option: paymentModel.PaymentHistoryDetailOption{
				PaymentID:  "valid-payment-id",
				MerchantID: "valid-merchant-id",
			},
			setupMock: func() {
				// Mock payment retrieval
				mockPaymentRepo.On("GetPaymentByIdAndMerchantId", mock.Anything, "valid-payment-id", "valid-merchant-id").
					Return(&paymentModel.Payment{
						UUID:       "valid-payment-uuid",
						MerchantID: "valid-merchant-id",
						CustomerID: "customer-id",
						PaymentMethod: paymentModel.PaymentMethod{
							Type:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
							BankName: util.ValueToPtr("BRUH"),
						},
						Metadata: &map[string]interface{}{
							"snapCore": map[string]interface{}{
								"Number":         "1234567890",
								"AccountName":    "Test VA",
								"IsClosedAmount": false,
								"IsSingleUse":    false,
							},
						},
						Currency:                 "IDR",
						Amount:                   decimal.NewFromInt(100000),
						Status:                   "success",
						CreatedAt:                now,
						UpdatedAt:                fiveMinLater,
						ExpiredAt:                nil,
						ProcessorReferenceNumber: util.ValueToPtr("test-ref-number"),
						PaymentURL:               "http://payment-url",
						CreatedFrom:              constant.DisbursementCreatedFromMerchantPortal,
					}, nil).Once()

				// Mock GetChargeList
				mockPaymentRepo.On("GetChargeList", mock.Anything, &unifiedPaymentModel.FilterChargeRequest{
					MerchantID:       "valid-merchant-id",
					PaymentSessionID: "valid-payment-uuid",
					Page:             1,
					PerPage:          1000,
				}).Return(&commonModel.PaginationResponse{
					Data: []*unifiedPaymentModel.ChargeResponse{
						{
							ID:                              constant.EmptyUUID,
							PaymentSessionID:                "valid-payment-uuid",
							PaymentSessionClientReferenceID: "",
							Amount: unifiedPaymentModel.Amount{
								Currency: "IDR",
								Value:    100000,
							},
							Status: constant.StatusSuccess,
							AuthorizedAmount: &unifiedPaymentModel.Amount{
								Currency: "IDR",
								Value:    100000,
							},
							IsCaptured: true,
							CapturedAmount: &unifiedPaymentModel.Amount{
								Currency: "IDR",
								Value:    100000,
							},
							ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{},
						},
					},
				}, nil).Once()

				// Mock account transaction retrieval for FDS data
				mockAccountTransactionRepo.On("FindByReference", mock.Anything, "valid-payment-uuid", constant.TypePayment).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						Credit:   100000,
						Currency: "IDR",
						Status:   constant.StatusSuccess,
					}, nil).Once()

				// mock status history
				mockStatusHistoriesRepo.On("GetByReference", mock.Anything, constant.TypePayment, "valid-payment-uuid").
					Return([]*statusHistoryModel.StatusHistory{}, nil).Once()

				// mock customer info
				mockCustomerRepo.On("GetCustomerById", mock.Anything, "customer-id", "valid-merchant-id").Return(nil, nil).Once()

				refundSvc.On("GetExistingRefundList", mock.Anything, mock.Anything).Return([]refundModel.RefundResponse{}, nil).Once()

			},
			want: &paymentModel.PaymentHistoryDetailResponse{
				UUID:                  "valid-payment-uuid",
				MerchantID:            "valid-merchant-id",
				CustomerID:            "customer-id",
				PaymentMethod:         paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				PaymentMethodCategory: "OPEN_STATIC",
				Amount: commonModel.Amount{
					Currency: "IDR",
					Value:    "100000.00",
				},
				AmountPaid: commonModel.Amount{
					Currency: "IDR",
					Value:    "100000.00",
				},
				BankReferenceID:    "-",
				Channel:            "BRUH",
				ProcessorRefNumber: "test-ref-number",
				Status:             "success",
				PaymentLink:        "http://payment-url",
				CreatedAt:          now,
				UpdatedAt:          fiveMinLater,
				TypeDetail: paymentModel.PaymentTypeDetail{
					VirtualAccountName:   util.ValueToPtr("Test VA"),
					VirtualAccountNumber: util.ValueToPtr("1234567890"),
				},
				TotalSplitAmount: commonModel.Amount{
					Currency: "IDR",
					Value:    "0.00",
				},
				Fee: commonModel.Amount{
					Currency: "IDR",
					Value:    "0.00",
				},
				SettledAmount: commonModel.Amount{
					Currency: "IDR",
					Value:    "100000.00",
				},
				SettlementModel: constant.PaymentMethodChannelTypeAggregator,
				Charges: []*unifiedPaymentModel.ChargeResponse{
					{
						ID:                              constant.EmptyUUID,
						PaymentSessionID:                "valid-payment-uuid",
						PaymentSessionClientReferenceID: "",
						Amount: unifiedPaymentModel.Amount{
							Currency: "IDR",
							Value:    100000,
						},
						Status: constant.StatusSuccess,
						AuthorizedAmount: &unifiedPaymentModel.Amount{
							Currency: "IDR",
							Value:    100000,
						},
						IsCaptured: true,
						CapturedAmount: &unifiedPaymentModel.Amount{
							Currency: "IDR",
							Value:    100000,
						},
						ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{},
					},
				},
			},
		},
		{
			name: "when CC payment and transaction found, should return payment history detail",
			option: paymentModel.PaymentHistoryDetailOption{
				PaymentID:  "valid-payment-id",
				MerchantID: "valid-merchant-id",
			},
			setupMock: func() {
				// Mock payment retrieval
				mockPaymentRepo.On("GetPaymentByIdAndMerchantId", mock.Anything, "valid-payment-id", "valid-merchant-id").
					Return(&paymentModel.Payment{
						UUID:       "valid-payment-uuid",
						MerchantID: "valid-merchant-id",
						CustomerID: "customer-id",
						PaymentMethod: paymentModel.PaymentMethod{
							Type:     paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
							BankName: util.ValueToPtr("BRUH"),
						},
						Metadata: &map[string]interface{}{
							"cardData": map[string]interface{}{
								"cardIssuing": "VISA",
								"last4Digit":  "1234",
								"cardBrand":   "VISA",
							},
						},
						Currency:                 "IDR",
						Amount:                   decimal.NewFromInt(100000),
						Status:                   "success",
						CreatedAt:                now,
						UpdatedAt:                fiveMinLater,
						ExpiredAt:                nil,
						ProcessorReferenceNumber: util.ValueToPtr("test-ref-number"),
					}, nil).Once()

				// Mock GetChargeList
				mockPaymentRepo.On("GetChargeList", mock.Anything, &unifiedPaymentModel.FilterChargeRequest{
					MerchantID:       "valid-merchant-id",
					PaymentSessionID: "valid-payment-uuid",
					Page:             1,
					PerPage:          1000,
				}).Return(&commonModel.PaginationResponse{
					Data: []*unifiedPaymentModel.ChargeResponse{
						{
							ID:                              constant.EmptyUUID,
							PaymentSessionID:                "valid-payment-uuid",
							PaymentSessionClientReferenceID: "",
							Amount: unifiedPaymentModel.Amount{
								Currency: "IDR",
								Value:    100000,
							},
							Status: constant.StatusSuccess,
							AuthorizedAmount: &unifiedPaymentModel.Amount{
								Currency: "IDR",
								Value:    100000,
							},
							IsCaptured: true,
							CapturedAmount: &unifiedPaymentModel.Amount{
								Currency: "IDR",
								Value:    100000,
							},
							ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{},
						},
					},
				}, nil).Once()

				// Mock account transaction retrieval
				mockAccountTransactionRepo.On("FindByReference", mock.Anything, "valid-payment-uuid", constant.TypePayment).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						Credit:   100000,
						Currency: "IDR",
						Status:   constant.StatusSuccess,
					}, nil).Once()

				// mock status history
				mockStatusHistoriesRepo.On("GetByReference", mock.Anything, constant.TypePayment, "valid-payment-uuid").
					Return([]*statusHistoryModel.StatusHistory{}, nil).Once()

				mockCustomerRepo.On("GetCustomerById", mock.Anything, "customer-id", "valid-merchant-id").Return(nil, nil).Once()

				refundSvc.On("GetExistingRefundList", mock.Anything, mock.Anything).Return([]refundModel.RefundResponse{}, nil).Once()

			},
			want: &paymentModel.PaymentHistoryDetailResponse{
				UUID:                  "valid-payment-uuid",
				MerchantID:            "valid-merchant-id",
				CustomerID:            "customer-id",
				PaymentMethod:         paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				PaymentMethodCategory: "",
				Amount: commonModel.Amount{
					Currency: "IDR",
					Value:    "100000.00",
				},
				AmountPaid: commonModel.Amount{
					Currency: "IDR",
					Value:    "100000.00",
				},
				BankReferenceID:    "-",
				Channel:            "VISA",
				ProcessorRefNumber: "test-ref-number",
				Status:             "success",
				CreatedAt:          now,
				UpdatedAt:          fiveMinLater,
				TypeDetail: paymentModel.PaymentTypeDetail{
					CardIssuer: util.ValueToPtr("VISA"),
					CardNumber: util.ValueToPtr("1234"),
					CardBrand:  "VISA",
				},
				TotalSplitAmount: commonModel.Amount{
					Currency: "IDR",
					Value:    "0.00",
				},
				Fee: commonModel.Amount{
					Currency: "IDR",
					Value:    "0.00",
				},
				SettledAmount: commonModel.Amount{
					Currency: "IDR",
					Value:    "100000.00",
				},
				SettlementModel: constant.PaymentMethodChannelTypeAggregator,
				Charges: []*unifiedPaymentModel.ChargeResponse{
					{
						ID:                              constant.EmptyUUID,
						PaymentSessionID:                "valid-payment-uuid",
						PaymentSessionClientReferenceID: "",
						Amount: unifiedPaymentModel.Amount{
							Currency: "IDR",
							Value:    100000,
						},
						Status: constant.StatusSuccess,
						AuthorizedAmount: &unifiedPaymentModel.Amount{
							Currency: "IDR",
							Value:    100000,
						},
						IsCaptured: true,
						CapturedAmount: &unifiedPaymentModel.Amount{
							Currency: "IDR",
							Value:    100000,
						},
						ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{},
					},
				},
			},
		},
		{
			name: "when CC payment and transaction found but the status was not success, should return payment history detail",
			option: paymentModel.PaymentHistoryDetailOption{
				PaymentID:  "valid-payment-id",
				MerchantID: "valid-merchant-id",
			},
			setupMock: func() {
				// Mock payment retrieval
				mockPaymentRepo.On("GetPaymentByIdAndMerchantId", mock.Anything, "valid-payment-id", "valid-merchant-id").
					Return(&paymentModel.Payment{
						UUID:        "valid-payment-uuid",
						MerchantID:  "valid-merchant-id",
						CustomerID:  "customer-id",
						ReferenceID: util.ValueToPtr("sample-ref-id"),
						PaymentMethod: paymentModel.PaymentMethod{
							Type:     paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
							BankName: util.ValueToPtr("BRUH"),
						},
						Metadata: &map[string]interface{}{
							"cardData": map[string]interface{}{
								"cardIssuing": "VISA",
								"last4Digit":  "1234",
								"cardBrand":   "VISA",
							},
						},
						Currency:                 "IDR",
						Amount:                   decimal.NewFromInt(100000),
						Status:                   "success",
						CreatedAt:                now,
						UpdatedAt:                fiveMinLater,
						ExpiredAt:                nil,
						ProcessorReferenceNumber: util.ValueToPtr("test-ref-number"),
					}, nil).Once()

				// Mock GetChargeList
				mockPaymentRepo.On("GetChargeList", mock.Anything, &unifiedPaymentModel.FilterChargeRequest{
					MerchantID:       "valid-merchant-id",
					PaymentSessionID: "valid-payment-uuid",
					Page:             1,
					PerPage:          1000,
				}).Return(&commonModel.PaginationResponse{
					Data: []*unifiedPaymentModel.ChargeResponse{
						{
							ID:                              constant.EmptyUUID,
							PaymentSessionID:                "valid-payment-uuid",
							PaymentSessionClientReferenceID: "sample-ref-id",
							Amount: unifiedPaymentModel.Amount{
								Currency: "IDR",
								Value:    100000,
							},
							IsCaptured:                 false,
							ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{},
						},
					},
				}, nil).Once()

				// Mock account transaction retrieval
				mockAccountTransactionRepo.On("FindByReference", mock.Anything, "valid-payment-uuid", constant.TypePayment).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						Credit:   100000,
						Currency: "IDR",
						Status:   constant.StatusPending,
					}, nil).Once()

				// mock status history
				mockStatusHistoriesRepo.On("GetByReference", mock.Anything, constant.TypePayment, "valid-payment-uuid").
					Return([]*statusHistoryModel.StatusHistory{}, nil).Once()

				mockCustomerRepo.On("GetCustomerById", mock.Anything, "customer-id", "valid-merchant-id").Return(nil, nil).Once()

				refundSvc.On("GetExistingRefundList", mock.Anything, mock.Anything).Return([]refundModel.RefundResponse{}, nil).Once()

			},
			want: &paymentModel.PaymentHistoryDetailResponse{
				UUID:                  "valid-payment-uuid",
				MerchantID:            "valid-merchant-id",
				CustomerID:            "customer-id",
				ReferenceID:           "sample-ref-id",
				PaymentMethod:         paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				PaymentMethodCategory: "",
				Amount: commonModel.Amount{
					Currency: "IDR",
					Value:    "100000.00",
				},
				AmountPaid: commonModel.Amount{
					Currency: "IDR",
					Value:    "0.00",
				},
				BankReferenceID:    "-",
				Channel:            "VISA",
				ProcessorRefNumber: "test-ref-number",
				Status:             "success",
				CreatedAt:          now,
				UpdatedAt:          fiveMinLater,
				TypeDetail: paymentModel.PaymentTypeDetail{
					CardIssuer: util.ValueToPtr("VISA"),
					CardNumber: util.ValueToPtr("1234"),
					CardBrand:  "VISA",
				},
				TotalSplitAmount: commonModel.Amount{
					Currency: "IDR",
					Value:    "0.00",
				},
				Fee: commonModel.Amount{
					Currency: "IDR",
					Value:    "0.00",
				},
				SettledAmount: commonModel.Amount{
					Currency: "IDR",
					Value:    "100000.00",
				},
				SettlementModel: constant.PaymentMethodChannelTypeAggregator,
				Charges: []*unifiedPaymentModel.ChargeResponse{
					{
						ID:                              constant.EmptyUUID,
						PaymentSessionID:                "valid-payment-uuid",
						PaymentSessionClientReferenceID: "sample-ref-id",
						Amount: unifiedPaymentModel.Amount{
							Currency: "IDR",
							Value:    100000,
						},
						IsCaptured:                 false,
						ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{},
					},
				},
			},
		},
		{
			name: "when CC payment and transaction not found, should return payment history detail",
			option: paymentModel.PaymentHistoryDetailOption{
				PaymentID:  "valid-payment-id",
				MerchantID: "valid-merchant-id",
			},
			setupMock: func() {
				// Mock payment retrieval
				mockPaymentRepo.On("GetPaymentByIdAndMerchantId", mock.Anything, "valid-payment-id", "valid-merchant-id").
					Return(&paymentModel.Payment{
						UUID:        "valid-payment-uuid",
						MerchantID:  "valid-merchant-id",
						CustomerID:  "customer-id",
						ReferenceID: util.ValueToPtr("sample-ref-id"),
						PaymentMethod: paymentModel.PaymentMethod{
							Type:     paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
							BankName: util.ValueToPtr("BRUH"),
						},
						Metadata: &map[string]interface{}{
							"cardData": map[string]interface{}{
								"cardIssuing": "VISA",
								"last4Digit":  "1234",
								"cardBrand":   "VISA",
							},
						},
						Currency:                 "IDR",
						Amount:                   decimal.NewFromInt(100000),
						Status:                   "success",
						CreatedAt:                now,
						UpdatedAt:                fiveMinLater,
						ExpiredAt:                nil,
						ProcessorReferenceNumber: util.ValueToPtr("test-ref-number"),
					}, nil).Once()

				// Mock GetChargeList
				mockPaymentRepo.On("GetChargeList", mock.Anything, &unifiedPaymentModel.FilterChargeRequest{
					MerchantID:       "valid-merchant-id",
					PaymentSessionID: "valid-payment-uuid",
					Page:             1,
					PerPage:          1000,
				}).Return(&commonModel.PaginationResponse{
					Data: []*unifiedPaymentModel.ChargeResponse{},
				}, nil).Once()

				// Mock account transaction retrieval
				mockAccountTransactionRepo.On("FindByReference", mock.Anything, "valid-payment-uuid", constant.TypePayment).
					Return(nil, nil).Once()

				// mock status history
				mockStatusHistoriesRepo.On("GetByReference", mock.Anything, constant.TypePayment, "valid-payment-uuid").
					Return([]*statusHistoryModel.StatusHistory{}, nil).Once()

				mockCustomerRepo.On("GetCustomerById", mock.Anything, "customer-id", "valid-merchant-id").Return(nil, nil).Once()

				refundSvc.On("GetExistingRefundList", mock.Anything, mock.Anything).Return([]refundModel.RefundResponse{}, nil).Once()

			},
			want: &paymentModel.PaymentHistoryDetailResponse{
				UUID:                  "valid-payment-uuid",
				MerchantID:            "valid-merchant-id",
				CustomerID:            "customer-id",
				ReferenceID:           "sample-ref-id",
				PaymentMethod:         paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				PaymentMethodCategory: "",
				Amount: commonModel.Amount{
					Currency: "IDR",
					Value:    "100000.00",
				},
				AmountPaid: commonModel.Amount{
					Currency: "IDR",
					Value:    "0.00",
				},
				BankReferenceID:    "-",
				Channel:            "VISA",
				ProcessorRefNumber: "test-ref-number",
				Status:             "success",
				CreatedAt:          now,
				UpdatedAt:          fiveMinLater,
				TypeDetail: paymentModel.PaymentTypeDetail{
					CardIssuer: util.ValueToPtr("VISA"),
					CardNumber: util.ValueToPtr("1234"),
					CardBrand:  "VISA",
				},
				TotalSplitAmount: commonModel.Amount{
					Currency: "IDR",
					Value:    "0.00",
				},
				Fee: commonModel.Amount{
					Currency: "IDR",
					Value:    "0.00",
				},
				SettledAmount: commonModel.Amount{
					Currency: "IDR",
					Value:    "100000.00",
				},
				SettlementModel: constant.PaymentMethodChannelTypeAggregator,
				Charges:         []*unifiedPaymentModel.ChargeResponse{},
			},
		},
		{
			name: "when unable marshal split routing from json, should return error",
			option: paymentModel.PaymentHistoryDetailOption{
				PaymentID:  "valid-payment-id",
				MerchantID: "valid-merchant-id",
			},
			setupMock: func() {
				// Mock payment retrieval
				chanErr := make(chan error)
				mockPaymentRepo.On("GetPaymentByIdAndMerchantId", mock.Anything, "valid-payment-id", "valid-merchant-id").
					Return(&paymentModel.Payment{
						UUID:       "valid-payment-uuid",
						MerchantID: "valid-merchant-id",
						CustomerID: "customer-id",
						PaymentMethod: paymentModel.PaymentMethod{
							Type:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
							BankName: util.ValueToPtr("BRUH"),
						},
						Metadata: &map[string]interface{}{
							"snapCore": map[string]interface{}{
								"Number":         "1234567890",
								"AccountName":    "Test VA",
								"IsClosedAmount": false,
								"IsSingleUse":    false,
							},
							"splitRoutingConfigurations": []interface{}{
								map[string]interface{}{
									"type":             "FIXED",
									"remarks":          "Split payment for merchant A",
									"currency":         chanErr, // trigger marshal error
									"merchantId":       "merchant-a",
									"transferId":       transfer1UUID.String(),
									"fixedAmount":      float64(50000),
									"percentageAmount": float64(0),
									"finalAmount":      float64(0),
								},
							},
						},
						Currency:                 "IDR",
						Amount:                   decimal.NewFromInt(100000),
						Status:                   "success",
						CreatedAt:                now,
						UpdatedAt:                fiveMinLater,
						ExpiredAt:                nil,
						ProcessorReferenceNumber: util.ValueToPtr("test-ref-number"),
					}, nil).Once()

				// Mock GetChargeList
				mockPaymentRepo.On("GetChargeList", mock.Anything, &unifiedPaymentModel.FilterChargeRequest{
					MerchantID:       "valid-merchant-id",
					PaymentSessionID: "valid-payment-uuid",
					Page:             1,
					PerPage:          1000,
				}).Return(&commonModel.PaginationResponse{
					Data: []*unifiedPaymentModel.ChargeResponse{
						{
							ID:                              constant.EmptyUUID,
							PaymentSessionID:                "valid-payment-uuid",
							PaymentSessionClientReferenceID: "",
							Amount: unifiedPaymentModel.Amount{
								Currency: "IDR",
								Value:    100000,
							},
							Status: constant.StatusSuccess,
							AuthorizedAmount: &unifiedPaymentModel.Amount{
								Currency: "IDR",
								Value:    100000,
							},
							IsCaptured: true,
							CapturedAmount: &unifiedPaymentModel.Amount{
								Currency: "IDR",
								Value:    100000,
							},
							ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{},
						},
					},
				}, nil).Once()

				mockAccountTransactionRepo.On("FindByReference", mock.Anything, "valid-payment-uuid", constant.TypePayment).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						Credit:   100000,
						Currency: "IDR",
						Status:   constant.StatusSuccess,
					}, nil).Once()

				mockCustomerRepo.On("GetCustomerById", mock.Anything, "customer-id", "valid-merchant-id").Return(nil, nil).Once()
			},
			wantErr: pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrPaymentSplitRoutingIsNotProcessed),
		},
		{
			name: "when unable unmarshal split routing from json, should return error",
			option: paymentModel.PaymentHistoryDetailOption{
				PaymentID:  "valid-payment-id",
				MerchantID: "valid-merchant-id",
			},
			setupMock: func() {
				// Mock payment retrieval
				mockPaymentRepo.On("GetPaymentByIdAndMerchantId", mock.Anything, "valid-payment-id", "valid-merchant-id").
					Return(&paymentModel.Payment{
						UUID:       "valid-payment-uuid",
						MerchantID: "valid-merchant-id",
						CustomerID: "customer-id",
						PaymentMethod: paymentModel.PaymentMethod{
							Type:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
							BankName: util.ValueToPtr("BRUH"),
						},
						Metadata: &map[string]interface{}{
							"snapCore": map[string]interface{}{
								"Number":         "1234567890",
								"AccountName":    "Test VA",
								"IsClosedAmount": false,
								"IsSingleUse":    false,
							},
							"splitRoutingConfigurations": []interface{}{
								map[string]interface{}{
									"type":             "FIXED",
									"remarks":          "Split payment for merchant A",
									"currency":         990, // trigger error
									"merchantId":       "merchant-a",
									"transferId":       transfer1UUID.String(),
									"fixedAmount":      float64(50000),
									"percentageAmount": float64(0),
									"finalAmount":      float64(0),
								},
								map[string]interface{}{
									"type":             "PERCENTAGE",
									"remarks":          "Split payment for merchant B",
									"currency":         "IDR",
									"merchantId":       "merchant-b",
									"transferId":       transfer2UUID.String(),
									"fixedAmount":      float64(0),
									"percentageAmount": float64(50),
									"finalAmount":      float64(0),
								},
							},
						},
						Currency:                 "IDR",
						Amount:                   decimal.NewFromInt(100000),
						Status:                   "success",
						CreatedAt:                now,
						UpdatedAt:                fiveMinLater,
						ExpiredAt:                nil,
						ProcessorReferenceNumber: util.ValueToPtr("test-ref-number"),
					}, nil).Once()

				// Mock GetChargeList
				mockPaymentRepo.On("GetChargeList", mock.Anything, &unifiedPaymentModel.FilterChargeRequest{
					MerchantID:       "valid-merchant-id",
					PaymentSessionID: "valid-payment-uuid",
					Page:             1,
					PerPage:          1000,
				}).Return(&commonModel.PaginationResponse{
					Data: []*unifiedPaymentModel.ChargeResponse{
						{
							ID:                              constant.EmptyUUID,
							PaymentSessionID:                "valid-payment-uuid",
							PaymentSessionClientReferenceID: "",
							Amount: unifiedPaymentModel.Amount{
								Currency: "IDR",
								Value:    100000,
							},
							Status: constant.StatusSuccess,
							AuthorizedAmount: &unifiedPaymentModel.Amount{
								Currency: "IDR",
								Value:    100000,
							},
							IsCaptured: true,
							CapturedAmount: &unifiedPaymentModel.Amount{
								Currency: "IDR",
								Value:    100000,
							},
							ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{},
						},
					},
				}, nil).Once()

				mockAccountTransactionRepo.On("FindByReference", mock.Anything, "valid-payment-uuid", constant.TypePayment).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						Credit:   100000,
						Currency: "IDR",
						Status:   constant.StatusSuccess,
					}, nil).Once()

				mockCustomerRepo.On("GetCustomerById", mock.Anything, "customer-id", "valid-merchant-id").Return(nil, nil).Once()
			},
			wantErr: pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrPaymentSplitRoutingIsNotProcessed),
		},
		{
			name: "when payment has split routing configurations, should return payment history detail with split routing configs",
			option: paymentModel.PaymentHistoryDetailOption{
				PaymentID:  "valid-payment-id",
				MerchantID: "valid-merchant-id",
			},
			setupMock: func() {
				// Mock payment retrieval with split routing configurations
				mockPaymentRepo.On("GetPaymentByIdAndMerchantId", mock.Anything, "valid-payment-id", "valid-merchant-id").
					Return(&paymentModel.Payment{
						UUID:       "valid-payment-uuid",
						MerchantID: "valid-merchant-id",
						CustomerID: "customer-id",
						PaymentMethod: paymentModel.PaymentMethod{
							Type:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
							BankName: util.ValueToPtr("BRUH"),
						},
						Metadata: &map[string]interface{}{
							"snapCore": map[string]interface{}{
								"Number":         "1234567890",
								"AccountName":    "Test VA",
								"IsClosedAmount": false,
								"IsSingleUse":    false,
							},
							"splitRoutingConfigurations": []interface{}{
								map[string]interface{}{
									"type":             "FIXED",
									"remarks":          "Split payment for merchant A",
									"currency":         "IDR",
									"merchantId":       "merchant-a",
									"transferId":       transfer1UUID.String(),
									"fixedAmount":      float64(50000),
									"percentageAmount": float64(0),
									"finalAmount":      float64(0),
								},
								map[string]interface{}{
									"type":             "PERCENTAGE",
									"remarks":          "Split payment for merchant B",
									"currency":         "IDR",
									"merchantId":       "merchant-b",
									"transferId":       transfer2UUID.String(),
									"fixedAmount":      float64(0),
									"percentageAmount": float64(50),
									"finalAmount":      float64(0),
								},
							},
						},
						Currency:                 "IDR",
						Amount:                   decimal.NewFromInt(100000),
						Status:                   "success",
						CreatedAt:                now,
						UpdatedAt:                fiveMinLater,
						ExpiredAt:                nil,
						ProcessorReferenceNumber: util.ValueToPtr("test-ref-number"),
					}, nil).Once()

				// Mock GetChargeList
				mockPaymentRepo.On("GetChargeList", mock.Anything, &unifiedPaymentModel.FilterChargeRequest{
					MerchantID:       "valid-merchant-id",
					PaymentSessionID: "valid-payment-uuid",
					Page:             1,
					PerPage:          1000,
				}).Return(&commonModel.PaginationResponse{
					Data: []*unifiedPaymentModel.ChargeResponse{
						{
							ID:                              constant.EmptyUUID,
							PaymentSessionID:                "valid-payment-uuid",
							PaymentSessionClientReferenceID: "",
							Amount: unifiedPaymentModel.Amount{
								Currency: "IDR",
								Value:    100000,
							},
							Status: constant.StatusSuccess,
							AuthorizedAmount: &unifiedPaymentModel.Amount{
								Currency: "IDR",
								Value:    100000,
							},
							IsCaptured: true,
							CapturedAmount: &unifiedPaymentModel.Amount{
								Currency: "IDR",
								Value:    100000,
							},
							ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{},
						},
					},
				}, nil).Once()

				mockAccountTransactionRepo.On("FindByReference", mock.Anything, "valid-payment-uuid", constant.TypePayment).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						Credit:   100000,
						Currency: "IDR",
						Status:   constant.StatusSuccess,
					}, nil).Once()

				transferSvc.On("GetById", mock.Anything, transfer1UUID.String(), "valid-merchant-id").
					Return(&transfer.Transfer{
						UUID:   transfer1UUID,
						Status: constant.StatusSuccess,
					}, nil).Once()

				transferSvc.On("GetById", mock.Anything, transfer2UUID.String(), "valid-merchant-id").
					Return(&transfer.Transfer{
						UUID:   transfer2UUID,
						Status: constant.StatusPending,
					}, nil).Once()

				// mock status history
				mockStatusHistoriesRepo.On("GetByReference", mock.Anything, constant.TypePayment, "valid-payment-uuid").
					Return([]*statusHistoryModel.StatusHistory{}, nil).Once()

				mockCustomerRepo.On("GetCustomerById", mock.Anything, "customer-id", "valid-merchant-id").Return(nil, nil).Once()

				refundSvc.On("GetExistingRefundList", mock.Anything, mock.Anything).Return([]refundModel.RefundResponse{}, nil).Once()

			},
			want: &paymentModel.PaymentHistoryDetailResponse{
				UUID:                  "valid-payment-uuid",
				MerchantID:            "valid-merchant-id",
				CustomerID:            "customer-id",
				PaymentMethod:         paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				PaymentMethodCategory: "OPEN_STATIC",
				Amount: commonModel.Amount{
					Currency: "IDR",
					Value:    "100000.00",
				},
				AmountPaid: commonModel.Amount{
					Currency: "IDR",
					Value:    "100000.00",
				},
				BankReferenceID:    "-",
				Channel:            "BRUH",
				ProcessorRefNumber: "test-ref-number",
				Status:             "success",
				CreatedAt:          now,
				UpdatedAt:          fiveMinLater,
				SplitRoutingConfigs: []paymentModel.SplitRoutingConfiguration{
					{
						Type:             "FIXED",
						Remarks:          "Split payment for merchant A",
						Currency:         "IDR",
						MerchantID:       "merchant-a",
						TransferID:       transfer1UUID.String(),
						FixedAmount:      50000,
						PercentageAmount: 0,
						FinalAmount:      50000,
						Status:           constant.StatusSuccess,
					},
					{
						Type:             "PERCENTAGE",
						Remarks:          "Split payment for merchant B",
						Currency:         "IDR",
						MerchantID:       "merchant-b",
						TransferID:       transfer2UUID.String(),
						FixedAmount:      0,
						PercentageAmount: 50,
						FinalAmount:      50000,
						Status:           constant.StatusPending,
					},
				},
				TypeDetail: paymentModel.PaymentTypeDetail{
					VirtualAccountName:   util.ValueToPtr("Test VA"),
					VirtualAccountNumber: util.ValueToPtr("1234567890"),
				},
				TotalSplitAmount: commonModel.Amount{
					Currency: "IDR",
					Value:    "100000.00",
				},
				Fee: commonModel.Amount{
					Currency: "IDR",
					Value:    "0.00",
				},
				SettledAmount: commonModel.Amount{
					Currency: "IDR",
					Value:    "0.00",
				},
				SettlementModel: constant.PaymentMethodChannelTypeAggregator,
				Charges: []*unifiedPaymentModel.ChargeResponse{
					{
						ID:                              constant.EmptyUUID,
						PaymentSessionID:                "valid-payment-uuid",
						PaymentSessionClientReferenceID: "",
						Amount: unifiedPaymentModel.Amount{
							Currency: "IDR",
							Value:    100000,
						},
						Status: constant.StatusSuccess,
						AuthorizedAmount: &unifiedPaymentModel.Amount{
							Currency: "IDR",
							Value:    100000,
						},
						IsCaptured: true,
						CapturedAmount: &unifiedPaymentModel.Amount{
							Currency: "IDR",
							Value:    100000,
						},
						ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{},
					},
				},
				Metadata: nil,
			},
		},
		{
			name: "when transfer service returns error, should return error",
			option: paymentModel.PaymentHistoryDetailOption{
				PaymentID:  "valid-payment-id",
				MerchantID: "valid-merchant-id",
			},
			setupMock: func() {
				// Mock payment retrieval with split routing configurations
				mockPaymentRepo.On("GetPaymentByIdAndMerchantId", mock.Anything, "valid-payment-id", "valid-merchant-id").
					Return(&paymentModel.Payment{
						UUID:       "valid-payment-uuid",
						MerchantID: "valid-merchant-id",
						CustomerID: "customer-id",
						PaymentMethod: paymentModel.PaymentMethod{
							Type:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
							BankName: util.ValueToPtr("BRUH"),
						},
						Metadata: &map[string]interface{}{
							"snapCore": map[string]interface{}{
								"Number":         "1234567890",
								"AccountName":    "Test VA",
								"IsClosedAmount": false,
								"IsSingleUse":    false,
							},
							"splitRoutingConfigurations": []interface{}{
								map[string]interface{}{
									"type":             "FIXED",
									"remarks":          "Split payment for merchant A",
									"currency":         "IDR",
									"merchantId":       "merchant-a",
									"transferId":       transfer1UUID.String(),
									"fixedAmount":      float64(50000),
									"percentageAmount": float64(0),
								},
							},
						},
						Currency:                 "IDR",
						Amount:                   decimal.NewFromInt(100000),
						Status:                   "success",
						CreatedAt:                now,
						UpdatedAt:                fiveMinLater,
						ExpiredAt:                nil,
						ProcessorReferenceNumber: util.ValueToPtr("test-ref-number"),
					}, nil).Once()

				// Mock GetChargeList
				mockPaymentRepo.On("GetChargeList", mock.Anything, &unifiedPaymentModel.FilterChargeRequest{
					MerchantID:       "valid-merchant-id",
					PaymentSessionID: "valid-payment-uuid",
					Page:             1,
					PerPage:          1000,
				}).Return(&commonModel.PaginationResponse{
					Data: []*unifiedPaymentModel.ChargeResponse{
						{
							ID:                              constant.EmptyUUID,
							PaymentSessionID:                "valid-payment-uuid",
							PaymentSessionClientReferenceID: "",
							Amount: unifiedPaymentModel.Amount{
								Currency: "IDR",
								Value:    100000,
							},
							Status: constant.StatusSuccess,
							AuthorizedAmount: &unifiedPaymentModel.Amount{
								Currency: "IDR",
								Value:    100000,
							},
							IsCaptured: true,
							CapturedAmount: &unifiedPaymentModel.Amount{
								Currency: "IDR",
								Value:    100000,
							},
							ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{},
						},
					},
				}, nil).Once()

				mockAccountTransactionRepo.On("FindByReference", mock.Anything, "valid-payment-uuid", constant.TypePayment).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						Credit:   100000,
						Currency: "IDR",
						Status:   constant.StatusSuccess,
					}, nil).Once()

				transferSvc.On("GetById", mock.Anything, transfer1UUID.String(), "valid-merchant-id").
					Return(nil, errors.New("transfer service error")).Once()

				mockCustomerRepo.On("GetCustomerById", mock.Anything, "customer-id", "valid-merchant-id").Return(nil, nil).Once()

			},
			wantErr: pkgErrors.New(response.HttpErrDatabase, errors.New("transfer service error")),
		},
		{
			name: "when transfer is not found, should return error",
			option: paymentModel.PaymentHistoryDetailOption{
				PaymentID:  "valid-payment-id",
				MerchantID: "valid-merchant-id",
			},
			setupMock: func() {
				mockPaymentRepo.On("GetPaymentByIdAndMerchantId", mock.Anything, "valid-payment-id", "valid-merchant-id").
					Return(&paymentModel.Payment{
						UUID:       "valid-payment-uuid",
						MerchantID: "valid-merchant-id",
						CustomerID: "customer-id",
						PaymentMethod: paymentModel.PaymentMethod{
							Type:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
							BankName: util.ValueToPtr("BRUH"),
						},
						Metadata: &map[string]interface{}{
							"snapCore": map[string]interface{}{
								"Number":         "1234567890",
								"AccountName":    "Test VA",
								"IsClosedAmount": false,
								"IsSingleUse":    false,
							},
							"splitRoutingConfigurations": []interface{}{
								map[string]interface{}{
									"type":             "FIXED",
									"remarks":          "Split payment for merchant A",
									"currency":         "IDR",
									"merchantId":       "merchant-a",
									"transferId":       transfer1UUID.String(),
									"fixedAmount":      float64(50000),
									"percentageAmount": float64(0),
								},
							},
						},
						Currency:                 "IDR",
						Amount:                   decimal.NewFromInt(100000),
						Status:                   "success",
						CreatedAt:                now,
						UpdatedAt:                fiveMinLater,
						ExpiredAt:                nil,
						ProcessorReferenceNumber: util.ValueToPtr("test-ref-number"),
					}, nil).Once()

				// Mock GetChargeList
				mockPaymentRepo.On("GetChargeList", mock.Anything, &unifiedPaymentModel.FilterChargeRequest{
					MerchantID:       "valid-merchant-id",
					PaymentSessionID: "valid-payment-uuid",
					Page:             1,
					PerPage:          1000,
				}).Return(&commonModel.PaginationResponse{
					Data: []*unifiedPaymentModel.ChargeResponse{
						{
							ID:                              constant.EmptyUUID,
							PaymentSessionID:                "valid-payment-uuid",
							PaymentSessionClientReferenceID: "",
							Amount: unifiedPaymentModel.Amount{
								Currency: "IDR",
								Value:    100000,
							},
							Status: constant.StatusSuccess,
							AuthorizedAmount: &unifiedPaymentModel.Amount{
								Currency: "IDR",
								Value:    100000,
							},
							IsCaptured: true,
							CapturedAmount: &unifiedPaymentModel.Amount{
								Currency: "IDR",
								Value:    100000,
							},
							ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{},
						},
					},
				}, nil).Once()

				mockAccountTransactionRepo.On("FindByReference", mock.Anything, "valid-payment-uuid", constant.TypePayment).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						Credit:   100000,
						Currency: "IDR",
						Status:   constant.StatusSuccess,
					}, nil).Once()

				transferSvc.On("GetById", mock.Anything, transfer1UUID.String(), "valid-merchant-id").
					Return(nil, nil).Once()

				mockCustomerRepo.On("GetCustomerById", mock.Anything, "customer-id", "valid-merchant-id").Return(nil, nil).Once()
			},
			wantErr: pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrPaymentSplitRoutingIsNotProcessed),
		},
		{
			name: "when failed to get customer, should return error",
			option: paymentModel.PaymentHistoryDetailOption{
				PaymentID:  "valid-payment-id",
				MerchantID: "valid-merchant-id",
			},
			setupMock: func() {
				mockPaymentRepo.On("GetPaymentByIdAndMerchantId", mock.Anything, "valid-payment-id", "valid-merchant-id").
					Return(&paymentModel.Payment{
						UUID:       "valid-payment-uuid",
						MerchantID: "valid-merchant-id",
						CustomerID: "customer-id",
						PaymentMethod: paymentModel.PaymentMethod{
							Type:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
							BankName: util.ValueToPtr("BRUH"),
						},
						Metadata: &map[string]interface{}{
							"snapCore": map[string]interface{}{
								"Number":         "1234567890",
								"AccountName":    "Test VA",
								"IsClosedAmount": false,
								"IsSingleUse":    false,
							},
							"splitRoutingConfigurations": []interface{}{
								map[string]interface{}{
									"type":             "FIXED",
									"remarks":          "Split payment for merchant A",
									"currency":         "IDR",
									"merchantId":       "merchant-a",
									"transferId":       transfer1UUID.String(),
									"fixedAmount":      float64(50000),
									"percentageAmount": float64(0),
								},
							},
						},
						Currency:                 "IDR",
						Amount:                   decimal.NewFromInt(100000),
						Status:                   "success",
						CreatedAt:                now,
						UpdatedAt:                fiveMinLater,
						ExpiredAt:                nil,
						ProcessorReferenceNumber: util.ValueToPtr("test-ref-number"),
					}, nil).Once()

				// Mock GetChargeList
				mockPaymentRepo.On("GetChargeList", mock.Anything, &unifiedPaymentModel.FilterChargeRequest{
					MerchantID:       "valid-merchant-id",
					PaymentSessionID: "valid-payment-uuid",
					Page:             1,
					PerPage:          1000,
				}).Return(&commonModel.PaginationResponse{
					Data: []*unifiedPaymentModel.ChargeResponse{
						{
							ID:                              constant.EmptyUUID,
							PaymentSessionID:                "valid-payment-uuid",
							PaymentSessionClientReferenceID: "",
							Amount: unifiedPaymentModel.Amount{
								Currency: "IDR",
								Value:    100000,
							},
							Status: constant.StatusSuccess,
							AuthorizedAmount: &unifiedPaymentModel.Amount{
								Currency: "IDR",
								Value:    100000,
							},
							IsCaptured: true,
							CapturedAmount: &unifiedPaymentModel.Amount{
								Currency: "IDR",
								Value:    100000,
							},
							ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{},
						},
					},
				}, nil).Once()

				mockAccountTransactionRepo.On("FindByReference", mock.Anything, "valid-payment-uuid", constant.TypePayment).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						Credit:   100000,
						Currency: "IDR",
						Status:   constant.StatusSuccess,
					}, nil).Once()

				mockCustomerRepo.On("GetCustomerById", mock.Anything, "customer-id", "valid-merchant-id").Return(nil, constant.ErrSomeErrorForUnitTest).Once()

			},
			wantErr: constant.ErrSomeErrorForUnitTest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()
			paymentSvc := PaymentService{
				logger:                 mockLogger,
				paymentRepo:            mockPaymentRepo,
				accountTransactionRepo: mockAccountTransactionRepo,
				customerRepo:           mockCustomerRepo,
				statusHistoriesRepo:    mockStatusHistoriesRepo,
				transferSvc:            transferSvc,
				refundSvc:              refundSvc,
			}

			ctx := context.Background()
			data, err := paymentSvc.GetPaymentHistoryDetail(ctx, tc.option)
			if tc.wantErr != nil {
				assert.Equal(t, tc.wantErr, err)
				require.Empty(t, data)
			} else {
				assert.NoError(t, err)
				assert.EqualValues(t, tc.want, data)
			}

			mockPaymentRepo.AssertExpectations(t)
			mockAccountTransactionRepo.AssertExpectations(t)
			mockStatusHistoriesRepo.AssertExpectations(t)
			transferSvc.AssertExpectations(t)
			refundSvc.AssertExpectations(t)
		})
	}
}

func TestGetPaymentTypeDetail(t *testing.T) {
	var (
		ctx                = context.Background()
		mockLogger, _      = loggerMocks.NewZapLogger(loggerMocks.Config{})
		mockPaymentService = PaymentService{
			logger: mockLogger,
		}

		vaAccountName   = "john"
		vaAccountNumber = "123"
	)

	testCases := []struct {
		name               string
		method             string
		metadata           *map[string]interface{}
		ledgerMetadataFunc func() []byte
		want               paymentModel.PaymentTypeDetail
		shouldErr          bool
	}{
		{
			name:   "when method is VA and invalid metadata, should return default VA details",
			method: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
			metadata: &map[string]interface{}{
				"invalid": "metadata",
			},
			want: paymentModel.PaymentTypeDetail{
				VirtualAccountName:   util.ValueToPtr(""),
				VirtualAccountNumber: util.ValueToPtr(""),
			},
		},
		{
			name:   "when method is VA and valid metadata, should return VA details",
			method: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
			metadata: &map[string]interface{}{
				"snapCore": map[string]any{
					"accountName": vaAccountName,
					"number":      vaAccountNumber,
				},
			},
			want: paymentModel.PaymentTypeDetail{
				VirtualAccountName:   &vaAccountName,
				VirtualAccountNumber: &vaAccountNumber,
			},
		},
		{
			name:     "when method is QRIS and metadata is missing snapCore, should return default QRIS details",
			method:   paymentConstant.PAYMENT_METHOD_QRIS,
			metadata: &map[string]interface{}{},
			want: paymentModel.PaymentTypeDetail{
				QRISMerchantName: util.ValueToPtr(""),
				QRISURL:          util.ValueToPtr(""),
			},
		},
		{
			name:   "when method is QRIS and metadata has snapCore, should return QRIS DYNAMIC",
			method: paymentConstant.PAYMENT_METHOD_QRIS,
			metadata: &map[string]interface{}{
				"snapCore": map[string]interface{}{
					"merchantName": "sample-merchant",
					"qrUrl":        "sample-url",
				},
			},
			want: paymentModel.PaymentTypeDetail{
				QRISMerchantName: util.ValueToPtr("sample-merchant"),
				QRISURL:          util.ValueToPtr("sample-url"),
			},
		},
		{
			name:   "when method is credit card and valid metadata, should return credit card details",
			method: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
			metadata: &map[string]interface{}{
				"cardData": map[string]interface{}{
					"cardIssuing": "VISA",
					"last4Digit":  "1234",
					"cardBrand":   "VISA",
				},
			},
			want: paymentModel.PaymentTypeDetail{
				CardIssuer: util.ValueToPtr("VISA"),
				CardNumber: util.ValueToPtr("1234"),
				CardBrand:  "VISA",
			},
		},
		{
			name:   "when method is credit card and valid metadata but don't have cardData, should return default value",
			method: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
			metadata: &map[string]interface{}{
				"otherCardData": map[string]interface{}{
					"cardIssuing": "VISA",
					"last4Digit":  "1234",
					"cardBrand":   "VISA",
				},
			},
			want: paymentModel.PaymentTypeDetail{
				CardIssuer: util.ValueToPtr(""),
				CardNumber: util.ValueToPtr(""),
				CardBrand:  "",
			},
		},
		{
			name:     "when method is ewallet and don't have ewallet, should return default value",
			method:   paymentConstant.PAYMENT_METHOD_EWALLET,
			metadata: &map[string]interface{}{},
			want: paymentModel.PaymentTypeDetail{
				EWalletChannel:        "",
				EWalletAppRedirectURL: "",
				EWalletWebRedirectURL: "",
			},
		},
		{
			name:   "when method is ewallet, should return ewallet value",
			method: paymentConstant.PAYMENT_METHOD_EWALLET,
			metadata: &map[string]interface{}{
				"isUnifiedPaymentV2": true,
			},
			ledgerMetadataFunc: func() []byte {
				return []byte(`{
					"methodDetail": {
						"ewallet": {
							"appRedirectUrl": "ewallet://redirectUrl",
							"webRedirectUrl": "ewallet://webRedirectUrl",
							"channel":        "DANA"
						}
					}
				}`)
			},
			want: paymentModel.PaymentTypeDetail{
				EWalletChannel:        "DANA",
				EWalletAppRedirectURL: "ewallet://redirectUrl",
				EWalletWebRedirectURL: "ewallet://webRedirectUrl",
			},
		},
		{
			name:     "when metadata is nil, should return empty PaymentTypeDetail",
			method:   paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
			metadata: nil,
			want:     paymentModel.PaymentTypeDetail{},
		},
		{
			name:   "when payment method is unknown, then should return empty data",
			method: "unknown",
			metadata: &map[string]interface{}{
				"otherCardData": map[string]interface{}{
					"cardIssuing": "VISA",
					"last4Digit":  "1234",
					"cardBrand":   "VISA",
				},
			},
			want: paymentModel.PaymentTypeDetail{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ledgerMetadata := []byte{}
			if tc.ledgerMetadataFunc != nil {
				ledgerMetadata = tc.ledgerMetadataFunc()
			}
			result := mockPaymentService.GetPaymentTypeDetail(ctx, tc.method, tc.metadata, ledgerMetadata)

			if tc.metadata == nil {
				assert.Nil(t, result.VirtualAccountName)
				assert.Nil(t, result.VirtualAccountNumber)
				assert.Nil(t, result.QRISMerchantName)
				assert.Nil(t, result.QRISURL)
				assert.Nil(t, result.CardIssuer)
				assert.Nil(t, result.CardNumber)
				return
			}

			if tc.method == paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT {
				assert.Equal(t, *tc.want.VirtualAccountName, *result.VirtualAccountName)
				assert.Equal(t, *tc.want.VirtualAccountNumber, *result.VirtualAccountNumber)
			}

			if tc.method == paymentConstant.PAYMENT_METHOD_QRIS {
				assert.Equal(t, *tc.want.QRISMerchantName, *result.QRISMerchantName)
				assert.Equal(t, *tc.want.QRISURL, *result.QRISURL)
			}

			if tc.method == paymentConstant.PAYMENT_METHOD_CREDIT_CARD {
				assert.Equal(t, result.CardIssuer, tc.want.CardIssuer)
				assert.Equal(t, result.CardNumber, tc.want.CardNumber)
				assert.Equal(t, result.CardBrand, tc.want.CardBrand)
			}
		})
	}
}

func TestGetPaymentSubType(t *testing.T) {
	var (
		ctx                = context.Background()
		mockLogger, _      = loggerMocks.NewZapLogger(loggerMocks.Config{})
		mockPaymentService = PaymentService{
			logger: mockLogger,
		}
	)

	testCases := []struct {
		name     string
		method   string
		metadata *map[string]interface{}
		want     string
	}{
		{
			name:     "when metadata is nil, should return empty string",
			method:   paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
			metadata: nil,
			want:     "",
		},
		{
			name:   "when method is VA and invalid metadata, should return empty string",
			method: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
			metadata: &map[string]interface{}{
				"invalid": "metadata",
			},
			want: "",
		},
		{
			name:   "when method is VA and valid metadata, should return VA VIRTUAL_ACCOUNT_TRX_TYPE_CLOSED_DYNAMIC",
			method: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
			metadata: &map[string]interface{}{
				"snapCore": map[string]any{
					"isClosedAmount": true,
					"isSingleUse":    true,
				},
			},
			want: paymentConstant.VIRTUAL_ACCOUNT_TRX_TYPE_CLOSED_DYNAMIC,
		},
		{
			name:   "when method is VA and valid metadata, should return VA VIRTUAL_ACCOUNT_TRX_TYPE_CLOSED_STATIC",
			method: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
			metadata: &map[string]interface{}{
				"snapCore": map[string]any{
					"isClosedAmount": true,
					"isSingleUse":    false,
				},
			},
			want: paymentConstant.VIRTUAL_ACCOUNT_TRX_TYPE_CLOSED_STATIC,
		},
		{
			name:   "when method is VA and valid metadata, should return VA VIRTUAL_ACCOUNT_TRX_TYPE_OPEN_STATIC",
			method: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
			metadata: &map[string]interface{}{
				"snapCore": map[string]any{
					"isClosedAmount": false,
					"isSingleUse":    false,
				},
			},
			want: paymentConstant.VIRTUAL_ACCOUNT_TRX_TYPE_OPEN_STATIC,
		},
		{
			name:   "when method is QRIS and valid metadata, should return QRIS type",
			method: paymentConstant.PAYMENT_METHOD_QRIS,
			metadata: &map[string]interface{}{
				"qrType": "dynamic",
			},
			want: "dynamic",
		},
		{
			name:   "when method is QRIS and invalid metadata, should return empty string",
			method: paymentConstant.PAYMENT_METHOD_QRIS,
			metadata: &map[string]interface{}{
				"invalid": "data",
			},
			want: "",
		},
		{
			name:     "when method is QRIS and metadata is nil, should return error",
			method:   paymentConstant.PAYMENT_METHOD_QRIS,
			metadata: nil,
			want:     "",
		},
		{
			name:   "when method is credit card and valid metadata, should return card type",
			method: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
			metadata: &map[string]interface{}{
				"cardData": map[string]interface{}{
					"cardType": "debit",
				},
			},
			want: "debit",
		},
		{
			name:     "when method is credit card and missing card data, should return empty string",
			method:   paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
			metadata: &map[string]interface{}{},
			want:     "",
		},
		{
			name:     "when method is unknown, then should return empty string",
			method:   "unknown",
			metadata: &map[string]interface{}{},
			want:     "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := mockPaymentService.GetPaymentSubType(ctx, tc.method, tc.metadata)
			assert.Equal(t, tc.want, result)
		})
	}
}

func TestGetTodayPaymentInsight(t *testing.T) {
	var (
		ctx             = context.Background()
		mockPaymentRepo repositoryMocks.IPaymentRepository
		mockAccountRepo repositoryMocks.IAccountRepository
		paymentService  = PaymentService{
			paymentRepo: &mockPaymentRepo,
			accountRepo: &mockAccountRepo,
		}
		validMerchantID = uuid.New().String()
	)
	testCases := []struct {
		name      string
		payload   paymentModel.PaymentInsightOption
		callMock  func()
		want      paymentModel.PaymentInsightItem
		wantErr   error
		shouldErr bool
	}{
		{
			name: "when merchant id was invalid, then should return error",
			payload: paymentModel.PaymentInsightOption{
				MerchantID: "invalid-merchant-id",
				Status:     paymentConstant.PAYMENT_STATUS_SUCCESS,
			},
			callMock:  func() {},
			wantErr:   fmt.Errorf("invalid UUID length: %d", len("invalid-merchant-id")),
			shouldErr: true,
		},
		{
			name: "when failed to get merchant account, then should return error",
			payload: paymentModel.PaymentInsightOption{
				MerchantID: validMerchantID,
				Status:     paymentConstant.PAYMENT_STATUS_SUCCESS,
			},
			callMock: func() {
				mockAccountRepo.On("FindMerchantAccountByName", mock.Anything, mock.Anything, constant.TypePayment).
					Return(nil, errors.New("database down")).Once()

				mockPaymentRepo.On("GetTodayPaymentStatusInsight", mock.Anything, paymentModel.PaymentInsightOption{
					MerchantID: validMerchantID,
					Status:     paymentConstant.PAYMENT_STATUS_SUCCESS,
				}).Return(&paymentModel.PaymentInsightItem{
					Total: 3,
					TotalAmount: commonModel.Amount{
						Value: strconv.FormatFloat(12000000, 'f', 2, 64),
					},
				}, nil).Once()
			},
			wantErr:   errors.New("database down"),
			shouldErr: true,
		},
		{
			name: "when merchant account was nil, then should return default value even though the insight item was found",
			payload: paymentModel.PaymentInsightOption{
				MerchantID: validMerchantID,
				Status:     paymentConstant.PAYMENT_STATUS_SUCCESS,
			},
			callMock: func() {
				mockAccountRepo.On("FindMerchantAccountByName", mock.Anything, mock.Anything, constant.TypePayment).
					Return(nil, nil).Once()

				mockPaymentRepo.On("GetTodayPaymentStatusInsight", mock.Anything, paymentModel.PaymentInsightOption{
					MerchantID: validMerchantID,
					Status:     paymentConstant.PAYMENT_STATUS_SUCCESS,
				}).Return(&paymentModel.PaymentInsightItem{
					Total: 3,
					TotalAmount: commonModel.Amount{
						Value: strconv.FormatFloat(12000000, 'f', 2, 64),
					},
				}, nil).Once()
			},
			want: paymentModel.PaymentInsightItem{
				Total: 0,
				TotalAmount: commonModel.Amount{
					Value: strconv.FormatFloat(0, 'f', 2, 64),
				},
			},
		},
		{
			name: "when failed to get payment insight, then should return error",
			payload: paymentModel.PaymentInsightOption{
				MerchantID: validMerchantID,
				Status:     paymentConstant.PAYMENT_STATUS_SUCCESS,
			},
			callMock: func() {
				mockAccountRepo.On("FindMerchantAccountByName", mock.Anything, mock.Anything, constant.TypePayment).
					Return(&accountModel.Account{
						UUID:     uuid.New(),
						Name:     constant.TypePayment,
						Currency: constant.CurrencyIDR,
					}, nil).Once()

				mockPaymentRepo.On("GetTodayPaymentStatusInsight", mock.Anything, paymentModel.PaymentInsightOption{
					MerchantID: validMerchantID,
					Status:     paymentConstant.PAYMENT_STATUS_SUCCESS,
				}).Return(nil, errors.New("invalid arg of insight")).Once()
			},
			shouldErr: true,
			wantErr:   errors.New("invalid arg of insight"),
		},
		{
			name: "when payment insight not found, then should not return error",
			payload: paymentModel.PaymentInsightOption{
				MerchantID: validMerchantID,
				Status:     paymentConstant.PAYMENT_STATUS_SUCCESS,
			},
			callMock: func() {
				mockAccountRepo.On("FindMerchantAccountByName", mock.Anything, mock.Anything, constant.TypePayment).
					Return(&accountModel.Account{
						UUID:     uuid.New(),
						Name:     constant.TypePayment,
						Currency: constant.CurrencyIDR,
					}, nil).Once()

				mockPaymentRepo.On("GetTodayPaymentStatusInsight", mock.Anything, paymentModel.PaymentInsightOption{
					MerchantID: validMerchantID,
					Status:     paymentConstant.PAYMENT_STATUS_SUCCESS,
				}).Return(nil, nil).Once()
			},
			want: paymentModel.PaymentInsightItem{
				TotalAmount: commonModel.Amount{
					Value: strconv.FormatFloat(0, 'f', 2, 64),
				},
			},
		},
		{
			name: "when payment insight found, then return the insight with the currency",
			payload: paymentModel.PaymentInsightOption{
				MerchantID: validMerchantID,
				Status:     paymentConstant.PAYMENT_STATUS_SUCCESS,
			},
			callMock: func() {
				mockAccountRepo.On("FindMerchantAccountByName", mock.Anything, mock.Anything, constant.TypePayment).
					Return(&accountModel.Account{
						UUID:     uuid.New(),
						Name:     constant.TypePayment,
						Currency: constant.CurrencyIDR,
					}, nil).Once()

				mockPaymentRepo.On("GetTodayPaymentStatusInsight", mock.Anything, paymentModel.PaymentInsightOption{
					MerchantID: validMerchantID,
					Status:     paymentConstant.PAYMENT_STATUS_SUCCESS,
				}).Return(&paymentModel.PaymentInsightItem{
					Total: 3,
					TotalAmount: commonModel.Amount{
						Value: strconv.FormatFloat(12000000, 'f', 2, 64),
					},
				}, nil).Once()
			},
			want: paymentModel.PaymentInsightItem{
				Total: 3,
				TotalAmount: commonModel.Amount{
					Currency: constant.CurrencyIDR,
					Value:    strconv.FormatFloat(12000000, 'f', 2, 64),
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.callMock()

			insight, err := paymentService.GetTodayPaymentInsight(ctx, tc.payload)

			if tc.shouldErr {
				assert.NotNil(t, err)
				assert.Equal(t, tc.wantErr.Error(), err.Error())
				return
			}

			assert.Nil(t, err)
			assert.Equal(t, tc.want.Total, insight.Total)
			assert.Equal(t, tc.want.TotalAmount.Currency, insight.TotalAmount.Currency)
			assert.Equal(t, tc.want.TotalAmount.Value, insight.TotalAmount.Value)
		})
	}
}

func TestGetSplitRoutingByTransferID(t *testing.T) {
	paymentRepo := repositoryMocks.NewIPaymentRepository(t)
	transferSvc := serviceMocks.NewITransferService(t)
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	//expectedResponse := &splitRoutingPaymentModel.SplitRoutingPaymentDetailResponse{}
	//validPayment := &paymentModel.Payment{}

	testCases := []struct {
		name      string
		wantErr   bool
		setupMock func()
	}{
		{
			name:    "ERROR: get GetPaymentById repo",
			wantErr: true,
			setupMock: func() {
				paymentRepo.On("GetPaymentById",
					constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: get GetPaymentById not found",
			wantErr: true,
			setupMock: func() {
				paymentRepo.On("GetPaymentById",
					constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(nil, nil)
			},
		},
		{
			name:    "ERROR: empty metadata",
			wantErr: true,
			setupMock: func() {
				paymentRepo.On("GetPaymentById",
					constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(&paymentModel.Payment{}, nil)
			},
		},
		{
			name:    "ERROR: payment does not have split routing destination",
			wantErr: true,
			setupMock: func() {
				paymentRepo.On("GetPaymentById",
					constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(&paymentModel.Payment{
					Metadata: &map[string]any{},
				}, nil)
			},
		},
		{
			name:    "ERROR: split routing destination is not found",
			wantErr: true,
			setupMock: func() {
				paymentRepo.On("GetPaymentById",
					constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(&paymentModel.Payment{
					Metadata: &map[string]any{
						constant.SplitRoutingPaymentConfigKey: []*splitRoutingPaymentModel.PaymentSplitRoutingConfiguration{{
							TransferID: "not-match",
						}},
					},
				}, nil)
			},
		},
		{
			name:    "ERROR: get transfer error",
			wantErr: true,
			setupMock: func() {
				referenceID := "ref-id"
				paymentRepo.On("GetPaymentById",
					constant.ValueCtxMockType(), constant.StringMockType()).
					Return(&paymentModel.Payment{
						ReferenceID: &referenceID,
						Metadata: &map[string]any{
							constant.SplitRoutingPaymentConfigKey: []*splitRoutingPaymentModel.PaymentSplitRoutingConfiguration{{
								TransferID: "transfer-id",
							}},
						},
					}, nil)

				transferSvc.On("GetById",
					constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType()).
					Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: get transfer not found",
			wantErr: true,
			setupMock: func() {
				transferSvc.On("GetById",
					constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType()).
					Once().Return(nil, nil)
			},
		},
		{
			name:    "ERROR: transfer is not SUCCESS yet",
			wantErr: true,
			setupMock: func() {
				transferSvc.On("GetById",
					constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType()).
					Once().Return(&transfer.Transfer{Status: constant.StatusPending}, nil)
			},
		},
		{
			name:    "SUCCESS",
			wantErr: false,
			setupMock: func() {
				transferSvc.On("GetById",
					constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType()).
					Return(&transfer.Transfer{Status: constant.StatusSuccess}, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()
			paymentSvc := New(paymentRepo, logger, nil, nil, nil, nil, nil,
				WithTransferService(transferSvc),
			)

			ctx := context.Background()
			data, err := paymentSvc.GetSplitRoutingByTransferID(ctx, "payment-id", "transfer-id")
			if tc.wantErr {
				assert.Error(t, err)
				require.Empty(t, data)
			} else {
				assert.NoError(t, err)
				require.NotEmpty(t, data)
			}

			paymentRepo.AssertExpectations(t)
		})
	}
}

func TestGetDetailForInternalDashboard(t *testing.T) {
	paymentRepo := repositoryMocks.NewIPaymentRepository(t)
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	testCases := []struct {
		name      string
		wantErr   bool
		setupMock func()
	}{
		{
			name:    "ERROR: get GetPaymentById repo",
			wantErr: true,
			setupMock: func() {
				paymentRepo.On("GetPaymentById",
					constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: get GetPaymentById not found",
			wantErr: true,
			setupMock: func() {
				paymentRepo.On("GetPaymentById",
					constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(nil, nil)
			},
		},
		{
			name:    "SUCCESS",
			wantErr: false,
			setupMock: func() {
				paymentRepo.On("GetPaymentById",
					constant.ValueCtxMockType(), constant.StringMockType()).
					Return(&paymentModel.Payment{
						UUID: uuid.NewString(),
					}, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()
			paymentSvc := New(paymentRepo, logger, nil, nil, nil, nil, nil)

			ctx := context.Background()
			data, err := paymentSvc.GetDetailByID(ctx, "payment-id")
			if tc.wantErr {
				assert.Error(t, err)
				require.Empty(t, data)
			} else {
				assert.NoError(t, err)
				require.NotEmpty(t, data)
			}

			paymentRepo.AssertExpectations(t)
		})
	}
}
func TestPaymentServiceGetPaymentDetailForPaymentUI(t *testing.T) {
	var (
		mockLogger, _              = loggerMocks.NewZapLogger(loggerMocks.Config{})
		mockPaymentRepo            = repositoryMocks.NewIPaymentRepository(t)
		mockAccountTransactionRepo = repositoryMocks.NewIAccountTransactionRepository(t)
		mockMerchantRepo           = repositoryMocks.NewIMerchantRepository(t)
		mockUnifiedPaymentSvc      = serviceMocks.NewIUnifiedPaymentService(t)
		vaAccountName              = "john-doe"
		vaAccountNumber            = "123"
	)

	errDB := errors.New("database error")
	validRefID := util.ValueToPtr("valid-ref-id")
	validPaymentID := "valid-payment-id"
	trxID := uuid.New()
	now := time.Now()

	testCases := []struct {
		name      string
		paymentID string
		setupMock func()
		want      *paymentModel.PaymentDetailForPaymentUIResponse
		wantErr   error
	}{
		{
			name:      "when failed to get payment, then should return error",
			paymentID: "invalid-payment-id",
			setupMock: func() {
				mockPaymentRepo.On("GetPaymentById", mock.Anything, "invalid-payment-id").
					Return(nil, errDB).
					Once()
			},
			wantErr: pkgErrors.New(response.HttpErrDatabase, errDB),
		},
		{
			name:      "when payment is not found, then should return error",
			paymentID: "missing-payment-id",
			setupMock: func() {
				mockPaymentRepo.On("GetPaymentById", mock.Anything, "missing-payment-id").
					Return(nil, nil).
					Once()
			},
			wantErr: pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrPaymentNotFound),
		},
		{
			name:      "when failed to get merchant, then should return error",
			paymentID: validPaymentID,
			setupMock: func() {
				mockPaymentRepo.On("GetPaymentById", mock.Anything, validPaymentID).
					Return(&paymentModel.Payment{
						UUID:       "valid-payment-uuid",
						MerchantID: "valid-merchant-id",
					}, nil).Once()
				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "valid-merchant-id").
					Return(nil, errDB).
					Once()
			},
			wantErr: pkgErrors.New(response.HttpErrDatabase, errDB),
		},
		{
			name:      "when merchant not found, then should return error",
			paymentID: validPaymentID,
			setupMock: func() {
				mockPaymentRepo.On("GetPaymentById", mock.Anything, validPaymentID).
					Return(&paymentModel.Payment{
						UUID:       "valid-payment-uuid",
						MerchantID: "valid-merchant-id",
					}, nil).Once()
				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "valid-merchant-id").
					Return(nil, nil).
					Once()
			},
			wantErr: pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrMerchantNotFound),
		},
		{
			name:      "when failed to get account transaction, then should return error",
			paymentID: validPaymentID,
			setupMock: func() {
				mockPaymentRepo.On("GetPaymentById", mock.Anything, validPaymentID).
					Return(&paymentModel.Payment{
						UUID:        validPaymentID,
						MerchantID:  "valid-merchant-id",
						ReferenceID: validRefID,
					}, nil).Once()
				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "valid-merchant-id").
					Return(&merchantModel.Merchant{
						UUID: "valid-merchant-id",
						KYCStatus: sql.NullString{
							String: constant.KYCStatusApproved,
							Valid:  true,
						},
						ParentID: sql.NullString{
							String: "parent-id",
							Valid:  true,
						},
					}, nil).
					Once()
				mockAccountTransactionRepo.On("FindByReference", mock.Anything, validPaymentID, constant.TypePayment).
					Return(nil, errDB).
					Once()
			},
			wantErr: pkgErrors.New(response.HttpErrDatabase, errDB),
		},
		{
			name:      "when its VA Payment, then should contain the Bank Name as payment channel",
			paymentID: validPaymentID,
			setupMock: func() {
				mockPaymentRepo.On("GetPaymentById", mock.Anything, validPaymentID).
					Return(&paymentModel.Payment{
						UUID:        validPaymentID,
						MerchantID:  "valid-merchant-id",
						ReferenceID: validRefID,
						Metadata: &map[string]interface{}{
							"snapCore": map[string]any{
								"accountName":    vaAccountName,
								"number":         vaAccountNumber,
								"isClosedAmount": true,
								"isSingleUse":    true,
							},
							"bypassStatusPage": true,
							"mode":             constant.UnifiedPaymentModeRedirect,
						},
						PaymentMethod: paymentModel.PaymentMethod{
							Name:     "VA BRI",
							Type:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
							BankName: util.ValueToPtr("BRI"),
						},
						Amount:    decimal.NewFromInt(100000),
						Currency:  "IDR",
						Status:    constant.StatusSuccess,
						CreatedAt: now,
						UpdatedAt: now,
						ExpiredAt: util.ValueToPtr(now),
						DeletedAt: nil,
					}, nil).Once()
				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "valid-merchant-id").
					Return(&merchantModel.Merchant{
						UUID:      "valid-merchant-id",
						ShortName: "Test Merchant",
						KYCStatus: sql.NullString{
							String: constant.KYCStatusApproved,
							Valid:  true,
						},
						ParentID: sql.NullString{
							String: "parent-id",
							Valid:  true,
						},
					}, nil).
					Once()
				mockAccountTransactionRepo.On("FindByReference", mock.Anything, validPaymentID, constant.TypePayment).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						Status:               constant.StatusSuccess,
						Credit:               100000,
						Currency:             "IDR",
						UUID:                 trxID,
						CreatedAt:            now,
						TransactionTimestamp: now,
					}, nil).
					Once()

			},
			wantErr: nil,
			want: &paymentModel.PaymentDetailForPaymentUIResponse{
				UUID:        validPaymentID,
				MerchantID:  "valid-merchant-id",
				ReferenceID: *validRefID,
				PaymentMethod: paymentModel.PaymentMethodDetail{
					Name:     "VA BRI",
					Method:   paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
					Category: paymentConstant.VIRTUAL_ACCOUNT_TRX_TYPE_CLOSED_DYNAMIC,
				},
				Merchant: paymentModel.MerchantDetail{
					Name: "Test Merchant",
					Logo: "",
				},
				TypeDetail: paymentModel.PaymentTypeDetail{
					VirtualAccountName:   &vaAccountName,
					VirtualAccountNumber: &vaAccountNumber,
				},
				Amount: commonModel.Amount{
					Currency: "IDR",
					Value:    "100000.00",
				},
				AmountPaid: commonModel.Amount{
					Currency: "IDR",
					Value:    "100000.00",
				},
				BankReferenceID: "-",
				Channel:         "BRI",
				Status:          constant.StatusSuccess,
				TransactionID:   trxID.String(),
				CreatedAt:       now,
				UpdatedAt:       now,
				ExpiredAt:       &now,
				PaidAt:          &now,
			},
		},
		{
			name:      "when its CC Payment, then should contain the Bank Name as payment channel",
			paymentID: validPaymentID,
			setupMock: func() {
				mockPaymentRepo.On("GetPaymentById", mock.Anything, validPaymentID).
					Return(&paymentModel.Payment{
						UUID:        validPaymentID,
						MerchantID:  "valid-merchant-id",
						ReferenceID: validRefID,
						Metadata: &map[string]interface{}{
							"cardData": map[string]interface{}{
								"cardIssuing": "VISA",
								"last4Digit":  "1234",
								"cardBrand":   "VISA",
								"cardType":    "debit",
							},
						},
						PaymentMethod: paymentModel.PaymentMethod{
							Name: "CC Mastercard",
							Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
						},
						Amount:    decimal.NewFromInt(100000),
						Currency:  "IDR",
						Status:    constant.StatusSuccess,
						CreatedAt: now,
						UpdatedAt: now,
						ExpiredAt: util.ValueToPtr(now),
						DeletedAt: nil,
					}, nil).Once()
				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "valid-merchant-id").
					Return(&merchantModel.Merchant{
						UUID:      "valid-merchant-id",
						ShortName: "Test Merchant",
						KYCStatus: sql.NullString{
							String: constant.KYCStatusApproved,
							Valid:  true,
						},
						ParentID: sql.NullString{
							String: "parent-id",
							Valid:  true,
						},
					}, nil).
					Once()
				mockAccountTransactionRepo.On("FindByReference", mock.Anything, validPaymentID, constant.TypePayment).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						Status:               constant.StatusSuccess,
						Credit:               100000,
						Currency:             "IDR",
						UUID:                 trxID,
						CreatedAt:            now,
						TransactionTimestamp: now,
					}, nil).
					Once()

			},
			wantErr: nil,
			want: &paymentModel.PaymentDetailForPaymentUIResponse{
				UUID:        validPaymentID,
				MerchantID:  "valid-merchant-id",
				ReferenceID: *validRefID,
				PaymentMethod: paymentModel.PaymentMethodDetail{
					Name:     "CC Mastercard",
					Method:   paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
					Category: "debit",
				},
				Merchant: paymentModel.MerchantDetail{
					Name: "Test Merchant",
					Logo: "",
				},
				TypeDetail: paymentModel.PaymentTypeDetail{
					CardIssuer: util.ValueToPtr("VISA"),
					CardNumber: util.ValueToPtr("1234"),
					CardBrand:  "VISA",
				},
				Amount: commonModel.Amount{
					Currency: "IDR",
					Value:    "100000.00",
				},
				AmountPaid: commonModel.Amount{
					Currency: "IDR",
					Value:    "100000.00",
				},
				BankReferenceID: "-",
				Channel:         "VISA",
				Status:          constant.StatusSuccess,
				TransactionID:   trxID.String(),
				CreatedAt:       now,
				UpdatedAt:       now,
				ExpiredAt:       &now,
				PaidAt:          &now,
			},
		},
		{
			name:      "when its CC Payment and the merchant was non-kyc",
			paymentID: validPaymentID,
			setupMock: func() {
				mockPaymentRepo.On("GetPaymentById", mock.Anything, validPaymentID).
					Return(&paymentModel.Payment{
						UUID:        validPaymentID,
						MerchantID:  "valid-merchant-id",
						ReferenceID: validRefID,
						Metadata: &map[string]interface{}{
							"cardData": map[string]interface{}{
								"cardIssuing": "VISA",
								"last4Digit":  "1234",
								"cardBrand":   "VISA",
								"cardType":    "debit",
							},
						},
						PaymentMethod: paymentModel.PaymentMethod{
							Name: "CC Mastercard",
							Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
						},
						Amount:    decimal.NewFromInt(100000),
						Currency:  "IDR",
						Status:    constant.StatusSuccess,
						CreatedAt: now,
						UpdatedAt: now,
						ExpiredAt: util.ValueToPtr(now),
						DeletedAt: nil,
					}, nil).Once()
				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "valid-merchant-id").
					Return(&merchantModel.Merchant{
						UUID:      "valid-merchant-id",
						ShortName: "Test Merchant",
						KYCStatus: sql.NullString{
							String: constant.KYCStatusNotRequired,
							Valid:  true,
						},
						ParentID: sql.NullString{
							String: "parent-id",
							Valid:  true,
						},
					}, nil).
					Once()
				mockAccountTransactionRepo.On("FindByReference", mock.Anything, validPaymentID, constant.TypePayment).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						Status:               constant.StatusSuccess,
						Credit:               100000,
						Currency:             "IDR",
						UUID:                 trxID,
						CreatedAt:            now,
						TransactionTimestamp: now,
					}, nil).
					Once()

			},
			wantErr: nil,
			want: &paymentModel.PaymentDetailForPaymentUIResponse{
				UUID:              validPaymentID,
				MerchantID:        "valid-merchant-id",
				DerivedMerchantID: "parent-id",
				ReferenceID:       *validRefID,
				PaymentMethod: paymentModel.PaymentMethodDetail{
					Name:     "CC Mastercard",
					Method:   paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
					Category: "debit",
				},
				Merchant: paymentModel.MerchantDetail{
					Name: "Test Merchant",
					Logo: "",
				},
				TypeDetail: paymentModel.PaymentTypeDetail{
					CardIssuer: util.ValueToPtr("VISA"),
					CardNumber: util.ValueToPtr("1234"),
					CardBrand:  "VISA",
				},
				Amount: commonModel.Amount{
					Currency: "IDR",
					Value:    "100000.00",
				},
				AmountPaid: commonModel.Amount{
					Currency: "IDR",
					Value:    "100000.00",
				},
				BankReferenceID: "-",
				Channel:         "VISA",
				Status:          constant.StatusSuccess,
				TransactionID:   trxID.String(),
				CreatedAt:       now,
				UpdatedAt:       now,
				ExpiredAt:       &now,
				PaidAt:          &now,
			},
		},
		{
			name:      "when its Wallet Payment, then status should be update session on defer",
			paymentID: validPaymentID,
			setupMock: func() {
				mockPaymentRepo.On("GetPaymentById", mock.Anything, validPaymentID).
					Return(&paymentModel.Payment{
						UUID:        validPaymentID,
						MerchantID:  "valid-merchant-id",
						ReferenceID: validRefID,
						Metadata:    &map[string]interface{}{},
						PaymentMethod: paymentModel.PaymentMethod{
							Type: paymentConstant.PAYMENT_METHOD_EWALLET,
						},
						Amount:    decimal.NewFromInt(100000),
						Currency:  "IDR",
						Status:    constant.UnifiedPaymentSessionStatusRequireAction,
						CreatedAt: now,
						UpdatedAt: now,
						ExpiredAt: util.ValueToPtr(now),
						DeletedAt: nil,
					}, nil).Once()
				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "valid-merchant-id").
					Return(&merchantModel.Merchant{
						UUID:      "valid-merchant-id",
						ShortName: "Test Merchant",
						KYCStatus: sql.NullString{
							String: constant.KYCStatusApproved,
							Valid:  true,
						},
						ParentID: sql.NullString{
							String: "parent-id",
							Valid:  true,
						},
					}, nil).
					Once()
				mockAccountTransactionRepo.On("FindByReference", mock.Anything, validPaymentID, constant.TypePayment).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						Status:               constant.StatusSuccess,
						Credit:               100000,
						Currency:             "IDR",
						UUID:                 trxID,
						CreatedAt:            now,
						TransactionTimestamp: now,
					}, nil).
					Once()

				mockUnifiedPaymentSvc.On("UpdateEWalletPaymentSession", mock.Anything, mock.Anything).Return(&paymentModel.Payment{
					UUID:        validPaymentID,
					MerchantID:  "valid-merchant-id",
					ReferenceID: validRefID,
					Metadata:    &map[string]interface{}{},
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_EWALLET,
					},
					Amount:    decimal.NewFromInt(100000),
					Currency:  "IDR",
					Status:    constant.UnifiedPaymentSessionStatusProcessing,
					CreatedAt: now,
					UpdatedAt: now,
					ExpiredAt: util.ValueToPtr(now),
					DeletedAt: nil,
				}, nil).Once()

			},
			wantErr: nil,
			want: &paymentModel.PaymentDetailForPaymentUIResponse{
				UUID:        validPaymentID,
				MerchantID:  "valid-merchant-id",
				ReferenceID: *validRefID,
				PaymentMethod: paymentModel.PaymentMethodDetail{
					Method: paymentConstant.PAYMENT_METHOD_EWALLET,
				},
				Merchant: paymentModel.MerchantDetail{
					Name: "Test Merchant",
					Logo: "",
				},
				TypeDetail: paymentModel.PaymentTypeDetail{
					EWalletChannel:        "",
					EWalletAppRedirectURL: "",
					EWalletWebRedirectURL: "",
				},
				Amount: commonModel.Amount{
					Currency: "IDR",
					Value:    "100000.00",
				},
				AmountPaid: commonModel.Amount{
					Currency: "IDR",
					Value:    "100000.00",
				},
				BankReferenceID: "-",
				Channel:         "",
				Status:          constant.UnifiedPaymentSessionStatusRequireAction,
				TransactionID:   trxID.String(),
				CreatedAt:       now,
				UpdatedAt:       now,
				ExpiredAt:       &now,
				PaidAt:          &now,
			},
		},
		{
			name:      "when its Wallet Payment, status processing then inquiry to processor",
			paymentID: validPaymentID,
			setupMock: func() {
				mockPaymentRepo.On("GetPaymentById", mock.Anything, validPaymentID).
					Return(&paymentModel.Payment{
						UUID:        validPaymentID,
						MerchantID:  "valid-merchant-id",
						ReferenceID: validRefID,
						Metadata:    &map[string]interface{}{},
						PaymentMethod: paymentModel.PaymentMethod{
							Type: paymentConstant.PAYMENT_METHOD_EWALLET,
						},
						Amount:    decimal.NewFromInt(100000),
						Currency:  "IDR",
						Status:    constant.UnifiedPaymentSessionStatusProcessing,
						CreatedAt: now,
						UpdatedAt: now,
						ExpiredAt: util.ValueToPtr(now),
						DeletedAt: nil,
					}, nil).Once()
				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "valid-merchant-id").
					Return(&merchantModel.Merchant{
						UUID:      "valid-merchant-id",
						ShortName: "Test Merchant",
						KYCStatus: sql.NullString{
							String: constant.KYCStatusApproved,
							Valid:  true,
						},
						ParentID: sql.NullString{
							String: "parent-id",
							Valid:  true,
						},
					}, nil).
					Once()
				mockAccountTransactionRepo.On("FindByReference", mock.Anything, validPaymentID, constant.TypePayment).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						Status:               constant.StatusSuccess,
						Credit:               100000,
						Currency:             "IDR",
						UUID:                 trxID,
						CreatedAt:            now,
						TransactionTimestamp: now,
					}, nil).
					Once()

				mockUnifiedPaymentSvc.On("InquiryEWalletPayment", mock.Anything, mock.Anything).Return(&paymentModel.Payment{
					UUID:        validPaymentID,
					MerchantID:  "valid-merchant-id",
					ReferenceID: validRefID,
					Metadata:    &map[string]interface{}{},
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_EWALLET,
					},
					Amount:    decimal.NewFromInt(100000),
					Currency:  "IDR",
					Status:    constant.UnifiedPaymentSessionStatusProcessing,
					CreatedAt: now,
					UpdatedAt: now,
					ExpiredAt: util.ValueToPtr(now),
					DeletedAt: nil,
					InquiryDetail: &unifiedPaymentModel.InquiryDetail{
						Status: constant.StatusFailed,
					},
				}, nil).Once()

				mockUnifiedPaymentSvc.On("UpdateEWalletPaymentSession", mock.Anything, mock.Anything).Return(&paymentModel.Payment{
					UUID:        validPaymentID,
					MerchantID:  "valid-merchant-id",
					ReferenceID: validRefID,
					Metadata:    &map[string]interface{}{},
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_EWALLET,
					},
					Amount:    decimal.NewFromInt(100000),
					Currency:  "IDR",
					Status:    constant.UnifiedPaymentSessionStatusProcessing,
					CreatedAt: now,
					UpdatedAt: now,
					ExpiredAt: util.ValueToPtr(now),
					DeletedAt: nil,
				}, nil).Once()

			},
			wantErr: nil,
			want: &paymentModel.PaymentDetailForPaymentUIResponse{
				UUID:        validPaymentID,
				MerchantID:  "valid-merchant-id",
				ReferenceID: *validRefID,
				PaymentMethod: paymentModel.PaymentMethodDetail{
					Method: paymentConstant.PAYMENT_METHOD_EWALLET,
				},
				Merchant: paymentModel.MerchantDetail{
					Name: "Test Merchant",
					Logo: "",
				},
				TypeDetail: paymentModel.PaymentTypeDetail{
					EWalletChannel:        "",
					EWalletAppRedirectURL: "",
					EWalletWebRedirectURL: "",
				},
				Amount: commonModel.Amount{
					Currency: "IDR",
					Value:    "100000.00",
				},
				AmountPaid: commonModel.Amount{
					Currency: "IDR",
					Value:    "100000.00",
				},
				BankReferenceID: "-",
				Channel:         "",
				Status:          constant.UnifiedPaymentSessionStatusProcessing,
				TransactionID:   trxID.String(),
				CreatedAt:       now,
				UpdatedAt:       now,
				ExpiredAt:       &now,
				PaidAt:          &now,
				InquiryDetail: &paymentModel.InquiryDetailResponse{
					HasFinalStatus: true,
				},
			},
		},
		{
			name:      "when its Wallet Payment, status processing then got error inquiry to processor",
			paymentID: validPaymentID,
			setupMock: func() {
				mockPaymentRepo.On("GetPaymentById", mock.Anything, validPaymentID).
					Return(&paymentModel.Payment{
						UUID:        validPaymentID,
						MerchantID:  "valid-merchant-id",
						ReferenceID: validRefID,
						Metadata:    &map[string]interface{}{},
						PaymentMethod: paymentModel.PaymentMethod{
							Type: paymentConstant.PAYMENT_METHOD_EWALLET,
						},
						Amount:    decimal.NewFromInt(100000),
						Currency:  "IDR",
						Status:    constant.UnifiedPaymentSessionStatusProcessing,
						CreatedAt: now,
						UpdatedAt: now,
						ExpiredAt: util.ValueToPtr(now),
						DeletedAt: nil,
					}, nil).Once()
				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "valid-merchant-id").
					Return(&merchantModel.Merchant{
						UUID:      "valid-merchant-id",
						ShortName: "Test Merchant",
						KYCStatus: sql.NullString{
							String: constant.KYCStatusApproved,
							Valid:  true,
						},
						ParentID: sql.NullString{
							String: "parent-id",
							Valid:  true,
						},
					}, nil).
					Once()
				mockAccountTransactionRepo.On("FindByReference", mock.Anything, validPaymentID, constant.TypePayment).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						Status:               constant.StatusSuccess,
						Credit:               100000,
						Currency:             "IDR",
						UUID:                 trxID,
						CreatedAt:            now,
						TransactionTimestamp: now,
					}, nil).
					Once()

				mockUnifiedPaymentSvc.On("InquiryEWalletPayment", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("error")).Once()

				mockUnifiedPaymentSvc.On("UpdateEWalletPaymentSession", mock.Anything, mock.Anything).Return(&paymentModel.Payment{
					UUID:        validPaymentID,
					MerchantID:  "valid-merchant-id",
					ReferenceID: validRefID,
					Metadata:    &map[string]interface{}{},
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_EWALLET,
					},
					Amount:    decimal.NewFromInt(100000),
					Currency:  "IDR",
					Status:    constant.UnifiedPaymentSessionStatusProcessing,
					CreatedAt: now,
					UpdatedAt: now,
					ExpiredAt: util.ValueToPtr(now),
					DeletedAt: nil,
				}, nil).Once()

			},
			wantErr: nil,
			want: &paymentModel.PaymentDetailForPaymentUIResponse{
				UUID:        validPaymentID,
				MerchantID:  "valid-merchant-id",
				ReferenceID: *validRefID,
				PaymentMethod: paymentModel.PaymentMethodDetail{
					Method: paymentConstant.PAYMENT_METHOD_EWALLET,
				},
				Merchant: paymentModel.MerchantDetail{
					Name: "Test Merchant",
					Logo: "",
				},
				TypeDetail: paymentModel.PaymentTypeDetail{
					EWalletChannel:        "",
					EWalletAppRedirectURL: "",
					EWalletWebRedirectURL: "",
				},
				Amount: commonModel.Amount{
					Currency: "IDR",
					Value:    "100000.00",
				},
				AmountPaid: commonModel.Amount{
					Currency: "IDR",
					Value:    "100000.00",
				},
				BankReferenceID: "-",
				Channel:         "",
				Status:          constant.UnifiedPaymentSessionStatusProcessing,
				TransactionID:   trxID.String(),
				CreatedAt:       now,
				UpdatedAt:       now,
				ExpiredAt:       &now,
				PaidAt:          &now,
			},
		},
		{
			name:      "when its Wallet Payment, failed update status to PROCESSING should not break the flow",
			paymentID: validPaymentID,
			setupMock: func() {
				mockPaymentRepo.On("GetPaymentById", mock.Anything, validPaymentID).
					Return(&paymentModel.Payment{
						UUID:        validPaymentID,
						MerchantID:  "valid-merchant-id",
						ReferenceID: validRefID,
						Metadata:    &map[string]interface{}{},
						PaymentMethod: paymentModel.PaymentMethod{
							Type: paymentConstant.PAYMENT_METHOD_EWALLET,
						},
						Amount:    decimal.NewFromInt(100000),
						Currency:  "IDR",
						Status:    constant.UnifiedPaymentSessionStatusRequireAction,
						CreatedAt: now,
						UpdatedAt: now,
						ExpiredAt: util.ValueToPtr(now),
						DeletedAt: nil,
					}, nil).Once()
				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "valid-merchant-id").
					Return(&merchantModel.Merchant{
						UUID:      "valid-merchant-id",
						ShortName: "Test Merchant",
						KYCStatus: sql.NullString{
							String: constant.KYCStatusApproved,
							Valid:  true,
						},
						ParentID: sql.NullString{
							String: "parent-id",
							Valid:  true,
						},
					}, nil).
					Once()
				mockAccountTransactionRepo.On("FindByReference", mock.Anything, validPaymentID, constant.TypePayment).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						Status:    constant.StatusSuccess,
						Credit:    100000,
						Currency:  "IDR",
						UUID:      trxID,
						CreatedAt: now,
					}, nil).
					Once()

				mockUnifiedPaymentSvc.On("UpdateEWalletPaymentSession", mock.Anything, mock.Anything).Return(nil, pkgErrors.New(response.HttpErrDatabase, constant.ErrUpdatePayment)).Once()

			},
			wantErr: nil,
			want: &paymentModel.PaymentDetailForPaymentUIResponse{
				UUID:        validPaymentID,
				MerchantID:  "valid-merchant-id",
				ReferenceID: *validRefID,
				PaymentMethod: paymentModel.PaymentMethodDetail{
					Method: paymentConstant.PAYMENT_METHOD_EWALLET,
				},
				Merchant: paymentModel.MerchantDetail{
					Name: "Test Merchant",
					Logo: "",
				},
				TypeDetail: paymentModel.PaymentTypeDetail{
					EWalletChannel:        "",
					EWalletAppRedirectURL: "",
					EWalletWebRedirectURL: "",
				},
				Amount: commonModel.Amount{
					Currency: "IDR",
					Value:    "100000.00",
				},
				AmountPaid: commonModel.Amount{
					Currency: "IDR",
					Value:    "100000.00",
				},
				BankReferenceID: "-",
				Channel:         "",
				Status:          constant.UnifiedPaymentSessionStatusRequireAction,
				TransactionID:   trxID.String(),
				CreatedAt:       now,
				UpdatedAt:       now,
				ExpiredAt:       &now,
				PaidAt:          &time.Time{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			svc := New(mockPaymentRepo, mockLogger, nil, nil, mockMerchantRepo, nil, nil, WithAccountTransactionRepository(mockAccountTransactionRepo))
			svc.unifiedPaymentSvc = mockUnifiedPaymentSvc
			got, err := svc.GetPaymentDetailForPaymentUI(context.Background(), tc.paymentID)

			if tc.wantErr != nil {
				assert.EqualError(t, err, tc.wantErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.want.UUID, got.UUID)
				assert.Equal(t, tc.want.TypeDetail, got.TypeDetail)
				assert.Equal(t, tc.want.Amount, got.Amount)
				assert.Equal(t, tc.want.AmountPaid, got.AmountPaid)
				assert.Equal(t, tc.want.BankReferenceID, got.BankReferenceID)
				assert.Equal(t, tc.want.Channel, got.Channel)
				assert.Equal(t, tc.want.CreatedAt, got.CreatedAt)
				assert.Equal(t, tc.want.CustomerID, got.CustomerID)
				assert.Equal(t, tc.want.ExpiredAt, got.ExpiredAt)
				assert.Equal(t, tc.want.FdsRiskAssessment, got.FdsRiskAssessment)
				assert.Equal(t, tc.want.Merchant, got.Merchant)
				assert.Equal(t, tc.want.MerchantID, got.MerchantID)
				assert.Equal(t, tc.want.PaidAt, got.PaidAt)
				assert.Equal(t, tc.want.PaymentMethod, got.PaymentMethod)
				assert.Equal(t, tc.want.ProcessorRefNumber, got.ProcessorRefNumber)
				assert.Equal(t, tc.want.RedirectUrl, got.RedirectUrl)
				assert.Equal(t, tc.want.ReferenceID, got.ReferenceID)
				assert.Equal(t, tc.want.Status, got.Status)
				assert.Equal(t, tc.want.TransactionID, got.TransactionID)
				assert.Equal(t, tc.want.TypeDetail, got.TypeDetail)

			}

			mockPaymentRepo.AssertExpectations(t)
			mockMerchantRepo.AssertExpectations(t)
			mockAccountTransactionRepo.AssertExpectations(t)
		})
	}
}

func TestPaymentService_shouldShowAuthCaptureBanner(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name               string
		payment            *paymentModel.Payment
		accountTransaction *orchestratorModel.AccountTransactionWithUseCase
		want               bool
	}{
		{
			name: "SUCCESS - Credit card payment with WAITING_FOR_CAPTURE status should show banner",
			payment: &paymentModel.Payment{
				Status: constant.UnifiedPaymentSessionStatusProcessing,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				},
			},
			accountTransaction: &orchestratorModel.AccountTransactionWithUseCase{
				Status: constant.ChargeStatusWaitingForCapture,
			},
			want: true,
		},
		{
			name: "SUCCESS - Credit card payment with chargeStatus in additionalInfo should show banner",
			payment: &paymentModel.Payment{
				Status: constant.UnifiedPaymentSessionStatusProcessing,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				},
			},
			accountTransaction: &orchestratorModel.AccountTransactionWithUseCase{
				Status: constant.StatusSuccess,
				AdditionalInfo: types.NullJSONText{
					Valid:    true,
					JSONText: []byte(`{"chargeStatus":"WAITING_FOR_CAPTURE"}`),
				},
			},
			want: true,
		},
		{
			name: "SUCCESS - chargeStatus in additionalInfo overrides accountTransaction.Status",
			payment: &paymentModel.Payment{
				Status: constant.UnifiedPaymentSessionStatusProcessing,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				},
			},
			accountTransaction: &orchestratorModel.AccountTransactionWithUseCase{
				Status: constant.StatusPending,
				AdditionalInfo: types.NullJSONText{
					Valid:    true,
					JSONText: []byte(`{"chargeStatus":"WAITING_FOR_CAPTURE","otherField":"value"}`),
				},
			},
			want: true,
		},
		{
			name: "FAILED - chargeStatus is nil, should use accountTransaction.Status",
			payment: &paymentModel.Payment{
				Status: constant.UnifiedPaymentSessionStatusProcessing,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				},
			},
			accountTransaction: &orchestratorModel.AccountTransactionWithUseCase{
				Status: constant.StatusPending,
				AdditionalInfo: types.NullJSONText{
					Valid:    true,
					JSONText: []byte(`{"chargeStatus":null}`),
				},
			},
			want: false,
		},
		{
			name: "FAILED - chargeStatus is not a string, should use accountTransaction.Status",
			payment: &paymentModel.Payment{
				Status: constant.UnifiedPaymentSessionStatusProcessing,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				},
			},
			accountTransaction: &orchestratorModel.AccountTransactionWithUseCase{
				Status: constant.StatusSuccess,
				AdditionalInfo: types.NullJSONText{
					Valid:    true,
					JSONText: []byte(`{"chargeStatus":123}`),
				},
			},
			want: false,
		},
		{
			name: "FAILED - chargeStatus key doesn't exist, should use accountTransaction.Status",
			payment: &paymentModel.Payment{
				Status: constant.UnifiedPaymentSessionStatusProcessing,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				},
			},
			accountTransaction: &orchestratorModel.AccountTransactionWithUseCase{
				Status: constant.StatusFailed,
				AdditionalInfo: types.NullJSONText{
					Valid:    true,
					JSONText: []byte(`{"otherField":"value"}`),
				},
			},
			want: false,
		},
		{
			name: "FAILED - AdditionalInfo.Valid is false, should use accountTransaction.Status",
			payment: &paymentModel.Payment{
				Status: constant.UnifiedPaymentSessionStatusProcessing,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				},
			},
			accountTransaction: &orchestratorModel.AccountTransactionWithUseCase{
				Status: constant.StatusPending,
				AdditionalInfo: types.NullJSONText{
					Valid:    false,
					JSONText: []byte(`{"chargeStatus":"WAITING_FOR_CAPTURE"}`),
				},
			},
			want: false,
		},
		{
			name: "FAILED - Invalid JSON in AdditionalInfo, should use accountTransaction.Status",
			payment: &paymentModel.Payment{
				Status: constant.UnifiedPaymentSessionStatusProcessing,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				},
			},
			accountTransaction: &orchestratorModel.AccountTransactionWithUseCase{
				Status: constant.StatusFailed,
				AdditionalInfo: types.NullJSONText{
					Valid:    true,
					JSONText: []byte(`{invalid json}`),
				},
			},
			want: false,
		},
		{
			name: "FAILED - Credit card payment with SUCCESS status should not show banner",
			payment: &paymentModel.Payment{
				Status: constant.StatusSuccess,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				},
			},
			accountTransaction: &orchestratorModel.AccountTransactionWithUseCase{
				Status: constant.ChargeStatusWaitingForCapture,
			},
			want: false,
		},
		{
			name: "FAILED - Credit card payment with PENDING status should not show banner",
			payment: &paymentModel.Payment{
				Status: constant.StatusPending,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				},
			},
			accountTransaction: &orchestratorModel.AccountTransactionWithUseCase{
				Status: constant.ChargeStatusWaitingForCapture,
			},
			want: false,
		},
		{
			name: "FAILED - Credit card payment with FAILED status should not show banner",
			payment: &paymentModel.Payment{
				Status: constant.StatusFailed,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				},
			},
			accountTransaction: &orchestratorModel.AccountTransactionWithUseCase{
				Status: constant.ChargeStatusWaitingForCapture,
			},
			want: false,
		},
		{
			name: "FAILED - Virtual account payment with WAITING_FOR_CAPTURE status should not show banner",
			payment: &paymentModel.Payment{
				Status: constant.ChargeStatusWaitingForCapture,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				},
			},
			accountTransaction: &orchestratorModel.AccountTransactionWithUseCase{
				Status: constant.ChargeStatusWaitingForCapture,
			},
			want: false,
		},
		{
			name: "FAILED - QRIS payment with WAITING_FOR_CAPTURE status should not show banner",
			payment: &paymentModel.Payment{
				Status: constant.ChargeStatusWaitingForCapture,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_QRIS,
				},
			},
			accountTransaction: &orchestratorModel.AccountTransactionWithUseCase{
				Status: constant.ChargeStatusWaitingForCapture,
			},
			want: false,
		},
		{
			name: "FAILED - Ewallet payment with WAITING_FOR_CAPTURE status should not show banner",
			payment: &paymentModel.Payment{
				Status: constant.ChargeStatusWaitingForCapture,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_EWALLET,
				},
			},
			accountTransaction: &orchestratorModel.AccountTransactionWithUseCase{
				Status: constant.ChargeStatusWaitingForCapture,
			},
			want: false,
		},
		{
			name: "FAILED - Bank transfer payment with WAITING_FOR_CAPTURE status should not show banner",
			payment: &paymentModel.Payment{
				Status: constant.ChargeStatusWaitingForCapture,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_BANK_TRANSFER,
				},
			},
			accountTransaction: &orchestratorModel.AccountTransactionWithUseCase{
				Status: constant.ChargeStatusWaitingForCapture,
			},
			want: false,
		},
		{
			name: "FAILED - Virtual account with SUCCESS status should not show banner",
			payment: &paymentModel.Payment{
				Status: constant.StatusSuccess,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				},
			},
			accountTransaction: &orchestratorModel.AccountTransactionWithUseCase{
				Status: constant.ChargeStatusWaitingForCapture,
			},
			want: false,
		},
	}

	svc := PaymentService{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.shouldShowAuthCaptureBanner(ctx, tt.payment, tt.accountTransaction)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPaymentService_getChargeStatusFromTransaction(t *testing.T) {
	tests := []struct {
		name               string
		accountTransaction *orchestratorModel.AccountTransactionWithUseCase
		want               string
	}{
		{
			name: "SUCCESS - Returns accountTransaction.Status when additionalInfo is not valid",
			accountTransaction: &orchestratorModel.AccountTransactionWithUseCase{
				Status: constant.StatusSuccess,
				AdditionalInfo: types.NullJSONText{
					Valid: false,
				},
			},
			want: constant.StatusSuccess,
		},
		{
			name: "SUCCESS - Returns chargeStatus from additionalInfo when available",
			accountTransaction: &orchestratorModel.AccountTransactionWithUseCase{
				Status: constant.StatusSuccess,
				AdditionalInfo: types.NullJSONText{
					Valid:    true,
					JSONText: []byte(`{"chargeStatus":"WAITING_FOR_CAPTURE"}`),
				},
			},
			want: constant.ChargeStatusWaitingForCapture,
		},
		{
			name: "SUCCESS - chargeStatus overrides accountTransaction.Status",
			accountTransaction: &orchestratorModel.AccountTransactionWithUseCase{
				Status: constant.StatusPending,
				AdditionalInfo: types.NullJSONText{
					Valid:    true,
					JSONText: []byte(`{"chargeStatus":"WAITING_FOR_CAPTURE","otherField":"value"}`),
				},
			},
			want: constant.ChargeStatusWaitingForCapture,
		},
		{
			name: "SUCCESS - Returns accountTransaction.Status when chargeStatus is nil",
			accountTransaction: &orchestratorModel.AccountTransactionWithUseCase{
				Status: constant.StatusPending,
				AdditionalInfo: types.NullJSONText{
					Valid:    true,
					JSONText: []byte(`{"chargeStatus":null}`),
				},
			},
			want: constant.StatusPending,
		},
		{
			name: "SUCCESS - Returns accountTransaction.Status when chargeStatus is not a string",
			accountTransaction: &orchestratorModel.AccountTransactionWithUseCase{
				Status: constant.StatusSuccess,
				AdditionalInfo: types.NullJSONText{
					Valid:    true,
					JSONText: []byte(`{"chargeStatus":123}`),
				},
			},
			want: constant.StatusSuccess,
		},
		{
			name: "SUCCESS - Returns accountTransaction.Status when chargeStatus key doesn't exist",
			accountTransaction: &orchestratorModel.AccountTransactionWithUseCase{
				Status: constant.StatusFailed,
				AdditionalInfo: types.NullJSONText{
					Valid:    true,
					JSONText: []byte(`{"otherField":"value"}`),
				},
			},
			want: constant.StatusFailed,
		},
		{
			name: "SUCCESS - Returns accountTransaction.Status when JSON is invalid",
			accountTransaction: &orchestratorModel.AccountTransactionWithUseCase{
				Status: constant.StatusFailed,
				AdditionalInfo: types.NullJSONText{
					Valid:    true,
					JSONText: []byte(`{invalid json}`),
				},
			},
			want: constant.StatusFailed,
		},
		{
			name:               "SUCCESS - Returns empty string when accountTransaction is nil",
			accountTransaction: nil,
			want:               "",
		},
	}

	svc := PaymentService{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.getChargeStatusFromTransaction(tt.accountTransaction)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPaymentServiceGetCardFundedPayoutMetadata(t *testing.T) {
	referenceID := "payout-reference-123"
	merchantID := "merchant-id-123"
	paymentUUID := "payment-uuid-123"

	tests := []struct {
		name       string
		payment    *paymentModel.Payment
		setupMocks func(disbursementRepo *repositoryMocks.IDisbursementRepository)
		want       *paymentModel.CardFundedPayoutMetadata
	}{
		{
			name:    "FAILED: payment is nil",
			payment: nil,
			setupMocks: func(disbursementRepo *repositoryMocks.IDisbursementRepository) {
				// No mock setup needed since early return
			},
			want: nil,
		},
		{
			name: "FAILED: repository returns error",
			payment: &paymentModel.Payment{
				UUID:        paymentUUID,
				ReferenceID: &referenceID,
				MerchantID:  merchantID,
			},
			setupMocks: func(disbursementRepo *repositoryMocks.IDisbursementRepository) {
				disbursementRepo.
					On(
						"GetCardFundedPayoutDetail",
						mock.Anything,
						mock.Anything).
					Return(nil, assert.AnError)
			},
			want: nil,
		},
		{
			name: "FAILED: card funded payout not found",
			payment: &paymentModel.Payment{
				UUID:        paymentUUID,
				ReferenceID: &referenceID,
				MerchantID:  merchantID,
			},
			setupMocks: func(disbursementRepo *repositoryMocks.IDisbursementRepository) {
				disbursementRepo.
					On(
						"GetCardFundedPayoutDetail",
						mock.Anything,
						mock.Anything).
					Return(nil, nil)
			},
			want: nil,
		},
		{
			name: "SUCCESS: returns card funded payout metadata",
			payment: &paymentModel.Payment{
				UUID:        paymentUUID,
				ReferenceID: &referenceID,
				MerchantID:  merchantID,
			},
			setupMocks: func(disbursementRepo *repositoryMocks.IDisbursementRepository) {
				createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
				cardFundedPayout := &cardFundedPayoutModel.GetPayoutDetailResponse{
					UUID:          "payout-uuid-123",
					ReferenceID:   "ref-123",
					VendorName:    "Vendor ABC",
					AccountNumber: "1234567890",
					AccountName:   "John Doe",
					BankName:      "Bank BCA",
					Remarks:       "Test payout",
					Amount:        "100000.00",
					Fee:           "5000.00",
					TotalAmount:   "105000.00",
					CreatedAt:     createdAt,
					MetadataObj: disbursementModel.Metadata{
						CardFundedDetail: &disbursementModel.CardFundedDetailMetadata{
							VendorID:   "vendor-123",
							VendorName: "Vendor ABC",
						},
					},
				}
				disbursementRepo.
					On(
						"GetCardFundedPayoutDetail",
						mock.Anything,
						&cardFundedPayoutModel.GetPayoutDetailRequest{
							PayoutID:   referenceID,
							MerchantID: merchantID,
						}).
					Return(cardFundedPayout, nil)
			},
			want: &paymentModel.CardFundedPayoutMetadata{
				PayoutID:          "payout-uuid-123",
				ReferenceID:       "ref-123",
				VendorName:        "Vendor ABC",
				BankAccountNumber: "1234567890",
				BankAccountName:   "John Doe",
				BankName:          "Bank BCA",
				Remarks:           "Test payout",
				Amount:            "100000.00",
				Fee:               "5000.00",
				TotalAmount:       "105000.00",
				CreatedAt:         time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDisbursementRepo := repositoryMocks.NewIDisbursementRepository(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tt.setupMocks(mockDisbursementRepo)

			svc := &PaymentService{
				disbursementRepo: mockDisbursementRepo,
				logger:           mockLogger,
			}

			got := svc.getCardFundedPayoutMetadata(context.Background(), tt.payment)
			assert.Equal(t, tt.want, got)

			mockDisbursementRepo.AssertExpectations(t)
		})
	}
}
