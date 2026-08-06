package requestaccountinquiry

import (
	"context"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	requestAccountInquiries "github.com/paper-indonesia/pivot-backoffice/internal/model/requestAccountInquiry"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdate(t *testing.T) {
	testCases := []struct {
		desc      string
		wantErr   bool
		mockSetup func(db *mysqlMocks.IMySqlExt)
	}{
		{
			desc:    "error when update request account inquiries",
			wantErr: true,
			mockSetup: func(db *mysqlMocks.IMySqlExt) {
				db.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("*requestAccountInquiries.RequestAccountInquiryWithMaster"),
				).Return(false, assert.AnError)
			},
		},
		{
			desc:    "error when update request account inquiries",
			wantErr: false,
			mockSetup: func(db *mysqlMocks.IMySqlExt) {
				db.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("*requestAccountInquiries.RequestAccountInquiryWithMaster"),
				).Return(true, nil)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			logger, _ := logger.NewZapLogger(logger.Config{})
			db := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(db)

			repo := New(db, logger)
			err := repo.Update(context.Background(), &requestAccountInquiries.RequestAccountInquiryWithMaster{})
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
