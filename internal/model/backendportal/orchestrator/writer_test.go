package orchestrator_model_test

import (
	"bytes"
	"net/http/httptest"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"

	"github.com/stretchr/testify/assert"
)

func TestWriteHeader(t *testing.T) {
	rec := httptest.NewRecorder()

	wr := FileWriter{Writer: rec}
	wr.WriteHeader("test.xlsx")

	headers := map[string]string{
		c.HeaderContentType:         "application/octet-stream",
		"Content-Disposition":       "attachment; filename=test.xlsx",
		"Content-Transfer-Encoding": "binary",
		"Expires":                   "0",
	}
	for key, want := range headers {
		assert.Equal(t, want, rec.Header().Get(key))
	}

	buff := new(bytes.Buffer)

	wr2 := FileWriter{Writer: buff}
	wr2.WriteHeader("xxxxx.xlsx")
	assert.Empty(t, buff.Bytes())
}
