package encryption

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	kms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/kms/apiv1/kmspb"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"google.golang.org/api/iterator"
)

const (
	PathKmsVersion = "/cryptoKeyVersions/"
)

type GCSClient interface {
	EncryptSymmetric(
		ctx context.Context,
		message string,
		version string,
	) (string, error)
	DecryptSymmetric(
		ctx context.Context,
		ciphertext []byte,
	) (string, error)

	GetLatestKMSVersion(ctx context.Context) (string, error)
}

type gcsClient struct {
	cfg *config.Secret
}

func NewGCS(cfg *config.Secret) GCSClient {
	return &gcsClient{
		cfg: cfg,
	}
}

func (gcs *gcsClient) GetLatestKMSVersion(ctx context.Context) (string, error) {
	client, err := kms.NewKeyManagementClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create kms client: %w", err)
	}
	defer client.Close()

	req := &kmspb.ListCryptoKeyVersionsRequest{
		Parent: gcs.cfg.GoogleKMSKey,
		Filter: "state=ENABLED", // Only list enabled key versions
	}

	it := client.ListCryptoKeyVersions(ctx, req)

	var latestVersion *kmspb.CryptoKeyVersion
	for {
		version, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to list key versions: %w", err)
		}
		if latestVersion == nil || version.CreateTime.AsTime().After(latestVersion.CreateTime.AsTime()) {
			latestVersion = version
		}

	}

	if latestVersion == nil {
		return "", fmt.Errorf("no enabled key versions found")
	}

	latest := strings.Split(latestVersion.Name, "/")
	if len(latest) == 0 {
		return "", fmt.Errorf("no enabled key versions found")
	}

	latestVersionNumber := latest[len(latest)-1]
	if _, err := strconv.Atoi(latestVersionNumber); err != nil {
		return "", fmt.Errorf("invalid version: %w", err)
	}

	return latestVersionNumber, nil
}

func (gcs *gcsClient) DecryptSymmetric(
	ctx context.Context,
	ciphertext []byte,
) (string, error) {
	client, err := kms.NewKeyManagementClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create kms client: %w", err)
	}
	defer client.Close()

	req := &kmspb.DecryptRequest{
		Name:       gcs.cfg.GoogleKMSKey,
		Ciphertext: ciphertext,
	}

	result, err := client.Decrypt(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(result.Plaintext), nil
}

func (gcs *gcsClient) EncryptSymmetric(
	ctx context.Context,
	message string,
	version string,
) (string, error) {
	client, err := kms.NewKeyManagementClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create kms client: %w", err)
	}
	defer client.Close()

	if version == "" {
		// Create a new version of the key
		version, err = gcs.createKeyVersion(ctx, client, gcs.cfg.GoogleKMSKey)
		if err != nil {
			return "", err
		}
	}

	versionUse := gcs.cfg.GoogleKMSKey + PathKmsVersion + version

	req := &kmspb.EncryptRequest{
		Name:      versionUse,
		Plaintext: []byte(message),
	}

	result, err := client.Encrypt(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt data: %w", err)
	}

	return base64.StdEncoding.EncodeToString(result.Ciphertext), nil
}

func (gcs *gcsClient) createKeyVersion(
	ctx context.Context,
	client *kms.KeyManagementClient,
	keyName string,
) (string, error) {
	// Call the API to create a new key version
	req := &kmspb.CreateCryptoKeyVersionRequest{
		Parent: keyName,
	}
	result, err := client.CreateCryptoKeyVersion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to create new key version: %w", err)
	}

	return result.Name, nil
}
