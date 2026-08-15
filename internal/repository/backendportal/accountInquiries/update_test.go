package accountInquiries

import (
	"context"
	"errors"
	"testing"
	"time"

	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/accountInquiries"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdate(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	repo := New(db, logger)

	validData := &accountInquiries.AccountInquiries{
		UUID:                   "uuid-uuid-uuid",
		BeneficiaryAccountNo:   "1234567890",
		BeneficiaryAccountName: "test",
		BeneficiaryBankCode:    "1234",
		BeneficiaryBankName:    "test",
		Response:               "response",
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}

	testCases := []struct {
		desc      string
		wantErr   bool
		mockSetup func()
	}{
		{
			desc:    "error Named ExecContext",
			wantErr: true,
			mockSetup: func() {
				db.On("NamedExecContext", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*accountInquiries.AccountInquiries")).Return(false, errors.New("error")).Once()
			},
		},
		{
			desc:    "success Update data",
			wantErr: false,
			mockSetup: func() {
				db.On("NamedExecContext", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*accountInquiries.AccountInquiries")).Return(true, nil)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {

			tc.mockSetup()

			err := repo.Update(context.Background(), validData)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

		})
	}
}
