package outbound

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/outbound"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestFindByID(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	ptrOutboundMockType := mock.AnythingOfType("*outbound.Outbound")

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
					"GetContext", c.ValueCtxMockType(), ptrOutboundMockType, c.StringMockType(), c.StringMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: "some error",
		},
		{
			name: "ERROR:err no row",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), ptrOutboundMockType, c.StringMockType(), c.StringMockType(),
				).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"GetContext",
					c.ValueCtxMockType(),
					ptrOutboundMockType,
					c.StringMockType(),
					c.StringMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					merchantAuth := args.Get(1).(*outbound.Outbound)
					*merchantAuth = outbound.Outbound{
						Id: uuid.NewString(),
					}
				})
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			_, err := repo.FindByID(context.Background(), uuid.NewString())
			if tc.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.wantErr)
			}
		})
	}
}
