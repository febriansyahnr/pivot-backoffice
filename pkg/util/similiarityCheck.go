package util

import (
	"fmt"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
)

var (
	InvalidSimilarityMessage      = "The account name entered does not match the bank's records. Please check the account information and try again. Bank record: %s"
	PartialValidSimilarityMessage = "The account name entered is not an exact match. Please check the account information and try again. Bank record: %s"
	ValidSimilarityMessage        = "The account name entered is an exact match. Bank record: %s"
	ValidSimilarityIndicator      = "is an exact match"
	NamePrefixes                  = []string{
		"Bpk.", "Ibu.", "Sdri.", "Sdr.",
		"Bpk", "Ibu", "Sdri", "Sdr",
	}
)

func NameHasPrefix(str string) bool {
	for _, item := range NamePrefixes {
		if strings.HasPrefix(strings.ToLower(str), strings.ToLower(item)) {
			return true
		}
	}
	return false
}

// RemovePrefix removes specified prefixes from the account name
func RemovePrefix(name string) string {

	for _, prefix := range NamePrefixes {
		if strings.HasPrefix(strings.ToLower(name), strings.ToLower(prefix)) {
			// need to string trimming to handle case insensitive-ness
			name = name[len(prefix):]
			name = strings.TrimSpace(name)
			break
		}
	}
	return name
}

// NormalizeNameForComparison removes prefixes and normalizes for comparison
func NormalizeNameForComparison(name, merchantId string) string {
	if constant.IsAccountInquiryIgnoreNamePrefix(merchantId) {
		// Remove prefixes if feature flag is enabled - but keep exact formatting
		cleanName := RemovePrefix(name)
		return strings.ToUpper(cleanName)
	}
	// Original behavior - just uppercase, keep spaces and punctuation
	return strings.ToUpper(name)
}

// CompareNamesWithFeatureFlag compares two names using feature flag configuration
func CompareNamesWithFeatureFlag(inputName, bankRecord, merchantId string) bool {
	if constant.IsAccountInquiryIgnoreNamePrefix(merchantId) {
		// With prefix ignore: remove prefixes but compare exactly (spaces, dots matter)
		cleanInput := strings.ToUpper(RemovePrefix(inputName))
		cleanBank := strings.ToUpper(RemovePrefix(bankRecord))
		return cleanInput == cleanBank
	} else {
		// Without prefix ignore: exact match (including spaces and punctuation)
		return strings.ToUpper(inputName) == strings.ToUpper(bankRecord)
	}
}

// LevenshteinDistance calculates the Levenshtein distance between two strings
func levenshtein(a, b string) int {
	lenA, lenB := len(a), len(b)
	matrix := make([][]int, lenA+1)
	for i := range matrix {
		matrix[i] = make([]int, lenB+1)
	}
	for i := 0; i <= lenA; i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= lenB; j++ {
		matrix[0][j] = j
	}
	for i := 1; i <= lenA; i++ {
		for j := 1; j <= lenB; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			matrix[i][j] = min(
				matrix[i-1][j]+1,
				min(matrix[i][j-1]+1, matrix[i-1][j-1]+cost),
			)
		}
	}
	return matrix[lenA][lenB]
}

// Count common words in two names
func commonWords(name1, name2 string) int {
	words1 := strings.Fields(strings.ToLower(name1))
	words2 := strings.Fields(strings.ToLower(name2))

	wordSet := make(map[string]bool)
	for _, word := range words1 {
		wordSet[word] = true
	}

	commonCount := 0
	for _, word := range words2 {
		if wordSet[word] {
			commonCount++
		}
	}

	return commonCount
}

// SimilarityCheck computes the similarity between two names
func SimilarityCheck(inputName, bankRecord, currentStatus, merchantId string) (string, string) {

	inputName = RemoveNameExtraSpace(inputName)
	inputName = RemovePrefix(inputName)
	bankRecord = RemoveNameExtraSpace(bankRecord)
	bankRecord = RemovePrefix(bankRecord)

	if constant.IsAccountInquiryMerchantNameUseRuneCheck(merchantId) {
		if strings.EqualFold(inputName, bankRecord) {
			return constant.RequestAccountInquiryStatusValid, ""
		} else {
			return constant.RequestAccountInquiryStatusWarning, fmt.Sprintf(InvalidSimilarityMessage, bankRecord)
		}
	}
	// Compute word-level similarity
	commonWordCount := commonWords(inputName, bankRecord)

	// Calculate Levenshtein distance - set to case insensitive
	dist := levenshtein(strings.ToUpper(inputName), strings.ToUpper(bankRecord))

	// Compute similarity percentage
	maxLength := max(len(inputName), len(bankRecord))
	if maxLength == 0 {
		return constant.RequestAccountInquiryStatusValid, fmt.Sprintf(ValidSimilarityMessage, bankRecord)
	}
	characterSimilarity := 100 - (dist * 100 / maxLength)

	// Improved logic for partial match
	hasCommonWords := commonWordCount > 0
	isLevenshteinClose := characterSimilarity >= 60

	// Exact match condition
	if characterSimilarity == 100 {
		return constant.RequestAccountInquiryStatusValid, ""
	}

	// Partial match condition
	if hasCommonWords || isLevenshteinClose {
		return currentStatus, fmt.Sprintf(PartialValidSimilarityMessage, bankRecord)
	}

	return currentStatus, fmt.Sprintf(InvalidSimilarityMessage, bankRecord)
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
