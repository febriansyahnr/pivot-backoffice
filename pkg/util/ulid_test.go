package util_test

import (
	"sync"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/stretchr/testify/require"
)

func TestGenerateULID(t *testing.T) {
	maps := new(sync.Map)

	for i := 0; i < 256; i++ {
		_, ok := maps.LoadOrStore(util.GenerateULID(), true)
		require.False(t, ok)
	}
}
