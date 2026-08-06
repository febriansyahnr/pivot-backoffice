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

func TestSubmitRfiDetails(t *testing.T) {
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
				disbursementRepo.On("FindByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(nil, nil)
			},
		},
		{
			name:    "ERROR: XbCore GetPayoutById service error",
			wantErr: true,
			setupMock: func() {
				disbursementRepo.On("FindByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Return(validDisbursement, nil)

				xbCoreProcessorRepo.On("GetPayoutById",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.GetPayoutRequest"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: status not rfi when request get payout details to xb core processor",
			wantErr: true,
			setupMock: func() {
				disbursementRepo.On("FindByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Return(validDisbursement, nil)

				xbCoreProcessorRepo.On("GetPayoutById",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.GetPayoutRequest"),
				).Once().Return(&xbCoreProcessorModel.GetPayoutResponseData{
					Status: c.XbStatusInProcess,
				}, nil)
			},
		},
		{
			name:    "ERROR: XbCore SubmitRfiDetails service error",
			wantErr: true,
			setupMock: func() {
				disbursementRepo.On("FindByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Return(validDisbursement, nil)

				xbCoreProcessorRepo.On("GetPayoutById",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.GetPayoutRequest"),
				).Once().Return(&xbCoreProcessorModel.GetPayoutResponseData{
					Status: c.XbStatusInfoRequested,
				}, nil)

				xbCoreProcessorRepo.On("SubmitRfiDetails",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.SubmitRfiDetailsRequest"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "SUCCESS",
			wantErr: false,
			setupMock: func() {
				xbCoreProcessorRepo.On("GetPayoutById",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.GetPayoutRequest"),
				).Once().Return(&xbCoreProcessorModel.GetPayoutResponseData{
					Status: c.XbStatusInfoRequested,
				}, nil)
				xbCoreProcessorRepo.On("SubmitRfiDetails",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.SubmitRfiDetailsRequest"),
				).Return(&xbCoreProcessorModel.SubmitRfiDetailsResponseData{}, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			svc := New(log, disbursementRepo, nil, xbCoreProcessorRepo)
			_, err := svc.SubmitRfiDetails(context.Background(), &xbModel.SubmitRfiDetailsRequest{})
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
