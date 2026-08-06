package constant_test

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"

	"github.com/stretchr/testify/assert"
)

func TestTypes(t *testing.T) {
	tests := []struct {
		input json.Marshaler
		want  string
	}{
		{
			input: &constant.NullString{sql.NullString{Valid: false}},
			want:  "null",
		},
		{
			input: &constant.NullString{sql.NullString{Valid: true, String: "Hello World"}},
			want:  "\"Hello World\"",
		},
		{
			input: &constant.NullTime{sql.NullTime{Valid: false}},
			want:  "null",
		},
		{
			input: &constant.NullTime{sql.NullTime{Valid: true, Time: time.Date(2024, 6, 13, 0, 0, 0, 0, time.UTC)}},
			want:  "\"2024-06-13T00:00:00Z\"",
		},
	}
	for _, test := range tests {

		buf, err := test.input.MarshalJSON()

		assert.NoError(t, err)
		assert.Equal(t, test.want, string(buf))
	}
}
