package refundModel

import "github.com/paper-indonesia/pivot-backoffice/pkg/util"

type TransferDestination struct {
	ChannelCode        string             `json:"channelCode" validate:"required"`
	ChannelInformation ChannelInformation `json:"channelInformation" validate:"required"`
	Description        string             `json:"description" validate:"alphanumspace,max=20"`
}

type ChannelInformation struct {
	AccountNumber string `json:"accountNumber" validate:"required"`
	AccountName   string `json:"accountName" validate:"required"`
}

type MetadataObj struct {
	TransferDestination *TransferDestination `json:"transferDestination,omitempty"`
	ClientMetadata      interface{}          `json:"clientMetadata,omitempty"`
}

func (t *TransferDestination) Scan(value interface{}) error {
	return util.ScanJSON(value, t)
}
