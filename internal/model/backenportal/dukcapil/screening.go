package dukcapilmodel

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/config"
)

type DukcapilScreeningData struct {
	Status       string                `json:"status"`
	FieldResults []DukcapilFieldResult `json:"field_results"`
}

type DukcapilFieldResult struct {
	Field     string `json:"field"`
	Score     int    `json:"score"`
	Threshold int    `json:"threshold"`
	Status    string `json:"status"` // MATCHED, NOT_MATCHED
}

// Constants for Dukcapil screening
const (
	DukcapilProviderName = "DUKCAPIL"

	// Status values
	StatusMatched    = "MATCHED"
	StatusNotMatched = "NOT_MATCHED"

	// Result values
	ResultCompleted = "COMPLETED"
	ResultFailed    = "FAILED"

	// Field names for verification (matching config field names)
	FieldNIK      = "NIK"
	FieldName     = "NAME"
	FieldGender   = "GENDER"
	FieldDOB      = "DOB"
	FieldPOB      = "POB"
	FieldJob      = "JOB"
	FieldAddress  = "ADDRESS"
	FieldRT       = "RT"
	FieldRW       = "RW"
	FieldVillage  = "VILLAGE"
	FieldDistrict = "DISTRICT"
	FieldRegency  = "REGENCY"
	FieldProvince = "PROVINCE"

	// Dukcapil API field names (constants for the hardcoded strings)
	DukcapilFieldNIK      = "NIK"
	DukcapilFieldName     = "NAMA_LGKP"
	DukcapilFieldGender   = "JENIS_KLMIN"
	DukcapilFieldDOB      = "TGL_LHR"
	DukcapilFieldPOB      = "TMPT_LHR"
	DukcapilFieldJob      = "JENIS_PKRJN"
	DukcapilFieldAddress  = "ALAMAT"
	DukcapilFieldRT       = "NO_RT"
	DukcapilFieldRW       = "NO_RW"
	DukcapilFieldVillage  = "KEL_NAME"
	DukcapilFieldDistrict = "KEC_NAME"
	DukcapilFieldRegency  = "KAB_NAME"
	DukcapilFieldProvince = "PROP_NAME"
)

var DefaultFieldThresholds = config.DukcapilFieldThresholds{
	Name:     100,
	Gender:   100,
	DOB:      100,
	POB:      100,
	Job:      100,
	Address:  95,
	RT:       95,
	RW:       95,
	Village:  95,
	District: 95,
	Regency:  95,
	Province: 95,
}

type IdentityVerificationRequest struct {
	MerchantID string `json:"merchantId"`
	*VerifyRequest
}

type IdentityVerificationResponse struct {
	ReferenceID  string                `json:"referenceId"`
	Status       string                `json:"status"`
	FieldResults []DukcapilFieldResult `json:"fieldResults"`
}

func GetFieldThresholds(cfg *config.DukcapilConfig) config.DukcapilFieldThresholds {
	thresholds := DefaultFieldThresholds

	if cfg != nil {
		ft := cfg.FieldThresholds
		if ft.Name > 0 {
			thresholds.Name = ft.Name
		}
		if ft.Gender > 0 {
			thresholds.Gender = ft.Gender
		}
		if ft.DOB > 0 {
			thresholds.DOB = ft.DOB
		}
		if ft.POB > 0 {
			thresholds.POB = ft.POB
		}
		if ft.Job > 0 {
			thresholds.Job = ft.Job
		}
		if ft.Address > 0 {
			thresholds.Address = ft.Address
		}
		if ft.RT > 0 {
			thresholds.RT = ft.RT
		}
		if ft.RW > 0 {
			thresholds.RW = ft.RW
		}
		if ft.Village > 0 {
			thresholds.Village = ft.Village
		}
		if ft.District > 0 {
			thresholds.District = ft.District
		}
		if ft.Regency > 0 {
			thresholds.Regency = ft.Regency
		}
		if ft.Province > 0 {
			thresholds.Province = ft.Province
		}
	}

	return thresholds
}

func GetMinimumThreshold(cfg *config.DukcapilConfig) int {
	thresholds := GetFieldThresholds(cfg)
	minThreshold := 100

	thresholdValues := []int{
		thresholds.Name, thresholds.Gender,
		thresholds.DOB, thresholds.POB, thresholds.Job,
		thresholds.Address, thresholds.RT, thresholds.RW,
		thresholds.Village, thresholds.District, thresholds.Regency,
		thresholds.Province,
	}

	for _, threshold := range thresholdValues {
		if threshold < minThreshold {
			minThreshold = threshold
		}
	}

	return minThreshold
}

// GetThresholdForField returns the threshold value for a specific field from config.DukcapilFieldThresholds
func GetThresholdForField(ft config.DukcapilFieldThresholds, fieldName string) int {
	thresholds := map[string]int{
		FieldName:     ft.Name,
		FieldGender:   ft.Gender,
		FieldDOB:      ft.DOB,
		FieldPOB:      ft.POB,
		FieldJob:      ft.Job,
		FieldAddress:  ft.Address,
		FieldRT:       ft.RT,
		FieldRW:       ft.RW,
		FieldVillage:  ft.Village,
		FieldDistrict: ft.District,
		FieldRegency:  ft.Regency,
		FieldProvince: ft.Province,
	}

	if threshold, exists := thresholds[fieldName]; exists {
		return threshold
	}
	return 100
}

// ParseDukcapilFieldScore extracts the numerical score from Dukcapil response format
// Examples: "Sesuai (92)" -> 92, "Tidak Sesuai" -> 0, "Sesuai" -> 100
func ParseDukcapilFieldScore(responseValue string) int {
	if responseValue == "" {
		return 0
	}

	// this is a bit hardcoded, but this is how dukcapil gave response
	if strings.Contains(strings.ToLower(responseValue), "tidak sesuai") {
		return 0
	}

	// extract score from format like "Sesuai (92)"
	re := regexp.MustCompile(`\((\d+)\)`)
	matches := re.FindStringSubmatch(responseValue)
	if len(matches) > 1 {
		if score, err := strconv.Atoi(matches[1]); err == nil {
			return score
		}
	}

	// this too is hardcoded
	if strings.Contains(strings.ToLower(responseValue), "sesuai") {
		return 100
	}

	return 0
}

// OrderedFieldMapping represents a single field mapping with order preserved
type OrderedFieldMapping struct {
	DukcapilField string
	StandardField string
	Value         string
}

// DukcapilFieldMappings represents all ordered field mappings
type DukcapilFieldMappings struct {
	Fields []OrderedFieldMapping
}

// NewDukcapilFieldMappings creates an ordered field mapping from a VerifyResult
func NewDukcapilFieldMappings(result *VerifyResult) *DukcapilFieldMappings {
	return &DukcapilFieldMappings{
		Fields: []OrderedFieldMapping{
			{DukcapilField: DukcapilFieldName, StandardField: FieldName, Value: result.FullName},
			{DukcapilField: DukcapilFieldGender, StandardField: FieldGender, Value: result.Gender},
			{DukcapilField: DukcapilFieldDOB, StandardField: FieldDOB, Value: result.BirthDate},
			{DukcapilField: DukcapilFieldPOB, StandardField: FieldPOB, Value: result.BirthPlace},
			{DukcapilField: DukcapilFieldJob, StandardField: FieldJob, Value: result.Occupation},
			{DukcapilField: DukcapilFieldAddress, StandardField: FieldAddress, Value: result.Address},
			{DukcapilField: DukcapilFieldRT, StandardField: FieldRT, Value: result.RT},
			{DukcapilField: DukcapilFieldRW, StandardField: FieldRW, Value: result.RW},
			{DukcapilField: DukcapilFieldVillage, StandardField: FieldVillage, Value: result.SubDistrict2},
			{DukcapilField: DukcapilFieldDistrict, StandardField: FieldDistrict, Value: result.SubDistrict},
			{DukcapilField: DukcapilFieldRegency, StandardField: FieldRegency, Value: result.District},
			{DukcapilField: DukcapilFieldProvince, StandardField: FieldProvince, Value: result.Province},
		},
	}
}

func MapDukcapilFieldName(dukcapilFieldName string) string {
	fieldMapping := map[string]string{
		DukcapilFieldNIK:      FieldNIK,
		DukcapilFieldName:     FieldName,
		DukcapilFieldGender:   FieldGender,
		DukcapilFieldDOB:      FieldDOB,
		DukcapilFieldPOB:      FieldPOB,
		DukcapilFieldJob:      FieldJob,
		DukcapilFieldAddress:  FieldAddress,
		DukcapilFieldRT:       FieldRT,
		DukcapilFieldRW:       FieldRW,
		DukcapilFieldVillage:  FieldVillage,
		DukcapilFieldDistrict: FieldDistrict,
		DukcapilFieldRegency:  FieldRegency,
		DukcapilFieldProvince: FieldProvince,
	}

	if mappedField, exists := fieldMapping[dukcapilFieldName]; exists {
		return mappedField
	}

	return dukcapilFieldName
}
