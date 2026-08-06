package util

import (
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/stretchr/testify/assert"
)

func TestMapAccountInquirySnapResponseToDetailStatus(t *testing.T) {
	testCases := []struct {
		name         string
		responseCode string
		expected     string
	}{
		{
			name:         "Dormant account pattern",
			responseCode: "403XX09",
			expected:     constant.ReqInquiryDetailStatusDormant,
		},
		{
			name:         "Inactive account pattern", 
			responseCode: "403XX18",
			expected:     constant.ReqInquiryDetailStatusInactive,
		},
		{
			name:         "Suspected fraud pattern",
			responseCode: "403XX03",
			expected:     constant.ReqInquiryDetailStatusSuspectedFraud,
		},
		{
			name:         "Activity count limit exceeded pattern",
			responseCode: "403XX04",
			expected:     constant.ReqInquiryDetailStatusLimitExceeded,
		},
		{
			name:         "Do not honor pattern",
			responseCode: "403XX05",
			expected:     constant.ReqInquiryDetailStatusDoNotHonor,
		},
		{
			name:         "Invalid account pattern",
			responseCode: "404XX11",
			expected:     constant.ReqInquiryDetailStatusInvalid,
		},
		{
			name:         "Unknown response code defaults to invalid",
			responseCode: "9999999",
			expected:     constant.ReqInquiryDetailStatusInvalid,
		},
		{
			name:         "Empty response code defaults to invalid",
			responseCode: "",
			expected:     constant.ReqInquiryDetailStatusInvalid,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := MapAccountInquirySnapResponseToDetailStatus(tc.responseCode)
			assert.Equal(t, tc.expected, result)
		})
	}
}