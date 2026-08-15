package paymentRepository_test

import (
	"context"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/payment"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/payment"
	mySqlExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestChangePaymentMethod(t *testing.T) {
	db := mySqlExtMock.NewIMySqlExt(t)

	repo := New(db, nil)

	argMockType := mock.AnythingOfType("*paymentModel.PaymentDTO")

	tests := []struct {
		name      string
		setupMock func()
		wantErr   error
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				db.On(
					"NamedExecContext", c.ValueCtxMockType(), c.StringMockType(), argMockType,
				).Once().Return(false, c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				db.On(
					"NamedExecContext", c.ValueCtxMockType(), c.StringMockType(), argMockType,
				).Return(true, nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			assert.Equal(t, test.wantErr, repo.ChangePaymentMethod(context.Background(), &paymentModel.PaymentDTO{}))
		})
	}
}
