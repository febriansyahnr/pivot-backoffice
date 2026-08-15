package bankAccount

import "testing"

func TestToResponse(t *testing.T) {
	b := BankAccount{
		BeneficiaryBankCode:    "008",
		BeneficiaryBankName:    "Bank 008",
		BeneficiaryAccountNo:   "8000800808",
		BeneficiaryAccountName: "Yories Yolanda",
	}

	bankAccountResponse := b.ToResponse()

	if bankAccountResponse.BeneficiaryBankCode != b.BeneficiaryBankCode {
		t.Errorf("Expected BeneficiaryBankCode to be %s, got %s", b.BeneficiaryBankCode, bankAccountResponse.BeneficiaryBankCode)
	}

	if bankAccountResponse.BeneficiaryBankName != b.BeneficiaryBankName {
		t.Errorf("Expected BeneficiaryBankName to be %s, got %s", b.BeneficiaryBankName, bankAccountResponse.BeneficiaryBankName)
	}

	if bankAccountResponse.BeneficiaryAccountNo != b.BeneficiaryAccountNo {
		t.Errorf("Expected BeneficiaryAccountNo to be %s, got %s", b.BeneficiaryAccountNo, bankAccountResponse.BeneficiaryAccountNo)
	}

	if bankAccountResponse.BeneficiaryAccountName != b.BeneficiaryAccountName {
		t.Errorf("Expected BeneficiaryAccountName to be %s, got %s", b.BeneficiaryAccountName, bankAccountResponse.BeneficiaryAccountName)
	}

}
