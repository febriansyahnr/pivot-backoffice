package credential_test

import (
	"context"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/credential"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/credential"
	mysqlMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateClientSecretById(t *testing.T) {
	db := mysqlMock.NewIMySqlExt(t)

	repo := New(db)
	tests := []struct {
		name      string
		mockSetup func(db *mysqlMock.IMySqlExt)
		wantErr   string
	}{
		{
			name: "ERROR:Some internal error",
			mockSetup: func(db *mysqlMock.IMySqlExt) {
				db.On(
					"ExecContext", constant.ValueCtxMockType(),
					constant.StringMockType(), constant.StringMockType(), uint(1), timeType, constant.StringMockType(), constant.StringMockType(),
				).Once().Return(false, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: "some error",
		},
		{
			name: "SUCCESS",
			mockSetup: func(db *mysqlMock.IMySqlExt) {
				db.On(
					"ExecContext", constant.ValueCtxMockType(),
					constant.StringMockType(), constant.StringMockType(), uint(1), timeType, constant.StringMockType(), constant.StringMockType(),
				).Return(true, nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.mockSetup(db)

			if _, err := repo.UpdateClientSecretById(context.Background(), uuid.NewString(), uuid.NewString(), &credential.ClientSecret{SecretVersion: 1}); test.wantErr == "" {
				require.NoError(t, err)

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}
