package util

import (
	"crypto/rand"
	"io"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
)

var (
	ulmx sync.Mutex
)

func GenerateULID() string {
	ulmx.Lock()
	defer ulmx.Unlock()

	// Use crypto/rand which is cryptographically secure
	id, _ := ulid.New(ulid.Timestamp(time.Now()), io.Reader(rand.Reader))
	return id.String()
}
