package routingProcessorModel

type ProcessorPriority struct {
	ProcessorName       string   `json:"processorName"`
	AllowedDestinations []string `json:"allowedDestinations"`
	Priority            int      `json:"priority"`
	IsActive            bool     `json:"isActive"`
}
