package bankTransferConsumer_test

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	bankTransferProto "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/bankTransfer"
	redisMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/consumer/bankTransfer"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/redis/go-redis/v9"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

func TestUpdateTransferStatus(t *testing.T) {
	logger := logger.NewSlogger(logger.Config{})
	ledgerSvc := serviceMocks.NewIOrchestratorService(t)
	withdrawalSvc := serviceMocks.NewIWithdrawalService(t)
	disbursementSvc := serviceMocks.NewIDisbursementService(t)
	refundProcSvc := serviceMocks.NewIRefundProcessorService(t)
	cardFundedPayoutSvc := serviceMocks.NewICardFundedPayoutService(t)

	handler := New(logger, &Service{
		LedgerSvc:           ledgerSvc,
		DisbursementSvc:     disbursementSvc,
		RefundProcSvc:       refundProcSvc,
		WithdrawalSvc:       withdrawalSvc,
		CardFundedPayoutSvc: cardFundedPayoutSvc,
	})

	externalId := "9ed2968f-2cc4-4fa0-a23f-487fd3d1d765"
	payload := fmt.Appendf(nil, `{"externalId": "%s"}`, externalId)

	tests := []struct {
		name      string
		payload   []byte
		setupMock func()
		wantError error
	}{
		{
			name:      "ERROR:Invalid payload",
			payload:   []byte("{B}"),
			setupMock: func() { /* Empty */ },
			wantError: errors.New("json Unmarshal: invalid character 'B' looking for beginning of object key string"),
		},
		{
			name: "WARN:Transaction not found",
			setupMock: func() {
				ledgerSvc.On("FindByID", mock.Anything, externalId).Once().Return(nil, constant.ErrDataNotFound)
			},
			wantError: nil,
		},
		{
			name: "ERROR:Find transaction by id",
			setupMock: func() {
				ledgerSvc.On("FindByID", mock.Anything, externalId).Once().Return(nil, assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name: "WARN:Transaction type not registered",
			setupMock: func() {
				ledgerSvc.On(
					"FindByID", mock.Anything, externalId,
				).Once().Return(&orchestrator_model.AccountTransactionWithUseCase{
					Type: constant.TypeFee,
				}, nil)
			},
			wantError: nil,
		},
		{
			name: "SUCCESS:Bulk Disbursement",
			setupMock: func() {
				ledgerSvc.On(
					"FindByID", mock.Anything, externalId,
				).Once().Return(&orchestrator_model.AccountTransactionWithUseCase{
					Type: constant.TypeBulkDisbursement,
				}, nil)
				disbursementSvc.On("ProcessUpdateTransferStatus", mock.Anything, mock.Anything).Once().Return(nil)
			},
			wantError: nil,
		},
		{
			name: "SUCCESS:Single Disbursement",
			setupMock: func() {
				ledgerSvc.On(
					"FindByID", mock.Anything, externalId,
				).Once().Return(&orchestrator_model.AccountTransactionWithUseCase{
					Type: constant.TypeDisbursement,
				}, nil)
				disbursementSvc.On("ProcessUpdateTransferStatus", mock.Anything, mock.Anything).Once().Return(nil)
			},
			wantError: nil,
		},
		{
			name: "SUCCESS:Refund",
			setupMock: func() {
				ledgerSvc.On(
					"FindByID", mock.Anything, externalId,
				).Once().Return(&orchestrator_model.AccountTransactionWithUseCase{
					Type: constant.TypeRefund,
				}, nil)
				refundProcSvc.On("ProcessUpdateBankTransferStatus", mock.Anything, mock.Anything).Once().Return(nil)
			},
			wantError: nil,
		},
		{
			name: "SUCCESS:Withdrawal",
			setupMock: func() {
				ledgerSvc.On(
					"FindByID", mock.Anything, externalId,
				).Once().Return(&orchestrator_model.AccountTransactionWithUseCase{
					Type: constant.TypeWithdrawal,
				}, nil)
				withdrawalSvc.On("UpdateBankTransferStatus", mock.Anything, mock.Anything).Once().Return(nil)
			},
			wantError: nil,
		},
		{
			name: "SUCCESS:Card-funded payout",
			setupMock: func() {
				ledgerSvc.On(
					"FindByID", mock.Anything, externalId,
				).Once().Return(&orchestrator_model.AccountTransactionWithUseCase{
					Type:      constant.TypeDisbursement,
					Reference: constant.TypePaymentFundedPayout,
				}, nil)
				cardFundedPayoutSvc.On("UpdateBankTransferStatus", mock.Anything, mock.Anything).Once().Return(nil)
			},
			wantError: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()
			if test.payload == nil {
				test.payload = payload
			}

			assert.Equal(t, test.wantError, handler.UpdateTransferStatus(t.Context(), test.payload, ""))

			ledgerSvc.AssertExpectations(t)
			withdrawalSvc.AssertExpectations(t)
			refundProcSvc.AssertExpectations(t)
			disbursementSvc.AssertExpectations(t)
			cardFundedPayoutSvc.AssertExpectations(t)
		})
	}
}

func TestCutOffReportTrigger(t *testing.T) {
	startAt := time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC)
	endAt := time.Date(2026, 2, 10, 2, 0, 0, 0, time.UTC)
	validPayload := &bankTransferProto.CutOffReportTrigger{
		PartnerCode:         "014",
		PartnerName:         "BCA",
		CutoffWindowStartAt: startAt.Format(time.RFC3339),
		CutoffWindowEndAt:   endAt.Format(time.RFC3339),
		ExternalID:          "trx-1",
	}
	validBody, err := proto.Marshal(validPayload)
	assert.NoError(t, err)

	tests := []struct {
		name             string
		body             []byte
		setupMock        func(redisExt *redisMocks.IRedisExt, disbursementSvc *serviceMocks.IDisbursementService, reportDone chan struct{})
		wantError        string
		expectReportCall bool
	}{
		{
			name:      "ERROR: invalid protobuf payload",
			body:      []byte("invalid-proto"),
			wantError: "unmarshal cutoff report trigger payload",
		},
		{
			name: "ERROR: missing mandatory payload field",
			body: func() []byte {
				payload := proto.Clone(validPayload).(*bankTransferProto.CutOffReportTrigger)
				payload.ExternalID = ""
				body, _ := proto.Marshal(payload)
				return body
			}(),
			wantError: "invalid cutoff report trigger payload",
		},
		{
			name: "ERROR: invalid cutoff start time format",
			body: func() []byte {
				payload := proto.Clone(validPayload).(*bankTransferProto.CutOffReportTrigger)
				payload.CutoffWindowStartAt = "not-rfc3339"
				body, _ := proto.Marshal(payload)
				return body
			}(),
			wantError: "parse cutoff_window_start_at",
		},
		{
			name: "ERROR: member dedup redis failure",
			body: validBody,
			setupMock: func(redisExt *redisMocks.IRedisExt, _ *serviceMocks.IDisbursementService, _ chan struct{}) {
				redisExt.On("SetNX", mock.Anything, mock.AnythingOfType("string"), true, 72*time.Hour).Return(redis.NewBoolResult(false, assert.AnError)).Once()
			},
			wantError: "set member dedup key",
		},
		{
			name: "SUCCESS: duplicate member no-op",
			body: validBody,
			setupMock: func(redisExt *redisMocks.IRedisExt, _ *serviceMocks.IDisbursementService, _ chan struct{}) {
				redisExt.On("SetNX", mock.Anything, mock.MatchedBy(func(key string) bool {
					return strings.Contains(key, ":cutoff-report:member:")
				}), true, 72*time.Hour).Return(redis.NewBoolResult(false, nil)).Once()
			},
		},
		{
			name: "ERROR: member index hset failure",
			body: validBody,
			setupMock: func(redisExt *redisMocks.IRedisExt, _ *serviceMocks.IDisbursementService, _ chan struct{}) {
				redisExt.On("SetNX", mock.Anything, mock.MatchedBy(func(key string) bool {
					return strings.Contains(key, ":cutoff-report:member:")
				}), true, 72*time.Hour).Return(redis.NewBoolResult(true, nil)).Once()
				redisExt.On("HSet", mock.Anything, mock.AnythingOfType("string"), validPayload.ExternalID, 1).Return(redis.NewIntResult(0, assert.AnError)).Once()
			},
			wantError: "add member to index",
		},
		{
			name: "SUCCESS: not executor after indexing",
			body: validBody,
			setupMock: func(redisExt *redisMocks.IRedisExt, _ *serviceMocks.IDisbursementService, _ chan struct{}) {
				redisExt.On("SetNX", mock.Anything, mock.MatchedBy(func(key string) bool {
					return strings.Contains(key, ":cutoff-report:member:")
				}), true, 72*time.Hour).Return(redis.NewBoolResult(true, nil)).Once()
				redisExt.On("HSet", mock.Anything, mock.AnythingOfType("string"), validPayload.ExternalID, 1).Return(redis.NewIntResult(1, nil)).Once()
				redisExt.On("Expire", mock.Anything, mock.AnythingOfType("string"), 72*time.Hour).Return(redis.NewBoolResult(true, nil)).Once()
				redisExt.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.Anything, 72*time.Hour).Return(redis.NewStatusResult("OK", nil)).Once()
				redisExt.On("SetNX", mock.Anything, mock.MatchedBy(func(key string) bool {
					return strings.Contains(key, ":cutoff-report:dedup:")
				}), true, 72*time.Hour).Return(redis.NewBoolResult(false, nil)).Once()
			},
		},
		{
			name: "SUCCESS: executor generates report after settle",
			body: validBody,
			setupMock: func(redisExt *redisMocks.IRedisExt, disbursementSvc *serviceMocks.IDisbursementService, reportDone chan struct{}) {
				redisExt.On("SetNX", mock.Anything, mock.MatchedBy(func(key string) bool {
					return strings.Contains(key, ":cutoff-report:member:")
				}), true, 72*time.Hour).Return(redis.NewBoolResult(true, nil)).Once()
				redisExt.On("HSet", mock.Anything, mock.AnythingOfType("string"), validPayload.ExternalID, 1).Return(redis.NewIntResult(1, nil)).Once()
				redisExt.On("Expire", mock.Anything, mock.AnythingOfType("string"), 72*time.Hour).Return(redis.NewBoolResult(true, nil)).Once()
				redisExt.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.Anything, 72*time.Hour).Return(redis.NewStatusResult("OK", nil)).Once()
				redisExt.On("SetNX", mock.Anything, mock.MatchedBy(func(key string) bool {
					return strings.Contains(key, ":cutoff-report:dedup:")
				}), true, 72*time.Hour).Return(redis.NewBoolResult(true, nil)).Once()
				redisExt.On("Get", mock.Anything, mock.AnythingOfType("string")).Return(redis.NewStringResult(strconv.FormatInt(time.Now().Add(-10*time.Second).UnixNano(), 10), nil)).Maybe()
				redisExt.On("HGetAllScan", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Run(func(args mock.Arguments) {
					if dst, ok := args.Get(2).(*map[string]string); ok {
						(*dst)[validPayload.ExternalID] = "1"
					}
				}).Return(nil).Once()
				disbursementSvc.On("ReportAfterPayoutCutOffTimeByPartnerWindow", mock.Anything, mock.AnythingOfType("*disbursementModel.PartnerWindowCutOffReportRequest")).Run(func(args mock.Arguments) {
					reportDone <- struct{}{}
				}).Return(disbursementModel.AfterPayoutCutOffTimeSummary{}, nil).Once()
			},
			expectReportCall: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			log := logger.NewSlogger(logger.Config{})
			ledgerSvc := serviceMocks.NewIOrchestratorService(t)
			withdrawalSvc := serviceMocks.NewIWithdrawalService(t)
			disbursementSvc := serviceMocks.NewIDisbursementService(t)
			refundProcSvc := serviceMocks.NewIRefundProcessorService(t)
			redisExt := redisMocks.NewIRedisExt(t)
			reportDone := make(chan struct{}, 1)

			handler := New(log, &Service{
				LedgerSvc:       ledgerSvc,
				DisbursementSvc: disbursementSvc,
				RefundProcSvc:   refundProcSvc,
				WithdrawalSvc:   withdrawalSvc,
				RedisExt:        redisExt,
			})

			if tc.setupMock != nil {
				tc.setupMock(redisExt, disbursementSvc, reportDone)
			}

			err := handler.CutOffReportTrigger(t.Context(), tc.body, "")
			if tc.wantError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantError)
			}

			if tc.expectReportCall {
				select {
				case <-reportDone:
				case <-time.After(2 * time.Second):
					t.Fatal("expected partner window report to be generated")
				}
			}

			redisExt.AssertExpectations(t)
			disbursementSvc.AssertExpectations(t)
		})
	}
}
