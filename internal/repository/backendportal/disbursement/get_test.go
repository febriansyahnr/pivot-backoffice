package disbursementRepository_test

import (
	"database/sql"
	"testing"

	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/disbursement"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"

	"github.com/jmoiron/sqlx/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetDetailForCardFundedPayoutByID(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)

	payoutID := "3acd51f8-1029-4f27-899b-9dd91793acc2"

	tests := []struct {
		name       string
		setupMock  func()
		wantError  error
		wantResult *disbursementModel.Disbursement
	}{
		{
			name: "ERROR:Data not found", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, payoutID,
				).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, payoutID,
				).Once().Return(assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, payoutID,
				).Once().Run(func(args mock.Arguments) {
					*args.Get(1).(*disbursementModel.Disbursement) = disbursementModel.Disbursement{
						Metadata: types.NullJSONText{Valid: true, JSONText: []byte(`{}`)},
					}
				}).Return(nil)
			},
			wantResult: &disbursementModel.Disbursement{
				Metadata: types.NullJSONText{Valid: true, JSONText: []byte(`{}`)},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.GetDetailForCardFundedPayoutByID(t.Context(), payoutID)
			assert.Equal(t, test.wantError, err)
			assert.Equal(t, test.wantResult, result)

			db.AssertExpectations(t)
		})
	}
}
