package role

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	menuModel "github.com/paper-indonesia/pivot-backoffice/internal/model/menu"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/role"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockRole "github.com/paper-indonesia/pivot-backoffice/mocks/repository/role"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	now := time.Now()

	createdRole := &role.Role{
		UUID:      "uuid-uuid-uuid",
		Name:      "admin",
		Slug:      "admin",
		CreatedAt: now,
		UpdatedAt: now,
	}

	testCases := []struct {
		name       string
		input      *role.Role
		mocksSetup func(trxRepo *mockRole.IRoleRepository)
		wantErr    bool
	}{
		{
			name:  "SUCCESS: successfully create role",
			input: createdRole,
			mocksSetup: func(trxRepo *mockRole.IRoleRepository) {
				trxRepo.On(
					"Create",
					mock.Anything,
					constant.PtrRoleMockType(),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name:  "ERROR: error create role",
			input: createdRole,
			mocksSetup: func(trxRepo *mockRole.IRoleRepository) {
				trxRepo.On(
					"Create",
					mock.Anything,
					constant.PtrRoleMockType(),
				).Return(errors.New("insert error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			roleRepo := mockRole.NewIRoleRepository(t)
			loggerMock, _ := mockLogger.NewZapLogger(mockLogger.Config{})

			tc.mocksSetup(roleRepo)

			trxSvc := New(roleRepo, loggerMock)

			ctx := context.Background()
			err := trxSvc.Create(ctx, tc.input)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			roleRepo.AssertExpectations(t)
		})
	}
}

func TestCreateRoleAndPermissions(t *testing.T) {

	menuRepoMock := repoMocks.NewIMenuRepository(t)
	roleRepoMock := mockRole.NewIRoleRepository(t)
	roleMenuPermMock := repoMocks.NewIRoleMenuPermissionRepository(t)

	service := New(roleRepoMock, nil, WithMenuRepository(menuRepoMock), WithRoleMenuPermissionRepository(roleMenuPermMock))

	request := &role.CreateRoleRequest{
		Name: "Tester",
		Menus: []role.RoleMenuRequest{
			{
				Slug:        "dashboard-test",
				Permissions: []string{"dashboard-test.view"},
			},
		},
		MerchantID: "unique-id",
	}
	reqWithCombinationIsNotAllowed := &role.CreateRoleRequest{
		Name: "Tester",
		Menus: []role.RoleMenuRequest{
			{
				Slug:        "disbursement.disbursement-create",
				Permissions: []string{"disbursement-create.create"},
			},
			{
				Slug:        "disbursement.disbursement-approval",
				Permissions: []string{"disbursement-approval.create"},
			},
		},
		MerchantID: "unique-id",
	}
	ptrRoleMenuPermMockType := mock.AnythingOfType("*roleMenuPermissionModel.RoleMenuPermission")

	tests := []struct {
		name         string
		request      *role.CreateRoleRequest
		mockModifier func()
		wantErr      string
	}{
		{
			name: "ERROR:Total role by merchant ID/Invalid session",
			mockModifier: func() {
				roleRepoMock.On(
					"TotalRoleByMerchantID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Once().Return(uint64(0), errors.New("db.GetContext: Total Role: Invalid session"))
			},
			wantErr: "db.GetContext: Total Role: Invalid session",
		},
		{
			name: "ERROR:Total role by merchant ID/Limit exceeded",
			mockModifier: func() {
				roleRepoMock.On(
					"TotalRoleByMerchantID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Once().Return(uint64(10), nil)
			},
			wantErr: "role limit have been exceeded",
		},
		{
			name: "ERROR:Check available role name/Invalid session",
			mockModifier: func() {
				roleRepoMock.On(
					"TotalRoleByMerchantID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(uint64(1), nil)

				roleRepoMock.On(
					"CheckAvailableRoleName", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Return(false, errors.New("db.GetContext: Check available role name: Invalid session"))
			},
			wantErr: "db.GetContext: Check available role name: Invalid session",
		},
		{
			name: "ERROR:Check available role name/Not available",
			mockModifier: func() {
				roleRepoMock.On(
					"CheckAvailableRoleName", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Return(false, nil)
			},
			wantErr: "role name is already in use",
		},
		{
			name:    "ERROR:Combination of menu access is not allowed",
			request: reqWithCombinationIsNotAllowed,
			mockModifier: func() {
				roleRepoMock.On(
					"CheckAvailableRoleName", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(),
				).Return(true, nil)
			},
			wantErr: "combination of menu access is not allowed",
		},
		{
			name: "ERROR:Get menu and permissions/Invalid session",
			mockModifier: func() {
				menuRepoMock.On(
					"GetMenuAndPermissionIDs", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Return(nil, errors.New("Invalid Session"))
			},
			wantErr: "Invalid Session",
		},
		{
			name: "ERROR:Get menu and permissions/Data not registered",
			mockModifier: func() {
				menuRepoMock.On(
					"GetMenuAndPermissionIDs", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Return(nil, nil)
			},
			wantErr: "menu or permission not registered",
		},
		{
			name: "ERROR:Begin transaction",
			mockModifier: func() {
				menuRepoMock.On(
					"GetMenuAndPermissionIDs", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(),
				).Return(&menuModel.MenuAndPermissionIDs{
					Permissions: []menuModel.ObjPermission{{}},
				}, nil)
				roleRepoMock.On(
					"BeginTransaction", constant.ValueCtxMockType(),
				).Once().Return(nil, errors.New("begin transaction: invalid session"))
			},
			wantErr: "begin transaction: invalid session",
		},
		{
			name: "ERROR:Rollback transaction",
			mockModifier: func() {
				roleRepoMock.On(
					"BeginTransaction", constant.ValueCtxMockType(),
				).Return(context.WithValue(context.Background(), mySqlExt.CtxSqlTx, &sqlx.Tx{}), nil)
				roleRepoMock.On(
					"Create", constant.ValueCtxMockType(), constant.PtrRoleMockType(),
				).Once().Return(errors.New("create: invalid session"))
				roleRepoMock.On(
					"RollbackTransaction", constant.ValueCtxMockType(),
				).Once().Return(errors.New("rollback transaction: invalid session"))
			},
			wantErr: "rollback transaction: invalid session",
		},
		{
			name: "ERROR:Create role",
			mockModifier: func() {
				roleRepoMock.On(
					"RollbackTransaction", constant.ValueCtxMockType(),
				).Return(nil)

				roleRepoMock.On(
					"Create", constant.ValueCtxMockType(), constant.PtrRoleMockType(),
				).Once().Return(errors.New("create role: invalid session"))
			},
			wantErr: "create role: invalid session",
		},
		{
			name: "ERROR:Create role and menu permission",
			mockModifier: func() {
				roleRepoMock.On(
					"Create", constant.ValueCtxMockType(), constant.PtrRoleMockType(),
				).Return(nil)
				roleMenuPermMock.On(
					"Create", constant.ValueCtxMockType(), ptrRoleMenuPermMockType,
				).Once().Return(errors.New("create role and menu permission: invalid session"))
			},
			wantErr: "create role and menu permission: invalid session",
		},
		{
			name: "ERROR:Commit transaction",
			mockModifier: func() {
				roleMenuPermMock.On(
					"Create", constant.ValueCtxMockType(), ptrRoleMenuPermMockType,
				).Return(nil)

				roleRepoMock.On(
					"CommitTransaction", constant.ValueCtxMockType(),
				).Once().Return(errors.New("commit transaction: invalid session"))
			},
			wantErr: "commit transaction: invalid session",
		},
		{
			name: "SUCCESS",
			mockModifier: func() {
				roleRepoMock.On(
					"CommitTransaction", constant.ValueCtxMockType(),
				).Once().Return(nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.mockModifier()

			if test.request == nil {
				test.request = request
			}
			if res, err := service.CreateRoleAndPermissions(context.Background(), test.request); test.wantErr != "" {
				require.Error(t, err)
				require.ErrorContains(t, err, test.wantErr)

			} else {
				require.NoError(t, err)
				require.NoError(t, uuid.Validate(res.ID))
			}
		})
	}
}
