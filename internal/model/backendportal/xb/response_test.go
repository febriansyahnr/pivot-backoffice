package xbModel_test

import (
	"testing"

	pb "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/proto/messages/callback"
	. "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/xb"

	"github.com/stretchr/testify/assert"
)

func TestToProtoSenderDataCallback(t *testing.T) {
	input := SenderDataResponse{
		Name:                 "name",
		CountryCode:          "ID",
		CountryName:          "indonesia",
		State:                "state",
		City:                 "city",
		Address:              "address",
		Postcode:             "post_code",
		AccountType:          "account_type",
		IdentificationType:   "identification_type",
		IdentificationNumber: "identification_number",
	}

	want := &pb.SenderXbData{
		Name:                 "name",
		Country:              "indonesia",
		CountryCode:          "ID",
		State:                "state",
		City:                 "city",
		Address:              "address",
		Postcode:             "post_code",
		AccountType:          "account_type",
		IdentificationType:   "identification_type",
		IdentificationNumber: "identification_number",
	}

	assert.Equal(t, want, input.ToProtoSenderDataCallback())
}

func TestToProtoBeneficiaryXbDataCallback(t *testing.T) {
	input := BeneficiaryDataResponse{
		Name:          "name",
		Address:       "address",
		City:          "city",
		Postcode:      "post_code",
		State:         "state",
		CountryCode:   "ID",
		CountryName:   "indonesia",
		AccountNumber: "account_number",
		BankName:      "bank_name",
		BankCode:      "bank_code",
	}

	want := &pb.BeneficiaryXbData{
		Name:          "name",
		Address:       "address",
		City:          "city",
		Postcode:      "post_code",
		State:         "state",
		CountryCode:   "ID",
		Country:       "indonesia",
		AccountNumber: "account_number",
		BankName:      "bank_name",
		BankCode:      "bank_code",
	}

	assert.Equal(t, want, input.ToProtoBeneficiaryXbDataCallback())
}
