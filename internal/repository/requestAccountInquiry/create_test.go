package requestaccountinquiry

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	requestAccountInquiry "github.com/paper-indonesia/pivot-backoffice/internal/model/requestAccountInquiry"
	mockDB "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
)

func TestCreateRequestAccountInquiry(t *testing.T) {
	db := mockDB.NewIMySqlExt(t)
	logger, _ := mockLogger.NewZapLogger(mockLogger.Config{})

	repo := New(db, logger)

	requestMock := requestAccountInquiry.RequestAccountInquiries{
		UUID:       uuid.NewString(),
		MerchantID: uuid.NewString(),
		AccountInquiryId: sql.NullString{
			String: uuid.NewString(),
			Valid:  true,
		},
		BeneficiaryBankCode: "008",
		BeneficiaryAccountNo: sql.NullString{
			String: "00809090909090909",
			Valid:  true,
		},
		CreatedAt: time.Now(),
	}

	testCases := []struct {
		desc      string
		wantErr   bool
		mockSetup func()
	}{
		{
			desc:    "error when insert into request_account_inquiries",
			wantErr: true,
			mockSetup: func() {
				db.On("NamedExecContext",
					mock.Anything,
					mock.Anything,
					constant.PtrRequestAccountInquiriesMockType()).
					Return(false, errors.New("error when insert into request_account_inquiries")).Once()

			},
		},
		{
			desc:    "error no row affected",
			wantErr: true,
			mockSetup: func() {
				db.On("NamedExecContext",
					mock.Anything,
					mock.Anything,
					constant.PtrRequestAccountInquiriesMockType()).
					Return(false, nil).Once()

			},
		},
		{
			desc:    "success create request account inquiries",
			wantErr: false,
			mockSetup: func() {
				db.On("NamedExecContext",
					mock.Anything,
					mock.Anything,
					constant.PtrRequestAccountInquiriesMockType()).
					Return(true, nil).Once()
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			tc.mockSetup()

			err := repo.Create(context.Background(), &requestMock)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
