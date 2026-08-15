package orchestrator_model_test

import (
	"database/sql"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/common"
	. "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/orchestrator"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestModelTransactionHistoryResponse(t *testing.T) {
	id, _ := uuid.Parse("e39d1d21-0a7b-49e5-8cd9-404ac75d54be")

	// Store the time in a variable
	settlementTime := util.TimeNow

	response := ToTransactionHistoryResponse(&AccountTransactionWithUseCase{
		UUID:                   id,
		ReferenceID:            "REF-0001",
		MerchantID:             id,
		AccountID:              id,
		Currency:               "IDR",
		Credit:                 125_500,
		Type:                   constant.TypePayment,
		Remarks:                "NOTES",
		SenderName:             "John Wick",
		Fee:                    1_500,
		BeneficiaryAccountNo:   "123456",
		BeneficiaryAccountName: "Bejo",
		BeneficiaryBankName:    "Bejo PG",
		ClientReferenceID:      "CLIENT-0001",
		SettlementAt:           sql.NullTime{Time: settlementTime, Valid: true},
	})
	assert.Equal(t, TransactionHistoryResponse{
		ID:                     id.String(),
		TrxType:                constant.TypePayment,
		Amount:                 125_500,
		BeneficiaryAccountName: "Bejo",
		TrxID:                  "REF-0001",
		SettlementAt:           &settlementTime,
		Fee:                    1_500,
	}, response)
}

func TestModelToTransactionHistoryOpenApiResponse(t *testing.T) {
	id, _ := uuid.Parse("4c427ab9-539c-493f-9062-5df00be50b57")

	// Store the time in a variable
	settlementTime := util.TimeNow

	response := ToTransactionHistoryOpenApiResponse(&AccountTransactionWithUseCase{
		UUID:                   id,
		MerchantID:             id,
		AccountID:              id,
		Currency:               "IDR",
		Credit:                 125_500,
		Type:                   "BULK_DISBURSEMENT",
		Remarks:                "NOTES",
		SenderName:             "John Wick",
		Fee:                    1_500,
		BeneficiaryAccountNo:   "123456",
		BeneficiaryAccountName: "Bejo",
		BeneficiaryBankName:    "Bejo PG",
		ClientReferenceID:      "CLIENT-0001",
		SettlementAt:           sql.NullTime{Time: settlementTime, Valid: true},
	})
	assert.Equal(t, TransactionHistoryOpenApiResponse{
		ID:      id.String(),
		TrxType: "Bulk Payout",
		Amount: commonModel.Amount2{
			Value:    125_500,
			Currency: "IDR",
		},
		BeneficiaryAccountName: "Bejo",
		SettlementAt:           &settlementTime,
		Fee: commonModel.Amount2{
			Value:    1_500,
			Currency: "IDR",
		},
	}, response)
}
