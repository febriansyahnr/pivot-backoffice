package qris

import "time"

type RegistrationResp struct {
	Id string `json:"id"`
}

type RegistrationListResp struct {
	Id                       string          `json:"id"`
	ExternalId               string          `json:"externalId"`
	Acquirer                 string          `json:"acquirer"`
	MerchantType             string          `json:"merchantType"`
	AcquirerMerchantParentId string          `json:"acquirerMerchantParentId"`
	MerchantName             string          `json:"merchantName"`
	Status                   string          `json:"status"`
	AcquirerMerchantId       *string         `json:"acquirerMerchantId"`
	CallbackDetail           *CallbackDetail `json:"callbackDetail"`
	CallbackDatetime         *time.Time      `json:"callbackDatetime"`
	CreatedAt                time.Time       `json:"createdAt"`
	UpdatedAt                time.Time       `json:"updatedAt"`
}

type ReuploadDocumentResp struct {
	Uploaded bool `json:"uploaded"`
}

type DuplicateRegistrationResp struct {
	Id string `json:"id"`
}

func (r *RegistrationListResp) FromRegistration(reg Registration) {
	r.Id = reg.Id
	r.ExternalId = reg.ExternalId
	r.Acquirer = reg.Acquirer
	r.MerchantType = reg.MerchantType
	r.AcquirerMerchantParentId = reg.AcquirerParentMerchantId
	r.MerchantName = reg.MerchantName
	r.Status = reg.Status
	r.CreatedAt = reg.CreatedAt
	r.UpdatedAt = reg.UpdatedAt

	if reg.AcquirerMerchantId != nil && *reg.AcquirerMerchantId != "" {
		r.AcquirerMerchantId = reg.AcquirerMerchantId
	}
	if reg.CallbackDetailRaw.Valid {
		r.CallbackDatetime = &reg.CallbackDatetime.Time
		_ = reg.CallbackDetailRaw.Unmarshal(&r.CallbackDetail)
	}
}
