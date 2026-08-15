package accounttransaction_repository

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"testing"

	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetLedgerDetail(t *testing.T) {
	testCases := []struct {
		name        string
		referenceId string
		mockSetup   func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr     bool
	}{
		{
			name:        "SUCCESS: Get ledger detail",
			referenceId: "valid-reference-id",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					"valid-reference-id",
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name:        "ERROR: Cant get ledger detail with given reference id",
			referenceId: "invalid-reference-id",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					"invalid-reference-id",
				).Return(errors.New("error"))
			},
			wantErr: true,
		},
		{
			name:        "ERROR: Data not found",
			referenceId: "other-reference-id",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.Anything, mock.Anything, mock.Anything, "other-reference-id",
				).Return(sql.ErrNoRows)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var wg sync.WaitGroup
			wg.Add(2)
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.Background()
			_, err := repo.GetLedgerDetail(ctx, tc.referenceId)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockMysql.AssertExpectations(t)

		})
	}
}
