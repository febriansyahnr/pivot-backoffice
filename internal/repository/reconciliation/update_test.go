package reconciliation

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/reconciliation"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestReconciliationRepositoryUpdate(t *testing.T) {
	testCases := []struct {
		name      string
		input     *reconciliation.Reconciliation
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Update reconciliation record",
			input: &reconciliation.Reconciliation{
				UUID:           uuid.NewString(),
				ResultFilePath: "test/updated-result.csv",
				Status:         "PROCESSED",
				UpdatedAt:      time.Now(),
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("*reconciliation.Reconciliation"),
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: No rows affected",
			input: &reconciliation.Reconciliation{
				UUID:           uuid.NewString(),
				ResultFilePath: "test/updated-result.csv",
				Status:         "PROCESSED",
				UpdatedAt:      time.Now(),
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("*reconciliation.Reconciliation"),
				).Return(false, constant.ErrNoRowsAffected)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Database error during update",
			input: &reconciliation.Reconciliation{
				UUID:           uuid.NewString(),
				ResultFilePath: "test/updated-result.csv",
				Status:         "PROCESSED",
				UpdatedAt:      time.Now(),
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("*reconciliation.Reconciliation"),
				).Return(false, errors.New("database error"))
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

			err := repo.Update(context.Background(), tc.input)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
