package role

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/role"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	roleRepositoryMock "github.com/paper-indonesia/pivot-backoffice/mocks/repository/role"
	userRoleRepositoryMock "github.com/paper-indonesia/pivot-backoffice/mocks/repository/userRole"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDelete(t *testing.T) {
	roleID := "cb3c4f47-9bcd-4876-a18e-a36492f8b089"
	merchantID := "597236a2-df14-44cc-836d-567f89b97d55"

	logMock, _ := loggerMock.NewZapLogger(loggerMock.Config{})

	roleRepoMock := roleRepositoryMock.NewIRoleRepository(t)
	userRoleRepoMock := userRoleRepositoryMock.NewIUserRoleRepository(t)
	roleMenuPermRepoMock := repositoryMocks.NewIRoleMenuPermissionRepository(t)

	service := New(roleRepoMock, logMock, WithRoleMenuPermissionRepository(roleMenuPermRepoMock), WithUserRoleRepository(userRoleRepoMock))

	const errRoleMenuInvalidSession = "role menu permission: invalid session"
	var mcb, mcv = constant.ValueCtxMockType(), constant.ValueCtxMockType()

	tests := []struct {
		name         string
		mockModifier func()
		wantErr      string
	}{
		{
			name: "ERROR:Find role by ID",
			mockModifier: func() {
				roleRepoMock.On(
					"FindRoleByID", mcb, constant.StringMockType(),
				).Once().Return(nil, errors.New("find role by id: invalid session"))
			},
			wantErr: "find role by id: invalid session",
		},
		{
			name: "ERROR:Role data not found",
			mockModifier: func() {
				roleRepoMock.On("FindRoleByID", mcb, constant.StringMockType()).Once().Return(nil, nil)
			},
			wantErr: "role not found",
		},
		{
			name: "ERROR:Total active users by role ID",
			mockModifier: func() {
				roleRepoMock.On(
					"FindRoleByID", mcb, constant.StringMockType(),
				).Return(&role.Role{MerchantID: sql.NullString{Valid: true, String: merchantID}}, nil)

				userRoleRepoMock.On(
					"TotalActiveUsersByRoleID", mcb, constant.StringMockType(),
				).Once().Return(uint64(0), errors.New("total active users by role id: invalid session"))
			},
			wantErr: "total active users by role id: invalid session",
		},
		{
			name: "ERROR:Role cannot be deleted",
			mockModifier: func() {
				userRoleRepoMock.On("TotalActiveUsersByRoleID", mcb, constant.StringMockType()).Once().Return(uint64(10), nil)
			},
			wantErr: "role cannot be deleted",
		},
		{
			name: "ERROR:Begin transaction",
			mockModifier: func() {
				userRoleRepoMock.On("TotalActiveUsersByRoleID", mcb, constant.StringMockType()).Return(uint64(0), nil)

				roleRepoMock.On(
					"BeginTransaction", constant.ValueCtxMockType(),
				).Once().Return(context.WithValue(context.Background(), mySqlExt.CtxSqlTx, &sqlx.Tx{}), errors.New("begin transaction: invalid session"))
			},
			wantErr: "begin transaction: invalid session",
		},
		{
			name: "ERROR:Rollback transaction",
			mockModifier: func() {
				roleRepoMock.On(
					"BeginTransaction", mcb,
				).Return(context.WithValue(context.Background(), mySqlExt.CtxSqlTx, &sqlx.Tx{}), nil)

				roleMenuPermRepoMock.On(
					"Delete", mcv, constant.StringMockType(),
				).Once().Return(errors.New(errRoleMenuInvalidSession))
				roleRepoMock.On("RollbackTransaction", mcv).Once().Return(errors.New("rollback transaction: invalid session"))
			},
			wantErr: "rollback transaction: invalid session",
		},
		{
			name: "ERROR:Delete role menu permission",
			mockModifier: func() {
				roleRepoMock.On("RollbackTransaction", mcv).Return(nil)

				roleMenuPermRepoMock.On(
					"Delete", mcv, constant.StringMockType(),
				).Once().Return(errors.New(errRoleMenuInvalidSession))
			},
			wantErr: errRoleMenuInvalidSession,
		},
		{
			name: "ERROR:Delete role",
			mockModifier: func() {
				roleMenuPermRepoMock.On(
					"Delete", mcv, constant.StringMockType(),
				).Return(nil)

				roleRepoMock.On(
					"Delete", mcv, constant.StringMockType(),
				).Once().Return(errors.New("role: invalid session"))
			},
			wantErr: "role: invalid session",
		},
		{
			name: "ERROR:Commit transaction",
			mockModifier: func() {
				roleRepoMock.On("Delete", mcv, constant.StringMockType()).Return(nil)

				roleRepoMock.On("CommitTransaction", mcv).Once().Return(errors.New("commit transaction: invalid session"))
			},
			wantErr: "commit transaction: invalid session",
		},
		{
			name:         "SUCCESS",
			mockModifier: func() { roleRepoMock.On("CommitTransaction", mcv).Return(nil) },
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			test.mockModifier()

			if err := service.Delete(context.Background(), merchantID, roleID); test.wantErr == "" {
				require.NoError(t, err)

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}
