package location

type LocationResp struct {
	ProvinceList interface{} `json:"provinceList,omitempty"`
	CityList     interface{} `json:"cityList,omitempty"`
	DistrictList interface{} `json:"districtList,omitempty"`
}
