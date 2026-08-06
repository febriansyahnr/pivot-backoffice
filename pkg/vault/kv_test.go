package vault_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	. "github.com/paper-indonesia/pivot-backoffice/pkg/vault"
	"github.com/paper-indonesia/pivot-backoffice/test/containers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func TestIntegrationVaultKeyValue(t *testing.T) {
	if os.Getenv(constant.IntegrationTestEnv) != "1" {
		t.Skip(constant.SkipIntegrationTest)
	}

	const (
		mountPath   = "secret"  // NOSONAR
		secretPath  = "testing" // NOSONAR
		secretKey   = "key1"    // NOSONAR
		secretValue = "value1"  // NOSONAR

		secretKeyNotFound = "KEY_NOT_FOUND" // NOSONAR
	)

	ctx := t.Context()

	container, err := containers.NewVault(ctx)
	require.NoError(t, err)

	defer container.Stop(ctx, nil)

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, containers.VaultPort.Port())
	require.NoError(t, err)

	_, _, err = container.Exec(ctx, []string{
		"vault", "kv", "put", fmt.Sprintf("%s/%s", mountPath, secretPath), fmt.Sprintf("%s=%s", secretKey, secretValue),
	})
	require.NoError(t, err)

	client, err := New(Config{
		Address: fmt.Sprintf("http://%s:%s", host, port.Port()), // NOSONAR
		Token:   containers.VaultToken,
	})
	require.NoError(t, err)

	kv := client.NewKeyValue(mountPath, secretPath)

	t.Run("KV:Get Secret Key String", func(t *testing.T) {
		secret, err := kv.GetSecretKeyString(ctx, secretKey)
		require.NoError(t, err)
		assert.Equal(t, int(1), secret.Version)
		assert.Equal(t, secretValue, secret.Value)
	})

	t.Run("KV:Get Secret Key String - Key Not Found", func(t *testing.T) {
		secret, err := kv.GetSecretKeyString(ctx, secretKeyNotFound) // NOSONAR
		assert.Equal(t, fmt.Errorf("key %s: %w", secretKeyNotFound, ErrKeyNotFound), err)
		assert.Nil(t, secret)
	})

	t.Run("KV:Get Secret Key Version String", func(t *testing.T) {
		group := new(errgroup.Group)

		_, err := kv.GetSecretKeyVersionString(ctx, 1, secretKey) // Init Request
		require.NoError(t, err)

		for range 12 {
			group.Go(func() error {
				secret, err := kv.GetSecretKeyVersionString(ctx, 1, secretKey)
				if err != nil {
					return err
				}
				assert.Equal(t, int(1), secret.Version)
				assert.Equal(t, secretValue, secret.Value)

				return nil
			})
		}

		require.NoError(t, group.Wait())
	})

	t.Run("KV:Get Secret Key Version String - Key Not Found", func(t *testing.T) {
		secret, err := kv.GetSecretKeyVersionString(ctx, 1, secretKeyNotFound) // NOSONAR
		assert.Equal(t, fmt.Errorf("key %s: %w", secretKeyNotFound, ErrKeyNotFound), err)
		assert.Nil(t, secret)
	})
}
