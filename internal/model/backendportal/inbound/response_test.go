package inboundModel_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/jmoiron/sqlx/types"
	"github.com/stretchr/testify/assert"

	inboundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/inbound"
)

func TestToResponse(t *testing.T) {
	var client = inboundModel.Client{
		Feature:     "payment",
		TraceId:     "trace-123",
		OriginId:    "origin-456",
		ReferenceId: "ref-789",
	}
	clientJSON, _ := json.Marshal(client)

	headers := types.JSONText(`{"Content-Type": ["application/json"]}`)
	body := types.NullJSONText{
		Valid:    true,
		JSONText: types.JSONText(`{"amount":1000,"currency":"USD"}`),
	}
	metadata := types.NullJSONText{
		Valid:    true,
		JSONText: types.JSONText(`{"platform":"web"}`),
	}
	responseBody := types.NullJSONText{
		Valid:    true,
		JSONText: types.JSONText(`{"status":"success"}`),
	}

	createdAt := time.Now().UTC()
	updatedAt := createdAt

	inbound := &inboundModel.Inbound{
		ID:                "req-123",
		ReferenceID:       "ref-789",
		OriginID:          "origin-456",
		TraceID:           "trace-123",
		IP:                "192.168.1.1",
		Method:            "POST",
		URL:               "https://api.example.com/pay",
		Headers:           headers,
		Body:              body,
		StatusCode:        200,
		ResponseTimeMs:    123.45,
		ResponseBody:      responseBody,
		Metadata:          metadata,
		SnapCompatibility: true,
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
		Client:            clientJSON,
	}

	response := inbound.ToResponse()

	assert.Equal(t, inbound.ID, response.ID)
	assert.Equal(t, inbound.ReferenceID, response.ReferenceID)
	assert.Equal(t, inbound.OriginID, response.OriginID)
	assert.Equal(t, inbound.TraceID, response.TraceID)
	assert.Equal(t, inbound.IP, response.IP)
	assert.Equal(t, inbound.Method, response.Method)
	assert.Equal(t, inbound.URL, response.URL)
	assert.Equal(t, inbound.Headers, response.Headers)
	assert.Equal(t, inbound.Body, response.Body)
	assert.Equal(t, inbound.StatusCode, response.StatusCode)
	assert.Equal(t, inbound.ResponseTimeMs, response.ResponseTimeMs)
	assert.Equal(t, inbound.ResponseBody, response.ResponseBody)
	assert.Equal(t, inbound.Metadata, response.Metadata)
	assert.Equal(t, inbound.SnapCompatibility, response.SnapCompatibility)
	assert.Equal(t, inbound.CreatedAt, response.CreatedAt)
	assert.Equal(t, inbound.UpdatedAt, response.UpdatedAt)

	assert.NotNil(t, response.Client)
	assert.Equal(t, client.Feature, response.Client.Feature)
	assert.Equal(t, client.Feature, response.Feature)
}
