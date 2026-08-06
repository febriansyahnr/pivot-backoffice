package xbPayoutService_test

import (
	"context"
	"testing"
	"time"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	xbCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xbCoreProcessor"
	repositoryMock "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/xbPayout"
)

func TestUploadUnderlyingDocument(t *testing.T) {
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})
	disbursementRepo := repositoryMock.NewIDisbursementRepository(t)
	xbCoreProcessorRepo := repositoryMock.NewIXbCoreProcessorRepository(t)

	processorReferenceId := "proc001"
	remark := "remark"
	validDisbursement := &disbursementModel.DisbursementWithTransaction{
		Disbursement: disbursementModel.Disbursement{
			ProcessorReferenceID: &processorReferenceId,
			Remark:               &remark,
			MetadataObj: disbursementModel.Metadata{
				XbDetail: &xbModel.XbPayoutMetadata{
					ExpiredAt:   time.Now().Add(2 * time.Hour),
					TotalAmount: decimal.NewFromFloat(1_000_000),
				},
			},
		},
	}

	testCases := []struct {
		name      string
		wantErr   bool
		setupMock func()
	}{
		{
			name:    "ERROR: Find payout error",
			wantErr: true,
			setupMock: func() {
				disbursementRepo.On("FindByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: Find payout not found",
			wantErr: true,
			setupMock: func() {
				disbursementRepo.On("FindByID", c.ValueCtxMockType(), c.StringMockType()).Once().Return(nil, nil)
			},
		},
		{
			name:    "ERROR: Merchant ID does not match",
			wantErr: true,
			setupMock: func() {
				disbursementRepo.On(
					"FindByID", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(&disbursementModel.DisbursementWithTransaction{
					Disbursement: disbursementModel.Disbursement{MerchantID: "123456"}, // NOSONAR
				}, nil)
			},
		},
		{
			name:    "ERROR: Payout was expired",
			wantErr: true,
			setupMock: func() {
				disbursementRepo.On("FindByID", c.ValueCtxMockType(), c.StringMockType()).Once().Return(&disbursementModel.DisbursementWithTransaction{
					Disbursement: disbursementModel.Disbursement{
						MetadataObj: disbursementModel.Metadata{
							XbDetail: &xbModel.XbPayoutMetadata{
								ExpiredAt: time.Now().Add(-time.Hour),
							},
						},
					},
				}, nil)
			},
		},
		{
			name:    "ERROR: XbCore UploadUnderlyingDocument service error",
			wantErr: true,
			setupMock: func() {
				disbursementRepo.On("FindByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Return(validDisbursement, nil)

				xbCoreProcessorRepo.On("UploadUnderlyingDocument",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.UploadUnderlyingDocumentRequest"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "SUCCESS",
			wantErr: false,
			setupMock: func() {
				xbCoreProcessorRepo.On("UploadUnderlyingDocument",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.UploadUnderlyingDocumentRequest"),
				).Return(&xbCoreProcessorModel.UploadUnderlyingDocumentResponseData{}, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			svc := New(log, disbursementRepo, nil, xbCoreProcessorRepo)
			_, err := svc.UploadUnderlyingDocument(context.Background(), &xbModel.UploadUnderlyingDocumentRequest{})
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
