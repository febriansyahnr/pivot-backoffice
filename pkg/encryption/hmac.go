package encryption

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func GenerateHMAC(key, message string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}
