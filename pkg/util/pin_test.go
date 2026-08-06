package util

import (
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/stretchr/testify/assert"
)

func TestIsValidPin(t *testing.T) {
	type payload struct {
		pin string
	}

	testCases := []struct {
		desc    string
		payload payload
		want    error
	}{
		{
			desc:    "when PIN length is not 6, should return ErrInvalidPINLength",
			payload: payload{pin: "12345"}, // less than 6 digits
			want:    constant.ErrInvalidPINLength,
		},
		{
			desc:    "when PIN has identical digits, should return ErrIdenticalPIN",
			payload: payload{pin: "111111"}, // identical digits
			want:    constant.ErrIdenticalPIN,
		},
		{
			desc:    "when PIN has ascending sequential digits, should return ErrSequentialPIN",
			payload: payload{pin: "123456"}, // sequential digits
			want:    constant.ErrSequentialPIN,
		},
		{
			desc:    "when PIN has descending sequential digits, should return ErrSequentialPIN",
			payload: payload{pin: "654321"}, // sequential digits
			want:    constant.ErrSequentialPIN,
		},
		{
			desc:    "when PIN is valid (non-sequential, non-identical), should return nil",
			payload: payload{pin: "246809"}, // valid non-sequential digits
			want:    nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			err := IsValidPin(tc.payload.pin)
			assert.Equal(t, tc.want, err)
		})
	}
}

// You can add additional utility functions if needed
