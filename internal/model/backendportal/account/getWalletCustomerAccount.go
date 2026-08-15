package account_model

type GetCustomerAccountRequest struct {
	CustomerID string `json:"customerId" validate:"required,uuid"`
	MerchantID string `json:"merchantId" validate:"required,uuid"`
}

type CalculateBulkLedgerBalanceRequest struct {
	MerchantID string
	AccountIDs []string `json:"accountIds"`
}
