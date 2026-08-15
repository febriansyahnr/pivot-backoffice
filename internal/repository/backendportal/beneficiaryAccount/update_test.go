package beneficiaryAccountRepository_test

import (
	"context"
	"errors"
	"testing"

	loggerMock "github.com/paper-indonesia/pdk/v2/logger"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	beneficiaryAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/beneficiaryAccount"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/beneficiaryAccount"
	mysqlMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdate(t *testing.T) {
	db := mysqlMock.NewIMySqlExt(t)
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})

	repo := New(db, logger)

	updateMockType := mock.AnythingOfType("*beneficiaryAccountModel.BeneficiaryAccount")

	tests := []struct {
		name      string
		setupMock func()
		wantErr   error
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				db.On(
					"NamedExecContext", c.ValueCtxMockType(), c.StringMockType(), updateMockType,
				).Once().Return(false, c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "ERROR:Now row affected", // NOSONAR
			setupMock: func() {
				db.On(
					"NamedExecContext", c.ValueCtxMockType(), c.StringMockType(), updateMockType,
				).Once().Return(false, nil)
			},
			wantErr: errors.New("no rows affected"),
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				db.On(
					"NamedExecContext", c.ValueCtxMockType(), c.StringMockType(), updateMockType,
				).Return(true, nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()
			assert.Equal(t, test.wantErr, repo.Update(context.Background(), &beneficiaryAccountModel.BeneficiaryAccount{}))
		})
	}
}
