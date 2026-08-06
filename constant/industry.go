package constant

import "slices"

// Industry constants for merchant classification
type IndustryData struct {
	ParentIndustry string
	ChildIndustry  string
	RiskLevel      string
	MCC            string
	CommonMCC      string
}

// Risk level constants for industry
const (
	IndustryRiskLevelLow    = "Low"
	IndustryRiskLevelMedium = "Medium"
	IndustryRiskLevelHigh   = "High"
)

// DigitalStatusOptions represent the available digital status options
var DigitalStatusOptions = []string{
	"Digital",
	"Non-digital",
}

// IndustryRiskLevelOptions represent the available risk levels for industry
var IndustryRiskLevelOptions = []string{
	IndustryRiskLevelLow,
	IndustryRiskLevelMedium,
	IndustryRiskLevelHigh,
}

// IsValidIndustryRiskLevel checks if the given risk level is valid
func IsValidIndustryRiskLevel(level string) bool {
	return slices.Contains(IndustryRiskLevelOptions, level)

}

// Country entity codes for merchant origin
var CountryEntityCodes = map[string]string{
	"ID": "Indonesia",     // Local
	"SG": "Singapore",     // Foreign
	"MY": "Malaysia",      // Foreign
	"TH": "Thailand",      // Foreign
	"VN": "Vietnam",       // Foreign
	"PH": "Philippines",   // Foreign
	"TW": "Taiwan",        // Foreign
	"US": "United States", // Foreign
	// Add more country codes as needed
}
