package cardFundedPayoutService

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/statusHistory"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/jmoiron/sqlx/types"
)

var statusHistoryMaps = map[string]struct {
	label          string
	description    string
	recommendation string
}{
	constant.DisbursementStatusWaiting:           {"Payout Submitted", "Payout request confirmed.", ""},
	constant.DisbursementStatusRejected:          {"Rejected", "Payout is rejected by %s.", ""},
	constant.DisbursementStatusApproved:          {"Approved", "Payout is approved by %s.", ""},
	constant.DisbursementStatusHistoryProcessing: {"Processing", "Payout is being processed by our bank partner.", ""},
	constant.DisbursementStatusHistorySuccess:    {"Success", "Funds are being transferred to the recipient bank.", ""},
	constant.DisbursementStatusHistoryFailed:     {"Failed", "We couldn't transfer the funds to the recipient bank. Please try again.", ""},
}

func (s *service) recordStatusHistory(ctx context.Context, referenceID, status, actorID, actorName string) error {
	metadata := &model.StatusHistoryMetadata{
		Actor:       actorID,
		Label:       util.ToTitle(status),
		Description: "Transaction status changed.",
	}
	if info, ok := statusHistoryMaps[status]; ok {
		metadata.Label = info.label
		metadata.Description = info.description
		metadata.Recommendation = info.recommendation
		if actorName != "" {
			metadata.Description = fmt.Sprintf(metadata.Description, actorName)
		}
	}
	rawMetadata, _ := json.Marshal(metadata)

	now := time.Now().UTC()
	return s.statusHistoryRepo.Insert(ctx, &model.StatusHistory{
		ID:            util.GenerateUUID().String(),
		ReferenceType: constant.DisbursementTypeCardFundedPayout,
		ReferenceID:   referenceID,
		Status:        status,
		MetadataObj:   metadata,
		Metadata:      types.NullJSONText{Valid: true, JSONText: rawMetadata},
		CreatedAt:     now,
		UpdatedAt:     now,
	})
}

func (s *service) recordPaymentFailedStatus(ctx context.Context, referenceID string) error {
	metadata := model.StatusHistoryMetadata{
		Label:       "Failed",
		Description: "Charge has failed.",
		Actor:       constant.StatusHistoryActorSystem,
	}
	now := time.Now().UTC()
	metadataBytes, _ := json.Marshal(metadata)

	return s.statusHistoryRepo.Insert(ctx, &model.StatusHistory{
		ID:            util.GenerateUUID().String(),
		ReferenceType: constant.TypePayment,
		ReferenceID:   referenceID,
		Status:        constant.PaymentStatusHistoryFailed,
		MetadataObj:   &metadata,
		Metadata:      types.NullJSONText{Valid: true, JSONText: metadataBytes},
		CreatedAt:     now,
		UpdatedAt:     now,
	})
}
