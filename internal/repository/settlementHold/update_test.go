package settlementHold

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	settlementHold "github.com/paper-indonesia/pivot-backoffice/internal/model/settlementHolds"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdate(t *testing.T) {
	now := time.Now().UTC()
	data := &settlementHold.SettlementHold{
		UUID:      "uuid-123",
		Status:    "RELEASE",
		UpdatedAt: now,
	}

	history := &settlementHold.SettlementHoldHistory{
		UUID:             "history-uuid-456",
		SettlementHoldID: "uuid-123",
		Status:           "RELEASE",
		Reason:           "Investigation complete",
		CreatedBy:        "admin",
		CreatedAt:        now,
	}

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Update settlement hold with history",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				tx := &sqlx.Tx{}
				ctx := context.WithValue(context.Background(), mySqlExt.CtxSqlTx, tx)
				mysqlMock.On("BeginTxx", mock.Anything).Return(ctx, nil)
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					data.Status,
					data.UpdatedAt,
					data.UUID,
				).Return(true, nil)
				mysqlMock.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("*settlementHold.SettlementHoldHistory"),
				).Return(true, nil)
				mysqlMock.On("Commit", mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Begin transaction fails",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On("BeginTxx", mock.Anything).Return(context.Background(), errors.New("begin error"))
			},
			wantErr: true,
		},
		{
			name: "ERROR: Update settlement hold fails",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				tx := &sqlx.Tx{}
				ctx := context.WithValue(context.Background(), mySqlExt.CtxSqlTx, tx)
				mysqlMock.On("BeginTxx", mock.Anything).Return(ctx, nil)
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					data.Status,
					data.UpdatedAt,
					data.UUID,
				).Return(false, errors.New("update error"))
				mysqlMock.On("Rollback", mock.Anything).Return(nil)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Insert history fails",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				tx := &sqlx.Tx{}
				ctx := context.WithValue(context.Background(), mySqlExt.CtxSqlTx, tx)
				mysqlMock.On("BeginTxx", mock.Anything).Return(ctx, nil)
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					data.Status,
					data.UpdatedAt,
					data.UUID,
				).Return(true, nil)
				mysqlMock.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("*settlementHold.SettlementHoldHistory"),
				).Return(false, errors.New("history insert error"))
				mysqlMock.On("Rollback", mock.Anything).Return(nil)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Commit fails",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				tx := &sqlx.Tx{}
				ctx := context.WithValue(context.Background(), mySqlExt.CtxSqlTx, tx)
				mysqlMock.On("BeginTxx", mock.Anything).Return(ctx, nil)
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					data.Status,
					data.UpdatedAt,
					data.UUID,
				).Return(true, nil)
				mysqlMock.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("*settlementHold.SettlementHoldHistory"),
				).Return(true, nil)
				mysqlMock.On("Commit", mock.Anything).Return(errors.New("commit error"))
				mysqlMock.On("Rollback", mock.Anything).Return(nil)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			err := repo.Update(context.Background(), data, history)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)
		})
	}
}
