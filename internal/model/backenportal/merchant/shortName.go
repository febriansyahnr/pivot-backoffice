package merchant

import "mime/multipart"

type ReservedMerchantShortNameRequest struct {
	File *multipart.FileHeader
}

type ReservedMerchantShortNameItem struct {
	ShortName        string   `json:"shortName"`
	AllowedMerchants []string `json:"allowedMerchants"`
}
