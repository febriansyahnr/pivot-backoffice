package cdcModel

import "encoding/json"

// Operation types from Debezium CDC events
const (
	// OpCreate represents INSERT operation
	OpCreate = "c"
	// OpUpdate represents UPDATE operation
	OpUpdate = "u"
	// OpDelete represents DELETE operation
	OpDelete = "d"
)

// ParseEvent unmarshals JSON data into a Debezium CDC Event.
// Returns an error if the JSON cannot be parsed.
func ParseEvent[T comparable](data []byte) (*Event[T], error) {
	var event Event[T]
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, err
	}
	return &event, nil
}

// Event represents the Debezium CDC change event wrapper.
// It contains the before and after states of the changed record.
type Event[T comparable] struct {
	Before      *T     `json:"before"`
	After       *T     `json:"after"`
	Source      Source `json:"source"`
	Transaction any    `json:"transaction"`
	Op          string `json:"op"`
	TsMs        int64  `json:"ts_ms"`
	TsUs        int64  `json:"ts_us"`
	TsNs        int64  `json:"ts_ns"`
}

// IsEmpty returns true if both Before and After states are nil.
func (e *Event[T]) IsEmpty() bool { return e.Before == nil && e.After == nil }

// IsCreate returns true if the operation is an INSERT.
func (e *Event[T]) IsCreate() bool { return e.Op == OpCreate }

// IsUpdate returns true if the operation is an UPDATE.
func (e *Event[T]) IsUpdate() bool { return e.Op == OpUpdate }

// IsDelete returns true if the operation is a DELETE.
func (e *Event[T]) IsDelete() bool { return e.Op == OpDelete }

// GetCurrent returns the current state of the record.
// For DELETE operations, this returns Before state.
// For all other operations, this returns After state.
func (e *Event[T]) GetCurrent() *T {
	if e.IsDelete() {
		return e.Before
	}
	return e.After
}

// GetPrevious returns the previous state of the record.
// For CREATE operations, this returns nil since there is no previous state.
// For all other operations, this returns Before state.
func (e *Event[T]) GetPrevious() *T {
	if e.IsCreate() {
		return nil
	}
	return e.Before
}

// Source contains metadata about the database source
type Source struct {
	Version   string `json:"version"`
	Connector string `json:"connector"`
	Name      string `json:"name"`
	TsMs      int64  `json:"ts_ms"`
	Snapshot  string `json:"snapshot"`
	Db        string `json:"db"`
	Sequence  any    `json:"sequence"`
	TsUs      int64  `json:"ts_us"`
	TsNs      int64  `json:"ts_ns"`
	Table     string `json:"table"`
	ServerID  int    `json:"server_id"`
	File      string `json:"file"`
	Pos       int    `json:"pos"`
	Row       int    `json:"row"`
	Thread    int    `json:"thread"`
}
