package util

import "strings"

func MaskCreditCardNumber(cardNumber string) string {
	if len(cardNumber) < 10 {
		return cardNumber
	}

	maskedLength := len(cardNumber) - 10
	maskedCard := cardNumber[:6] + strings.Repeat("*", maskedLength) + cardNumber[len(cardNumber)-4:]
	return maskedCard
}

// MaskFullName masks the full name, keeping only the first and last letter of each part.
func MaskFullName(fullName string) string {
	parts := strings.Fields(fullName) // Split by spaces
	if len(parts) == 0 {
		return ""
	}

	for i := range parts {
		if len(parts[i]) > 2 {
			parts[i] = string(parts[i][0]) + strings.Repeat("*", len(parts[i])-2) + string(parts[i][len(parts[i])-1])
		} else {
			parts[i] = strings.Repeat("*", len(parts[i])) // Mask completely if too short
		}
	}

	return strings.Join(parts, " ")
}
