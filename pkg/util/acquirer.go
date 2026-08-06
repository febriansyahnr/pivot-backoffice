package util

import "strings"

// CleanAcquirerName removes common suffixes from acquirer strings returned by payment processors.
// Examples:
//   - "BNC_QRIS" -> "bnc"
//   - "BNI_VA" -> "bni"
//   - "BRI_EWALLET" -> "bri"
//   - "BNC" -> "bnc"
func CleanAcquirerName(acquirer string) string {
	if acquirer == "" {
		return ""
	}

	// Convert to lowercase
	cleaned := strings.ToLower(acquirer)

	// Remove common suffixes
	suffixes := []string{"_qris", "_va", "_ewallet", "_cc", "_ovo", "_dana", "_gopay", "_linkaja"}
	for _, suffix := range suffixes {
		cleaned = strings.TrimSuffix(cleaned, suffix)
	}

	return cleaned
}
