package disbursementRepository_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/disbursement"
	mySqlExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetAvgDurationOfBankTransferProcessInMs(t *testing.T) {
	db := mySqlExtMock.NewIMySqlExt(t)

	repo := New(db, nil)

	now := time.Now().UTC()

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult float64
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), mock.Anything, c.StringMockType(), c.TimeMockType(), c.TimeMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), mock.Anything, c.StringMockType(), c.TimeMockType(), c.TimeMockType(),
				).Run(func(args mock.Arguments) {
					*args.Get(1).(*float64) = 10.125
				}).Return(nil)
			},
			wantResult: 10.125,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.GetAvgDurationOfBankTransferProcessInMs(context.Background(), now, now)
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestGetSummaryOfDelayedTransactionBeforeProcessed(t *testing.T) {
	db := mySqlExtMock.NewIMySqlExt(t)

	repo := New(db, nil)

	now := time.Now().UTC()
	banks := []disbursementModel.AfterPayoutCutOffTimeBankSummary{
		{Name: "Bank Dummy", Total: 1, Amount: 12_500},
	}
	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult disbursementModel.AfterPayoutCutOffTimeSummary
	}{
		{
			name: "SUCCESS:Data not found",
			setupMock: func() {
				db.On(
					"SelectContext", c.ValueCtxMockType(), mock.Anything, c.StringMockType(), c.TimeMockType(), c.TimeMockType(),
				).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"SelectContext", c.ValueCtxMockType(), mock.Anything, c.StringMockType(), c.TimeMockType(), c.TimeMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"SelectContext", c.ValueCtxMockType(), mock.Anything, c.StringMockType(), c.TimeMockType(), c.TimeMockType(),
				).Run(func(args mock.Arguments) {
					*args.Get(1).(*[]disbursementModel.AfterPayoutCutOffTimeBankSummary) = banks
				}).Return(nil)
			},
			wantResult: disbursementModel.AfterPayoutCutOffTimeSummary{
				Total:  1,
				Amount: 12_500,
				Banks:  banks,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.GetSummaryOfDelayedTransactionBeforeProcessed(context.Background(), now, now)
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
