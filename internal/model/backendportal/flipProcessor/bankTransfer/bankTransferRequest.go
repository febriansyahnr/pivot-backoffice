package flipProcessorModel

import (
	"strconv"
	"strings"

	routingProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/bankTransfer"
)

type SpecialMoneyTransferRequest struct {
	AccountNumber        string `json:"account_number,omitempty"`
	BankCode             string `json:"bank_code,omitempty"`
	Amount               int    `json:"amount,omitempty"`
	Remark               string `json:"remark,omitempty"`
	RecipientCity        string `json:"recipient_city,omitempty"`
	SenderCountry        int    `json:"sender_country,omitempty"`
	SenderPlaceOfBirth   int    `json:"sender_place_of_birth,omitempty"`
	SenderDateOfBirth    string `json:"sender_date_of_birth,omitempty"`
	SenderIdentityType   string `json:"sender_identity_type,omitempty"`
	SenderName           string `json:"sender_name,omitempty"`
	SenderAddress        string `json:"sender_address,omitempty"`
	SenderIdentityNumber string `json:"sender_identity_number,omitempty"`
	SenderJob            string `json:"sender_job,omitempty"`
	Direction            string `json:"direction,omitempty"`
	BeneficiaryEmail     string `json:"beneficiary_email,omitempty"`
}

const (
	FlipIndonesiaCountryCode         = 100252
	FlipTransferDefaultSenderName    = "PT Harsya Remitindo"
	FlipTransferDefaultSenderAddress = "Biomedical Campus, Knowledge Tower Lt. 3, Kav. Digital Hub"
)

func NewSpecialTransferRequestFromProcessor(rp *routingProcessorModel.BankTransferRequest) *SpecialMoneyTransferRequest {
	amountF, _ := strconv.ParseFloat(rp.Amount.Value, 64)

	// trim remark maximum 18 character
	if len(rp.Remark) > 18 {
		rp.Remark = rp.Remark[:18]
	}

	return &SpecialMoneyTransferRequest{
		AccountNumber: rp.Beneficiary.AccountNo,
		BankCode:      strings.ToLower(rp.Beneficiary.BankCode),
		Amount:        int(amountF),
		Remark:        rp.Remark,
		Direction:     "DOMESTIC_SPECIAL_TRANSFER",
		SenderName:    FlipTransferDefaultSenderName,
		SenderAddress: FlipTransferDefaultSenderAddress,
		SenderCountry: FlipIndonesiaCountryCode, //required, fill for indonesia code
		SenderJob:     "company",
	}
}
