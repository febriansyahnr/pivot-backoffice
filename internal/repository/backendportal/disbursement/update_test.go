package disbursementRepository

import (
	"context"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateProcessorReferenceId(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType(constant.MockTypeTime),
					mock.AnythingOfType("string"),
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Failure Update to Database",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType(constant.MockTypeTime),
					mock.AnythingOfType("string"),
				).Return(false, constant.ErrSomeErrorForUnitTest)

			},
			wantErr: true,
		},
		{
			name: "ERROR: No Rows Affected",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType(constant.MockTypeTime),
					mock.AnythingOfType("string"),
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
			err := repo.UpdateProcessorReferenceIdAndBankReferenceNo(ctx, uuid.NewString(), uuid.NewString(), "WRef1234")
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestUpdateReasonByIDs(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
		inputIDs  []string
	}{
		{
			name: "SUCCESS",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType(constant.MockTypeTime),
				).Return(true, nil)
			},
			wantErr:  false,
			inputIDs: []string{uuid.NewString(), uuid.NewString()},
		},
		{
			name: "ERROR: Failure Update to Database",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType(constant.MockTypeTime),
				).Return(false, constant.ErrSomeErrorForUnitTest)

			},
			wantErr:  true,
			inputIDs: []string{uuid.NewString(), uuid.NewString()},
		},
		{
			name: "ERROR: No Rows Affected",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType(constant.MockTypeTime),
				).Return(false, constant.ErrNoRowsAffected)
			},
			wantErr:  false,
			inputIDs: []string{uuid.NewString(), uuid.NewString()},
		},
		{
			name: "ERROR: No disbursement to update",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
			},
			wantErr:  true,
			inputIDs: []string{},
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			ctx := context.Background()

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			err := repo.UpdateReasonByIDs(ctx, tc.inputIDs, "", "")
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestUpdateReversalTransaction(t *testing.T) {
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
					"ExecContext", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(), constant.StringMockType(), constant.TimeMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Return(false, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				db.On(
					"ExecContext", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(), constant.StringMockType(), constant.TimeMockType(), constant.StringMockType(), constant.StringMockType(),
				).Return(true, nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			assert.Equal(t, test.wantErr, repo.UpdateReversalTransaction(context.Background(), "123", constant.ReasonTypeReversal, "bla bla", "John"))
		})
	}
}

func TestUpdateProcessorReferenceById(t *testing.T) {
	testCases := []struct {
		desc      string
		wantErr   bool
		setupMock func(db *mysqlMocks.IMySqlExt)
	}{
		{
			desc:    "ERROR Update Processor Reference",
			wantErr: true,
			setupMock: func(db *mysqlMocks.IMySqlExt) {
				db.On(
					`NamedExecContext`,
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("*disbursementModel.Disbursement"),
				).Return(false, constant.ErrSomeErrorForUnitTest)

			},
		},
		{
			desc:    "SUCCESS Update Processor Reference",
			wantErr: false,
			setupMock: func(db *mysqlMocks.IMySqlExt) {
				db.On(
					`NamedExecContext`,
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("*disbursementModel.Disbursement"),
				).Return(true, nil)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			db := mysqlMocks.NewIMySqlExt(t)
			logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.setupMock(db)

			repo := New(db, logger)

			err := repo.UpdateProcessorReferenceByID(context.Background(), &disbursementModel.Disbursement{})
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateStatusAndReasonByID(t *testing.T) {
	reasonType := "REVERSAL"
	reasonDescription := "Transaction reversed"

	testCases := []struct {
		desc      string
		wantErr   bool
		setupMock func(db *mysqlMocks.IMySqlExt)
	}{
		{
			desc:    "SUCCESS",
			wantErr: false,
			setupMock: func(db *mysqlMocks.IMySqlExt) {
				db.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("*string"),
					mock.AnythingOfType("*string"),
					mock.AnythingOfType(constant.MockTypeTime),
					mock.AnythingOfType("string"),
				).Return(true, nil)
			},
		},
		{
			desc:    "ERROR: Failure Update to Database",
			wantErr: true,
			setupMock: func(db *mysqlMocks.IMySqlExt) {
				db.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("*string"),
					mock.AnythingOfType("*string"),
					mock.AnythingOfType(constant.MockTypeTime),
					mock.AnythingOfType("string"),
				).Return(false, constant.ErrSomeErrorForUnitTest)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			db := mysqlMocks.NewIMySqlExt(t)
			logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.setupMock(db)

			repo := New(db, logger)

			err := repo.UpdateStatusAndReasonByID(context.Background(), uuid.NewString(), constant.StatusSuccess, &reasonType, &reasonDescription)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			db.AssertExpectations(t)
		})
	}
}

func TestUpdateBankReferenceNo(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType(constant.MockTypeTime),
					mock.AnythingOfType("string"),
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Failure Update to Database",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType(constant.MockTypeTime),
					mock.AnythingOfType("string"),
				).Return(false, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name: "ERROR: No Rows Affected",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType(constant.MockTypeTime),
					mock.AnythingOfType("string"),
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
			err := repo.UpdateBankReferenceNo(ctx, uuid.NewString(), "WRef1234")
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)
		})
	}
}
