package bigquery

import (
	"context"
	"fmt"
	"sync"
	"time"

	"cloud.google.com/go/bigquery"
	"go.opentelemetry.io/otel"
	"google.golang.org/api/iterator"
)

var (
	otelTracer = otel.Tracer("pkg/bigquery")
	once       sync.Once
	client     *bigquery.Client
	rootErr    error
)

type bigQueryService struct {
	config Config
}

func NewBigQueryService(config Config) IBigQueryService {
	once.Do(func() {
		client, rootErr = bigquery.NewClient(context.Background(), config.ProjectID)
	})

	return &bigQueryService{
		config: config,
	}
}

// NewClient creates a new BigQuery client using Application Default Credentials
func (b *bigQueryService) NewClient(ctx context.Context) (*bigquery.Client, error) {
	ctx, segment := otelTracer.Start(ctx, "pkg/bigquery/NewClient")
	defer segment.End()

	if client != nil {
		return client, nil
	}

	return nil, rootErr
}

// Close closes the BigQuery client
func (b *bigQueryService) Close() error {
	if client != nil {
		return client.Close()
	}
	return nil
}

// ExecuteQuery executes a SQL query and returns results
func (b *bigQueryService) ExecuteQuery(ctx context.Context, sql string) (*QueryResult, error) {
	ctx, segment := otelTracer.Start(ctx, "pkg/bigquery/ExecuteQuery")
	defer segment.End()

	if client == nil {
		return nil, rootErr
	}

	// Create query with configuration
	q := client.Query(sql)
	q.Location = b.config.Location
	if b.config.QueryTimeoutSeconds > 0 {
		q.QueryConfig.JobTimeout = time.Duration(b.config.QueryTimeoutSeconds) * time.Second
	}

	// Execute query and read results
	it, err := q.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrQueryExecutionFailed, err)
	}

	return b.iteratorToResult(it)
}

// ExecuteQueryWithParams executes a parameterized SQL query
func (b *bigQueryService) ExecuteQueryWithParams(ctx context.Context, sql string, params map[string]any) (*QueryResult, error) {
	ctx, segment := otelTracer.Start(ctx, "pkg/bigquery/ExecuteQueryWithParams")
	defer segment.End()

	if client == nil {
		return nil, rootErr
	}

	// Create query with parameters
	q := client.Query(sql)
	q.Location = b.config.Location
	if b.config.QueryTimeoutSeconds > 0 {
		q.QueryConfig.JobTimeout = time.Duration(b.config.QueryTimeoutSeconds) * time.Second
	}

	// Set parameters
	for key, value := range params {
		q.Parameters = append(q.Parameters, bigquery.QueryParameter{
			Name:  key,
			Value: value,
		})
	}

	// Execute query and read results
	it, err := q.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrQueryExecutionFailed, err)
	}

	return b.iteratorToResult(it)
}

// QueryAsync executes a query asynchronously and returns the job ID
func (b *bigQueryService) QueryAsync(ctx context.Context, sql string) (string, error) {
	ctx, segment := otelTracer.Start(ctx, "pkg/bigquery/QueryAsync")
	defer segment.End()

	if client == nil {
		return "", rootErr
	}

	q := client.Query(sql)
	q.Location = b.config.Location
	if b.config.QueryTimeoutSeconds > 0 {
		q.QueryConfig.JobTimeout = time.Duration(b.config.QueryTimeoutSeconds) * time.Second
	}

	job, err := q.Run(ctx)
	if err != nil {
		return "", fmt.Errorf("%s: %w", ErrQueryExecutionFailed, err)
	}

	return job.ID(), nil
}

// GetJobStatus retrieves the status of a BigQuery job
func (b *bigQueryService) GetJobStatus(ctx context.Context, jobID string) (*JobStatus, error) {
	ctx, segment := otelTracer.Start(ctx, "pkg/bigquery/GetJobStatus")
	defer segment.End()

	if client == nil {
		return nil, rootErr
	}

	job, err := client.JobFromID(ctx, jobID)
	if err != nil {
		return nil, err
	}

	status, err := job.Status(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrJobNotFound, err)
	}

	jobStatus := &JobStatus{
		JobID:   jobID,
		Created: status.Statistics.CreationTime,
	}

	// Map BigQuery job state to our constants
	switch status.State {
	case bigquery.Pending:
		jobStatus.State = JobStatePending
	case bigquery.Running:
		jobStatus.State = JobStateRunning
		jobStatus.Started = status.Statistics.StartTime
	case bigquery.Done:
		jobStatus.State = JobStateDone
		jobStatus.Started = status.Statistics.StartTime
		jobStatus.Ended = status.Statistics.EndTime

		// Check for errors
		if status.Err() != nil {
			jobStatus.ErrorMsg = status.Err().Error()
		}
	}

	return jobStatus, nil
}

// iteratorToResult converts a BigQuery RowIterator to our QueryResult format
func (b *bigQueryService) iteratorToResult(it *bigquery.RowIterator) (*QueryResult, error) {
	var rows []map[string]any
	totalRows := int64(0)

	// Read all rows
	for {
		var values []bigquery.Value
		err := it.Next(&values)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("%s: %w", ErrResultReadFailed, err)
		}

		// Convert BigQuery values to map
		row := make(map[string]any)
		for i, field := range it.Schema {
			row[field.Name] = values[i]
		}
		rows = append(rows, row)
		totalRows++
	}

	// Convert schema
	schema := b.convertSchema(it.Schema)

	return &QueryResult{
		Rows:      rows,
		Schema:    schema,
		TotalRows: totalRows,
	}, nil
}

// convertSchema converts BigQuery schema to our FieldSchema format
func (b *bigQueryService) convertSchema(bqSchema bigquery.Schema) []FieldSchema {
	var schema []FieldSchema
	for _, field := range bqSchema {
		fieldSchema := FieldSchema{
			Name:        field.Name,
			Type:        string(field.Type),
			Description: field.Description,
		}

		// Map BigQuery field modes to our constants
		switch field.Required {
		case true:
			fieldSchema.Mode = FieldModeRequired
		case false:
			if field.Repeated {
				fieldSchema.Mode = FieldModeRepeated
			} else {
				fieldSchema.Mode = FieldModeNullable
			}
		}

		schema = append(schema, fieldSchema)
	}
	return schema
}
