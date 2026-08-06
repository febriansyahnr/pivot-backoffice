package bigquery

import (
	"context"

	"cloud.google.com/go/bigquery"
)

// IBigQueryService defines the interface for BigQuery operations
type IBigQueryService interface {
	// Connection management
	NewClient(ctx context.Context) (*bigquery.Client, error)
	Close() error

	// Query operations
	ExecuteQuery(ctx context.Context, sql string) (*QueryResult, error)
	ExecuteQueryWithParams(ctx context.Context, sql string, params map[string]any) (*QueryResult, error)

	// Async operations
	QueryAsync(ctx context.Context, sql string) (string, error) // Returns JobID
	GetJobStatus(ctx context.Context, jobID string) (*JobStatus, error)
}

// Constants for BigQuery operations
const (
	// Query job states
	JobStatePending = "PENDING"
	JobStateRunning = "RUNNING"
	JobStateDone    = "DONE"

	// Field modes
	FieldModeNullable = "NULLABLE"
	FieldModeRequired = "REQUIRED"
	FieldModeRepeated = "REPEATED"

	// Common field types
	FieldTypeString    = "STRING"
	FieldTypeInteger   = "INTEGER"
	FieldTypeFloat     = "FLOAT"
	FieldTypeBoolean   = "BOOLEAN"
	FieldTypeTimestamp = "TIMESTAMP"
	FieldTypeDate      = "DATE"
	FieldTypeTime      = "TIME"
	FieldTypeDateTime  = "DATETIME"

	// Error messages
	ErrClientCreationFailed = "failed to create BigQuery client"
	ErrQueryExecutionFailed = "query execution failed"
	ErrResultReadFailed     = "error reading query results"
	ErrTableNotFound        = "table not found"
	ErrDatasetNotFound      = "dataset not found"
	ErrJobNotFound          = "job not found"
)
