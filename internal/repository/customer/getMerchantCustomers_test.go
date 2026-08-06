package customerRepository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetMerchantCustomersByID(t *testing.T) {
	email := "VJ2jK@example.com"
	phone := "081234567890"

	validExpected := []*customerModel.CustomerDBModel{
		{
			UUID:        "123",
			FirstName:   "John Doe",
			Email:       sql.NullString{String: email, Valid: true},
			PhoneNumber: phone,
			MerchantID:  uuid.NewString(),
		},
	}

	testCases := []struct {
		desc       string
		customerID string
		wantErr    bool
		mockSetup  func(mysqlMock *mysqlMocks.IMySqlExt)
		expected   []*customerModel.CustomerDBModel
	}{
		{
			desc:       "SUCCESS: get merchant customers",
			customerID: "123",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]*customerModel.CustomerDBModel"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil).Run(func(args mock.Arguments) {
					data := args.Get(1).(*[]*customerModel.CustomerDBModel)
					*data = validExpected
				})
			},
			expected: validExpected,
		},
		{
			desc:       "ERROR: Customer Not Found",
			customerID: "123",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]*customerModel.CustomerDBModel"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(sql.ErrNoRows)
			},
			expected: nil,
			wantErr:  false,
		},
		{
			desc:       "ERROR: Database Error",
			customerID: "123",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]*customerModel.CustomerDBModel"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(errors.New("database error"))
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "customers")
			_, err := repo.GetMerchantCustomersByID(ctx, uuid.NewString(), []string{uuid.NewString(), uuid.NewString()})
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockMysql.AssertExpectations(t)

		})
	}
}
