package merchantTopUp

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/merchantTopUp"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"

	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreate(t *testing.T) {
	data := &model.MerchantTopUp{
		ID:              "uuid-uuid-uuid",
		MerchantID:      "merchant-id",
		PaymentMethodID: "payment-method-id",
		ReferenceNumber: "reference-number",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	buf := new(bytes.Buffer)
	log := logger.NewSlogger(logger.Config{}, logger.WithSlogOutput(buf))

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		data      *model.MerchantTopUp
		wantErr   bool
	}{
		{
			name: "Valid merchant top up reference",
			data: data,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext", mock.Anything, mock.Anything, mock.Anything,
				).Return(true, nil)
			},
		},
		{
			name: "ERROR: Failure Insert to Database",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext", mock.Anything, mock.Anything, mock.Anything,
				).Return(false, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf.Reset()

			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, log)
			err := repo.Create(context.Background(), tc.data)

			if tc.wantErr {
				assert.Error(t, err)

			} else {
				assert.NoError(t, err)
			}
			mockMysql.AssertExpectations(t)
		})
	}
}
