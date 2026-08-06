package xbPayoutController

import (
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	statusHistoryModel "github.com/paper-indonesia/pivot-backoffice/internal/model/statusHistory"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"

	"github.com/stretchr/testify/assert"
)

func TestBuildStatusHistories(t *testing.T) {

	tests := []struct {
		name          string
		histories     []*statusHistoryModel.StatusHistory
		wantHistories []xbModel.XbPayoutStatusHistoryResponse
	}{
		{
			name:          "Nil histories",
			histories:     nil,
			wantHistories: []xbModel.XbPayoutStatusHistoryResponse{},
		},
		{
			name: "Only payout created",
			histories: []*statusHistoryModel.StatusHistory{
				{
					MetadataObj: &statusHistoryModel.StatusHistoryMetadata{
						Label: "Payout Created",
					},
				},
			},
			wantHistories: []xbModel.XbPayoutStatusHistoryResponse{},
		},
		{
			name: "Payout expired",
			histories: []*statusHistoryModel.StatusHistory{
				{
					MetadataObj: &statusHistoryModel.StatusHistoryMetadata{Label: "Payout Created"},
				},
				{Status: constant.XbStatusExpired},
			},
			wantHistories: []xbModel.XbPayoutStatusHistoryResponse{},
		},
		{
			name: "Payout confirmed",
			histories: []*statusHistoryModel.StatusHistory{
				{MetadataObj: &statusHistoryModel.StatusHistoryMetadata{Label: "Payout Created"}},
				{
					MetadataObj: &statusHistoryModel.StatusHistoryMetadata{
						Label:       "Payout Created",            // NOSONAR
						Description: "Payout request confirmed.", // NOSONAR
					},
				},
			},
			wantHistories: []xbModel.XbPayoutStatusHistoryResponse{
				{
					Label:       "Payout Created",            // NOSONAR
					Description: "Payout request confirmed.", // NOSONAR
				},
			},
		},
		{
			name: "Payout processing",
			histories: []*statusHistoryModel.StatusHistory{
				{MetadataObj: &statusHistoryModel.StatusHistoryMetadata{Label: "Payout Created"}},
				{
					MetadataObj: &statusHistoryModel.StatusHistoryMetadata{
						Label:       "Payout Created",            // NOSONAR
						Description: "Payout request confirmed.", // NOSONAR
					},
				},
				{
					MetadataObj: &statusHistoryModel.StatusHistoryMetadata{
						Label:       "Processing",                                          // NOSONAR
						Description: "Transaction is being processed by our bank partner.", // NOSONAR
					},
				},
				{
					MetadataObj: &statusHistoryModel.StatusHistoryMetadata{
						Label:       "Processing",                                          // NOSONAR
						Description: "Transaction is being processed by our bank partner.", // NOSONAR
					},
				},
				{
					MetadataObj: &statusHistoryModel.StatusHistoryMetadata{
						Label:       "Processing",                                          // NOSONAR
						Description: "Transaction is being processed by our bank partner.", // NOSONAR
					},
					CreatedAt: time.Date(2025, 12, 10, 9, 12, 10, 0, time.UTC),
				},
			},
			wantHistories: []xbModel.XbPayoutStatusHistoryResponse{
				{
					Label:       "Payout Created",            // NOSONAR
					Description: "Payout request confirmed.", // NOSONAR
				},
				{
					Label:       "Processing",                                          // NOSONAR
					Description: "Transaction is being processed by our bank partner.", // NOSONAR
					Timestamp:   time.Date(2025, 12, 10, 9, 12, 10, 0, time.UTC),
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.wantHistories, buildStatusHistories(test.histories))
		})
	}
}
