package merchant_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/product"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/merchant"
	pdkLoggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	redisExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTransactionConfig(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
	rdb := redisExtMock.NewIRedisExt(t)
	repo := repoMocks.NewIMerchantRepository(t)
	productRepo := repoMocks.NewIProductRepository(t)

	merchantId := "e4a5c634-13cd-43f8-ad40-8d92cc773acf"
	service := New(repo, logger, nil, nil, nil, nil, WithRedisClient(rdb), WithProductRepository(productRepo))

	rdb.On(
		"Del", c.ValueCtxMockType(),
		fmt.Sprintf(c.DisbursementTransactionConfigFmt, merchantId),
		fmt.Sprintf(c.DailyDisbursementTransactionConfigFmt, merchantId, c.DisbursementDailyLimitMerchant),
		fmt.Sprintf(c.DailyDisbursementTransactionConfigFmt, merchantId, c.DisbursementDailyLimitMerchantPlatform),
	).Return(&redis.IntCmd{})
	repo.On(
		"UpdateTransactionConfig", c.ValueCtxMockType(), merchantId, mock.Anything,
	).Return(nil)

	withdrawalConfig := merchant.WithdrawalConfig{
		MinAmount: 10_000,
		MaxAmount: 250_000_000,
	}
	disbursementConfig := merchant.DisbursementConfig{
		MinAmount: 10_000,
		MaxAmount: 250_000_000,
	}

	tests := []struct {
		name      string
		config    *merchant.TransactionConfigs
		setupMock func()
		wantErr   error
	}{
		{
			name: "ERROR:Find merchant by id",
			setupMock: func() {
				repo.On(
					"FindMerchantByID", c.ValueCtxMockType(), merchantId,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: pkgErr.New(response.HttpErrDatabase, c.ErrSomeErrorForUnitTest),
		},
		{
			name: "ERROR:Merchant not found",
			setupMock: func() {
				repo.On(
					"FindMerchantByID", c.ValueCtxMockType(), merchantId,
				).Once().Return(nil, nil)
			},
			wantErr: pkgErr.New(response.HttpErrRequest, c.ErrMerchantNotFound),
		},
		{
			name:   "ERROR:Non KYC sub-merchant",
			config: &merchant.TransactionConfigs{},
			setupMock: func() {
				repo.On(
					"FindMerchantByID", c.ValueCtxMockType(), merchantId,
				).Once().Return(&merchant.Merchant{ParentID: sql.NullString{Valid: true}, KYCStatus: sql.NullString{String: c.KYCStatusNotRequired}}, nil)
			},
			wantErr: pkgErr.New(response.HttpErrUnprocessableContent, errors.New("transaction config applies only to platform merchants, merchants, and KYC sub merchants")),
		},
		{
			name:   "ERROR:Merchant does not set daily limit",
			config: &merchant.TransactionConfigs{},
			setupMock: func() {
				repo.On("FindMerchantByID", c.ValueCtxMockType(), merchantId).Once().Return(&merchant.Merchant{}, nil)
			},
			wantErr: pkgErr.New(response.HttpErrUnprocessableContent, errors.New("daily transaction limit must be set")),
		},
		{
			name: "ERROR:Invalid config",
			config: &merchant.TransactionConfigs{
				DailyDisbursement: &merchant.DailyDisbursementConfig{},
			},
			setupMock: func() {
				repo.On("FindMerchantByID", c.ValueCtxMockType(), merchantId).Once().Return(&merchant.Merchant{}, nil)
			},
			wantErr: assert.AnError,
		},
		{
			name: "ERROR:Daily transaction limit below max transaction",
			config: &merchant.TransactionConfigs{
				Disbursement:      disbursementConfig,
				Withdrawal:        withdrawalConfig,
				DailyDisbursement: &merchant.DailyDisbursementConfig{},
			},
			setupMock: func() {
				repo.On("FindMerchantByID", c.ValueCtxMockType(), merchantId).Return(&merchant.Merchant{}, nil)
			},
			wantErr: pkgErr.New(response.HttpErrUnprocessableContent, errors.New("daily transaction limit (merchant) must be gte to the max transaction")),
		},
		{
			name: "ERROR:Get merchant selected product by name",
			setupMock: func() {
				productRepo.On(
					"GetMerchantSelectedProductByName", c.ValueCtxMockType(), merchantId, c.ProductPlatform,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: pkgErr.New(response.HttpErrDatabase, c.ErrSomeErrorForUnitTest),
		},
		{
			name: "ERROR:Not a merchant platform",
			config: &merchant.TransactionConfigs{
				Disbursement: disbursementConfig,
				Withdrawal:   withdrawalConfig,
				DailyDisbursement: &merchant.DailyDisbursementConfig{
					Merchant:         250_000_000, // NOSONAR
					MerchantPlatform: util.ValueToPtr(1.0),
				},
			},
			setupMock: func() {
				productRepo.On(
					"GetMerchantSelectedProductByName", c.ValueCtxMockType(), merchantId, c.ProductPlatform,
				).Once().Return(nil, nil)
			},
			wantErr: pkgErr.New(response.HttpErrUnprocessableContent, errors.New("merchant does not activate platform product")),
		},
		{
			name: "ERROR:Merchant platform not set daily limit",
			setupMock: func() {
				productRepo.On(
					"GetMerchantSelectedProductByName", c.ValueCtxMockType(), merchantId, c.ProductPlatform,
				).Return(&product.MerchantWithProductName{}, nil)
			},
			wantErr: pkgErr.New(response.HttpErrUnprocessableContent, errors.New("merchant platform daily transaction limit must be set")),
		},
		{
			name: "ERROR:Daily transaction limit (Platform) below max transaction",
			config: &merchant.TransactionConfigs{
				Disbursement: disbursementConfig,
				Withdrawal:   withdrawalConfig,
				DailyDisbursement: &merchant.DailyDisbursementConfig{
					Merchant:         250_000_000,          // NOSONAR
					MerchantPlatform: util.ValueToPtr(1.0), // NOSONAR
				},
			},
			setupMock: func() { /* No Body */ }, // NOSONAR
			wantErr:   pkgErr.New(response.HttpErrUnprocessableContent, errors.New("daily transaction limit (merchant platform) must be gte to the max transaction")),
		},
		{
			name: "SUCCESS",
			config: &merchant.TransactionConfigs{
				Disbursement: disbursementConfig,
				Withdrawal:   withdrawalConfig,
				DailyDisbursement: &merchant.DailyDisbursementConfig{
					Merchant:         250_000_000,                     // NOSONAR
					MerchantPlatform: util.ValueToPtr(250_000_000.00), // NOSONAR
				},
			},
			setupMock: func() { /* No Body */ }, // NOSONAR
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			if test.config == nil {
				test.config = &merchant.TransactionConfigs{
					Disbursement:      disbursementConfig,
					Withdrawal:        withdrawalConfig,
					DailyDisbursement: &merchant.DailyDisbursementConfig{Merchant: 250_000_000},
				}
			}

			err := service.TransactionConfig(context.Background(), merchantId, test.config)

			if test.wantErr != nil {
				require.Error(t, err)

				var vldErrs validator.ValidationErrors
				if !errors.As(err, &vldErrs) {
					assert.Equal(t, test.wantErr, err)
				}
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestUpdateSettlementConfig(t *testing.T) {
	repo := repoMocks.NewIMerchantRepository(t)
	service := New(repo, nil, nil, nil, nil, nil)

	tests := []struct {
		name           string
		setupMock      func()
		wantErr        string
		config         *merchant.SettlementConfig
		isEmptyRequest bool
	}{
		{
			name: "ERROR: Empty settlement config",
			setupMock: func() {
				// empty setup
			},
			wantErr:        "invalid config",
			isEmptyRequest: true,
		},
		{
			name: "ERROR: GetMerchantFeeByID repo",
			setupMock: func() {
				repo.On(
					"GetMerchantFeeByID", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: "some error",
		},
		{
			name: "ERROR: GetMerchantFeeByID not found",
			setupMock: func() {
				repo.On(
					"GetMerchantFeeByID", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(nil, nil)
			},
			wantErr: "data not found",
		},
		{
			name: "SUCCESS: UpdateMerchantFee now rows affected",
			setupMock: func() {
				repo.On(
					"GetMerchantFeeByID", c.ValueCtxMockType(), c.StringMockType(),
				).Return(&merchant.MerchantFee{}, nil).Once()

				repo.On(
					"UpdateMerchantFee", c.ValueCtxMockType(), c.PtrMerchantFeeMockType(),
				).Once().Return(c.ErrNoRowsAffected)
			},
		},
		{
			name: "ERROR: UpdateMerchantFee general error",
			setupMock: func() {
				repo.On(
					"GetMerchantFeeByID", c.ValueCtxMockType(), c.StringMockType(),
				).Return(&merchant.MerchantFee{}, nil).Once()

				repo.On(
					"UpdateMerchantFee", c.ValueCtxMockType(), c.PtrMerchantFeeMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: "some error",
		},
		{
			name: "ERROR: Invalid config for INSTANT merchant fee settlement method",
			setupMock: func() {
				repo.On(
					"GetMerchantFeeByID", c.ValueCtxMockType(), c.StringMockType(),
				).Return(&merchant.MerchantFee{SettlementMethod: util.ValueToPtr("INSTANT")}, nil).Once()
			},
			wantErr: "invalid settlement config type for instant settlement",
			config: &merchant.SettlementConfig{
				Type: "T+1",
			},
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				repo.On(
					"GetMerchantFeeByID", c.ValueCtxMockType(), c.StringMockType(),
				).Return(&merchant.MerchantFee{}, nil).Once()

				repo.On(
					"UpdateMerchantFee", c.ValueCtxMockType(), c.PtrMerchantFeeMockType(),
				).Return(nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()
			settlementConfig := &merchant.SettlementConfig{}
			if test.config != nil {
				settlementConfig = test.config
			}
			if test.isEmptyRequest {
				settlementConfig = nil
			}

			if err := service.UpdateSettlementConfig(context.Background(), uuid.NewString(), settlementConfig); test.wantErr == "" {
				assert.NoError(t, err)

			} else {
				assert.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}

func TestGetTransactionConfig(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
	repo := repoMocks.NewIMerchantRepository(t)

	service := New(repo, logger, nil, nil, nil, nil)

	result := &merchant.TransactionConfigResp{
		MerchantId: "935c9e6f-9930-488a-a8cb-bc41d4ddeb8c",
	}

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult *merchant.TransactionConfigResp
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				repo.On(
					"GetTransactionConfig", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: pkgErr.New(response.HttpErrDatabase, c.ErrSomeErrorForUnitTest),
		},
		{
			name: "ERROR:Data not found",
			setupMock: func() {
				repo.On(
					"GetTransactionConfig", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(nil, nil)
			},
			wantErr: pkgErr.New(response.HttpErrUnprocessableContent, c.ErrDataNotFound),
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				repo.On(
					"GetTransactionConfig", c.ValueCtxMockType(), c.StringMockType(),
				).Return(result, nil)
			},
			wantResult: result,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := service.GetTransactionConfig(context.Background(), result.MerchantId)
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestGetMerchantIdForConfigs(t *testing.T) {
	log := pdkLoggerMock.NewILogger(t)
	merchantRepo := repoMocks.NewIMerchantRepository(t)

	service := New(merchantRepo, log, nil, nil, nil, nil)

	ctx := context.Background()
	merchantId := "f5d7c4bf-3e34-4a13-98b4-57f21465be98"
	parentMerchantId := "cb05f54f-e8be-4973-8749-1dcaddce8041"
	subMerchantData := &merchant.Merchant{
		UUID:      merchantId,
		ParentID:  sql.NullString{String: parentMerchantId},
		KYCStatus: sql.NullString{String: c.KYCStatusNotRequired},
	}

	tests := []struct {
		name           string
		setMerchantCtx bool
		setupMock      func()
		wantErr        error
		wantResult     *merchant.MerchantIdForConfigs
	}{
		{
			name: "ERROR:Find merchant by id",
			setupMock: func() {
				merchantRepo.On("FindMerchantByID", mock.Anything, merchantId).Once().Return(nil, c.ErrSomeErrorForUnitTest)
				log.On(
					"Error", mock.Anything, "Failed while find merchant by id", mock.Anything,
				).Once().Return()
			},
			wantErr: pkgErr.New(response.HttpErrDatabase, c.ErrSomeErrorForUnitTest),
		},
		{
			name: "ERROR:Find merchant by id",
			setupMock: func() {
				merchantRepo.On("FindMerchantByID", mock.Anything, merchantId).Once().Return(nil, nil)
			},
			wantErr: pkgErr.New(response.HttpErrRequest, c.ErrMerchantNotFound),
		},
		{
			name: "SUCCESS:Merchant",
			setupMock: func() {
				merchantRepo.On("FindMerchantByID", mock.Anything, merchantId).Once().Return(&merchant.Merchant{
					UUID: merchantId,
				}, nil)
			},
			wantResult: &merchant.MerchantIdForConfigs{
				MerchantType:              c.MerchantTypeMerchant,
				MerchantTransactionConfig: merchantId,
			},
		},
		{
			name:           "SUCCESS:Sub-Merchant",
			setMerchantCtx: true,
			setupMock: func() {
				merchantRepo.On("FindMerchantByID", mock.Anything, merchantId).Once().Return(subMerchantData, nil)

				ctx = context.WithValue(ctx, c.CtxMerchantData, subMerchantData)
				ctx = context.WithValue(ctx, c.CtxParentMerchantId, parentMerchantId)
			},
			wantResult: &merchant.MerchantIdForConfigs{
				MerchantType:              c.MerchantTypeSubMerchant,
				MerchantTransactionConfig: parentMerchantId,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			c, result, err := service.GetMerchantIdForConfigs(context.Background(), merchantId, test.setMerchantCtx)
			assert.NotNil(t, c)
			assert.Equal(t, ctx, c)
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestFDSConfig(t *testing.T) {
	logger := pdkLoggerMock.NewILogger(t)
	repo := repoMocks.NewIMerchantRepository(t)

	service := New(repo, logger, nil, nil, nil, nil)

	config := merchant.FDSConfigRequest{
		FDSConfig: merchant.FDSConfig{
			ProofOfPayment: &merchant.FDSFeatureProofOfPayment{
				Velocity: merchant.FDSRuleVelocityConfig{
					Enabled: true,
					Window: merchant.FDSWindowConfig{
						Interval: 1,
						Unit:     "HOUR",
					},
					Threshold: merchant.FDSThresholdConfig{
						Count: 5,
					},
					Action: "BLOCK", // NOSONAR
				},
			},
		},
	}
	merchantID := "4f7288ce-f599-4bc9-9dcc-c9bc2bd5a55c"

	tests := []struct {
		name       string
		setupMock  func()
		wantError  error
		wantResult *merchant.FDSConfigResponse
	}{
		{
			name: "ERROR:Find merchant by ID",
			setupMock: func() {
				repo.On("FindMerchantByID", mock.Anything, merchantID).Once().Return(nil, assert.AnError)
				logger.On("Error", mock.Anything, "Failed to retrieve merchant details by ID", mock.Anything).Once().Return()
			},
			wantError: pkgErr.New(response.HttpErrDatabase, assert.AnError),
		},
		{
			name: "ERROR:Merchant not found",
			setupMock: func() {
				repo.On("FindMerchantByID", mock.Anything, merchantID).Once().Return(nil, nil)
			},
			wantError: pkgErr.New(response.HttpErrRequest, c.ErrMerchantNotFound),
		},
		{
			name: "ERROR:FDS config is not for Non-KYC sub-merchants",
			setupMock: func() {
				repo.On("FindMerchantByID", mock.Anything, merchantID).Once().Return(&merchant.Merchant{
					UUID:      merchantID,
					ParentID:  sql.NullString{Valid: true, String: "dd1cbc4a-5ed3-477c-bf0d-edd5e802fd9e"},
					KYCStatus: sql.NullString{Valid: true, String: c.KYCStatusNotRequired},
				}, nil)
			},
			wantError: pkgErr.New(response.HttpErrUnprocessableContent, fmt.Errorf("%s", "FDS configuration can only be used by merchants or KYC sub-merchants")),
		},
		{
			name: "ERROR:Update FDS config",
			setupMock: func() {
				repo.On("FindMerchantByID", mock.Anything, merchantID).Return(&merchant.Merchant{
					UUID:   merchantID,
					Status: c.StatusActive,
				}, nil)
				repo.On("UpdateFDSConfig", mock.Anything, merchantID, mock.Anything).Once().Return(assert.AnError)
				logger.On("Error", mock.Anything, "Failed to update FDS configuration", mock.Anything).Once().Return()
			},
			wantError: pkgErr.New(response.HttpErrDatabase, c.ErrInternalServerForUser),
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				repo.On("UpdateFDSConfig", mock.Anything, merchantID, mock.Anything).Once().Return(nil)
			},
			wantError:  nil,
			wantResult: &merchant.FDSConfigResponse{FDSConfig: config.FDSConfig},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			test.setupMock()

			result, err := service.FDSConfig(t.Context(), merchantID, config)
			assert.Equal(t, test.wantError, err)
			assert.Equal(t, test.wantResult, result)

			repo.AssertExpectations(t)
			logger.AssertExpectations(t)
		})
	}
}

func TestGetFDSConfig(t *testing.T) {
	logger := pdkLoggerMock.NewILogger(t)
	repo := repoMocks.NewIMerchantRepository(t)

	service := New(repo, logger, nil, nil, nil, nil)

	merchantID := "4d154e93-1285-4f3e-aeaa-eca8cb3541d5"
	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult *merchant.GetFDSConfigResponse
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				repo.On("GetFDSConfig", mock.Anything, merchantID).Once().Return(nil, assert.AnError)
				logger.On("Error", mock.Anything, "Failed to fetch merchant configuration", mock.Anything).Once().Return()
			},
			wantErr: pkgErr.New(response.HttpErrDatabase, c.ErrInternalServerForUser),
		},
		{
			name: "ERROR:Merchant not found", // NOSONAR
			setupMock: func() {
				repo.On("GetFDSConfig", mock.Anything, merchantID).Once().Return(nil, nil)
			},
			wantErr: pkgErr.New(response.HttpErrNotFound, c.ErrMerchantNotFound),
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				repo.On("GetFDSConfig", mock.Anything, merchantID).Once().Return(&merchant.GetFDSConfigResponse{}, nil)
			},
			wantResult: &merchant.GetFDSConfigResponse{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := service.GetFDSConfig(t.Context(), merchantID)
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)

			repo.AssertExpectations(t)
			logger.AssertExpectations(t)
		})
	}
}

func TestPaymentInvestigationConfig(t *testing.T) {
	logger := pdkLoggerMock.NewILogger(t)
	repo := repoMocks.NewIMerchantRepository(t)

	service := New(repo, logger, nil, nil, nil, nil)

	merchantID := "63d2d174-7473-4fb6-8d10-dbfa1b11e8e8"
	config := merchant.PaymentInvestigationConfigRequest{
		Enabled:             true,    // NOSONAR
		PivotPercentageLoss: 50,      // NOSONAR
		PivotMaxLoss:        500_000, // NOSONAR
	}

	tests := []struct {
		name       string
		setupMock  func()
		wantError  error
		wantResult *merchant.PaymentInvestigationConfigResponse
	}{
		{
			name: "ERROR:Find merchant by ID",
			setupMock: func() {
				repo.On("FindMerchantByID", mock.Anything, merchantID).Once().Return(nil, assert.AnError)
				logger.On("Error", mock.Anything, "Failed to retrieve merchant details by ID", mock.Anything).Once().Return()
			},
			wantError: pkgErr.New(response.HttpErrDatabase, assert.AnError),
		},
		{
			name: "ERROR:Merchant not found",
			setupMock: func() {
				repo.On("FindMerchantByID", mock.Anything, merchantID).Once().Return(nil, nil)
			},
			wantError: pkgErr.New(response.HttpErrRequest, c.ErrMerchantNotFound),
		},
		{
			name: "ERROR:Payment investigation config is not for Non-KYC sub-merchants",
			setupMock: func() {
				repo.On("FindMerchantByID", mock.Anything, merchantID).Once().Return(&merchant.Merchant{
					UUID:      "84fc7135-2e66-4e0b-b558-727505e04177",
					ParentID:  sql.NullString{Valid: true, String: merchantID},
					KYCStatus: sql.NullString{Valid: true, String: c.KYCStatusNotRequired},
				}, nil)
			},
			wantError: pkgErr.New(response.HttpErrUnprocessableContent, fmt.Errorf("%s", "Payment investigation config can only be used by merchants or KYC sub-merchants")),
		},
		{
			name: "ERROR:Update payment investigation config",
			setupMock: func() {
				repo.On("FindMerchantByID", mock.Anything, merchantID).Return(&merchant.Merchant{
					UUID:   merchantID,
					Name:   "TEST", // NOSONAR
					Status: c.StatusActive,
				}, nil)
				repo.On("UpdatePaymentInvestigationConfig", mock.Anything, merchantID, config).Once().Return(assert.AnError)
				logger.On("Error", mock.Anything, "Failed to update payment investigation config", mock.Anything).Once().Return()
			},
			wantError: pkgErr.New(response.HttpErrDatabase, c.ErrInternalServerForUser),
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				repo.On("UpdatePaymentInvestigationConfig", mock.Anything, merchantID, config).Once().Return(nil)
			},
			wantError: nil,
			wantResult: &merchant.PaymentInvestigationConfigResponse{
				Enabled:             true, // NOSONAR
				MerchantID:          merchantID,
				MerchantName:        "TEST",  // NOSONAR
				PivotPercentageLoss: 50,      // NOSONAR
				PivotMaxLoss:        500_000, // NOSONAR
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			test.setupMock()

			result, err := service.PaymentInvestigationConfig(t.Context(), merchantID, config)
			assert.Equal(t, test.wantError, err)
			assert.Equal(t, test.wantResult, result)

			repo.AssertExpectations(t)
			logger.AssertExpectations(t)
		})
	}
}
