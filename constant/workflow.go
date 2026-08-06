package constant

const (
	// Workflow List
	WorkflowMerchantCallback = "merchant_callbacks"

	// Task List
	WorkflowTaskMerchantCallbackPreparation          = "merchant_callback_preparation"
	WorkflowTaskMerchantCallbackDelivery             = "merchant_callback_delivery"
	WorkflowTaskMerchantCallbackDeliveryWithoutRetry = "merchant_callback_delivery_without_retry"
	WorkflowTaskMerchantCallbackWriteLog             = "merchant_callback_write_log"
	WorkflowTaskMerchantCallbackWriteMetric          = "merchant_callback_write_metric"
)
