package constant

const (
	SimulationPhoneSuccess    = "08111111001"
	SimulationPhoneProcessing = "08111111002"
	SimulationPhoneFailed     = "08111111003"
	SimulationPhoneExpired    = "08111111004"
)

const (
	PhoneNumberUnknown = "UNKNOWN"
)

func GetChargeStatusByPhoneNumber(phoneNumber string) string {
	// Default mapping
	mockData := map[string]string{
		SimulationPhoneSuccess:    ChargeStatusSuccess,
		SimulationPhoneProcessing: ChargeStatusProcessing,
		SimulationPhoneFailed:     ChargeStatusFailed,
		SimulationPhoneExpired:    ChargeStatusExpired,
	}

	// Return mapped status
	if status, ok := mockData[phoneNumber]; ok {
		return status
	}

	return PhoneNumberUnknown
}
