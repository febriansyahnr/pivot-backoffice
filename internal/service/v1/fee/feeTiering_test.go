package feeService_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/fee"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestDetermineFeeTierLevel(t *testing.T) {

	feeService := &FeeService{}

	tests := []struct {
		name       string
		value      float64 // TPV or FREQUENCY
		tiers      []merchant.FeeTieringConfig
		wantResult *merchant.FeeTieringConfig
	}{
		{
			name:  "Flat-fee",
			value: 10_000_000,
		},
		{
			name:  "Single-tier",
			value: 10_000_000,
			tiers: []merchant.FeeTieringConfig{
				{
					Tier:       1,
					Min:        0,
					Max:        99_999_999_999,
					AmountType: "AMOUNT",
					Amount:     10_000,
					TaxType:    "NON_PKP",
				},
			},
			wantResult: &merchant.FeeTieringConfig{
				Tier:       1,
				Min:        0,
				Max:        99_999_999_999,
				AmountType: "AMOUNT",
				Amount:     10_000,
				TaxType:    "NON_PKP",
			},
		},
		{
			name:  "Two-tier",
			value: 10_000_000,
			tiers: []merchant.FeeTieringConfig{
				{
					Tier:          2,
					Min:           5_000_001,
					Max:           10_000_000,
					AmountType:    "PERCENTAGE",
					Percentage:    1,
					TaxType:       "EXCLUSIVE",
					TaxPercentage: 11,
				},
				{
					Tier:       1,
					Min:        0,
					Max:        5_000_000,
					AmountType: "AMOUNT",
					Amount:     10_000,
					TaxType:    "NON_PKP",
				},
			},
			wantResult: &merchant.FeeTieringConfig{
				Tier:          2,
				Min:           5_000_001,
				Max:           10_000_000,
				AmountType:    "PERCENTAGE",
				Percentage:    1,
				TaxType:       "EXCLUSIVE",
				TaxPercentage: 11,
			},
		},
		{
			name:  "Three-tier",
			value: 10_000_001,
			tiers: []merchant.FeeTieringConfig{
				{
					Tier:          3,
					Min:           10_000_001,
					Max:           99_999_999,
					AmountType:    "AMOUNT",
					Amount:        4_000,
					TaxType:       "EXCLUSIVE",
					TaxPercentage: 11,
				},
				{
					Tier:       1,
					Min:        0,
					Max:        5_000_000,
					AmountType: "AMOUNT",
					Amount:     10_000,
					TaxType:    "NON_PKP",
				},
				{
					Tier:          2,
					Min:           5_000_001,
					Max:           10_000_000,
					AmountType:    "AMOUNT",
					Amount:        5_000,
					TaxType:       "EXCLUSIVE",
					TaxPercentage: 11,
				},
			},
			wantResult: &merchant.FeeTieringConfig{
				Tier:          3,
				Min:           10_000_001,
				Max:           99_999_999,
				AmountType:    "AMOUNT",
				Amount:        4_000,
				TaxType:       "EXCLUSIVE",
				TaxPercentage: 11,
			},
		},
		{
			name:  "Four-tier",
			value: 7_500_000,
			tiers: []merchant.FeeTieringConfig{
				{
					Tier:          4,
					Min:           10_000_001,
					Max:           99_999_999_999,
					AmountType:    "AMOUNT",
					Amount:        2_500,
					TaxType:       "EXCLUSIVE",
					TaxPercentage: 11,
				},
				{
					Tier:          3,
					Min:           5_000_001,
					Max:           10_000_000,
					AmountType:    "AMOUNT",
					Amount:        5_000,
					TaxType:       "EXCLUSIVE",
					TaxPercentage: 11,
				},
				{
					Tier:       1,
					Min:        0,
					Max:        2_500_000,
					AmountType: "AMOUNT",
					Amount:     10_000,
					TaxType:    "NON_PKP",
				},
				{
					Tier:          2,
					Min:           2_500_001,
					Max:           5_000_000,
					AmountType:    "AMOUNT",
					Amount:        7_500,
					TaxType:       "INCLUSIVE",
					TaxPercentage: 11,
				},
			},
			wantResult: &merchant.FeeTieringConfig{
				Tier:          3,
				Min:           5_000_001,
				Max:           10_000_000,
				AmountType:    "AMOUNT",
				Amount:        5_000,
				TaxType:       "EXCLUSIVE",
				TaxPercentage: 11,
			},
		},
		{
			name:  "Five-tier",
			value: 1_000_00,
			tiers: []merchant.FeeTieringConfig{
				{
					Tier:          4,
					Min:           10_000_001,
					Max:           20_000_000,
					AmountType:    "AMOUNT",
					Amount:        2_500,
					TaxType:       "EXCLUSIVE",
					TaxPercentage: 11,
				},
				{
					Tier:          5,
					Min:           20_000_001,
					Max:           99_999_999_999,
					AmountType:    "AMOUNT",
					Amount:        2_500,
					TaxType:       "INCLUSIVE",
					TaxPercentage: 11,
				},
				{
					Tier:          3,
					Min:           5_000_001,
					Max:           10_000_000,
					AmountType:    "AMOUNT",
					Amount:        5_000,
					TaxType:       "EXCLUSIVE",
					TaxPercentage: 11,
				},
				{
					Tier:          1,
					Min:           0,
					Max:           2_500_000,
					AmountType:    "AMOUNT_PERCENTAGE",
					Amount:        10_000,
					Percentage:    1,
					MaxFeeAmount:  util.ValueToPtr(15_000.00),
					TaxType:       "EXCLUSIVE",
					TaxPercentage: 11,
				},
				{
					Tier:          2,
					Min:           2_500_001,
					Max:           5_000_000,
					AmountType:    "AMOUNT",
					Amount:        7_500,
					TaxType:       "INCLUSIVE",
					TaxPercentage: 11,
				},
			},
			wantResult: &merchant.FeeTieringConfig{
				Tier:          1,
				Min:           0,
				Max:           2_500_000,
				AmountType:    "AMOUNT_PERCENTAGE",
				Amount:        10_000,
				Percentage:    1,
				MaxFeeAmount:  util.ValueToPtr(15_000.00),
				TaxType:       "EXCLUSIVE",
				TaxPercentage: 11,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			assert.Equal(t, test.wantResult, feeService.DetermineFeeTierLevel(context.Background(), test.value, test.tiers))

			for i := 0; i < len(test.tiers); i++ {
				assert.Equal(t, i+1, test.tiers[i].Tier)
			}
		})
	}
}

func TestDetermineFeeTierLvlFromMonthlyTPV(t *testing.T) {

	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
	merchantRepo := repoMocks.NewIMerchantRepository(t)
	accountTrxRepo := repoMocks.NewIAccountTransactionRepository(t)

	feeService := New(logger, nil, merchantRepo, WithAccountTransactionRepository(accountTrxRepo))

	defDate := time.Date(2024, 10, 1, 0, 15, 0, 0, tz)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Merchant Id -> Merchant fee tiering config
	merchantId := uuid.NewString()
	merchantFeeId := uuid.NewString()
	subMerchantIds := []string{uuid.NewString(), uuid.NewString(), uuid.NewString()}
	merchantFeeThatUseTiers := map[string][]merchant.MerchantFeeThatUseTier{
		merchantId: {
			{
				Id:             merchantFeeId,
				MerchantId:     merchantId,
				Reference:      c.ReferenceDisbursement,
				TieringType:    c.FrequencyTieringType,
				TieringConfigs: []merchant.FeeTieringConfig{{}},
			},
		},
	}
	platformActivityFee := map[string][]merchant.MerchantFeeThatUseTier{
		merchantId + "_" + c.ReferencePlatformActivity: {
			{
				Id:             merchantFeeId,
				MerchantId:     merchantId,
				Reference:      c.ReferencePlatformActivity,
				TieringType:    c.TPVTieringType,
				TieringConfigs: []merchant.FeeTieringConfig{{}},
			},
		},
	}

	tests := []struct {
		name      string
		date      time.Time
		setupMock func()
		wantErr   error
	}{
		{
			name:    "ERROR:Can only be run on the 1st",
			date:    time.Date(2024, 10, 2, 1, 0, 0, 0, tz),
			wantErr: nil,
			setupMock: func() {
			},
		},
		{
			name: "ERROR:Get list merchant fee that use tiers",
			setupMock: func() {
				merchantRepo.On(
					"GetListMerchantFeeThatUseTiers", c.ValueCtxMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Errorf("get list merchant fee that use tiers: %w", c.ErrSomeErrorForUnitTest),
		},
		{
			name: "SUCCESS:No merchant fee config using tiers",
			setupMock: func() {
				merchantRepo.On(
					"GetListMerchantFeeThatUseTiers", c.ValueCtxMockType(),
				).Once().Return(nil, nil)
			},
		},
		{
			name: "ERROR:Empty merchant fee list",
			setupMock: func() {
				merchantRepo.On(
					"GetListMerchantFeeThatUseTiers", c.ValueCtxMockType(),
				).Once().Return(map[string][]merchant.MerchantFeeThatUseTier{"K": nil}, nil)
			},
			wantErr: errors.New("merchant fee list not found"),
		},
		{
			name: "ERROR:Get sub merchant id list by parent id",
			date: time.Date(2024, 8, 1, 0, 30, 0, 0, tz),
			setupMock: func() {
				merchantRepo.On(
					"GetListMerchantFeeThatUseTiers", c.ValueCtxMockType(),
				).Once().Return(platformActivityFee, nil)

				merchantRepo.On(
					"GetSubMerchantIdListByParentId", c.ValueCtxMockType(), merchantId,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
				// accountTrxRepo.On(
				// 	"CalculatingMerchantTPVToDetermineFeeTierLevel", c.ValueCtxMockType(), c.StringMockType(), time.Date(2024, 6, 30, 17, 0, 0, 0, time.UTC), time.Date(2024, 7, 31, 16, 59, 59, 0, time.UTC),
				// ).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			// wantErr: fmt.Errorf("calculating merchant tpv: %w", c.ErrSomeErrorForUnitTest),
			wantErr: fmt.Errorf("get sub merchant id list by parent id: %w", c.ErrSomeErrorForUnitTest),
		},
		{
			name: "ERROR:Calculating platform TPV",
			date: time.Date(2024, 8, 1, 0, 30, 0, 0, tz),
			setupMock: func() {
				merchantRepo.On(
					"GetListMerchantFeeThatUseTiers", c.ValueCtxMockType(),
				).Once().Return(platformActivityFee, nil)

				merchantRepo.On(
					"GetSubMerchantIdListByParentId", c.ValueCtxMockType(), merchantId,
				).Return(subMerchantIds, nil)
				accountTrxRepo.On(
					"CalculatingTPVForPlatformActivitiesToDetermineFeeTierLevel", c.ValueCtxMockType(), subMerchantIds, time.Date(2024, 6, 30, 17, 0, 0, 0, time.UTC), time.Date(2024, 7, 31, 16, 59, 59, 0, time.UTC),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Errorf("calculating merchant tpv: %w", c.ErrSomeErrorForUnitTest),
		},
		{
			name: "ERROR:Calculating merchant TPV",
			setupMock: func() {
				merchantRepo.On(
					"GetListMerchantFeeThatUseTiers", c.ValueCtxMockType(),
				).Return(merchantFeeThatUseTiers, nil)

				accountTrxRepo.On(
					"CalculatingMerchantTPVToDetermineFeeTierLevel", c.ValueCtxMockType(), c.StringMockType(), c.TimeMockType(), c.TimeMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Errorf("calculating merchant tpv: %w", c.ErrSomeErrorForUnitTest),
		},
		{
			name:      "ERROR:Context canceled",
			setupMock: func() { cancel() },
			wantErr:   context.Canceled,
		},
		{
			name: "ERROR:Applied fee from tiers",
			setupMock: func() {

				ctx = context.Background() // Reactivating context after canceled (ERROR:Context canceled)

				accountTrxRepo.On(
					"CalculatingMerchantTPVToDetermineFeeTierLevel", c.ValueCtxMockType(), c.StringMockType(), c.TimeMockType(), c.TimeMockType(),
				).Return(map[string]orchestrator_model.CalculatingMerchantTPVSummary{
					"DISBURSEMENT_BANK_TRANSFER": {
						Type:      "DISBURSEMENT",
						Channel:   "BANK_TRANSFER",
						Frequency: 1, Volume: 2,
					},
				}, nil)

				merchantRepo.On(
					"AppliedFeeFromTiers", c.ValueCtxMockType(), merchantFeeId, &merchant.FeeTieringConfig{},
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Errorf("applied fee from tiers: %w", c.ErrSomeErrorForUnitTest),
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				merchantRepo.On(
					"AppliedFeeFromTiers", c.ValueCtxMockType(), merchantFeeId, &merchant.FeeTieringConfig{},
				).Return(nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			if test.date.IsZero() {
				test.date = defDate
			}
			if test.setupMock != nil {
				test.setupMock()
			}

			assert.Equal(t, test.wantErr, feeService.DetermineFeeTierLvlFromMonthlyTPV(ctx, test.date))
		})
	}
}
