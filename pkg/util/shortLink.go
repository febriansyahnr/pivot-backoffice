package util

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"hash/fnv"
	"strings"
	"time"
)

func GenerateShortCode(uniqueID string) string {
	var (
		ts      = uint64(time.Now().UnixNano())
		entropy [8]byte
	)

	if uniqueID != "" {
		// Use 64-bit hash for better collision resistance
		h := fnv.New64a()
		h.Write([]byte(uniqueID))
		binary.BigEndian.PutUint64(entropy[:], h.Sum64())
	} else {
		// Cryptographically secure random
		rand.Read(entropy[:])
	}

	// 16 bytes total: 8 timestamp + 8 entropy
	buf := make([]byte, 16)
	binary.BigEndian.PutUint64(buf[:8], ts)
	copy(buf[8:], entropy[:])

	codes := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(buf)
	codes = strings.ReplaceAll(codes, "-", "")
	codes = strings.ReplaceAll(codes, "_", "")
	return codes[:11]
}
