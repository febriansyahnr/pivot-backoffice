package settlementService_test

import (
	"bytes"
	"context"
	"database/sql"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestraModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	settlementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/settlement"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/settlement"
	loggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	repositoryMock "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	logger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProcessPaymentSettlement(t *testing.T) {
	buf := new(bytes.Buffer)
	defer buf.Reset()

	log := logger.NewSlogger(logger.Config{}, logger.WithSlogOutput(buf))

	transactionID := "cd1f4928-5bd2-4a76-bdc2-564fcc5dbfa7"
	feeTransactionID := "010d8f69-ea7d-4e15-abc7-7038f50fb3bc"
	merchantID := "9f60d59b-3d38-46a8-9709-5c73f140d24a"

	validTransaction := &orchestraModel.AccountTransactionWithUseCase{
		UUID:             util.ParseUUID(transactionID),
		MerchantID:       util.ParseUUID(merchantID),
		AdditionalInfo:   types.NullJSONText{Valid: true},
		SettlementStatus: sql.NullString{String: constant.SettlementStatusPending, Valid: true},
		Type:             constant.TypePayment,
		ReferenceID:      "payment-ref-123",
	}

	validMerchantTransaction := &orchestraModel.AccountTransactionWithUseCase{
		UUID:             util.ParseUUID(transactionID),
		MerchantID:       util.ParseUUID(merchantID),
		AdditionalInfo:   types.NullJSONText{Valid: true},
		SettlementStatus: sql.NullString{String: constant.SettlementStatusPending, Valid: true},
		Type:             constant.TypeMerchantPayment,
		ReferenceID:      "merchant-payment-ref-123",
	}

	validCardFundedPayoutTransaction := &orchestraModel.AccountTransactionWithUseCase{
		UUID:             util.ParseUUID(transactionID),
		MerchantID:       util.ParseUUID(merchantID),
		AdditionalInfo:   types.NullJSONText{Valid: true},
		SettlementStatus: sql.NullString{String: constant.SettlementStatusPending, Valid: true},
		Type:             constant.TypePayment,
		Reference:        constant.ReferencePaymentFundedPayout,
		ReferenceID:      "card-funded-payout-ref-123",
	}

	testCases := []struct {
		name                 string
		byPassSettlementHold bool
		wantErr              bool
		setupMock            func(paymentSvc *serviceMock.IPaymentService, merchantSvc *serviceMock.IMerchantService, accountTransactionRepo *repositoryMock.IAccountTransactionRepository, cardFundedPayoutSvc *serviceMock.ICardFundedPayoutService)
		expectedError        string
	}{
		{
			name:    "ERROR: Transaction error",
			wantErr: true,
			setupMock: func(paymentSvc *serviceMock.IPaymentService, merchantSvc *serviceMock.IMerchantService, accountTransactionRepo *repositoryMock.IAccountTransactionRepository, cardFundedPayoutSvc *serviceMock.ICardFundedPayoutService) {
				accountTransactionRepo.On("FindByID", mock.Anything, transactionID).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			expectedError: constant.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name:    "ERROR: Transaction not found",
			wantErr: true,
			setupMock: func(paymentSvc *serviceMock.IPaymentService, merchantSvc *serviceMock.IMerchantService, accountTransactionRepo *repositoryMock.IAccountTransactionRepository, cardFundedPayoutSvc *serviceMock.ICardFundedPayoutService) {
				accountTransactionRepo.On("FindByID", constant.ValueCtxMockType(), transactionID).Once().Return(nil, nil)
			},
			expectedError: "ERROR_UNPROCESSABLE_CONTENT | payment not found",
		},
		{
			name:    "ERROR: Transaction merchantID not match",
			wantErr: true,
			setupMock: func(paymentSvc *serviceMock.IPaymentService, merchantSvc *serviceMock.IMerchantService, accountTransactionRepo *repositoryMock.IAccountTransactionRepository, cardFundedPayoutSvc *serviceMock.ICardFundedPayoutService) {
				accountTransactionRepo.On("FindByID",
					constant.ValueCtxMockType(), constant.StringMockType(),
				).Once().Return(&orchestraModel.AccountTransactionWithUseCase{
					MerchantID: uuid.New(),
				}, nil)
			},
			expectedError: "ERROR_REQUEST | merchant id is not match",
		},
		{
			name:    "ERROR: BeginTransaction",
			wantErr: true,
			setupMock: func(paymentSvc *serviceMock.IPaymentService, merchantSvc *serviceMock.IMerchantService, accountTransactionRepo *repositoryMock.IAccountTransactionRepository, cardFundedPayoutSvc *serviceMock.ICardFundedPayoutService) {
				accountTransactionRepo.On("FindByID", mock.Anything, transactionID).Once().Return(validTransaction, nil)
				accountTransactionRepo.On("BeginTransaction", mock.Anything).Once().Return(context.Background(), constant.ErrSomeErrorForUnitTest)
			},
			expectedError: constant.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name:    "ERROR: RollbackTransaction",
			wantErr: true,
			setupMock: func(paymentSvc *serviceMock.IPaymentService, merchantSvc *serviceMock.IMerchantService, accountTransactionRepo *repositoryMock.IAccountTransactionRepository, cardFundedPayoutSvc *serviceMock.ICardFundedPayoutService) {
				accountTransactionRepo.On("FindByID", mock.Anything, transactionID).Once().Return(validMerchantTransaction, nil)
				accountTransactionRepo.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				accountTransactionRepo.On("UpdatePaymentTransactionStatusAndMetadataByID", mock.Anything, mock.Anything, mock.Anything).Once().Return(constant.ErrSomeErrorForUnitTest)
				accountTransactionRepo.On("RollbackTransaction", mock.Anything).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			expectedError: pkgErr.New(response.HttpErrDatabase, constant.ErrSomeErrorForUnitTest).Error(),
		},
		{
			name:    "ERROR: Update payment not found",
			wantErr: true,
			setupMock: func(paymentSvc *serviceMock.IPaymentService, merchantSvc *serviceMock.IMerchantService, accountTransactionRepo *repositoryMock.IAccountTransactionRepository, cardFundedPayoutSvc *serviceMock.ICardFundedPayoutService) {
				accountTransactionRepo.On("FindByID", mock.Anything, transactionID).Once().Return(validMerchantTransaction, nil)
				accountTransactionRepo.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				accountTransactionRepo.On("UpdatePaymentTransactionStatusAndMetadataByID", mock.Anything, mock.Anything, mock.Anything).Once().Return(constant.ErrDataNotFound)
				accountTransactionRepo.On("RollbackTransaction", mock.Anything).Return(nil)
			},
			expectedError: pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrDataNotFound).Error(),
		},
		{
			name:    "ERROR: Merchant Error",
			wantErr: true,
			setupMock: func(paymentSvc *serviceMock.IPaymentService, merchantSvc *serviceMock.IMerchantService, accountTransactionRepo *repositoryMock.IAccountTransactionRepository, cardFundedPayoutSvc *serviceMock.ICardFundedPayoutService) {
				accountTransactionRepo.On("FindByID", mock.Anything, transactionID).Times(1).Return(validMerchantTransaction, nil)
				accountTransactionRepo.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				accountTransactionRepo.On("UpdatePaymentTransactionStatusAndMetadataByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
				accountTransactionRepo.On("RollbackTransaction", mock.Anything).Return(nil)
			},
			expectedError: constant.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name:    "ERROR: Merchant Not Found",
			wantErr: true,
			setupMock: func(paymentSvc *serviceMock.IPaymentService, merchantSvc *serviceMock.IMerchantService, accountTransactionRepo *repositoryMock.IAccountTransactionRepository, cardFundedPayoutSvc *serviceMock.ICardFundedPayoutService) {
				accountTransactionRepo.On("FindByID", mock.Anything, transactionID).Times(1).Return(validMerchantTransaction, nil)
				accountTransactionRepo.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				accountTransactionRepo.On("UpdatePaymentTransactionStatusAndMetadataByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Once().Return(nil, nil)
				accountTransactionRepo.On("RollbackTransaction", mock.Anything).Return(nil)
			},
			expectedError: pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrMerchantNotFound).Error(),
		},
		{
			name:    "ERROR: Commit transaction main settlement",
			wantErr: true,
			setupMock: func(paymentSvc *serviceMock.IPaymentService, merchantSvc *serviceMock.IMerchantService, accountTransactionRepo *repositoryMock.IAccountTransactionRepository, cardFundedPayoutSvc *serviceMock.ICardFundedPayoutService) {
				accountTransactionRepo.On("FindByID", mock.Anything, transactionID).Times(1).Return(validTransaction, nil)
				accountTransactionRepo.On("BeginTransaction", mock.Anything).Once().Return(context.Background(), nil)
				accountTransactionRepo.On("UpdatePaymentTransactionStatusAndMetadataByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Times(1).Return(&merchant.Merchant{ParentID: sql.NullString{Valid: true, String: uuid.NewString()}}, nil)
				accountTransactionRepo.On("FindByID", mock.Anything, feeTransactionID).Times(1).Return(nil, nil)
				accountTransactionRepo.On("CommitTransaction", mock.Anything).Once().Return(constant.ErrSomeErrorForUnitTest)
				accountTransactionRepo.On("RollbackTransaction", mock.Anything).Return(nil)
			},
			expectedError: pkgErr.New(response.HttpErrDatabase, constant.ErrSomeErrorForUnitTest).Error(),
		},
		{
			name:    "ERROR: Split route begin transaction",
			wantErr: false, // Main settlement succeeds, split route failure doesn't fail overall process
			setupMock: func(paymentSvc *serviceMock.IPaymentService, merchantSvc *serviceMock.IMerchantService, accountTransactionRepo *repositoryMock.IAccountTransactionRepository, cardFundedPayoutSvc *serviceMock.ICardFundedPayoutService) {
				// Main settlement flow
				accountTransactionRepo.On("FindByID", mock.Anything, transactionID).Times(1).Return(validTransaction, nil)
				accountTransactionRepo.On("BeginTransaction", mock.Anything).Once().Return(context.Background(), nil)
				accountTransactionRepo.On("UpdatePaymentTransactionStatusAndMetadataByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Times(1).Return(&merchant.Merchant{ParentID: sql.NullString{Valid: true, String: uuid.NewString()}}, nil)
				accountTransactionRepo.On("FindByID", mock.Anything, feeTransactionID).Times(1).Return(nil, nil)
				accountTransactionRepo.On("CommitTransaction", mock.Anything).Once().Return(nil)
				// Split route fails at BeginTransaction
				accountTransactionRepo.On("BeginTransaction", mock.Anything).Once().Return(context.Background(), constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: Split route process error",
			wantErr: false, // Main settlement succeeds, split route failure doesn't fail overall process
			setupMock: func(paymentSvc *serviceMock.IPaymentService, merchantSvc *serviceMock.IMerchantService, accountTransactionRepo *repositoryMock.IAccountTransactionRepository, cardFundedPayoutSvc *serviceMock.ICardFundedPayoutService) {
				accountTransactionRepo.On("FindByID", mock.Anything, transactionID).Times(1).Return(validTransaction, nil)
				accountTransactionRepo.On("BeginTransaction", mock.Anything).Once().Return(context.Background(), nil)
				accountTransactionRepo.On("UpdatePaymentTransactionStatusAndMetadataByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Times(1).Return(&merchant.Merchant{ParentID: sql.NullString{Valid: true, String: uuid.NewString()}}, nil)
				accountTransactionRepo.On("FindByID", mock.Anything, feeTransactionID).Times(1).Return(nil, nil)
				accountTransactionRepo.On("CommitTransaction", mock.Anything).Once().Return(nil)
				accountTransactionRepo.On("BeginTransaction", mock.Anything).Once().Return(context.Background(), nil)
				paymentSvc.On("ProcessSplitRoute", mock.Anything, mock.Anything).Once().Return(constant.ErrSomeErrorForUnitTest)
				accountTransactionRepo.On("RollbackTransaction", mock.Anything).Once().Return(nil)
			},
		},
		{
			name:    "ERROR: Split route commit transaction",
			wantErr: false, // Main settlement succeeds, split route failure doesn't fail overall process
			setupMock: func(paymentSvc *serviceMock.IPaymentService, merchantSvc *serviceMock.IMerchantService, accountTransactionRepo *repositoryMock.IAccountTransactionRepository, cardFundedPayoutSvc *serviceMock.ICardFundedPayoutService) {
				accountTransactionRepo.On("FindByID", mock.Anything, transactionID).Times(1).Return(validTransaction, nil)
				accountTransactionRepo.On("BeginTransaction", mock.Anything).Once().Return(context.Background(), nil)
				accountTransactionRepo.On("UpdatePaymentTransactionStatusAndMetadataByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Times(1).Return(&merchant.Merchant{ParentID: sql.NullString{Valid: true, String: uuid.NewString()}}, nil)
				accountTransactionRepo.On("FindByID", mock.Anything, feeTransactionID).Times(1).Return(nil, nil)
				accountTransactionRepo.On("CommitTransaction", mock.Anything).Once().Return(nil)
				accountTransactionRepo.On("BeginTransaction", mock.Anything).Once().Return(context.Background(), nil)
				paymentSvc.On("ProcessSplitRoute", mock.Anything, mock.Anything).Once().Return(nil)
				accountTransactionRepo.On("CommitTransaction", mock.Anything).Once().Return(constant.ErrSomeErrorForUnitTest)
				accountTransactionRepo.On("RollbackTransaction", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "SUCCESS: Complete settlement with split route",
			setupMock: func(paymentSvc *serviceMock.IPaymentService, merchantSvc *serviceMock.IMerchantService, accountTransactionRepo *repositoryMock.IAccountTransactionRepository, cardFundedPayoutSvc *serviceMock.ICardFundedPayoutService) {
				accountTransactionRepo.On("FindByID", mock.Anything, transactionID).Times(1).Return(validTransaction, nil)
				accountTransactionRepo.On("BeginTransaction", mock.Anything).Once().Return(context.Background(), nil)
				accountTransactionRepo.On("UpdatePaymentTransactionStatusAndMetadataByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Times(1).Return(&merchant.Merchant{ParentID: sql.NullString{Valid: true, String: uuid.NewString()}}, nil)
				accountTransactionRepo.On("FindByID", mock.Anything, feeTransactionID).Times(1).Return(nil, nil)
				accountTransactionRepo.On("CommitTransaction", mock.Anything).Once().Return(nil)
				accountTransactionRepo.On("BeginTransaction", mock.Anything).Once().Return(context.Background(), nil)
				paymentSvc.On("ProcessSplitRoute", mock.Anything, mock.Anything).Once().Return(nil)
				accountTransactionRepo.On("CommitTransaction", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "SUCCESS: Merchant payment settlement without split route",
			setupMock: func(paymentSvc *serviceMock.IPaymentService, merchantSvc *serviceMock.IMerchantService, accountTransactionRepo *repositoryMock.IAccountTransactionRepository, cardFundedPayoutSvc *serviceMock.ICardFundedPayoutService) {
				accountTransactionRepo.On("FindByID", mock.Anything, transactionID).Times(1).Return(validMerchantTransaction, nil)
				accountTransactionRepo.On("BeginTransaction", mock.Anything).Once().Return(context.Background(), nil)
				accountTransactionRepo.On("UpdatePaymentTransactionStatusAndMetadataByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Times(1).Return(&merchant.Merchant{ParentID: sql.NullString{Valid: true, String: uuid.NewString()}}, nil)
				accountTransactionRepo.On("FindByID", mock.Anything, feeTransactionID).Times(1).Return(nil, nil)
				accountTransactionRepo.On("CommitTransaction", mock.Anything).Once().Return(nil)
			},
		},
		{
			name:                 "SUCCESS: Settlement on hold but being bypassed (Refund)",
			byPassSettlementHold: true,
			wantErr:              false,
			expectedError:        pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrPaymentChargeIsOnHold).Error(),
			setupMock: func(paymentSvc *serviceMock.IPaymentService, merchantSvc *serviceMock.IMerchantService, accountTransactionRepo *repositoryMock.IAccountTransactionRepository, cardFundedPayoutSvc *serviceMock.ICardFundedPayoutService) {
				accountTransactionRepo.On("FindByID", mock.Anything, transactionID).Times(1).Return(&orchestraModel.AccountTransactionWithUseCase{
					UUID:           util.ParseUUID(transactionID),
					MerchantID:     util.ParseUUID(merchantID),
					AdditionalInfo: types.NullJSONText{Valid: true},
					AdditionalInfoObj: orchestraModel.FeeTransactionMetadataObject{
						AccountTransactionMetadataObject: &settlementModel.AccountTransactionMetadataObject{
							SettlementDetail: merchant.SettlementConfig{
								IsOnHold: true,
							},
						},
					},
					Type:             constant.TypeMerchantPayment,
					SettlementStatus: sql.NullString{String: constant.SettlementStatusPending, Valid: true},
				}, nil)
				accountTransactionRepo.On("BeginTransaction", mock.Anything).Once().Return(context.Background(), nil)
				accountTransactionRepo.On("UpdatePaymentTransactionStatusAndMetadataByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Times(1).Return(&merchant.Merchant{ParentID: sql.NullString{Valid: true, String: uuid.NewString()}}, nil)
				accountTransactionRepo.On("FindByID", mock.Anything, feeTransactionID).Times(1).Return(nil, nil)
				accountTransactionRepo.On("CommitTransaction", mock.Anything).Once().Return(nil)

			},
		},
		{
			name:          "SUCCESS: Settlement on hold",
			wantErr:       true,
			expectedError: pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrPaymentChargeIsOnHold).Error(),
			setupMock: func(paymentSvc *serviceMock.IPaymentService, merchantSvc *serviceMock.IMerchantService, accountTransactionRepo *repositoryMock.IAccountTransactionRepository, cardFundedPayoutSvc *serviceMock.ICardFundedPayoutService) {
				accountTransactionRepo.On("FindByID", mock.Anything, transactionID).Times(1).Return(&orchestraModel.AccountTransactionWithUseCase{
					UUID:           util.ParseUUID(transactionID),
					MerchantID:     util.ParseUUID(merchantID),
					AdditionalInfo: types.NullJSONText{Valid: true},
					AdditionalInfoObj: orchestraModel.FeeTransactionMetadataObject{
						AccountTransactionMetadataObject: &settlementModel.AccountTransactionMetadataObject{
							SettlementDetail: merchant.SettlementConfig{
								IsOnHold: true,
							},
						},
					},
					Type:             constant.TypeMerchantPayment,
					SettlementStatus: sql.NullString{String: constant.SettlementStatusPending, Valid: true},
				}, nil)
			},
		},
		{
			name: "SUCCESS: Settlement cancelled",
			setupMock: func(paymentSvc *serviceMock.IPaymentService, merchantSvc *serviceMock.IMerchantService, accountTransactionRepo *repositoryMock.IAccountTransactionRepository, cardFundedPayoutSvc *serviceMock.ICardFundedPayoutService) {
				accountTransactionRepo.On("FindByID", mock.Anything, transactionID).Times(1).Return(&orchestraModel.AccountTransactionWithUseCase{
					UUID:             util.ParseUUID(transactionID),
					MerchantID:       util.ParseUUID(merchantID),
					AdditionalInfo:   types.NullJSONText{Valid: true},
					Type:             constant.TypeMerchantPayment,
					SettlementStatus: sql.NullString{String: constant.SettlementStatusCancelled, Valid: true},
				}, nil)
			},
		},
		{
			name: "SUCCESS: Settlement success",
			setupMock: func(paymentSvc *serviceMock.IPaymentService, merchantSvc *serviceMock.IMerchantService, accountTransactionRepo *repositoryMock.IAccountTransactionRepository, cardFundedPayoutSvc *serviceMock.ICardFundedPayoutService) {
				accountTransactionRepo.On("FindByID", mock.Anything, transactionID).Times(1).Return(&orchestraModel.AccountTransactionWithUseCase{
					UUID:             util.ParseUUID(transactionID),
					MerchantID:       util.ParseUUID(merchantID),
					AdditionalInfo:   types.NullJSONText{Valid: true},
					Type:             constant.TypePayment,
					SettlementStatus: sql.NullString{String: constant.StatusSuccess, Valid: true},
				}, nil)
			},
		},
		{
			name: "SUCCESS: Settlement failed",
			setupMock: func(paymentSvc *serviceMock.IPaymentService, merchantSvc *serviceMock.IMerchantService, accountTransactionRepo *repositoryMock.IAccountTransactionRepository, cardFundedPayoutSvc *serviceMock.ICardFundedPayoutService) {
				accountTransactionRepo.On("FindByID", mock.Anything, transactionID).Times(1).Return(&orchestraModel.AccountTransactionWithUseCase{
					UUID:             util.ParseUUID(transactionID),
					MerchantID:       util.ParseUUID(merchantID),
					AdditionalInfo:   types.NullJSONText{Valid: true},
					Type:             constant.TypePayment,
					SettlementStatus: sql.NullString{String: constant.StatusFailed, Valid: true},
				}, nil)
			},
		},
		{
			name: "SUCCESS: Card funded payout settlement",
			setupMock: func(paymentSvc *serviceMock.IPaymentService, merchantSvc *serviceMock.IMerchantService, accountTransactionRepo *repositoryMock.IAccountTransactionRepository, cardFundedPayoutSvc *serviceMock.ICardFundedPayoutService) {
				accountTransactionRepo.On("FindByID", mock.Anything, transactionID).Times(1).Return(validCardFundedPayoutTransaction, nil)
				accountTransactionRepo.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				accountTransactionRepo.On("UpdatePaymentTransactionStatusAndMetadataByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Times(1).Return(&merchant.Merchant{ParentID: sql.NullString{Valid: false}}, nil)
				accountTransactionRepo.On("FindByID", mock.Anything, feeTransactionID).Times(1).Return(nil, nil)
				paymentSvc.On("ProcessSplitRoute", mock.Anything, mock.Anything).Once().Return(nil)
				accountTransactionRepo.On("CommitTransaction", mock.Anything).Return(nil)
				cardFundedPayoutSvc.On("ProcessFinishCardFundedPayoutSettlement", mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			name:    "SUCCESS: Card funded payout settlement - ProcessFinishCardFundedPayoutSettlement error (non-blocking)",
			wantErr: false, // Error from ProcessFinishCardFundedPayoutSettlement is logged as warning, not returned
			setupMock: func(paymentSvc *serviceMock.IPaymentService, merchantSvc *serviceMock.IMerchantService, accountTransactionRepo *repositoryMock.IAccountTransactionRepository, cardFundedPayoutSvc *serviceMock.ICardFundedPayoutService) {
				accountTransactionRepo.On("FindByID", mock.Anything, transactionID).Times(1).Return(validCardFundedPayoutTransaction, nil)
				accountTransactionRepo.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				accountTransactionRepo.On("UpdatePaymentTransactionStatusAndMetadataByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Times(1).Return(&merchant.Merchant{ParentID: sql.NullString{Valid: false}}, nil)
				accountTransactionRepo.On("FindByID", mock.Anything, feeTransactionID).Times(1).Return(nil, nil)
				paymentSvc.On("ProcessSplitRoute", mock.Anything, mock.Anything).Once().Return(nil)
				accountTransactionRepo.On("CommitTransaction", mock.Anything).Return(nil)
				cardFundedPayoutSvc.On("ProcessFinishCardFundedPayoutSettlement", mock.Anything, mock.Anything).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf.Reset()

			// Create fresh mocks for each test case
			paymentSvc := serviceMock.NewIPaymentService(t)
			merchantSvc := serviceMock.NewIMerchantService(t)
			cardFundedPayoutSvc := serviceMock.NewICardFundedPayoutService(t)
			accountTransactionRepo := repositoryMock.NewIAccountTransactionRepository(t)
			svc := New(log, accountTransactionRepo, WithPaymentSvc(paymentSvc), WithMerchantSvc(merchantSvc))
			WithCardFundedPayoutSvc(svc, cardFundedPayoutSvc)

			if tc.setupMock != nil {
				tc.setupMock(paymentSvc, merchantSvc, accountTransactionRepo, cardFundedPayoutSvc)
			}

			err := svc.ProcessSettlement(context.Background(), &settlementModel.ProcessSettlementRequest{
				TransactionID: transactionID, FeeTransactionID: feeTransactionID, MerchantID: merchantID,
				ByPassSettlementHold: tc.byPassSettlementHold,
			})
			if tc.wantErr {
				assert.Error(t, err)
				assert.Equal(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProcessSettlementTransactionFee(t *testing.T) {
	log := loggerMock.NewILogger(t)
	accountTrxRepo := repositoryMock.NewIAccountTransactionRepository(t)
	merchantSvc := serviceMock.NewIMerchantService(t)

	service := New(log, accountTrxRepo, WithMerchantSvc(merchantSvc))

	merchantId := "add571b3-282b-458f-8f3d-ec03bca8810b"

	tests := []struct {
		name      string
		setupMock func()
		wantErr   error
	}{
		{
			name: "ERROR:Find transaction by id",
			setupMock: func() {
				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchant.Merchant{ParentID: sql.NullString{Valid: true, String: merchantId}}, nil)
				accountTrxRepo.On("FindByID", mock.Anything, mock.Anything).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
				log.On(
					"Error", mock.Anything, "Failed while find transaction fee by id", mock.Anything,
				).Once().Return()
			},
			wantErr: constant.ErrSomeErrorForUnitTest,
		},
		{
			name: "Data not found",
			setupMock: func() {
				accountTrxRepo.On("FindByID", mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
		},
		{
			name: "ERROR:Merchant forbidden",
			setupMock: func() {
				accountTrxRepo.On("FindByID", mock.Anything, mock.Anything).Once().Return(&orchestraModel.AccountTransactionWithUseCase{Status: constant.StatusSuccess}, nil)
			},
			wantErr: pkgErr.New(response.HttpErrRequest, constant.ErrMerchantIsNotMatch),
		},
		{
			name: "SUCCESS:Skip settlement process for fee with deduction type manual",
			setupMock: func() {
				accountTrxRepo.On(
					"FindByID", mock.Anything, mock.Anything,
				).Once().Return(&orchestraModel.AccountTransactionWithUseCase{
					MerchantID:       util.ParseUUID(merchantId),
					Status:           constant.StatusSuccess,
					SettlementStatus: sql.NullString{Valid: true, String: constant.StatusPending},
					AdditionalInfo: types.NullJSONText{
						Valid:    true,
						JSONText: []byte(`{"deductionType": "MANUAL"}`),
					},
				}, nil)
			},
		},
		{
			name: "ERROR:Update additional info",
			setupMock: func() {
				accountTrxRepo.On(
					"FindByID", mock.Anything, mock.Anything,
				).Return(&orchestraModel.AccountTransactionWithUseCase{
					MerchantID: util.ParseUUID(merchantId),
					Status:     constant.StatusSuccess,
					AdditionalInfo: types.NullJSONText{
						Valid:    true,
						JSONText: []byte(`{"settlementDetail": {}, "settlementStatus": "PENDING"}`),
					},
				}, nil)

				accountTrxRepo.On("UpdateSettlementStatusAndSettlementAtByID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(constant.ErrSomeErrorForUnitTest)
				log.On(
					"Error", mock.Anything, "Failed while update settlement status and settlement at by id (fee)", mock.Anything,
				).Once().Return()
			},
			wantErr: pkgErr.New(response.HttpErrDatabase, constant.ErrSomeErrorForUnitTest),
		},
		{
			name: "ERROR:Update status account transaction",
			setupMock: func() {
				accountTrxRepo.On(
					"FindByID", mock.Anything, mock.Anything,
				).Return(&orchestraModel.AccountTransactionWithUseCase{
					MerchantID: util.ParseUUID(merchantId),
					Status:     constant.StatusSuccess,
					AdditionalInfo: types.NullJSONText{
						Valid:    true,
						JSONText: []byte(`{"settlementDetail": {}, "settlementStatus": "PENDING"}`),
					},
				}, nil)

				accountTrxRepo.On("UpdateSettlementStatusAndSettlementAtByID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

				accountTrxRepo.On(
					"UpdateStatusAccountTransaction", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(constant.ErrSomeErrorForUnitTest)
				log.On(
					"Error", mock.Anything, "Failed while update status account transaction (fee)", mock.Anything,
				).Once().Return()
			},
			wantErr: pkgErr.New(response.HttpErrDatabase, constant.ErrSomeErrorForUnitTest),
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				accountTrxRepo.On(
					"FindByID", mock.Anything, mock.Anything,
				).Return(&orchestraModel.AccountTransactionWithUseCase{
					MerchantID: util.ParseUUID(merchantId),
					Status:     constant.StatusSuccess,
					AdditionalInfo: types.NullJSONText{
						Valid:    true,
						JSONText: []byte(`{"settlementDetail": {}, "settlementStatus": "PENDING"}`),
					},
				}, nil)

				accountTrxRepo.On("UpdateSettlementStatusAndSettlementAtByID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

				accountTrxRepo.On(
					"UpdateStatusAccountTransaction", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			assert.Equal(t, test.wantErr, service.ProcessSettlementTransactionFee(context.Background(), merchantId, ""))
		})
	}
}
