package random_test

import (
	"sync"
	"testing"

	. "github.com/paper-indonesia/pivot-backoffice/pkg/random"

	"github.com/stretchr/testify/require"
)

var otps = new(sync.Map)

func TestGenerateOTP(t *testing.T) {
	const length = 6

	for i := 0; i < 100; i++ {
		otp := GenerateOTP(length)
		require.Len(t, otp, length)

		_, ok := otps.LoadOrStore(otp, 1)
		require.False(t, ok)
	}
}
