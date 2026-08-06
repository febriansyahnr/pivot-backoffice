package encryption

type DataEncryption struct {
	EncryptedKey    string `json:"encryptedKey" validate:"required,base64"`
	Nonce           string `json:"nonce" validate:"required,base64"`
	Ciphertext      string `json:"ciphertext" validate:"required,base64"`
	PrivateKeyPEM   string `json:"-" validate:"-"` // Plaintext
	PrivateKeyPKCS8 string `json:"-" validate:"-"` // Encoded base64
}
