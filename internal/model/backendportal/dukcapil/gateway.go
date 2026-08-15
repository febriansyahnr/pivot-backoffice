package dukcapilmodel

type GatewayVerifyRequest struct {
	UserID       string `json:"USER_ID"`
	Password     string `json:"PASSWORD"`
	IPUser       string `json:"IP_USER"`
	Threshold    int    `json:"TRESHOLD"`
	NIK          string `json:"NIK"`
	Address      string `json:"ALAMAT"`
	Gender       string `json:"JENIS_KLMIN"`
	Occupation   string `json:"JENIS_PKRJN"`
	District     string `json:"KAB_NAME"`
	SubDistrict  string `json:"KEC_NAME"`
	SubDistrict2 string `json:"KEL_NAME"`
	Province     string `json:"PROP_NAME"`
	FullName     string `json:"NAMA_LGKP"`
	RT           string `json:"NO_RT"`
	RW           string `json:"NO_RW"`
	BirthDate    string `json:"TGL_LHR"`
	BirthPlace   string `json:"TMPT_LHR"`
}

func ToGatewayRequest(r *VerifyRequest) *GatewayVerifyRequest {
	return &GatewayVerifyRequest{
		NIK:          r.NIK,
		Address:      r.Address,
		Gender:       r.Gender,
		Occupation:   r.Job,
		District:     r.Regency,
		SubDistrict:  r.District,
		SubDistrict2: r.Village,
		Province:     r.Province,
		FullName:     r.Name,
		RT:           r.RT,
		RW:           r.RW,
		BirthDate:    r.DOB,
		BirthPlace:   r.POB,
	}
}

type GatewayVerifyResponse struct {
	Content          []VerifyResult `json:"content"`
	LastPage         bool           `json:"lastPage"`
	NumberOfElements int            `json:"numberOfElements"`
	TotalElements    int            `json:"totalElements"`
	FirstPage        bool           `json:"firstPage"`
	Number           int            `json:"number"`
	Size             int            `json:"size"`
	QuotaLimiter     int            `json:"quotaLimiter"`
}
