package paymentCapture

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentCaptureModel "github.com/paper-indonesia/pivot-backoffice/internal/model/paymentCapture"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetByPaymentID(t *testing.T) {
	testCase := []struct {
		name      string
		paymentID string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name:      "SUCCESS: Get payment captures by payment ID",
			paymentID: uuid.NewString(),
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]*paymentCaptureModel.PaymentCapture"),
					constant.StringMockType(),
					mock.Anything,
				).Run(func(args mock.Arguments) {
					// Populate the slice with mock data
					captures := args.Get(1).(*[]*paymentCaptureModel.PaymentCapture)
					*captures = []*paymentCaptureModel.PaymentCapture{
						{
							ID:        uuid.NewString(),
							PaymentID: args.Get(3).(string),
							Status:    "CAPTURED",
							Currency:  "IDR",
							Amount:    10000,
						},
					}
				}).Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "SUCCESS: Get payment captures by payment ID - empty result",
			paymentID: uuid.NewString(),
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]*paymentCaptureModel.PaymentCapture"),
					constant.StringMockType(),
					mock.Anything,
				).Run(func(args mock.Arguments) {
					// Return empty slice
					captures := args.Get(1).(*[]*paymentCaptureModel.PaymentCapture)
					*captures = []*paymentCaptureModel.PaymentCapture{}
				}).Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "ERROR: Database error",
			paymentID: uuid.NewString(),
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]*paymentCaptureModel.PaymentCapture"),
					constant.StringMockType(),
					mock.Anything,
				).Return(errors.New("database connection failed"))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			result, err := repo.GetByPaymentID(context.Background(), tc.paymentID)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.IsType(t, []*paymentCaptureModel.PaymentCapture{}, result)
			}

			mockMysql.AssertExpectations(t)
		})
	}
}
