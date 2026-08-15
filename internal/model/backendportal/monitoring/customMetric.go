package monitoring

type CustomMetric struct {
	MetricInstrumentType string // COUNTER, GAUGE, HISTOGRAM
	ComponentName        string
	MetricName           string
	MetricValue          int64
	Attributes           map[string]any
}
