package accounttransaction_repository_test

import (
	"context"
	"testing"
	"time"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/accountTransaction"
	mySqlExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"

	"github.com/jmoiron/sqlx/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetListOfTransactionReferenceIdsWithPendingStatus(t *testing.T) {
	db := mySqlExtMock.NewIMySqlExt(t)

	repo := New(db, nil)
	ptrJSONTextType := mock.AnythingOfType("*types.JSONText")

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult []string
	}{
		{
			name: "ERROR:Get context",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), ptrJSONTextType, c.StringMockType(), c.StringMockType(), c.StringMockType(), c.TimeMockType(), c.TimeMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), ptrJSONTextType, c.StringMockType(), c.StringMockType(), c.StringMockType(), c.TimeMockType(), c.TimeMockType(),
				).Run(func(args mock.Arguments) {
					*args.Get(1).(*types.JSONText) = types.JSONText([]byte(`["12345"]`))
				}).Return(nil)
			},
			wantResult: []string{"12345"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.GetListOfTransactionReferenceIdsWithPendingStatus(
				context.Background(), "123456", "6543321", time.Now().UTC(), time.Now().UTC(),
			)
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
