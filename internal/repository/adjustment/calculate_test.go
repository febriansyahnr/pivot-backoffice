package adjustment_test

import (
	"context"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/adjustment"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestCalculateAmountBalanceAdjustmentForTopUp(t *testing.T) {
	sumAmount := 0.0
	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*float64"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil).Run(func(args mock.Arguments) {
					amountPtr := args.Get(1).(*float64)
					*amountPtr = sumAmount
				})
			},
			wantErr: false,
		},
		{
			name: "ERROR",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*float64"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mysqlMock := &mysqlMocks.IMySqlExt{}

			tc.mockSetup(mysqlMock)
			repo := New(mysqlMock)
			ctx := context.Background()
			res, err := repo.CalculateAmountBalanceAdjustmentForTopUp(ctx, uuid.NewString())

			if tc.wantErr {
				assert.Error(t, err)
				assert.Empty(t, res)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
