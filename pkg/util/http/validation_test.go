package httputil_test

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	. "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"

	"github.com/stretchr/testify/require"
)

func TestValidateReportDateRangeFromRequest(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name      string
		request   any
		validate  func(any) error
		wantError string
	}{
		{
			name: "ERROR:Invalid key pair",
			validate: func(r any) error {
				return ValidateReportDateRangeFromRequest(r, "startDate")
			},
			wantError: "parameters must consist of paired initial and final values",
		},
		{
			name:    "ERROR:Invalid decode request",
			request: []string{},
			validate: func(r any) error {
				return ValidateReportDateRangeFromRequest(r, "startDate", "endDate")
			},
			wantError: "json unmarshal request: json: cannot unmarshal array into Go value of type map[string]interface {}",
		},
		{
			name: "ERROR:Invalid start value",
			request: &http.Request{
				URL: &url.URL{
					RawQuery: "startDate=X&endDate=Y",
				},
			},
			validate: func(r any) error {
				return ValidateReportDateRangeFromRequest(r, "startDate", "endDate")
			},
			wantError: `Key: startDate Value: X Error: Value format must be yyyy-mm-ddThh:nn:ssZ`,
		},
		{
			name: "ERROR:Invalid end value",
			request: &http.Request{
				URL: &url.URL{
					RawQuery: "startDate=2025-01-01&endDate=Y",
				},
			},
			validate: func(r any) error {
				return ValidateReportDateRangeFromRequest(r, "startDate", "endDate")
			},
			wantError: `Key: endDate Value: Y Error: Value format must be yyyy-mm-ddThh:nn:ssZ`,
		},
		{
			name: "ERROR:Start date is greater than end date",
			request: &http.Request{
				URL: &url.URL{
					RawQuery: fmt.Sprintf("startDate=%s&endDate=%s", now.Add(time.Hour).Format(time.RFC3339), now.Format(time.RFC3339)),
				},
			},
			validate: func(r any) error {
				return ValidateReportDateRangeFromRequest(r, "startDate", "endDate")
			},
			wantError: `startDate must not be greater than endDate`,
		},
		{
			name: "ERROR:Exceeding the maximum backdate limit",
			request: &http.Request{
				URL: &url.URL{
					RawQuery: fmt.Sprintf("startDate=%s&endDate=%s", now.AddDate(0, -7, 0).Format(time.RFC3339), now.AddDate(0, -6, 0).Format(time.RFC3339)),
				},
			},
			validate: func(r any) error {
				return ValidateReportDateRangeFromRequest(r, "startDate", "endDate")
			},
			wantError: `The date range exceeds the allowed backdate limit. Maximum allowed is the last 6 months.`,
		},
		{
			name: "ERROR:Exceeding the maximum backdate limit from request body",
			request: map[string]time.Time{
				"paymentStartDate": now.AddDate(0, -7, 0),
				"paymentEndDate":   now.AddDate(0, -6, 0),
			},
			validate: func(r any) error {
				return ValidateReportDateRangeFromRequest(r, "paymentStartDate", "paymentEndDate")
			},
			wantError: `The date range exceeds the allowed backdate limit. Maximum allowed is the last 6 months.`,
		},
		{
			name: "ERROR:Exceeding the date range limit",
			request: &http.Request{
				URL: &url.URL{
					RawQuery: fmt.Sprintf("startDate=%s&endDate=%s", now.AddDate(0, -2, 0).Format(time.RFC3339), now.Format(time.RFC3339)),
				},
			},
			validate: func(r any) error {
				return ValidateReportDateRangeFromRequest(r, "startDate", "endDate")
			},
			wantError: `The date range exceeds the allowed limit. Maximum permitted is 31 days.`,
		},
		{
			name: "ERROR:Exceeding the date range limit from request body",
			request: map[string]string{
				"startDate": now.AddDate(0, -2, 0).Format(time.RFC3339),
				"endDate":   now.Format(time.RFC3339),
			},
			validate: func(r any) error {
				return ValidateReportDateRangeFromRequest(r, "startDate", "endDate")
			},
			wantError: `The date range exceeds the allowed limit. Maximum permitted is 31 days.`,
		},
		{
			name: "SUCCESS:Params not found",
			request: &http.Request{
				URL: &url.URL{
					RawQuery: fmt.Sprintf("startDate=%s&endDate=%s", now.AddDate(0, -2, 0).Format(time.RFC3339), now.Format(time.RFC3339)),
				},
			},
			validate: func(r any) error {
				return ValidateReportDateRangeFromRequest(r, "paymentStartDate", "paymentEndDate")
			},
			wantError: "",
		},
		{
			name: "SUCCESS:Valid date range",
			request: &http.Request{
				URL: &url.URL{
					RawQuery: fmt.Sprintf("startDate=%s&endDate=%s", now.AddDate(0, 0, -31).Format(time.RFC3339), now.Format(time.RFC3339)),
				},
			},
			validate: func(r any) error {
				return ValidateReportDateRangeFromRequest(r, "startDate", "endDate", "paymentStartDate", "paymentEndDate")
			},
			wantError: "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.validate(test.request)
			if test.wantError == "" {
				require.Nil(t, err)

			} else {
				require.NotNil(t, err)
				require.Equal(t, test.wantError, err.Error())
			}
		})
	}
}
