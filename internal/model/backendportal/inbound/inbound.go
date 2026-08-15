package inboundModel

import (
	"time"

	"github.com/jmoiron/sqlx/types"
)

// Inbound represents the schema of the inbound table
type Inbound struct {
	ID                string             `json:"id" db:"id"`
	ReferenceID       string             `json:"referenceId" db:"reference_id"`
	OriginID          string             `json:"originId" db:"origin_id"`
	TraceID           string             `json:"traceId" db:"trace_id"`
	IP                string             `json:"ip" db:"ip"`
	Client            types.JSONText     `json:"client" db:"client"`
	Method            string             `json:"method" db:"method"`
	URL               string             `json:"url" db:"url"`
	Headers           types.JSONText     `json:"headers" db:"headers"`
	Body              types.NullJSONText `json:"body" db:"body"`
	StatusCode        int                `json:"statusCode" db:"status_code"`
	ResponseTimeMs    float64            `json:"responseTimeMs" db:"response_time_ms"`
	ResponseBody      types.NullJSONText `json:"responseBody" db:"response_body"`
	Metadata          types.NullJSONText `json:"metadata" db:"metadata"`
	SnapCompatibility bool               `json:"snapCompatibility" db:"snap_compatibility"`
	CreatedAt         time.Time          `json:"createdAt" db:"created_at"`
	UpdatedAt         time.Time          `json:"updatedAt" db:"updated_at"`
}
