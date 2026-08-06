package snapPaymentModel

import commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"

type PublishRequest struct {
	PaymentMethod     string             `json:"payment_method"`
	InternalReference string             `json:"internal_reference"`
	PartnerReference  string             `json:"partner_reference"`
	Amount            commonModel.Amount `json:"amount"`
	ForceSuccess      bool               `json:"force_success"`
}
