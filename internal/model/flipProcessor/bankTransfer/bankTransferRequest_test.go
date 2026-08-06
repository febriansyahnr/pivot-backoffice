package flipProcessorModel

import (
	"testing"

	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	routingProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/bankTransfer"
	"github.com/stretchr/testify/assert"
)

func TestNewSpecialMoneyTransferRequestFromProcessor(t *testing.T) {
	testCases := []struct {
		desc    string
		request *routingProcessorModel.BankTransferRequest
		want    *SpecialMoneyTransferRequest
	}{
		{
			desc: "success wrap request",
			request: &routingProcessorModel.BankTransferRequest{
				Beneficiary: routingProcessorModel.SubjectRequest{
					AccountNo: "1234567890",
					BankCode:  "123",
				},
				Amount: commonModel.Amount{
					Value:    "1000000.00",
					Currency: "IDR",
				},
				Remark: "test",
				Source: routingProcessorModel.SubjectRequest{
					Name:    "test",
					Address: "test",
				},
			},
			want: &SpecialMoneyTransferRequest{
				AccountNumber: "1234567890",
				BankCode:      "123",
				Amount:        1000000,
				Remark:        "test",
				Direction:     "DOMESTIC_SPECIAL_TRANSFER",
				SenderName:    "PT Harsya Remitindo",
				SenderAddress: "Biomedical Campus, Knowledge Tower Lt. 3, Kav. Digital Hub",
				SenderCountry: 100252, //required, fill for indonesia code
				SenderJob:     "company",
			},
		},
		{
			desc: "long remark should be trimmed to 18 characters",
			request: &routingProcessorModel.BankTransferRequest{
				Beneficiary: routingProcessorModel.SubjectRequest{
					AccountNo: "9876543210",
					BankCode:  "456",
				},
				Amount: commonModel.Amount{
					Value:    "2000000.00",
					Currency: "IDR",
				},
				Remark: "This is a very long remark that should be trimmed to 18 characters only",
				Source: routingProcessorModel.SubjectRequest{
					Name:    "test2",
					Address: "test2",
				},
			},
			want: &SpecialMoneyTransferRequest{
				AccountNumber: "9876543210",
				BankCode:      "456",
				Amount:        2000000,
				Remark:        "This is a very lon", // Trimmed to 18 characters
				Direction:     "DOMESTIC_SPECIAL_TRANSFER",
				SenderName:    "PT Harsya Remitindo",
				SenderAddress: "Biomedical Campus, Knowledge Tower Lt. 3, Kav. Digital Hub",
				SenderCountry: 100252,
				SenderJob:     "company",
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got := NewSpecialTransferRequestFromProcessor(tC.request)
			assert.Equal(t, tC.want, got)
		})
	}
}
