package flipProcessorModel

import (
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	routingProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/bankTransfer"
	"github.com/stretchr/testify/assert"
)

func TestToBankTransferResponse(t *testing.T) {
	datetime := "2025-01-23 23:23:00"
	location, _ := time.LoadLocation("Asia/Jakarta")
	timestamp, _ := time.ParseInLocation(constant.DatetimeFormat, datetime, location)

	bt := BankTransferResponse{
		ID:             12898999,
		IdempotencyKey: "test-idempotency-key",
		AccountNo:      "test-account-no",
		BankCode:       "bri",
		Amount:         1000000,
		Timestamp:      datetime,
	}

	var metadata map[string]any
	btB, _ := json.Marshal(bt)
	json.Unmarshal(btB, &metadata)

	wantBt := &routingProcessorModel.BankTransferResponseData{
		ResponseCode:         "2001800",
		ResponseMessage:      "Successful",
		UUID:                 strconv.Itoa(bt.ID),
		BankReferenceNo:      bt.IdempotencyKey,
		BeneficiaryAccountNo: bt.AccountNo,
		BeneficiaryBankCode:  bt.BankCode,
		ExternalID:           bt.IdempotencyKey,
		Amount: commonModel.Amount{
			Value:    strconv.Itoa(bt.Amount),
			Currency: "IDR",
		},
		Metadata:           metadata,
		ProcessorReference: constant.FlipPGProcessor,
		TransactionDate:    timestamp.UTC(),
	}

	testCases := []struct {
		desc string
		bt   *BankTransferResponse
		want *routingProcessorModel.BankTransferResponseData
	}{
		{
			desc: "success wrapping request",
			bt:   &bt,
			want: wantBt,
		},
		{
			desc: "status DONE should be mapped to SUCCESS",
			bt: &BankTransferResponse{
				ID:             12898999,
				IdempotencyKey: "test-idempotency-key",
				AccountNo:      "test-account-no",
				BankCode:       "bri",
				Amount:         1000000,
				Timestamp:      datetime,
				Status:         constant.FlipBankTransferStatusDone,
			},
			want: func() *routingProcessorModel.BankTransferResponseData {
				btDone := BankTransferResponse{
					ID:             12898999,
					IdempotencyKey: "test-idempotency-key",
					AccountNo:      "test-account-no",
					BankCode:       "bri",
					Amount:         1000000,
					Timestamp:      datetime,
					Status:         constant.FlipBankTransferStatusDone,
				}
				var metadataDone map[string]any
				btDoneB, _ := json.Marshal(btDone)
				json.Unmarshal(btDoneB, &metadataDone)

				return &routingProcessorModel.BankTransferResponseData{
					ResponseCode:         "2001800",
					ResponseMessage:      "Successful",
					UUID:                 "12898999",
					BankReferenceNo:      "test-idempotency-key",
					BeneficiaryAccountNo: "test-account-no",
					BeneficiaryBankCode:  "bri",
					ExternalID:           "test-idempotency-key",
					Status:               constant.SnapCoreBankTransferStatusSuccess,
					Amount: commonModel.Amount{
						Value:    "1000000",
						Currency: "IDR",
					},
					ProcessorReference: constant.FlipPGProcessor,
					Metadata:           metadataDone,
					TransactionDate:    timestamp.UTC(),
				}
			}(),
		},
		{
			desc: "status CANCELLED should be mapped to FAILED",
			bt: &BankTransferResponse{
				ID:             12898999,
				IdempotencyKey: "test-idempotency-key",
				AccountNo:      "test-account-no",
				BankCode:       "bri",
				Amount:         1000000,
				Timestamp:      datetime,
				Status:         constant.FlipBankTransferStatusCancelled,
			},
			want: func() *routingProcessorModel.BankTransferResponseData {
				btCancelled := BankTransferResponse{
					ID:             12898999,
					IdempotencyKey: "test-idempotency-key",
					AccountNo:      "test-account-no",
					BankCode:       "bri",
					Amount:         1000000,
					Timestamp:      datetime,
					Status:         constant.FlipBankTransferStatusCancelled,
				}
				var metadataCancelled map[string]any
				btCancelledB, _ := json.Marshal(btCancelled)
				json.Unmarshal(btCancelledB, &metadataCancelled)

				return &routingProcessorModel.BankTransferResponseData{
					ResponseCode:         "2001800",
					ResponseMessage:      "Successful",
					UUID:                 "12898999",
					BankReferenceNo:      "test-idempotency-key",
					BeneficiaryAccountNo: "test-account-no",
					BeneficiaryBankCode:  "bri",
					ExternalID:           "test-idempotency-key",
					Status:               constant.SnapCoreBankTransferStatusFailed,
					Amount: commonModel.Amount{
						Value:    "1000000",
						Currency: "IDR",
					},
					ProcessorReference: constant.FlipPGProcessor,
					Metadata:           metadataCancelled,
					TransactionDate:    timestamp.UTC(),
				}
			}(),
		},
		{
			desc: "status PENDING should be mapped to PENDING",
			bt: &BankTransferResponse{
				ID:             12898999,
				IdempotencyKey: "test-idempotency-key",
				AccountNo:      "test-account-no",
				BankCode:       "bri",
				Amount:         1000000,
				Timestamp:      datetime,
				Status:         constant.FlipBankTransferStatusPending,
			},
			want: func() *routingProcessorModel.BankTransferResponseData {
				btPending := BankTransferResponse{
					ID:             12898999,
					IdempotencyKey: "test-idempotency-key",
					AccountNo:      "test-account-no",
					BankCode:       "bri",
					Amount:         1000000,
					Timestamp:      datetime,
					Status:         constant.FlipBankTransferStatusPending,
				}
				var metadataPending map[string]any
				btPendingB, _ := json.Marshal(btPending)
				json.Unmarshal(btPendingB, &metadataPending)

				return &routingProcessorModel.BankTransferResponseData{
					ResponseCode:         "2001800",
					ResponseMessage:      "Successful",
					UUID:                 "12898999",
					BankReferenceNo:      "test-idempotency-key",
					BeneficiaryAccountNo: "test-account-no",
					BeneficiaryBankCode:  "bri",
					ExternalID:           "test-idempotency-key",
					Status:               constant.SnapCoreBankTransferStatusPending,
					Amount: commonModel.Amount{
						Value:    "1000000",
						Currency: "IDR",
					},
					ProcessorReference: constant.FlipPGProcessor,
					Metadata:           metadataPending,
					TransactionDate:    timestamp.UTC(),
				}
			}(),
		},
		{
			desc: "reason INACTIVE_ACCOUNT should set appropriate response code",
			bt: &BankTransferResponse{
				ID:             12898999,
				IdempotencyKey: "test-idempotency-key",
				AccountNo:      "test-account-no",
				BankCode:       "bri",
				Amount:         1000000,
				Timestamp:      datetime,
				Reason:         constant.FlipReasonInactiveAccount,
			},
			want: func() *routingProcessorModel.BankTransferResponseData {
				btInactive := BankTransferResponse{
					ID:             12898999,
					IdempotencyKey: "test-idempotency-key",
					AccountNo:      "test-account-no",
					BankCode:       "bri",
					Amount:         1000000,
					Timestamp:      datetime,
					Reason:         constant.FlipReasonInactiveAccount,
				}
				var metadataInactive map[string]any
				btInactiveB, _ := json.Marshal(btInactive)
				json.Unmarshal(btInactiveB, &metadataInactive)

				return &routingProcessorModel.BankTransferResponseData{
					ResponseCode:         "4031818",
					ResponseMessage:      "Inactive Account",
					UUID:                 "12898999",
					BankReferenceNo:      "test-idempotency-key",
					BeneficiaryAccountNo: "test-account-no",
					BeneficiaryBankCode:  "bri",
					ExternalID:           "test-idempotency-key",
					Reason:               constant.FlipReasonInactiveAccount,
					Amount: commonModel.Amount{
						Value:    "1000000",
						Currency: "IDR",
					},
					ProcessorReference: constant.FlipPGProcessor,
					Metadata:           metadataInactive,
					TransactionDate:    timestamp.UTC(),
				}
			}(),
		},
		{
			desc: "reason INVALID_ACCOUNT should set appropriate response code",
			bt: &BankTransferResponse{
				ID:             12898999,
				IdempotencyKey: "test-idempotency-key",
				AccountNo:      "test-account-no",
				BankCode:       "bri",
				Amount:         1000000,
				Timestamp:      datetime,
				Reason:         constant.FlipReasonInvalidAccount,
			},
			want: func() *routingProcessorModel.BankTransferResponseData {
				btInactive := BankTransferResponse{
					ID:             12898999,
					IdempotencyKey: "test-idempotency-key",
					AccountNo:      "test-account-no",
					BankCode:       "bri",
					Amount:         1000000,
					Timestamp:      datetime,
					Reason:         constant.FlipReasonInvalidAccount,
				}
				var metadataInactive map[string]any
				btInactiveB, _ := json.Marshal(btInactive)
				json.Unmarshal(btInactiveB, &metadataInactive)

				return &routingProcessorModel.BankTransferResponseData{
					ResponseCode:         "4031811",
					ResponseMessage:      "Invalid Account",
					UUID:                 "12898999",
					BankReferenceNo:      "test-idempotency-key",
					BeneficiaryAccountNo: "test-account-no",
					BeneficiaryBankCode:  "bri",
					ExternalID:           "test-idempotency-key",
					Reason:               constant.FlipReasonInvalidAccount,
					Amount: commonModel.Amount{
						Value:    "1000000",
						Currency: "IDR",
					},
					ProcessorReference: constant.FlipPGProcessor,
					Metadata:           metadataInactive,
					TransactionDate:    timestamp.UTC(),
				}
			}(),
		},
		{
			desc: "reason INSUFFICIENT_BALANCE should set appropriate response code",
			bt: &BankTransferResponse{
				ID:             12898999,
				IdempotencyKey: "test-idempotency-key",
				AccountNo:      "test-account-no",
				BankCode:       "bri",
				Amount:         1000000,
				Timestamp:      datetime,
				Reason:         constant.FlipReasonInsufficientBalance,
			},
			want: func() *routingProcessorModel.BankTransferResponseData {
				btInsufficient := BankTransferResponse{
					ID:             12898999,
					IdempotencyKey: "test-idempotency-key",
					AccountNo:      "test-account-no",
					BankCode:       "bri",
					Amount:         1000000,
					Timestamp:      datetime,
					Reason:         constant.FlipReasonInsufficientBalance,
				}
				var metadataInsufficient map[string]any
				btInsufficientB, _ := json.Marshal(btInsufficient)
				json.Unmarshal(btInsufficientB, &metadataInsufficient)

				return &routingProcessorModel.BankTransferResponseData{
					ResponseCode:         "4031814",
					ResponseMessage:      "Insufficient Fund",
					UUID:                 "12898999",
					BankReferenceNo:      "test-idempotency-key",
					BeneficiaryAccountNo: "test-account-no",
					BeneficiaryBankCode:  "bri",
					ExternalID:           "test-idempotency-key",
					Reason:               constant.FlipReasonInsufficientBalance,
					Amount: commonModel.Amount{
						Value:    "1000000",
						Currency: "IDR",
					},
					ProcessorReference: constant.FlipPGProcessor,
					Metadata:           metadataInsufficient,
					TransactionDate:    timestamp.UTC(),
				}
			}(),
		},
		{
			desc: "reason INVALID_AMOUNT should set appropriate response code",
			bt: &BankTransferResponse{
				ID:             12898999,
				IdempotencyKey: "test-idempotency-key",
				AccountNo:      "test-account-no",
				BankCode:       "bri",
				Amount:         1000000,
				Timestamp:      datetime,
				Reason:         constant.FlipReasonInvalidAmount,
			},
			want: func() *routingProcessorModel.BankTransferResponseData {
				btInvalidAmount := BankTransferResponse{
					ID:             12898999,
					IdempotencyKey: "test-idempotency-key",
					AccountNo:      "test-account-no",
					BankCode:       "bri",
					Amount:         1000000,
					Timestamp:      datetime,
					Reason:         constant.FlipReasonInvalidAmount,
				}
				var metadataInvalidAmount map[string]any
				btInvalidAmountB, _ := json.Marshal(btInvalidAmount)
				json.Unmarshal(btInvalidAmountB, &metadataInvalidAmount)

				return &routingProcessorModel.BankTransferResponseData{
					ResponseCode:         "4041813",
					ResponseMessage:      "Invalid Amount",
					UUID:                 "12898999",
					BankReferenceNo:      "test-idempotency-key",
					BeneficiaryAccountNo: "test-account-no",
					BeneficiaryBankCode:  "bri",
					ExternalID:           "test-idempotency-key",
					Reason:               constant.FlipReasonInvalidAmount,
					Amount: commonModel.Amount{
						Value:    "1000000",
						Currency: "IDR",
					},
					ProcessorReference: constant.FlipPGProcessor,
					Metadata:           metadataInvalidAmount,
					TransactionDate:    timestamp.UTC(),
				}
			}(),
		},
		{
			desc: "reason DORMANT_ACCOUNT should set appropriate response code",
			bt: &BankTransferResponse{
				ID:             12898999,
				IdempotencyKey: "test-idempotency-key",
				AccountNo:      "test-account-no",
				BankCode:       "bri",
				Amount:         1000000,
				Timestamp:      datetime,
				Reason:         constant.FlipReasonDormantAccount,
			},
			want: func() *routingProcessorModel.BankTransferResponseData {
				btDormant := BankTransferResponse{
					ID:             12898999,
					IdempotencyKey: "test-idempotency-key",
					AccountNo:      "test-account-no",
					BankCode:       "bri",
					Amount:         1000000,
					Timestamp:      datetime,
					Reason:         constant.FlipReasonDormantAccount,
				}
				var metadataDormant map[string]any
				btDormantB, _ := json.Marshal(btDormant)
				json.Unmarshal(btDormantB, &metadataDormant)

				return &routingProcessorModel.BankTransferResponseData{
					ResponseCode:         "4031809",
					ResponseMessage:      "Dormant Account",
					UUID:                 "12898999",
					BankReferenceNo:      "test-idempotency-key",
					BeneficiaryAccountNo: "test-account-no",
					BeneficiaryBankCode:  "bri",
					ExternalID:           "test-idempotency-key",
					Reason:               constant.FlipReasonDormantAccount,
					Amount: commonModel.Amount{
						Value:    "1000000",
						Currency: "IDR",
					},
					ProcessorReference: constant.FlipPGProcessor,
					Metadata:           metadataDormant,
					TransactionDate:    timestamp.UTC(),
				}
			}(),
		},
		{
			desc: "invalid timestamp should use current time",
			bt: &BankTransferResponse{
				ID:             12898999,
				IdempotencyKey: "test-idempotency-key",
				AccountNo:      "test-account-no",
				BankCode:       "bri",
				Amount:         1000000,
				Timestamp:      "invalid-timestamp",
			},
			want: func() *routingProcessorModel.BankTransferResponseData {
				btInvalidTime := BankTransferResponse{
					ID:             12898999,
					IdempotencyKey: "test-idempotency-key",
					AccountNo:      "test-account-no",
					BankCode:       "bri",
					Amount:         1000000,
					Timestamp:      "invalid-timestamp",
				}
				var metadataInvalidTime map[string]any
				btInvalidTimeB, _ := json.Marshal(btInvalidTime)
				json.Unmarshal(btInvalidTimeB, &metadataInvalidTime)

				return &routingProcessorModel.BankTransferResponseData{
					ResponseCode:         "2001800",
					ResponseMessage:      "Successful",
					UUID:                 "12898999",
					BankReferenceNo:      "test-idempotency-key",
					BeneficiaryAccountNo: "test-account-no",
					BeneficiaryBankCode:  "bri",
					ExternalID:           "test-idempotency-key",
					Amount: commonModel.Amount{
						Value:    "1000000",
						Currency: "IDR",
					},
					ProcessorReference: constant.FlipPGProcessor,
					Metadata:           metadataInvalidTime,
					// We can't assert exact time since it will be current time
				}
			}(),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got := tC.bt.ToBankTransferResponse()

			// For the invalid timestamp case, we need special handling
			if tC.desc == "invalid timestamp should use current time" {
				// Check everything except the transaction date
				assert.Equal(t, tC.want.ResponseCode, got.ResponseCode)
				assert.Equal(t, tC.want.ResponseMessage, got.ResponseMessage)
				assert.Equal(t, tC.want.UUID, got.UUID)
				assert.Equal(t, tC.want.BankReferenceNo, got.BankReferenceNo)
				assert.Equal(t, tC.want.BeneficiaryAccountNo, got.BeneficiaryAccountNo)
				assert.Equal(t, tC.want.BeneficiaryBankCode, got.BeneficiaryBankCode)
				assert.Equal(t, tC.want.ExternalID, got.ExternalID)
				assert.Equal(t, tC.want.Amount, got.Amount)
				assert.Equal(t, tC.want.ProcessorReference, got.ProcessorReference)

				// Verify that the transaction date is close to now
				now := time.Now().UTC()
				diff := now.Sub(got.TransactionDate)
				assert.LessOrEqual(t, diff.Seconds(), float64(5), "Transaction date should be close to current time")
			} else {
				assert.Equal(t, tC.want, got)
			}
		})
	}
}
