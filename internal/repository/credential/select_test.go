package credential_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/credential"
	mysqlMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	db := mysqlMock.NewIMySqlExt(t)

	repo := New(db)
	tests := []struct {
		name      string
		mockSetup func(db *mysqlMock.IMySqlExt)
		wantErr   string
	}{
		{
			name: "ERROR:Get context for merchant data",
			mockSetup: func(db *mysqlMock.IMySqlExt) {
				db.On(
					"GetContext", constant.ValueCtxMockType(),
					constant.PtrStringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: "some error",
		},
		{
			name: "ERROR:Data not found",
			mockSetup: func(db *mysqlMock.IMySqlExt) {
				db.On(
					"GetContext", constant.ValueCtxMockType(),
					constant.PtrStringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "ERROR:Select context for merchant auths",
			mockSetup: func(db *mysqlMock.IMySqlExt) {
				db.On(
					"GetContext", constant.ValueCtxMockType(),
					constant.PtrStringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Return(nil)
				db.On(
					"SelectContext", constant.ValueCtxMockType(),
					clientSecretSumPtrSliceType, constant.StringMockType(), constant.StringMockType(),
				).Once().Return(fmt.Errorf("Select Context: %v", constant.ErrSomeErrorForUnitTest))
			},
			wantErr: "Select Context: some error",
		},
		{
			name: "SUCCESS",
			mockSetup: func(db *mysqlMock.IMySqlExt) {
				db.On(
					"SelectContext", constant.ValueCtxMockType(),
					clientSecretSumPtrSliceType, constant.StringMockType(), constant.StringMockType(),
				).Return(nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.mockSetup(db)

			if _, err := repo.Get(context.Background(), uuid.NewString()); test.wantErr == "" {
				require.NoError(t, err)

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}

func TestGetClientSecretById(t *testing.T) {
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
					"GetContext", constant.ValueCtxMockType(),
					clientSecretPtrType, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: "some error",
		},
		{
			name: "ERROR:Data not found",
			mockSetup: func(db *mysqlMock.IMySqlExt) {
				db.On(
					"GetContext", constant.ValueCtxMockType(),
					clientSecretPtrType, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "SUCCESS",
			mockSetup: func(db *mysqlMock.IMySqlExt) {
				db.On(
					"GetContext", constant.ValueCtxMockType(),
					clientSecretPtrType, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Return(nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.mockSetup(db)

			if _, err := repo.GetClientSecretById(context.Background(), uuid.NewString(), uuid.NewString()); test.wantErr == "" {
				require.NoError(t, err)

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}
