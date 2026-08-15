package paymentCapture

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentCaptureModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/paymentCapture"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetByID(t *testing.T) {
	testCase := []struct {
		name      string
		id        string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
		wantNil   bool
	}{
		{
			name: "SUCCESS: Get payment capture by ID",
			id:   uuid.NewString(),
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*paymentCaptureModel.PaymentCapture"),
					constant.StringMockType(),
					mock.Anything,
				).Return(nil)
			},
			wantErr: false,
			wantNil: false,
		},
		{
			name: "SUCCESS: No rows found - returns nil, nil",
			id:   uuid.NewString(),
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*paymentCaptureModel.PaymentCapture"),
					constant.StringMockType(),
					mock.Anything,
				).Return(sql.ErrNoRows)
			},
			wantErr: false,
			wantNil: true,
		},
		{
			name: "ERROR: Database error",
			id:   uuid.NewString(),
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*paymentCaptureModel.PaymentCapture"),
					constant.StringMockType(),
					mock.Anything,
				).Return(errors.New("database connection failed"))
			},
			wantErr: true,
			wantNil: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			result, err := repo.GetByID(context.Background(), tc.id)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tc.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.IsType(t, &paymentCaptureModel.PaymentCapture{}, result)
			}

			mockMysql.AssertExpectations(t)
		})
	}
}
