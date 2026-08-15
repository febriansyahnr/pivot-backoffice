package accounttransaction_repository

import (
	"context"
	"errors"
	"testing"

	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/orchestrator"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBulkInsert(t *testing.T) {
	input := []*orchestrator_model.AccountTransaction{{}}

	args := []interface{}{mock.AnythingOfType(constant.MockTypeValueContextReference)}
	for range 25 {
		args = append(args, mock.Anything)
	}

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Insert to Database",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On("ExecContext", args...).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Error insert",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On("ExecContext", args...).Return(false, errors.New("error"))

			},
			wantErr: true,
		},
		{
			name: "ERROR: No rows affected",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On("ExecContext", args...).Return(false, nil)

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
			err := repo.BulkInsert(context.Background(), input)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
