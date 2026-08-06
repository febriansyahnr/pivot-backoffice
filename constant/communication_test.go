package constant_test

import (
	"encoding/json"
	"testing"

	. "github.com/paper-indonesia/pivot-backoffice/constant"

	"github.com/stretchr/testify/assert"
)

func TestEmailPriority(t *testing.T) {
	tests := []struct {
		input EmailPriority
		want  string
	}{
		{
			input: EmailPrioritL0,
			want:  "\"0\"",
		},
		{
			input: EmailPrioritL1,
			want:  "\"1\"",
		},
		{
			input: EmailPrioritL2,
			want:  "\"2\"",
		},
	}
	for _, test := range tests {
		buf, err := json.Marshal(test.input)
		assert.Nil(t, err)
		assert.Equal(t, test.want, string(buf))
	}
}
