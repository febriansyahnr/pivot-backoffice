package orchestrator_model

import (
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	settlementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/settlement"
)

type FeeTransactionMetadataObject struct {
	feeModel.FeeMetadataObject
	*settlementModel.AccountTransactionMetadataObject
}
