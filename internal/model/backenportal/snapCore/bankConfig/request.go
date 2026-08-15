package snapCoreModel

type GetBankCodeListRequest struct {
	IsActive     int    `json:"isActive"`
	TransferType string `json:"transferType"`
}
