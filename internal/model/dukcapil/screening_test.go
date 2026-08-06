package dukcapilmodel

import (
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"

	"github.com/stretchr/testify/assert"
)

func TestVerifyResultIsEmpty(t *testing.T) {
	tests := []struct {
		name string
		r    *VerifyResult
		want bool
	}{
		{
			name: "empty response code returns true",
			r:    &VerifyResult{},
			want: true,
		},
		{
			name: "non-empty response code returns false",
			r: &VerifyResult{
				ResponseCode: "200",
			},
			want: false,
		},
		{
			name: "result with other fields but empty response code returns true",
			r: &VerifyResult{
				FullName:   "John Doe",
				BirthPlace: "Jakarta",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.r.IsEmpty())
		})
	}
}

func TestToGatewayRequest(t *testing.T) {
	tests := []struct {
		name string
		req  *VerifyRequest
		want *GatewayVerifyRequest
	}{
		{
			name: "maps all fields correctly",
			req: &VerifyRequest{
				NIK:      "3201234567890001",
				Name:     "John Doe",
				Gender:   "LAKI-LAKI",
				DOB:      "1990-01-15",
				POB:      "Jakarta",
				Job:      "Programmer",
				Address:  "Jl. Sudirman No. 1",
				RT:       "001",
				RW:       "002",
				Village:  "Karet",
				District: "Karet Semanggi",
				Regency:  "Kota Jakarta Selatan",
				Province: "DKI Jakarta",
			},
			want: &GatewayVerifyRequest{
				NIK:          "3201234567890001",
				FullName:     "John Doe",
				Gender:       "LAKI-LAKI",
				BirthDate:    "1990-01-15",
				BirthPlace:   "Jakarta",
				Occupation:   "Programmer",
				Address:      "Jl. Sudirman No. 1",
				RT:           "001",
				RW:           "002",
				SubDistrict2: "Karet",
				SubDistrict:  "Karet Semanggi",
				District:     "Kota Jakarta Selatan",
				Province:     "DKI Jakarta",
			},
		},
		{
			name: "empty fields mapped as empty",
			req:  &VerifyRequest{},
			want: &GatewayVerifyRequest{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ToGatewayRequest(tt.req))
		})
	}
}

func TestGetFieldThresholds(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.DukcapilConfig
		want config.DukcapilFieldThresholds
	}{
		{
			name: "nil config returns defaults",
			cfg:  nil,
			want: DefaultFieldThresholds,
		},
		{
			name: "empty config returns defaults",
			cfg:  &config.DukcapilConfig{},
			want: DefaultFieldThresholds,
		},
		{
			name: "zero value thresholds are ignored and defaults used",
			cfg: &config.DukcapilConfig{
				FieldThresholds: config.DukcapilFieldThresholds{},
			},
			want: DefaultFieldThresholds,
		},
		{
			name: "partial config overrides only specified fields",
			cfg: &config.DukcapilConfig{
				FieldThresholds: config.DukcapilFieldThresholds{
					Name:    90,
					Address: 80,
				},
			},
			want: config.DukcapilFieldThresholds{
				Name:     90,
				Gender:   DefaultFieldThresholds.Gender,
				DOB:      DefaultFieldThresholds.DOB,
				POB:      DefaultFieldThresholds.POB,
				Job:      DefaultFieldThresholds.Job,
				Address:  80,
				RT:       DefaultFieldThresholds.RT,
				RW:       DefaultFieldThresholds.RW,
				Village:  DefaultFieldThresholds.Village,
				District: DefaultFieldThresholds.District,
				Regency:  DefaultFieldThresholds.Regency,
				Province: DefaultFieldThresholds.Province,
			},
		},
		{
			name: "full config overrides all defaults",
			cfg: &config.DukcapilConfig{
				FieldThresholds: config.DukcapilFieldThresholds{
					Name:     85,
					Gender:   90,
					DOB:      95,
					POB:      88,
					Job:      92,
					Address:  80,
					RT:       75,
					RW:       78,
					Village:  82,
					District: 84,
					Regency:  86,
					Province: 88,
				},
			},
			want: config.DukcapilFieldThresholds{
				Name:     85,
				Gender:   90,
				DOB:      95,
				POB:      88,
				Job:      92,
				Address:  80,
				RT:       75,
				RW:       78,
				Village:  82,
				District: 84,
				Regency:  86,
				Province: 88,
			},
		},
		{
			name: "single field override preserves all other defaults",
			cfg: &config.DukcapilConfig{
				FieldThresholds: config.DukcapilFieldThresholds{
					Name: 50,
				},
			},
			want: config.DukcapilFieldThresholds{
				Name:     50,
				Gender:   DefaultFieldThresholds.Gender,
				DOB:      DefaultFieldThresholds.DOB,
				POB:      DefaultFieldThresholds.POB,
				Job:      DefaultFieldThresholds.Job,
				Address:  DefaultFieldThresholds.Address,
				RT:       DefaultFieldThresholds.RT,
				RW:       DefaultFieldThresholds.RW,
				Village:  DefaultFieldThresholds.Village,
				District: DefaultFieldThresholds.District,
				Regency:  DefaultFieldThresholds.Regency,
				Province: DefaultFieldThresholds.Province,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, GetFieldThresholds(tt.cfg))
		})
	}
}

func TestGetMinimumThreshold(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.DukcapilConfig
		want int
	}{
		{
			name: "nil config returns minimum of defaults",
			cfg:  nil,
			want: 95, // min of DefaultFieldThresholds (95 is the lowest)
		},
		{
			name: "custom config returns lowest custom value",
			cfg: &config.DukcapilConfig{
				FieldThresholds: config.DukcapilFieldThresholds{
					Name:     100,
					Address:  80,
					Province: 90,
				},
			},
			want: 80,
		},
		{
			name: "all fields equal returns that value",
			cfg: &config.DukcapilConfig{
				FieldThresholds: config.DukcapilFieldThresholds{
					Name: 100, Gender: 100, DOB: 100, POB: 100,
					Job: 100, Address: 100, RT: 100, RW: 100,
					Village: 100, District: 100, Regency: 100, Province: 100,
				},
			},
			want: 100,
		},
		{
			name: "single low field determines minimum",
			cfg: &config.DukcapilConfig{
				FieldThresholds: config.DukcapilFieldThresholds{
					Name:    50,
					Address: 100,
				},
			},
			want: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, GetMinimumThreshold(tt.cfg))
		})
	}
}

func TestGetThresholdForField(t *testing.T) {
	ft := config.DukcapilFieldThresholds{
		Name:     90,
		Gender:   95,
		DOB:      100,
		POB:      85,
		Job:      80,
		Address:  75,
		RT:       70,
		RW:       72,
		Village:  78,
		District: 82,
		Regency:  88,
		Province: 92,
	}

	tests := []struct {
		name      string
		ft        config.DukcapilFieldThresholds
		fieldName string
		want      int
	}{
		{
			name:      "return threshold for NAME field",
			ft:        ft,
			fieldName: FieldName,
			want:      90,
		},
		{
			name:      "return threshold for GENDER field",
			ft:        ft,
			fieldName: FieldGender,
			want:      95,
		},
		{
			name:      "return threshold for ADDRESS field",
			ft:        ft,
			fieldName: FieldAddress,
			want:      75,
		},
		{
			name:      "return 100 for unknown field",
			ft:        ft,
			fieldName: "UNKNOWN_FIELD",
			want:      100,
		},
		{
			name:      "return 100 for empty field name",
			ft:        ft,
			fieldName: "",
			want:      100,
		},
		{
			name:      "all fields present",
			ft:        ft,
			fieldName: FieldProvince,
			want:      92,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, GetThresholdForField(tt.ft, tt.fieldName))
		})
	}
}

func TestParseDukcapilFieldScore(t *testing.T) {
	tests := []struct {
		name          string
		responseValue string
		want          int
	}{
		{
			name:          "empty string returns 0",
			responseValue: "",
			want:          0,
		},
		{
			name:          "tidak sesuai returns 0",
			responseValue: "Tidak Sesuai",
			want:          0,
		},
		{
			name:          "tidak sesuai mixed case returns 0",
			responseValue: "TIDAK SESUAI",
			want:          0,
		},
		{
			name:          "sesuai with score in parentheses",
			responseValue: "Sesuai (92)",
			want:          92,
		},
		{
			name:          "sesuai with score 100 in parentheses",
			responseValue: "Sesuai (100)",
			want:          100,
		},
		{
			name:          "sesuai without score returns 100",
			responseValue: "Sesuai",
			want:          100,
		},
		{
			name:          "sesuai lowercase returns 100",
			responseValue: "sesuai",
			want:          100,
		},
		{
			name:          "sesuai with score 85",
			responseValue: "Sesuai (85)",
			want:          85,
		},
		{
			name:          "unrecognized string returns 0",
			responseValue: "some random text",
			want:          0,
		},
		{
			name:          "score with parentheses only no text",
			responseValue: "(75)",
			want:          75,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ParseDukcapilFieldScore(tt.responseValue))
		})
	}
}

func TestNewDukcapilFieldMappings(t *testing.T) {
	result := &VerifyResult{
		FullName:     "John Doe",
		Gender:       "LAKI-LAKI",
		BirthDate:    "1990-01-15",
		BirthPlace:   "Jakarta",
		Occupation:   "Programmer",
		Address:      "Jl. Sudirman No. 1",
		RT:           "001",
		RW:           "002",
		SubDistrict2: "Karet",
		SubDistrict:  "Karet Semanggi",
		District:     "Kota Jakarta Selatan",
		Province:     "DKI Jakarta",
	}

	want := &DukcapilFieldMappings{
		Fields: []OrderedFieldMapping{
			{DukcapilField: DukcapilFieldName, StandardField: FieldName, Value: "John Doe"},
			{DukcapilField: DukcapilFieldGender, StandardField: FieldGender, Value: "LAKI-LAKI"},
			{DukcapilField: DukcapilFieldDOB, StandardField: FieldDOB, Value: "1990-01-15"},
			{DukcapilField: DukcapilFieldPOB, StandardField: FieldPOB, Value: "Jakarta"},
			{DukcapilField: DukcapilFieldJob, StandardField: FieldJob, Value: "Programmer"},
			{DukcapilField: DukcapilFieldAddress, StandardField: FieldAddress, Value: "Jl. Sudirman No. 1"},
			{DukcapilField: DukcapilFieldRT, StandardField: FieldRT, Value: "001"},
			{DukcapilField: DukcapilFieldRW, StandardField: FieldRW, Value: "002"},
			{DukcapilField: DukcapilFieldVillage, StandardField: FieldVillage, Value: "Karet"},
			{DukcapilField: DukcapilFieldDistrict, StandardField: FieldDistrict, Value: "Karet Semanggi"},
			{DukcapilField: DukcapilFieldRegency, StandardField: FieldRegency, Value: "Kota Jakarta Selatan"},
			{DukcapilField: DukcapilFieldProvince, StandardField: FieldProvince, Value: "DKI Jakarta"},
		},
	}

	got := NewDukcapilFieldMappings(result)
	assert.Equal(t, want, got)
	assert.Equal(t, 12, len(got.Fields), "should have 12 field mappings")
}

func TestMapDukcapilFieldName(t *testing.T) {
	tests := []struct {
		name              string
		dukcapilFieldName string
		want              string
	}{
		{name: "NIK", dukcapilFieldName: DukcapilFieldNIK, want: FieldNIK},
		{name: "NAME", dukcapilFieldName: DukcapilFieldName, want: FieldName},
		{name: "GENDER", dukcapilFieldName: DukcapilFieldGender, want: FieldGender},
		{name: "DOB", dukcapilFieldName: DukcapilFieldDOB, want: FieldDOB},
		{name: "POB", dukcapilFieldName: DukcapilFieldPOB, want: FieldPOB},
		{name: "JOB", dukcapilFieldName: DukcapilFieldJob, want: FieldJob},
		{name: "ADDRESS", dukcapilFieldName: DukcapilFieldAddress, want: FieldAddress},
		{name: "RT", dukcapilFieldName: DukcapilFieldRT, want: FieldRT},
		{name: "RW", dukcapilFieldName: DukcapilFieldRW, want: FieldRW},
		{name: "VILLAGE", dukcapilFieldName: DukcapilFieldVillage, want: FieldVillage},
		{name: "DISTRICT", dukcapilFieldName: DukcapilFieldDistrict, want: FieldDistrict},
		{name: "REGENCY", dukcapilFieldName: DukcapilFieldRegency, want: FieldRegency},
		{name: "PROVINCE", dukcapilFieldName: DukcapilFieldProvince, want: FieldProvince},
		{
			name:              "unknown field returns original",
			dukcapilFieldName: "UNKNOWN_FIELD",
			want:              "UNKNOWN_FIELD",
		},
		{
			name:              "empty string returns empty",
			dukcapilFieldName: "",
			want:              "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, MapDukcapilFieldName(tt.dukcapilFieldName))
		})
	}
}
