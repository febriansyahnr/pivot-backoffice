package qris

import (
	repository "github.com/paper-indonesia/pivot-backoffice/internal/repository/backendportal"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"go.opentelemetry.io/otel"
)

const qrRegistrationTableName = "qr_registrations"

var otelTracer = otel.Tracer("QrisRepository")

var documents = map[string][]string{
	"NationalIdentityCard":       {"nationalIdentityCardNumber", "nationalIdentityCardFile"},
	"BusinessLicense":            {"businessLicenseNumber", "businessLicenseFile"},
	"TaxIdentification":          {"taxIdentificationNumber", "taxIdentificationFile"},
	"BusinessRegistration":       {"businessRegistrationNumber", "businessRegistrationFile"},
	"CertificateIncorporation":   {"certificateIncorporation"},
	"CertificateNo40":            {"certificateNo40"},
	"CertificateLastAmendment":   {"certificateLastAmendment"},
	"CertificateDeedAmendment":   {"certificateDeedAmendment"},
	"CertificateAmendmentAct":    {"certificateAmendmentAct"},
	"CertificateEstablishment":   {"certificateEstablishment"},
	"CertificateTaxRegistration": {"certificateTaxRegistration"},
	"BusinessEnvironmentPhoto":   {"businessEnvironmentPhoto"},
}

type qrisRepository struct {
	db mySqlExt.IMySqlExt
}

func New(db mySqlExt.IMySqlExt) repository.IQrisRepository {
	return &qrisRepository{db}
}
