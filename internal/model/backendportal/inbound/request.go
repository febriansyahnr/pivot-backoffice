package inboundModel

import (
	"encoding/json"
	"net/http"
	"regexp"
	"time"

	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/unifiedPayment"
)

type InboundRequest struct {
	ID                string
	Client            *Client
	IP                string
	Method            string
	URL               string
	Headers           map[string][]string
	Body              interface{}
	StatusCode        int
	ResponseTimeMs    float64
	ResponseBody      []byte
	SnapCompatibility bool
	Metadata          map[string]interface{}
}

type Client struct {
	Feature     string `json:"feature"`
	TraceId     string `json:"trace_id"`
	OriginId    string `json:"origin_id,omitempty"`
	ReferenceId string `json:"reference_id,omitempty"`
}

func (r *InboundRequest) ToInbound() *Inbound {
	r.SetSnapCompatible()

	data := &Inbound{
		ID:                r.ID,
		IP:                r.IP,
		Method:            r.Method,
		URL:               r.URL,
		StatusCode:        r.StatusCode,
		ResponseTimeMs:    r.ResponseTimeMs,
		SnapCompatibility: r.SnapCompatibility,
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}
	data.Client, _ = json.Marshal(r.Client)
	data.Headers, _ = json.Marshal(r.Headers)

	if r.Body != nil {
		data.Body.Valid = true
		data.Body.JSONText, _ = json.Marshal(r.Body)
	}
	if r.Metadata != nil {
		data.Metadata.Valid = true
		data.Metadata.JSONText, _ = json.Marshal(r.Metadata)
	}
	if r.ResponseBody != nil {
		data.ResponseBody.Valid = true
		_ = data.ResponseBody.JSONText.Scan(r.ResponseBody)
	}

	return data
}

func (r *InboundRequest) SetSnapCompatible() {
	if r.URL == "/internal/v1/access-token/b2b" {
		r.SnapCompatibility = true
		return
	}

	isPaymentRequest := r.Method == http.MethodPost && r.URL == "/open-api/v2/payments"
	isPaymentConfirm := regexp.MustCompile(`^/open-api/v2/payments/[a-f0-9-]+/confirm$`).MatchString(r.URL)

	if !(isPaymentRequest || isPaymentConfirm) {
		return
	}

	var temp struct {
		Data json.RawMessage `json:"data"`
	}
	if json.Unmarshal(r.ResponseBody, &temp) != nil {
		return
	}

	var unifiedResponse unifiedPaymentModel.UnifiedPaymentSessionResponse
	if json.Unmarshal(temp.Data, &unifiedResponse) != nil || len(unifiedResponse.ChargeDetails) == 0 {
		return
	}

	r.SnapCompatibility = true
}

type GetInboundFilterRequest struct {
	MerchantID     string     `json:"-"`
	OriginID       string     `json:"-"`
	StartCreatedAt *time.Time `json:"startCreatedAt"`
	EndCreatedAt   *time.Time `json:"endCreatedAt"`
	Status         string     `json:"status" example:"SUCCESS | FAILED | ERROR | REDIRECT | REDIRECTION"`
	Method         string     `json:"method"`
	Product        string     `json:"product"`

	Page    int64 `json:"-"`
	PerPage int64 `json:"-"`
}
