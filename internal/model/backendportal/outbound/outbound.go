package outbound

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"

	"github.com/jmoiron/sqlx/types"
)

type Client struct {
	RequestId      string `json:"request_id"`
	From           string `json:"from,omitempty"`
	OriginId       string `json:"origin_id,omitempty"`
	ReferenceId    string `json:"reference_id,omitempty"`
	ReplyToAddress string `json:"reply_to_address,omitempty"`
}

type OutboundRequest struct {
	Id           string
	Client       *Client
	Date         time.Time
	Method       string
	URL          string
	Headers      map[string]string
	Body         interface{}
	StatusCode   int
	ResponseTime string
	ResponseBody []byte
	Error        error
}

type Outbound struct {
	Id           string              `db:"id"`
	Client       types.JSONText      `db:"client"`
	Date         time.Time           `db:"date"`
	Method       string              `db:"method"`
	URL          string              `db:"url"`
	Headers      types.JSONText      `db:"headers"`
	Body         types.NullJSONText  `db:"body"`
	StatusCode   int                 `db:"status_code"`
	ResponseTime string              `db:"response_time"`
	ResponseBody types.NullJSONText  `db:"response_body"`
	Error        constant.NullString `db:"error_message"`
}

func (r *OutboundRequest) ToOutbound() *Outbound {

	data := &Outbound{
		Id:           r.Id,
		Date:         r.Date,
		Method:       r.Method,
		URL:          r.URL,
		StatusCode:   r.StatusCode,
		ResponseTime: r.ResponseTime,
	}
	data.Client, _ = json.Marshal(r.Client)
	data.Headers, _ = json.Marshal(r.Headers)

	if r.Body != nil {
		data.Body.Valid = true
		data.Body.JSONText, _ = json.Marshal(r.Body)
	}
	if r.ResponseBody != nil {
		data.ResponseBody.Valid = true
		_ = data.ResponseBody.JSONText.Scan(r.ResponseBody)
	}
	if r.Error != nil {
		data.Error = constant.NullString{
			NullString: sql.NullString{
				Valid:  true,
				String: r.Error.Error(),
			}}
	}
	return data
}
