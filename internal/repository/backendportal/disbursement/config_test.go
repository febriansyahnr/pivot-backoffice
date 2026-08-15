package disbursementRepository_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/disbursement"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/disbursement"
	mySqlExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetTransactionConfig(t *testing.T) {
	db := mySqlExtMock.NewIMySqlExt(t)

	repo := New(db, nil)

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    string
		wantResult *disbursementModel.TransactionConfig
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), c.PtrNullJSONText(), c.StringMockType(), c.StringMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: "some error", // NOSONAR
		},
		{
			name: "SUCCESS:Default config",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), c.PtrNullJSONText(), c.StringMockType(), c.StringMockType(),
				).Once().Return(sql.ErrNoRows)
			},
			wantResult: &disbursementModel.TransactionConfig{
				MinAmount: c.DisbursementMinAmount,
				MaxAmount: c.DisbursementMaxAmount,
			},
		},
		{
			name: "ERROR:Malformed JSON configuration",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), c.PtrNullJSONText(), c.StringMockType(), c.StringMockType(),
				).Once().Run(func(args mock.Arguments) {
					*args.Get(1).(*types.NullJSONText) = types.NullJSONText{
						Valid:    true,
						JSONText: []byte(`{A}`), // NOSONAR
					}
				}).Return(nil)
			},
			wantErr: "invalid character 'A' looking for beginning of object key string",
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), c.PtrNullJSONText(), c.StringMockType(), c.StringMockType(),
				).Run(func(args mock.Arguments) {
					*args.Get(1).(*types.NullJSONText) = types.NullJSONText{
						Valid:    true,
						JSONText: []byte(`{"minAmount": 30000, "maxAmount": 45000}`),
					}
				}).Return(nil)
			},
			wantResult: &disbursementModel.TransactionConfig{
				MinAmount: 30_000, MaxAmount: 45_000,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.GetTransactionConfig(context.Background(), uuid.NewString())
			if test.wantErr == "" {
				require.NoError(t, err)
				assert.Equal(t, test.wantResult, result)

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}

func TestGetDailyTransactionLimit(t *testing.T) {
	db := mySqlExtMock.NewIMySqlExt(t)

	config := &config.DisbursementConfig{
		DailyLimitMerchant:         10_000,
		DailyLimitMerchantPlatform: 50_000,
	}
	repo := New(db, nil, WithConfig(config))

	merchantId := "dc491715-4d3f-4aca-807a-1a7368af3f74"

	tests := []struct {
		name         string
		merchantType string
		setupMock    func()
		wantErr      error
		wantResult   *disbursementModel.DailyTransactionLimitResponse
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), mock.Anything, c.StringMockType(), c.TimeMockType(), c.TimeMockType(), c.TimeMockType(), c.TimeMockType(), merchantId,
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS:Merchant default value",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), mock.Anything, c.StringMockType(), c.TimeMockType(), c.TimeMockType(), c.TimeMockType(), c.TimeMockType(), merchantId,
				).Once().Run(func(args mock.Arguments) {
					(*args.Get(1).(*disbursementModel.DailyTransactionLimitResponse)) = disbursementModel.DailyTransactionLimitResponse{}
				}).Return(nil)
			},
			wantResult: &disbursementModel.DailyTransactionLimitResponse{
				Limit:     &config.DailyLimitMerchant,
				Remaining: config.DailyLimitMerchant,
			},
		},
		{
			name:         "SUCCESS:Merchant platform default value",
			merchantType: c.DisbursementDailyLimitMerchantPlatform,
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), mock.Anything, c.StringMockType(), c.TimeMockType(), c.TimeMockType(), c.TimeMockType(), c.TimeMockType(), merchantId,
				).Once().Run(func(args mock.Arguments) {
					(*args.Get(1).(*disbursementModel.DailyTransactionLimitResponse)) = disbursementModel.DailyTransactionLimitResponse{}
				}).Return(nil)
			},
			wantResult: &disbursementModel.DailyTransactionLimitResponse{
				Limit:     &config.DailyLimitMerchantPlatform,
				Remaining: config.DailyLimitMerchantPlatform,
			},
		},
		{
			name: "SUCCESS:Merchant already config",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), mock.Anything, c.StringMockType(), c.TimeMockType(), c.TimeMockType(), c.TimeMockType(), c.TimeMockType(), merchantId,
				).Once().Run(func(args mock.Arguments) {
					(*args.Get(1).(*disbursementModel.DailyTransactionLimitResponse)) = disbursementModel.DailyTransactionLimitResponse{
						Limit:     util.ValueToPtr(250_000.00),
						Processed: 10_000,
					}
				}).Return(nil)
			},
			wantResult: &disbursementModel.DailyTransactionLimitResponse{
				Limit:     util.ValueToPtr(250_000.00),
				Processed: 10_000,
				Remaining: 240_000,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.merchantType == "" {
				test.merchantType = c.DisbursementDailyLimitMerchant
			}
			test.setupMock()

			result, err := repo.GetDailyTransactionLimit(context.Background(), merchantId, test.merchantType)
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
