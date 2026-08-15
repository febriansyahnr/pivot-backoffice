package cdcModel_test

import (
	"encoding/json"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/cdc"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/stretchr/testify/assert"
)

func TestAccountTransactionUnmarshalJSON(t *testing.T) {
	tests := []struct {
		input      string
		wantError  error
		wantResult AccountTransaction
	}{
		{
			input: `B`,
			wantError: func() error {
				return json.Unmarshal([]byte("B"), &struct{}{})
			}(),
		},
		{
			input: `{"additional_info": "{\"amount\":0"}`,
			wantError: func() error {
				return json.Unmarshal([]byte("{\"amount\":0"), &struct{}{})
			}(),
		},
		{
			input:      `{"uuid":"9c5df4be-4c8c-4fbc-8746-edff4c52061e","additional_info": null}`,
			wantResult: AccountTransaction{UUID: "9c5df4be-4c8c-4fbc-8746-edff4c52061e"},
		},
		{
			input: `{"uuid":"5cd39d84-b1e9-4415-834f-fc3596894d1f","additional_info": "{\"amount\":0,\"amountType\":\"PERCENTAGE\",\"deductionType\":\"MANUAL\",\"finalAmount\":8,\"maxFeeAmount\":10000,\"method\":\"\",\"percentage\":0.5,\"referenceType\":\"\",\"taxAmount\":0,\"taxPercentage\":0,\"taxType\":\"NON_PKP\",\"trxAmount\":1500,\"type\":\"PLATFORM_TRANSACTION\"}"}`,
			wantResult: AccountTransaction{
				UUID:           "5cd39d84-b1e9-4415-834f-fc3596894d1f",
				AdditionalInfo: AccountTransactionAdditionalInfo{Type: "PLATFORM_TRANSACTION"},
			},
		},
	}
	for _, test := range tests {

		result := AccountTransaction{}

		assert.Equal(t, test.wantError, json.Unmarshal([]byte(test.input), &result))
		assert.Equal(t, test.wantResult, result)
	}
}

func TestAccountTransactionGetSettlementModel(t *testing.T) {
	tests := []struct {
		name        string
		transaction AccountTransaction
		wantResult  string
	}{
		{
			name:       "Settlement model is nil",
			wantResult: constant.SettlementModelAggregator,
		},
		{
			name: "Settlement model is aggregator",
			transaction: AccountTransaction{
				SettlementModel: util.ValueToPtr(constant.SettlementModelAggregator),
			},
			wantResult: constant.SettlementModelAggregator,
		},
		{
			name: "Settlement model is facilitator",
			transaction: AccountTransaction{
				SettlementModel: util.ValueToPtr(constant.SettlementModelFacilitator),
			},
			wantResult: constant.SettlementModelDirect,
		},
		{
			name: "Settlement model is direct",
			transaction: AccountTransaction{
				SettlementModel: util.ValueToPtr(constant.SettlementModelDirect),
			},
			wantResult: constant.SettlementModelDirect,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantResult, tt.transaction.GetSettlementModel())
		})
	}
}
