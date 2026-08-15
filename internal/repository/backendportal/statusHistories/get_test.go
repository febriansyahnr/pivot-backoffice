package statusHistoriesRepository_test

import (
	"context"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	statusHistoriesModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/statusHistory"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/backendportal/statusHistories"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetByReference(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	resultMockType := mock.AnythingOfType("*[]*statusHistoryModel.StatusHistory")

	repo := New(db)

	tests := []struct {
		name          string
		setupMock     func()
		referenceType string
		referenceID   string
		wantErr       string
		wantResult    []*statusHistoriesModel.StatusHistory
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"SelectContext", c.ValueCtxMockType(), resultMockType, c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			referenceType: "disbursement",
			referenceID:   "test-id",
			wantErr:       "some error",
		},
		{
			name: "SUCCESS:Empty result",
			setupMock: func() {
				db.On(
					"SelectContext", c.ValueCtxMockType(), resultMockType, c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Run(func(args mock.Arguments) {
					result := args.Get(1).(*[]*statusHistoriesModel.StatusHistory)
					*result = []*statusHistoriesModel.StatusHistory{}
				}).Return(nil)
			},
			referenceType: "disbursement",
			referenceID:   "test-id",
			wantResult:    []*statusHistoriesModel.StatusHistory{},
		},
		{
			name: "SUCCESS:With metadata",
			setupMock: func() {
				db.On(
					"SelectContext", c.ValueCtxMockType(), resultMockType, c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Return(nil)
			},
			referenceType: "disbursement",
			referenceID:   "test-reference-id",
			wantResult:    []*statusHistoriesModel.StatusHistory{},
		},
		{
			name: "SUCCESS:Without metadata",
			setupMock: func() {
				db.On(
					"SelectContext", c.ValueCtxMockType(), resultMockType, c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Return(nil)
			},
			referenceType: "disbursement",
			referenceID:   "test-reference-id-2",
			wantResult:    []*statusHistoriesModel.StatusHistory{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.GetByReference(context.Background(), test.referenceType, test.referenceID)

			if test.wantErr == "" {
				require.NoError(t, err)
				assert.Equal(t, test.wantResult, result)
			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
				assert.Nil(t, result)
			}
		})
	}
}
