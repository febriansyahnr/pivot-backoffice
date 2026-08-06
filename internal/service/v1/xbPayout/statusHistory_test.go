package xbPayoutService

import (
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	statusHistoryModel "github.com/paper-indonesia/pivot-backoffice/internal/model/statusHistory"
	logMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	repoMock "github.com/paper-indonesia/pivot-backoffice/mocks/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRecordStatusHistory(t *testing.T) {
	logger := logMock.NewILogger(t)
	historyRepo := repoMock.NewIStatusHistoriesRepository(t)

	service := &xbPayoutService{
		logger:              logger,
		statusHistoriesRepo: historyRepo,
	}

	tests := []struct {
		name      string
		request   *statusHistoryModel.RecordDisbursementStatusHistoryRequest
		setupMock func()
		wantError error
	}{
		{
			name:      "Payout History - Skipped",
			setupMock: func() { /* Empty Function */ },
		},
		{
			name:    "Payout History - Insert status history",
			request: &statusHistoryModel.RecordDisbursementStatusHistoryRequest{},
			setupMock: func() {
				historyRepo.On("Insert", mock.Anything, mock.Anything).Once().Return(assert.AnError)
				logger.On("Error", mock.Anything, "Failed to insert status history", mock.Anything, mock.Anything).Once().Return()
			},
			wantError: assert.AnError,
		},
		{
			name: "Payout History - Payout created", // NOSONAR
			request: &statusHistoryModel.RecordDisbursementStatusHistoryRequest{
				Status: constant.XbStatusCreated,
			},
			setupMock: func() {
				historyRepo.On("Insert", mock.Anything, mock.MatchedBy(func(req *statusHistoryModel.StatusHistory) bool {
					return req.Status == constant.XbStatusCreated &&
						assert.JSONEq(t, `{"label":"Payout Created", "description":"Waiting for merchant to confirm payout.", "actor":""}`, string(req.Metadata.JSONText))
				})).Once().Return(nil)
			},
		},
		{
			name: "Payout History - Payout Confirmed", // NOSONAR
			request: &statusHistoryModel.RecordDisbursementStatusHistoryRequest{
				Status: constant.XbStatusConfirmed,
			},
			setupMock: func() {
				historyRepo.On("Insert", mock.Anything, mock.MatchedBy(func(req *statusHistoryModel.StatusHistory) bool {
					return req.Status == constant.XbStatusConfirmed &&
						assert.JSONEq(t, `{"label":"Payout Created", "description":"Payout request confirmed.", "actor":""}`, string(req.Metadata.JSONText))
				})).Once().Return(nil)
			},
		},
		{
			name: "Payout History - Information Requested", // NOSONAR
			request: &statusHistoryModel.RecordDisbursementStatusHistoryRequest{
				Status: constant.XbStatusInfoRequested,
			},
			setupMock: func() {
				historyRepo.On("Insert", mock.Anything, mock.MatchedBy(func(req *statusHistoryModel.StatusHistory) bool {
					return req.Status == constant.XbStatusInfoRequested &&
						assert.JSONEq(t, `{"label":"Information Requested", "description":"Further information requested by bank partner.", "actor":"", "recommendation":"Please submit supporting information to Helpdesk."}`, string(req.Metadata.JSONText))
				})).Once().Return(nil)
			},
		},
		{
			name: "Payout History - Information In Review", // NOSONAR
			request: &statusHistoryModel.RecordDisbursementStatusHistoryRequest{
				Status: constant.XbStatusComplianceApproved,
			},
			setupMock: func() {
				historyRepo.On("Insert", mock.Anything, mock.MatchedBy(func(req *statusHistoryModel.StatusHistory) bool {
					return req.Status == constant.XbDisbursementReasonTypeInReview &&
						assert.JSONEq(t, `{"label":"Information In Review", "description":"Information submitted and in review by bank partner.", "actor":""}`, string(req.Metadata.JSONText))
				})).Once().Return(nil)
			},
		},
		{
			name: "Payout History - Processing", // NOSONAR
			request: &statusHistoryModel.RecordDisbursementStatusHistoryRequest{
				Status: constant.XbStatusSentToBank,
			},
			setupMock: func() {
				historyRepo.On("Insert", mock.Anything, mock.MatchedBy(func(req *statusHistoryModel.StatusHistory) bool {
					return req.Status == constant.XbDisbursementReasonTypeProcessing &&
						assert.JSONEq(t, `{"label":"Processing", "description":"Transaction is being processed by our bank partner.", "actor":""}`, string(req.Metadata.JSONText))
				})).Once().Return(nil)
			},
		},
		{
			name: "Payout History - Rejected", // NOSONAR
			request: &statusHistoryModel.RecordDisbursementStatusHistoryRequest{
				Status: constant.XbStatusRejected,
			},
			setupMock: func() {
				historyRepo.On("Insert", mock.Anything, mock.MatchedBy(func(req *statusHistoryModel.StatusHistory) bool {
					return req.Status == constant.XbStatusRejected &&
						assert.JSONEq(t, `{"label":"Rejected", "description":"Transaction rejected by beneficiary.", "actor":""}`, string(req.Metadata.JSONText))
				})).Once().Return(nil)
			},
		},
		{
			name: "Payout History - Compliance rejected", // NOSONAR
			request: &statusHistoryModel.RecordDisbursementStatusHistoryRequest{
				Status: constant.XbStatusComplianceRejected,
			},
			setupMock: func() {
				historyRepo.On("Insert", mock.Anything, mock.MatchedBy(func(req *statusHistoryModel.StatusHistory) bool {
					return req.Status == constant.XbStatusComplianceRejected &&
						assert.JSONEq(t, `{"label":"Rejected", "description":"Transaction rejected by compliance.", "actor":""}`, string(req.Metadata.JSONText))
				})).Once().Return(nil)
			},
		},
		{
			name: "Payout History - Error (from XB Core Processor)", // NOSONAR
			request: &statusHistoryModel.RecordDisbursementStatusHistoryRequest{
				Status: constant.XbStatusError,
			},
			setupMock: func() {
				historyRepo.On("Insert", mock.Anything, mock.MatchedBy(func(req *statusHistoryModel.StatusHistory) bool {
					return req.Status == constant.XbDisbursementReasonTypeFailed &&
						assert.JSONEq(t, `{"label":"Failed", "description":"Transaction failed due to error from provider.", "actor":""}`, string(req.Metadata.JSONText))
				})).Once().Return(nil)
			},
		},
		{
			name: "Payout History - HTTP Error (confirmation failure)", // NOSONAR
			request: &statusHistoryModel.RecordDisbursementStatusHistoryRequest{
				Status: constant.XbStatusHttpError,
			},
			setupMock: func() {
				historyRepo.On("Insert", mock.Anything, mock.MatchedBy(func(req *statusHistoryModel.StatusHistory) bool {
					return req.Status == constant.XbDisbursementReasonTypeError &&
						assert.JSONEq(t, `{"label":"Error", "description":"HTTP confirmation error. Transaction status unknown.", "actor":""}`, string(req.Metadata.JSONText))
				})).Once().Return(nil)
			},
		},
		{
			name: "Payout History - Paid", // NOSONAR
			request: &statusHistoryModel.RecordDisbursementStatusHistoryRequest{
				Status: constant.XbStatusPaid,
			},
			setupMock: func() {
				historyRepo.On("Insert", mock.Anything, mock.MatchedBy(func(req *statusHistoryModel.StatusHistory) bool {
					return req.Status == constant.XbStatusPaid &&
						assert.JSONEq(t, `{"label":"Success", "description":"Transaction has been successfully completed.", "actor":""}`, string(req.Metadata.JSONText))
				})).Once().Return(nil)
			},
		},
		{
			name: "Payout History - Refunded", // NOSONAR
			request: &statusHistoryModel.RecordDisbursementStatusHistoryRequest{
				Status: constant.XbStatusReturned,
			},
			setupMock: func() {
				historyRepo.On("Insert", mock.Anything, mock.MatchedBy(func(req *statusHistoryModel.StatusHistory) bool {
					return req.Status == constant.XbDisbursementReasonTypeRefunded &&
						assert.JSONEq(t, `{"label":"Refunded", "description":"Transaction has been refunded.", "actor":""}`, string(req.Metadata.JSONText))
				})).Once().Return(nil)
			},
		},
		{
			name: "Payout History - Canceled", // NOSONAR
			request: &statusHistoryModel.RecordDisbursementStatusHistoryRequest{
				Status: constant.XbStatusCanceled,
			},
			setupMock: func() {
				historyRepo.On("Insert", mock.Anything, mock.MatchedBy(func(req *statusHistoryModel.StatusHistory) bool {
					return req.Status == constant.XbStatusCanceled &&
						assert.JSONEq(t, `{"label":"Canceled", "description":"Transaction has been cancelled.", "actor":""}`, string(req.Metadata.JSONText))
				})).Once().Return(nil)
			},
		},
		{
			name: "Payout History - Expired", // NOSONAR
			request: &statusHistoryModel.RecordDisbursementStatusHistoryRequest{
				Status: constant.XbStatusExpired,
			},
			setupMock: func() {
				historyRepo.On("Insert", mock.Anything, mock.MatchedBy(func(req *statusHistoryModel.StatusHistory) bool {
					return req.Status == constant.XbStatusExpired &&
						assert.JSONEq(t, `{"label":"Expired", "description":"Transaction has been expired.", "actor":""}`, string(req.Metadata.JSONText))
				})).Once().Return(nil)
			},
		},
		{
			name: "Payout History - Others", // NOSONAR
			request: &statusHistoryModel.RecordDisbursementStatusHistoryRequest{
				Status: "OTHERS",
			},
			setupMock: func() {
				historyRepo.On("Insert", mock.Anything, mock.MatchedBy(func(req *statusHistoryModel.StatusHistory) bool {
					return req.Status == "OTHERS" &&
						assert.JSONEq(t, `{"label":"Others", "description":"Transaction status changed.", "actor":""}`, string(req.Metadata.JSONText))
				})).Once().Return(nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()
			assert.Equal(t, test.wantError, service.RecordStatusHistory(t.Context(), test.request))

			logger.AssertExpectations(t)
			historyRepo.AssertExpectations(t)
		})
	}
}
