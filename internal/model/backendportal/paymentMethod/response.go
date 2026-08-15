package paymentMethodModel

type MerchantRequiredDocumentsResponse struct {
	Name   string `json:"name"`
	Format string `json:"format"` // data / document

	MerchantDocumentID     string `json:"merchantDocumentID"`
	MerchantDocumentStatus string `json:"merchantDocumentStatus"`
}
