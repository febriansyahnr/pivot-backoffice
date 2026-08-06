package xbPayoutService

import (
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/stretchr/testify/assert"
)

func TestMapStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected string
	}{
		{
			name:     "SUCCESS: map status XbStatusInProcess",
			status:   constant.XbStatusInProcess,
			expected: constant.XbDisbursementReasonTypePending,
		},
		{
			name:     "SUCCESS: map status XbStatusInfoRequested",
			status:   constant.XbStatusInfoRequested,
			expected: constant.XbDisbursementReasonTypeDocumentRequested,
		},
		{
			name:     "SUCCESS: map status XbStatusComplianceVerification",
			status:   constant.XbStatusComplianceVerification,
			expected: constant.XbDisbursementReasonTypeInReview,
		},
		{
			name:     "SUCCESS: map status XbStatusInfoResponded",
			status:   constant.XbStatusInfoResponded,
			expected: constant.XbDisbursementReasonTypeInReview,
		},
		{
			name:     "SUCCESS: map status XbStatusComplianceRejected",
			status:   constant.XbStatusComplianceRejected,
			expected: constant.XbDisbursementReasonTypePartnerRejected,
		},
		{
			name:     "SUCCESS: map status XbStatusComplianceApproved",
			status:   constant.XbStatusComplianceApproved,
			expected: constant.XbDisbursementReasonTypeInReview,
		},
		{
			name:     "SUCCESS: map status XbStatusPGProcessing",
			status:   constant.XbStatusPGProcessing,
			expected: constant.XbDisbursementReasonTypeProcessing,
		},
		{
			name:     "SUCCESS: map status XbStatusSentToBank",
			status:   constant.XbStatusSentToBank,
			expected: constant.XbDisbursementReasonTypeProcessing,
		},
		{
			name:     "SUCCESS: map status XbStatusRejected",
			status:   constant.XbStatusRejected,
			expected: constant.XbDisbursementReasonTypeBeneficiaryRejected,
		},
		{
			name:     "SUCCESS: map status XbStatusError",
			status:   constant.XbStatusError,
			expected: constant.XbDisbursementReasonTypeFailed,
		},
		{
			name:     "SUCCESS: map status XbStatusReturned",
			status:   constant.XbStatusReturned,
			expected: constant.XbDisbursementReasonTypeRefunded,
		},
		{
			name:     "SUCCESS: map status XbStatusPaid",
			status:   constant.XbStatusPaid,
			expected: constant.XbDisbursementReasonTypeSuccess,
		},
		{
			name:     "SUCCESS: map status default case",
			status:   "someUnknownStatus",
			expected: "someUnknownStatus",
		},
		{
			name:     "SUCCESS: map status XbStatusCanceled",
			status:   constant.XbStatusCanceled,
			expected: constant.XbDisbursementReasonTypeFailed,
		},
		{
			name:     "SUCCESS: map status XbStatusRemindRecipient",
			status:   constant.XbStatusRemindRecipient,
			expected: constant.XbDisbursementReasonTypePending,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := &xbPayoutService{}
			assert.Equal(t, tc.expected, s.mapStatus(tc.status))
		})
	}
}
