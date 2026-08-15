package disbursementModel_test

import (
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/disbursement"
	pb "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/proto/messages/disbursement"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestTransformArrayCreateSingleRequestToProtobufType(t *testing.T) {

	input := []CreateSingleRequest{
		{
			ReferenceID:            "REF001",
			BeneficiaryBankCode:    "001",
			BeneficiaryBankName:    "Bank Dummy",
			BeneficiaryAccountNo:   "0000001",
			BeneficiaryAccountName: "John",
			Amount:                 decimal.NewFromInt(125_000),
			Remark:                 "TEST",
			PurposeID:              "P",
			InquiryID:              "I",
		},
	}
	want := []*pb.CreateSingleRequest{
		{
			ReferenceId:            "REF001",
			BeneficiaryBankCode:    "001",
			BeneficiaryBankName:    "Bank Dummy",
			BeneficiaryAccountNo:   "0000001",
			BeneficiaryAccountName: "John",
			Amount:                 "125000",
			Remark:                 "TEST",
			PurposeId:              "P",
			InquiryId:              "I",
		},
	}
	assert.Equal(t, want, TransformArrayCreateSingleRequestToProtobufType(input))
}

func TestValidate(t *testing.T) {
	testCases := []struct {
		name    string
		input   GetDisbursementFilterRequest
		wantErr bool
	}{
		{
			name:  "SUCCESS: Valid Request",
			input: GetDisbursementFilterRequest{},
		},
		{
			name: "SUCCESS: Valid Disbursement Type",
			input: GetDisbursementFilterRequest{
				Type: constant.DisbursementTypeBulk,
			},
		},
		{
			name: "SUCCESS: Valid Disbursement Approval Status",
			input: GetDisbursementFilterRequest{
				Status: constant.DisbursementStatusWaiting,
			},
		},
		{
			name: "SUCCESS: Valid Disbursement Payment Status",
			input: GetDisbursementFilterRequest{
				TransactionStatus: constant.StatusSuccess,
			},
		},
		{
			name: "SUCCESS: Valid Disbursement Sort",
			input: GetDisbursementFilterRequest{
				SortBy: "updatedAt",
			},
		},
		{
			name: "ERROR: Invalid Disbursement Type",
			input: GetDisbursementFilterRequest{
				Type: "HULK",
			},
			wantErr: true,
		},
		{
			name: "ERROR: Invalid Disbursement Approval Status",
			input: GetDisbursementFilterRequest{
				Status: "UNKNOWN",
			},
			wantErr: true,
		},
		{
			name: "ERROR: Invalid Disbursement Payment Status",
			input: GetDisbursementFilterRequest{
				TransactionStatus: "UNKNOWN",
			},
			wantErr: true,
		},
		{
			name: "ERROR: Invalid Disbursement Sort",
			input: GetDisbursementFilterRequest{
				SortBy: "updatedAts",
			},
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.input.Validate()
			if tc.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}

}
