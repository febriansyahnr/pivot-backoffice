package withdrawal_test

import (
	"testing"
	"time"

	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	. "github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"

	"github.com/jmoiron/sqlx/types"
	"github.com/stretchr/testify/assert"
)

func TestWithdrawalDetailResponse(t *testing.T) {
	tests := []struct {
		name       string
		response   WithdrawalDetailResponse
		wantResult OpenAPIWithdrawalResponse
	}{
		{
			name: "Without metadata",
			response: WithdrawalDetailResponse{
				Id:                     "01991460-38b6-7860-901c-317accd1cd85",         // NOSONAR
				CreatedAt:              time.Date(2025, 9, 4, 10, 57, 54, 0, time.UTC), // NOSONAR
				UpdatedAt:              time.Date(2025, 9, 4, 10, 57, 54, 0, time.UTC), // NOSONAR
				Type:                   "MANUAL",                                       // NOSONAR
				Amount:                 10_000,                                         // NOSONAR
				Status:                 "SUCCESS",                                      // NOSONAR
				BeneficiaryAccountName: "Payout Balance",                               // NOSONAR
				Currency:               "IDR",                                          // NOSONAR
				MerchantID:             "aec6636d-7a02-4d93-a4c5-006b9c235068",         // NOSONAR
			},
			wantResult: OpenAPIWithdrawalResponse{
				ID:         "01991460-38b6-7860-901c-317accd1cd85", // NOSONAR
				MerchantId: "aec6636d-7a02-4d93-a4c5-006b9c235068", // NOSONAR
				Withdrawal: OpenAPIWithdrawalDetailResponse{
					ReferenceID:  "",                 // NOSONAR
					WithdrawType: "BALANCE_TRANSFER", // NOSONAR
					BalanceType:  "PAYOUT_BALANCE",   // NOSONAR
					IsFullAmount: false,              // NOSONAR
					Amount: &commonModel.Amount{
						Currency: "IDR",   // NOSONAR
						Value:    "10000", // NOSONAR
					},
				},
				Status:    "SUCCESS",              // NOSONAR
				CreatedAt: "2025-09-04T10:57:54Z", // NOSONAR
				UpdatedAt: "2025-09-04T10:57:54Z", // NOSONAR
			},
		},
		{
			name: "with metadata",
			response: WithdrawalDetailResponse{
				Id:          "01991460-38b6-7860-901c-317accd1cd85",         // NOSONAR
				CreatedAt:   time.Date(2025, 9, 4, 10, 57, 54, 0, time.UTC), // NOSONAR
				UpdatedAt:   time.Date(2025, 9, 4, 10, 57, 54, 0, time.UTC), // NOSONAR
				Type:        "MANUAL",                                       // NOSONAR
				Amount:      10_000,                                         // NOSONAR
				Status:      "SUCCESS",                                      // NOSONAR
				Currency:    "IDR",                                          // NOSONAR
				MerchantID:  "aec6636d-7a02-4d93-a4c5-006b9c235068",         // NOSONAR
				RawMetadata: types.NullJSONText{Valid: true, JSONText: []byte(`{"source": "OPEN_API", "balanceType": "", "isFullAmount": true, "withdrawType": "BANK_TRANSFER"}`)},
			},
			wantResult: OpenAPIWithdrawalResponse{
				ID:         "01991460-38b6-7860-901c-317accd1cd85", // NOSONAR
				MerchantId: "aec6636d-7a02-4d93-a4c5-006b9c235068", // NOSONAR
				Withdrawal: OpenAPIWithdrawalDetailResponse{
					ReferenceID:  "",              // NOSONAR
					WithdrawType: "BANK_TRANSFER", // NOSONAR
					BalanceType:  "",              // NOSONAR
					IsFullAmount: true,            // NOSONAR
					Amount: &commonModel.Amount{
						Currency: "IDR",   // NOSONAR
						Value:    "10000", // NOSONAR
					},
				},
				Status:    "SUCCESS",              // NOSONAR
				CreatedAt: "2025-09-04T10:57:54Z", // NOSONAR
				UpdatedAt: "2025-09-04T10:57:54Z", // NOSONAR
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.wantResult, test.response.ToOpenAPIWithdrawalResponse())
		})
	}
}
