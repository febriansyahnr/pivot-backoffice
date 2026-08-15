package cimbProcessorModel

type InquiryCorporateCreditCardRequest struct {
	BankCardNo string `json:"bankCardNo"`
}

type InquiryTransactionCorporateCreditCardRequest struct {
	PartnerReferenceNo string
	RecordType         string `json:"recordType"`
	BillingCycle       string `json:"billingCycle"`
	PostingDate        string `json:"postingDate"`
	BankCardNo         string `json:"-"`
	Page               int
}
