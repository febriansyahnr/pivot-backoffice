package util

import (
	"crypto/sha256"
	"fmt"
	"io"
)

// IOToSHA256 computes the SHA-256 hash of the data read from the provided io.Reader.
// It returns the hash as a string in the format "sha256:<hexadecimal hash>".
// If the provided io.Reader is nil, it returns an empty string.
func IOToSHA256(src io.Reader) string {
	if src == nil {
		return ""
	}

	raw, _ := io.ReadAll(src)
	sha := sha256.New()
	sha.Write(raw)
	return fmt.Sprintf("sha256:%x", sha.Sum(nil))
}
