package feeService_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/fee"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/go-redis/redismock/v9"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var tz, _ = time.LoadLocation(c.TimeLoc)

func TestPlatformActivitiesFee(t *testing.T) {

	rdb, clientMock := redismock.NewClientMock()

	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
	feeRepo := repoMocks.NewIFeeRepository(t)
	merchantRepo := repoMocks.NewIMerchantRepository(t)
	orchestratorSvc := serviceMocks.NewIOrchestratorService(t)
	accountTransactionRepo := repoMocks.NewIAccountTransactionRepository(t)

	service := New(
		logger, feeRepo, merchantRepo,
		WithOrchestratorService(orchestratorSvc),
		WithAccountTransactionRepository(accountTransactionRepo),
		WithPaymentMethodService(nil),
		WithRedisClient(redisExt.WrapRedisClient(rdb, nil)),
	)

	merchantId := uuid.NewString()
	subMerchants := []string{uuid.NewString(), uuid.NewString()}
	rawSubMerchants, _ := json.Marshal(subMerchants)

	cacheKey := fmt.Sprintf(
		c.NonPaymentFeeConfigsFmt, merchantId, strings.ToLower(c.ReferencePlatformActivity),
	)

	merchants := []merchant.MerchantWithSubMerchantList{
		{
			ID:              merchantId,
			CreatedAt:       time.Date(2024, 8, 1, 7, 0, 0, 0, time.UTC),
			RawSubMerchants: rawSubMerchants,
			SubMerchants:    subMerchants,
		},
	}

	tests := []struct {
		name      string
		date      time.Time
		setupMock func()
		wantErr   error
	}{
		{
			name: "ERROR:Get list of merchants who have sub merchant",
			setupMock: func() {
				merchantRepo.On(
					"GetListOfMerchantsWhoHaveSubMerchant", c.ValueCtxMockType(),
				).Once().Return(nil, fmt.Errorf("get list merchants: %v", c.ErrSomeErrorForUnitTest))
			},
			wantErr: fmt.Errorf("get list merchants: %v", c.ErrSomeErrorForUnitTest), // NOSONAR
		},
		{
			name: "SUCCESS:List of merchants is not found",
			setupMock: func() {
				merchantRepo.On(
					"GetListOfMerchantsWhoHaveSubMerchant", c.ValueCtxMockType(),
				).Once().Return(nil, nil)

			},
		},
		{
			name: "ERROR:Get merchant fee by merchant id and type",
			setupMock: func() {
				merchantRepo.On("GetListOfMerchantsWhoHaveSubMerchant", c.ValueCtxMockType()).Return(merchants, nil)

				merchantRepo.On(
					"GetMerchantFeeByMerchantIDAndType", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, c.ErrSomeErrorForUnitTest), // NOSONAR
		},
		{
			name: "SUCCESS:There is no calculation schedule",
			date: time.Date(2024, 2, 27, 0, 10, 0, 0, tz),
			setupMock: func() {
				merchantRepo.On(
					"GetMerchantFeeByMerchantIDAndType", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(nil, nil)
			},
		},
		{
			name: "ERROR:Begin transaction",
			date: time.Date(2024, 2, 29, 0, 10, 0, 0, tz),
			setupMock: func() {
				merchantRepo.On(
					"GetMerchantFeeByMerchantIDAndType", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(nil, nil)
				feeRepo.On("BeginTransaction", c.ValueCtxMockType()).Once().Return(nil, fmt.Errorf("Bagin Trx: %v", c.ErrSomeErrorForUnitTest))
			},
			wantErr: fmt.Errorf("Bagin Trx: %v", c.ErrSomeErrorForUnitTest), // NOSONAR
		},
		{
			name: "ERROR:Get platform transaction activities",
			date: time.Date(2024, 2, 29, 0, 10, 0, 0, tz),
			setupMock: func() {
				feeRepo.On(
					"BeginTransaction", c.ValueCtxMockType(),
				).Return(context.WithValue(context.Background(), mySqlExt.CtxSqlTx, &sql.Tx{}), nil)

				merchantRepo.On(
					"GetMerchantFeeByMerchantIDAndType", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(nil, nil)
				feeRepo.On(
					"RollbackTransaction", c.ValueCtxMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
				accountTransactionRepo.On(
					"GetPlatformTransactionActivities", c.ValueCtxMockType(), mock.AnythingOfType("[]string"), c.TimeMockType(), c.TimeMockType(),
				).Once().Return(nil, fmt.Errorf("Get platform activities: %v", c.ErrSomeErrorForUnitTest))
			},
			wantErr: fmt.Errorf("Get platform activities: %v", c.ErrSomeErrorForUnitTest), // NOSONAR
		},
		{
			name: "ERROR:Create default merchant fee",
			date: time.Date(2024, 2, 29, 0, 10, 0, 0, tz),
			setupMock: func() {
				feeRepo.On("RollbackTransaction", c.ValueCtxMockType()).Return(nil)

				merchantRepo.On(
					"GetMerchantFeeByMerchantIDAndType", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(nil, nil)
				accountTransactionRepo.On(
					"GetPlatformTransactionActivities", c.ValueCtxMockType(), mock.AnythingOfType("[]string"), c.TimeMockType(), c.TimeMockType(),
				).Once().Return(nil, nil)
				merchantRepo.On(
					"CreateMerchantFee", c.ValueCtxMockType(), mock.AnythingOfType("*merchant.MerchantFee"),
				).Once().Return(fmt.Errorf("Create merchant fee: %v", c.ErrSomeErrorForUnitTest))
			},
			wantErr: fmt.Errorf("Create merchant fee: %v", c.ErrSomeErrorForUnitTest), // NOSONAR
		},
		{
			name: "ERROR:Update merchant fee last deduction date",
			date: time.Date(2024, 2, 29, 0, 10, 0, 0, tz),
			setupMock: func() {
				merchantRepo.On(
					"GetMerchantFeeByMerchantIDAndType", c.ValueCtxMockType(), merchantId, c.ReferencePlatformActivity,
				).Return(&merchant.MerchantFee{
					Reference:     c.ReferencePlatformActivity,
					AmountType:    c.MerchantFeeAmountType,
					Amount:        15_000,
					DeductionType: c.MerchantFeeDeductionTypeAutomated,
					DeductionDay:  util.ValueToPtr(int16(31)),
					TaxType:       c.MerchantTaxTypeExclusive,
					TaxPercentage: 10,
				}, nil)

				accountTransactionRepo.On(
					"GetPlatformTransactionActivities", c.ValueCtxMockType(), mock.AnythingOfType("[]string"), c.TimeMockType(), c.TimeMockType(),
				).Once().Return(nil, nil)
				merchantRepo.On(
					"UpdateMerchantFeeLastDeductionDate", c.ValueCtxMockType(), merchantId, c.ReferencePlatformActivity, c.TimeMockType(),
				).Once().Return(fmt.Errorf("Update deduction date: %v", c.ErrSomeErrorForUnitTest))
			},
			wantErr: fmt.Errorf("Update deduction date: %v", c.ErrSomeErrorForUnitTest), // NOSONAR
		},
		{
			name: "SUCCESS:Transaction activities is not found",
			date: time.Date(2024, 2, 29, 0, 10, 0, 0, tz),
			setupMock: func() {
				accountTransactionRepo.On(
					"GetPlatformTransactionActivities", c.ValueCtxMockType(), mock.AnythingOfType("[]string"), c.TimeMockType(), c.TimeMockType(),
				).Once().Return(nil, nil)

				merchantRepo.On(
					"UpdateMerchantFeeLastDeductionDate", c.ValueCtxMockType(), merchantId, c.ReferencePlatformActivity, c.TimeMockType(),
				).Once().Return(nil)
				feeRepo.On(
					"CommitTransaction", c.ValueCtxMockType(),
				).Once().Return(nil)
			},
		},
		{
			name: "ERROR:Post account transaction",
			date: time.Date(2024, 2, 29, 0, 10, 0, 0, tz),
			setupMock: func() {
				accountTransactionRepo.On(
					"GetPlatformTransactionActivities", c.ValueCtxMockType(), mock.AnythingOfType("[]string"), c.TimeMockType(), c.TimeMockType(),
				).Return([]model.TransactionActivity{{}, {}}, nil)
				orchestratorSvc.On(
					"PostAccountTransaction", c.ValueCtxMockType(), mock.MatchedBy(func(req *model.CreateAccountTransactionRequest) bool {
						return req.Status == c.StatusSuccess
					}),
				).Once().Return(fmt.Errorf("Post account transaction: %v", c.ErrSomeErrorForUnitTest))
			},
			wantErr: fmt.Errorf("Post account transaction: %v", c.ErrSomeErrorForUnitTest), // NOSONAR
		},
		{
			name: "ERROR:Update last deduct balance",
			date: time.Date(2024, 2, 29, 0, 10, 0, 0, tz),
			setupMock: func() {
				orchestratorSvc.On(
					"PostAccountTransaction", c.ValueCtxMockType(), mock.Anything,
				).Return(nil)

				merchantRepo.On(
					"UpdateMerchantFeeLastDeductionDate", c.ValueCtxMockType(), merchantId, c.ReferencePlatformActivity, c.TimeMockType(),
				).Once().Return(fmt.Errorf("Update last deduct balance: %v", c.ErrSomeErrorForUnitTest))
			},
			wantErr: fmt.Errorf("Update last deduct balance: %v", c.ErrSomeErrorForUnitTest), // NOSONAR
		},
		{
			name: "ERROR:Commit transaction",
			date: time.Date(2024, 2, 29, 0, 10, 0, 0, tz),
			setupMock: func() {
				merchantRepo.On(
					"UpdateMerchantFeeLastDeductionDate", c.ValueCtxMockType(), merchantId, c.ReferencePlatformActivity, c.TimeMockType(),
				).Return(nil)

				feeRepo.On(
					"CommitTransaction", c.ValueCtxMockType(),
				).Once().Return(fmt.Errorf("Commit transaction: %v", c.ErrSomeErrorForUnitTest))
			},
			wantErr: fmt.Errorf("Commit transaction: %v", c.ErrSomeErrorForUnitTest),
		},
		{
			name: "SUCCESS",
			date: time.Date(2024, 2, 29, 0, 10, 0, 0, tz),
			setupMock: func() {

				feeRepo.On(
					"CommitTransaction", c.ValueCtxMockType(),
				).Return(nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			test.setupMock()
			clientMock.ExpectGet(cacheKey).SetErr(redisExt.ErrNil)

			assert.Equal(t, test.wantErr, service.PlatformActivitiesFee(context.Background(), test.date))
		})
	}
}
