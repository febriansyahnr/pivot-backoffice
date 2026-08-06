package accountinquiry

import (
	"context"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"

	"github.com/google/uuid"
	requestAccountInquiry "github.com/paper-indonesia/pivot-backoffice/internal/model/requestAccountInquiry"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/bankAccount"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	mocks_repo "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/jmoiron/sqlx/types"
	mocks_logger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFindLatestByInquiryID(t *testing.T) {
	logger, _ := mocks_logger.NewZapLogger(mocks_logger.Config{})
	snapCore := mocks_repo.NewISnapCoreRepository(t)
	reqAccountRepo := mocks_repo.NewIRequestAccountInquiryRepository(t)
	accountInquiryRepo := mocks_repo.NewIAccountInquiriesRepository(t)
	accountTransactionSvc := mocks.NewIOrchestratorService(t)
	merchantService := mocks.NewIMerchantService(t)
	feeSvc := mocks.NewIFeeService(t)

	svc := New(logger, snapCore, reqAccountRepo, accountInquiryRepo, accountTransactionSvc, merchantService, feeSvc,
		WithConfig(&config.Config{Environment: "test"}),
	)

	testCases := []struct {
		desc      string
		wantErr   bool
		mockSetup func()
	}{
		{
			desc:    "ERROR: FindLatestByInquiryID database error",
			wantErr: true,
			mockSetup: func() {
				reqAccountRepo.On("FindLatestByInquiryID", mock.Anything, c.StringMockType(), c.StringMockType()).Return(nil, c.ErrSomeErrorForUnitTest).Once()
			},
		},
		{
			desc:    "ERROR: FindLatestByInquiryID not found",
			wantErr: true,
			mockSetup: func() {
				reqAccountRepo.On("FindLatestByInquiryID", mock.Anything, c.StringMockType(), c.StringMockType()).Return(nil, nil).Once()
			},
		},
		{
			desc:    "SUCCESS: FindLatestByInquiryID",
			wantErr: false,
			mockSetup: func() {
				reqAccountRepo.On("FindLatestByInquiryID", mock.Anything, c.StringMockType(), c.StringMockType()).Return(&requestAccountInquiry.RequestAccountInquiryWithMaster{}, nil).Once()
			},
		},
		{
			desc:    "SUCCESS: FindLatestByInquiryID with valid metadata and SnapCoreResponse",
			wantErr: false,
			mockSetup: func() {
				reqAccountRepo.On("FindLatestByInquiryID", mock.Anything, c.StringMockType(), c.StringMockType()).Return(&requestAccountInquiry.RequestAccountInquiryWithMaster{
					RequestAccountInquiries: requestAccountInquiry.RequestAccountInquiries{
						Metadata: types.NullJSONText{
							Valid: true,
						},
						MetadataObj: requestAccountInquiry.Metadata{
							SnapCoreResponse: &snapCoreModel.InquiryAccountResponseData{
								BeneficiaryAccountName: "Test Account Name From SnapCore",
							},
						},
					},
					MasterBeneficiaryAccountName: "Original Master Name",
				}, nil).Once()
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {

			tc.mockSetup()
			res, err := svc.FindLatestByInquiryID(context.Background(), uuid.NewString(), uuid.NewString())
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
			}

		})
	}
}
