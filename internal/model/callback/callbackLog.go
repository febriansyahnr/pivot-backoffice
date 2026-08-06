package callback_model

import (
	"encoding/json"
	"time"

	"github.com/jmoiron/sqlx/types"
	"github.com/paper-indonesia/pivot-backoffice/constant"

	"github.com/google/uuid"
)

type CallbackLog struct {
	UUID        uuid.UUID          `db:"uuid" json:"uuid" example:"b3b3b3b3-3b3b-3b3b-3b3b-3b3b3b3b3b3b"`
	CallbackID  uuid.UUID          `db:"callback_id" json:"callbackId" example:"b3b3b3b3-3b3b-3b3b-3b3b-3b3b3b3b3b3b"`
	Event       *string            `db:"event" json:"event" example:"PAYOUT.DONE"`
	Request     string             `db:"request" json:"request" example:"{}"`
	Response    *string            `db:"response" json:"response" example:"{}"`
	Status      string             `db:"status" json:"status" example:"DELIVERED"`
	Retry       int                `db:"retry" json:"retry" example:"1"`
	RawMetadata types.NullJSONText `db:"metadata" json:"-"`
	ReferenceId *string            `db:"reference_id" json:"referenceId" example:"AUTO17582633126772"`
	CreatedAt   time.Time          `db:"created_at" json:"createdAt" example:"2024-01-01T00:00:00Z"`
	UpdatedAt   time.Time          `db:"updated_at" json:"updatedAt" example:"2024-01-01T00:00:00Z"`
	// Intended for internal processes only
	Metadata *CallbackLogMetadata `db:"-" json:"-"`
}

type CallbackLogMetadata struct {
	WorkflowId string `json:"workflowId"`
}

type CallbackLogWithMaster struct {
	UUID        uuid.UUID `db:"uuid" json:"uuid" example:"b3b3b3b3-3b3b-3b3b-3b3b-3b3b3b3b3b3b"`
	CallbackID  uuid.UUID `db:"callback_id" json:"callbackId" example:"b3b3b3b3-3b3b-3b3b-3b3b-3b3b3b3b3b3b"`
	Type        string    `db:"type" json:"type" example:"PAYOUT"`
	BaseURL     *string   `db:"base_url" json:"baseUrl" example:"http://test"`
	URL         string    `db:"url" json:"url" example:"http://test"`
	Event       *string   `db:"event" json:"event" example:"PAYOUT.DONE"`
	Request     string    `db:"request" json:"request" example:"{}"`
	Response    *string   `db:"response" json:"response" example:"{}"`
	Status      string    `db:"status" json:"status" example:"DELIVERED"`
	Retry       int       `db:"retry" json:"retry" example:"1"`
	ReferenceId *string   `db:"reference_id" json:"referenceId" example:"AUTO17582633126772"`
	CreatedAt   time.Time `db:"created_at" json:"createdAt" example:"2024-01-01T00:00:00Z"`
	UpdatedAt   time.Time `db:"updated_at" json:"updatedAt" example:"2024-01-01T00:00:00Z"`

	MerchantID string `db:"merchant_id" json:"-"`
}

// MarshalJSON customizes the JSON output for the CallbackLog struct.
func (c CallbackLogWithMaster) MarshalJSON() ([]byte, error) {

	// Create an alias to avoid recursion
	type Alias CallbackLogWithMaster

	// Determine the event title
	eventTitle := ""
	if c.Event != nil {
		eventTitle = constant.GetCallbackEventTitle(*c.Event)
	}

	// Create an anonymous struct to include EventTitle
	return json.Marshal(&struct {
		Alias
		EventTitle string `json:"eventTitle"`
	}{
		Alias:      Alias(c),
		EventTitle: eventTitle,
	})
}

func (c CallbackLogWithMaster) ToCallbackLog() *CallbackLog {
	return &CallbackLog{
		UUID:        c.UUID,
		CallbackID:  c.CallbackID,
		Event:       c.Event,
		Request:     c.Request,
		Response:    c.Response,
		Status:      c.Status,
		Retry:       c.Retry,
		ReferenceId: c.ReferenceId,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}
