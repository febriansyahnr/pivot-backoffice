package outbound_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOutboundModel(t *testing.T) {
	req := OutboundRequest{
		Id: "outbound-id",
		Client: &Client{
			RequestId: "request-id",
		},
		Method:       http.MethodGet,
		URL:          "http://localhost:3000/api",
		Headers:      map[string]string{"Header1": "Value1"},
		Body:         map[string]interface{}{"Data": 1},
		StatusCode:   http.StatusOK,
		ResponseTime: "32 ms",
		ResponseBody: []byte(`{"message":"OK"}`),
		Error:        constant.ErrSomeErrorForUnitTest,
	}

	want, err := json.Marshal(req.ToOutbound())
	require.NoError(t, err)

	expected := `{"Id":"outbound-id","Client":{"request_id":"request-id"},"Date":"0001-01-01T00:00:00Z","Method":"GET","URL":"http://localhost:3000/api","Headers":{"Header1":"Value1"},"Body":{"Data":1},"StatusCode":200,"ResponseTime":"32 ms","ResponseBody":{"message":"OK"},"Error":"some error"}`
	assert.JSONEq(t, expected, string(want))
}
