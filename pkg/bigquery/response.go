package bigquery

import "time"

type QueryResult struct {
	Rows      []map[string]any `json:"rows"`
	Schema    []FieldSchema    `json:"schema"`
	TotalRows int64            `json:"totalRows"`
	JobID     string           `json:"jobId,omitempty"`
}

type FieldSchema struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // STRING, INTEGER, FLOAT, BOOLEAN, TIMESTAMP, etc.
	Mode        string `json:"mode"` // NULLABLE, REQUIRED, REPEATED
	Description string `json:"description"`
}

type TableInfo struct {
	TableID      string    `json:"tableId"`
	Description  string    `json:"description"`
	CreatedTime  time.Time `json:"createdTime"`
	LastModified time.Time `json:"lastModified"`
	NumRows      int64     `json:"numRows"`
	NumBytes     int64     `json:"numBytes"`
}

type JobStatus struct {
	JobID    string    `json:"jobId"`
	State    string    `json:"state"` // PENDING, RUNNING, DONE
	Created  time.Time `json:"created"`
	Started  time.Time `json:"started,omitempty"`
	Ended    time.Time `json:"ended,omitempty"`
	ErrorMsg string    `json:"error,omitempty"`
}
