package merchant_test

import (
	"context"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/merchant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/merchant"
	mysqlMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetDepositSetting(t *testing.T) {
	db := mysqlMock.NewIMySqlExt(t)

	repo := New(db, nil)

	response := merchant.DepositSettingResponse{
		AutoWithdrawal: "OFF",
	}
	resultMockType := mock.AnythingOfType("*merchant.DepositSettingResponse")

	tests := []struct {
		name       string
		setupMock  func()
		wantResult *merchant.DepositSettingResponse
		wantErr    error
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), resultMockType, c.StringMockType(), c.StringMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), resultMockType, c.StringMockType(), c.StringMockType(),
				).Run(func(args mock.Arguments) {
					(*args.Get(1).(*merchant.DepositSettingResponse)) = response
				}).Return(nil)
			},
			wantResult: &response,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.GetDepositSetting(context.Background(), "123456")
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestSetAutoWithdrawal(t *testing.T) {
	db := mysqlMock.NewIMySqlExt(t)

	repo := New(db, nil)

	tests := []struct {
		name      string
		setupMock func()
		wantErr   error
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				db.On(
					"ExecContext", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.TimeMockType(), c.StringMockType(),
				).Once().Return(false, c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				db.On(
					"ExecContext", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.TimeMockType(), c.StringMockType(),
				).Return(true, nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			assert.Equal(t, test.wantErr, repo.SetAutoWithdrawal(context.Background(), &merchant.AutoWithdrawalSettingRequest{}))
		})
	}
}
