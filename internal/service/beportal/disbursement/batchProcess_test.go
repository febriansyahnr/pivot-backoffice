package disbursementService

import (
	"context"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	disbursementDashboardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursementDashboard"
	rabbitMqExtMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	redisMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestBatchProcessDisbursement(t *testing.T) {
	type mocker struct {
		disbursementRepo *repositoryMocks.IDisbursementRepository
		internal         *serviceMocks.IDisbursementInternalService
		rmq              *rabbitMqExtMocks.RabbitMQExt
		redis            *redisMock.IRedisExt
	}

	disbursementID := uuid.NewString()
	validRequest := &disbursementModel.BatchProcessDisbursementRequest{BulkID: uuid.NewString(), DisbursementIDs: []string{disbursementID}}
	feeDecimal := decimal.NewFromFloat(1000)

	testCases := []struct {
		name      string
		wantErr   bool
		mockSetup func(m *mocker)
		input     *disbursementModel.BatchProcessDisbursementRequest
	}{
		{
			name:    "SUCCESS",
			wantErr: false,
			mockSetup: func(m *mocker) {
				m.disbursementRepo.On(
					"FindByID", mock.Anything, constant.StringMockType(),
				).Times(2).Return(&disbursementModel.DisbursementWithTransaction{Disbursement: disbursementModel.Disbursement{MerchantID: uuid.NewString(), Fee: &feeDecimal}}, nil)

				m.internal.On(
					"CreateBankTransfer", constant.ValueCtxMockType(), mock.Anything,
				).Once().Return(nil, nil)

				m.disbursementRepo.On(
					"CountStatusInProgressByBulkID", mock.Anything, constant.StringMockType(),
				).Once().Return(0)

				m.disbursementRepo.On(
					"UpdateBulkDisbursementStatusByID", mock.Anything, constant.StringMockType(), constant.StringMockType(),
				).Once().Return(nil)
			},
			input: validRequest,
		},
		{
			name:    "SUCCESS:With Send Callback",
			wantErr: false,
			mockSetup: func(m *mocker) {
				m.disbursementRepo.On(
					"FindByID", mock.Anything, constant.StringMockType(),
				).Times(2).Return(&disbursementModel.DisbursementWithTransaction{Disbursement: disbursementModel.Disbursement{
					MerchantID:  uuid.NewString(),
					Fee:         &feeDecimal,
					CreatedFrom: util.ValueToPtr(constant.DisbursementCreatedFromOpenApi),
				}}, nil)
				m.internal.On(
					"CreateBankTransfer", constant.ValueCtxMockType(), mock.Anything,
				).Once().Return(nil, nil)

				m.disbursementRepo.On(
					"CountStatusInProgressByBulkID", mock.Anything, constant.StringMockType(),
				).Return(0)

				m.disbursementRepo.On(
					"CountByBulkID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Once().Return(2)

				m.disbursementRepo.On(
					"UpdateBulkDisbursementStatusByID", mock.Anything, constant.StringMockType(), constant.StringMockType(),
				).Once().Return(nil)

				m.disbursementRepo.On(
					"SummaryPendingByBulkID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Once().Return(disbursementDashboardModel.SummaryTransactionDTO{})

				m.disbursementRepo.On(
					"SummarySuccessByBulkID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Once().Return(disbursementDashboardModel.SummaryTransactionDTO{})

				m.disbursementRepo.On(
					"SummaryFailedByBulkID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Once().Return(disbursementDashboardModel.SummaryTransactionDTO{})

				m.disbursementRepo.On(
					"SummaryCancelledByBulkID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Once().Return(disbursementDashboardModel.SummaryTransactionDTO{})

				m.disbursementRepo.On(
					"GetMerchantIDsForPayoutCallback", mock.Anything, mock.Anything,
				).Once().Return(nil, nil)
				m.rmq.On(
					"PublishMerchantCallback", constant.ValueCtxMockType(), mock.Anything,
				).Once().Return(nil)

				// Callback event lock SetNX (value="1")
				m.redis.On(
					"SetNX", constant.ValueCtxMockType(), constant.StringMockType(), "1", mock.AnythingOfType("time.Duration"),
				).Once().Return(redis.NewBoolResult(true, nil))
			},
			input: validRequest,
		},
		{
			name:    "SUCCESS:But got error in Process service",
			wantErr: false,
			mockSetup: func(m *mocker) {
				m.disbursementRepo.On(
					"FindByID", mock.Anything, constant.StringMockType(),
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)

				m.disbursementRepo.On(
					"CountStatusInProgressByBulkID", mock.Anything, constant.StringMockType(),
				).Once().Return(1)
			},
			input: validRequest,
		},
		{
			name:    "ERROR:Update Status Parent And Send Callback #1",
			wantErr: true,
			mockSetup: func(m *mocker) {
				m.disbursementRepo.On(
					"FindByID", mock.Anything, constant.StringMockType(),
				).Once().Return(&disbursementModel.DisbursementWithTransaction{Disbursement: disbursementModel.Disbursement{MerchantID: uuid.NewString(), Fee: &feeDecimal}}, nil)
				m.internal.On(
					"CreateBankTransfer", constant.ValueCtxMockType(), mock.Anything,
				).Once().Return(nil, nil)

				m.disbursementRepo.On(
					"CountStatusInProgressByBulkID", mock.Anything, constant.StringMockType(),
				).Once().Return(0)

				m.disbursementRepo.On(
					"UpdateBulkDisbursementStatusByID", mock.Anything, constant.StringMockType(), constant.StringMockType(),
				).Return(constant.ErrSomeErrorForUnitTest)
			},
			input: validRequest,
		},
		{
			name:    "ERROR:Update Status Parent And Send Callback #2",
			wantErr: true,
			mockSetup: func(m *mocker) {
				m.disbursementRepo.On(
					"FindByID", mock.Anything, constant.StringMockType(),
				).Times(1).Return(&disbursementModel.DisbursementWithTransaction{Disbursement: disbursementModel.Disbursement{MerchantID: uuid.NewString(), Fee: &feeDecimal}}, nil)
				m.internal.On(
					"CreateBankTransfer", constant.ValueCtxMockType(), mock.Anything,
				).Once().Return(nil, nil)

				m.disbursementRepo.On(
					"CountStatusInProgressByBulkID", mock.Anything, constant.StringMockType(),
				).Once().Return(0)

				m.disbursementRepo.On(
					"UpdateBulkDisbursementStatusByID", mock.Anything, constant.StringMockType(), constant.StringMockType(),
				).Return(nil)
				m.disbursementRepo.On(
					"FindByID", mock.Anything, constant.StringMockType(),
				).Times(1).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			input: validRequest,
		},
		{
			name:    "ERROR:Update Status Parent And Send Callback #3",
			wantErr: true,
			mockSetup: func(m *mocker) {
				m.disbursementRepo.On(
					"FindByID", mock.Anything, constant.StringMockType(),
				).Times(1).Return(&disbursementModel.DisbursementWithTransaction{Disbursement: disbursementModel.Disbursement{MerchantID: uuid.NewString(), Fee: &feeDecimal}}, nil)
				m.internal.On(
					"CreateBankTransfer", constant.ValueCtxMockType(), mock.Anything,
				).Once().Return(nil, nil)

				m.disbursementRepo.On(
					"CountStatusInProgressByBulkID", mock.Anything, constant.StringMockType(),
				).Once().Return(0)

				m.disbursementRepo.On(
					"UpdateBulkDisbursementStatusByID", mock.Anything, constant.StringMockType(), constant.StringMockType(),
				).Return(nil)
				m.disbursementRepo.On(
					"FindByID", mock.Anything, constant.StringMockType(),
				).Times(1).Return(nil, nil)
			},
			input: validRequest,
		},
		{
			name:    "ERROR:Send Callback Lock Acquisition Failed",
			wantErr: true,
			mockSetup: func(m *mocker) {
				m.disbursementRepo.On(
					"FindByID", mock.Anything, constant.StringMockType(),
				).Times(2).Return(&disbursementModel.DisbursementWithTransaction{Disbursement: disbursementModel.Disbursement{
					MerchantID:  uuid.NewString(),
					Fee:         &feeDecimal,
					CreatedFrom: util.ValueToPtr(constant.DisbursementCreatedFromOpenApi),
				}}, nil)
				m.internal.On(
					"CreateBankTransfer", constant.ValueCtxMockType(), mock.Anything,
				).Once().Return(nil, nil)

				m.disbursementRepo.On(
					"CountStatusInProgressByBulkID", mock.Anything, constant.StringMockType(),
				).Return(0)

				m.disbursementRepo.On(
					"UpdateBulkDisbursementStatusByID", mock.Anything, constant.StringMockType(), constant.StringMockType(),
				).Once().Return(nil)

				m.disbursementRepo.On(
					"CountByBulkID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Once().Return(2)

				m.disbursementRepo.On(
					"SummaryPendingByBulkID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Once().Return(disbursementDashboardModel.SummaryTransactionDTO{})

				m.disbursementRepo.On(
					"SummarySuccessByBulkID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Once().Return(disbursementDashboardModel.SummaryTransactionDTO{})

				m.disbursementRepo.On(
					"SummaryFailedByBulkID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Once().Return(disbursementDashboardModel.SummaryTransactionDTO{})

				m.disbursementRepo.On(
					"SummaryCancelledByBulkID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Once().Return(disbursementDashboardModel.SummaryTransactionDTO{})

				m.redis.On(
					"SetNX", constant.ValueCtxMockType(), constant.StringMockType(), "1", mock.AnythingOfType("time.Duration"),
				).Once().Return(redis.NewBoolResult(false, constant.ErrSomeErrorForUnitTest))
			},
			input: validRequest,
		},
		{
			name:    "SUCCESS:Send Callback Lock Not Acquired",
			wantErr: false,
			mockSetup: func(m *mocker) {
				m.disbursementRepo.On(
					"FindByID", mock.Anything, constant.StringMockType(),
				).Times(2).Return(&disbursementModel.DisbursementWithTransaction{Disbursement: disbursementModel.Disbursement{
					MerchantID:  uuid.NewString(),
					Fee:         &feeDecimal,
					CreatedFrom: util.ValueToPtr(constant.DisbursementCreatedFromOpenApi),
				}}, nil)
				m.internal.On(
					"CreateBankTransfer", constant.ValueCtxMockType(), mock.Anything,
				).Once().Return(nil, nil)

				m.disbursementRepo.On(
					"CountStatusInProgressByBulkID", mock.Anything, constant.StringMockType(),
				).Return(0)

				m.disbursementRepo.On(
					"UpdateBulkDisbursementStatusByID", mock.Anything, constant.StringMockType(), constant.StringMockType(),
				).Once().Return(nil)

				m.disbursementRepo.On(
					"CountByBulkID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Once().Return(2)

				m.disbursementRepo.On(
					"SummaryPendingByBulkID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Once().Return(disbursementDashboardModel.SummaryTransactionDTO{})

				m.disbursementRepo.On(
					"SummarySuccessByBulkID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Once().Return(disbursementDashboardModel.SummaryTransactionDTO{})

				m.disbursementRepo.On(
					"SummaryFailedByBulkID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Once().Return(disbursementDashboardModel.SummaryTransactionDTO{})

				m.disbursementRepo.On(
					"SummaryCancelledByBulkID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Once().Return(disbursementDashboardModel.SummaryTransactionDTO{})

				m.redis.On(
					"SetNX", constant.ValueCtxMockType(), constant.StringMockType(), "1", mock.AnythingOfType("time.Duration"),
				).Once().Return(redis.NewBoolResult(false, nil))
			},
			input: validRequest,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := &mocker{
				disbursementRepo: repositoryMocks.NewIDisbursementRepository(t),
				internal:         serviceMocks.NewIDisbursementInternalService(t),
				rmq:              rabbitMqExtMocks.NewRabbitMQExt(t),
				redis:            redisMock.NewIRedisExt(t),
			}
			// Register fallback Maybe mocks FIRST so specific mocks in mockSetup take priority (testify uses LIFO matching)
			m.redis.On(
				"Del", constant.ValueCtxMockType(), constant.StringMockType(),
			).Return(&redis.IntCmd{}).Maybe()

			// Queue lock SetNX uses bool value (true); callback lock uses string ("1").
			// This mock only matches the queue lock call (value=true).
			m.redis.On(
				"SetNX", constant.ValueCtxMockType(), constant.StringMockType(), true, mock.AnythingOfType("time.Duration"),
			).Return(redis.NewBoolResult(true, nil)).Maybe()

			tc.mockSetup(m)

			conf := config.Config{
				Environment: constant.EnvironmentStaging,
			}
			svc := New(
				&conf, pdkLoggerMock, nil, m.disbursementRepo, nil, nil,
				WithRabbitMQClient(m.rmq),
				WithDisbursementInternalService(m.internal),
				WithRedisClient(m.redis),
			)

			if err := svc.BatchProcessDisbursement(context.Background(), tc.input); tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
