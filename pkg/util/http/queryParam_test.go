package httputil

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetDateTimeQueryParam(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		key      string
		layout   string
		expected time.Time
		found    bool
	}{
		{
			name:     "Valid datetime with timezone",
			query:    "datetime=2024-10-13T07:19:01.638Z",
			key:      "datetime",
			layout:   time.RFC3339,
			expected: time.Date(2024, 10, 13, 07, 19, 01, 638000000, time.UTC),
			found:    true,
		},
		{
			name:     "Valid datetime without timezone",
			query:    "datetime=2024-10-13T07:19:01",
			key:      "datetime",
			layout:   "2006-01-02T15:04:05",
			expected: time.Date(2024, 10, 13, 07, 19, 01, 0, time.UTC),
			found:    true,
		},
		{
			name:     "Invalid datetime",
			query:    "datetime=invalid",
			key:      "datetime",
			layout:   time.RFC3339,
			expected: time.Time{},
			found:    false,
		},
		{
			name:     "Missing datetime",
			query:    "",
			key:      "datetime",
			layout:   time.RFC3339,
			expected: time.Time{},
			found:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &http.Request{
				URL: &url.URL{
					RawQuery: tt.query,
				},
			}
			result, found := GetDateTimeQueryParam(req, tt.key, tt.layout)
			assert.Equal(t, tt.found, found)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetQueryParam(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		key      string
		expected string
		found    bool
	}{
		{
			name:     "Parameter found",
			query:    "param=value",
			key:      "param",
			expected: "value",
			found:    true,
		},
		{
			name:     "Parameter not found",
			query:    "param=value",
			key:      "other",
			expected: "",
			found:    false,
		},
		{
			name:     "Empty parameter value",
			query:    "param=",
			key:      "param",
			expected: "",
			found:    false,
		},
		{
			name:     "Multiple parameters",
			query:    "param1=value1&param2=value2",
			key:      "param2",
			expected: "value2",
			found:    true,
		},
		{
			name:     "No query parameters",
			query:    "",
			key:      "param",
			expected: "",
			found:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &http.Request{
				URL: &url.URL{
					RawQuery: tt.query,
				},
			}
			result, found := GetQueryParam(req, tt.key)
			assert.Equal(t, tt.found, found)
			assert.Equal(t, tt.expected, result)
		})
	}
}
