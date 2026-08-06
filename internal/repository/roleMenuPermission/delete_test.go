package roleMenuPermissionRepository

import (
	"context"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	mySqlExt "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDelete(t *testing.T) {
	mysqlMock := mySqlExt.NewIMySqlExt(t)

	repo := New(mysqlMock, nil)

	tests := []struct {
		name         string
		mockModifier func()
		wantErr      string
	}{
		{
			name: "ERROR:Invalid session",
			mockModifier: func() {
				mysqlMock.On(
					"ExecContext", constant.ValueCtxMockType(), mock.AnythingOfType("string"), mock.AnythingOfType("string"),
				).Once().Return(false, errors.New("invalid session"))
			},
			wantErr: "invalid session",
		},
		{
			name: "SUCCESS",
			mockModifier: func() {
				mysqlMock.On(
					"ExecContext", constant.ValueCtxMockType(), mock.AnythingOfType("string"), mock.AnythingOfType("string"),
				).Return(true, nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			test.mockModifier()

			if err := repo.Delete(context.Background(), "role-id"); test.wantErr == "" {
				require.NoError(t, err)

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}
