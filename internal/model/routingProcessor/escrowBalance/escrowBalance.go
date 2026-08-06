package routingProcessorModelEscrowBalance

type EscrowBalanceResponse struct {
	ResponseCode       string  `json:"responseCode"`
	ResponseMessage    string  `json:"responseMessage"`
	ProcessorReference string  `json:"processorReference,omitempty"`
	BalanceAmount      float64 `json:"balance"`
}
