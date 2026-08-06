package snapCoreModel

type RegistrationReq struct {
	RegistrationId    string  `json:"registrationId"`
	Acquirer          string  `json:"acquirer"`
	MerchantId        string  `json:"merchantId"` // Used external_id
	MerchantType      string  `json:"merchantType"`
	ParentMerchantId  string  `json:"parentMerchantId"`
	MerchantName      string  `json:"merchantName"`
	Address           Address `json:"address"`
	BusinessShortname string  `json:"businessShortname"`
	MCC               string  `json:"mcc"`
}

type SyncRegistrationDataRequest struct {
	Acquirer                 string   `json:"acquirer" validate:"required"`
	AcquirerMerchantID       string   `json:"acquirerMerchantId" validate:"required"`
	AcquirerParentMerchantID string   `json:"acquirerParentMerchantId" validate:"required"`
	ApplymentCode            string   `json:"applymentCode" validate:"required"`
	ExternalID               string   `json:"externalId" validate:"required"`
	MerchantID               string   `json:"merchantId" validate:"required"`
	MerchantType             string   `json:"merchantType" validate:"required"`
	RegistrationID           string   `json:"registrationId" validate:"required"`
	TerminalID               string   `json:"terminalId"`
	MCC                      string   `json:"mcc" validate:"required"`
	BusinessShortname        string   `json:"businessShortname" validate:"required"`
	StoreIDs                 []string `json:"storeIds,omitempty"`
	Status                   string   `json:"status,omitempty"`
	IsActive                 *bool    `json:"isActive,omitempty"`
}

type SyncRegistrationOption struct {
	MerchantID string
	StoreIDs   []string
	IsActive   *bool
}

type Address struct {
	ProvinceId uint16 `json:"provinceId"`
	CityId     uint16 `json:"cityId"`
	DistrictId uint16 `json:"districtId"`
	PostCode   string `json:"postcode"`
	Detail     string `json:"detail"`
}
