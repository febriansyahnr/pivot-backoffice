package installmentPlanModel

type FilterRequest struct {
	MerchantID     string
	InstallmentIDs []string
	Acquirer       string
	SettlementType string
	PaymentMethod  string
	Tenor          int
	Status         string
	MidID          string
	Page           int
	PageSize       int
}
