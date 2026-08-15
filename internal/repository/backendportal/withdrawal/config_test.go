package withdrawalRepository_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/merchant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/withdrawal"
	mysqlMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/jmoiron/sqlx/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetTransactionConfig(t *testing.T) {
	db := mysqlMock.NewIMySqlExt(t)

	merchantId := "5a814581-04c0-4154-be37-7de10169c5d4"

	repo := New(db)

	tests := []struct {
		name       string
		setupMock  func()
		wantError  error
		wantResult *merchant.WithdrawalConfig
	}{
		{
			name: "ERROR:Merchant not found",
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, merchantId,
				).Once().Return(sql.ErrNoRows)
			},
			wantError: pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrMerchantNotFound),
		},
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, merchantId,
				).Once().Return(assert.AnError)
			},
			wantError: pkgErrs.New(response.HttpErrDatabase, assert.AnError),
		},
		{
			name: "SUCCESS:Not config", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, merchantId,
				).Once().Return(nil)
			},
		},
		{
			name: "SUCCESS:Custom config", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, merchantId,
				).Once().Run(func(args mock.Arguments) {
					*args.Get(1).(*types.NullJSONText) = types.NullJSONText{
						Valid:    true,
						JSONText: []byte(`{"maxAmount": 20000, "minAmount": 10000}`),
					}
				}).Return(nil)
			},
			wantResult: &merchant.WithdrawalConfig{
				MinAmount: 10_000, MaxAmount: 20_000,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.GetTransactionConfig(context.Background(), merchantId)
			assert.Equal(t, test.wantError, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
