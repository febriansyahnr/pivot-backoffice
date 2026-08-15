package banktransfer_test

import (
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/snapCore/bankTransfer"

	"github.com/stretchr/testify/assert"
)

func TestMappingAccountTransactionErrStatus(t *testing.T) {
	tests := []struct {
		input           BankTransferResponseData
		wantStatus      string
		wantType        string
		wantDescription string
	}{
		{
			input:      BankTransferResponseData{},
			wantStatus: c.StatusFailed,
			wantType:   c.ReasonTypeOtherReason,
		},
		{
			input: BankTransferResponseData{
				Status:          c.SnapCoreBankTransferStatusPending,
				ResponseMessage: "Test",
			},
			wantStatus:      c.StatusPending,
			wantType:        c.ReasonTypeOtherReason,
			wantDescription: "Test",
		},
		{
			input: BankTransferResponseData{
				Status:          c.SnapCoreBankTransferStatusFailed,
				ResponseCode:    "4030014",
				ResponseMessage: "Insufficient fund",
			},
			wantStatus:      c.StatusPending,
			wantType:        c.ReasonTypeInsufficientEscrowFund,
			wantDescription: "Insufficient fund",
		},
		{
			input: BankTransferResponseData{
				ResponseCode: "4030018",
			},
			wantStatus:      c.StatusFailed,
			wantType:        c.ReasonTypeBeneficiaryAccountReason,
			wantDescription: c.SnapCoreResponseInactiveAccountMessage,
		},
		{
			input: BankTransferResponseData{
				ResponseCode: "4030009",
			},
			wantStatus:      c.StatusFailed,
			wantType:        c.ReasonTypeBeneficiaryAccountReason,
			wantDescription: c.SnapCoreResponseDormantAccountMessage,
		},
		{
			input: BankTransferResponseData{
				ResponseCode: "4040011",
			},
			wantStatus:      c.StatusFailed,
			wantType:        c.ReasonTypeBeneficiaryAccountReason,
			wantDescription: c.SnapCoreResponseInvalidAccountMessage,
		},
		{
			input: BankTransferResponseData{
				Status:          c.SnapCoreBankTransferStatusFailed,
				ResponseMessage: "Other error message",
			},
			wantStatus:      c.StatusFailed,
			wantType:        c.ReasonTypeOtherReason,
			wantDescription: "Other error message",
		},
	}
	for _, test := range tests {
		status, reasonType, reasonDescription := test.input.MappingAccountTransactionErrStatus()

		assert.Equal(t, test.wantStatus, status)
		assert.Equal(t, test.wantType, reasonType)
		assert.Equal(t, test.wantDescription, reasonDescription)
	}
}

func TestMappingInquiryTransactionStatus(t *testing.T) {
	tests := []struct {
		input           BankTransferResponseData
		wantStatus      string
		wantType        string
		wantDescription string
	}{
		{
			wantStatus: c.StatusPending,
		},
		{
			input: BankTransferResponseData{
				Status: c.StatusPending,
			},
			wantStatus: c.StatusPending,
		},
		{
			input: BankTransferResponseData{
				Status: c.StatusSuccess,
			},
			wantStatus: c.StatusSuccess,
		},
		{
			input: BankTransferResponseData{
				Status:          c.StatusFailed,
				ResponseMessage: c.ReasonDescInvalidBeneficiaryAccount,
			},
			wantStatus:      c.StatusFailed,
			wantType:        c.ReasonTypeOtherReason,
			wantDescription: c.ReasonDescInvalidBeneficiaryAccount,
		},
		{
			input: BankTransferResponseData{
				Status:          c.StatusFailed,
				ResponseCode:    "4030014",
				ResponseMessage: "TEST",
			},
			wantStatus:      c.StatusPending,
			wantType:        c.ReasonTypeInsufficientEscrowFund,
			wantDescription: "TEST",
		},
	}
	for _, test := range tests {
		status, reasonType, reasonDescription := test.input.MappingInquiryTransactionStatus()

		assert.Equal(t, test.wantStatus, status)
		assert.Equal(t, test.wantType, reasonType)
		assert.Equal(t, test.wantDescription, reasonDescription)
	}
}
