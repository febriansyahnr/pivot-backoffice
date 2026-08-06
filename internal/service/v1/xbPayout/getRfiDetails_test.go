package xbPayoutService

import (
	"context"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disburmentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	xbCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xbCoreProcessor"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetRfiDetails(t *testing.T) {
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
			setupMock: func(
				mockXbCoreProcessorRepo *repositoryMocks.IXbCoreProcessorRepository,
				mockDisbursementRepo *repositoryMocks.IDisbursementRepository,
			) {
				mockDisbursementRepo.On("FindByID", mock.Anything, mock.Anything).Return(nil, nil)
			},
		},
		{
			name:      "ERROR: error when request get payout details to xb core processor",
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
			name:      "ERROR: status not rfi when request get payout details to xb core processor",
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
				mockXbCoreProcessorRepo.On("GetPayoutById", mock.Anything, mock.Anything).Return(&xbCoreProcessorModel.GetPayoutResponseData{
					Status: constant.XbStatusInProcess,
				}, nil)
			},
		},
		{
			name:      "ERROR: error when request get rfi details to xb core processor",
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
			name:      "SUCCESS: success get rfi details",
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
					Status: constant.XbStatusInfoRequested,
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

			_, err := s.GetRfiDetails(context.Background(), &xbModel.GetRfiDetailsRequest{
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
