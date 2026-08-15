package inboundModel

import (
	"encoding/json"
	"github.com/jmoiron/sqlx/types"
	"time"
)

type InboundResponse struct {
	ID                string             `json:"id"`
	ReferenceID       string             `json:"referenceId"`
	OriginID          string             `json:"originId"`
	TraceID           string             `json:"traceId"`
	IP                string             `json:"ip"`
	Method            string             `json:"method"`
	URL               string             `json:"url"`
	Headers           types.JSONText     `json:"headers"`
	Body              types.NullJSONText `json:"body"`
	StatusCode        int                `json:"statusCode"`
	ResponseTimeMs    float64            `json:"responseTimeMs"`
	ResponseBody      types.NullJSONText `json:"responseBody"`
	Metadata          types.NullJSONText `json:"metadata"`
	SnapCompatibility bool               `json:"snapCompatibility"`
	CreatedAt         time.Time          `json:"createdAt"`
	UpdatedAt         time.Time          `json:"updatedAt"`

	Client  *Client `json:"-"`
	Feature string  `json:"feature"`
}

func (i *Inbound) ToResponse() *InboundResponse {
	resp := &InboundResponse{
		ID:                i.ID,
		ReferenceID:       i.ReferenceID,
		OriginID:          i.OriginID,
		TraceID:           i.TraceID,
		IP:                i.IP,
		Method:            i.Method,
		URL:               i.URL,
		Headers:           i.Headers,
		Body:              i.Body,
		StatusCode:        i.StatusCode,
		ResponseTimeMs:    i.ResponseTimeMs,
		ResponseBody:      i.ResponseBody,
		Metadata:          i.Metadata,
		SnapCompatibility: i.SnapCompatibility,
		CreatedAt:         i.CreatedAt,
		UpdatedAt:         i.UpdatedAt,
		Client:            &Client{},
	}

	_ = json.Unmarshal(i.Client, &resp.Client)
	resp.Feature = resp.Client.Feature

	return resp
}
