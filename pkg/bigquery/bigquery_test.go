package bigquery

import (
	"context"
	"testing"

	"cloud.google.com/go/bigquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBigQueryService(t *testing.T) {
	config := Config{
		ProjectID:           "test-project",
		Location:            "US",
		QueryTimeoutSeconds: 300,
	}

	service := NewBigQueryService(config)
	require.NotNil(t, service)

	bqService, ok := service.(*bigQueryService)
	require.True(t, ok)
	assert.Equal(t, config, bqService.config)
}

func TestBigQueryService_ClientNotAvailable(t *testing.T) {
	config := Config{
		ProjectID:           "test-project",
		Location:            "US",
		QueryTimeoutSeconds: 300,
	}

	service := NewBigQueryService(config)
	bqService := service.(*bigQueryService)

	ctx := context.Background()

	if rootErr != nil {
		t.Run("ExecuteQuery returns error when client unavailable", func(t *testing.T) {
			result, err := bqService.ExecuteQuery(ctx, "SELECT 1")
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Equal(t, rootErr, err)
		})

		t.Run("ExecuteQueryWithParams returns error when client unavailable", func(t *testing.T) {
			params := map[string]interface{}{"param": "value"}
			result, err := bqService.ExecuteQueryWithParams(ctx, "SELECT @param", params)
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Equal(t, rootErr, err)
		})

		t.Run("QueryAsync returns error when client unavailable", func(t *testing.T) {
			jobID, err := bqService.QueryAsync(ctx, "SELECT 1")
			assert.Error(t, err)
			assert.Empty(t, jobID)
			assert.Equal(t, rootErr, err)
		})

		t.Run("GetJobStatus returns error when client unavailable", func(t *testing.T) {
			status, err := bqService.GetJobStatus(ctx, "job_id")
			assert.Error(t, err)
			assert.Nil(t, status)
			assert.Equal(t, rootErr, err)
		})
	} else {
		t.Skip("BigQuery client is available, cannot test error cases")
	}
}

func TestBigQueryService_NewClient(t *testing.T) {
	config := Config{
		ProjectID:           "test-project",
		Location:            "US",
		QueryTimeoutSeconds: 300,
	}

	service := NewBigQueryService(config)
	bqService := service.(*bigQueryService)

	ctx := context.Background()
	clientResult, err := bqService.NewClient(ctx)

	if rootErr != nil {
		assert.Error(t, err)
		assert.Nil(t, clientResult)
		assert.Equal(t, rootErr, err)
	} else {
		assert.NoError(t, err)
		assert.NotNil(t, clientResult)
		assert.Equal(t, client, clientResult)
	}
}

func TestBigQueryService_Close(t *testing.T) {
	config := Config{
		ProjectID:           "test-project",
		Location:            "US",
		QueryTimeoutSeconds: 300,
	}

	service := NewBigQueryService(config)
	bqService := service.(*bigQueryService)

	err := bqService.Close()
	assert.NoError(t, err)
}

func TestBigQueryService_convertSchema(t *testing.T) {
	config := Config{
		ProjectID: "test-project",
		Location:  "US",
	}

	service := NewBigQueryService(config)
	bqService := service.(*bigQueryService)

	tests := []struct {
		name     string
		input    bigquery.Schema
		expected []FieldSchema
	}{
		{
			name: "Convert schema with various field types and modes",
			input: bigquery.Schema{
				{
					Name:        "required_string",
					Type:        bigquery.StringFieldType,
					Required:    true,
					Description: "A required string field",
				},
				{
					Name:     "nullable_integer",
					Type:     bigquery.IntegerFieldType,
					Required: false,
					Repeated: false,
				},
				{
					Name:     "repeated_float",
					Type:     bigquery.FloatFieldType,
					Required: false,
					Repeated: true,
				},
				{
					Name:     "boolean_field",
					Type:     bigquery.BooleanFieldType,
					Required: false,
				},
			},
			expected: []FieldSchema{
				{
					Name:        "required_string",
					Type:        "STRING",
					Mode:        FieldModeRequired,
					Description: "A required string field",
				},
				{
					Name: "nullable_integer",
					Type: "INTEGER",
					Mode: FieldModeNullable,
				},
				{
					Name: "repeated_float",
					Type: "FLOAT",
					Mode: FieldModeRepeated,
				},
				{
					Name: "boolean_field",
					Type: "BOOLEAN",
					Mode: FieldModeNullable,
				},
			},
		},
		{
			name:     "Empty schema",
			input:    bigquery.Schema{},
			expected: nil,
		},
		{
			name: "Date and time fields",
			input: bigquery.Schema{
				{
					Name:     "timestamp_field",
					Type:     bigquery.TimestampFieldType,
					Required: false,
				},
				{
					Name:     "date_field",
					Type:     bigquery.DateFieldType,
					Required: true,
				},
				{
					Name:     "datetime_field",
					Type:     bigquery.DateTimeFieldType,
					Required: false,
				},
			},
			expected: []FieldSchema{
				{
					Name: "timestamp_field",
					Type: "TIMESTAMP",
					Mode: FieldModeNullable,
				},
				{
					Name: "date_field",
					Type: "DATE",
					Mode: FieldModeRequired,
				},
				{
					Name: "datetime_field",
					Type: "DATETIME",
					Mode: FieldModeNullable,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := bqService.convertSchema(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfig(t *testing.T) {
	t.Run("Config struct initialization", func(t *testing.T) {
		config := Config{
			ProjectID:           "data-teams-staging",
			Location:            "asia-southeast1",
			QueryTimeoutSeconds: 600,
		}

		assert.Equal(t, "data-teams-staging", config.ProjectID)
		assert.Equal(t, "asia-southeast1", config.Location)
		assert.Equal(t, 600, config.QueryTimeoutSeconds)
	})

	t.Run("Zero value config", func(t *testing.T) {
		config := Config{}
		assert.Empty(t, config.ProjectID)
		assert.Empty(t, config.Location)
		assert.Zero(t, config.QueryTimeoutSeconds)
	})
}

func TestConstants(t *testing.T) {
	t.Run("Job state constants", func(t *testing.T) {
		assert.Equal(t, "PENDING", JobStatePending)
		assert.Equal(t, "RUNNING", JobStateRunning)
		assert.Equal(t, "DONE", JobStateDone)
	})

	t.Run("Field mode constants", func(t *testing.T) {
		assert.Equal(t, "NULLABLE", FieldModeNullable)
		assert.Equal(t, "REQUIRED", FieldModeRequired)
		assert.Equal(t, "REPEATED", FieldModeRepeated)
	})

	t.Run("Field type constants", func(t *testing.T) {
		assert.Equal(t, "STRING", FieldTypeString)
		assert.Equal(t, "INTEGER", FieldTypeInteger)
		assert.Equal(t, "FLOAT", FieldTypeFloat)
		assert.Equal(t, "BOOLEAN", FieldTypeBoolean)
		assert.Equal(t, "TIMESTAMP", FieldTypeTimestamp)
		assert.Equal(t, "DATE", FieldTypeDate)
		assert.Equal(t, "TIME", FieldTypeTime)
		assert.Equal(t, "DATETIME", FieldTypeDateTime)
	})

	t.Run("Error message constants", func(t *testing.T) {
		assert.Equal(t, "failed to create BigQuery client", ErrClientCreationFailed)
		assert.Equal(t, "query execution failed", ErrQueryExecutionFailed)
		assert.Equal(t, "error reading query results", ErrResultReadFailed)
		assert.Equal(t, "table not found", ErrTableNotFound)
		assert.Equal(t, "dataset not found", ErrDatasetNotFound)
		assert.Equal(t, "job not found", ErrJobNotFound)
	})
}

func TestResponseStructs(t *testing.T) {
	t.Run("QueryResult struct", func(t *testing.T) {
		result := QueryResult{
			Rows:      []map[string]interface{}{{"test": "value"}},
			Schema:    []FieldSchema{{Name: "test", Type: "STRING"}},
			TotalRows: 1,
		}

		assert.Len(t, result.Rows, 1)
		assert.Equal(t, "value", result.Rows[0]["test"])
		assert.Len(t, result.Schema, 1)
		assert.Equal(t, "test", result.Schema[0].Name)
		assert.Equal(t, int64(1), result.TotalRows)
	})

	t.Run("FieldSchema struct", func(t *testing.T) {
		field := FieldSchema{
			Name:        "test_field",
			Type:        FieldTypeString,
			Mode:        FieldModeRequired,
			Description: "Test description",
		}

		assert.Equal(t, "test_field", field.Name)
		assert.Equal(t, "STRING", field.Type)
		assert.Equal(t, "REQUIRED", field.Mode)
		assert.Equal(t, "Test description", field.Description)
	})

	t.Run("TableInfo struct", func(t *testing.T) {
		info := TableInfo{
			TableID:     "test_table",
			Description: "Test table",
			NumRows:     100,
			NumBytes:    1024,
		}

		assert.Equal(t, "test_table", info.TableID)
		assert.Equal(t, "Test table", info.Description)
		assert.Equal(t, int64(100), info.NumRows)
		assert.Equal(t, int64(1024), info.NumBytes)
	})

	t.Run("JobStatus struct", func(t *testing.T) {
		status := JobStatus{
			JobID:    "test_job",
			State:    JobStateDone,
			ErrorMsg: "",
		}

		assert.Equal(t, "test_job", status.JobID)
		assert.Equal(t, "DONE", status.State)
		assert.Empty(t, status.ErrorMsg)
	})
}
