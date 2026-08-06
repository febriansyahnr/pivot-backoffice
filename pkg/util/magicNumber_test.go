package util

import (
	"net/http"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
)

func TestValidateMagicNumber(t *testing.T) {
	testCases := []struct {
		name               string
		beneficiaryAccount string
		method             string
		expectedError      error
	}{
		{
			name:               "SUCCESS: should return nil",
			method:             http.MethodPost,
			beneficiaryAccount: "999966660001",
			expectedError:      nil,
		},
		{
			name:               "ERROR: should return error constant.ErrDuplicateDisbursementReferenceId",
			method:             http.MethodPost,
			beneficiaryAccount: "999966660004",
			expectedError:      constant.ErrDuplicateDisbursementReferenceId,
		},
		{
			name:               "ERROR: should return error constant.ErrInsufficientBalance",
			method:             http.MethodGet,
			beneficiaryAccount: "999966660007",
			expectedError:      constant.ErrInsufficientBalance,
		},
		{
			name:               "ERROR: should return error constant.ErrTimeout",
			method:             http.MethodGet,
			beneficiaryAccount: "999966660008",
			expectedError:      constant.ErrTimeout,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ValidateMagicNumber(tc.method, tc.beneficiaryAccount)
			if err != tc.expectedError {
				t.Errorf("expected error: %v, got: %v", tc.expectedError, err)
			}
		})
	}
}
