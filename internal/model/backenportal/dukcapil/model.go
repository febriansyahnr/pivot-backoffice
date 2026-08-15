package dukcapilmodel

type VerifyRequest struct {
	NIK      string `json:"nik" validate:"required"`
	Name     string `json:"name" validate:"required"`
	Gender   string `json:"gender"`
	DOB      string `json:"dob" validate:"required"`
	POB      string `json:"pob"`
	Job      string `json:"job"`
	Address  string `json:"address"`
	RT       string `json:"rt"`
	RW       string `json:"rw"`
	Village  string `json:"village"`
	District string `json:"district"`
	Regency  string `json:"regency"`
	Province string `json:"province"`
}

type VerifyResult struct {
	Address        string `json:"ALAMAT,omitempty"`
	Gender         string `json:"JENIS_KLMIN,omitempty"`
	Occupation     string `json:"JENIS_PKRJN,omitempty"`
	District       string `json:"KAB_NAME,omitempty"`
	SubDistrict    string `json:"KEC_NAME,omitempty"`
	SubDistrict2   string `json:"KEL_NAME,omitempty"`
	Province       string `json:"PROP_NAME,omitempty"`
	FullName       string `json:"NAMA_LGKP,omitempty"`
	DistrictNo     string `json:"NO_KAB,omitempty"`
	SubDistrictNo  string `json:"NO_KEC,omitempty"`
	SubDistrict2No string `json:"NO_KEL,omitempty"`
	ProvinceNo     string `json:"NO_PROP,omitempty"`
	RT             string `json:"NO_RT,omitempty"`
	RW             string `json:"NO_RW,omitempty"`
	MaritalStatus  string `json:"STATUS_KAWIN,omitempty"`
	BirthDate      string `json:"TGL_LHR,omitempty"`
	BirthPlace     string `json:"TMPT_LHR,omitempty"`
	ResponseCode   string `json:"RESPONSE_CODE,omitempty"`
	ResponseDesc   string `json:"RESPONSE_DESC,omitempty"`
	Response       string `json:"RESPONSE,omitempty"`
}

func (r *VerifyResult) IsEmpty() bool {
	return r.ResponseCode == ""
}
