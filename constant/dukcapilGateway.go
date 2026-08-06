package constant

const DUKCAPIL_GATEWAY = "DUKCAPIL_GATEWAY"

var (
	DataFound              = "01"
	LoginFailed            = "03"
	IncorrectIP            = "04"
	DailyQuotaReached      = "05"
	IncorrectAccessTime    = "06"
	IncompleteRequestParam = "07"
	NIKUnprocessable       = "08"
	InvalidCredentials     = "09" // Empty user, password, ip
	InvalidIP              = "10"
	DataFoundDeceased      = "11"
	DataFoundMultipleData  = "12"
	NIKDataNotFound        = "13"
	InactiveData           = "14"
	DataNotFound           = "15" // Data not found, NIK not aligned with dukcapil format
	InstanceNotActive      = "16"
	InactiveUserID         = "17" // Need admin
)

var SuccessResponseCode = map[string]bool{
	DataFound:             true,
	DataFoundMultipleData: true,
}

var FailedResponseCode = map[string]bool{
	NIKUnprocessable:  true,
	NIKDataNotFound:   true,
	InactiveData:      true,
	DataNotFound:      true,
	DataFoundDeceased: true,
}

var IncorrectSetupResponseCode = map[string]bool{
	LoginFailed:            true,
	IncorrectIP:            true,
	DailyQuotaReached:      true,
	IncorrectAccessTime:    true,
	IncompleteRequestParam: true,
	InvalidCredentials:     true,
	InvalidIP:              true,
	InstanceNotActive:      true,
	InactiveUserID:         true,
}
