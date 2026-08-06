package reconciliation

import (
	"context"
	"database/sql"
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

func TestReconciliationRepositoryCreate(t *testing.T) {
	filePath := "test/path.csv"
	fileResultPath := "test/result.csv"
	testCases := []struct {
		name      string
		input     *reconciliation.Reconciliation
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Create reconciliation record",
			input: &reconciliation.Reconciliation{
				UUID:           uuid.NewString(),
				FilePath:       filePath,
				ResultFilePath: fileResultPath,
				Status:         "PENDING",
				Reasons: sql.NullString{
					String: "test",
					Valid:  true,
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
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
			name: "ERROR: Database error during creation",
			input: &reconciliation.Reconciliation{
				UUID:           uuid.NewString(),
				FilePath:       filePath,
				ResultFilePath: fileResultPath,
				Status:         "PENDING",
				CreatedAt:      time.Now(),
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
		{
			name: "ERROR: No rows affected",
			input: &reconciliation.Reconciliation{
				UUID:           uuid.NewString(),
				FilePath:       filePath,
				ResultFilePath: fileResultPath,
				Status:         "PENDING",
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("*reconciliation.Reconciliation"),
				).Return(false, nil)
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

			err := repo.Create(context.Background(), tc.input)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
