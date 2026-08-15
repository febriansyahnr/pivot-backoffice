package paymentRepository

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBeginTransaction(t *testing.T) {
	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Begin transaction",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				tx := &sqlx.Tx{}
				ctx := context.WithValue(context.Background(), mySqlExt.CtxSqlTx, tx)
				mysqlMock.On("BeginTxx", mock.Anything).Return(ctx, nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Database begin transaction fails",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On("BeginTxx", mock.Anything).Return(nil, fmt.Errorf("some-error"))

			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			ctx := context.Background()

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			newCtx, err := repo.BeginTransaction(ctx)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, newCtx)
			}

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestCommitTransaction(t *testing.T) {
	// Define test cases
	tests := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "Success commit",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On("Commit", constant.ValueCtxMockType()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "Fail commit",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On("Commit", constant.ValueCtxMockType()).Return(errors.New("commit error"))

			},
			wantErr: true,
		},
	}

	// Execute test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			ctx := context.Background()

			tt.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			err := repo.CommitTransaction(ctx)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestRollbackTransaction(t *testing.T) {
	// Define test cases
	tests := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "Success rollback",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On("Rollback", constant.ValueCtxMockType()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "Fail rollback",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On("Rollback", constant.ValueCtxMockType()).Return(errors.New("rollback error"))

			},
			wantErr: true,
		},
	}

	// Execute test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			ctx := context.Background()

			tt.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			err := repo.RollbackTransaction(ctx)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)

		})
	}
}
