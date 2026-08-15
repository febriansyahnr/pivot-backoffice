package qris_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/qris"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/qris"
	mySqlExt "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestFindQrRegistrationByExternalID(t *testing.T) {
	db := mySqlExt.NewIMySqlExt(t)
	repo := New(db)

	qrisRegistration := &qris.Registration{}

	tests := []struct {
		name      string
		setupMock func()
		wantErr   error
	}{
		{
			name: "ERROR:Not found",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), mock.AnythingOfType("*qris.Registration"), c.StringMockType(), c.StringMockType(),
				).Once().Return(sql.ErrNoRows)
			},
			wantErr: nil,
		},
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), mock.AnythingOfType("*qris.Registration"), c.StringMockType(), c.StringMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), mock.AnythingOfType("*qris.Registration"), c.StringMockType(), c.StringMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					ptr := args.Get(1).(*qris.Registration)
					*ptr = *qrisRegistration
				})
			},
			wantErr: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			if _, err := repo.FindQrRegistrationByExternalID(context.Background(), util.GenerateULID()); test.wantErr == nil {
				assert.NoError(t, err)

			} else {
				require.Error(t, err)
				require.Equal(t, test.wantErr, err)
			}
		})
	}
}

func TestFindQrRegistrationByExternalIDAndAcquirer(t *testing.T) {
	db := mySqlExt.NewIMySqlExt(t)
	repo := New(db)

	qrisRegistration := &qris.Registration{}

	tests := []struct {
		name      string
		setupMock func()
		wantErr   error
	}{
		{
			name: "ERROR:Not found",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), mock.AnythingOfType("*qris.Registration"), c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(sql.ErrNoRows)
			},
			wantErr: nil,
		},
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), mock.AnythingOfType("*qris.Registration"), c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), mock.AnythingOfType("*qris.Registration"), c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					ptr := args.Get(1).(*qris.Registration)
					*ptr = *qrisRegistration
				})
			},
			wantErr: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			if _, err := repo.FindQrRegistrationByExternalIDAndAcquirer(context.Background(), "external-id", "acquirer"); test.wantErr == nil {
				assert.NoError(t, err)

			} else {
				require.Error(t, err)
				require.Equal(t, test.wantErr, err)
			}
		})
	}
}
