package statusHistoriesRepository_test

import (
	"context"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	statusHistoriesModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/statusHistory"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/statusHistories"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestInsert(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	ptrStatusHistoryMockType := mock.AnythingOfType("*statusHistoryModel.StatusHistory")

	repo := New(db)

	tests := []struct {
		name      string
		setupMock func()
		wantErr   string
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"NamedExecContext", c.ValueCtxMockType(), c.StringMockType(), ptrStatusHistoryMockType,
				).Once().Return(false, c.ErrSomeErrorForUnitTest)
			},
			wantErr: "some error",
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"NamedExecContext", c.ValueCtxMockType(), c.StringMockType(), ptrStatusHistoryMockType,
				).Return(true, nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			if err := repo.Insert(context.Background(), &statusHistoriesModel.StatusHistory{}); test.wantErr == "" {
				require.NoError(t, err)

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}
