package danaProcessorModel

type EscrowBalanceHeadRequest struct {
	Function     string `json:"function"`
	ClientId     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	Version      string `json:"version"`
	ReqTime      string `json:"reqTime"`
	ReqMsgId     string `json:"reqMsgId"`
	Reserve      string `json:"reserve"`
}

type EscrowBalanceBodyRequest struct {
	RequestMerchantId        string   `json:"requestMerchantId"`
	MerchantResourceInfoList []string `json:"merchantResourceInfoList"`
}

type EscrowBalanceRequestPayload struct {
	Head EscrowBalanceHeadRequest `json:"head"`
	Body EscrowBalanceBodyRequest `json:"body"`
}

type EscrowBalanceRequest struct {
	Request   EscrowBalanceRequestPayload `json:"request"`
	Signature string                      `json:"signature"`
}

type EscrowBalanceHeadResponse struct {
	Function string `json:"function"`
	ClientId string `json:"clientId"`
	Version  string `json:"version"`
	RespTime string `json:"respTime"`
	ReqMsgId string `json:"reqMsgId"`
}

type ResultInfo struct {
	ResultCodeId string `json:"resultCodeId"`
	ResultCode   string `json:"resultCode"`
	ResultMsg    string `json:"resultMsg"`
	ResultStatus string `json:"resultStatus"`
}

type ValueInfo struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

type MerchantResourceInformation struct {
	ResourceType string `json:"resourceType"`
	Value        string `json:"value"`
}

type EscrowBalanceBodyResponse struct {
	ResultInfo                   ResultInfo                    `json:"resultInfo"`
	MerchantResourceInformations []MerchantResourceInformation `json:"merchantResourceInformations"`
}

type EscrowBalanceResponsePayload struct {
	Head EscrowBalanceHeadResponse `json:"head"`
	Body EscrowBalanceBodyResponse `json:"body"`
}

type EscrowBalanceResponse struct {
	Response  EscrowBalanceResponsePayload `json:"response"`
	Signature string                       `json:"signature"`
}
