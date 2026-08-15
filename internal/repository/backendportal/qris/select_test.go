package qris_test

import (
	"context"
	"database/sql"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/qris"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/qris"
	mySqlExt "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestFindRegistrationById(t *testing.T) {
	db := mySqlExt.NewIMySqlExt(t)

	repo := New(db)

	ptrRegistrationMerchantMockType := mock.AnythingOfType("*qris.RegistrationMerchant")

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    string
		wantResult *qris.RegistrationMerchant
	}{
		{
			name: "ERROR:Data not found",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), ptrRegistrationMerchantMockType, c.StringMockType(), c.StringMockType(),
				).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), ptrRegistrationMerchantMockType, c.StringMockType(), c.StringMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), ptrRegistrationMerchantMockType, c.StringMockType(), c.StringMockType(),
				).Run(func(args mock.Arguments) {
					*args.Get(1).(*qris.RegistrationMerchant) = qris.RegistrationMerchant{Id: "ID", ExternalId: "EX"}
				}).Return(nil)
			},
			wantResult: &qris.RegistrationMerchant{Id: "ID", ExternalId: "EX"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			if result, err := repo.FindRegistrationById(context.Background(), uuid.NewString()); test.wantErr == "" {
				require.NoError(t, err)
				assert.Equal(t, test.wantResult, result)

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}
