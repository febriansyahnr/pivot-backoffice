package disbursementRepository

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	pdkLoggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFindByID(t *testing.T) {
	metadata := disbursementModel.Metadata{
		FeeDetail: feeModel.FeeMetadataObject{
			Type:                constant.TypeDisbursement,
			DeductionType:       constant.MerchantFeeDeductionTypeAutomated,
			AmountType:          constant.MerchantFeeAmountPercentageType,
			Amount:              1_000,
			Percentage:          2.5,
			LinkedTransactionId: uuid.NewString(),
		},
	}
	rawMetadata, _ := json.Marshal(metadata)

	testCase := []struct {
		name       string
		mockSetup  func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr    bool
		wantResult *disbursementModel.DisbursementWithTransaction
	}{
		{
			name: "SUCCESS",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeDisbursementWithTransactionReference), constant.StringMockType(), constant.StringMockType(),
				).Run(func(args mock.Arguments) {
					args.Get(1).(*disbursementModel.DisbursementWithTransaction).Metadata.Valid = true
					args.Get(1).(*disbursementModel.DisbursementWithTransaction).Metadata.JSONText = rawMetadata
				}).Return(nil)
			},
			wantResult: &disbursementModel.DisbursementWithTransaction{
				Disbursement: disbursementModel.Disbursement{
					Metadata: types.NullJSONText{
						Valid: true, JSONText: rawMetadata,
					},
					MetadataObj: metadata,
				},
			},
		},
		{
			name: "ERROR: Mysql error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeDisbursementWithTransactionReference),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(constant.ErrSomeErrorForUnitTest)

			},
			wantErr: true,
		},
		{
			name: "ERROR: Data not found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeDisbursementWithTransactionReference),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(sql.ErrNoRows)

			},
			wantErr: false,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			ctx := context.Background()

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)

			if result, err := repo.FindByID(ctx, uuid.NewString()); tc.wantErr {
				assert.Error(t, err)

			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantResult, result)
			}

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestFindForReversalDisbursementById(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)

	fee := disbursementModel.TransactionMetadataForReversal{
		Id:     "9876543",
		Amount: 2_000,
		Status: constant.StatusPending,
	}
	transaction := disbursementModel.TransactionMetadataForReversal{
		Id:     "123456",
		Amount: 100_000,
		Status: constant.StatusSuccess,
	}
	disbursement := &disbursementModel.DisbursementForReversal{
		Status:      constant.StatusApproved,
		Fee:         fee,
		Transaction: transaction,
	}
	disbursement.RawFee, _ = json.Marshal(fee)
	disbursement.RawTransaction, _ = json.Marshal(transaction)

	disbursementReversalMockType := mock.AnythingOfType("*disbursementModel.DisbursementForReversal")

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult *disbursementModel.DisbursementForReversal
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", constant.ValueCtxMockType(), disbursementReversalMockType, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest,
		},
		{
			name: "ERROR:Data not found", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", constant.ValueCtxMockType(), disbursementReversalMockType, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"GetContext", constant.ValueCtxMockType(), disbursementReversalMockType, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Run(func(args mock.Arguments) {
					*args.Get(1).(*disbursementModel.DisbursementForReversal) = disbursementModel.DisbursementForReversal{
						Status:         constant.StatusApproved,
						RawFee:         disbursement.RawFee,
						RawTransaction: disbursement.RawTransaction,
					}
				}).Return(nil)
			},
			wantResult: disbursement,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.FindForReversalDisbursementById(context.Background(), uuid.NewString(), uuid.NewString())
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestGetMerchantIDsForPayoutCallback(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)
	log := pdkLoggerMock.NewILogger(t)

	repo := New(db, log)

	merchanIds := []string{
		"1c3f0bda-0c10-4a5d-ac0a-d3083e1e0a4e", "3a6ebcc9-09ba-436a-b130-202fcd47372d",
	}

	tests := []struct {
		name       string
		setupMock  func()
		wantError  error
		wantResult []string
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(assert.AnError)
				log.On("Error", mock.Anything, "Failed while getting merchant ids for payout callback", mock.Anything).Once().Return()
			},
			wantError: assert.AnError,
		},
		{
			name: "SUCCESS:Data not found", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(sql.ErrNoRows)
			},
			wantError: nil, wantResult: nil,
		},
		{
			name: "SUCCESS:Data not found", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Run(func(args mock.Arguments) {
					*args.Get(1).(*disbursementModel.MerchantIDForPayoutCallback) = disbursementModel.MerchantIDForPayoutCallback{
						MerchantId: merchanIds[0], ParentMerchantId: merchanIds[1],
					}
				}).Return(nil)
			},
			wantError: nil, wantResult: merchanIds,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.GetMerchantIDsForPayoutCallback(t.Context(), "f3c5294c-2095-4d37-b4ca-8530979104f1")
			assert.Equal(t, test.wantError, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
