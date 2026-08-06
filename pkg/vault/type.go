package vault

import "context"

type IVaultKeyValue interface {
	GetSecret(ctx context.Context) (*Secret, error)
	GetSecretVersion(ctx context.Context, version int) (*Secret, error)
	GetSecretKeyString(ctx context.Context, key string) (*SecretKey[string], error)
	GetSecretKeyVersionString(ctx context.Context, version int, key string) (*SecretKey[string], error)
}

type IVaultTransit interface {
	// Encrypts plaintext data using transit secret engine. Supports optional key versioning and context-based key derivation.
	Encrypt(ctx context.Context, request EncryptRequest) (*EncryptResponse, error)

	// Encrypts multiple plaintext items in a single Vault API call using batch operations.
	BatchEncrypt(ctx context.Context, request BatchEncryptRequest) ([]EncryptResponse, error)

	// Decrypts ciphertext using Transit secret engine. Supports context-based key derivation if required by the key configuration.
	Decrypt(ctx context.Context, request DecryptRequest) (*DecryptResponse, error)
}
