package disbursementService_test

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	disbursementService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/disbursement"
	rabbitMqExtMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"

	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetCutOffTimeStatus(t *testing.T) {
	config := &config.Config{
		DisbursementConfig: config.DisbursementConfig{
			CutOffTimeWindow: config.DisbursementCutOffTimeWindow{
				SameDay:                       true,
				GMT:                           7,
				BannerStatus:                  "Any transaction bla bla",
				TimeLagForSendingReportSecond: 15,
			},
		},
	}

	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	defer buf.Reset()

	defaultMerchantId := "7c246a7b-01e3-441d-b810-5466229dfadc"
	log := logger.NewSlogger(logger.Config{}, logger.WithSlogOutput(buf))

	timeToTestStr := "2025-01-16T14:45:00+07:00"
	timeToTest := time.Date(2025, 1, 16, 7, 45, 0, 0, time.UTC)

	statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
	statusHistoriesRepo.On("Insert", mock.Anything, mock.Anything).Return(nil).Maybe()

	service := disbursementService.New(config, log, nil, nil, nil, nil,
		disbursementService.WithStatusHistoriesRepository(statusHistoriesRepo),
	)

	tests := []struct {
		name       string
		merchantId string
		setupFlow  func()
		wantErr    error
		wantResult *disbursementModel.CutOffTimeStatusResponse
	}{
		{
			name: "Payout cut-off time feature is deactive",
			wantResult: &disbursementModel.CutOffTimeStatusResponse{
				Status: constant.DisbursementCutOffTimeStatusDeactive,
			},
		},
		{
			name: "Error parse start time window",
			setupFlow: func() {
				config.DisbursementConfig.CutOffTimeWindow.Enabled = true
				config.DisbursementConfig.CutOffTimeWindow.StartTime = "HH:mm"
			},
			wantErr: errors.New(`parse start time window: parsing time "2025-01-16 HH:mm:00" as "2006-01-02 15:04:05": cannot parse "HH:mm:00" as "15"`),
		},
		{
			name: "Error parse end time window",
			setupFlow: func() {
				config.DisbursementConfig.CutOffTimeWindow.StartTime = "12:00"
				config.DisbursementConfig.CutOffTimeWindow.EndTime = "HH:mm"
			},
			wantErr: errors.New(`parse end time window: parsing time "2025-01-16 HH:mm:59" as "2006-01-02 15:04:05": cannot parse "HH:mm:59" as "15"`),
		},
		{
			name: "Payout cut-off time is off schedule",
			setupFlow: func() {
				config.DisbursementConfig.CutOffTimeWindow.StartTime = "00:00"
				config.DisbursementConfig.CutOffTimeWindow.EndTime = "04:59"
			},
			wantResult: &disbursementModel.CutOffTimeStatusResponse{
				Status: constant.DisbursementCutOffTimeStatusOffSchedule, Time: timeToTestStr,
			},
		},
		{
			name:       "Whitelisted merchant id",
			merchantId: "3fc96de8-f65e-4b16-90a1-e2a00d1bae29",
			setupFlow: func() {
				config.DisbursementConfig.CutOffTimeWindow.StartTime = "12:00"
				config.DisbursementConfig.CutOffTimeWindow.EndTime = "14:59"
			},
			wantResult: &disbursementModel.CutOffTimeStatusResponse{
				Status: constant.DisbursementCutOffTimeStatusWhitelisted, Time: timeToTestStr,
			},
		},
		{
			name: "Payout cut_off time ongoing",
			wantResult: &disbursementModel.CutOffTimeStatusResponse{
				Status:      constant.DisbursementCutOffTimeStatusOngoing,
				Time:        timeToTestStr,
				Banner:      "Any transaction bla bla",
				ProcessedAt: time.Date(2025, 1, 16, 8, 0, 14, 0, time.UTC),
			},
		},
		{
			name: "Payout cut_off time config cross day ongoing",
			setupFlow: func() {
				timeToTest = time.Date(2025, 1, 16, 7, 59, 40, 0, time.UTC) // Modify input param

				config.DisbursementConfig.CutOffTimeWindow.SameDay = false
				config.DisbursementConfig.CutOffTimeWindow.StartTime = "23:00"
				config.DisbursementConfig.CutOffTimeWindow.EndTime = "14:59"
			},
			wantResult: &disbursementModel.CutOffTimeStatusResponse{
				Status:      constant.DisbursementCutOffTimeStatusOngoing,
				Time:        "2025-01-16T14:59:40+07:00",
				Banner:      "Any transaction bla bla",
				ProcessedAt: time.Date(2025, 1, 16, 8, 0, 14, 0, time.UTC),
			},
		},
		{
			name:       "Payout cut_off time config cross day whitelisted",
			merchantId: "3fc96de8-f65e-4b16-90a1-e2a00d1bae29",
			wantResult: &disbursementModel.CutOffTimeStatusResponse{
				Status: constant.DisbursementCutOffTimeStatusWhitelisted, Time: "2025-01-16T14:59:40+07:00",
			},
		},
		{
			name: "Payout cut_off time config cross day off schedule",
			setupFlow: func() {
				config.DisbursementConfig.CutOffTimeWindow.SameDay = false
				config.DisbursementConfig.CutOffTimeWindow.StartTime = "23:00"
				config.DisbursementConfig.CutOffTimeWindow.EndTime = "04:59"
			},
			wantResult: &disbursementModel.CutOffTimeStatusResponse{
				Status: constant.DisbursementCutOffTimeStatusOffSchedule, Time: "2025-01-16T14:59:40+07:00",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf.Reset()

			if test.setupFlow != nil {
				test.setupFlow()
			}
			if test.merchantId == "" {
				test.merchantId = defaultMerchantId
			}

			result, err := service.GetCutOffTimeStatus(context.Background(), timeToTest, test.merchantId, nil)
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestReportAfterPayoutCutOffTimeByPartnerWindow(t *testing.T) {
	type mocker struct {
		disbursementRepo       *repositoryMocks.IDisbursementRepository
		accountTransactionRepo *repositoryMocks.IAccountTransactionRepository
		rabbitMqExt            *rabbitMqExtMocks.RabbitMQExt
	}

	config := &config.Config{}
	config.SlackConfig.PayoutCutOffTimeWebHookURL = "https://webhook.test/payout"

	log := logger.NewSlogger(logger.Config{})

	testCases := []struct {
		name      string
		request   *disbursementModel.PartnerWindowCutOffReportRequest
		wantTotal int64
		setupMock func(m mocker)
	}{
		{
			name: "empty request returns empty report",
			request: &disbursementModel.PartnerWindowCutOffReportRequest{
				PartnerCode: "",
			},
			wantTotal: 0,
			setupMock: func(m mocker) {},
		},
		{
			name: "success generate partner window report",
			request: &disbursementModel.PartnerWindowCutOffReportRequest{
				PartnerCode:   "014",
				PartnerName:   "BCA",
				WindowStartAt: time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC),
				WindowEndAt:   time.Date(2026, 2, 10, 2, 0, 0, 0, time.UTC),
				ExternalIDs:   []string{"trx-1"},
			},
			wantTotal: 1,
			setupMock: func(m mocker) {
				m.accountTransactionRepo.On("FindByID", mock.Anything, "trx-1").Return(&orchestratorModel.AccountTransactionWithUseCase{
					Type:        constant.TypeDisbursement,
					ReferenceID: "disb-1",
					ReasonType:  sql.NullString{String: constant.ReasonTypePayoutCutOffTime, Valid: true},
				}, nil).Once()

				bankName := "Bank BCA"
				m.disbursementRepo.On("FindByID", mock.Anything, "disb-1").Return(&disbursementModel.DisbursementWithTransaction{
					Disbursement: disbursementModel.Disbursement{
						MerchantID:          "merchant-1",
						BeneficiaryBankName: &bankName,
						Amount:              decimal.NewFromInt(100000),
					},
				}, nil).Once()

				m.disbursementRepo.On("GetAvgDurationOfBankTransferProcessInMs", mock.Anything, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return(1200.0, nil).Once()
				m.rabbitMqExt.On("Publish", mock.Anything, rabbitMqExt.SlackPostWebhookRoutingKey, mock.Anything, mock.AnythingOfType("[]uint8")).Return(nil).Once()
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			disbursementRepo := repositoryMocks.NewIDisbursementRepository(t)
			accountTransactionRepo := repositoryMocks.NewIAccountTransactionRepository(t)
			rabbitMqExt := rabbitMqExtMocks.NewRabbitMQExt(t)
			tc.setupMock(mocker{disbursementRepo: disbursementRepo, accountTransactionRepo: accountTransactionRepo, rabbitMqExt: rabbitMqExt})

			statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
			statusHistoriesRepo.On("Insert", mock.Anything, mock.Anything).Return(nil).Maybe()

			svc := disbursementService.New(config, log, nil, disbursementRepo, nil, nil,
				disbursementService.WithAccountTransactionRepository(accountTransactionRepo),
				disbursementService.WithRabbitMQClient(rabbitMqExt),
				disbursementService.WithStatusHistoriesRepository(statusHistoriesRepo),
			)

			reporter, ok := svc.(interface {
				ReportAfterPayoutCutOffTimeByPartnerWindow(ctx context.Context, req *disbursementModel.PartnerWindowCutOffReportRequest) (disbursementModel.AfterPayoutCutOffTimeSummary, error)
			})
			assert.True(t, ok)

			report, err := reporter.ReportAfterPayoutCutOffTimeByPartnerWindow(context.Background(), tc.request)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantTotal, report.Total)

			disbursementRepo.AssertExpectations(t)
			accountTransactionRepo.AssertExpectations(t)
			rabbitMqExt.AssertExpectations(t)
		})
	}
}
