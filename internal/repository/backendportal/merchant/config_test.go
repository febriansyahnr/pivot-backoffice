package merchant_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	loggerMock "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/merchant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/merchant"
	mySqlExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/jmoiron/sqlx/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateTransactionConfig(t *testing.T) {
	db := mySqlExtMock.NewIMySqlExt(t)
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})

	repo := New(db, logger)

	merchantId := "4495457f-7664-4a7d-9526-6df1ca8a28be"

	tests := []struct {
		name      string
		config    *merchant.TransactionConfigs
		setupMock func()
		wantErr   error
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"ExecContext", c.ValueCtxMockType(), c.StringMockType(), c.SliceUint8MockType(), c.SliceUint8MockType(), c.TimeMockType(), c.StringMockType(),
				).Once().Return(false, c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS:With DailyDisbursement",
			config: &merchant.TransactionConfigs{
				DailyDisbursement: &merchant.DailyDisbursementConfig{},
			},
			setupMock: func() {
				db.On(
					"ExecContext", c.ValueCtxMockType(), c.StringMockType(), c.SliceUint8MockType(), c.SliceUint8MockType(), c.SliceUint8MockType(), c.TimeMockType(), c.StringMockType(),
				).Once().Return(true, nil)
			},
			wantErr: nil,
		},
		{
			name: "SUCCESS:Without DailyDisbursement",
			config: &merchant.TransactionConfigs{
				Disbursement: merchant.DisbursementConfig{
					MinAmount: 1000,
					MaxAmount: 10000,
				},
			},
			setupMock: func() {
				db.On(
					"ExecContext", c.ValueCtxMockType(), c.StringMockType(), c.SliceUint8MockType(), c.SliceUint8MockType(), c.TimeMockType(), c.StringMockType(),
				).Once().Return(true, nil)
			},
			wantErr: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			if test.config == nil {
				test.config = &merchant.TransactionConfigs{}
			}
			assert.Equal(t, test.wantErr, repo.UpdateTransactionConfig(context.Background(), merchantId, test.config))
		})
	}
}

func TestGetTransactionConfig(t *testing.T) {
	db := mySqlExtMock.NewIMySqlExt(t)

	disbursementConfig := config.DisbursementConfig{
		MinAmount:                  10_000,
		MaxAmount:                  1_000_000,
		DailyLimitMerchant:         10_000_000,
		DailyLimitMerchantPlatform: 50_000_000,
	}
	config := &config.Config{
		DisbursementConfig: disbursementConfig,
	}

	repo := New(db, nil, WithServiceConfig(config))

	merchantId := "662ea841-60ca-4461-9d6a-499df401af74"
	rawTransactionConfigMockType := mock.AnythingOfType("*merchant.RawTransactionConfig")

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult *merchant.TransactionConfigResp
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), rawTransactionConfigMockType, c.StringMockType(), c.StringMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "ERROR:Date not found",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), rawTransactionConfigMockType, c.StringMockType(), c.StringMockType(),
				).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "SUCCESS:Default value for merchant",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), rawTransactionConfigMockType, c.StringMockType(), c.StringMockType(),
				).Once().Run(func(args mock.Arguments) {
					(*args.Get(1).(*merchant.RawTransactionConfig)) = merchant.RawTransactionConfig{
						MerchantType: "MERCHANT",
					}
				}).Return(nil)
			},
			wantResult: &merchant.TransactionConfigResp{
				MerchantId:   merchantId,
				MerchantType: "MERCHANT",
				TransactionConfigs: merchant.TransactionConfigs{
					Disbursement: merchant.DisbursementConfig{
						MinAmount: disbursementConfig.MinAmount,
						MaxAmount: disbursementConfig.MaxAmount,
					},
					DailyDisbursement: &merchant.DailyDisbursementConfig{
						Merchant: disbursementConfig.DailyLimitMerchant,
					},
				},
			},
		},
		{
			name: "SUCCESS:Default value for merchant platform",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), rawTransactionConfigMockType, c.StringMockType(), c.StringMockType(),
				).Once().Run(func(args mock.Arguments) {
					(*args.Get(1).(*merchant.RawTransactionConfig)) = merchant.RawTransactionConfig{
						MerchantType: "MERCHANT_PLATFORM",
					}
				}).Return(nil)
			},
			wantResult: &merchant.TransactionConfigResp{
				MerchantId:   merchantId,
				MerchantType: "MERCHANT_PLATFORM",
				TransactionConfigs: merchant.TransactionConfigs{
					Disbursement: merchant.DisbursementConfig{
						MinAmount: disbursementConfig.MinAmount,
						MaxAmount: disbursementConfig.MaxAmount,
					},
					DailyDisbursement: &merchant.DailyDisbursementConfig{
						Merchant:         disbursementConfig.DailyLimitMerchant,
						MerchantPlatform: &disbursementConfig.DailyLimitMerchantPlatform,
					},
				},
			},
		},
		{
			name: "SUCCESS:Default value for sub-merchant",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), rawTransactionConfigMockType, c.StringMockType(), c.StringMockType(),
				).Once().Run(func(args mock.Arguments) {
					(*args.Get(1).(*merchant.RawTransactionConfig)) = merchant.RawTransactionConfig{
						MerchantType: "NON_KYC_SUB_MERCHANT",
					}
				}).Return(nil)
			},
			wantResult: &merchant.TransactionConfigResp{
				MerchantId:   merchantId,
				MerchantType: "NON_KYC_SUB_MERCHANT",
				TransactionConfigs: merchant.TransactionConfigs{
					Disbursement: merchant.DisbursementConfig{
						MinAmount: disbursementConfig.MinAmount,
						MaxAmount: disbursementConfig.MaxAmount,
					},
				},
			},
		},
		{
			name: "SUCCESS:Merchant platform",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), rawTransactionConfigMockType, c.StringMockType(), c.StringMockType(),
				).Once().Run(func(args mock.Arguments) {
					(*args.Get(1).(*merchant.RawTransactionConfig)) = merchant.RawTransactionConfig{
						MerchantType: "MERCHANT_PLATFORM",
						Disbursement: types.NullJSONText{
							JSONText: []byte(`{"minAmount": 10001, "maxAmount": 1000001}`),
						},
						DailyDisbursement: types.NullJSONText{
							JSONText: []byte(`{"merchant": 10000001, "merchantPlatform": 50000001}`),
						},
					}
				}).Return(nil)
			},
			wantResult: &merchant.TransactionConfigResp{
				MerchantId:   merchantId,
				MerchantType: "MERCHANT_PLATFORM",
				TransactionConfigs: merchant.TransactionConfigs{
					Disbursement: merchant.DisbursementConfig{
						MinAmount: 10_001,
						MaxAmount: 1_000_001,
					},
					DailyDisbursement: &merchant.DailyDisbursementConfig{
						Merchant:         10_000_001,
						MerchantPlatform: util.ValueToPtr[float64](50_000_001),
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.GetTransactionConfig(context.Background(), merchantId)
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestGetDisbursementMerchantConfig(t *testing.T) {
	db := mySqlExtMock.NewIMySqlExt(t)

	repo := New(db, nil)

	result := merchant.DisbursementMerchantConfig{
		DailyLimitMerchantId:   "ccc1a739-671f-4c06-a382-28acdfbcee50",
		DailyLimitMerchantType: c.DisbursementDailyLimitMerchant,
	}

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult *merchant.DisbursementMerchantConfig
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), mock.Anything, c.StringMockType(), c.StringMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr:    c.ErrSomeErrorForUnitTest,
			wantResult: &merchant.DisbursementMerchantConfig{},
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), mock.Anything, c.StringMockType(), c.StringMockType(),
				).Run(func(args mock.Arguments) {
					(*args.Get(1).(*merchant.DisbursementMerchantConfig)) = result
				}).Return(nil)
			},
			wantResult: &result,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.GetDisbursementMerchantConfig(context.Background(), "123456")
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestUpdateFDSConfig(t *testing.T) {
	db := mySqlExtMock.NewIMySqlExt(t)

	repo := New(db, nil)

	config := merchant.FDSConfigRequest{}
	raw, _ := json.Marshal(config)
	merchantID := "1e515704-836f-4d52-8fc9-d93c8b5c6ab0"

	tests := []struct {
		name      string
		setupMock func()
		wantError error
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				db.On(
					"ExecContext", mock.Anything, mock.Anything, types.JSONText(raw), mock.Anything, merchantID,
				).Once().Return(false, assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name: "ERROR:Data not found", // NOSONAR
			setupMock: func() {
				db.On(
					"ExecContext", mock.Anything, mock.Anything, types.JSONText(raw), mock.Anything, merchantID,
				).Once().Return(false, nil)
			},
			wantError: c.ErrNoRowsAffected,
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				db.On(
					"ExecContext", mock.Anything, mock.Anything, types.JSONText(raw), mock.Anything, merchantID,
				).Once().Return(true, nil)
			},
			wantError: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			assert.Equal(t, test.wantError, repo.UpdateFDSConfig(t.Context(), merchantID, config))
		})
	}
}

func TestGetFDSConfig(t *testing.T) {
	db := mySqlExtMock.NewIMySqlExt(t)

	fdsConfig := config.FDSFeatureProofOfPaymentConfig{
		Velocity: config.FDSRuleVelocityConfig{
			Enabled: true,
			Window: config.FDSWindowConfig{
				Interval: 1,      // NOSONAR
				Unit:     "HOUR", // NOSONAR
			},
			Threshold: config.FDSThresholdConfig{
				Count: 5,
			},
			Action: "BLOCK", // NOSONAR
		},
	}
	repo := New(
		db, nil, WithServiceConfig(&config.Config{
			FdsConfig: config.FdsConfig{
				Features: config.FDSFeaturesConfig{
					ProofOfPayment: fdsConfig,
				},
			},
		}),
	)

	merchantID := "cac48c05-045e-42c1-af74-3f8d4bbea111"

	tests := []struct {
		name       string
		setupMock  func()
		wantError  error
		wantResult *merchant.GetFDSConfigResponse
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				db.On("GetContext", mock.Anything, mock.Anything, mock.Anything, merchantID).Once().Return(assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name: "ERROR:Data not found", // NOSONAR
			setupMock: func() {
				db.On("GetContext", mock.Anything, mock.Anything, mock.Anything, merchantID).Once().Return(sql.ErrNoRows)
			},
			wantError: nil, wantResult: nil,
		},
		{
			name: "SUCCESS:Default FDS config", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, merchantID,
				).Once().Return(nil)
			},
			wantResult: &merchant.GetFDSConfigResponse{
				FDSConfig: merchant.FDSConfig{
					ProofOfPayment: &merchant.FDSFeatureProofOfPayment{
						Velocity: merchant.FDSRuleVelocityConfig{
							Enabled: fdsConfig.Velocity.Enabled,
							Window: merchant.FDSWindowConfig{
								Interval: fdsConfig.Velocity.Window.Interval,
								Unit:     fdsConfig.Velocity.Window.Unit,
							},
							Threshold: merchant.FDSThresholdConfig{
								Count: fdsConfig.Velocity.Threshold.Count,
							},
							Action: fdsConfig.Velocity.Action,
						},
					},
				},
			},
		},
		{
			name: "SUCCESS:Merchant FDS config", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, merchantID,
				).Once().Run(func(args mock.Arguments) {
					*args.Get(1).(*merchant.GetFDSConfigResponse) = merchant.GetFDSConfigResponse{
						MerchantID: merchantID,
						RawFDSConfig: types.NullJSONText{
							Valid: true, JSONText: []byte(`{"proofOfPayment": {"velocity": {"action": "BLOCK", "window": {"unit": "MINUTE", "interval": 1}, "enabled": true, "threshold": {"count": 10}}}}`),
						},
					}
				}).Return(nil)
			},
			wantResult: &merchant.GetFDSConfigResponse{
				MerchantID: merchantID,
				FDSConfig: merchant.FDSConfig{
					ProofOfPayment: &merchant.FDSFeatureProofOfPayment{
						Velocity: merchant.FDSRuleVelocityConfig{
							Enabled: true,
							Window: merchant.FDSWindowConfig{
								Interval: 1,
								Unit:     "MINUTE",
							},
							Threshold: merchant.FDSThresholdConfig{
								Count: 10,
							},
							Action: "BLOCK",
						},
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			config, err := repo.GetFDSConfig(t.Context(), merchantID)
			assert.Equal(t, test.wantError, err)
			assert.Equal(t, test.wantResult, config)

			db.AssertExpectations(t)
		})
	}
}

func TestUpdatePaymentInvestigationConfig(t *testing.T) {
	db := mySqlExtMock.NewIMySqlExt(t)

	repo := New(db, nil)

	merchantID := "107725b5-8444-406b-9bf8-9d3d8cb06322"

	tests := []struct {
		name      string
		setupMock func()
		wantError error
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				db.On(
					"ExecContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, merchantID,
				).Once().Return(false, assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				db.On(
					"ExecContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, merchantID,
				).Once().Return(true, nil)
			},
			wantError: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			assert.Equal(t, test.wantError, repo.UpdatePaymentInvestigationConfig(t.Context(), merchantID, merchant.PaymentInvestigationConfigRequest{}))
		})
	}
}

func TestIsInvestigationFlowEnabled(t *testing.T) {
	db := mySqlExtMock.NewIMySqlExt(t)

	repo := New(db, nil)

	enabled := false
	merchantID := "598c9032-3791-4114-8622-89532cbb61ce"

	tests := []struct {
		name       string
		setupMock  func()
		wantError  error
		wantResult bool
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, &enabled, mock.Anything, merchantID,
				).Once().Return(assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name: "SUCCESS:Investigation disabled", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, &enabled, mock.Anything, merchantID,
				).Once().Return(nil)
			},
			wantError:  nil,
			wantResult: false,
		},
		{
			name: "SUCCESS:Investigation enabled", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, &enabled, mock.Anything, merchantID,
				).Once().Run(func(args mock.Arguments) {
					*args.Get(1).(*bool) = true
				}).Return(nil)
			},
			wantError:  nil,
			wantResult: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.IsInvestigationFlowEnabled(t.Context(), merchantID)
			assert.Equal(t, test.wantError, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
