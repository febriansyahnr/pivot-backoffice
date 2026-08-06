package containers

import (
	"context"
	"fmt"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	VaultToken string   = "ddd4e7a1-1580-4beb-ab42-7b89bb0aaf10" // NOSONAR
	VaultPort  nat.Port = "8200"                                 // NOSONAR
)

// Creating a Hashicorp Vault container for testing purposes. After stopping the container will be automatically deleted.
func NewVault(ctx context.Context) (testcontainers.Container, error) {
	request := testcontainers.ContainerRequest{
		Image:        "hashicorp/vault:1.20", // NOSONAR
		ExposedPorts: []string{"8200/tcp"},
		WaitingFor:   wait.ForLog("Development mode should NOT be used in production installations!"),
		Env: map[string]string{
			"VAULT_ADDR":              fmt.Sprintf("http://127.0.0.1:%s", VaultPort), // NOSONAR
			"VAULT_TOKEN":             VaultToken,
			"VAULT_DEV_ROOT_TOKEN_ID": VaultToken,
		},
	}

	return testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: request,
		Started:          true,
	})
}
