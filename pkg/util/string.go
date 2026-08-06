package util

import (
	"crypto/rand"
	cryptoRand "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func GenerateDeviceID(userAgent string) string {
	// Hash the User-Agent string using SHA-256
	hash := sha256.Sum256([]byte(userAgent))

	// Convert the hash to a hexadecimal string
	deviceID := fmt.Sprintf("%x", hash)

	return deviceID
}

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		// Generate a random index using crypto/rand
		randomIndex, err := cryptoRand.Int(cryptoRand.Reader, big.NewInt(int64(len(letterBytes))))
		if err != nil {
			// Fallback in case of error, though this should rarely happen
			b[i] = letterBytes[0]
			continue
		}
		b[i] = letterBytes[randomIndex.Int64()]
	}
	return string(b)
}

func CreateSlug(input string) string {
	// Convert the string to lowercase
	slug := strings.ToLower(input)

	// Replace non-alphanumeric characters (excluding hyphen) with hyphens
	reg := regexp.MustCompile("[^a-z0-9-]+")
	slug = reg.ReplaceAllString(slug, "-")

	// Remove leading and trailing hyphens
	slug = strings.Trim(slug, "-")

	// Replace multiple consecutive hyphens with a single hyphen
	slug = strings.ReplaceAll(slug, "--", "-")

	// Convert to UPPER
	slug = strings.ToUpper(slug)

	return slug
}

func ToTitle(str string) string {
	// Change text to Lower Case
	str = strings.ToLower(str)

	// Remove Underscore
	str = strings.ReplaceAll(str, "_", " ")

	// Change text to Title
	c := cases.Title(language.English)
	title := c.String(str)

	return title
}

func ParseUUID(id string) uuid.UUID {
	result, _ := uuid.Parse(id)
	return result
}

func GenerateRandomString(n int) (string, error) {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var result string
	for i := 0; i < n; i++ {
		num, err := cryptoRand.Int(cryptoRand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err // return an empty string and the error
		}
		result += string(letters[num.Int64()])
	}
	return result, nil
}

func InArray(arr []string, target string) bool {
	for _, v := range arr {
		if v == target {
			return true
		}
	}
	return false
}

func IsNumericValue(value string) bool {
	return regexp.MustCompile("^[0-9]+$").Match([]byte(value))
}

func HashString(value string) string {
	// Create a new SHA-256 hasher
	hasher := sha256.New()

	// Write the PIN bytes to the hasher
	hasher.Write([]byte(value))

	// Get the hashed PIN bytes
	hashedPinBytes := hasher.Sum(nil)

	// Convert the hashed string bytes to a hexadecimal string
	hashedString := hex.EncodeToString(hashedPinBytes)

	return hashedString
}

// GenerateUniqueRandomToken generates a random token of specified byte length
func GenerateUniqueRandomToken(length int) (string, error) {
	// Create a byte slice to hold the random bytes
	bytes := make([]byte, length)

	// Read random bytes from the crypto/rand reader
	if _, err := io.ReadFull(cryptoRand.Reader, bytes); err != nil {
		return "", err
	}

	// Encode the bytes to a base64 URL-encoded string
	token := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(bytes)

	return token, nil
}

// ConvertCamelToSnake converts camelCase or PascalCase strings to snake_case
func ConvertCamelToSnake(str string) string {
	re := regexp.MustCompile("([a-z0-9])([A-Z])")
	snake := re.ReplaceAllString(str, "${1}_${2}")
	return strings.ToLower(snake)
}

func IsPatternMatch(pattern, str string) bool {
	// Compile the regular expression
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}

	// Compare the string with the pattern
	if re.MatchString(str) {
		return true
	}

	return false
}

func GetValueAsString(data map[string]any, key, defaultValue string) string {
	if val, ok := data[key]; ok {
		if str, okStr := val.(string); okStr {
			return str
		}
	}
	return defaultValue
}

// trim left with length of max length
func TrimLength(str string, maxLength int) string {
	// trim, max length maxLength
	strLen := len(str) - maxLength
	if strLen <= 0 {
		strLen = 0
	}
	return str[strLen:]
}

// trim right with length of max length
func TrimLengthRight(str string, maxLength int) string {
	// trim, max length maxLength
	if maxLength <= 0 {
		maxLength = 0
	}

	strLen := len(str)
	if maxLength > strLen {
		maxLength = strLen
	}

	return str[:maxLength]
}

func RemoveNameExtraSpace(str string) string {
	str = strings.TrimSpace(str)

	var builder strings.Builder
	builder.Grow(len(str))
	for i, char := range str {
		if char != ' ' {
			builder.WriteRune(char)
		} else if char == ' ' {
			if i > 0 && str[i-1] == ' ' {
				continue
			} else {
				builder.WriteRune(' ')
			}
		}
	}

	return builder.String()
}

func GenerateVccRandomPartnerReferenceNo() string {
	timestamp := fmt.Sprintf("%d", time.Now().UnixNano())
	base := timestamp[:15]

	// Generate cryptographically secure 3-digit suffix
	randomBytes := make([]byte, 2)
	if _, err := rand.Read(randomBytes); err != nil {
		// Fallback if crypto/rand fails
		suffix := fmt.Sprintf("%03d", time.Now().UnixNano()%1000)
		return base + suffix
	}

	// Convert random bytes to a number between 0-999
	randomNum := int(randomBytes[0])<<8 + int(randomBytes[1])
	suffix := fmt.Sprintf("%03d", randomNum%1000)
	return base + suffix

}
