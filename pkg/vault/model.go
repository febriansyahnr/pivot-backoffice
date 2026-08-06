package vault

type EncryptRequest struct {
	// Specifies the version of the key to use for the operation.
	// Leave KeyVersion unset to use the latest version. KeyVersion must be unset or
	// greater than or equal to the associated min_encryption_version value.
	KeyVersion uint

	// Plaintext to be encrypt
	Plaintext []byte

	// This is required if key derivation is enabled for this key
	Context []byte
}

type BatchEncryptRequest struct {
	// Specifies the version of the key to use for the operation.
	// Leave KeyVersion unset to use the latest version. KeyVersion must be unset or
	// greater than or equal to the associated min_encryption_version value.
	KeyVersion uint

	// Specifies a list of items to be encrypted in a single batch.
	BatchInput []BatchEncryptInput
}

type BatchEncryptInput struct {
	// Plaintext to be encrypt
	Plaintext []byte

	// This is required if key derivation is enabled for this key
	Context []byte
}

type DecryptRequest struct {
	// Specifies the ciphertext to decrypt.
	Ciphertext string

	// Specifies the context for key derivation. This is required if key derivation is enabled.
	Context []byte
}

type EncryptResponse struct {
	Ciphertext string `json:"ciphertext"`
	KeyVersion uint   `json:"key_version"`
}

type DecryptResponse struct {
	Plaintext []byte
}
