package conductor

const mimeApplicationJson = "application/json"

type TaskStatus string

const (
	TaskStatusSuccess           TaskStatus = "SUCCESS"
	TaskStatusRetryableError    TaskStatus = "RETRYABLE_ERROR"
	TaskStatusNonRetryableError TaskStatus = "NON_RETRYABLE_ERROR"
)
