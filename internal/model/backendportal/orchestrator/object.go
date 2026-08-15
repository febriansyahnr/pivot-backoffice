package orchestrator_model

import (
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/fee"
	settlementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/settlement"
)

type FeeTransactionMetadataObject struct {
	feeModel.FeeMetadataObject
	*settlementModel.AccountTransactionMetadataObject
}
