package util_test

import (
	"testing"

	. "github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/stretchr/testify/assert"
)

func TestJSONBind(t *testing.T) {
	type data struct {
		Message string `json:"message"`
	}

	tests := []struct {
		dst        any
		src        any
		wantError  string
		wantResult any
	}{
		{
			dst:       nil,
			wantError: "destination pointer is nil - cannot bind source input",
		},
		{
			src:       make(chan any, 1),
			dst:       &data{},
			wantError: "failed to marshal input data",
		},
		{
			src: map[string]string{"message": "Hello World!"},
			dst: &data{},
			wantResult: &data{
				Message: "Hello World!",
			},
		},
		{
			src: []byte(`{"message": "Hi, John!"}`),
			dst: &data{},
			wantResult: &data{
				Message: "Hi, John!",
			},
		},
		{
			src: `{"message": "Hi, Soleh!"}`,
			dst: &data{},
			wantResult: &data{
				Message: "Hi, Soleh!",
			},
		},
	}
	for _, test := range tests {

		err := JSONBind(test.dst, test.src)

		if test.wantError == "" {
			assert.NoError(t, err)
			assert.Equal(t, test.wantResult, test.dst)

		} else if assert.Error(t, err) {
			assert.Contains(t, err.Error(), test.wantError)
		}
	}
}
