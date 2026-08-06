package encryption_test

import (
	"testing"

	. "github.com/paper-indonesia/pivot-backoffice/pkg/encryption"

	"github.com/stretchr/testify/require"
)

func TestGenerateHMAC(t *testing.T) {
	const (
		key     = "2764f0f2-c9ac-4fe1-80ad-ae1eca541814"
		message = "This is response data"
		result  = "37512a3681369c68fd3934eb8afb7a5bac995c58bdc6acb1324248d7c369f1f5"
	)
	require.Equal(t, result, GenerateHMAC(key, message))
}
