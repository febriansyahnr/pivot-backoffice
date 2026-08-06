package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ICrypto for cryptography operations
type ICrypto interface {
	GenerateRandomPKCS8Key() ([]byte, error)
	Encrypt(plainText, key string) (string, error)
	Decrypt(ciphertext, secretKey string) (*[]byte, error)
	SecretKeyFromUUID(id uuid.UUID) string
	ToPublicKey(data []byte) (pubKey *rsa.PublicKey, err error)
}

// Crypto struct for cryptography operations
type Crypto struct {
	MerchantSecretKey   string
	PublicKey           []byte
	PrivateKey          []byte
	EncryptedPublicKey  string
	EncryptedPrivateKey string
	Error               error
}

// New create new instance of Crypto
func New() ICrypto {
	return &Crypto{}
}

func (c *Crypto) SecretKeyFromUUID(id uuid.UUID) string {
	secretKey := strings.ReplaceAll(id.String(), "-", "")
	return secretKey
}

// GenerateRandomPKCS8Key generate random pkcs#8 key pair
func (c *Crypto) GenerateRandomPKCS8Key() ([]byte, error) {

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		zap.Error(err)
		return nil, err
	}

	privateKetBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		zap.Error(err)
		return nil, err
	}

	privateKeyPem := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKetBytes,
	})

	return privateKeyPem, nil
}

func (c *Crypto) Encrypt(plainText, key string) (string, error) {
	aes, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(aes)
	if err != nil {
		return "", err
	}

	// We need a 12-byte nonce for GCM (modifiable if you use cipher.NewGCMWithNonceSize())
	// A nonce should always be randomly generated for every encryption.
	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		return "", err
	}

	// ciphertext here is actually nonce+ciphertext
	// So that when we decrypt, just knowing the nonce size
	// is enough to separate it from the ciphertext.
	ciphertext := gcm.Seal(nonce, nonce, []byte(plainText), nil)

	result := base64.StdEncoding.EncodeToString(ciphertext)

	return result, nil
}

func (c *Crypto) Decrypt(ciphertext, secretKey string) (result *[]byte, err error) {

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovery panic: %v", r)
		}
	}()

	ciphertextBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, err
	}

	// Create cipher block
	block, err := aes.NewCipher([]byte(secretKey))
	if err != nil {
		return nil, err
	}

	// Create GCM cipher mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Get the nonce size
	nonceSize := gcm.NonceSize()

	// Check if the ciphertext is at least as long as the nonce
	if len(ciphertextBytes) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	// Extract the nonce from the beginning of the ciphertext
	// This is the nonce that was randomly generated during encryption
	extractedNonce, encryptedData := ciphertextBytes[:nonceSize], ciphertextBytes[nonceSize:]

	// Decrypt the data using the extracted nonce
	// #nosec G407 - This is not a hardcoded nonce, it's extracted from the ciphertext
	plaintext, err := gcm.Open(nil, extractedNonce, encryptedData, nil)
	if err != nil {
		return nil, err
	}

	return &plaintext, nil
}

// EncryptPassword encrypt password using built-in sha256
func EncryptPassword(pass string) string {
	hasher := sha256.New()
	hasher.Write([]byte(pass))

	return hex.EncodeToString(hasher.Sum(nil))
}

func (c *Crypto) ToPublicKey(data []byte) (pubKey *rsa.PublicKey, err error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	switch block.Type {
	case "PUBLIC KEY":
		k, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		pubKey, _ = k.(*rsa.PublicKey)

	case "RSA PUBLIC KEY":
		pubKey, err = x509.ParsePKCS1PublicKey(block.Bytes)

	default:
		return nil, errors.New("invalid key type")
	}

	return
}
