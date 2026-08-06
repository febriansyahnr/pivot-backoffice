package util_test

import (
	"net/http"
	"net/url"
	"testing"

	. "github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/stretchr/testify/assert"
)

func TestPostFormBoolValue(t *testing.T) {
	tests := []struct {
		req       *http.Request
		wantError bool
		wantValue bool
	}{
		{
			req: &http.Request{
				PostForm: url.Values{
					"key": {""},
				},
			},
			wantError: false,
			wantValue: false,
		},
		{
			req: &http.Request{
				PostForm: url.Values{
					"key": {"true"},
				},
			},
			wantError: false,
			wantValue: true,
		},
		{
			req: &http.Request{
				PostForm: url.Values{
					"key": {"false"},
				},
			},
			wantError: false,
			wantValue: false,
		},
		{
			req: &http.Request{
				PostForm: url.Values{
					"key": {"invalid"},
				},
			},
			wantError: true,
			wantValue: false,
		},
	}
	for _, tc := range tests {
		val, err := PostFormBoolValue(tc.req, "key")
		assert.Equal(t, tc.wantError, err != nil)
		assert.Equal(t, tc.wantValue, val)
	}
}
