package xbPayoutService_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	xbCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xbCoreProcessor"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/xbPayout"
	repositoryMock "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestReConfirm(t *testing.T) {
	cfg := &config.Config{
		Environment: "test",
		XbCoreProcessorConfig: config.XbCoreProcessorConfig{
			ExtendedExpireAt: 5,
		},
	}
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})
	disbursementRepo := repositoryMock.NewIDisbursementRepository(t)
	xbCoreProcessorRepo := repositoryMock.NewIXbCoreProcessorRepository(t)

	payoutID := uuid.NewString()
	merchantID := uuid.NewString()
	processorReferenceID := "proc001"
	xbPayoutUUID := uuid.NewString()
	reasonTypeInsufficientBalance := c.XbDisbursementReasonTypeInsufficientBalance

	validDisbursement := &disbursementModel.DisbursementWithTransaction{
		Disbursement: disbursementModel.Disbursement{
			UUID:                 payoutID,
			MerchantID:           merchantID,
			ProcessorReferenceID: &processorReferenceID,
			MetadataObj: disbursementModel.Metadata{
				XbDetail: &xbModel.XbPayoutMetadata{
					Uuid: xbPayoutUUID,
				},
			},
		},
	}

	disbursementWithInsufficientBalance := &disbursementModel.DisbursementWithTransaction{
		Disbursement: disbursementModel.Disbursement{
			UUID:                 payoutID,
			MerchantID:           merchantID,
			ProcessorReferenceID: &processorReferenceID,
			ReasonType:           &reasonTypeInsufficientBalance,
			MetadataObj: disbursementModel.Metadata{
				XbDetail: &xbModel.XbPayoutMetadata{
					Uuid: xbPayoutUUID,
				},
			},
		},
	}

	testCases := []struct {
		name                    string
		request                 *xbModel.ConfirmPayoutRequest
		wantErr                 bool
		expectedNeedAutoConfirm bool
		expectedMerchantID      string
		setupMock               func()
	}{
		{
			name: "ERROR: Find disbursement by ID returns error",
			request: &xbModel.ConfirmPayoutRequest{
				PayoutId: payoutID,
			},
			wantErr: true,
			setupMock: func() {
				disbursementRepo.On("FindByID",
					c.ValueCtxMockType(),
					payoutID,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name: "ERROR: Disbursement not found",
			request: &xbModel.ConfirmPayoutRequest{
				PayoutId: payoutID,
			},
			wantErr: true,
			setupMock: func() {
				disbursementRepo.On("FindByID",
					c.ValueCtxMockType(),
					payoutID,
				).Once().Return(nil, nil)
			},
		},
		{
			name: "ERROR: ReConfirmPayout returns error",
			request: &xbModel.ConfirmPayoutRequest{
				PayoutId: payoutID,
			},
			wantErr: true,
			setupMock: func() {
				disbursementRepo.On("FindByID",
					c.ValueCtxMockType(),
					payoutID,
				).Once().Return(validDisbursement, nil)

				xbCoreProcessorRepo.On("ReConfirmPayout",
					c.ValueCtxMockType(),
					mock.MatchedBy(func(req *xbCoreProcessorModel.ConfirmPayoutRequest) bool {
						return req.XbPayoutId == xbPayoutUUID &&
							req.MerchantId == merchantID &&
							req.AcquirerTransactionId == processorReferenceID
					}),
				).Once().Return(xbCoreProcessorModel.ReConfirmPayoutResponse{}, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name: "ERROR: ReconfirmXB returns error",
			request: &xbModel.ConfirmPayoutRequest{
				PayoutId: payoutID,
			},
			wantErr: true,
			setupMock: func() {
				disbursementRepo.On("FindByID",
					c.ValueCtxMockType(),
					payoutID,
				).Once().Return(validDisbursement, nil)

				xbCoreProcessorRepo.On("ReConfirmPayout",
					c.ValueCtxMockType(),
					mock.MatchedBy(func(req *xbCoreProcessorModel.ConfirmPayoutRequest) bool {
						return req.XbPayoutId == xbPayoutUUID &&
							req.MerchantId == merchantID &&
							req.AcquirerTransactionId == processorReferenceID
					}),
				).Once().Return(xbCoreProcessorModel.ReConfirmPayoutResponse{}, nil)

				disbursementRepo.On("ReconfirmXB",
					c.ValueCtxMockType(),
					mock.MatchedBy(func(req *disbursementModel.ReconfirmXBRequest) bool {
						return req.PayoutId == payoutID &&
							req.XBStatus == c.XbStatusWaiting &&
							req.ExtendedTime.After(time.Now())
					}),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name: "SUCCESS: ReConfirm payout with insufficient balance",
			request: &xbModel.ConfirmPayoutRequest{
				PayoutId: payoutID,
			},
			wantErr:                 false,
			expectedNeedAutoConfirm: true,
			expectedMerchantID:      merchantID,
			setupMock: func() {
				disbursementRepo.On("FindByID",
					c.ValueCtxMockType(),
					payoutID,
				).Once().Return(validDisbursement, nil)

				xbCoreProcessorRepo.On("ReConfirmPayout",
					c.ValueCtxMockType(),
					mock.MatchedBy(func(req *xbCoreProcessorModel.ConfirmPayoutRequest) bool {
						return req.XbPayoutId == xbPayoutUUID &&
							req.MerchantId == merchantID &&
							req.AcquirerTransactionId == processorReferenceID
					}),
				).Once().Return(xbCoreProcessorModel.ReConfirmPayoutResponse{}, nil)

				disbursementRepo.On("ReconfirmXB",
					c.ValueCtxMockType(),
					mock.MatchedBy(func(req *disbursementModel.ReconfirmXBRequest) bool {
						return req.PayoutId == payoutID &&
							req.XBStatus == c.XbStatusWaiting &&
							req.ExtendedTime.After(time.Now())
					}),
				).Once().Return(nil)
			},
		},
		{
			name: "SUCCESS: ReConfirm payout with insufficient balance (no need auto confirm)",
			request: &xbModel.ConfirmPayoutRequest{
				PayoutId: payoutID,
			},
			wantErr:                 false,
			expectedNeedAutoConfirm: false,
			expectedMerchantID:      merchantID,
			setupMock: func() {
				disbursementRepo.On("FindByID",
					c.ValueCtxMockType(),
					payoutID,
				).Once().Return(disbursementWithInsufficientBalance, nil)

				xbCoreProcessorRepo.On("ReConfirmPayout",
					c.ValueCtxMockType(),
					mock.MatchedBy(func(req *xbCoreProcessorModel.ConfirmPayoutRequest) bool {
						return req.XbPayoutId == xbPayoutUUID &&
							req.MerchantId == merchantID &&
							req.AcquirerTransactionId == processorReferenceID
					}),
				).Once().Return(xbCoreProcessorModel.ReConfirmPayoutResponse{}, nil)

				disbursementRepo.On("ReconfirmXB",
					c.ValueCtxMockType(),
					mock.MatchedBy(func(req *disbursementModel.ReconfirmXBRequest) bool {
						return req.PayoutId == payoutID &&
							req.XBStatus == c.XbStatusWaiting &&
							req.ExtendedTime.After(time.Now())
					}),
				).Once().Return(nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			svc := New(log, disbursementRepo, nil, xbCoreProcessorRepo,
				WithConfig(cfg),
			)

			result, err := svc.ReConfirm(context.Background(), tc.request)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tc.expectedNeedAutoConfirm, result.NeedAutoConfirm)
				assert.Equal(t, tc.expectedMerchantID, result.MerchantID)
			}
		})
	}
}
