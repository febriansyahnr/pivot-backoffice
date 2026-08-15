package withdrawalRepository_test

import (
	"context"
	"database/sql"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/withdrawal"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/backendportal/withdrawal"
	mysqlMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFindById(t *testing.T) {
	db := mysqlMock.NewIMySqlExt(t)

	repo := New(db)

	resultMockType := mock.AnythingOfType("*withdrawal.Withdrawal")

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult *withdrawal.Withdrawal
	}{
		{
			name: "ERROR: Database error",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), resultMockType, c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS: Not found",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), resultMockType, c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(sql.ErrNoRows)
			},
			wantResult: nil,
		},
		{
			name: "SUCCESS: Found",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), resultMockType, c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Run(func(args mock.Arguments) {
					w := args.Get(1).(*withdrawal.Withdrawal)
					w.Id = "test-id"
					w.MerchantId = "test-merchant"
					w.BeneficiaryBankCode = "014"
					w.BeneficiaryAccountNo = "1234567890"
					w.BeneficiaryAccountName = "Test User"
					w.Amount = 100_000
				}).Return(nil)
			},
			wantResult: &withdrawal.Withdrawal{
				Id:                     "test-id",
				MerchantId:             "test-merchant",
				BeneficiaryBankCode:    "014",
				BeneficiaryAccountNo:   "1234567890",
				BeneficiaryAccountName: "Test User",
				Amount:                 100_000,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.FindById(context.Background(), "some-id", "some-merchant")
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
