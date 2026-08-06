package industry

import (
	"github.com/paper-indonesia/pivot-backoffice/constant"
)

// IsValidDigitalStatus checks if the given digital status is valid
func IsValidDigitalStatus(status string) bool {
	for _, validStatus := range constant.DigitalStatusOptions {
		if validStatus == status {
			return true
		}
	}
	return false
}

// IsValidCountryEntity checks if the given country code is valid
func IsValidCountryEntity(countryCode string) bool {
	_, exists := constant.CountryEntityCodes[countryCode]
	return exists
}
