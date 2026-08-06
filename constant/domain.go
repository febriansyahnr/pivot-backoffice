package constant

const (
	// CtxSQLTableNameKey is the context key for sql table name
	CtxSQLTableNameKey ContextKey = "table_name"
	// CtxRabbitMQStartTime is the context key for rabbitmq start time
	CtxRabbitMQStartTime ContextKey = "start_time"
	// CtxSQLTableNameKey is the context key for sql table name
	CtxRabbitMQReplyTo    ContextKey = "reply_to"
	CtxRabbitMQExpiration ContextKey = "rabbitmq_pub_with_expiration" // Use int64 data type for the value of this key.
	CtxRabbitMQRetryCount ContextKey = "retry_count"
)
