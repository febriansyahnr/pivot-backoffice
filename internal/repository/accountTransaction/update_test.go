package accounttransaction_repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	fdscommon "github.com/paper-indonesia/pivot-backoffice/internal/model/fdsProcessor/fdsCommon"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateStatusAccountTransaction(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Update status account transaction",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.PtrStringMockType(),
					constant.PtrStringMockType(),
					constant.TimeMockType(),
					constant.StringMockType(),
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Failure Update to Database",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.PtrStringMockType(),
					constant.PtrStringMockType(),
					constant.TimeMockType(),
					constant.StringMockType(),
				).Return(false, constant.ErrSomeErrorForUnitTest)

			},
			wantErr: true,
		},
		{
			name: "ERROR: No Rows Affected",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.PtrStringMockType(),
					constant.PtrStringMockType(),
					constant.TimeMockType(),
					constant.StringMockType(),
				).Return(false, constant.ErrNoRowsAffected)
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
			err := repo.UpdateStatusAccountTransaction(ctx, uuid.NewString(), constant.StatusSuccess, nil, nil)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestCancelIndirectTransactionFee(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)

	tests := []struct {
		name      string
		setupMock func()
		wantErr   error
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				db.On(
					"ExecContext", constant.ValueCtxMockType(), constant.StringMockType(), constant.TimeMockType(), constant.StringMockType(),
				).Once().Return(false, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"ExecContext", constant.ValueCtxMockType(), constant.StringMockType(), constant.TimeMockType(), constant.StringMockType(),
				).Return(true, nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			assert.Equal(t, test.wantErr, repo.CancelIndirectTransactionFee(context.Background(), uuid.NewString(), time.Now()))
		})
	}
}

func TestDeductBalanceForIndirectFeeType(t *testing.T) {

	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)

	rawQuery := "SELECT ???"
	trxId1 := "123"
	trxId2 := "124"
	trxIds := []string{trxId1, trxId2}
	merchantId := uuid.NewString()

	tests := []struct {
		name      string
		setupMock func()
		wantErr   error
	}{
		{
			name: "ERROR:Build IN query statement",
			setupMock: func() {
				db.On(
					"In", constant.StringMockType(), merchantId, trxIds,
				).Once().Return("", nil, fmt.Errorf("query builder: %v", constant.ErrSomeErrorForUnitTest)) // NOSONAR
			},
			wantErr: fmt.Errorf("query builder: %v", constant.ErrSomeErrorForUnitTest), // NOSONAR
		},
		{
			name: "ERROR:Exec context",
			setupMock: func() {
				db.On(
					"In", constant.StringMockType(), merchantId, trxIds,
				).Return(rawQuery, []interface{}{merchantId, trxId1, trxId2}, nil)
				db.On("Rebind", rawQuery).Return(rawQuery)

				db.On(
					"ExecContext", constant.ValueCtxMockType(), rawQuery, merchantId, trxId1, trxId2,
				).Once().Return(false, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"ExecContext", constant.ValueCtxMockType(), rawQuery, merchantId, trxId1, trxId2,
				).Return(true, nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			test.setupMock()
			assert.Equal(t, test.wantErr, repo.DeductBalanceForIndirectFeeType(context.Background(), merchantId, trxIds))
		})
	}
}

func TestUpdateTransactionsStatusAndAdditionalInfoByID(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Update status and additional info account transaction by id",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType("types.NullJSONText"),
					constant.TimeMockType(),
					constant.StringMockType(),
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Failure Update to Database",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType("types.NullJSONText"),
					constant.TimeMockType(),
					constant.StringMockType(),
				).Return(false, constant.ErrSomeErrorForUnitTest)

			},
			wantErr: true,
		},
		{
			name: "ERROR: No Rows Affected",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType("types.NullJSONText"),
					constant.TimeMockType(),
					constant.StringMockType(),
				).Return(false, constant.ErrNoRowsAffected)
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
			err := repo.UpdateTransactionsStatusAndAdditionalInfoByID(ctx, uuid.NewString(), constant.StatusSuccess, constant.ReasonTypeOtherReason, constant.ReasonDescTransactionVoidByProcessor, types.NullJSONText{})
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestVoidTransaction(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Void transaction",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.TimeMockType(),
					constant.StringMockType(),
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Failure Update to Database",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.TimeMockType(),
					constant.StringMockType(),
				).Return(false, constant.ErrSomeErrorForUnitTest)

			},
			wantErr: true,
		},
		{
			name: "ERROR: No Rows Affected",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.TimeMockType(),
					constant.StringMockType(),
				).Return(false, constant.ErrNoRowsAffected)
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
			err := repo.VoidTransaction(ctx, &orchestratorModel.VoidTransactionRequest{
				Status:            constant.StatusSuccess,
				ReasonType:        constant.ReasonTypeOtherReason,
				ReasonDescription: constant.ReasonDescTransactionVoidByProcessor,
				SettlementStatus:  constant.SettlementStatusCancelled,
				TrxID:             uuid.NewString(),
			})
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestUpdateProcessorReferenceName(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Update processor refere",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.TimeMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Failure Update to Database",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.TimeMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(false, constant.ErrSomeErrorForUnitTest)

			},
			wantErr: true,
		},
		{
			name: "ERROR: No Rows Affected",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.TimeMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(false, constant.ErrNoRowsAffected)
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
			err := repo.UpdateProcessorAndReconReference(ctx, uuid.NewString(), constant.SnapCoreProcessor, uuid.NewString(), uuid.NewString())
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestUpdateTransactionTimestamp(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Update transaction timestamp",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.TimeMockType(),
					constant.TimeMockType(),
					constant.StringMockType(),
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Failure Update to Database",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.TimeMockType(),
					constant.TimeMockType(),
					constant.StringMockType(),
				).Return(false, constant.ErrSomeErrorForUnitTest)

			},
			wantErr: true,
		},
		{
			name: "ERROR: No Rows Affected",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.TimeMockType(),
					constant.TimeMockType(),
					constant.StringMockType(),
				).Return(false, constant.ErrNoRowsAffected)
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
			err := repo.UpdateTransactionTimestamp(ctx, uuid.NewString(), time.Now())
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestRearrangeUpdatedAtForTransactionWithPendingStatus(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)

	referenceIds := []string{"fd28e99e-685c-4323-92b1-70c6f30cddb2"}
	tests := []struct {
		name         string
		referenceIds []string
		setupMock    func()
		wantErr      error
	}{
		{
			name:      "SUCCESS:Empty reference ids",
			setupMock: func() { /* empty function */ },
		},
		{
			name:         "ERROR:Query builder",
			referenceIds: referenceIds,
			setupMock: func() {
				db.On(
					"In", constant.StringMockType(), constant.TimeMockType(), mock.Anything,
				).Once().Return("", nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest,
		},
		{
			name:         "ERROR:Execute query",
			referenceIds: referenceIds,
			setupMock: func() {
				db.On(
					"In", constant.StringMockType(), constant.TimeMockType(), mock.Anything,
				).Return("SELECT ???", []interface{}{time.Now().UTC(), "fd28e99e-685c-4323-92b1-70c6f30cddb2"}, nil)
				db.On("Rebind", constant.StringMockType()).Return("SELECT ???")
				db.On(
					"ExecContext", constant.ValueCtxMockType(), constant.StringMockType(), constant.TimeMockType(), constant.StringMockType(),
				).Once().Return(false, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest,
		},
		{
			name:         "SUCCESS",
			referenceIds: referenceIds,
			setupMock: func() {
				db.On(
					"ExecContext", constant.ValueCtxMockType(), constant.StringMockType(), constant.TimeMockType(), constant.StringMockType(),
				).Return(true, nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			assert.Equal(t, test.wantErr, repo.RearrangeUpdatedAtForTransactionWithPendingStatus(
				context.Background(), test.referenceIds, time.Now().UTC(),
			))
		})
	}
}

func TestUpdatePaymentTransactionStatusAndMetadataByID(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)

	ledgerId := uuid.NewString()

	tests := []struct {
		name       string
		request    orchestratorModel.UpdatePaymentTransactionRequest
		metadata   orchestratorModel.MetadataPayment[any]
		setupMocks func()
		wantErr    error
	}{
		{
			name: "ERROR:Some error",
			request: orchestratorModel.UpdatePaymentTransactionRequest{
				Status: constant.StatusPending,
			},
			setupMocks: func() {
				db.On(
					"ExecContext", constant.ValueCtxMockType(), constant.StringMockType(), constant.StatusPending, constant.TimeMockType(), ledgerId,
				).Once().Return(false, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest,
		},
		{
			name: "ERROR:Data not found",
			request: orchestratorModel.UpdatePaymentTransactionRequest{
				Status: constant.StatusPending,
			},
			setupMocks: func() {
				db.On(
					"ExecContext", constant.ValueCtxMockType(), constant.StringMockType(), constant.StatusPending, constant.TimeMockType(), ledgerId,
				).Once().Return(false, nil)
			},
			wantErr: constant.ErrDataNotFound,
		},
		{
			name: "SUCCESS:Update account transaction",
			request: orchestratorModel.UpdatePaymentTransactionRequest{
				UpdatedAt:              time.Now().UTC(),
				TransactionTimestamp:   time.Now().UTC(),
				ProcessorTransactionId: "1234",
				MethodDetail:           map[string]string{},
				SettlementStatus:       util.ValueToPtr(constant.StatusSuccess),
				SettlementAt:           util.ValueToPtr(time.Now().UTC()),
			},
			metadata: orchestratorModel.MetadataPayment[any]{
				SettlementDetail: &orchestratorModel.MetadataPaymentSettlementDetail{},
				FeeDetail:        &feeModel.FeeMetadataObject{},
				FeeOnBehalf:      &feeModel.TrxFeeOnBehalfMetadata{},
			},
			setupMocks: func() {
				db.On(
					"ExecContext", constant.ValueCtxMockType(), constant.StringMockType(), constant.TimeMockType(), constant.TimeMockType(), "1234", constant.StatusSuccess, mock.Anything, mock.AnythingOfType("[]uint8"), mock.AnythingOfType("[]uint8"), mock.AnythingOfType("[]uint8"), ledgerId,
				).Once().Return(true, nil)
			},
		},
		{
			name: "SUCCESS:Update with ReconReferenceNo",
			request: orchestratorModel.UpdatePaymentTransactionRequest{
				Status:   constant.StatusSuccess,
				LedgerId: ledgerId,
			},
			metadata: orchestratorModel.MetadataPayment[any]{
				ReconReferenceNo: "recon-ref-123",
			},
			setupMocks: func() {
				db.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					mock.MatchedBy(func(s string) bool {
						return strings.Contains(s, "JSON_SET(additional_info, '$.reconReferenceNo', ?)")
					}),
					constant.StatusSuccess,
					constant.TimeMockType(),
					"recon-ref-123",
					ledgerId,
				).Once().Return(true, nil)
			},
		},
		{
			name: "SUCCESS:Update with MethodDetail and ChargeStatus",
			request: orchestratorModel.UpdatePaymentTransactionRequest{
				Status:   constant.StatusSuccess,
				LedgerId: ledgerId,
			},
			metadata: orchestratorModel.MetadataPayment[any]{
				MethodDetail: map[string]interface{}{
					"payment_method": "credit_card",
					"card_type":      "visa",
				},
				ChargeStatus: "charged",
			},
			setupMocks: func() {
				db.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					mock.MatchedBy(func(s string) bool {
						return strings.Contains(s, "JSON_SET(additional_info,")
					}),
					constant.StatusSuccess,
					constant.TimeMockType(),
					mock.AnythingOfType("[]uint8"),
					mock.AnythingOfType("[]uint8"),
					ledgerId,
				).Once().Return(true, nil)
			},
		},
		{
			name: "SUCCESS:Update with ReconDetail",
			request: orchestratorModel.UpdatePaymentTransactionRequest{
				Status:   constant.StatusSuccess,
				LedgerId: ledgerId,
			},
			metadata: orchestratorModel.MetadataPayment[any]{
				ReconDetail: &orchestratorModel.MetadataReconDetail{
					Status:   "matched",
					DateTime: time.Now().Format(time.RFC3339),
				},
			},
			setupMocks: func() {
				db.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					mock.MatchedBy(func(s string) bool {
						return strings.Contains(s, "JSON_SET(additional_info,")
					}),
					constant.StatusSuccess,
					constant.TimeMockType(),
					mock.AnythingOfType("[]uint8"),
					ledgerId,
				).Once().Return(true, nil)
			},
		},
		{
			name: "SUCCESS:Update with SettlementStatus and SettlementAt",
			request: orchestratorModel.UpdatePaymentTransactionRequest{
				Status:           constant.StatusSuccess,
				LedgerId:         ledgerId,
				SettlementStatus: util.ValueToPtr(constant.StatusSuccess),
				SettlementAt:     util.ValueToPtr(time.Now().UTC()),
			},
			setupMocks: func() {
				db.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					mock.MatchedBy(func(s string) bool {
						return strings.Contains(s, "settlement_status = ?, settlement_at = ?")
					}),
					constant.StatusSuccess,
					constant.TimeMockType(),
					constant.StatusSuccess,
					mock.AnythingOfType("time.Time"),
					ledgerId,
				).Once().Return(true, nil)
			},
		},
		{
			name: "SUCCESS:Update with ProcessorReferenceId",
			request: orchestratorModel.UpdatePaymentTransactionRequest{
				Status:                 constant.StatusSuccess,
				LedgerId:               ledgerId,
				ProcessorReferenceName: constant.SnapCoreProcessor,
				ProcessorReferenceId:   "processor-ref-12345",
			},
			setupMocks: func() {
				db.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					mock.MatchedBy(func(s string) bool {
						return strings.Contains(s, "processor_reference_id = ?")
					}),
					constant.StatusSuccess,
					constant.TimeMockType(),
					constant.SnapCoreProcessor,
					"processor-ref-12345",
					ledgerId,
				).Once().Return(true, nil)
			},
		},
		{
			name: "SUCCESS:Update with StatementDescriptor and expiredAt",
			request: orchestratorModel.UpdatePaymentTransactionRequest{
				Status:   constant.StatusSuccess,
				LedgerId: ledgerId,
			},
			metadata: orchestratorModel.MetadataPayment[any]{
				StatementDescriptor: "Statement desc",
				ExpiredAt:           time.Now().Add(15 * time.Minute).UTC(),
			},
			setupMocks: func() {
				db.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.StatusSuccess,
					constant.TimeMockType(),
					mock.AnythingOfType("[]uint8"),
					mock.AnythingOfType("[]uint8"),
					ledgerId,
				).Once().Return(true, nil)
			},
		},
		{
			name: "SUCCESS:Update with SettlementModel",
			request: orchestratorModel.UpdatePaymentTransactionRequest{
				Status:          constant.StatusSuccess,
				LedgerId:        ledgerId,
				SettlementModel: util.ValueToPtr(constant.PaymentMethodChannelTypeFacilitator),
			},
			setupMocks: func() {
				db.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					mock.MatchedBy(func(s string) bool {
						return strings.Contains(s, "settlement_model = ?")
					}),
					constant.StatusSuccess,
					constant.TimeMockType(),
					constant.PaymentMethodChannelTypeFacilitator,
					ledgerId,
				).Once().Return(true, nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMocks()

			test.request.LedgerId = ledgerId
			assert.Equal(
				t, test.wantErr, repo.UpdatePaymentTransactionStatusAndMetadataByID(
					context.Background(), test.request, test.metadata,
				),
			)

			db.AssertExpectations(t)
		})
	}
}

func TestUpdateTransactionWithPendingStatusByReferenceIdTypeAndChannel(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)

	tests := []struct {
		name       string
		setupMocks func()
		wantErr    error
	}{
		{
			name: "ERROR:Some error",
			setupMocks: func() {
				db.On(
					"ExecContext", constant.ValueCtxMockType(), constant.StringMockType(),
					constant.JSONTextMockType(), constant.StringMockType(), constant.StringMockType(), constant.TimeMockType(), constant.StringMockType(), constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Return(false, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest,
		},
		{
			name: "ERROR:Data not found",
			setupMocks: func() {
				db.On(
					"ExecContext", constant.ValueCtxMockType(), constant.StringMockType(),
					constant.JSONTextMockType(), constant.StringMockType(), constant.StringMockType(), constant.TimeMockType(), constant.StringMockType(), constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Return(false, nil)
			},
			wantErr: fmt.Errorf("update metadata for transaction with pending status: %w", constant.ErrDataNotFound),
		},
		{
			name: "SUCCESS",
			setupMocks: func() {
				db.On(
					"ExecContext", constant.ValueCtxMockType(), constant.StringMockType(),
					constant.JSONTextMockType(), constant.StringMockType(), constant.StringMockType(), constant.TimeMockType(), constant.StringMockType(), constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Return(true, nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMocks()

			assert.Equal(
				t, test.wantErr, repo.UpdateTransactionWithPendingStatusByReferenceIdTypeAndChannel(context.Background(), "1", "2", "3", orchestratorModel.UpdateTransactionWithPendingStatus{}),
			)
		})
	}
}

func TestBulkUpdateTransactions(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	repo := New(db, logger)

	tests := []struct {
		name       string
		setupMocks func()
		wantErr    error
	}{
		{
			name: "SUCCESS",
			setupMocks: func() {
				db.On("BeginTxx", mock.Anything).Once().Return(nil, nil)

				db.On(
					"ExecContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(true, nil)

				db.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "ERROR: Begin Trx",
			setupMocks: func() {
				tx := &sqlx.Tx{}
				ctx := context.WithValue(context.Background(), mySqlExt.CtxSqlTx, tx)

				db.On("BeginTxx", mock.Anything).Once().Return(ctx, errors.New("errors"))
			},
			wantErr: errors.New("errors"),
		},
		{
			name: "ERROR: Exec context",
			setupMocks: func() {
				tx := &sqlx.Tx{}
				ctx := context.WithValue(context.Background(), mySqlExt.CtxSqlTx, tx)

				db.On("BeginTxx", mock.Anything).Once().Return(ctx, nil)

				db.On(
					"ExecContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(false, errors.New("errors"))

				db.On("Rollback", mock.Anything).Once().Return(nil)
			},
			wantErr: errors.New("errors"),
		},
		{
			name: "ERROR: Rollback",
			setupMocks: func() {
				tx := &sqlx.Tx{}
				ctx := context.WithValue(context.Background(), mySqlExt.CtxSqlTx, tx)

				db.On("BeginTxx", mock.Anything).Once().Return(ctx, nil)

				db.On(
					"ExecContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(false, errors.New("errors"))

				db.On("Rollback", ctx).Once().Return(errors.New("errors"))
			},
			wantErr: errors.New("errors"),
		},
		{
			name: "ERROR: Commit error",
			setupMocks: func() {
				tx := &sqlx.Tx{}
				ctx := context.WithValue(context.Background(), mySqlExt.CtxSqlTx, tx)

				db.On("BeginTxx", mock.Anything).Once().Return(ctx, nil)

				db.On(
					"ExecContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(true, nil)

				db.On("Commit", ctx).Once().Return(errors.New("commit error"))
			},
			wantErr: errors.New("commit error"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMocks()
			assert.Equal(
				t, test.wantErr, repo.BulkUpdateTransactions(context.Background(), []*orchestratorModel.AccountTransaction{
					{
						UUID:                   uuid.New(),
						Status:                 constant.StatusSuccess,
						ReasonType:             sql.NullString{String: "reason", Valid: true},
						ReasonDescription:      sql.NullString{String: "description", Valid: true},
						SettlementStatus:       sql.NullString{String: "settlement", Valid: true},
						UpdatedAt:              time.Now().UTC(),
						Processor:              "1",
						ProcessorID:            "1",
						ProcessorTransactionID: "1",
						Reference:              "1",
					},
				}),
			)
		})
	}
}

func TestUpdateAdditionalInfoByID(t *testing.T) {
	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Update additional info",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					mock.AnythingOfType("types.NullJSONText"),
					constant.StringMockType(),
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Database error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					mock.AnythingOfType("types.NullJSONText"),
					constant.StringMockType(),
				).Return(false, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name: "ERROR: No rows affected",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					mock.AnythingOfType("types.NullJSONText"),
					constant.StringMockType(),
				).Return(false, constant.ErrNoRowsAffected)
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			ctx := context.Background()

			additionalInfo := types.NullJSONText{}
			additionalInfo.Scan(`{"test": "value"}`)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			err := repo.UpdateAdditionalInfoByID(ctx, uuid.NewString(), additionalInfo)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)
		})
	}
}

func TestUpdateSettlementStatusAndSettlementAtByID(t *testing.T) {
	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Update settlement status and date",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.TimeMockType(),
					constant.TimeMockType(),
					constant.StringMockType(),
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Database error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.TimeMockType(),
					constant.TimeMockType(),
					constant.StringMockType(),
				).Return(false, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name: "ERROR: No rows affected",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.TimeMockType(),
					constant.TimeMockType(),
					constant.StringMockType(),
				).Return(false, constant.ErrNoRowsAffected)
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			ctx := context.Background()
			settlementTime := time.Now().UTC()
			settlementStatus := "SETTLED"

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			err := repo.UpdateSettlementStatusAndSettlementAtByID(ctx, uuid.NewString(), settlementStatus, settlementTime)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)
		})
	}
}

func TestUpdateReasonOnly(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Update reason only without changing updated_at",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.PtrStringMockType(),
					constant.PtrStringMockType(),
					constant.StringMockType(),
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Failure Update to Database",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.PtrStringMockType(),
					constant.PtrStringMockType(),
					constant.StringMockType(),
				).Return(false, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name: "ERROR: No Rows Affected",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.PtrStringMockType(),
					constant.PtrStringMockType(),
					constant.StringMockType(),
				).Return(false, constant.ErrNoRowsAffected)
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

			reasonType := constant.XbDisbursementReasonTypeRefunded
			reasonDesc := constant.XbDisbursementReasonDescRefunded

			repo := New(mockMysql, mockLogger)
			err := repo.UpdateReasonOnly(ctx, uuid.NewString(), &reasonType, &reasonDesc)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)
		})
	}
}

func TestUpdateStatusAccountTransactionByReferenceID(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Update status account transaction by reference ID",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.PtrStringMockType(),
					constant.PtrStringMockType(),
					constant.TimeMockType(),
					constant.StringMockType(),
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Failure Update to Database",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.PtrStringMockType(),
					constant.PtrStringMockType(),
					constant.TimeMockType(),
					constant.StringMockType(),
				).Return(false, constant.ErrSomeErrorForUnitTest)

			},
			wantErr: true,
		},
		{
			name: "ERROR: No Rows Affected",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.PtrStringMockType(),
					constant.PtrStringMockType(),
					constant.TimeMockType(),
					constant.StringMockType(),
				).Return(false, constant.ErrNoRowsAffected)
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
			err := repo.UpdateStatusAccountTransactionByReferenceID(ctx, "payment-session-123", constant.StatusSuccess, nil, nil)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestUpdateFDSRiskAssessmentResultByID(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)

	tests := []struct {
		name      string
		setupMock func()
		wantError error
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				db.On(
					"ExecContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(false, assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				db.On(
					"ExecContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(true, nil)
			},
			wantError: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			assert.Equal(t, test.wantError, repo.UpdateFDSRiskAssessmentResultByID(t.Context(), "123456", fdscommon.TransactionAssessmentResponse{}))

			db.AssertExpectations(t)
		})
	}
}

func TestUpdateSettlementDetailByIDs(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)

	tests := []struct {
		name      string
		request   orchestratorModel.UpdateSettlementDetailRequest
		setupMock func()
		wantError error
	}{
		{
			name:      "ERROR:Empty parameters", // NOSONAR
			setupMock: func() { /* Empty Function */ },
			wantError: errors.New("at least one settlement detail field must be provided for update"),
		},
		{
			name: "ERROR:Some error", // NOSONAR
			request: orchestratorModel.UpdateSettlementDetailRequest{
				EstimateSettlementAt: util.ValueToPtr(time.Now()),
			},
			setupMock: func() {
				db.On(
					"ExecContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(false, assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name: "SUCCESS", // NOSONAR
			request: orchestratorModel.UpdateSettlementDetailRequest{
				EstimateSettlementAt: util.ValueToPtr(time.Now()),
			},
			setupMock: func() {
				db.On(
					"ExecContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(true, nil)
			},
			wantError: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			assert.Equal(t, test.wantError, repo.UpdateSettlementDetailByIDs(t.Context(), []string{"123456"}, test.request))

			db.AssertExpectations(t)
		})
	}
}

func TestUpdateSettlementHoldByReferenceID(t *testing.T) {
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	db := mysqlMocks.NewIMySqlExt(t)
	repo := New(db, logger)

	referenceID := uuid.NewString()

	testCases := []struct {
		name      string
		refID     string
		flag      bool
		mockSetup func()
		wantErr   bool
	}{
		{
			name:  "SUCCESS: Update settlement hold flag to true",
			refID: referenceID,
			flag:  true,
			mockSetup: func() {
				db.On(
					"ExecContext",
					mock.Anything,
					mock.Anything,
					"true",
					referenceID,
				).Once().Return(true, nil)
			},
			wantErr: false,
		},
		{
			name:  "SUCCESS: Update settlement hold flag to false",
			refID: referenceID,
			flag:  false,
			mockSetup: func() {
				db.On(
					"ExecContext",
					mock.Anything,
					mock.Anything,
					"false",
					referenceID,
				).Once().Return(true, nil)
			},
			wantErr: false,
		},
		{
			name:  "ERROR: Database update fails",
			refID: referenceID,
			flag:  true,
			mockSetup: func() {
				db.On(
					"ExecContext",
					mock.Anything,
					mock.Anything,
					"true",
					referenceID,
				).Once().Return(false, assert.AnError)
			},
			wantErr: true,
		},
		{
			name:  "SUCCESS: No matching rows (no error returned)",
			refID: referenceID,
			flag:  true,
			mockSetup: func() {
				db.On(
					"ExecContext",
					mock.Anything,
					mock.Anything,
					"true",
					referenceID,
				).Once().Return(false, nil)
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			err := repo.UpdateSettlementHoldByReferenceID(context.Background(), tc.refID, tc.flag)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			db.AssertExpectations(t)
		})
	}
}

func TestUpdateTransactionDetail(t *testing.T) {
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	db := mysqlMocks.NewIMySqlExt(t)
	repo := New(db, logger)

	defaultPayload := orchestratorModel.UpdateTransactionRequest{
		TransactionID: uuid.NewString(),
		Channel:       "CHANNEL",
	}

	testCases := []struct {
		name      string
		payload   orchestratorModel.UpdateTransactionRequest
		mockSetup func()
		wantErr   bool
	}{
		{
			name:    "SUCCESS: Update transaction detail",
			payload: defaultPayload,
			mockSetup: func() {
				db.On(
					"ExecContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(true, nil)
			},
			wantErr: false,
		},
		{
			name:    "ERROR: Empty Update Payload",
			payload: orchestratorModel.UpdateTransactionRequest{},
			mockSetup: func() {
			},
			wantErr: true,
		},
		{
			name:    "ERROR: Update transaction detail",
			payload: defaultPayload,
			mockSetup: func() {
				db.On(
					"ExecContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(false, fmt.Errorf("error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			err := repo.UpdateTransactionDetail(context.Background(), tc.payload)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			db.AssertExpectations(t)
		})
	}
}
