package accounttransaction_repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	mySqlExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetPastDueSettlementTransactions(t *testing.T) {
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	db := mySqlExtMock.NewIMySqlExt(t)
	repo := New(db, mockLogger)

	now := time.Now().UTC()
	testUUID := uuid.New()
	merchantUUID := uuid.New()
	accountUUID := uuid.New()

	tests := []struct {
		name       string
		request    *orchestrator_model.GetPastDueSettlementTransactionsRequest
		setupMock  func()
		wantErr    error
		wantResult []*orchestrator_model.AccountTransaction
	}{
		{
			name: "ERROR: SelectContext fails",
			request: &orchestrator_model.GetPastDueSettlementTransactionsRequest{
				ReferenceID: "REF123",
				Datetime:    now,
			},
			setupMock: func() {
				db.On(
					"SelectContext",
					c.ValueCtxMockType(),
					mock.Anything,
					c.StringMockType(),
					c.StringMockType(),
					c.TimeMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr:    c.ErrSomeErrorForUnitTest,
			wantResult: nil,
		},
		{
			name: "SUCCESS: Returns transactions",
			request: &orchestrator_model.GetPastDueSettlementTransactionsRequest{
				ReferenceID: "REF123",
				Datetime:    now,
			},
			setupMock: func() {
				db.On(
					"SelectContext",
					c.ValueCtxMockType(),
					mock.Anything,
					c.StringMockType(),
					c.StringMockType(),
					c.TimeMockType(),
				).Run(func(args mock.Arguments) {
					result := args.Get(1).(*[]*orchestrator_model.AccountTransaction)
					*result = []*orchestrator_model.AccountTransaction{
						{
							UUID:                 testUUID,
							ReferenceID:          "REF123",
							MerchantID:           merchantUUID,
							AccountID:            accountUUID,
							Currency:             "IDR",
							Credit:               100000,
							Debit:                0,
							Reference:            "PMT123",
							Type:                 "PAYMENT",
							Channel:              "QRIS",
							Status:               "SUCCESS",
							SettlementStatus:     sql.NullString{String: "PENDING", Valid: true},
							TransactionTimestamp: now,
							CreatedAt:            now,
							UpdatedAt:            now,
							AdditionalInfo:       types.NullJSONText{JSONText: types.JSONText(`{"settlementDetail":{"type":"T+1"}}`), Valid: true},
						},
					}
				}).Return(nil)
			},
			wantErr: nil,
			wantResult: []*orchestrator_model.AccountTransaction{
				{
					UUID:                 testUUID,
					ReferenceID:          "REF123",
					MerchantID:           merchantUUID,
					AccountID:            accountUUID,
					Currency:             "IDR",
					Credit:               100000,
					Debit:                0,
					Reference:            "PMT123",
					Type:                 "PAYMENT",
					Channel:              "QRIS",
					Status:               "SUCCESS",
					SettlementStatus:     sql.NullString{String: "PENDING", Valid: true},
					TransactionTimestamp: now,
					CreatedAt:            now,
					UpdatedAt:            now,
					AdditionalInfo:       types.NullJSONText{JSONText: types.JSONText(`{"settlementDetail":{"type":"T+1"}}`), Valid: true},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			result, err := repo.GetPastDueSettlementTransactions(context.Background(), tt.request)

			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.wantResult, result)
		})
	}
}
