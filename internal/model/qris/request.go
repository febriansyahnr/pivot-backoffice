package qris

type RegistrationReq struct {
	Type         string `json:"-" validate:"-"`
	MerchantId   string `json:"merchantId" validate:"required,uuid"`
	Acquirer     string `json:"acquirer" validate:"required,oneof=BNC BRI BNI"`
	MerchantType string `json:"merchant_type" validate:"oneof=Merchant Franchisee Sub-Merchant"`
	CreatedBy    string `json:"createdBy" validate:"required,max=50"`
}

type ReuploadDocumentReq struct {
	RegistrationId string `json:"registrationId" validate:"required,numeric"`
	DocumentType   string `json:"documentType" validate:"required,oneof=NationalIdentityCard BusinessLicense TaxIdentification BusinessRegistration CertificateIncorporation CertificateNo40 CertificateLastAmendment CertificateDeedAmendment CertificateAmendmentAct CertificateEstablishment CertificateTaxRegistration BusinessEnvironmentPhoto"`
}

type DuplicateRegistrationReq struct {
	SourceMerchantId string `json:"sourceMerchantId" validate:"required,uuid"`
	TargetMerchantId string `json:"targetMerchantId" validate:"required,uuid"`
}
