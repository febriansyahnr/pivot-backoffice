package unifiedPaymentModel

import commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"

type GetPaymentMethodConfigResponse struct {
	Card           *GetPaymentMethodConfigCard           `json:"card"`
	VirtualAccount *GetPaymentMethodConfigVirtualAccount `json:"virtualAccount"`
	Qr             *GetPaymentMethodConfigQr             `json:"qr"`
	Ewallet        *GetPaymentMethodConfigEWallet        `json:"ewallet,omitempty"`
	Installment    *GetPaymentMethodConfigInstallment    `json:"installment,omitempty"`
}

type GetPaymentMethodConfigCard struct {
	Enabled           bool        `json:"enabled"`
	AcceptedChannels  []string    `json:"acceptedChannels"`
	MinimumAmount     *Amount     `json:"minimumAmount"`
	MaximumAmount     *Amount     `json:"maximumAmount"`
	MaximumExpiry     string      `json:"maximumExpiry"`
	InstallmentConfig interface{} `json:"installmentConfig"`
}

type GetPaymentMethodConfigVirtualAccount struct {
	Enabled          bool     `json:"enabled"`
	AcceptedChannels []string `json:"acceptedChannels"`
	MinimumAmount    *Amount  `json:"minimumAmount"`
	MaximumAmount    *Amount  `json:"maximumAmount"`
	MaximumExpiry    string   `json:"maximumExpiry"`
}

type GetPaymentMethodConfigQr struct {
	Enabled       bool    `json:"enabled"`
	MinimumAmount *Amount `json:"minimumAmount"`
	MaximumAmount *Amount `json:"maximumAmount"`
	MaximumExpiry string  `json:"maximumExpiry"`
}

type GetPaymentMethodConfigEWallet struct {
	Enabled          bool     `json:"enabled"`
	AcceptedChannels []string `json:"acceptedChannels"`
	MinimumAmount    *Amount  `json:"minimumAmount"`
	MaximumAmount    *Amount  `json:"maximumAmount"`
	MaximumExpiry    string   `json:"maximumExpiry"`
}

type GetPaymentMethodConfigInstallment struct {
	Enabled          bool                                      `json:"enabled"`
	AcceptedChannels map[string][]*InstallmentAcceptedChannels `json:"acceptedChannels"` // bank - channel list
}

type InstallmentAcceptedChannels struct {
	ID            string              `json:"id"`
	ProgramName   string              `json:"programName"`
	Tenure        string              `json:"tenure"`
	Interest      float64             `json:"interest"`
	AllowedBins   []string            `json:"allowedBin"`
	MinimumAmount commonModel.Amount2 `json:"minimumAmount"`
	MaximumAmount commonModel.Amount2 `json:"maximumAmount"`
}
