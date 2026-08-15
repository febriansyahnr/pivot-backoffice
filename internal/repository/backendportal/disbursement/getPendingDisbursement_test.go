package disbursementRepository

import (
	"context"
	"testing"
	"time"

	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/disbursement"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetPendingTransactionsBetweenTimeForInquiryTransaction(t *testing.T) {
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	mockMysql := mysqlMocks.NewIMySqlExt(t)

	ctx := context.Background()
	from := time.Now()
	to := time.Now()

	pendingDisbursements := []*disbursementModel.DisbursementWithTransaction{
		{
			Disbursement: disbursementModel.Disbursement{
				UUID:        "valid-uuid",
				ReferenceID: "valid-reference-id",
				Status:      constant.DisbursementStatusApproved,
			},
		},
		{
			Disbursement: disbursementModel.Disbursement{
				UUID:        "valid-uuid-2",
				ReferenceID: "valid-reference-id-2",
				Status:      constant.DisbursementStatusApproved,
			},
		},
	}

	repo := New(mockMysql, mockLogger)

	testCases := []struct {
		name        string
		from        time.Time
		to          time.Time
		setupMock   func()
		shouldError bool
		want        []*disbursementModel.DisbursementWithTransaction
		wantErr     error
	}{
		{
			name: "when the query is failed, then should return error",
			from: from,
			to:   to,
			setupMock: func() {
				mockMysql.On("SelectContext", constant.ValueCtxMockType(), mock.Anything, constant.StringMockType(), constant.TimeMockType(), constant.TimeMockType()).Return(constant.ErrSomeErrorForUnitTest).Once()
			},
			shouldError: true,
			wantErr:     constant.ErrSomeErrorForUnitTest,
		},
		{
			name: "when the query is success, then should not return error",
			from: from,
			to:   to,
			setupMock: func() {
				mockMysql.On("SelectContext", constant.ValueCtxMockType(), mock.Anything, constant.StringMockType(), constant.TimeMockType(), constant.TimeMockType()).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(1).(*[]*disbursementModel.DisbursementWithTransaction) = pendingDisbursements
				}).Once()
			},
			shouldError: false,
			want:        pendingDisbursements,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			pendingDisbursements, err := repo.GetPendingTransactionsBetweenTimeForInquiryTransaction(ctx, tc.from, tc.to)
			if tc.shouldError {
				assert.Error(t, err)
				assert.Equal(t, tc.wantErr, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.want, pendingDisbursements)
		})
	}
}
