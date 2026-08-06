package xbPayoutService

import (
	"context"
	"testing"

	disburmentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	xbCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xbCoreProcessor"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetPayoutById(t *testing.T) {
	testCases := []struct {
		name      string
		wantError bool
		setupMock func(
			mockXbCoreProcessorRepo *repositoryMocks.IXbCoreProcessorRepository,
			mockDisbursementRepo *repositoryMocks.IDisbursementRepository,
		)
	}{
		{
			name:      "ERROR: error when find disbursement by id",
			wantError: true,
			setupMock: func(
				mockXbCoreProcessorRepo *repositoryMocks.IXbCoreProcessorRepository,
				mockDisbursementRepo *repositoryMocks.IDisbursementRepository,
			) {
				mockDisbursementRepo.On("FindByID", mock.Anything, mock.Anything).Return(nil, assert.AnError)
			},
		},
		{
			name:      "ERROR: error not found disbursement by id",
			wantError: true,
			setupMock: func(_ *repositoryMocks.IXbCoreProcessorRepository, mockDisbursementRepo *repositoryMocks.IDisbursementRepository) {
				mockDisbursementRepo.On("FindByID", mock.Anything, mock.Anything).Return(nil, nil)
			},
		},
		{
			name:      "ERROR: merchant ID does not match",
			wantError: true,
			setupMock: func(_ *repositoryMocks.IXbCoreProcessorRepository, mockDisbursementRepo *repositoryMocks.IDisbursementRepository) {
				mockDisbursementRepo.On("FindByID", mock.Anything, mock.Anything).Return(&disburmentModel.DisbursementWithTransaction{
					Disbursement: disburmentModel.Disbursement{MerchantID: "123456"}, // NOSONAR
				}, nil)
			},
		},
		{
			name:      "ERROR: error when request get payout id to xb core processor",
			wantError: true,
			setupMock: func(
				mockXbCoreProcessorRepo *repositoryMocks.IXbCoreProcessorRepository,
				mockDisbursementRepo *repositoryMocks.IDisbursementRepository,
			) {
				mockDisbursementRepo.On("FindByID", mock.Anything, mock.Anything).Return(&disburmentModel.DisbursementWithTransaction{
					Disbursement: disburmentModel.Disbursement{
						MerchantID: "merchantId",
						MetadataObj: disburmentModel.Metadata{
							XbDetail: &xbModel.XbPayoutMetadata{
								Uuid: "uuid",
							},
						},
					},
				}, nil)
				mockXbCoreProcessorRepo.On("GetPayoutById", mock.Anything, mock.Anything).Return(nil, assert.AnError)
			},
		},
		{
			name:      "SUCCESS: success get payout id",
			wantError: false,
			setupMock: func(
				mockXbCoreProcessorRepo *repositoryMocks.IXbCoreProcessorRepository,
				mockDisbursementRepo *repositoryMocks.IDisbursementRepository,
			) {
				mockDisbursementRepo.On("FindByID", mock.Anything, mock.Anything).Return(&disburmentModel.DisbursementWithTransaction{
					Disbursement: disburmentModel.Disbursement{
						MerchantID: "merchantId",
						MetadataObj: disburmentModel.Metadata{
							XbDetail: &xbModel.XbPayoutMetadata{
								Uuid: "uuid",
							},
						},
					},
				}, nil)
				mockXbCoreProcessorRepo.On("GetPayoutById", mock.Anything, mock.Anything).Return(&xbCoreProcessorModel.GetPayoutResponseData{
					RfiDetails: []*xbCoreProcessorModel.RfiDetails{
						{
							PartnerDocumentID: "partnerDocumentID",
						},
						{
							PartnerDocumentID: "partnerDocumentID",
						},
					},
				}, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockXbCoreProcessorRepo := repositoryMocks.NewIXbCoreProcessorRepository(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockDisbursement := repositoryMocks.NewIDisbursementRepository(t)
			mockBeneficiaryAccount := repositoryMocks.NewIBeneficiaryAccountRepository(t)

			tc.setupMock(mockXbCoreProcessorRepo, mockDisbursement)

			s := New(mockLogger, mockDisbursement, mockBeneficiaryAccount, mockXbCoreProcessorRepo)

			_, err := s.GetPayoutById(context.Background(), &xbModel.GetPayoutRequest{
				MerchantId: "merchantId",
				PayoutId:   "id",
			})

			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockXbCoreProcessorRepo.AssertExpectations(t)

		})
	}
}
