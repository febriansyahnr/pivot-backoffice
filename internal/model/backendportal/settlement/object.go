package settlementModel

import (
	"time"

	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/merchant"
)

type AccountTransactionMetadataObject struct {
	SettlementStatus string                         `json:"settlementStatus"`
	SettlementAt     *time.Time                     `json:"settlementAt,omitempty"`
	SettlementDetail merchantModel.SettlementConfig `json:"settlementDetail"`
	ReconReferenceNo string                         `json:"reconReferenceNo"`
}
