package account_model

type CreateCustomerAccountRequest struct {
	CustomerID string `json:"customerId" validate:"required,uuid"`
}
