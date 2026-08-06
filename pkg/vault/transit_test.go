package vault_test

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	. "github.com/paper-indonesia/pivot-backoffice/pkg/vault"
	"github.com/paper-indonesia/pivot-backoffice/test/containers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrationNewTransit(t *testing.T) {
	if os.Getenv(constant.IntegrationTestEnv) != "1" {
		t.Skip(constant.SkipIntegrationTest)
	}

	const (
		secretPath = "transit-engine"
		secretKey  = "my-key"
	)

	ctx := t.Context()

	container, err := containers.NewVault(ctx)
	require.NoError(t, err, "Failed while creating Vault container")

	defer container.Stop(ctx, nil)

	_, _, err = container.Exec(ctx, []string{
		"vault", "secrets", "enable", "-path=" + secretPath, "transit",
	})
	require.NoError(t, err, "Failed to enable Transit secret engine")

	_, output, err := container.Exec(ctx, []string{
		"vault", "write", fmt.Sprintf("%s/keys/%s", secretPath, secretKey), `type=aes256-gcm96`,
	})
	require.NoError(t, err, "Failed while creating Transit key")

	console, _ := io.ReadAll(output)
	require.Contains(t, string(console), "aes256-gcm96")

	host, err := container.Host(ctx)
	require.NoError(t, err, "Failed to get container host")

	mappedPort, err := container.MappedPort(ctx, containers.VaultPort.Port())
	require.NoError(t, err, "Failed to mapped container pod")

	vault, err := New(Config{
		Address: fmt.Sprintf("http://%s:%s", host, mappedPort.Port()),
		Token:   containers.VaultToken,
	})
	require.NoError(t, err, "Failed while creating session to Vault")

	transit := vault.NewTransit(secretPath, secretKey)

	message := "Hello World!"

	wrapped, err := transit.Encrypt(ctx, EncryptRequest{
		Plaintext: []byte(message),
	})
	require.NoError(t, err, "Failed while encrypting message")
	require.Equal(t, wrapped.KeyVersion, uint(1), "key_version for single encryption is invalid")

	unwrapped, err := transit.Decrypt(ctx, DecryptRequest{
		Ciphertext: wrapped.Ciphertext,
	})
	require.NoError(t, err, "Failed while decrypt ciphertext")
	require.Equal(t, message, string(unwrapped.Plaintext), "The original value and the decrypted result do not match")

	messages := []string{
		"f8844747-35dc-4576-ad7f-27c40d8dd0ea", "3003311c-fc61-4bb9-ac10-ad58b9860f76", "2241c280-8183-4f8e-941a-e090f26ecc05", "2d3d32a3-47a1-4c8e-ad63-1e4b5b49c264",
	}

	batchInput := make([]BatchEncryptInput, len(messages))
	for i, message := range messages {
		batchInput[i] = BatchEncryptInput{
			Plaintext: []byte(message),
		}
	}
	wrappeds, err := transit.BatchEncrypt(ctx, BatchEncryptRequest{
		BatchInput: batchInput,
	})
	require.NoError(t, err, "Failed while performing batch encryption")
	require.Len(t, wrappeds, 4, "Total encrypted data does not match")

	for i, wrapped := range wrappeds {
		unwrapped, err := transit.Decrypt(ctx, DecryptRequest{
			Ciphertext: wrapped.Ciphertext,
		})
		require.NoError(t, err, "Failed while decrypt ciphertext on bacth proses")

		assert.Equal(t, wrapped.KeyVersion, uint(1))
		assert.Equal(t, messages[i], string(unwrapped.Plaintext))
	}
}
