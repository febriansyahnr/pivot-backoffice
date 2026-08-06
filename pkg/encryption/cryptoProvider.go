package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"

	"go.mozilla.org/pkcs7"
)

type CryptoProvider interface {
	DecryptHybrid(request *DataEncryption) ([]byte, error)
	DecryptAESCBC(key, iv, ciphertext []byte) ([]byte, error)
	EncryptPKCS7(certPem, plaintext []byte) (string, error)
}

type provider struct{}

func NewCryptoProvider() CryptoProvider {
	return &provider{}
}

func (p *provider) DecryptHybrid(request *DataEncryption) ([]byte, error) {

	if request.PrivateKeyPEM == "" && request.PrivateKeyPKCS8 == "" {
		return nil, errors.New("private key used cannot be empty")

	} else if request.PrivateKeyPEM != "" && request.PrivateKeyPKCS8 != "" {
		return nil, errors.New("can only use one private key, namely PEM or PKCS8")
	}

	var (
		err             error
		privateKeyBytes []byte
	)

	if request.PrivateKeyPKCS8 != "" {
		if privateKeyBytes, err = base64.StdEncoding.DecodeString(request.PrivateKeyPKCS8); err != nil {
			return nil, fmt.Errorf("decode base64: %w", err)
		}

	} else {
		pemBlock, _ := pem.Decode([]byte(request.PrivateKeyPEM))
		if pemBlock == nil {
			return nil, errors.New("failed to decode PEM block")
		}
		privateKeyBytes = pemBlock.Bytes
	}

	privateKeyAny, err := x509.ParsePKCS8PrivateKey(privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}

	privateKey, ok := privateKeyAny.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("not an RSA private key")
	}

	var encryptedKey, nonce, ciphertext []byte

	decodePayload := map[string][2]any{
		"encrypted key": {&encryptedKey, request.EncryptedKey},
		"nonce":         {&nonce, request.Nonce},
		"ciphertext":    {&ciphertext, request.Ciphertext},
	}
	for key, data := range decodePayload {
		if *data[0].(*[]byte), err = base64.StdEncoding.DecodeString(data[1].(string)); err != nil {
			return nil, fmt.Errorf("failed to decode base64 %s: %w", key, err)
		}
	}

	dataEncryptionKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, encryptedKey, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt data encryption key: %w", err)
	}

	block, err := aes.NewCipher(dataEncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("create aes cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create cipher aes gcm: %w", err)
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt ciphertext: %w", err)
	}
	return plaintext, nil
}

func (p *provider) DecryptAESCBC(key, iv, ciphertext []byte) ([]byte, error) {
	if len(iv) != aes.BlockSize {
		return nil, fmt.Errorf("iv must be %d bytes, got %d", aes.BlockSize, len(iv))
	}

	if len(ciphertext) == 0 || len(ciphertext)%aes.BlockSize != 0 {
		return nil, errors.New("ciphertext length must be a non-zero multiple of aes block size")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("new aes cipher: %w", err)
	}

	plaintext := make([]byte, len(ciphertext))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(plaintext, ciphertext)

	plaintext, err = pkcs7Unpad(plaintext, aes.BlockSize)
	if err != nil {
		return nil, fmt.Errorf("remove pkcs7 padding: %w", err)
	}
	return plaintext, nil
}

func (p *provider) EncryptPKCS7(certPem, plaintext []byte) (string, error) {
	pemBlock, _ := pem.Decode(certPem)
	if pemBlock == nil {
		return "", errors.New("failed to decode PEM block")
	}

	cert, err := x509.ParseCertificate(pemBlock.Bytes)
	if err != nil {
		return "", fmt.Errorf("parse PKIX public key: %w", err)
	}

	ciphertext, err := pkcs7.Encrypt(plaintext, []*x509.Certificate{cert})
	if err != nil {
		return "", fmt.Errorf("encrypt pkcs7 plaintext: %w", err)
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("empty data")
	}

	padLen := int(data[len(data)-1])
	if padLen == 0 || padLen > blockSize {
		return nil, errors.New("invalid pkcs7 padding")
	}

	for _, b := range data[len(data)-padLen:] {
		if int(b) != padLen {
			return nil, errors.New("invalid pkcs7 padding byte")
		}
	}

	return data[:len(data)-padLen], nil
}
