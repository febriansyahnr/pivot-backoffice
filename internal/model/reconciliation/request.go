package reconciliation

import (
	"fmt"
	"strings"

	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
)

type ReconciliationFilterRequest struct {
	Status  string
	PerPage int64
	Page    int64
}

func (r *ReconciliationFilterRequest) Query() string {
	queryArr := []string{}

	if r.Status != "" {
		queryArr = append(queryArr, fmt.Sprintf("status = '%s'", r.Status))
	}

	query := strings.Join(queryArr, " AND ")

	if query != "" {
		query = "WHERE " + query
	}

	return query
}

type ReconDetailRequest struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Reason string `json:"reason"`
}

type ReconciliationPayout struct {
	UUID                   string              `json:"uuid"`
	ExternalID             string              `json:"externalId"`
	PartnerReferenceNo     string              `json:"partnerReferenceNo"`
	Acquirer               string              `json:"acquirer"`
	Status                 string              `json:"status"`
	Reason                 string              `json:"reason"`
	Amount                 *commonModel.Amount `json:"amount"`
	ProcessorReferenceName string              `json:"processorReferenceName"`
}
