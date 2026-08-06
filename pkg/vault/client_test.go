package vault_test

import (
	"errors"
	"testing"

	. "github.com/paper-indonesia/pivot-backoffice/pkg/vault"

	"github.com/stretchr/testify/assert"
)

func TestClient(t *testing.T) {
	const (
		vaultAddr  = "http://localhost:8200" // NOSONAR
		vaultToken = "test-only-token"       // NOSONAR
	)
	tests := []struct {
		name      string
		config    Config
		wantError error
	}{
		{
			name:      "ERROR:Empty Address",
			config:    Config{},
			wantError: errors.New("vault address is required"), // NOSONAR
		},
		{
			name: "ERROR:Empty Token",
			config: Config{
				Address: vaultAddr,
			},
			wantError: errors.New("vault token is required"), // NOSONAR
		},
		{
			name: "SUCCESS",
			config: Config{
				Address: vaultAddr,
				Token:   vaultToken,
			},
			wantError: nil, // NOSONAR
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := New(test.config)
			assert.Equal(t, test.wantError, err)
		})
	}
}
