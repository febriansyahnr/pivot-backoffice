package util

// CleanUpIDNPhoneNumber normalizes Indonesian phone numbers by removing common prefixes.
// It strips the following prefixes from the input phone number:
// - "62" (Indonesian country code)
// - "0" (National prefix for Indonesian numbers)
// - "+62" (International format for Indonesian numbers)
// Examples:
//
//	CleanUpIDNPhoneNumber("+62812345678") returns "812345678"
//	CleanUpIDNPhoneNumber("62812345678") returns "812345678"
//	CleanUpIDNPhoneNumber("0812345678") returns "812345678"
//	CleanUpIDNPhoneNumber("812345678") returns "812345678"
//	CleanUpIDNPhoneNumber("") returns ""
func CleanUpIDNPhoneNumber(phoneNumber string) string {
	if phoneNumber == "" {
		return ""
	}

	// Remove Prefixes like +62, 62, 0, etc.
	if len(phoneNumber) > 2 && phoneNumber[:2] == "62" {
		return phoneNumber[2:]
	}

	if len(phoneNumber) > 1 && phoneNumber[:1] == "0" {
		return phoneNumber[1:]
	}

	if len(phoneNumber) > 3 && phoneNumber[:3] == "+62" {
		return phoneNumber[3:]
	}

	return phoneNumber
}
